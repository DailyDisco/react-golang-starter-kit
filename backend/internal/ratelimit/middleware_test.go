package ratelimit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"react-golang-starter/internal/auth"
)

// ============ getClientIP Tests ============

func TestGetClientIP_RemoteAddrOnly(t *testing.T) {
	config := &Config{
		TrustedProxies:       []string{},
		parsedTrustedProxies: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	ip := getClientIP(req, config)
	if ip != "192.168.1.100" {
		t.Errorf("getClientIP() = %q, want %q", ip, "192.168.1.100")
	}
}

func TestGetClientIP_RemoteAddrWithoutPort(t *testing.T) {
	config := &Config{
		TrustedProxies:       []string{},
		parsedTrustedProxies: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100" // No port

	ip := getClientIP(req, config)
	if ip != "192.168.1.100" {
		t.Errorf("getClientIP() = %q, want %q", ip, "192.168.1.100")
	}
}

func TestGetClientIP_TrustedProxyWithXFF(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.1"},
	}
	config.parseTrustedProxies()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.1")

	ip := getClientIP(req, config)
	// Should return the first non-trusted IP from X-Forwarded-For
	if ip != "203.0.113.50" {
		t.Errorf("getClientIP() = %q, want %q", ip, "203.0.113.50")
	}
}

func TestGetClientIP_TrustedProxyWithXRealIP(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.1"},
	}
	config.parseTrustedProxies()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.75")

	ip := getClientIP(req, config)
	if ip != "203.0.113.75" {
		t.Errorf("getClientIP() = %q, want %q", ip, "203.0.113.75")
	}
}

func TestGetClientIP_UntrustedProxyIgnoresHeaders(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.1"},
	}
	config.parseTrustedProxies()

	// Request NOT from a trusted proxy
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", "spoofed-ip")
	req.Header.Set("X-Real-IP", "spoofed-ip2")

	ip := getClientIP(req, config)
	// Should use RemoteAddr, NOT the spoofed headers
	if ip != "192.168.1.100" {
		t.Errorf("getClientIP() = %q, want %q (should ignore spoofed headers)", ip, "192.168.1.100")
	}
}

func TestGetClientIP_AllXFFAreTrusted(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.0/8"},
	}
	config.parseTrustedProxies()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	// All IPs in X-Forwarded-For are trusted
	req.Header.Set("X-Forwarded-For", "10.0.0.2, 10.0.0.3")

	ip := getClientIP(req, config)
	// Should fall back to RemoteAddr
	if ip != "10.0.0.1" {
		t.Errorf("getClientIP() = %q, want %q", ip, "10.0.0.1")
	}
}

func TestGetClientIP_XRealIPIsTrusted(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.0/8"},
	}
	config.parseTrustedProxies()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	// X-Real-IP is also a trusted proxy
	req.Header.Set("X-Real-IP", "10.0.0.99")

	ip := getClientIP(req, config)
	// Should fall back to RemoteAddr since X-Real-IP is trusted
	if ip != "10.0.0.1" {
		t.Errorf("getClientIP() = %q, want %q", ip, "10.0.0.1")
	}
}

// ============ setCORSErrorHeaders Tests ============

func TestSetCORSErrorHeaders_NoOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	setCORSErrorHeaders(rr, req)

	// No Origin header, so no CORS headers should be set
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set CORS headers when no Origin header present")
	}
}

func TestSetCORSErrorHeaders_AllowedOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,http://example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	setCORSErrorHeaders(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q",
			rr.Header().Get("Access-Control-Allow-Origin"), "http://localhost:3000")
	}
	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Should set Access-Control-Allow-Credentials to true")
	}
}

func TestSetCORSErrorHeaders_DisallowedOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	setCORSErrorHeaders(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set CORS headers for disallowed origin")
	}
}

// ============ createRateLimitHandler Tests ============

func TestCreateRateLimitHandler(t *testing.T) {
	handler := createRateLimitHandler(60)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	// Check status code
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}

	// Check content type
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rr.Header().Get("Content-Type"))
	}

	// Parse and check JSON response
	var response rateLimitResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != "RATE_LIMITED" {
		t.Errorf("response.Error = %q, want RATE_LIMITED", response.Error)
	}
	if response.Code != http.StatusTooManyRequests {
		t.Errorf("response.Code = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
	if response.RetryAfter != 60 {
		t.Errorf("response.RetryAfter = %d, want 60", response.RetryAfter)
	}
}

// ============ RateLimitError Tests ============

func TestRateLimitError_ErrorMethod(t *testing.T) {
	err := RateLimitError{
		Message:    "Test error message",
		RetryAfter: 30,
	}

	if err.Error() != "Test error message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "Test error message")
	}
}

