package cache

import (
	"context"
	"encoding/json"
	"time"

	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with an expiration time
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all keys matching a pattern
	Clear(ctx context.Context, pattern string) error

	// Close closes the cache connection
	Close() error

	// Ping checks the cache connection
	Ping(ctx context.Context) error

	// IsAvailable returns whether the cache is available
	IsAvailable() bool
}

// CacheError represents a cache-specific error
type CacheError struct {
	Op  string
	Key string
	Err error
}

func (e *CacheError) Error() string {
	if e.Key != "" {
		return "cache " + e.Op + " " + e.Key + ": " + e.Err.Error()
	}
	return "cache " + e.Op + ": " + e.Err.Error()
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// ErrCacheMiss is returned when a key is not found in the cache
var ErrCacheMiss = &CacheError{Op: "get", Err: nil}

// instance holds the global cache instance
var instance Cache

// Initialize sets up the cache based on configuration
func Initialize(config *Config) error {
	if !config.Enabled {
		log.Info().Msg("cache disabled, using no-op cache")
		instance = NewNoOpCache()
		return nil
	}

	if config.Type == "redis" && config.RedisURL != "" {
		redisCache, err := NewRedisCache(config)
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to Redis, falling back to in-memory cache")
			instance = NewMemoryCache(config)
			return nil
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisCache.Ping(ctx); err != nil {
			log.Warn().Err(err).Msg("Redis ping failed, falling back to in-memory cache")
			redisCache.Close()
			instance = NewMemoryCache(config)
			return nil
		}

		log.Info().Str("url", config.RedisURL).Msg("Redis cache initialized")
		instance = redisCache
		return nil
	}

	// Default to in-memory cache
	log.Info().Msg("using in-memory cache")
	instance = NewMemoryCache(config)
	return nil
}

// Get retrieves a value from the global cache instance
func Get(ctx context.Context, key string) ([]byte, error) {
	if instance == nil {
		return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
	}
	return instance.Get(ctx, key)
}

// Set stores a value in the global cache instance
func Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if instance == nil {
		return nil
	}
	return instance.Set(ctx, key, value, expiration)
}

// Delete removes a value from the global cache instance
func Delete(ctx context.Context, key string) error {
	if instance == nil {
		return nil
	}
	return instance.Delete(ctx, key)
}

// GetJSON retrieves and unmarshals a JSON value from the cache
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals and stores a JSON value in the cache
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}
	return Set(ctx, key, data, expiration)
}

// Instance returns the global cache instance
func Instance() Cache {
	return instance
}

// Close closes the global cache instance
func Close() error {
	if instance == nil {
		return nil
	}
	return instance.Close()
}

// IsAvailable returns whether the cache is available
func IsAvailable() bool {
	if instance == nil {
		return false
	}
	return instance.IsAvailable()
}

// CheckCacheHealth returns the health status of the cache
func CheckCacheHealth() models.ComponentStatus {
	if instance == nil {
		return models.ComponentStatus{
			Name:    "cache",
			Status:  "unhealthy",
			Message: "cache not initialized",
		}
	}

	if !instance.IsAvailable() {
		return models.ComponentStatus{
			Name:    "cache",
			Status:  "degraded",
			Message: "cache unavailable (no-op mode)",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := instance.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return models.ComponentStatus{
			Name:    "cache",
			Status:  "unhealthy",
			Message: "failed to ping cache: " + err.Error(),
		}
	}

	return models.ComponentStatus{
		Name:    "cache",
		Status:  "healthy",
		Message: "cache responding normally",
		Latency: latency.String(),
	}
}
