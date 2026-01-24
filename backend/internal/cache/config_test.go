package cache

import (
	"testing"
	"time"
)

// ============ DefaultConfig Tests ============

func TestDefaultConfig_AllFields(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// General settings
	if config.Enabled {
		t.Error("DefaultConfig().Enabled = true, want false (disabled by default)")
	}
	if config.Type != "memory" {
		t.Errorf("DefaultConfig().Type = %q, want 'memory'", config.Type)
	}
	if config.KeyPrefix != "app" {
		t.Errorf("DefaultConfig().KeyPrefix = %q, want 'app'", config.KeyPrefix)
	}

	// Redis defaults
	if config.RedisURL != "" {
		t.Errorf("DefaultConfig().RedisURL = %q, want empty", config.RedisURL)
	}
	if config.RedisPoolSize != 10 {
		t.Errorf("DefaultConfig().RedisPoolSize = %d, want 10", config.RedisPoolSize)
	}
	if config.RedisMinIdleConns != 2 {
		t.Errorf("DefaultConfig().RedisMinIdleConns = %d, want 2", config.RedisMinIdleConns)
	}
	if config.RedisMaxIdleConns != 5 {
		t.Errorf("DefaultConfig().RedisMaxIdleConns = %d, want 5", config.RedisMaxIdleConns)
	}
	if config.RedisConnMaxIdleTime != 5*time.Minute {
		t.Errorf("DefaultConfig().RedisConnMaxIdleTime = %v, want 5m", config.RedisConnMaxIdleTime)
	}
	if config.RedisConnMaxLifetime != 30*time.Minute {
		t.Errorf("DefaultConfig().RedisConnMaxLifetime = %v, want 30m", config.RedisConnMaxLifetime)
	}

	// Memory cache defaults
	if config.MemoryMaxSize != 10000 {
		t.Errorf("DefaultConfig().MemoryMaxSize = %d, want 10000", config.MemoryMaxSize)
	}
	if config.MemoryCleanupInterval != 1*time.Minute {
		t.Errorf("DefaultConfig().MemoryCleanupInterval = %v, want 1m", config.MemoryCleanupInterval)
	}

	// TTL defaults
	if config.DefaultTTL != 5*time.Minute {
		t.Errorf("DefaultConfig().DefaultTTL = %v, want 5m", config.DefaultTTL)
	}
	if config.HealthCheckTTL != 30*time.Second {
		t.Errorf("DefaultConfig().HealthCheckTTL = %v, want 30s", config.HealthCheckTTL)
	}
	if config.UserProfileTTL != 2*time.Minute {
		t.Errorf("DefaultConfig().UserProfileTTL = %v, want 2m", config.UserProfileTTL)
	}
	if config.SessionTTL != 15*time.Minute {
		t.Errorf("DefaultConfig().SessionTTL = %v, want 15m", config.SessionTTL)
	}
	if config.OrganizationTTL != 5*time.Minute {
		t.Errorf("DefaultConfig().OrganizationTTL = %v, want 5m", config.OrganizationTTL)
	}
	if config.MembershipTTL != 5*time.Minute {
		t.Errorf("DefaultConfig().MembershipTTL = %v, want 5m", config.MembershipTTL)
	}
}

// ============ LoadConfig Tests ============

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear env vars
	t.Setenv("CACHE_ENABLED", "")
	t.Setenv("CACHE_TYPE", "")
	t.Setenv("CACHE_KEY_PREFIX", "")
	t.Setenv("REDIS_URL", "")

	config := LoadConfig()

	if config == nil {
		t.Fatal("LoadConfig() returned nil")
	}

	if config.Enabled {
		t.Error("LoadConfig().Enabled should be false by default")
	}
	if config.Type != "memory" {
		t.Errorf("LoadConfig().Type = %q, want 'memory'", config.Type)
	}
}

func TestLoadConfig_Enabled(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed", "True", true},
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"1", "1", false}, // Only "true" enables
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CACHE_ENABLED", tt.envVal)
			config := LoadConfig()
			if config.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

