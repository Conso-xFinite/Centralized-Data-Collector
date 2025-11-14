package utils

import "time"

// 获取毫秒级时间戳
func GetCurrentTimestampMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// 获取秒级时间戳
func GetCurrentTimestampSec() int64 {
	return time.Now().UnixNano() / int64(time.Second)
}

// 获取纳秒级时间戳
func GetCurrentTimestampNano() int64 {
	return time.Now().UnixNano()
}
