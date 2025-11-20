package collector

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/internal/db"
	"Centralized-Data-Collector/internal/model"
	pkg "Centralized-Data-Collector/pkg/http"
	"Centralized-Data-Collector/pkg/logger"
	"context"
	"encoding/json"
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

	// 解析 pushedMsgs 为 TradeModel 和 PriceModel
	for i := 0; i < len(pushedMsgs); i++ {
		msg := pushedMsgs[i]

		//如果是重启或者重连后 isFirst字段为ture,此时需要写入数据库，用于后续按区间拉取数据
		// isFirst :=
		// logger.Debug("发现重连后的新代币对 %s %s", msg.IsFirst, msg.EventType)

		switch msg.EventType {
		case "aggTrade":
			tradeModels = append(tradeModels, c.parsePushedMsgToTradeModel(ctx, msg))
		case "kline":
			klineModel := c.parsePushedMsgToCandleModel(ctx, msg)

			// dataIndent, _ := json.MarshalIndent(klineModel, "", "  ")
			// logger.Debug("kline JSON:\n", string(dataIndent))

			klineModels = append(klineModels, klineModel)
			if msg.IsFirst {
				//需要填充表中增加数据
				//先查询是否有存在最新的
				klineLast, err := db.GetLatestKLine1mInfo(ctx, klineModel.Pair)
				if err != nil {
					logger.Debug("GetLatestKLine1mInfo err %s", err)
				}
				if klineLast != nil {
					logger.Debug("GetLatestKLine1mInfo %s", klineLast)
					logger.Debug("GetLatestKLine1mInfo %s %s", klineLast.Pair, klineLast.EventTime)
					marketDataFillModels = append(marketDataFillModels, createDateFillModel("kline", klineLast.StartTime, klineModel.StartTime, klineLast.Pair))
				}
			}
		case "24hrMiniTicker":
			tickerModels = append(tickerModels, c.parsePushedMsgToTickereModel(ctx, msg))
		}

		// TODO 后续需要检查一个数据，aggtrade中的数据中的交易id和kline中的交易id是否是一致的（逐条推送和聚合推送的在两个数据中id）

		// if len(tradeModels) > 0 {
		// 	for {
		// 		err := db.BatchInsertBinanceAggTrades(ctx, tradeModels)
		// 		if err == nil {
		// 			logger.Info("Stored %d trade records from Binance", len(tradeModels))
		// 			break
		// 		}
		// 		logger.Error("Failed to batch insert Binance trades: err = %v", err)
		// 		time.Sleep(100 * time.Millisecond)
		// 	}
		// }

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

		if len(marketDataFillModels) > 0 {
			for {
				err := db.BatchInsertBinanceDataFill(ctx, marketDataFillModels)
				if err == nil {
					logger.Info("Stored %d data fill records from Binance", len(marketDataFillModels))

					// 连续发信号（阻塞式入队）
					logger.Debug("send signal")
					c.signalProcessor.Signal()
					// p.Stop()
					// fmt.Println("processor stopped")
					break
				} else {
					logger.Error("Failed to batch insert Binance data fill: err = %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

			}

		}
	}
}

// func main() {
// 	symbol := "BTCUSDT"
// 	interval := "1m"
// 	limit := 5

// 	// 拉取Kline
// 	klines, err := FetchKlines(symbol, interval, limit, 0, 0)
// 	if err != nil {
// 		fmt.Println("FetchKlines error:", err)
// 		return
// 	}
// 	fmt.Println("Klines:")
// 	for _, k := range klines {
// 		fmt.Printf("Time: %v, Open: %s, Close: %s\n", time.UnixMilli(k.OpenTime), k.Open, k.Close)
// 	}

// 	// 拉取AggTrade
// 	trades, err := FetchAggTrades(symbol, limit, 0, 0)
// 	if err != nil {
// 		fmt.Println("FetchAggTrades error:", err)
// 		return
// 	}
// 	fmt.Println("\nAggTrades:")
// 	for _, t := range trades {
// 		fmt.Printf("ID: %d, Price: %s, Qty: %s, Timestamp: %v\n", t.AggTradeID, t.Price, t.Quantity, time.UnixMilli(t.Timestamp))
// 	}
// }

func dateFillHandler(ctx context.Context) error {
	logger.Debug("exampleHandler")

	//查询data_fill表中是否有需要使用短连接拉取数据
	for {

		fillData, err := db.GetBinanceDataFill(ctx)
		dataIndents, err := json.MarshalIndent(fillData, "", "  ")
		logger.Debug("data fill JSON:\n", string(dataIndents))

		if err != nil {
			logger.Debug("GetBinanceDataFill failed")
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if fillData != nil {
			//开始查询短连接
			logger.Debug("FetchKlines param %s %d %s ", fillData.Symbol, fillData.EventStartTime, fillData.EventEndTime)
			klineResponses, err := pkg.FetchKlines(fillData.Symbol, "1m", 1000, fillData.EventStartTime, fillData.EventEndTime-60000)

			dataIndents, err := json.MarshalIndent(klineResponses, "", "  ")
			logger.Debug("get from binance api JSON:\n", string(dataIndents))

			if err != nil {
				logger.Debug("FetchKlines failed: %s", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if len(klineResponses) == 0 {
				err := db.UpdateBinanceDataFill(ctx, fillData.Symbol, fillData.EventStartTime, fillData.EventStartTime, true)
				if err != nil {
					logger.Debug("UpdateBinanceDataFill failed: %s", err)
					time.Sleep(2000 * time.Millisecond)
				}
				continue
			}
			//如果第一条startTime与查询条件中的startTime一致，就更新它，其他插入
			//如果返回的数量少于limit 就说明还有，需要再次发起

			klineModels := ConvertRESTKlines(fillData.Symbol, klineResponses)

			//删除klineModels对应表中的第一条数据
			deleteErr := db.DeleteBinanceKline(ctx, fillData.Symbol, klineModels[0].StartTime, klineModels[0].EndTime)
			if deleteErr != nil {
				logger.Debug("DeleteBinanceKline failed: %s", deleteErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			logger.Debug("DeleteBinanceKline %s 成功", fillData.Symbol)
			insertErr := db.BatchInsertBinanceKline1M(ctx, klineModels)
			if insertErr != nil {
				logger.Debug("BatchInsertBinanceKline1M failed: %s", insertErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			logger.Debug("BatchInsertBinanceKline1M %s 成功", fillData.Symbol)
			if len(klineModels) < 1000 {
				err2 := db.UpdateBinanceDataFill(ctx, fillData.Symbol, fillData.EventStartTime, fillData.EventStartTime, true)
				if err2 != nil {
					logger.Debug("UpdateBinanceDataFill failed:%s", err2)
					continue
				}
				logger.Debug("UpdateBinanceDataFill %s 成功", fillData.Symbol)
				time.Sleep(2000 * time.Millisecond)
			} else {
				err2 := db.UpdateBinanceDataFill(ctx, fillData.Symbol, fillData.EventStartTime, klineModels[len(klineModels)-1].StartTime, false)
				if err2 != nil {
					logger.Debug("UpdateBinanceDataFill failed:%s", err2)
					continue
				}
				time.Sleep(2000 * time.Millisecond)
			}
			continue
		}
		break
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
			Interval:       "1m",
			FirstTradeID:   0, // REST 不提供
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

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
