package db

import (
	"Centralized-Data-Collector/internal/model"
	"context"
)

func BatchInsertBinanceTicker(ctx context.Context, tickers []*model.TickerModel) error {
	for _, t := range tickers {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).Create(tickers).Error
}
