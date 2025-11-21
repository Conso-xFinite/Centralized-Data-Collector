package db

import (
	"Centralized-Data-Collector/internal/model"
	"context"
)

// func InsertBinanceAggTrade(ctx context.Context, trades *model.OKXTrade) error {
// 	return GormDB.WithContext(ctx).Create(trades).Error
// }

func BatchInsertBinanceDataFill(ctx context.Context, fillDatas []*model.MarketDataFill) error {
	for _, t := range fillDatas {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).Create(fillDatas).Error
}

func DeleteBinanceDataFill(ctx context.Context, fillDatas *model.MarketDataFill) error {
	return GormDB.WithContext(ctx).Delete(fillDatas).Error
}

func UpdateBinanceDataFill(ctx context.Context, event string, symbol string, startTime int64, startTimeUpdateTo int64, isClosed bool) error {
	return GormDB.WithContext(ctx).
		Model(&model.MarketDataFill{}).
		Where("event = ?", event).
		Where("symbol = ?", symbol).
		Where("event_start_time = ?", startTime).
		Updates(map[string]interface{}{
			"event_start_time": startTimeUpdateTo,
			"is_closed":        isClosed,
		}).Error

}

func GetBinanceDataFill(ctx context.Context) (*model.MarketDataFill, error) {
	var fillData *model.MarketDataFill
	err := GormDB.
		WithContext(ctx).
		Model(&model.MarketDataFill{}).
		Where("is_closed = ?", false).
		Order("created_at").
		First(&fillData).Error
	if err != nil {
		if isErrRecordNotFound(err) {
			return nil, nil // Token not found
		}
		return nil, err
	}
	return fillData, nil
}
