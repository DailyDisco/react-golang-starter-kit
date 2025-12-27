package ratelimit

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	// General rate limiting settings
	Enabled bool

	// TrustedProxies is a list of IP addresses or CIDR ranges that are trusted
	// to provide the real client IP via X-Forwarded-For header.
	// If empty, X-Forwarded-For is ignored and only RemoteAddr is used.
	// Example: ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1"]
	TrustedProxies []string

	// parsedTrustedProxies holds the parsed CIDR networks for efficient checking
	parsedTrustedProxies []*net.IPNet

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

	// Load trusted proxies from environment (comma-separated)
	// Example: RATE_LIMIT_TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,127.0.0.1"
	if val := os.Getenv("RATE_LIMIT_TRUSTED_PROXIES"); val != "" {
		proxies := strings.Split(val, ",")
		for _, proxy := range proxies {
			proxy = strings.TrimSpace(proxy)
			if proxy != "" {
				config.TrustedProxies = append(config.TrustedProxies, proxy)
			}
		}
	}

	// Parse trusted proxies into CIDR networks
	config.parseTrustedProxies()

	return config
}

// parseTrustedProxies parses the TrustedProxies slice into net.IPNet for efficient checking
func (c *Config) parseTrustedProxies() {
	for _, proxy := range c.TrustedProxies {
		// Try parsing as CIDR first
		_, network, err := net.ParseCIDR(proxy)
		if err == nil {
			c.parsedTrustedProxies = append(c.parsedTrustedProxies, network)
			continue
		}

		// Try parsing as single IP
		ip := net.ParseIP(proxy)
		if ip != nil {
			// Convert single IP to /32 or /128 CIDR
			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}
			c.parsedTrustedProxies = append(c.parsedTrustedProxies, &net.IPNet{
				IP:   ip,
				Mask: mask,
			})
		}
	}
}

// IsTrustedProxy checks if the given IP is in the trusted proxy list
func (c *Config) IsTrustedProxy(ipStr string) bool {
	if len(c.parsedTrustedProxies) == 0 {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, network := range c.parsedTrustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
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
