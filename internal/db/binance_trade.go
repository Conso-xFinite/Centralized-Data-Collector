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

func GetLatestAggTradeInfo(ctx context.Context, symbol string) (*model.BinanceAggTrade, error) {
	var addTradeModel *model.BinanceAggTrade
	err := GormDB.
		WithContext(ctx).
		Model(&model.BinanceAggTrade{}).
		Where("symbol = ?", symbol).
		Order("event_time DESC").
		First(&addTradeModel).Error
	if err != nil {
		if isErrRecordNotFound(err) {
			return nil, nil // Token not found
		}
		return nil, err
	}
	return addTradeModel, nil
}

// // 批量插入
// func BatchInsertOKXCandle1s(ctx context.Context, candles []*model.OKXCandle) error {
// 	return GormDB.WithContext(ctx).Table("candle1s").Create(candles).Error
// }
