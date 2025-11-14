package utils

// SliceContains 检查切片是否包含指定元素
func SliceContains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
