package cache

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// MetricsCache wraps a Cache implementation with hit/miss logging
type MetricsCache struct {
	cache  Cache
	hits   int64
	misses int64
}

// NewMetricsCache creates a new metrics-aware cache wrapper
func NewMetricsCache(cache Cache) *MetricsCache {
	return &MetricsCache{
		cache: cache,
	}
}

// Get retrieves a value from the cache and logs hit/miss
func (m *MetricsCache) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	data, err := m.cache.Get(ctx, key)
	latency := time.Since(start)

	if err != nil {
		atomic.AddInt64(&m.misses, 1)
		log.Debug().
			Str("key", key).
			Dur("latency", latency).
			Msg("cache miss")
		return nil, err
	}

	atomic.AddInt64(&m.hits, 1)
	log.Debug().
		Str("key", key).
		Dur("latency", latency).
		Int("size", len(data)).
		Msg("cache hit")
	return data, nil
}

// Set stores a value in the cache
func (m *MetricsCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	start := time.Now()
	err := m.cache.Set(ctx, key, value, expiration)
	latency := time.Since(start)

	if err != nil {
		log.Debug().
			Str("key", key).
			Dur("latency", latency).
			Err(err).
			Msg("cache set failed")
		return err
	}

	log.Debug().
		Str("key", key).
		Dur("latency", latency).
		Dur("ttl", expiration).
		Int("size", len(value)).
		Msg("cache set")
	return nil
}

// Delete removes a value from the cache
func (m *MetricsCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := m.cache.Delete(ctx, key)
	latency := time.Since(start)

	log.Debug().
		Str("key", key).
		Dur("latency", latency).
		Err(err).
		Msg("cache delete")
	return err
}

// Exists checks if a key exists in the cache
func (m *MetricsCache) Exists(ctx context.Context, key string) (bool, error) {
	return m.cache.Exists(ctx, key)
}

// Clear removes all keys matching a pattern
func (m *MetricsCache) Clear(ctx context.Context, pattern string) error {
	start := time.Now()
	err := m.cache.Clear(ctx, pattern)
	latency := time.Since(start)

	log.Debug().
		Str("pattern", pattern).
		Dur("latency", latency).
		Err(err).
		Msg("cache clear")
	return err
}

// Close closes the cache connection
func (m *MetricsCache) Close() error {
	return m.cache.Close()
}

// Ping checks the cache connection
func (m *MetricsCache) Ping(ctx context.Context) error {
	return m.cache.Ping(ctx)
}

// IsAvailable returns whether the cache is available
func (m *MetricsCache) IsAvailable() bool {
	return m.cache.IsAvailable()
}

// Stats returns the current hit/miss statistics
func (m *MetricsCache) Stats() (hits, misses int64) {
	return atomic.LoadInt64(&m.hits), atomic.LoadInt64(&m.misses)
}

// HitRate returns the cache hit rate as a percentage (0-100)
func (m *MetricsCache) HitRate() float64 {
	hits := atomic.LoadInt64(&m.hits)
	misses := atomic.LoadInt64(&m.misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// ResetStats resets the hit/miss counters
func (m *MetricsCache) ResetStats() {
	atomic.StoreInt64(&m.hits, 0)
	atomic.StoreInt64(&m.misses, 0)
}

// LogStats logs current cache statistics at info level
func (m *MetricsCache) LogStats() {
	hits, misses := m.Stats()
	log.Info().
		Int64("hits", hits).
		Int64("misses", misses).
		Float64("hit_rate", m.HitRate()).
		Msg("cache statistics")
}
