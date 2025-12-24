package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using an in-memory map
type MemoryCache struct {
	data            map[string]*cacheItem
	mu              sync.RWMutex
	keyPrefix       string
	maxSize         int
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache(config *Config) *MemoryCache {
	mc := &MemoryCache{
		data:            make(map[string]*cacheItem),
		keyPrefix:       config.KeyPrefix,
		maxSize:         config.MemoryMaxSize,
		cleanupInterval: config.MemoryCleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.cleanupExpired()

	return mc
}

// prefixKey adds the configured prefix to a key
func (c *MemoryCache) prefixKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return c.keyPrefix + ":" + key
}

// cleanupExpired periodically removes expired items
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, item := range c.data {
				if !item.expiration.IsZero() && now.After(item.expiration) {
					delete(c.data, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCleanup:
			return
		}
	}
}

// Get retrieves a value from the memory cache
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[c.prefixKey(key)]
	if !exists {
		return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
	}

	// Check if expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
	}

	// Return a copy to prevent data races
	result := make([]byte, len(item.value))
	copy(result, item.value)
	return result, nil
}

// Set stores a value in the memory cache
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: if at max size, remove oldest expired items first
	if c.maxSize > 0 && len(c.data) >= c.maxSize {
		now := time.Now()
		for k, item := range c.data {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(c.data, k)
				if len(c.data) < c.maxSize {
					break
				}
			}
		}
		// If still at max, remove a random item
		if len(c.data) >= c.maxSize {
			for k := range c.data {
				delete(c.data, k)
				break
			}
		}
	}

	// Make a copy of the value
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	c.data[c.prefixKey(key)] = &cacheItem{
		value:      valueCopy,
		expiration: exp,
	}

	return nil
}

// Delete removes a value from the memory cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, c.prefixKey(key))
	return nil
}

// Exists checks if a key exists in the memory cache
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[c.prefixKey(key)]
	if !exists {
		return false, nil
	}

	// Check if expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return false, nil
	}

	return true, nil
}

// Clear removes all keys matching a pattern
func (c *MemoryCache) Clear(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefixedPattern := c.prefixKey(pattern)
	// Simple pattern matching: only support * at the end
	if strings.HasSuffix(prefixedPattern, "*") {
		prefix := strings.TrimSuffix(prefixedPattern, "*")
		for key := range c.data {
			if strings.HasPrefix(key, prefix) {
				delete(c.data, key)
			}
		}
	} else {
		delete(c.data, prefixedPattern)
	}

	return nil
}

// Close stops the cleanup goroutine
func (c *MemoryCache) Close() error {
	close(c.stopCleanup)
	return nil
}

// Ping always returns nil for memory cache
func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

// IsAvailable always returns true for memory cache
func (c *MemoryCache) IsAvailable() bool {
	return true
}

// NoOpCache is a cache that does nothing (used when caching is disabled)
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
}

func (c *NoOpCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *NoOpCache) Clear(ctx context.Context, pattern string) error {
	return nil
}

func (c *NoOpCache) Close() error {
	return nil
}

func (c *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

func (c *NoOpCache) IsAvailable() bool {
	return false
}
