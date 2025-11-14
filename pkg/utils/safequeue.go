package utils

import "sync"

// SafeQueue 是并发安全的泛型队列（FIFO）
type SafeQueue[T any] struct {
	mu   sync.Mutex
	vals []T
}

func NewSafeQueue[T any]() *SafeQueue[T] {
	return &SafeQueue[T]{vals: make([]T, 0)}
}

// Push 入队
func (q *SafeQueue[T]) Push(v T) {
	q.mu.Lock()
	q.vals = append(q.vals, v)
	q.mu.Unlock()
}

// PushBatch 批量入队
func (q *SafeQueue[T]) PushBatch(vals []T) {
	if len(vals) == 0 {
		return
	}
	q.mu.Lock()
	q.vals = append(q.vals, vals...)
	q.mu.Unlock()
}

// Pop 出队一个，若为空返回零值和 false
func (q *SafeQueue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.vals) == 0 {
		var zero T
		return zero, false
	}
	v := q.vals[0]
	// 避免内存泄漏，按需清除引用
	q.vals[0] = *new(T)
	q.vals = q.vals[1:]
	return v, true
}

// PopBatch 出队最多 n 个元素
func (q *SafeQueue[T]) PopBatch(n int) []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	if n <= 0 || len(q.vals) == 0 {
		return nil
	}
	if n > len(q.vals) {
		n = len(q.vals)
	}
	out := make([]T, n)
	copy(out, q.vals[:n])
	// 清理已取出的引用
	for i := 0; i < n; i++ {
		q.vals[i] = *new(T)
	}
	q.vals = q.vals[n:]
	return out
}

// PopAll 取出全部并清空队列
func (q *SafeQueue[T]) PopAll() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.vals) == 0 {
		return nil
	}
	out := q.vals
	q.vals = make([]T, 0)
	return out
}

// Len 返回队列长度
func (q *SafeQueue[T]) Len() int {
	q.mu.Lock()
	l := len(q.vals)
	q.mu.Unlock()
	return l
}

// Peek 返回队首元素但不出队
func (q *SafeQueue[T]) Peek() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.vals) == 0 {
		var zero T
		return zero, false
	}
	return q.vals[0], true
}

// PeekBatch 返回队列所有元素的副本但不出队
func (q *SafeQueue[T]) PeekBatch() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.vals) == 0 {
		return nil
	}
	out := make([]T, len(q.vals))
	copy(out, q.vals)
	return out
}

// IsEmpty 判断队列是否为空
func (q *SafeQueue[T]) IsEmpty() bool {
	q.mu.Lock()
	empty := len(q.vals) == 0
	q.mu.Unlock()
	return empty
}

// Clear 清空队列
func (q *SafeQueue[T]) Clear() {
	q.mu.Lock()
	q.vals = make([]T, 0)
	q.mu.Unlock()
}
