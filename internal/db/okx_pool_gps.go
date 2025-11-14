package db

// import (
// 	"Centralized-Data-Collector/internal/model"
// 	"context"
// )

// func GetOKXPoolByAddress(ctx context.Context, source string, chainIndex int, dexName, logoUrl string) (*model.OKXPool, error) {
// 	var pool model.OKXPool
// 	err := GormDB.WithContext(ctx).
// 		Where("source = ? AND chain_index = ? AND dex_name = ? AND logo_url = ?", source, chainIndex, dexName, logoUrl).
// 		First(&pool).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return nil, nil // Pool not found
// 		}
// 		return nil, err
// 	}
// 	return &pool, nil
// }

// func GetOKXPoolIdByAddress(ctx context.Context, source string, chainIndex int, dexName, logoUrl string) (string, error) {
// 	var poolId string
// 	err := GormDB.
// 		WithContext(ctx).
// 		Model(&model.OKXPool{}).
// 		Select("id").
// 		Where("source = ? AND chain_index = ? AND dex_name = ? AND logo_url= ?", source, chainIndex, dexName, logoUrl).
// 		First(&poolId).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return "", nil // Pool not found
// 		}
// 		return "", err
// 	}
// 	return poolId, nil
// }

// func InsertOKXPool(ctx context.Context, source string, chainIndex int, dexName, logoUrl string) (id string, err error) {
// 	pool := &model.OKXPool{
// 		Source:     source,
// 		ChainIndex: chainIndex,
// 		DexName:    dexName,
// 		LogoUrl:    logoUrl,
// 	}
// 	err = GormDB.WithContext(ctx).Create(pool).Error
// 	if err != nil {
// 		return "", err
// 	}
// 	return pool.ID, nil
// }
