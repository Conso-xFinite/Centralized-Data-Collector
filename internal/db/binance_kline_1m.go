package db

import (
	"Centralized-Data-Collector/internal/model"
	"context"
)

// func InsertBinanceAggTrade(ctx context.Context, trades *model.OKXTrade) error {
// 	return GormDB.WithContext(ctx).Create(trades).Error
// }

func BatchInsertBinanceKline1M(ctx context.Context, klines []*model.Kline1m) error {
	for _, t := range klines {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).Create(klines).Error
}

// // 批量插入
// func BatchInsertOKXCandle1s(ctx context.Context, candles []*model.OKXCandle) error {
// 	return GormDB.WithContext(ctx).Table("candle1s").Create(candles).Error
// }
