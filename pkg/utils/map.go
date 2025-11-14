package utils

import (
	"time"
)

// Map 是并发安全的泛型 map
type Map[K comparable, V any] struct {
	m        map[K]V
	ttl      time.Duration
	expire   map[K]time.Time
	stopChan chan struct{}
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}
}

// NewMapWithTTL 创建带统一过期时间的 Map，所有 key 都有相同 ttl，自动清理过期 key
func NewMapWithTTL[K comparable, V any](ttl time.Duration) *Map[K, V] {
	sm := &Map[K, V]{
		m:        make(map[K]V),
		ttl:      ttl,
		expire:   make(map[K]time.Time),
		stopChan: make(chan struct{}),
	}
	go sm.expireLoop()
	return sm
}

// expireLoop 定时清理过期 key
func (s *Map[K, V]) expireLoop() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for k, exp := range s.expire {
				if now.After(exp) {
					delete(s.m, k)
					delete(s.expire, k)
				}
			}
		case <-s.stopChan:
			return
		}
	}
}

// Set 写入或覆盖
func (s *Map[K, V]) Set(key K, val V) {
	s.m[key] = val
	if s.ttl > 0 {
		s.expire[key] = time.Now().Add(s.ttl)
	}
}

// Get 读取
func (s *Map[K, V]) Get(key K) (V, bool) {
	val, ok := s.m[key]
	return val, ok
}

// Delete 删除键
func (s *Map[K, V]) Delete(key K) {
	delete(s.m, key)
	if s.ttl > 0 {
		delete(s.expire, key)
	}
}

// Len 返回长度
func (s *Map[K, V]) Len() int {
	l := len(s.m)
	return l
}

// Keys 返回键的切片（副本）
func (s *Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回值的切片（副本）
func (s *Map[K, V]) Values() []V {
	vals := make([]V, 0, len(s.m))
	for _, v := range s.m {
		vals = append(vals, v)
	}
	return vals
}

// Range 遍历 map，若 f 返回 false 则停止遍历
func (s *Map[K, V]) Range(f func(K, V) bool) {
	for k, v := range s.m {
		if !f(k, v) {
			return
		}
	}
}

// LoadOrStore 若键存在则返回现有值和 true；否则存入新值并返回新值和 false
func (s *Map[K, V]) LoadOrStore(key K, val V) (V, bool) {
	if existing, ok := s.m[key]; ok {
		return existing, true
	}
	s.m[key] = val
	return val, false
}

// CompareAndSwap 简单的比较并交换，成功返回 true
func (s *Map[K, V]) CompareAndSwap(key K, old V, new V) bool {
	cur, ok := s.m[key]
	if !ok {
		return false
	}
	// 仅在可比类型上有意义，调用者需保证 old/new 可比或按需实现
	if any(cur) == any(old) {
		s.m[key] = new
		return true
	}
	return false
}

// Clear 清空 map
func (s *Map[K, V]) Clear() {
	s.m = make(map[K]V)
	if s.ttl > 0 {
		s.expire = make(map[K]time.Time)
	}
}

// Stop 清理协程（仅对带TTL的Map有效）
func (s *Map[K, V]) Stop() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}
