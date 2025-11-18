package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"sync/atomic"
)

// 生成 HMAC-SHA256 哈希并将前4个字节转换为 uint32,最后与给定的数字进行求模
func HMACSHA256FromStringToUint(s string, i *atomic.Int32) uint32 {
	//几个市值最高的代币分隔开
	hash := sha256.Sum256([]byte(s))
	b := hash[0:4]
	num := binary.BigEndian.Uint32(b[:])
	poolLength := uint32(i.Load())
	mod := num % uint32(poolLength)
	if strings.Contains(s, "btc") {
		mod = 0
	} else if strings.Contains(s, "eth") {
		mod = 1
	} else if strings.Contains(s, "bnb") {
		mod = 2
	} else if strings.Contains(s, "sol") {
		mod = 3
	}
	return mod
}

// 生成 HMAC-SHA256 哈希
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
