package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"sync/atomic"
)

// 生成 HMAC-SHA256 哈希并将前4个字节转换为 uint32,最后与给定的数字进行求模
func HMACSHA256FromStringToUint(s string, i *atomic.Int32) int32 {
	hash := sha256.Sum256([]byte(s))
	b := hash[0:4]
	num := int32(binary.BigEndian.Uint32(b[:]))
	poolLength := i.Load()
	mod := num % (poolLength - 1)
	if mod < 0 {
		mod += poolLength
	}
	return mod
}

// 生成 HMAC-SHA256 哈希
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