func TestLoadConfig_Type(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   string
	}{
		{"redis", "redis", "redis"},
		{"REDIS uppercase", "REDIS", "redis"},
		{"memory", "memory", "memory"},
		{"MEMORY uppercase", "MEMORY", "memory"},
		{"empty defaults to memory", "", "memory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CACHE_TYPE", tt.envVal)
			t.Setenv("REDIS_URL", "") // Don't auto-switch to redis
			config := LoadConfig()
			if config.Type != tt.want {
				t.Errorf("Type = %q, want %q", config.Type, tt.want)
			}
		})
	}
}

func TestLoadConfig_KeyPrefix(t *testing.T) {
	t.Run("custom prefix", func(t *testing.T) {
		t.Setenv("CACHE_KEY_PREFIX", "myapp")
		config := LoadConfig()
		if config.KeyPrefix != "myapp" {
			t.Errorf("KeyPrefix = %q, want 'myapp'", config.KeyPrefix)
		}
	})

	t.Run("empty defaults to app", func(t *testing.T) {
		t.Setenv("CACHE_KEY_PREFIX", "")
		config := LoadConfig()
		if config.KeyPrefix != "app" {
			t.Errorf("KeyPrefix = %q, want 'app'", config.KeyPrefix)
		}
	})
}

func TestLoadConfig_RedisURL(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	config := LoadConfig()

	if config.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("RedisURL = %q, want 'redis://localhost:6379/0'", config.RedisURL)
	}
	// Should auto-switch to redis
	if config.Type != "redis" {
		t.Errorf("Type = %q, want 'redis' when REDIS_URL is set", config.Type)
	}
}

func TestLoadConfig_RedisPoolSize(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid pool size", "20", 20},
		{"zero keeps default", "0", 10},
		{"negative keeps default", "-5", 10},
		{"invalid keeps default", "invalid", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_POOL_SIZE", tt.envVal)
			config := LoadConfig()
			if config.RedisPoolSize != tt.want {
				t.Errorf("RedisPoolSize = %d, want %d", config.RedisPoolSize, tt.want)
			}
		})
	}
}

func TestLoadConfig_RedisIdleConns(t *testing.T) {
	t.Run("min idle conns", func(t *testing.T) {
		t.Setenv("REDIS_MIN_IDLE_CONNS", "5")
		config := LoadConfig()
		if config.RedisMinIdleConns != 5 {
			t.Errorf("RedisMinIdleConns = %d, want 5", config.RedisMinIdleConns)
		}
	})

	t.Run("max idle conns", func(t *testing.T) {
		t.Setenv("REDIS_MAX_IDLE_CONNS", "10")
		config := LoadConfig()
		if config.RedisMaxIdleConns != 10 {
			t.Errorf("RedisMaxIdleConns = %d, want 10", config.RedisMaxIdleConns)
		}
	})

	t.Run("zero is valid for min", func(t *testing.T) {
		t.Setenv("REDIS_MIN_IDLE_CONNS", "0")
		config := LoadConfig()
		if config.RedisMinIdleConns != 0 {
			t.Errorf("RedisMinIdleConns = %d, want 0", config.RedisMinIdleConns)
		}
	})

	t.Run("negative keeps default", func(t *testing.T) {
		t.Setenv("REDIS_MIN_IDLE_CONNS", "-1")
		config := LoadConfig()
		if config.RedisMinIdleConns != 2 {
			t.Errorf("RedisMinIdleConns = %d, want 2 (default)", config.RedisMinIdleConns)
		}
	})
}

func TestLoadConfig_MemoryMaxSize(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid size", "5000", 5000},
		{"zero keeps default", "0", 10000},
		{"negative keeps default", "-100", 10000},
		{"invalid keeps default", "abc", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CACHE_MEMORY_MAX_SIZE", tt.envVal)
			config := LoadConfig()
			if config.MemoryMaxSize != tt.want {
				t.Errorf("MemoryMaxSize = %d, want %d", config.MemoryMaxSize, tt.want)
			}
		})
	}
}

