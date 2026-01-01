package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCache_Get_Hit(t *testing.T) {
	// Setup underlying cache with test data
	underlying := NewMemoryCache(&Config{Enabled: true})
	ctx := context.Background()

	key := "test_key"
	value := []byte("test_value")
	err := underlying.Set(ctx, key, value, time.Minute)
	require.NoError(t, err)

	// Wrap with metrics
	metrics := NewMetricsCache(underlying)

	// Get should be a hit
	data, err := metrics.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, data)

	// Check stats
	hits, misses := metrics.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(0), misses)
	assert.Equal(t, 100.0, metrics.HitRate())
}

func TestMetricsCache_Get_Miss(t *testing.T) {
	underlying := NewMemoryCache(&Config{Enabled: true})
	ctx := context.Background()

	metrics := NewMetricsCache(underlying)

	// Get non-existent key should be a miss
	_, err := metrics.Get(ctx, "nonexistent")
	assert.Error(t, err)

	// Check stats
	hits, misses := metrics.Stats()
	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(1), misses)
	assert.Equal(t, 0.0, metrics.HitRate())
}

func TestMetricsCache_MixedOperations(t *testing.T) {
	underlying := NewMemoryCache(&Config{Enabled: true})
	ctx := context.Background()

	metrics := NewMetricsCache(underlying)

	// Set a value
	err := metrics.Set(ctx, "key1", []byte("value1"), time.Minute)
	require.NoError(t, err)

	// Hit
	_, err = metrics.Get(ctx, "key1")
	require.NoError(t, err)

	// Miss
	_, err = metrics.Get(ctx, "key2")
	assert.Error(t, err)

	// Another hit
	_, err = metrics.Get(ctx, "key1")
	require.NoError(t, err)

	// Check stats: 2 hits, 1 miss
	hits, misses := metrics.Stats()
	assert.Equal(t, int64(2), hits)
	assert.Equal(t, int64(1), misses)

	// Hit rate should be 66.67%
	hitRate := metrics.HitRate()
	assert.InDelta(t, 66.67, hitRate, 0.1)
}

func TestMetricsCache_ResetStats(t *testing.T) {
	underlying := NewMemoryCache(&Config{Enabled: true})
	ctx := context.Background()

	metrics := NewMetricsCache(underlying)

	// Generate some stats
	_ = metrics.Set(ctx, "key", []byte("value"), time.Minute)
	_, _ = metrics.Get(ctx, "key")
	_, _ = metrics.Get(ctx, "missing")

	hits, misses := metrics.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)

	// Reset
	metrics.ResetStats()

	hits, misses = metrics.Stats()
	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(0), misses)
	assert.Equal(t, 0.0, metrics.HitRate())
}

func TestMetricsCache_DelegatesAllMethods(t *testing.T) {
	underlying := NewMemoryCache(&Config{Enabled: true})
	ctx := context.Background()

	metrics := NewMetricsCache(underlying)

	// Test Set
	err := metrics.Set(ctx, "key", []byte("value"), time.Minute)
	require.NoError(t, err)

	// Test Exists
	exists, err := metrics.Exists(ctx, "key")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	err = metrics.Delete(ctx, "key")
	require.NoError(t, err)

	exists, err = metrics.Exists(ctx, "key")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Clear
	_ = metrics.Set(ctx, "prefix:1", []byte("v1"), time.Minute)
	_ = metrics.Set(ctx, "prefix:2", []byte("v2"), time.Minute)
	err = metrics.Clear(ctx, "prefix:*")
	require.NoError(t, err)

	// Test Ping
	err = metrics.Ping(ctx)
	require.NoError(t, err)

	// Test IsAvailable
	assert.True(t, metrics.IsAvailable())
}

func TestMetricsCache_HitRateZeroTotal(t *testing.T) {
	underlying := NewMemoryCache(&Config{Enabled: true})
	metrics := NewMetricsCache(underlying)

	// No operations yet, hit rate should be 0 (not NaN)
	assert.Equal(t, 0.0, metrics.HitRate())
}
