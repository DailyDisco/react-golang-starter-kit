package ratelimit

import (
	"testing"
	"time"
)

// ============ LoadConfig Tests ============

func TestLoadConfig_AllDefaults(t *testing.T) {
	// Clear env vars
	t.Setenv("RATE_LIMIT_ENABLED", "")
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_IP_PER_HOUR", "")

	config := LoadConfig()

	if config == nil {
		t.Fatal("LoadConfig() returned nil")
	}

	// Check defaults
	if !config.Enabled {
		t.Error("Enabled should default to true")
	}
	if config.IPRequestsPerMinute != 60 {
		t.Errorf("IPRequestsPerMinute = %d, want 60", config.IPRequestsPerMinute)
	}
	if config.IPRequestsPerHour != 1000 {
		t.Errorf("IPRequestsPerHour = %d, want 1000", config.IPRequestsPerHour)
	}
	if config.IPBurstSize != 10 {
		t.Errorf("IPBurstSize = %d, want 10", config.IPBurstSize)
	}
}

func TestLoadConfig_UserDefaults(t *testing.T) {
	t.Setenv("RATE_LIMIT_USER_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_USER_PER_HOUR", "")
	t.Setenv("RATE_LIMIT_USER_BURST_SIZE", "")

	config := LoadConfig()

	if config.UserRequestsPerMinute != 120 {
		t.Errorf("UserRequestsPerMinute = %d, want 120", config.UserRequestsPerMinute)
	}
	if config.UserRequestsPerHour != 2000 {
		t.Errorf("UserRequestsPerHour = %d, want 2000", config.UserRequestsPerHour)
	}
	if config.UserBurstSize != 20 {
		t.Errorf("UserBurstSize = %d, want 20", config.UserBurstSize)
	}
}

func TestLoadConfig_AuthDefaults(t *testing.T) {
	t.Setenv("RATE_LIMIT_AUTH_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_AUTH_PER_HOUR", "")
	t.Setenv("RATE_LIMIT_AUTH_BURST_SIZE", "")

	config := LoadConfig()

	if config.AuthRequestsPerMinute != 20 {
		t.Errorf("AuthRequestsPerMinute = %d, want 20", config.AuthRequestsPerMinute)
	}
	if config.AuthRequestsPerHour != 100 {
		t.Errorf("AuthRequestsPerHour = %d, want 100", config.AuthRequestsPerHour)
	}
	if config.AuthBurstSize != 5 {
		t.Errorf("AuthBurstSize = %d, want 5", config.AuthBurstSize)
	}
}

func TestLoadConfig_APIDefaults(t *testing.T) {
	t.Setenv("RATE_LIMIT_API_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_API_PER_HOUR", "")
	t.Setenv("RATE_LIMIT_API_BURST_SIZE", "")

	config := LoadConfig()

	if config.APIRequestsPerMinute != 100 {
		t.Errorf("APIRequestsPerMinute = %d, want 100", config.APIRequestsPerMinute)
	}
	if config.APIRequestsPerHour != 1500 {
		t.Errorf("APIRequestsPerHour = %d, want 1500", config.APIRequestsPerHour)
	}
	if config.APIBurstSize != 15 {
		t.Errorf("APIBurstSize = %d, want 15", config.APIBurstSize)
	}
}

func TestLoadConfig_AIDefaults(t *testing.T) {
	t.Setenv("RATE_LIMIT_AI_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_AI_PER_HOUR", "")
	t.Setenv("RATE_LIMIT_AI_BURST_SIZE", "")

	config := LoadConfig()

	if config.AIRequestsPerMinute != 20 {
		t.Errorf("AIRequestsPerMinute = %d, want 20", config.AIRequestsPerMinute)
	}
	if config.AIRequestsPerHour != 200 {
		t.Errorf("AIRequestsPerHour = %d, want 200", config.AIRequestsPerHour)
	}
	if config.AIBurstSize != 5 {
		t.Errorf("AIBurstSize = %d, want 5", config.AIBurstSize)
	}
}

