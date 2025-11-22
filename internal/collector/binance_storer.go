package collector

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/internal/config"
	"Centralized-Data-Collector/internal/db"
	"Centralized-Data-Collector/internal/model"
	pkg "Centralized-Data-Collector/pkg/http"
	"Centralized-Data-Collector/pkg/logger"
	"context"
	"strconv"
	"time"
)

type BinanceStorer struct {
	signalProcessor *BlockingSignalProcessor
}

func NewBinanceStorer() *BinanceStorer {
	return &BinanceStorer{
		signalProcessor: NewBlockingSignalProcessor(100, dateFillHandler),
	}
}

func (c *BinanceStorer) parsePushedMsgToTradeModel(ctx context.Context, pushedMsg *binance_define.WSSinglePushMsg) *model.BinanceAggTrade {
	// 	// 解析 pushedMsg 为 TradeModel
	data := pushedMsg.Data.(binance_define.BinanceAggTrade)
	tradeModel := &model.BinanceAggTrade{
		Event:      data.Event,
		EventTime:  data.EventTime,
		Symbol:     data.Symbol,
		AggID:      data.AggID,
		FirstID:    data.FirstID,
		LastID:     data.LastID,
		Timestamp:  data.Timestamp,
		BuyerMaker: data.BuyerMaker,
		Ignore:     data.Ignore,
		CreatedAt:  time.Now(),
	}
	// Price 和 Quantity 从字符串转换为 float64
	if price, err := strconv.ParseFloat(data.Price, 64); err == nil {
		tradeModel.Price = price
	} else {
		// 处理解析错误
		tradeModel.Price = 0
	}

	if qty, err := strconv.ParseFloat(data.Quantity, 64); err == nil {
		tradeModel.Quantity = qty
	} else {
		tradeModel.Quantity = 0
	}
	return tradeModel

}

func (c *BinanceStorer) parsePushedMsgToCandleModel(ctx context.Context, pushedMsg *binance_define.WSSinglePushMsg) *model.Kline1m {
	// 解析 pushedMsg 为 CandleModel
	wsKline := pushedMsg.Data.(binance_define.KlineMessage)
	dbKline := &model.Kline1m{
		EventTime:    wsKline.Ev,
		Pair:         wsKline.K.S,
		StartTime:    wsKline.K.T,
		EndTime:      wsKline.K.TEnd,
		Interval:     wsKline.K.I,
		FirstTradeID: wsKline.K.F,
		LastTradeID:  wsKline.K.L,
		TradeCount:   int(wsKline.K.N),
		IsClosed:     wsKline.K.X,
		Block:        wsKline.K.B,
		CreatedAt:    time.Now(),
	}

	// 将字符串字段转换为 float64
	if f, err := strconv.ParseFloat(wsKline.K.O, 64); err == nil {
		dbKline.OpenPrice = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.C, 64); err == nil {
		dbKline.ClosePrice = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.H, 64); err == nil {
		dbKline.HighPrice = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.Lw, 64); err == nil {
		dbKline.LowPrice = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.V, 64); err == nil {
		dbKline.Volume = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.Q, 64); err == nil {
		dbKline.QuoteVolume = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.VB, 64); err == nil {
		dbKline.BuyVolume = f
	}
	if f, err := strconv.ParseFloat(wsKline.K.QB, 64); err == nil {
		dbKline.BuyQuoteVolume = f
	}
	return dbKline
}

func (c *BinanceStorer) parsePushedMsgToTickereModel(ctx context.Context, pushedMsg *binance_define.WSSinglePushMsg) *model.TickerModel {
	// 解析 pushedMsg 为 TickerModel
	ticker := pushedMsg.Data.(binance_define.Binance24hrMiniTicker)
	return &model.TickerModel{
		Event:       ticker.Event,
		EventTime:   ticker.EventTime,
		Symbol:      ticker.Symbol,
		Close:       ticker.Close,
		Open:        ticker.Open,
		High:        ticker.High,
		Low:         ticker.Low,
		Volume:      ticker.Volume,
		QuoteVolume: ticker.QuoteVolume,
		CreatedAt:   time.Now(), // 自动填充
	}
}

