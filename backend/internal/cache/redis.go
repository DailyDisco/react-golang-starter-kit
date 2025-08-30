package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	zerologlog "github.com/rs/zerolog/log"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Client represents a Redis client wrapper
type Client struct {
	*redis.Client
	ctx context.Context
}

// Cache represents the caching service
type Cache struct {
	client *Client
}

// NewRedisConfig creates Redis configuration from environment variables
func NewRedisConfig() *RedisConfig {
	host := getEnv("REDIS_HOST", "localhost")
	portStr := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	dbStr := getEnv("REDIS_DB", "0")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid REDIS_PORT: %s, using default 6379", portStr)
		port = 6379
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB: %s, using default 0", dbStr)
		db = 0
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

// NewClient creates a new Redis client
func NewClient(config *RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	client := &Client{
		Client: rdb,
		ctx:    context.Background(),
	}

	return client
}

// ConnectRedis initializes and tests the Redis connection
func ConnectRedis() *Client {
	config := NewRedisConfig()

	zerologlog.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Int("db", config.DB).
		Msg("connecting to Redis")

	client := NewClient(config)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		zerologlog.Error().
			Err(err).
			Str("host", config.Host).
			Int("port", config.Port).
			Msg("failed to connect to Redis")
		log.Fatal("Failed to connect to Redis:", err)
	}

	zerologlog.Info().Msg("Redis connected successfully")
	return client
}

// NewCache creates a new cache service
func NewCache(client *Client) *Cache {
	return &Cache{
		client: client,
	}
}

// Set stores a value in cache with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	return c.client.Set(c.client.ctx, key, data, ttl).Err()
}

// Get retrieves a value from cache
func (c *Cache) Get(key string, dest interface{}) error {
	data, err := c.client.Get(c.client.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *Cache) Delete(key string) error {
	return c.client.Del(c.client.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *Cache) DeletePattern(pattern string) error {
	keys, err := c.client.Keys(c.client.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find keys with pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(c.client.ctx, keys...).Err()
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(key string) bool {
	count, err := c.client.Exists(c.client.ctx, key).Result()
	return err == nil && count > 0
}

// SetTTL sets TTL for an existing key
func (c *Cache) SetTTL(key string, ttl time.Duration) error {
	return c.client.Expire(c.client.ctx, key, ttl).Err()
}

// GetTTL gets remaining TTL for a key
func (c *Cache) GetTTL(key string) (time.Duration, error) {
	ttl, err := c.client.TTL(c.client.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// FlushDB clears all keys in the current database
func (c *Cache) FlushDB() error {
	return c.client.FlushDB(c.client.ctx).Err()
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	return c.Client.Close()
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
