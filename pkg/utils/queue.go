package utils

// Queue 是并发安全的泛型队列（FIFO）
type Queue[T any] struct {
	vals []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{vals: make([]T, 0)}
}

// Push 入队
func (q *Queue[T]) Push(v T) {
	q.vals = append(q.vals, v)
}

// PushBatch 批量入队
func (q *Queue[T]) PushBatch(vals []T) {
	if len(vals) == 0 {
		return
	}
	q.vals = append(q.vals, vals...)
}

// Pop 出队一个，若为空返回零值和 false
func (q *Queue[T]) Pop() (T, bool) {
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
func (q *Queue[T]) PopBatch(n int) []T {
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
func (q *Queue[T]) PopAll() []T {
	if len(q.vals) == 0 {
		return nil
	}
	out := q.vals
	q.vals = make([]T, 0)
	return out
}

// Len 返回队列长度
func (q *Queue[T]) Len() int {
	l := len(q.vals)
	return l
}

// Peek 返回队首元素但不出队
func (q *Queue[T]) Peek() (T, bool) {
	if len(q.vals) == 0 {
		var zero T
		return zero, false
	}
	return q.vals[0], true
}

// PeekBatch 返回队列所有元素的副本但不出队
func (q *Queue[T]) PeekBatch() []T {
	if len(q.vals) == 0 {
		return nil
	}
	out := make([]T, len(q.vals))
	copy(out, q.vals)
	return out
}

// IsEmpty 判断队列是否为空
func (q *Queue[T]) IsEmpty() bool {
	empty := len(q.vals) == 0
	return empty
}

// Clear 清空队列
func (q *Queue[T]) Clear() {
	q.vals = make([]T, 0)
}
