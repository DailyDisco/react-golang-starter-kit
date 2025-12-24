package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client    *redis.Client
	keyPrefix string
	available bool
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *Config) (*RedisCache, error) {
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, &CacheError{Op: "connect", Err: err}
	}

	// Apply connection pool settings
	opt.PoolSize = config.RedisPoolSize
	opt.MinIdleConns = config.RedisMinIdleConns
	opt.MaxIdleConns = config.RedisMaxIdleConns
	opt.ConnMaxIdleTime = config.RedisConnMaxIdleTime
	opt.ConnMaxLifetime = config.RedisConnMaxLifetime

	client := redis.NewClient(opt)

	return &RedisCache{
		client:    client,
		keyPrefix: config.KeyPrefix,
		available: true,
	}, nil
}

// prefixKey adds the configured prefix to a key
func (c *RedisCache) prefixKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return c.keyPrefix + ":" + key
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.available {
		return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
	}

	val, err := c.client.Get(ctx, c.prefixKey(key)).Bytes()
	if err == redis.Nil {
		return nil, &CacheError{Op: "get", Key: key, Err: ErrCacheMiss.Err}
	}
	if err != nil {
		return nil, &CacheError{Op: "get", Key: key, Err: err}
	}
	return val, nil
}

// Set stores a value in Redis
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if !c.available {
		return nil
	}

	err := c.client.Set(ctx, c.prefixKey(key), value, expiration).Err()
	if err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}
	return nil
}

// Delete removes a value from Redis
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.available {
		return nil
	}

	err := c.client.Del(ctx, c.prefixKey(key)).Err()
	if err != nil {
		return &CacheError{Op: "delete", Key: key, Err: err}
	}
	return nil
}

// Exists checks if a key exists in Redis
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if !c.available {
		return false, nil
	}

	count, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	if err != nil {
		return false, &CacheError{Op: "exists", Key: key, Err: err}
	}
	return count > 0, nil
}

// Clear removes all keys matching a pattern
func (c *RedisCache) Clear(ctx context.Context, pattern string) error {
	if !c.available {
		return nil
	}

	prefixedPattern := c.prefixKey(pattern)
	iter := c.client.Scan(ctx, 0, prefixedPattern, 100).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return &CacheError{Op: "clear", Key: pattern, Err: err}
		}
	}

	if err := iter.Err(); err != nil {
		return &CacheError{Op: "clear", Key: pattern, Err: err}
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.client == nil {
		return nil
	}
	c.available = false
	return c.client.Close()
}

// Ping checks the Redis connection
func (c *RedisCache) Ping(ctx context.Context) error {
	if c.client == nil {
		return &CacheError{Op: "ping", Err: ErrCacheMiss.Err}
	}

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		c.available = false
		return &CacheError{Op: "ping", Err: err}
	}
	c.available = true
	return nil
}

// IsAvailable returns whether Redis is available
func (c *RedisCache) IsAvailable() bool {
	return c.available
}
