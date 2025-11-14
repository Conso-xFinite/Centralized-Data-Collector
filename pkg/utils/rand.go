package utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

var ErrInvalidLength = errors.New("invalid length")

func RandomBytes(length int) ([]byte, error) {
	// 这里要判断length是否合法
	if length <= 0 {
		return nil, ErrInvalidLength
	}
	buf := make([]byte, length)
	left := length
	for left > 0 {
		n, err := rand.Read(buf[length-left:])
		if err != nil {
			return nil, err
		}
		left -= n
	}
	return buf, nil
}

func RandomUint32(max uint32) (uint32, error) {
	buf, err := RandomBytes(4)
	if err != nil {
		return 0, err
	}
	num := binary.LittleEndian.Uint32(buf)
	return num % max, nil
}

func RandomUint64(max uint64) (uint64, error) {
	buf, err := RandomBytes(8)
	if err != nil {
		return 0, err
	}
	num := binary.LittleEndian.Uint64(buf)
	return num % max, nil
}
