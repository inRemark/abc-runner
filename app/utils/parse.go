package utils

import (
	"strconv"
)

// ParseInt 解析整数字符串
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// ParseInt64 解析64位整数字符串
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseFloat64 解析浮点数字符串
func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ParseBool 解析布尔值字符串
func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
