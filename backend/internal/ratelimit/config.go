package ratelimit

import (
	"os"
	"strconv"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	// General rate limiting settings
	Enabled bool

	// IP-based rate limiting
	IPRequestsPerMinute int
	IPRequestsPerHour   int
	IPBurstSize         int

	// User-based rate limiting (for authenticated endpoints)
	UserRequestsPerMinute int
	UserRequestsPerHour   int
	UserBurstSize         int

	// Auth endpoints rate limiting (more restrictive)
	AuthRequestsPerMinute int
	AuthRequestsPerHour   int
	AuthBurstSize         int

	// API endpoints rate limiting
	APIRequestsPerMinute int
	APIRequestsPerHour   int
	APIBurstSize         int
}

// LoadConfig loads rate limiting configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Enabled: true, // Default to enabled

		// Default IP limits
		IPRequestsPerMinute: 60,
		IPRequestsPerHour:   1000,
		IPBurstSize:         10,

		// Default user limits (less restrictive for authenticated users)
		UserRequestsPerMinute: 120,
		UserRequestsPerHour:   2000,
		UserBurstSize:         20,

		// Default auth limits (more restrictive)
		AuthRequestsPerMinute: 5,
		AuthRequestsPerHour:   20,
		AuthBurstSize:         2,

		// Default API limits
		APIRequestsPerMinute: 100,
		APIRequestsPerHour:   1500,
		APIBurstSize:         15,
	}

	// Override with environment variables
	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		if enabled == "false" || enabled == "0" {
			config.Enabled = false
		}
	}

	// IP-based limits
	if val := os.Getenv("RATE_LIMIT_IP_PER_MINUTE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.IPRequestsPerMinute = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_IP_PER_HOUR"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.IPRequestsPerHour = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_IP_BURST_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.IPBurstSize = parsed
		}
	}

	// User-based limits
	if val := os.Getenv("RATE_LIMIT_USER_PER_MINUTE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.UserRequestsPerMinute = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_USER_PER_HOUR"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.UserRequestsPerHour = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_USER_BURST_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.UserBurstSize = parsed
		}
	}

	// Auth limits
	if val := os.Getenv("RATE_LIMIT_AUTH_PER_MINUTE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.AuthRequestsPerMinute = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_AUTH_PER_HOUR"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.AuthRequestsPerHour = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_AUTH_BURST_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.AuthBurstSize = parsed
		}
	}

	// API limits
	if val := os.Getenv("RATE_LIMIT_API_PER_MINUTE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.APIRequestsPerMinute = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_API_PER_HOUR"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.APIRequestsPerHour = parsed
		}
	}
	if val := os.Getenv("RATE_LIMIT_API_BURST_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			config.APIBurstSize = parsed
		}
	}

	return config
}

// GetIPWindow returns the rate limiting window duration for IP-based limits
func (c *Config) GetIPWindow() time.Duration {
	return time.Minute
}

// GetUserWindow returns the rate limiting window duration for user-based limits
func (c *Config) GetUserWindow() time.Duration {
	return time.Minute
}

// GetAuthWindow returns the rate limiting window duration for auth endpoints
func (c *Config) GetAuthWindow() time.Duration {
	return time.Minute
}

// GetAPIWindow returns the rate limiting window duration for API endpoints
func (c *Config) GetAPIWindow() time.Duration {
	return time.Minute
}
