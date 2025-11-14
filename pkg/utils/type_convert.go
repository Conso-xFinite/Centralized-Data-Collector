package utils

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

// StringToUint64 字符串转 uint64
func StringToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// StringToUint32 字符串转 uint32
func StringToUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	return uint32(u), err
}

// StringToInt64 字符串转 int64
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StringToInt 字符串转 int
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StringToFloat64 字符串转 float64
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// StringToDecimal 字符串转 decimal
func StringToDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

// InterfaceSliceToUint64Slice []interface{} 转 []uint64
func InterfaceSliceToUint64Slice(src []interface{}) ([]uint64, error) {
	result := make([]uint64, 0, len(src))
	for _, v := range src {
		switch val := v.(type) {
		case uint64:
			result = append(result, val)
		case uint32:
			result = append(result, uint64(val))
		case int:
			result = append(result, uint64(val))
		case int64:
			result = append(result, uint64(val))
		case string:
			u, err := StringToUint64(val)
			if err != nil {
				return nil, err
			}
			result = append(result, u)
		default:
			return nil, fmt.Errorf("unsupported type: %T", v)
		}
	}
	return result, nil
}

// InterfaceSliceToStringSlice []interface{} 转 []string
func InterfaceSliceToStringSlice(src []interface{}) ([]string, error) {
	result := make([]string, 0, len(src))
	for _, v := range src {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case uint64, int, int64, uint32:
			result = append(result, fmt.Sprintf("%v", val))
		default:
			return nil, fmt.Errorf("unsupported type: %T", v)
		}
	}
	return result, nil
}

// Uint64SliceToInterfaceSlice []uint64 转 []interface{}
func Uint64SliceToInterfaceSlice(src []uint64) []interface{} {
	result := make([]interface{}, len(src))
	for i, v := range src {
		result[i] = v
	}
	return result
}

// StringSliceToInterfaceSlice []string 转 []interface{}
func StringSliceToInterfaceSlice(src []string) []interface{} {
	result := make([]interface{}, len(src))
	for i, v := range src {
		result[i] = v
	}
	return result
}

// int 转 string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Uint32ToString uint32 转 string
func Uint32ToString(u uint32) string {
	return strconv.FormatUint(uint64(u), 10)
}

// Uint64ToString uint64 转 string
func Uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// Int64ToString int64 转 string
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToString float64 转 string
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// DecimalToString decimal 转 string
func DecimalToString(d decimal.Decimal) string {
	return d.String()
}
