package utils

import (
	"crypto/hmac"
	"crypto/sha256"
)

// func HashPassword(password string) (string, error) {
// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	return string(bytes), err
// }

// func CheckPasswordHash(password, hash string) bool {
// 	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
// }

// 生成 HMAC-SHA256 哈希
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
