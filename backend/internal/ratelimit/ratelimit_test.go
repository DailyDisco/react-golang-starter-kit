package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("RATE_LIMIT_ENABLED")
	os.Unsetenv("RATE_LIMIT_IP_PER_MINUTE")
	os.Unsetenv("RATE_LIMIT_AUTH_PER_MINUTE")

	config := LoadConfig()

	if !config.Enabled {
		t.Error("Expected rate limiting to be enabled by default")
	}

	if config.IPRequestsPerMinute != 60 {
		t.Errorf("Expected IPRequestsPerMinute to be 60, got %d", config.IPRequestsPerMinute)
	}

	if config.AuthRequestsPerMinute != 5 {
		t.Errorf("Expected AuthRequestsPerMinute to be 5, got %d", config.AuthRequestsPerMinute)
	}

	if config.UserRequestsPerMinute != 120 {
		t.Errorf("Expected UserRequestsPerMinute to be 120, got %d", config.UserRequestsPerMinute)
	}

	if config.APIRequestsPerMinute != 100 {
		t.Errorf("Expected APIRequestsPerMinute to be 100, got %d", config.APIRequestsPerMinute)
	}
}

func TestLoadConfig_DisabledViaEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"false string", "false"},
		{"zero string", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("RATE_LIMIT_ENABLED", tt.value)
			defer os.Unsetenv("RATE_LIMIT_ENABLED")

			config := LoadConfig()

			if config.Enabled {
				t.Error("Expected rate limiting to be disabled")
			}
		})
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_IP_PER_MINUTE", "30")
	os.Setenv("RATE_LIMIT_IP_PER_HOUR", "500")
	os.Setenv("RATE_LIMIT_AUTH_PER_MINUTE", "3")
	os.Setenv("RATE_LIMIT_USER_PER_MINUTE", "60")
	os.Setenv("RATE_LIMIT_API_PER_MINUTE", "50")
	defer func() {
		os.Unsetenv("RATE_LIMIT_IP_PER_MINUTE")
		os.Unsetenv("RATE_LIMIT_IP_PER_HOUR")
		os.Unsetenv("RATE_LIMIT_AUTH_PER_MINUTE")
		os.Unsetenv("RATE_LIMIT_USER_PER_MINUTE")
		os.Unsetenv("RATE_LIMIT_API_PER_MINUTE")
	}()

	config := LoadConfig()

	if config.IPRequestsPerMinute != 30 {
		t.Errorf("Expected IPRequestsPerMinute to be 30, got %d", config.IPRequestsPerMinute)
	}

	if config.IPRequestsPerHour != 500 {
		t.Errorf("Expected IPRequestsPerHour to be 500, got %d", config.IPRequestsPerHour)
	}

	if config.AuthRequestsPerMinute != 3 {
		t.Errorf("Expected AuthRequestsPerMinute to be 3, got %d", config.AuthRequestsPerMinute)
	}

	if config.UserRequestsPerMinute != 60 {
		t.Errorf("Expected UserRequestsPerMinute to be 60, got %d", config.UserRequestsPerMinute)
	}

	if config.APIRequestsPerMinute != 50 {
		t.Errorf("Expected APIRequestsPerMinute to be 50, got %d", config.APIRequestsPerMinute)
	}
}

func TestLoadConfig_InvalidValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_IP_PER_MINUTE", "invalid")
	defer os.Unsetenv("RATE_LIMIT_IP_PER_MINUTE")

	config := LoadConfig()

	// Should fall back to default value
	if config.IPRequestsPerMinute != 60 {
		t.Errorf("Expected IPRequestsPerMinute to be default 60 for invalid input, got %d", config.IPRequestsPerMinute)
	}
}

func TestLoadConfig_NegativeValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_IP_PER_MINUTE", "-5")
	defer os.Unsetenv("RATE_LIMIT_IP_PER_MINUTE")

	config := LoadConfig()

	// Should fall back to default value for negative input
	if config.IPRequestsPerMinute != 60 {
		t.Errorf("Expected IPRequestsPerMinute to be default 60 for negative input, got %d", config.IPRequestsPerMinute)
	}
}

func TestGetWindows(t *testing.T) {
	config := LoadConfig()

	if config.GetIPWindow() != time.Minute {
		t.Errorf("Expected IP window to be 1 minute, got %v", config.GetIPWindow())
	}

	if config.GetUserWindow() != time.Minute {
		t.Errorf("Expected User window to be 1 minute, got %v", config.GetUserWindow())
	}

	if config.GetAuthWindow() != time.Minute {
		t.Errorf("Expected Auth window to be 1 minute, got %v", config.GetAuthWindow())
	}

	if config.GetAPIWindow() != time.Minute {
		t.Errorf("Expected API window to be 1 minute, got %v", config.GetAPIWindow())
	}
}

func TestRateLimitError(t *testing.T) {
	err := RateLimitError{
		Message:    "Rate limit exceeded",
		RetryAfter: 30 * time.Second,
	}

	if err.Error() != "Rate limit exceeded" {
		t.Errorf("Expected error message 'Rate limit exceeded', got '%s'", err.Error())
	}

	if err.RetryAfter != 30*time.Second {
		t.Errorf("Expected RetryAfter to be 30s, got %v", err.RetryAfter)
	}
}

func TestNewIPRateLimitMiddleware_Disabled(t *testing.T) {
	config := &Config{Enabled: false}

	middleware := NewIPRateLimitMiddleware(config)

	// Handler should be a passthrough when disabled
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestNewIPRateLimitMiddleware_Enabled(t *testing.T) {
	config := &Config{
		Enabled:             true,
		IPRequestsPerMinute: 2, // Low limit for testing
	}

	middleware := NewIPRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rr := httptest.NewRecorder()

		middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 (rate limited), got %d", rr.Code)
	}
}

func TestNewAuthRateLimitMiddleware_Disabled(t *testing.T) {
	config := &Config{Enabled: false}

	middleware := NewAuthRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/auth/login", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestNewUserRateLimitMiddleware_Disabled(t *testing.T) {
	config := &Config{Enabled: false}

	middleware := NewUserRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestNewAPIRateLimitMiddleware_Disabled(t *testing.T) {
	config := &Config{Enabled: false}

	middleware := NewAPIRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/resource", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRateLimitByIP(t *testing.T) {
	config := &Config{
		Enabled:             true,
		IPRequestsPerMinute: 1,
	}
	middleware := NewIPRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "10.0.0.1:5000"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second request from same IP should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.1:5000"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}

	// Request from different IP should succeed
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "10.0.0.2:5000"
	rr3 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("Third request (different IP): Expected status 200, got %d", rr3.Code)
	}
}
