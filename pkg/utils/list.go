package utils

// List 是并发安全的切片封装
type List[T any] struct {
	vals      []T
	equalFunc func(a, b T) bool
}

func NewList[T any]() *List[T] {
	return &List[T]{vals: make([]T, 0)}
}

func NewListWithEqualFunc[T any](equalFunc func(a, b T) bool) *List[T] {
	l := NewList[T]()
	l.equalFunc = equalFunc
	return l
}

// Append 追加元素
func (l *List[T]) Append(v T) {
	l.vals = append(l.vals, v)
}

// AppendBatch 批量追加
func (l *List[T]) AppendBatch(vals []T) {
	if len(vals) == 0 {
		return
	}
	l.vals = append(l.vals, vals...)
}

// Get 安全获取索引元素
func (l *List[T]) Get(i int) (T, bool) {
	if i < 0 || i >= len(l.vals) {
		var zero T
		return zero, false
	}
	return l.vals[i], true
}

// IndexOf 返回指定 value 在 List 中的位置，不存在则返回 -1。
// 此方法依赖于 equalFunc，使用前请确保已设置。
func (l *List[T]) IndexOf(value T) int {
	for i, v := range l.vals {
		if l.equalFunc(v, value) {
			return i
		}
	}
	return -1
}

// Len 返回长度
func (l *List[T]) Len() int {
	n := len(l.vals)
	return n
}

// All 返回值副本，防止外部修改内部切片
func (l *List[T]) All() []T {
	out := make([]T, len(l.vals))
	copy(out, l.vals)
	return out
}

// RemoveAt 删除指定索引元素（保持顺序）
func (l *List[T]) RemoveAt(i int) bool {
	if i < 0 || i >= len(l.vals) {
		return false
	}
	copy(l.vals[i:], l.vals[i+1:])
	l.vals[len(l.vals)-1] = *new(T)
	l.vals = l.vals[:len(l.vals)-1]
	return true
}

// Clear 清空列表
func (l *List[T]) Clear() {
	l.vals = make([]T, 0)
}

func (l *List[T]) RemoveRange(start, end int) {
	if start < 0 {
		start = 0
	}
	if end > len(l.vals) {
		end = len(l.vals)
	}
	if start >= end {
		return // 无需删除
	}

	// 删除区间[start, end)
	l.vals = append(l.vals[:start], l.vals[end:]...)
}
