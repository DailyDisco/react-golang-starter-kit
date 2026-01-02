package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// sfGroup coalesces concurrent requests for the same cache key during cache misses.
// This prevents cache stampedes where many goroutines simultaneously try to load
// the same data when the cache expires.
var sfGroup singleflight.Group

// CacheAside implements the cache-aside pattern with graceful fallback
// and singleflight protection against cache stampedes.
//
// On cache hit: returns cached value immediately
// On cache miss: uses singleflight to coalesce concurrent requests,
// calls loader once, caches result asynchronously, returns value to all waiters
// On cache error: logs warning, falls through to loader
func CacheAside[T any](
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	var result T

	// Try to get from cache first
	err := GetJSON(ctx, key, &result)
	if err == nil {
		// Cache hit
		return result, nil
	}

	// Cache miss - use singleflight to prevent stampede
	// All concurrent requests for the same key will wait for the first one to complete
	v, err, shared := sfGroup.Do(key, func() (interface{}, error) {
		// Double-check cache in case another goroutine populated it while we waited
		var cachedResult T
		if cacheErr := GetJSON(ctx, key, &cachedResult); cacheErr == nil {
			return cachedResult, nil
		}

		// Load from source
		loaded, loadErr := loader()
		if loadErr != nil {
			return loaded, loadErr
		}

		// Cache the result asynchronously (don't block the response)
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if setErr := SetJSON(cacheCtx, key, loaded, ttl); setErr != nil {
				log.Debug().Err(setErr).Str("key", key).Msg("failed to cache result")
			}
		}()

		return loaded, nil
	})

	if err != nil {
		return result, err
	}

	if shared {
		log.Debug().Str("key", key).Msg("cache request coalesced via singleflight")
	}

	return v.(T), nil
}

// CacheAsideSync is like CacheAside but caches synchronously.
// Use this when you need to ensure the value is cached before returning.
// Also uses singleflight for stampede protection.
func CacheAsideSync[T any](
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func() (T, error),
) (T, error) {
	var result T

	// Try to get from cache first
	err := GetJSON(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// Cache miss - use singleflight to prevent stampede
	v, err, _ := sfGroup.Do(key, func() (interface{}, error) {
		// Double-check cache
		var cachedResult T
		if cacheErr := GetJSON(ctx, key, &cachedResult); cacheErr == nil {
			return cachedResult, nil
		}

		// Load from source
		loaded, loadErr := loader()
		if loadErr != nil {
			return loaded, loadErr
		}

		// Cache synchronously
		if setErr := SetJSON(ctx, key, loaded, ttl); setErr != nil {
			log.Debug().Err(setErr).Str("key", key).Msg("failed to cache result")
		}

		return loaded, nil
	})

	if err != nil {
		return result, err
	}

	return v.(T), nil
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
	return "user:" + strconv.FormatUint(uint64(userID), 10)
}

// FeatureFlagsCacheKey is the cache key for all feature flags.
const FeatureFlagsCacheKey = "feature_flags:all"

// SessionCacheKey generates a cache key for session data.
func SessionCacheKey(sessionID string) string {
	return "session:" + sessionID
}
