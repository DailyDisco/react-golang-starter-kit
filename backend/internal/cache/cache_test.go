package cache

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

// ============ Config Tests ============

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Enabled {
		t.Error("DefaultConfig() Enabled should be false")
	}
	if config.Type != "memory" {
		t.Errorf("DefaultConfig() Type = %v, want memory", config.Type)
	}
	if config.KeyPrefix != "app" {
		t.Errorf("DefaultConfig() KeyPrefix = %v, want app", config.KeyPrefix)
	}
	if config.MemoryMaxSize != 10000 {
		t.Errorf("DefaultConfig() MemoryMaxSize = %v, want 10000", config.MemoryMaxSize)
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		check    func(*Config) bool
		checkMsg string
	}{
		{
			name:     "default values",
			envVars:  map[string]string{},
			check:    func(c *Config) bool { return !c.Enabled && c.Type == "memory" },
			checkMsg: "should have default values",
		},
		{
			name:     "CACHE_ENABLED true",
			envVars:  map[string]string{"CACHE_ENABLED": "true"},
			check:    func(c *Config) bool { return c.Enabled },
			checkMsg: "should be enabled",
		},
		{
			name:     "CACHE_ENABLED false",
			envVars:  map[string]string{"CACHE_ENABLED": "false"},
			check:    func(c *Config) bool { return !c.Enabled },
			checkMsg: "should be disabled",
		},
		{
			name:     "CACHE_TYPE redis",
			envVars:  map[string]string{"CACHE_TYPE": "redis"},
			check:    func(c *Config) bool { return c.Type == "redis" },
			checkMsg: "should be redis type",
		},
		{
			name:     "REDIS_URL auto-switches type",
			envVars:  map[string]string{"REDIS_URL": "redis://localhost:6379"},
			check:    func(c *Config) bool { return c.Type == "redis" && c.RedisURL == "redis://localhost:6379" },
			checkMsg: "should auto-switch to redis type",
		},
		{
			name:     "CACHE_KEY_PREFIX",
			envVars:  map[string]string{"CACHE_KEY_PREFIX": "myapp"},
			check:    func(c *Config) bool { return c.KeyPrefix == "myapp" },
			checkMsg: "should set key prefix",
		},
		{
			name:     "REDIS_POOL_SIZE",
			envVars:  map[string]string{"REDIS_POOL_SIZE": "20"},
			check:    func(c *Config) bool { return c.RedisPoolSize == 20 },
			checkMsg: "should set pool size",
		},
		{
			name:     "CACHE_MEMORY_MAX_SIZE",
			envVars:  map[string]string{"CACHE_MEMORY_MAX_SIZE": "5000"},
			check:    func(c *Config) bool { return c.MemoryMaxSize == 5000 },
			checkMsg: "should set memory max size",
		},
		{
			name:     "CACHE_DEFAULT_TTL",
			envVars:  map[string]string{"CACHE_DEFAULT_TTL": "120"},
			check:    func(c *Config) bool { return c.DefaultTTL == 120*time.Second },
			checkMsg: "should set default TTL",
		},
		{
			name:     "invalid REDIS_POOL_SIZE ignored",
			envVars:  map[string]string{"REDIS_POOL_SIZE": "invalid"},
			check:    func(c *Config) bool { return c.RedisPoolSize == 10 }, // default
			checkMsg: "should keep default pool size for invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars
			envKeys := []string{
				"CACHE_ENABLED", "CACHE_TYPE", "CACHE_KEY_PREFIX",
				"REDIS_URL", "REDIS_POOL_SIZE", "REDIS_MIN_IDLE_CONNS", "REDIS_MAX_IDLE_CONNS",
				"CACHE_MEMORY_MAX_SIZE", "CACHE_DEFAULT_TTL", "CACHE_HEALTH_CHECK_TTL",
				"CACHE_USER_PROFILE_TTL", "CACHE_SESSION_TTL",
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set test env vars
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			config := LoadConfig()
			if !tt.check(config) {
				t.Errorf("LoadConfig() %s", tt.checkMsg)
			}

			// Cleanup
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

// ============ MemoryCache Tests ============

func newTestMemoryCache() *MemoryCache {
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: 1 * time.Hour, // Long interval to avoid interference
	}
	return NewMemoryCache(config)
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	tests := []struct {
		name       string
		key        string
		value      []byte
		expiration time.Duration
		wantErr    bool
	}{
		{"simple value", "key1", []byte("value1"), time.Minute, false},
		{"empty value", "key2", []byte(""), time.Minute, false},
		{"binary value", "key3", []byte{0x00, 0x01, 0x02}, time.Minute, false},
		{"no expiration", "key4", []byte("persistent"), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := cache.Get(ctx, tt.key)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			if string(got) != string(tt.value) {
				t.Errorf("Get() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent key")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// Set a value
	cache.Set(ctx, "deletekey", []byte("value"), time.Minute)

	// Delete it
	err := cache.Delete(ctx, "deletekey")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = cache.Get(ctx, "deletekey")
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}
}

func TestMemoryCache_DeleteNonExistent(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// Deleting non-existent key should not error
	err := cache.Delete(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Delete() should not error for non-existent key, got %v", err)
	}
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// Non-existent key
	exists, err := cache.Exists(ctx, "missing")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false for non-existent key")
	}

	// Set a key
	cache.Set(ctx, "existskey", []byte("value"), time.Minute)

	// Check it exists
	exists, err = cache.Exists(ctx, "existskey")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true for existing key")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// Set with very short expiration
	cache.Set(ctx, "expiring", []byte("value"), 10*time.Millisecond)

	// Should exist immediately
	_, err := cache.Get(ctx, "expiring")
	if err != nil {
		t.Error("Get() should succeed immediately after Set()")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "expiring")
	if err == nil {
		t.Error("Get() should return error for expired key")
	}

	// Exists should also return false
	exists, _ := cache.Exists(ctx, "expiring")
	if exists {
		t.Error("Exists() should return false for expired key")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// Set multiple keys
	cache.Set(ctx, "user:1", []byte("data1"), time.Minute)
	cache.Set(ctx, "user:2", []byte("data2"), time.Minute)
	cache.Set(ctx, "session:1", []byte("session"), time.Minute)

	// Clear with wildcard pattern
	err := cache.Clear(ctx, "user:*")
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// User keys should be gone
	_, err = cache.Get(ctx, "user:1")
	if err == nil {
		t.Error("user:1 should be cleared")
	}
	_, err = cache.Get(ctx, "user:2")
	if err == nil {
		t.Error("user:2 should be cleared")
	}

	// Session key should remain
	_, err = cache.Get(ctx, "session:1")
	if err != nil {
		t.Error("session:1 should not be cleared")
	}
}

func TestMemoryCache_ClearExact(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	cache.Set(ctx, "exactkey", []byte("value"), time.Minute)
	cache.Set(ctx, "exactkey2", []byte("value2"), time.Minute)

	// Clear exact key (no wildcard)
	cache.Clear(ctx, "exactkey")

	_, err := cache.Get(ctx, "exactkey")
	if err == nil {
		t.Error("exactkey should be cleared")
	}

	_, err = cache.Get(ctx, "exactkey2")
	if err != nil {
		t.Error("exactkey2 should not be cleared")
	}
}

func TestMemoryCache_MaxSize(t *testing.T) {
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         3,
		MemoryCleanupInterval: 1 * time.Hour,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()
	ctx := context.Background()

	// Fill cache to max
	cache.Set(ctx, "key1", []byte("v1"), time.Minute)
	cache.Set(ctx, "key2", []byte("v2"), time.Minute)
	cache.Set(ctx, "key3", []byte("v3"), time.Minute)

	// Add one more - should evict one
	cache.Set(ctx, "key4", []byte("v4"), time.Minute)

	// key4 should exist
	_, err := cache.Get(ctx, "key4")
	if err != nil {
		t.Error("key4 should exist")
	}

	// Count remaining keys (should be at most maxSize)
	count := 0
	for _, key := range []string{"key1", "key2", "key3", "key4"} {
		if _, err := cache.Get(ctx, key); err == nil {
			count++
		}
	}
	if count > 3 {
		t.Errorf("Cache should have at most 3 items, got %d", count)
	}
}

func TestMemoryCache_Concurrency(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "concurrent"
				cache.Set(ctx, key, []byte("value"), time.Minute)
				cache.Get(ctx, key)
				cache.Exists(ctx, key)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions or panics occur
}

func TestMemoryCache_PrefixKey(t *testing.T) {
	// Test with prefix
	config := &Config{
		KeyPrefix:             "myprefix",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	prefixed := cache.prefixKey("testkey")
	if prefixed != "myprefix:testkey" {
		t.Errorf("prefixKey() = %v, want myprefix:testkey", prefixed)
	}

	// Test without prefix
	config2 := &Config{
		KeyPrefix:             "",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	cache2 := NewMemoryCache(config2)
	defer cache2.Close()

	unprefixed := cache2.prefixKey("testkey")
	if unprefixed != "testkey" {
		t.Errorf("prefixKey() = %v, want testkey", unprefixed)
	}
}

func TestMemoryCache_Ping(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestMemoryCache_IsAvailable(t *testing.T) {
	cache := newTestMemoryCache()
	defer cache.Close()

	if !cache.IsAvailable() {
		t.Error("IsAvailable() should return true for MemoryCache")
	}
}

// ============ NoOpCache Tests ============

func TestNoOpCache_Get(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	_, err := cache.Get(ctx, "anykey")
	if err == nil {
		t.Error("NoOpCache.Get() should always return error (cache miss)")
	}
}

func TestNoOpCache_Set(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	err := cache.Set(ctx, "key", []byte("value"), time.Minute)
	if err != nil {
		t.Errorf("NoOpCache.Set() error = %v, want nil", err)
	}
}

func TestNoOpCache_Delete(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	err := cache.Delete(ctx, "key")
	if err != nil {
		t.Errorf("NoOpCache.Delete() error = %v, want nil", err)
	}
}

func TestNoOpCache_Exists(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	exists, err := cache.Exists(ctx, "key")
	if err != nil {
		t.Errorf("NoOpCache.Exists() error = %v", err)
	}
	if exists {
		t.Error("NoOpCache.Exists() should always return false")
	}
}

func TestNoOpCache_Clear(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	err := cache.Clear(ctx, "pattern*")
	if err != nil {
		t.Errorf("NoOpCache.Clear() error = %v, want nil", err)
	}
}

func TestNoOpCache_Close(t *testing.T) {
	cache := NewNoOpCache()

	err := cache.Close()
	if err != nil {
		t.Errorf("NoOpCache.Close() error = %v, want nil", err)
	}
}

func TestNoOpCache_Ping(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Errorf("NoOpCache.Ping() error = %v, want nil", err)
	}
}

func TestNoOpCache_IsAvailable(t *testing.T) {
	cache := NewNoOpCache()

	if cache.IsAvailable() {
		t.Error("NoOpCache.IsAvailable() should return false")
	}
}

// ============ CacheError Tests ============

func TestCacheError_Error(t *testing.T) {
	// Create a real error for testing
	testErr := context.DeadlineExceeded

	tests := []struct {
		name     string
		err      *CacheError
		contains string
	}{
		{
			name:     "with key",
			err:      &CacheError{Op: "get", Key: "mykey", Err: testErr},
			contains: "cache get mykey",
		},
		{
			name:     "without key",
			err:      &CacheError{Op: "ping", Key: "", Err: testErr},
			contains: "cache ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestCacheError_Unwrap(t *testing.T) {
	innerErr := context.DeadlineExceeded
	cacheErr := &CacheError{Op: "get", Key: "key", Err: innerErr}

	if cacheErr.Unwrap() != innerErr {
		t.Error("Unwrap() should return inner error")
	}
}

// ============ Redis Integration Tests (skip if unavailable) ============

func skipIfNoRedis(t *testing.T) *RedisCache {
	t.Helper()
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("Skipping Redis test: TEST_REDIS_URL not set")
	}

	config := &Config{
		Enabled:              true,
		Type:                 "redis",
		RedisURL:             redisURL,
		KeyPrefix:            "test",
		RedisPoolSize:        5,
		RedisMinIdleConns:    1,
		RedisMaxIdleConns:    3,
		RedisConnMaxIdleTime: time.Minute,
		RedisConnMaxLifetime: 5 * time.Minute,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Skipping Redis test: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cache.Ping(ctx); err != nil {
		cache.Close()
		t.Skipf("Skipping Redis test: ping failed: %v", err)
	}

	return cache
}

func TestRedisCache_Integration(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	// Test Set and Get
	testKey := "integration:test"
	testValue := []byte("test-value")

	err := cache.Set(ctx, testKey, testValue, time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := cache.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != string(testValue) {
		t.Errorf("Get() = %v, want %v", string(got), string(testValue))
	}

	// Test Exists
	exists, err := cache.Exists(ctx, testKey)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true")
	}

	// Test Delete
	err = cache.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = cache.Get(ctx, testKey)
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}

	// Test IsAvailable
	if !cache.IsAvailable() {
		t.Error("IsAvailable() should return true after successful ping")
	}
}
