package utils

import (
	"Centralized-Data-Collector/pkg/logger"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"sync/atomic"
)

// 生成 HMAC-SHA256 哈希并将前4个字节转换为 uint32,最后与给定的数字进行求模
func HMACSHA256FromStringToUint(s string, i *atomic.Int32) uint32 {
	hash := sha256.Sum256([]byte(s))
	b := hash[0:4]
	num := binary.BigEndian.Uint32(b[:])
	poolLength := uint32(i.Load())
	mod := num % uint32(poolLength)
	logger.Debug("num s% s% s%", num, poolLength, mod)
	return mod
}

// 生成 HMAC-SHA256 哈希
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
