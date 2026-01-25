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

// ============ Global Cache Functions Tests ============

func TestIsAvailable_NilInstance(t *testing.T) {
	// Save and restore the global instance
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	if IsAvailable() {
		t.Error("IsAvailable() should return false when instance is nil")
	}
}

func TestIsAvailable_WithNoOpCache(t *testing.T) {
	oldInstance := instance
	instance = NewNoOpCache()
	defer func() { instance = oldInstance }()

	if IsAvailable() {
		t.Error("IsAvailable() should return false for NoOpCache")
	}
}

func TestIsAvailable_WithMemoryCache(t *testing.T) {
	oldInstance := instance
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	instance = NewMemoryCache(config)
	defer func() {
		instance.Close()
		instance = oldInstance
	}()

	if !IsAvailable() {
		t.Error("IsAvailable() should return true for MemoryCache")
	}
}

func TestCheckCacheHealth_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	status := CheckCacheHealth()
	if status.Status != "unhealthy" {
		t.Errorf("CheckCacheHealth() status = %q, want 'unhealthy' when instance is nil", status.Status)
	}
}

func TestCheckCacheHealth_WithMemoryCache(t *testing.T) {
	oldInstance := instance
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	instance = NewMemoryCache(config)
	defer func() {
		instance.Close()
		instance = oldInstance
	}()

	status := CheckCacheHealth()
	if status.Status != "healthy" {
		t.Errorf("CheckCacheHealth() status = %q, want 'healthy' for MemoryCache", status.Status)
	}
}

func TestInstance_NilReturnsNil(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	got := Instance()
	if got != nil {
		t.Error("Instance() should return nil when instance is nil")
	}
}

func TestInstance_ReturnsCache(t *testing.T) {
	oldInstance := instance
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	instance = NewMemoryCache(config)
	defer func() {
		instance.Close()
		instance = oldInstance
	}()

	got := Instance()
	if got == nil {
		t.Error("Instance() should return the cache instance")
	}
	if !got.IsAvailable() {
		t.Error("Instance() should return an available cache")
	}
}

func TestGet_WithNilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	_, err := Get(context.Background(), "key")
	if err == nil {
		t.Error("Get() should return error when instance is nil")
	}
}

func TestSet_WithNilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	// Set silently succeeds when instance is nil (no-op behavior)
	err := Set(context.Background(), "key", []byte("value"), time.Minute)
	if err != nil {
		t.Errorf("Set() should return nil when instance is nil (no-op), got %v", err)
	}
}

func TestGet_WithMemoryCache(t *testing.T) {
	oldInstance := instance
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	instance = NewMemoryCache(config)
	defer func() {
		instance.Close()
		instance = oldInstance
	}()

	ctx := context.Background()

	// Set a value
	err := Set(ctx, "testkey", []byte("testvalue"), time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get the value
	got, err := Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != "testvalue" {
		t.Errorf("Get() = %q, want 'testvalue'", string(got))
	}
}

func TestClose_WithNilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	// Close silently succeeds when instance is nil
	err := Close()
	if err != nil {
		t.Errorf("Close() should return nil when instance is nil, got %v", err)
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

// ============ Org Cache Key Generation Tests ============

func TestOrgSlugKey(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		expected string
	}{
		{"basic slug", "acme-corp", "org:slug:acme-corp"},
		{"empty slug", "", "org:slug:"},
		{"complex slug", "my-awesome-org-123", "org:slug:my-awesome-org-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orgSlugKey(tt.slug)
			if result != tt.expected {
				t.Errorf("orgSlugKey(%q) = %q, want %q", tt.slug, result, tt.expected)
			}
		})
	}
}

func TestOrgIDKey(t *testing.T) {
	tests := []struct {
		name     string
		id       uint
		expected string
	}{
		{"org 1", 1, "org:id:1"},
		{"org 0", 0, "org:id:0"},
		{"large ID", 999999, "org:id:999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orgIDKey(tt.id)
			if result != tt.expected {
				t.Errorf("orgIDKey(%d) = %q, want %q", tt.id, result, tt.expected)
			}
		})
	}
}

