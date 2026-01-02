package cache

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds cache configuration settings
type Config struct {
	// General settings
	Enabled bool
	Type    string // "redis" or "memory"

	// Key prefix for namespacing
	KeyPrefix string

	// Redis settings
	RedisURL             string
	RedisPoolSize        int
	RedisMinIdleConns    int
	RedisMaxIdleConns    int
	RedisConnMaxIdleTime time.Duration
	RedisConnMaxLifetime time.Duration

	// Memory cache settings
	MemoryMaxSize         int
	MemoryCleanupInterval time.Duration

	// Default TTL for cached items
	DefaultTTL time.Duration

	// TTL for specific cache types
	HealthCheckTTL  time.Duration
	UserProfileTTL  time.Duration
	SessionTTL      time.Duration
	OrganizationTTL time.Duration
	MembershipTTL   time.Duration
}

// DefaultConfig returns sensible default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:   false, // Disabled by default
		Type:      "memory",
		KeyPrefix: "app",

		// Redis defaults
		RedisURL:             "",
		RedisPoolSize:        10,
		RedisMinIdleConns:    2,
		RedisMaxIdleConns:    5,
		RedisConnMaxIdleTime: 5 * time.Minute,
		RedisConnMaxLifetime: 30 * time.Minute,

		// Memory cache defaults
		MemoryMaxSize:         10000, // 10k items max
		MemoryCleanupInterval: 1 * time.Minute,

		// TTL defaults
		DefaultTTL:      5 * time.Minute,
		HealthCheckTTL:  30 * time.Second,
		UserProfileTTL:  2 * time.Minute,
		SessionTTL:      15 * time.Minute,
		OrganizationTTL: 5 * time.Minute,
		MembershipTTL:   5 * time.Minute,
	}
}

// LoadConfig loads cache configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	// General settings
	if enabled := os.Getenv("CACHE_ENABLED"); enabled != "" {
		config.Enabled = strings.ToLower(enabled) == "true"
	}

	if cacheType := os.Getenv("CACHE_TYPE"); cacheType != "" {
		config.Type = strings.ToLower(cacheType)
	}

	if prefix := os.Getenv("CACHE_KEY_PREFIX"); prefix != "" {
		config.KeyPrefix = prefix
	}

	// Redis settings
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.RedisURL = redisURL
		config.Type = "redis" // Auto-switch to redis if URL is provided
	}

	if poolSize := os.Getenv("REDIS_POOL_SIZE"); poolSize != "" {
		if size, err := strconv.Atoi(poolSize); err == nil && size > 0 {
			config.RedisPoolSize = size
		}
	}

	if minIdle := os.Getenv("REDIS_MIN_IDLE_CONNS"); minIdle != "" {
		if conns, err := strconv.Atoi(minIdle); err == nil && conns >= 0 {
			config.RedisMinIdleConns = conns
		}
	}

	if maxIdle := os.Getenv("REDIS_MAX_IDLE_CONNS"); maxIdle != "" {
		if conns, err := strconv.Atoi(maxIdle); err == nil && conns >= 0 {
			config.RedisMaxIdleConns = conns
		}
	}

	// Memory cache settings
	if maxSize := os.Getenv("CACHE_MEMORY_MAX_SIZE"); maxSize != "" {
		if size, err := strconv.Atoi(maxSize); err == nil && size > 0 {
			config.MemoryMaxSize = size
		}
	}

	// TTL settings (in seconds)
	if ttl := os.Getenv("CACHE_DEFAULT_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.DefaultTTL = time.Duration(seconds) * time.Second
		}
	}

	if ttl := os.Getenv("CACHE_HEALTH_CHECK_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.HealthCheckTTL = time.Duration(seconds) * time.Second
		}
	}

	if ttl := os.Getenv("CACHE_USER_PROFILE_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.UserProfileTTL = time.Duration(seconds) * time.Second
		}
	}

	if ttl := os.Getenv("CACHE_SESSION_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.SessionTTL = time.Duration(seconds) * time.Second
		}
	}

	if ttl := os.Getenv("CACHE_ORGANIZATION_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.OrganizationTTL = time.Duration(seconds) * time.Second
		}
	}

	if ttl := os.Getenv("CACHE_MEMBERSHIP_TTL"); ttl != "" {
		if seconds, err := strconv.Atoi(ttl); err == nil && seconds > 0 {
			config.MembershipTTL = time.Duration(seconds) * time.Second
		}
	}

	return config
}