func TestLoadConfig_TTLSettings(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		envVal string
		getter func(*Config) time.Duration
		want   time.Duration
	}{
		{
			name:   "default TTL",
			envVar: "CACHE_DEFAULT_TTL",
			envVal: "120",
			getter: func(c *Config) time.Duration { return c.DefaultTTL },
			want:   120 * time.Second,
		},
		{
			name:   "health check TTL",
			envVar: "CACHE_HEALTH_CHECK_TTL",
			envVal: "60",
			getter: func(c *Config) time.Duration { return c.HealthCheckTTL },
			want:   60 * time.Second,
		},
		{
			name:   "user profile TTL",
			envVar: "CACHE_USER_PROFILE_TTL",
			envVal: "180",
			getter: func(c *Config) time.Duration { return c.UserProfileTTL },
			want:   180 * time.Second,
		},
		{
			name:   "session TTL",
			envVar: "CACHE_SESSION_TTL",
			envVal: "900",
			getter: func(c *Config) time.Duration { return c.SessionTTL },
			want:   900 * time.Second,
		},
		{
			name:   "organization TTL",
			envVar: "CACHE_ORGANIZATION_TTL",
			envVal: "600",
			getter: func(c *Config) time.Duration { return c.OrganizationTTL },
			want:   600 * time.Second,
		},
		{
			name:   "membership TTL",
			envVar: "CACHE_MEMBERSHIP_TTL",
			envVal: "300",
			getter: func(c *Config) time.Duration { return c.MembershipTTL },
			want:   300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envVal)
			config := LoadConfig()
			if tt.getter(config) != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.getter(config), tt.want)
			}
		})
	}
}

func TestLoadConfig_InvalidTTL(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		envVal      string
		getter      func(*Config) time.Duration
		wantDefault time.Duration
	}{
		{
			name:        "invalid default TTL keeps default",
			envVar:      "CACHE_DEFAULT_TTL",
			envVal:      "invalid",
			getter:      func(c *Config) time.Duration { return c.DefaultTTL },
			wantDefault: 5 * time.Minute,
		},
		{
			name:        "zero default TTL keeps default",
			envVar:      "CACHE_DEFAULT_TTL",
			envVal:      "0",
			getter:      func(c *Config) time.Duration { return c.DefaultTTL },
			wantDefault: 5 * time.Minute,
		},
		{
			name:        "negative default TTL keeps default",
			envVar:      "CACHE_DEFAULT_TTL",
			envVal:      "-60",
			getter:      func(c *Config) time.Duration { return c.DefaultTTL },
			wantDefault: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envVal)
			config := LoadConfig()
			if tt.getter(config) != tt.wantDefault {
				t.Errorf("%s = %v, want %v (default)", tt.name, tt.getter(config), tt.wantDefault)
			}
		})
	}
}

// ============ Config Structure Tests ============

func TestConfig_Structure(t *testing.T) {
	config := Config{
		Enabled:               true,
		Type:                  "redis",
		KeyPrefix:             "test",
		RedisURL:              "redis://localhost:6379",
		RedisPoolSize:         25,
		RedisMinIdleConns:     5,
		RedisMaxIdleConns:     10,
		RedisConnMaxIdleTime:  10 * time.Minute,
		RedisConnMaxLifetime:  60 * time.Minute,
		MemoryMaxSize:         5000,
		MemoryCleanupInterval: 2 * time.Minute,
		DefaultTTL:            10 * time.Minute,
		HealthCheckTTL:        1 * time.Minute,
		UserProfileTTL:        5 * time.Minute,
		SessionTTL:            30 * time.Minute,
		OrganizationTTL:       10 * time.Minute,
		MembershipTTL:         10 * time.Minute,
	}

	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	if config.Type != "redis" {
		t.Errorf("Type = %q, want 'redis'", config.Type)
	}
	if config.RedisPoolSize != 25 {
		t.Errorf("RedisPoolSize = %d, want 25", config.RedisPoolSize)
	}
}
