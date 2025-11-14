package db

// import (
// 	"Centralized-Data-Collector/internal/model"
// 	"context"
// )

// func GetOKXTokenContractByAddress(ctx context.Context, source string, chainIndex int, tokenContractAddress string) (*model.OKXTokenContract, error) {
// 	var tokenContract model.OKXTokenContract
// 	err := GormDB.WithContext(ctx).
// 		Where("source = ? AND chain_index = ? AND address = ?", source, chainIndex, tokenContractAddress).
// 		First(&tokenContract).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return nil, nil // Token not found
// 		}
// 		return nil, err
// 	}
// 	return &tokenContract, nil
// }

// func GetOKXTokenContractIdByAddress(ctx context.Context, source string, chainIndex int, tokenContractAddress string) (string, error) {
// 	var tokenContractId string
// 	err := GormDB.
// 		WithContext(ctx).
// 		Model(&model.OKXTokenContract{}).
// 		Select("id").
// 		Where("source = ? AND chain_index = ? AND address = ?", source, chainIndex, tokenContractAddress).
// 		First(&tokenContractId).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return "", nil // Token not found
// 		}
// 		return "", err
// 	}
// 	return tokenContractId, nil
// }

// func GetOKXTokenContractByTokenContractId(ctx context.Context, tokenContractId string) (*model.OKXTokenContract, error) {
// 	var tokenContract model.OKXTokenContract
// 	err := GormDB.
// 		WithContext(ctx).
// 		Where("token_contract_id = ?", tokenContractId).
// 		First(&tokenContract).Error
// 	if err != nil {
// 		if isErrRecordNotFound(err) {
// 			return nil, nil // Token not found
// 		}
// 		return nil, err
// 	}
// 	return &tokenContract, nil
// }

// func InsertOKXTokenContract(ctx context.Context, source string, chainIndex int, tokenContractAddress string) (id string, err error) {
// 	tokenContract := &model.OKXTokenContract{
// 		Source:     source,
// 		ChainIndex: chainIndex,
// 		Address:    tokenContractAddress,
// 	}
// 	err = GormDB.WithContext(ctx).Create(tokenContract).Error
// 	if err != nil {
// 		return "", err
// 	}
// 	return tokenContract.ID, nil
// }
