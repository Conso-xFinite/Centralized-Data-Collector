package db

// import (
// 	"Centralized-Data-Collector/internal/model"
// 	"context"
// 	"fmt"
// 	"strings"
// )

// func InsertOKXCandle1s(ctx context.Context, candle *model.OKXCandle) error {
// 	return GormDB.WithContext(ctx).Table("candle1s").Create(candle).Error
// }

// // 批量插入
// func BatchInsertOKXCandle1s(ctx context.Context, candles []*model.OKXCandle) error {
// 	return GormDB.WithContext(ctx).Table("candle1s").Create(candles).Error
// }

// func BatchDeleteAndInsertCandle1s(ctx context.Context, sameCandles []*model.OKXCandle, dedupedCandles []*model.OKXCandle) error {
// 	if len(dedupedCandles) == 0 && len(sameCandles) == 0 {
// 		return nil
// 	}

// 	return GormDB.WithContext(ctx).Transaction(func(tx *DB) error {
// 		// 1) 删除 sameCandles（如果有）
// 		if len(sameCandles) > 0 {
// 			var valuePlaceholders []string
// 			var params []interface{}
// 			for _, c := range sameCandles {
// 				valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?)")
// 				params = append(params, c.Source, c.ChainIndex, c.TokenContractID, c.Timestamp)
// 			}

// 			delQuery := fmt.Sprintf(
// 				"DELETE FROM %s WHERE (source, chain_index, token_contract_id, timestamp) IN (%s)",
// 				"candle1s",
// 				strings.Join(valuePlaceholders, ","),
// 			)

// 			if res := tx.Exec(delQuery, params...); res.Error != nil {
// 				return res.Error
// 			}
// 		}

// 		// 2) 插入 dedupedCandles（如果有）
// 		if len(dedupedCandles) > 0 {
// 			err := BatchInsertOKXCandle1s(ctx, dedupedCandles)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})
// }

// func GetOKXCandle1sByTimeRange(ctx context.Context, startTimestamp, endTimestamp int64) ([]*model.OKXCandle, error) {
// 	var candles []*model.OKXCandle
// 	err := GormDB.WithContext(ctx).
// 		Where("timestamp BETWEEN ? AND ?", startTimestamp, endTimestamp).
// 		Find(&candles).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return candles, nil
// }

// func MatchSameCandles(
// 	result []*model.OKXCandle,
// 	candles []*model.OKXCandle,
// ) (same []*model.OKXCandle) {
// 	candleMap := make(map[string]struct{}, len(candles))

// 	for _, c := range candles {
// 		key := fmt.Sprintf("%s_%d_%s_%d", c.Source, c.ChainIndex, c.TokenContractID, c.Timestamp)
// 		candleMap[key] = struct{}{}
// 	}

// 	for _, r := range result {
// 		key := fmt.Sprintf("%s_%d_%s_%d", r.Source, r.ChainIndex, r.TokenContractID, r.Timestamp)
// 		if _, exists := candleMap[key]; exists {
// 			same = append(same, r)
// 		}
// 	}

// 	return same
// }