func TestLoadConfig_EnabledFalse(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"false", "false", false},
		{"0", "0", false},
		{"true", "true", true},
		{"1", "1", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RATE_LIMIT_ENABLED", tt.envVal)
			config := LoadConfig()
			if config.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

func TestLoadConfig_IPLimitsFromEnv(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "100")
	t.Setenv("RATE_LIMIT_IP_PER_HOUR", "2000")
	t.Setenv("RATE_LIMIT_IP_BURST_SIZE", "20")

	config := LoadConfig()

	if config.IPRequestsPerMinute != 100 {
		t.Errorf("IPRequestsPerMinute = %d, want 100", config.IPRequestsPerMinute)
	}
	if config.IPRequestsPerHour != 2000 {
		t.Errorf("IPRequestsPerHour = %d, want 2000", config.IPRequestsPerHour)
	}
	if config.IPBurstSize != 20 {
		t.Errorf("IPBurstSize = %d, want 20", config.IPBurstSize)
	}
}

func TestLoadConfig_InvalidValuesExtended(t *testing.T) {
	tests := []struct {
		name       string
		envVar     string
		envVal     string
		checkField func(c *Config) int
		wantVal    int
	}{
		{
			name:       "invalid IP per minute",
			envVar:     "RATE_LIMIT_IP_PER_MINUTE",
			envVal:     "invalid",
			checkField: func(c *Config) int { return c.IPRequestsPerMinute },
			wantVal:    60, // default
		},
		{
			name:       "negative IP per minute",
			envVar:     "RATE_LIMIT_IP_PER_MINUTE",
			envVal:     "-10",
			checkField: func(c *Config) int { return c.IPRequestsPerMinute },
			wantVal:    60, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envVal)
			config := LoadConfig()
			if tt.checkField(config) != tt.wantVal {
				t.Errorf("%s: got %d, want %d", tt.name, tt.checkField(config), tt.wantVal)
			}
		})
	}
}

func TestLoadConfig_TrustedProxies(t *testing.T) {
	t.Setenv("RATE_LIMIT_TRUSTED_PROXIES", "10.0.0.0/8,172.16.0.0/12,127.0.0.1")

	config := LoadConfig()

	if len(config.TrustedProxies) != 3 {
		t.Errorf("TrustedProxies length = %d, want 3", len(config.TrustedProxies))
	}
}

func TestLoadConfig_TrustedProxies_Empty(t *testing.T) {
	t.Setenv("RATE_LIMIT_TRUSTED_PROXIES", "")

	config := LoadConfig()

	if len(config.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies length = %d, want 0", len(config.TrustedProxies))
	}
}

// ============ IsTrustedProxy Tests ============

func TestIsTrustedProxy_CIDR(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.0/8"},
	}
	config.parseTrustedProxies()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"in range start", "10.0.0.1", true},
		{"in range middle", "10.128.0.1", true},
		{"in range end", "10.255.255.254", true},
		{"out of range", "11.0.0.1", false},
		{"different subnet", "192.168.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if config.IsTrustedProxy(tt.ip) != tt.expected {
				t.Errorf("IsTrustedProxy(%q) = %v, want %v", tt.ip, !tt.expected, tt.expected)
			}
		})
	}
}

func TestIsTrustedProxy_SingleIP(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"127.0.0.1"},
	}
	config.parseTrustedProxies()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"exact match", "127.0.0.1", true},
		{"different IP", "127.0.0.2", false},
		{"other localhost", "192.168.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if config.IsTrustedProxy(tt.ip) != tt.expected {
				t.Errorf("IsTrustedProxy(%q) = %v, want %v", tt.ip, !tt.expected, tt.expected)
			}
		})
	}
}

