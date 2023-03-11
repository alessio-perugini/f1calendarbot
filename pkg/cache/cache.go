package cache

import (
	"sync"
	"time"
)

type Interface[K comparable, T any] interface {
	Get(k K) T
	Set(k K, v T, ttl time.Duration)
	Remove(key K)
	Close()
}

type TTLCache[K comparable, T any] struct {
	close chan struct{}
	mu    sync.RWMutex
	m     map[K]item[T]
}

func NewTTLCache[K comparable, T any](cleanupInterval time.Duration) *TTLCache[K, T] {
	c := &TTLCache[K, T]{close: make(chan struct{}), m: make(map[K]item[T], 16)}

	// cleanup func
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.close:
				return
			}
		}
	}()

	return c
}

type item[T any] struct {
	value      T
	expiration int64
}

func (c *TTLCache[K, T]) Get(k K) T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[k].value
}

func (c *TTLCache[K, T]) Set(k K, v T, ttl time.Duration) {
	c.mu.Lock()
	c.m[k] = item[T]{
		value:      v,
		expiration: time.Now().UnixNano() + int64(ttl),
	}
	c.mu.Unlock()
}

func (c *TTLCache[K, T]) Remove(key K) {
	c.mu.Lock()
	delete(c.m, key)
	c.mu.Unlock()
}

func (c *TTLCache[K, T]) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, v := range c.m {
		if v.expiration < now {
			delete(c.m, k)
		}
	}
	c.mu.Unlock()
}

func (c *TTLCache[K, T]) Close() {
	c.close <- struct{}{}
	c.m = make(map[K]item[T])
}
