package utils

import "sync"

// SafeList 是并发安全的切片封装
type SafeList[T any] struct {
	mu        sync.RWMutex
	vals      []T
	equalFunc func(a, b T) bool
}

func NewSafeList[T any]() *SafeList[T] {
	return &SafeList[T]{vals: make([]T, 0)}
}

func NewSafeListWithEqualFunc[T any](equalFunc func(a, b T) bool) *SafeList[T] {
	l := NewSafeList[T]()
	l.equalFunc = equalFunc
	return l
}

// Append 追加元素
func (l *SafeList[T]) Append(v T) {
	l.mu.Lock()
	l.vals = append(l.vals, v)
	l.mu.Unlock()
}

// AppendBatch 批量追加
func (l *SafeList[T]) AppendBatch(vals []T) {
	if len(vals) == 0 {
		return
	}
	l.mu.Lock()
	l.vals = append(l.vals, vals...)
	l.mu.Unlock()
}

// Get 安全获取索引元素
func (l *SafeList[T]) Get(i int) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if i < 0 || i >= len(l.vals) {
		var zero T
		return zero, false
	}
	return l.vals[i], true
}

// IndexOf 返回指定 value 在 SafeList 中的位置，不存在则返回 -1。
// 此方法依赖于 equalFunc，使用前请确保已设置。
func (l *SafeList[T]) IndexOf(value T) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i, v := range l.vals {
		if l.equalFunc(v, value) {
			return i
		}
	}
	return -1
}

// Len 返回长度
func (l *SafeList[T]) Len() int {
	l.mu.RLock()
	n := len(l.vals)
	l.mu.RUnlock()
	return n
}

// All 返回值副本，防止外部修改内部切片
func (l *SafeList[T]) All() []T {
	l.mu.RLock()
	out := make([]T, len(l.vals))
	copy(out, l.vals)
	l.mu.RUnlock()
	return out
}

// RemoveAt 删除指定索引元素（保持顺序）
func (l *SafeList[T]) RemoveAt(i int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if i < 0 || i >= len(l.vals) {
		return false
	}
	copy(l.vals[i:], l.vals[i+1:])
	l.vals[len(l.vals)-1] = *new(T)
	l.vals = l.vals[:len(l.vals)-1]
	return true
}

// Clear 清空列表
func (l *SafeList[T]) Clear() {
	l.mu.Lock()
	l.vals = make([]T, 0)
	l.mu.Unlock()
}