// ============ NewAIRateLimitMiddleware Tests ============

func TestNewAIRateLimitMiddleware_Disabled(t *testing.T) {
	config := &Config{Enabled: false}

	middleware := NewAIRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestNewAIRateLimitMiddleware_Enabled(t *testing.T) {
	config := &Config{
		Enabled:             true,
		AIRequestsPerMinute: 2, // Low limit for testing
	}

	middleware := NewAIRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rr := httptest.NewRecorder()

		middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 (rate limited), got %d", rr.Code)
	}
}

func TestNewAIRateLimitMiddleware_WithUserContext(t *testing.T) {
	config := &Config{
		Enabled:             true,
		AIRequestsPerMinute: 1,
	}

	middleware := NewAIRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request with user 1 should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
	ctx1 := context.WithValue(req1.Context(), auth.UserIDContextKey, uint(1))
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second request with same user should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
	ctx2 := context.WithValue(req2.Context(), auth.UserIDContextKey, uint(1))
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}

	// Request with different user should succeed
	req3 := httptest.NewRequest(http.MethodGet, "/ai/test", nil)
	ctx3 := context.WithValue(req3.Context(), auth.UserIDContextKey, uint(2))
	req3 = req3.WithContext(ctx3)
	rr3 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("Third request (different user): Expected status 200, got %d", rr3.Code)
	}
}

// ============ NewUserRateLimitMiddleware Tests ============

func TestNewUserRateLimitMiddleware_Enabled_WithUser(t *testing.T) {
	config := &Config{
		Enabled:               true,
		UserRequestsPerMinute: 1,
	}

	middleware := NewUserRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request with user context
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx1 := context.WithValue(req1.Context(), auth.UserIDContextKey, uint(42))
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second request with same user should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx2 := context.WithValue(req2.Context(), auth.UserIDContextKey, uint(42))
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}
}

func TestNewUserRateLimitMiddleware_Enabled_FallbackToIP(t *testing.T) {
	config := &Config{
		Enabled:               true,
		UserRequestsPerMinute: 1,
	}

	middleware := NewUserRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Request without user context (falls back to IP)
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.RemoteAddr = "192.168.1.50:1234"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second request from same IP should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "192.168.1.50:1234"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}
}

// ============ NewAPIRateLimitMiddleware Tests ============

func TestNewAPIRateLimitMiddleware_Enabled_WithUser(t *testing.T) {
	config := &Config{
		Enabled:              true,
		APIRequestsPerMinute: 1,
	}

	middleware := NewAPIRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request with user context
	req1 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	ctx1 := context.WithValue(req1.Context(), auth.UserIDContextKey, uint(100))
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", rr1.Code)
	}

	// Second request with same user should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	ctx2 := context.WithValue(req2.Context(), auth.UserIDContextKey, uint(100))
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected status 429, got %d", rr2.Code)
	}
}

// ============ NewAuthRateLimitMiddleware Tests ============

func TestNewAuthRateLimitMiddleware_Enabled(t *testing.T) {
	config := &Config{
		Enabled:               true,
		AuthRequestsPerMinute: 2,
	}

	middleware := NewAuthRateLimitMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "203.0.113.1:5000"
		rr := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "203.0.113.1:5000"
	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 (rate limited), got %d", rr.Code)
	}
}

// ============ keyByTrustedIP Tests ============

func TestKeyByTrustedIP(t *testing.T) {
	config := &Config{
		TrustedProxies: []string{"10.0.0.1"},
	}
	config.parseTrustedProxies()

	keyFunc := keyByTrustedIP(config)

	// Test request from trusted proxy
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.100")

	key, err := keyFunc(req)
	if err != nil {
		t.Fatalf("keyFunc returned error: %v", err)
	}

	if key != "203.0.113.100" {
		t.Errorf("keyFunc() = %q, want %q", key, "203.0.113.100")
	}
}
