package db

import (
	"Centralized-Data-Collector/internal/model"
	"Centralized-Data-Collector/pkg/logger"
	"context"
	"strings"
)

func BatchInsertBinanceKline1M(ctx context.Context, klines []*model.Kline1m) error {
	for _, t := range klines {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).Create(klines).Error
}

func BatchDeleteAndInsertKline1M(ctx context.Context, models []*model.Kline1m) error {
	return GormDB.WithContext(ctx).Transaction(func(tx *DB) error {
		// 1) 删除 models中k线的startTime已存在的行，不做其他过滤。直接全部删除
		dbQuery := GormDB.WithContext(ctx).Model(&model.Kline1m{})
		var conditions []string
		var values []interface{}

		for _, m := range models {
			conditions = append(conditions, "(pair = ? AND start_time = ?)")
			values = append(values, m.Pair, m.StartTime)
		}

		condStr := strings.Join(conditions, " OR ")
		err := dbQuery.Where(condStr, values...).Delete(&model.Kline1m{}).Error
		if err != nil {
			logger.Debug("批量删除失败:", err)
		}
		logger.Debug("批量删除成功")
		// 2) 插入 dedupedCandles（如果有）
		batchInsertErr := BatchInsertBinanceKline1M(ctx, models)
		if batchInsertErr != nil {
			return err
		}
		logger.Debug("批量插入成功")
		return nil
	})
}

func GetLatestKLine1mInfo(ctx context.Context, pair string) (*model.Kline1m, error) {
	var kline1m *model.Kline1m
	err := GormDB.
		WithContext(ctx).
		Model(&model.Kline1m{}).
		Where("pair = ?", pair).
		Order("event_time DESC").
		First(&kline1m).Error
	if err != nil {
		if isErrRecordNotFound(err) {
			return nil, nil // Token not found
		}
		return nil, err
	}
	return kline1m, nil
}

func DeleteBinanceKline(ctx context.Context, symbol string, startTime int64, endTime int64) error {
	return GormDB.
		WithContext(ctx).
		Where("pair = ?", symbol).
		Where("start_time = ?", startTime).
		Where("end_time = ?", endTime).
		Delete(&model.Kline1m{}).Error

}
