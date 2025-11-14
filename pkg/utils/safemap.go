package utils

import (
	"sync"
	"time"
)

// SafeMap 是并发安全的泛型 map
type SafeMap[K comparable, V any] struct {
	mu       sync.RWMutex
	m        map[K]V
	ttl      time.Duration
	expire   map[K]time.Time
	stopChan chan struct{}
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{m: make(map[K]V)}
}

// NewSafeMapWithTTL 创建带统一过期时间的 SafeMap，所有 key 都有相同 ttl，自动清理过期 key
func NewSafeMapWithTTL[K comparable, V any](ttl time.Duration) *SafeMap[K, V] {
	sm := &SafeMap[K, V]{
		m:        make(map[K]V),
		ttl:      ttl,
		expire:   make(map[K]time.Time),
		stopChan: make(chan struct{}),
	}
	go sm.expireLoop()
	return sm
}

// expireLoop 定时清理过期 key
func (s *SafeMap[K, V]) expireLoop() {
	ticker := time.NewTicker(s.ttl / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for k, exp := range s.expire {
				if now.After(exp) {
					delete(s.m, k)
					delete(s.expire, k)
				}
			}
			s.mu.Unlock()
		case <-s.stopChan:
			return
		}
	}
}

// Set 写入或覆盖
func (s *SafeMap[K, V]) Set(key K, val V) {
	s.mu.Lock()
	s.m[key] = val
	if s.ttl > 0 {
		s.expire[key] = time.Now().Add(s.ttl)
	}
	s.mu.Unlock()
}

// Get 读取, 若存在则更新过期时间
func (s *SafeMap[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	val, ok := s.m[key]
	if s.ttl > 0 {
		s.expire[key] = time.Now().Add(s.ttl)
	}
	s.mu.RUnlock()
	return val, ok
}

// Delete 删除键
func (s *SafeMap[K, V]) Delete(key K) {
	s.mu.Lock()
	delete(s.m, key)
	if s.ttl > 0 {
		delete(s.expire, key)
	}
	s.mu.Unlock()
}

// Len 返回长度
func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	l := len(s.m)
	s.mu.RUnlock()
	return l
}

// Keys 返回键的切片（副本）
func (s *SafeMap[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回值的切片（副本）
func (s *SafeMap[K, V]) Values() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vals := make([]V, 0, len(s.m))
	for _, v := range s.m {
		vals = append(vals, v)
	}
	return vals
}

// Range 遍历 map，若 f 返回 false 则停止遍历
func (s *SafeMap[K, V]) Range(f func(K, V) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.m {
		if !f(k, v) {
			return
		}
	}
}

// LoadOrStore 若键存在则返回现有值和 true；否则存入新值并返回新值和 false
func (s *SafeMap[K, V]) LoadOrStore(key K, val V) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.m[key]; ok {
		return existing, true
	}
	s.m[key] = val
	return val, false
}

// CompareAndSwap 简单的比较并交换，成功返回 true
func (s *SafeMap[K, V]) CompareAndSwap(key K, old V, new V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
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
func (s *SafeMap[K, V]) Clear() {
	s.mu.Lock()
	s.m = make(map[K]V)
	if s.ttl > 0 {
		s.expire = make(map[K]time.Time)
	}
	s.mu.Unlock()
}

// Stop 清理协程（仅对带TTL的SafeMap有效）
func (s *SafeMap[K, V]) Stop() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}
