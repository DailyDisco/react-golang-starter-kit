package cache

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// CacheAside implements the cache-aside pattern with graceful fallback.
// It first tries to get the value from cache, and if not found,
// loads it from the source and caches the result.
//
// On cache miss: calls loader, caches result, returns value
// On cache hit: returns cached value
// On cache error: logs warning, calls loader, returns value without caching
func CacheAside[T any](
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	var result T

	// Try to get from cache
	err := GetJSON(ctx, key, &result)
	if err == nil {
		// Cache hit
		return result, nil
	}

	// Cache miss or error - load from source
	result, err = loader()
	if err != nil {
		return result, err
	}

	// Cache the result asynchronously (don't block the response)
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if setErr := SetJSON(cacheCtx, key, result, ttl); setErr != nil {
			log.Debug().Err(setErr).Str("key", key).Msg("failed to cache result")
		}
	}()

	return result, nil
}

// CacheAsideSync is like CacheAside but caches synchronously.
// Use this when you need to ensure the value is cached before returning.
func CacheAsideSync[T any](
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	var result T

	// Try to get from cache
	err := GetJSON(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// Load from source
	result, err = loader()
	if err != nil {
		return result, err
	}

	// Cache synchronously
	if setErr := SetJSON(ctx, key, result, ttl); setErr != nil {
		log.Debug().Err(setErr).Str("key", key).Msg("failed to cache result")
	}

	return result, nil
}

// Invalidate removes a key from the cache.
// This should be called when the underlying data changes.
func Invalidate(ctx context.Context, key string) {
	if err := Delete(ctx, key); err != nil {
		log.Debug().Err(err).Str("key", key).Msg("failed to invalidate cache")
	}
}

// InvalidatePattern removes all keys matching a pattern from the cache.
func InvalidatePattern(ctx context.Context, pattern string) {
	if instance == nil {
		return
	}
	if err := instance.Clear(ctx, pattern); err != nil {
		log.Debug().Err(err).Str("pattern", pattern).Msg("failed to invalidate cache pattern")
	}
}

// Exists checks if a key exists in the cache.
// Returns false on cache error (fail-open behavior for existence checks).
func Exists(ctx context.Context, key string) bool {
	if instance == nil {
		return false
	}
	exists, err := instance.Exists(ctx, key)
	if err != nil {
		return false
	}
	return exists
}

// SetIfNotExists sets a key only if it doesn't already exist.
// Returns true if the key was set, false if it already existed.
func SetIfNotExists(ctx context.Context, key string, value []byte, ttl time.Duration) bool {
	if Exists(ctx, key) {
		return false
	}
	if err := Set(ctx, key, value, ttl); err != nil {
		return false
	}
	return true
}

// BlacklistCacheKey generates a cache key for token blacklist lookups.
func BlacklistCacheKey(tokenHashPrefix string) string {
	return "blacklist:" + tokenHashPrefix
}

// UserCacheKey generates a cache key for user data.
func UserCacheKey(userID uint) string {
	return "user:" + string(rune(userID))
}

// FeatureFlagsCacheKey is the cache key for all feature flags.
const FeatureFlagsCacheKey = "feature_flags:all"

// SessionCacheKey generates a cache key for session data.
func SessionCacheKey(sessionID string) string {
	return "session:" + sessionID
}
