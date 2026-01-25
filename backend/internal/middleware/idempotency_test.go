package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============ IdempotencyConfig Tests ============

func TestLoadIdempotencyConfig_Defaults(t *testing.T) {
	t.Setenv("IDEMPOTENCY_ENABLED", "")
	t.Setenv("IDEMPOTENCY_TTL", "")
	t.Setenv("IDEMPOTENCY_HEADER", "")

	config := LoadIdempotencyConfig()

	if !config.Enabled {
		t.Error("Enabled should be true by default")
	}

	if config.TTL != 24*time.Hour {
		t.Errorf("TTL = %v, want 24h", config.TTL)
	}

	if config.HeaderName != "Idempotency-Key" {
		t.Errorf("HeaderName = %q, want 'Idempotency-Key'", config.HeaderName)
	}
}

func TestLoadIdempotencyConfig_Disabled(t *testing.T) {
	t.Setenv("IDEMPOTENCY_ENABLED", "false")

	config := LoadIdempotencyConfig()

	if config.Enabled {
		t.Error("Enabled should be false when set to 'false'")
	}
}

func TestLoadIdempotencyConfig_CustomTTL(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   time.Duration
	}{
		{"1 hour", "1h", 1 * time.Hour},
		{"30 minutes", "30m", 30 * time.Minute},
		{"invalid uses default", "invalid", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("IDEMPOTENCY_TTL", tt.envVal)
			config := LoadIdempotencyConfig()
			if config.TTL != tt.want {
				t.Errorf("TTL = %v, want %v", config.TTL, tt.want)
			}
		})
	}
}

func TestLoadIdempotencyConfig_CustomHeader(t *testing.T) {
	t.Setenv("IDEMPOTENCY_HEADER", "X-Idempotency-Key")

	config := LoadIdempotencyConfig()

	if config.HeaderName != "X-Idempotency-Key" {
		t.Errorf("HeaderName = %q, want 'X-Idempotency-Key'", config.HeaderName)
	}
}