func TestIsTrustedProxy_Empty(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{},
	}
	config.parseTrustedProxies()

	if config.IsTrustedProxy("10.0.0.1") {
		t.Error("IsTrustedProxy should return false for empty trusted proxies list")
	}
}

func TestIsTrustedProxy_InvalidIP(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.0/8"},
	}
	config.parseTrustedProxies()

	if config.IsTrustedProxy("invalid-ip") {
		t.Error("IsTrustedProxy should return false for invalid IP")
	}
}

func TestIsTrustedProxy_IPv6(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"::1"},
	}
	config.parseTrustedProxies()

	if !config.IsTrustedProxy("::1") {
		t.Error("IsTrustedProxy should return true for IPv6 localhost")
	}
}

// ============ Window Duration Tests ============

func TestGetIPWindow(t *testing.T) {
	config := &Config{}
	if config.GetIPWindow() != time.Minute {
		t.Errorf("GetIPWindow() = %v, want %v", config.GetIPWindow(), time.Minute)
	}
}

func TestGetUserWindow(t *testing.T) {
	config := &Config{}
	if config.GetUserWindow() != time.Minute {
		t.Errorf("GetUserWindow() = %v, want %v", config.GetUserWindow(), time.Minute)
	}
}

func TestGetAuthWindow(t *testing.T) {
	config := &Config{}
	if config.GetAuthWindow() != time.Minute {
		t.Errorf("GetAuthWindow() = %v, want %v", config.GetAuthWindow(), time.Minute)
	}
}

func TestGetAPIWindow(t *testing.T) {
	config := &Config{}
	if config.GetAPIWindow() != time.Minute {
		t.Errorf("GetAPIWindow() = %v, want %v", config.GetAPIWindow(), time.Minute)
	}
}

func TestGetAIWindow(t *testing.T) {
	config := &Config{}
	if config.GetAIWindow() != time.Minute {
		t.Errorf("GetAIWindow() = %v, want %v", config.GetAIWindow(), time.Minute)
	}
}

// ============ Config Structure Tests ============

func TestConfig_Structure(t *testing.T) {
	config := Config{
		Enabled:               true,
		TrustedProxies:        []string{"10.0.0.0/8"},
		IPRequestsPerMinute:   100,
		IPRequestsPerHour:     1000,
		IPBurstSize:           10,
		UserRequestsPerMinute: 200,
		UserRequestsPerHour:   2000,
		UserBurstSize:         20,
		AuthRequestsPerMinute: 30,
		AuthRequestsPerHour:   150,
		AuthBurstSize:         5,
		APIRequestsPerMinute:  150,
		APIRequestsPerHour:    1500,
		APIBurstSize:          15,
		AIRequestsPerMinute:   25,
		AIRequestsPerHour:     250,
		AIBurstSize:           5,
	}

	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	if config.IPRequestsPerMinute != 100 {
		t.Errorf("IPRequestsPerMinute = %d, want 100", config.IPRequestsPerMinute)
	}
	if len(config.TrustedProxies) != 1 {
		t.Errorf("TrustedProxies length = %d, want 1", len(config.TrustedProxies))
	}
}

// ============ parseTrustedProxies Tests ============

func TestParseTrustedProxies_Mixed(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1", "192.168.1.0/24", "::1"},
	}
	config.parseTrustedProxies()

	// Should have 4 parsed networks
	if len(config.parsedTrustedProxies) != 4 {
		t.Errorf("parsedTrustedProxies length = %d, want 4", len(config.parsedTrustedProxies))
	}
}

func TestParseTrustedProxies_InvalidEntry(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"invalid-entry", "10.0.0.0/8"},
	}
	config.parseTrustedProxies()

	// Should only parse valid entries
	if len(config.parsedTrustedProxies) != 1 {
		t.Errorf("parsedTrustedProxies length = %d, want 1 (only valid CIDR)", len(config.parsedTrustedProxies))
	}
}
