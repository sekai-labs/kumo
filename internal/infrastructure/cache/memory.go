package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sekai-labs/kumo/internal/core/ports"
)

type cacheItem struct {
	value     any
	expiresAt time.Time
}

func (item cacheItem) isExpired(now time.Time) bool {
	if item.expiresAt.IsZero() {
		return false
	}
	return now.After(item.expiresAt)
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

var _ ports.Cache = (*MemoryCache)(nil)

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheItem),
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string, target any) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return false, nil
	}

	now := time.Now()
	if item.isExpired(now) {

		c.mu.Lock()

		if curItem, exists := c.items[key]; exists && curItem.isExpired(now) {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return false, nil
	}

	if target != nil {
		if err := assignTarget(item.value, target); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:     val,
		expiresAt: expiresAt,
	}
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

func (c *MemoryCache) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
	return nil
}

func (c *MemoryCache) PurgeExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0
	for k, item := range c.items {
		if item.isExpired(now) {
			delete(c.items, k)
			count++
		}
	}
	return count
}

func assignTarget(val any, target any) error {

	switch ptr := target.(type) {
	case *any:
		*ptr = val
		return nil
	}

	return assignReflect(val, target)
}