// ============ getEnvOrDefault Tests ============

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback string
		want     string
	}{
		{"use env value", "TEST_KEY_1", "custom_value", "default", "custom_value"},
		{"use fallback when empty", "TEST_KEY_2", "", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(tt.key, tt.envVal)
			}
			got := getEnvOrDefault(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvOrDefault(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

// ============ buildIdempotencyCacheKey Tests ============

func TestBuildIdempotencyCacheKey(t *testing.T) {
	key := buildIdempotencyCacheKey("abc123", 42, "/api/v1/users")

	if key == "" {
		t.Error("buildIdempotencyCacheKey returned empty string")
	}

	// Should contain the idempotency key
	if !bytes.Contains([]byte(key), []byte("abc123")) {
		t.Error("cache key should contain idempotency key")
	}

	// Should contain path
	if !bytes.Contains([]byte(key), []byte("/api/v1/users")) {
		t.Error("cache key should contain path")
	}
}

func TestBuildIdempotencyCacheKey_DifferentUsers(t *testing.T) {
	key1 := buildIdempotencyCacheKey("abc123", 1, "/api/v1/users")
	key2 := buildIdempotencyCacheKey("abc123", 2, "/api/v1/users")

	if key1 == key2 {
		t.Error("cache keys should be different for different users")
	}
}

func TestBuildIdempotencyCacheKey_DifferentPaths(t *testing.T) {
	key1 := buildIdempotencyCacheKey("abc123", 1, "/api/v1/users")
	key2 := buildIdempotencyCacheKey("abc123", 1, "/api/v1/orders")

	if key1 == key2 {
		t.Error("cache keys should be different for different paths")
	}
}

// ============ hashRequest Tests ============

func TestHashRequest_WithBody(t *testing.T) {
	body := []byte(`{"name":"test","value":123}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))

	hash := hashRequest(req)

	if hash == "" {
		t.Error("hashRequest returned empty string")
	}
	if hash == "empty" {
		t.Error("hashRequest should not return 'empty' for request with body")
	}
	if hash == "error" {
		t.Error("hashRequest should not return 'error' for valid request")
	}

	// Body should still be readable after hashing
	readBody, _ := io.ReadAll(req.Body)
	if !bytes.Equal(readBody, body) {
		t.Error("request body should be restored after hashing")
	}
}

func TestHashRequest_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Body = nil // Explicitly set to nil (httptest creates http.NoBody)

	hash := hashRequest(req)

	if hash != "empty" {
		t.Errorf("hashRequest for nil body = %q, want 'empty'", hash)
	}
}

func TestHashRequest_EmptyBody(t *testing.T) {
	// httptest.NewRequest with nil creates http.NoBody, not nil
	// This tests the case where body exists but is empty
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte{}))

	hash := hashRequest(req)

	// SHA256 of empty bytes is a known value
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expectedHash {
		t.Errorf("hashRequest for empty body = %q, want %q", hash, expectedHash)
	}
}

func TestHashRequest_Deterministic(t *testing.T) {
	body := []byte(`{"name":"test"}`)

	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	hash1 := hashRequest(req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	hash2 := hashRequest(req2)

	if hash1 != hash2 {
		t.Error("same body should produce same hash")
	}
}

func TestHashRequest_DifferentBodies(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{"a":1}`)))
	hash1 := hashRequest(req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{"b":2}`)))
	hash2 := hashRequest(req2)

	if hash1 == hash2 {
		t.Error("different bodies should produce different hashes")
	}
}

// ============ extractCacheableHeaders Tests ============

func TestExtractCacheableHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("X-Request-Id", "req-123")
	headers.Set("Location", "/api/v1/users/42")
	headers.Set("X-Custom-Header", "should-not-be-cached")

	cached := extractCacheableHeaders(headers)

	if cached["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q, want 'application/json'", cached["Content-Type"])
	}

	if cached["X-Request-Id"] != "req-123" {
		t.Errorf("X-Request-Id = %q, want 'req-123'", cached["X-Request-Id"])
	}

	if cached["Location"] != "/api/v1/users/42" {
		t.Errorf("Location = %q, want '/api/v1/users/42'", cached["Location"])
	}

	if _, exists := cached["X-Custom-Header"]; exists {
		t.Error("X-Custom-Header should not be cached")
	}
}

func TestExtractCacheableHeaders_EmptyHeaders(t *testing.T) {
	headers := http.Header{}
	cached := extractCacheableHeaders(headers)

	if len(cached) != 0 {
		t.Errorf("cached headers length = %d, want 0", len(cached))
	}
}

// ============ ShouldApplyIdempotency Tests ============

func TestShouldApplyIdempotency(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Billing paths
		{"/api/v1/billing/checkout", true},
		{"/api/v1/billing/subscribe", true},

		// Payment paths
		{"/api/v1/payment/process", true},

		// User paths
		{"/api/v1/users", true},
		{"/api/v1/users/123", true},

		// Organization paths
		{"/api/v1/organizations", true},
		{"/api/v1/organizations/myorg/members", true},

		// File paths
		{"/api/v1/files/upload", true},
		{"/api/v1/files", true},

		// Paths that should NOT apply
		{"/api/v1/health", false},
		{"/api/v1/settings", false},
		{"/api/v1/auth/login", false},
		{"/metrics", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ShouldApplyIdempotency(tt.path)
			if got != tt.want {
				t.Errorf("ShouldApplyIdempotency(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ============ IdempotencyResponse Tests ============

func TestIdempotencyResponse_JSON(t *testing.T) {
	response := IdempotencyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Location":     "/api/v1/users/123",
		},
		Body:      []byte(`{"id":123,"name":"Test"}`),
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded IdempotencyResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", decoded.StatusCode, http.StatusCreated)
	}

	if decoded.Headers["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q, want 'application/json'", decoded.Headers["Content-Type"])
	}
}

// ============ IdempotencyConfig Structure Tests ============

func TestIdempotencyConfig_Structure(t *testing.T) {
	config := IdempotencyConfig{
		Enabled:    true,
		TTL:        12 * time.Hour,
		HeaderName: "X-Custom-Idempotency",
	}

	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	if config.TTL != 12*time.Hour {
		t.Errorf("TTL = %v, want 12h", config.TTL)
	}
	if config.HeaderName != "X-Custom-Idempotency" {
		t.Errorf("HeaderName = %q, want 'X-Custom-Idempotency'", config.HeaderName)
	}
}

// ============ responseRecorder Tests ============

func TestResponseRecorder_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}

	recorder.WriteHeader(http.StatusCreated)

	if recorder.statusCode != http.StatusCreated {
		t.Errorf("statusCode = %d, want %d", recorder.statusCode, http.StatusCreated)
	}

	// Second call should not change the status
	recorder.WriteHeader(http.StatusBadRequest)
	if recorder.statusCode != http.StatusCreated {
		t.Errorf("statusCode changed to %d, should remain %d", recorder.statusCode, http.StatusCreated)
	}
}

func TestResponseRecorder_Write(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}

	testBody := []byte("test response body")
	n, err := recorder.Write(testBody)

	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	if n != len(testBody) {
		t.Errorf("Write returned %d, want %d", n, len(testBody))
	}
	if recorder.body.String() != "test response body" {
		t.Errorf("body = %q, want 'test response body'", recorder.body.String())
	}
}

func TestResponseRecorder_Unwrap(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}

	unwrapped := recorder.Unwrap()
	if unwrapped != w {
		t.Error("Unwrap() should return the underlying ResponseWriter")
	}
}

// ============ IdempotencyMiddleware Tests ============

func TestIdempotencyMiddleware_DisabledPassesThrough(t *testing.T) {
	config := &IdempotencyConfig{
		Enabled:    false,
		TTL:        time.Hour,
		HeaderName: "Idempotency-Key",
	}

	handler := IdempotencyMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Idempotency-Key", "test-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestIdempotencyMiddleware_GETPassesThrough(t *testing.T) {
	config := &IdempotencyConfig{
		Enabled:    true,
		TTL:        time.Hour,
		HeaderName: "Idempotency-Key",
	}

	handler := IdempotencyMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Idempotency-Key", "test-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestIdempotencyMiddleware_NoKeyPassesThrough(t *testing.T) {
	config := &IdempotencyConfig{
		Enabled:    true,
		TTL:        time.Hour,
		HeaderName: "Idempotency-Key",
	}

	handler := IdempotencyMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	// No Idempotency-Key header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}
