package cache

import (
	"sync"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/metrics"
)

type CacheEntry struct {
	Data      []byte
	Timestamp time.Time
}

type Cache struct {
	entries map[string]CacheEntry
	ttl     time.Duration
	mu      sync.RWMutex
}

func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}

	// Start cleanup routine
	go cache.cleanup()

	return cache
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.Timestamp) > c.ttl {
		metrics.CacheHits.WithLabelValues("miss_expired").Inc()
		return nil, false
	}

	metrics.CacheHits.WithLabelValues("hit").Inc()
	return entry.Data, true
}

func (c *Cache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.Timestamp) > c.ttl {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}
