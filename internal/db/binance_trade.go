package db

import (
	"Centralized-Data-Collector/internal/model"
	"context"
)

// func InsertBinanceAggTrade(ctx context.Context, trades *model.OKXTrade) error {
// 	return GormDB.WithContext(ctx).Create(trades).Error
// }

func BatchInsertBinanceAggTrades(ctx context.Context, trades []*model.BinanceAggTrade) error {
	for _, t := range trades {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).Create(trades).Error
}

// // 批量插入
// func BatchInsertOKXCandle1s(ctx context.Context, candles []*model.OKXCandle) error {
// 	return GormDB.WithContext(ctx).Table("candle1s").Create(candles).Error
// }
