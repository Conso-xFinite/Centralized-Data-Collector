package collector

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/internal/db"
	"Centralized-Data-Collector/internal/model"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"strconv"
	"time"
)

type BinanceStorer struct {
	source string
	// priceQueue         *utils.SafeQueue[*binance_define.WSPriceData]
	// tradeQueue         *utils.SafeQueue[*binance_define.WSTradeData]
	tokenContractIdMap *utils.SafeMap[string, string] // key: source:chainIndex:tokenContractAddress, value: tokenContractId
	tokenIdMap         *utils.SafeMap[string, string] // key: source:chainIndex:tokenAddress, value: tokenId
	poolIdMap          *utils.SafeMap[string, string] // key: source:chainIndex:dexName:logo_url, value: poolId
}

func NewBinanceStorer() *BinanceStorer {
	return &BinanceStorer{
		source: "binance",
		// priceQueue:         utils.NewSafeQueue[*binance_define.WSPriceData](),
		// tradeQueue:         utils.NewSafeQueue[*binance_define.WSTradeData](),
		tokenContractIdMap: utils.NewSafeMapWithTTL[string, string](10 * time.Minute),
		tokenIdMap:         utils.NewSafeMapWithTTL[string, string](10 * time.Minute),
		poolIdMap:          utils.NewSafeMapWithTTL[string, string](10 * time.Minute),
	}
}

func (c *BinanceStorer) parsePushedMsgToTradeModel(ctx context.Context, pushedMsg *binance_define.WSSinglePushMsg) *model.BinanceAggTrade {
	// 	// 解析 pushedMsg 为 TradeModel
	data := pushedMsg.Data.(binance_define.BinanceAggTrade)
	tradeModel := &model.BinanceAggTrade{
		Event:     data.Event,
		EventTime: data.EventTime,
		Symbol:    data.Symbol,
		AggID:     data.AggID,
		// Price:      utils.StringToDecimal(data.Price),
		// Quantity:   utils.StringToDecimal(data.Quantity),
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

	// candle1sModels := []*model.BinanceCandle1m{}

	// 解析 pushedMsgs 为 TradeModel 和 PriceModel
	for i := 0; i < len(pushedMsgs); i++ {
		msg := pushedMsgs[i]
		// logger.Error("Failed to batch insert OKX trades: err = %v", msg)
		switch msg.EventType {
		case "aggTrade":
			tradeModels = append(tradeModels, c.parsePushedMsgToTradeModel(ctx, msg))
		case "kline":
			klineModels = append(klineModels, c.parsePushedMsgToCandleModel(ctx, msg))
		case "24hrMiniTicker":
			tickerModels = append(tickerModels, c.parsePushedMsgToTickereModel(ctx, msg))
		}
		// TODO 后续需要检查一个数据，aggtrade中的数据中的交易id和kline中的交易id是否是一致的（逐条推送和聚合推送的在两个数据中id）

		if len(tradeModels) > 0 {
			for {
				err := db.BatchInsertBinanceAggTrades(ctx, tradeModels)
				if err == nil {
					logger.Info("Stored %d trade records from Binance", len(tradeModels))
					break
				}
				logger.Error("Failed to batch insert Binance trades: err = %v", err)
				time.Sleep(100 * time.Millisecond)
			}
		}

		if len(klineModels) > 0 {
			for {
				err := db.BatchInsertBinanceKline1M(ctx, klineModels)
				if err == nil {
					// logger.Info("Stored %d kline records from Binance", len(klineModels))
					break
				}
				logger.Error("Failed to batch insert Binance kline: err = %v", err)
				time.Sleep(100 * time.Millisecond)
			}
		}

		if len(klineModels) > 0 {
			for {
				err := db.BatchInsertBinanceKline1M(ctx, klineModels)
				if err == nil {
					// logger.Info("Stored %d kline records from Binance", len(klineModels))
					break
				}
				logger.Error("Failed to batch insert Binance kline: err = %v", err)
				time.Sleep(100 * time.Millisecond)
			}
		}
		if len(tickerModels) > 0 {
			for {
				err := db.BatchInsertBinanceTicker(ctx, tickerModels)
				if err == nil {
					logger.Info("Stored %d ticker records from Binance", len(tickerModels))
					break
				}
				logger.Error("Failed to batch insert Binance ticker: err = %v", err)
				time.Sleep(100 * time.Millisecond)
			}

		}
	}
}
