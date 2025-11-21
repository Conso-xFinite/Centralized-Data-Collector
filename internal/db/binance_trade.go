package db

import (
	"Centralized-Data-Collector/internal/model"
	"Centralized-Data-Collector/pkg/logger"
	"context"
	"time"

	"gorm.io/gorm/clause"
)

// func InsertBinanceAggTrade(ctx context.Context, trades *model.OKXTrade) error {
// 	return GormDB.WithContext(ctx).Create(trades).Error
// }

func BatchInsertBinanceAggTrades(ctx context.Context, trades []*model.BinanceAggTrade) error {
	for _, t := range trades {
		t.ID = "" // 确保为空
	}
	return GormDB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "agg_id"}},
			DoNothing: true,
		}).
		Create(trades).Error
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

func BatchInsertAggTradeAndFillData(ctx context.Context, tradeModels []*model.BinanceAggTrade, fillData *model.MarketDataFill) error {
	return GormDB.WithContext(ctx).Transaction(func(tx *DB) error {

		insertErr := BatchInsertBinanceAggTrades(ctx, tradeModels)
		if insertErr != nil {
			logger.Debug("BatchInsertBinanceKline1M failed: %s", insertErr)
			time.Sleep(200 * time.Millisecond)
			return insertErr
		}

		if len(tradeModels) < 1000 {
			updateErr := UpdateBinanceDataFill(ctx, fillData.Event, fillData.Symbol, fillData.EventStartTime, fillData.EventStartTime, true)
			if updateErr != nil {
				logger.Debug("UpdateBinanceDataFill failed:%s", updateErr)
				return updateErr
			}
		} else {
			//如果返回的数量少于limit 就说明还有，需要再次发起
			updateErr2 := UpdateBinanceDataFill(ctx, fillData.Event, fillData.Symbol, fillData.EventStartTime, tradeModels[len(tradeModels)-1].EventTime, false)
			if updateErr2 != nil {
				logger.Debug("UpdateBinanceDataFill failed:%s", updateErr2)
				return updateErr2
			}
		}
		logger.Debug("AggTrades数据补充 %d 条数据成功， 第一条的id: %d 最后一条id: %d", len(tradeModels), tradeModels[0].EventTime, tradeModels[len(tradeModels)-1].EventTime)
		return nil
	})
}

func DeleteSingleAggTrade(ctx context.Context, symbol string, eventTime int64) error {
	return GormDB.
		WithContext(ctx).
		Where("symbol = ?", symbol).
		Where("event_time = ?", eventTime).
		Delete(&model.BinanceAggTrade{}).Error

}