func TestMembershipKey(t *testing.T) {
	tests := []struct {
		name     string
		orgID    uint
		userID   uint
		expected string
	}{
		{"basic membership", 1, 2, "membership:1:2"},
		{"same IDs", 5, 5, "membership:5:5"},
		{"zero IDs", 0, 0, "membership:0:0"},
		{"large IDs", 12345, 67890, "membership:12345:67890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := membershipKey(tt.orgID, tt.userID)
			if result != tt.expected {
				t.Errorf("membershipKey(%d, %d) = %q, want %q", tt.orgID, tt.userID, result, tt.expected)
			}
		})
	}
}

// ============ Org Cache Constants Tests ============

func TestOrgCacheConstants(t *testing.T) {
	if OrgBySlugKeyPrefix != "org:slug:" {
		t.Errorf("OrgBySlugKeyPrefix = %q, want %q", OrgBySlugKeyPrefix, "org:slug:")
	}
	if OrgByIDKeyPrefix != "org:id:" {
		t.Errorf("OrgByIDKeyPrefix = %q, want %q", OrgByIDKeyPrefix, "org:id:")
	}
	if MembershipKeyPrefix != "membership:" {
		t.Errorf("MembershipKeyPrefix = %q, want %q", MembershipKeyPrefix, "membership:")
	}
}

// ============ Org Cache TTL Tests ============

func TestGetOrgTTL_Default(t *testing.T) {
	// Save and restore config
	oldConfig := orgCacheConfig
	orgCacheConfig = nil
	defer func() { orgCacheConfig = oldConfig }()

	ttl := getOrgTTL()
	if ttl != 5*time.Minute {
		t.Errorf("getOrgTTL() = %v, want %v", ttl, 5*time.Minute)
	}
}

func TestGetOrgTTL_WithConfig(t *testing.T) {
	oldConfig := orgCacheConfig
	orgCacheConfig = &Config{OrganizationTTL: 10 * time.Minute}
	defer func() { orgCacheConfig = oldConfig }()

	ttl := getOrgTTL()
	if ttl != 10*time.Minute {
		t.Errorf("getOrgTTL() = %v, want %v", ttl, 10*time.Minute)
	}
}

func TestGetOrgTTL_ZeroConfigTTL(t *testing.T) {
	oldConfig := orgCacheConfig
	orgCacheConfig = &Config{OrganizationTTL: 0}
	defer func() { orgCacheConfig = oldConfig }()

	// Should fall back to default when TTL is 0
	ttl := getOrgTTL()
	if ttl != 5*time.Minute {
		t.Errorf("getOrgTTL() with zero TTL = %v, want %v (default)", ttl, 5*time.Minute)
	}
}

func TestGetMembershipTTL_Default(t *testing.T) {
	oldConfig := orgCacheConfig
	orgCacheConfig = nil
	defer func() { orgCacheConfig = oldConfig }()

	ttl := getMembershipTTL()
	if ttl != 5*time.Minute {
		t.Errorf("getMembershipTTL() = %v, want %v", ttl, 5*time.Minute)
	}
}

func TestGetMembershipTTL_WithConfig(t *testing.T) {
	oldConfig := orgCacheConfig
	orgCacheConfig = &Config{MembershipTTL: 15 * time.Minute}
	defer func() { orgCacheConfig = oldConfig }()

	ttl := getMembershipTTL()
	if ttl != 15*time.Minute {
		t.Errorf("getMembershipTTL() = %v, want %v", ttl, 15*time.Minute)
	}
}

func TestSetOrgCacheConfig(t *testing.T) {
	oldConfig := orgCacheConfig
	defer func() { orgCacheConfig = oldConfig }()

	config := &Config{OrganizationTTL: 20 * time.Minute}
	SetOrgCacheConfig(config)

	if orgCacheConfig != config {
		t.Error("SetOrgCacheConfig() did not set config")
	}
}

// ============ Email Template Cache Tests ============

func TestEmailTemplateKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"welcome template", "welcome", "email_template:key:welcome"},
		{"password reset", "password_reset", "email_template:key:password_reset"},
		{"empty key", "", "email_template:key:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailTemplateKey(tt.key)
			if result != tt.expected {
				t.Errorf("emailTemplateKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestEmailTemplateKeyPrefix(t *testing.T) {
	if EmailTemplateKeyPrefix != "email_template:key:" {
		t.Errorf("EmailTemplateKeyPrefix = %q, want %q", EmailTemplateKeyPrefix, "email_template:key:")
	}
}

// ============ Cache Unavailable Scenarios ============

func TestGetOrganization_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	org, err := GetOrganization(ctx, "test-slug")
	if err != nil {
		t.Errorf("GetOrganization() error = %v, want nil", err)
	}
	if org != nil {
		t.Error("GetOrganization() should return nil when cache unavailable")
	}
}

func TestGetOrganizationByID_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	org, err := GetOrganizationByID(ctx, 1)
	if err != nil {
		t.Errorf("GetOrganizationByID() error = %v, want nil", err)
	}
	if org != nil {
		t.Error("GetOrganizationByID() should return nil when cache unavailable")
	}
}

func TestSetOrganization_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := SetOrganization(ctx, nil)
	if err != nil {
		t.Errorf("SetOrganization() error = %v, want nil", err)
	}
}

func TestGetMembership_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	membership, err := GetMembership(ctx, 1, 2)
	if err != nil {
		t.Errorf("GetMembership() error = %v, want nil", err)
	}
	if membership != nil {
		t.Error("GetMembership() should return nil when cache unavailable")
	}
}

func TestSetMembership_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := SetMembership(ctx, nil)
	if err != nil {
		t.Errorf("SetMembership() error = %v, want nil", err)
	}
}

func TestInvalidateOrganization_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := InvalidateOrganization(ctx, "test-slug", 1)
	if err != nil {
		t.Errorf("InvalidateOrganization() error = %v, want nil", err)
	}
}

func TestInvalidateMembership_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := InvalidateMembership(ctx, 1, 2)
	if err != nil {
		t.Errorf("InvalidateMembership() error = %v, want nil", err)
	}
}

func TestInvalidateOrgMemberships_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := InvalidateOrgMemberships(ctx, 1)
	if err != nil {
		t.Errorf("InvalidateOrgMemberships() error = %v, want nil", err)
	}
}

func TestGetEmailTemplate_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	template, err := GetEmailTemplate(ctx, "test")
	if err != nil {
		t.Errorf("GetEmailTemplate() error = %v, want nil", err)
	}
	if template != nil {
		t.Error("GetEmailTemplate() should return nil when cache unavailable")
	}
}

func TestSetEmailTemplate_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := SetEmailTemplate(ctx, nil)
	if err != nil {
		t.Errorf("SetEmailTemplate() error = %v, want nil", err)
	}
}

func TestInvalidateEmailTemplate_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := InvalidateEmailTemplate(ctx, "test")
	if err != nil {
		t.Errorf("InvalidateEmailTemplate() error = %v, want nil", err)
	}
}

func TestInvalidateAllEmailTemplates_Unavailable(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	err := InvalidateAllEmailTemplates(ctx)
	if err != nil {
		t.Errorf("InvalidateAllEmailTemplates() error = %v, want nil", err)
	}
}

// ============ SetIfNotExists Tests ============

func TestSetIfNotExists_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	ctx := context.Background()
	set := SetIfNotExists(ctx, "test-key", []byte("value"), time.Minute)
	// When instance is nil: Exists returns false, Set returns nil (no-op)
	// So SetIfNotExists returns true (key didn't exist, "set" succeeded)
	if !set {
		t.Error("SetIfNotExists() should return true when instance is nil (no-op behavior)")
	}
}

// ============ Delete Global Function Test ============

func TestDelete_WithNilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	err := Delete(context.Background(), "key")
	// Delete returns nil silently when instance is nil
	if err != nil {
		t.Errorf("Delete() with nil instance = %v, want nil", err)
	}
}

func TestDelete_WithMemoryCache(t *testing.T) {
	oldInstance := instance
	config := &Config{
		KeyPrefix:             "test",
		MemoryMaxSize:         100,
		MemoryCleanupInterval: time.Hour,
	}
	instance = NewMemoryCache(config)
	defer func() {
		instance.Close()
		instance = oldInstance
	}()

	ctx := context.Background()

	// Set a value
	Set(ctx, "deletetest", []byte("value"), time.Minute)

	// Delete it
	err := Delete(ctx, "deletetest")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = Get(ctx, "deletetest")
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}
}
