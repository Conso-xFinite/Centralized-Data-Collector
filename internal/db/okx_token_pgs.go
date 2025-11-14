package db

// import (
// 	"Centralized-Data-Collector/internal/model"
// 	"context"
// )

// func GetOKXTokenByAddress(ctx context.Context, source string, chainIndex int, tokenAddress string) (*model.OKXToken, error) {
// 	var token model.OKXToken
// 	err := GormDB.WithContext(ctx).Where("source = ? AND chain_index = ? AND address = ? AND is_active = 1", source, chainIndex, tokenAddress).First(&token).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return nil, nil // Token not found
// 		}
// 		return nil, err
// 	}
// 	return &token, nil
// }

// func GetOKXTokenIdByAddress(ctx context.Context, source string, chainIndex int, tokenAddress string) (string, error) {
// 	var tokenId string
// 	err := GormDB.
// 		WithContext(ctx).
// 		Model(&model.OKXToken{}).
// 		Select("id").
// 		Where("source = ? AND chain_index = ? AND address = ? AND is_active = 1", source, chainIndex, tokenAddress).
// 		First(&tokenId).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return "", nil // Token not found
// 		}
// 		return "", err
// 	}
// 	return tokenId, nil
// }

// func InsertOKXToken(ctx context.Context, source string, chainIndex int, tokenAddress string, symbol string) (id string, err error) {
// 	token := &model.OKXToken{
// 		Source:     source,
// 		ChainIndex: chainIndex,
// 		Address:    tokenAddress,
// 		Symbol:     symbol,
// 		IsActive:   true,
// 	}
// 	err = GormDB.WithContext(ctx).Create(token).Error
// 	if err != nil {
// 		return "", err
// 	}
// 	return token.ID, nil
// }