func (c *BinanceStorer) StoreData(ctx context.Context, pushedMsgs []*binance_define.WSSinglePushMsg) {
	tradeModels := []*model.BinanceAggTrade{}
	klineModels := []*model.Kline1m{}
	tickerModels := []*model.TickerModel{}

	marketDataFillModels := []*model.MarketDataFill{}
	for i := 0; i < len(pushedMsgs); i++ {
		msg := pushedMsgs[i]

		switch msg.EventType {
		case config.AggTradeStream:
			tradeModel := c.parsePushedMsgToTradeModel(ctx, msg)
			// logger.Debug("aggTrade %d %s %d %d", tradeModel.AggID, tradeModel.Event, tradeModel.EventTime, tradeModel.FirstID)
			tradeModels = append(tradeModels, tradeModel)
			if msg.IsFirst {
				tradeLast, err := db.GetLatestAggTradeInfo(ctx, tradeModel.Symbol)
				if err != nil {
					logger.Error("GetLatestKLine1mInfo err %s", err)
				}
				if tradeLast != nil {
					marketDataFillModels = append(marketDataFillModels, createDateFillModel(config.AggTradeStream, tradeLast.EventTime, tradeModel.EventTime, tradeModel.Symbol))
				}
			}
		case config.KlineStream:
			klineModel := c.parsePushedMsgToCandleModel(ctx, msg)

			klineModels = append(klineModels, klineModel)
			if msg.IsFirst {
				klineLast, err := db.GetLatestKLine1mInfo(ctx, klineModel.Pair)
				if err != nil {
					logger.Error("GetLatestKLine1mInfo err %s", err)
				}
				if klineLast != nil {
					marketDataFillModels = append(marketDataFillModels, createDateFillModel(config.KlineStream, klineLast.StartTime, klineModel.StartTime, klineLast.Pair))
				}
			}
		case config.MiniTickerStream:
			tickerModels = append(tickerModels, c.parsePushedMsgToTickereModel(ctx, msg))
		}
		// TODO 后续需要检查一个数据，aggtrade中的数据中的交易id和kline中的交易id是否是一致的（逐条推送和聚合推送的在两个数据中id）
	}

	if len(tradeModels) > 0 {
		for {
			err := db.BatchInsertBinanceAggTrades(ctx, tradeModels)
			if err == nil {
				logger.Info("Stored %d trade records from Binance", len(tradeModels))
				// tradeModels = tradeModels[:0]
				break
			}
			logger.Error("Failed to batch insert Binance trades: err = %v", err)
			time.Sleep(100 * time.Millisecond)

		}
	}

	if len(klineModels) > 0 {
		for {
			//先去重，保证每分钟只有一条线存在
			err := db.BatchDeleteAndInsertKline1M(ctx, klineModels)
			if err == nil {
				logger.Info("Stored %d kline records from Binance", len(klineModels))
				break
			}
			logger.Error("Failed to batch insert Binance kline: err = %v", err)
			time.Sleep(100 * time.Millisecond)

		}
	}
	if len(marketDataFillModels) > 0 {
		for {
			err := db.BatchInsertBinanceDataFill(ctx, marketDataFillModels)
			if err == nil {
				logger.Info("Stored %d data fill records from Binance", len(marketDataFillModels))
				c.signalProcessor.Signal()
				break
			} else {
				logger.Error("Failed to batch insert Binance data fill: err = %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

		}

	}

	// if len(tickerModels) > 0 {
	// 	for {
	// 		err := db.BatchInsertBinanceTicker(ctx, tickerModels)
	// 		if err == nil {
	// 			logger.Info("Stored %d ticker records from Binance", len(tickerModels))
	// 			break
	// 		}
	// 		logger.Error("Failed to batch insert Binance ticker: err = %v", err)
	// 		time.Sleep(100 * time.Millisecond)
	// 	}

	// }

}

// 处理数据补充
func dateFillHandler(ctx context.Context) error {
	//查询data_fill表中是否有需要使用短连接拉取数据
	for {
		fillData, err := db.GetBinanceDataFill(ctx)
		if err != nil {
			logger.Debug("GetBinanceDataFill failed")
			// time.Sleep(100 * time.Millisecond)
		}
		if fillData == nil {
			break
		}

		err2 := opDataFill(ctx, fillData)
		if err2 != nil {
			logger.Debug("opDataFill failed %s", err2)

		}
		time.Sleep(200 * time.Millisecond)

	}

	return nil
}

func fillingKLine(ctx context.Context, fillData *model.MarketDataFill) error {

	//开始查询短连接
	klineResponses, err := pkg.FetchKlines(fillData.Symbol, config.Interval1m, config.DefaultLimit, fillData.EventStartTime, fillData.EventEndTime-config.OneminuteTimestamp)
	if err != nil {
		logger.Debug("FetchKlines failed: %s", err)
		time.Sleep(500 * time.Millisecond)
		return err
	}
	logger.Debug("EventStartTime: %d", fillData.EventStartTime)
	logger.Debug("EventEndTime: %d", fillData.EventEndTime)
	logger.Debug("FetchKlines: %s", len(klineResponses))
	if len(klineResponses) == 0 {
		err := db.UpdateBinanceDataFill(ctx, fillData.Event, fillData.Symbol, fillData.EventStartTime, fillData.EventStartTime, true)
		if err != nil {
			logger.Debug("UpdateBinanceDataFill failed: %s", err)
			time.Sleep(200 * time.Millisecond)
			return err
		}
		return nil
	}

	klineModels := ConvertRESTKlines(fillData.Symbol, klineResponses)
	txErr := db.BatchInsertKline1mAndFillData(ctx, klineModels, fillData)
	if txErr != nil {
		logger.Debug("UpdateBinanceDataFill failed: %s", txErr)
		time.Sleep(200 * time.Millisecond)
		return txErr
	}
	return nil
}

func fillingTrade(ctx context.Context, fillData *model.MarketDataFill) error {

	//开始查询短连接
	tradesResponses, err := pkg.FetchAggTrades(fillData.Symbol, config.DefaultLimit, fillData.EventStartTime, fillData.EventEndTime)
	if err != nil {
		time.Sleep(200 * time.Millisecond)
		logger.Debug("FetchAggTrades failed: %s", err)
		return err
	}

	if len(tradesResponses) == 0 {
		err := db.UpdateBinanceDataFill(ctx, fillData.Event, fillData.Symbol, fillData.EventStartTime, fillData.EventStartTime, true)
		if err != nil {
			logger.Debug("UpdateBinanceDataFill failed: %s", err)
			time.Sleep(200 * time.Millisecond)
			return err
		}
		return nil
	}
	tradeModels := ConvertAggTrades(fillData.Symbol, tradesResponses)
	txErr := db.BatchInsertAggTradeAndFillData(ctx, tradeModels, fillData)
	if txErr != nil {
		logger.Debug("BatchInsertAggTradeAndFillData failed: %s", txErr)
		time.Sleep(200 * time.Millisecond)
		return txErr
	}
	return nil
}

func opDataFill(ctx context.Context, fillData *model.MarketDataFill) error {
	if fillData != nil {
		switch fillData.Event {
		case config.KlineStream:
			klineErr := fillingKLine(ctx, fillData)
			if klineErr != nil {
				logger.Debug("fillingKLine failed:%s", klineErr)
				return klineErr
			}
		case config.AggTradeStream:
			tradeErr := fillingTrade(ctx, fillData)
			if tradeErr != nil {
				logger.Debug("fillingTrade failed:%s", tradeErr)
				return tradeErr
			}
		default: //price需要调研能否取回历史数据
		}

	}
	return nil
}

func createDateFillModel(event string, eventStartime int64, eventEndTime int64, symbol string) *model.MarketDataFill {
	return &model.MarketDataFill{
		Event:          event,
		EventStartTime: eventStartime,
		EventEndTime:   eventEndTime,
		Symbol:         symbol,
		CreatedAt:      time.Now(),
	}
}

func ConvertRESTKlines(pair string, responses []pkg.KlineResponse) []*model.Kline1m {
	result := make([]*model.Kline1m, 0, len(responses))

	for _, r := range responses {
		k := model.Kline1m{
			Pair:           pair,
			StartTime:      r.OpenTime,
			EndTime:        r.CloseTime,
			Interval:       config.Interval1m,
			FirstTradeID:   0,
			LastTradeID:    0, // REST 不提供
			OpenPrice:      parseFloat(r.Open),
			ClosePrice:     parseFloat(r.Close),
			HighPrice:      parseFloat(r.High),
			LowPrice:       parseFloat(r.Low),
			Volume:         parseFloat(r.Volume),
			TradeCount:     int(r.NumberOfTrades),
			IsClosed:       true, // REST 返回的数据都是已收盘的
			QuoteVolume:    parseFloat(r.QuoteAssetVolume),
			BuyVolume:      parseFloat(r.TakerBuyBaseAssetVolume),
			BuyQuoteVolume: parseFloat(r.TakerBuyQuoteAssetVolume),
			Block:          r.Ignore,
			EventTime:      r.CloseTime,
		}

		result = append(result, &k)
	}

	return result
}

// ConvertAggTrades 将 Binance API 的 aggTrades 响应转换为 []*model.BinanceAggTrade
func ConvertAggTrades(symbol string, responses []pkg.AggTradeResponse) []*model.BinanceAggTrade {
	result := make([]*model.BinanceAggTrade, 0, len(responses))

	for _, r := range responses {
		k := model.BinanceAggTrade{
			Event:      config.AggTradeStream, // REST aggTrades 没有 e 字段，统一写 aggTrade
			EventTime:  r.Timestamp,           // 使用交易时间作为 event time（aggTrade 没有 E 字段）
			Symbol:     symbol,
			AggID:      r.AggTradeID,
			Price:      parseFloat(r.Price),
			Quantity:   parseFloat(r.Quantity),
			FirstID:    r.FirstID,
			LastID:     r.LastID,
			Timestamp:  r.Timestamp,
			BuyerMaker: r.IsBuyerMaker,
			Ignore:     r.Ignore,
			CreatedAt:  time.Now(),
		}

		result = append(result, &k)
	}

	return result
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
