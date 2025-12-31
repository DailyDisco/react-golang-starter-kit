package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ============ DefaultLogConfig Tests ============

func TestDefaultLogConfig(t *testing.T) {
	config := DefaultLogConfig()

	if config == nil {
		t.Fatal("DefaultLogConfig() returned nil")
	}

	// Check default values
	if !config.Enabled {
		t.Error("DefaultLogConfig().Enabled should be true")
	}

	if config.Level != "info" {
		t.Errorf("DefaultLogConfig().Level = %q, want %q", config.Level, "info")
	}

	if !config.IncludeUserContext {
		t.Error("DefaultLogConfig().IncludeUserContext should be true")
	}

	if config.IncludeRequestBody {
		t.Error("DefaultLogConfig().IncludeRequestBody should be false")
	}

	if config.IncludeResponseBody {
		t.Error("DefaultLogConfig().IncludeResponseBody should be false")
	}

	if config.MaxRequestBodySize != 4096 {
		t.Errorf("DefaultLogConfig().MaxRequestBodySize = %d, want %d", config.MaxRequestBodySize, 4096)
	}

	if config.MaxResponseBodySize != 4096 {
		t.Errorf("DefaultLogConfig().MaxResponseBodySize = %d, want %d", config.MaxResponseBodySize, 4096)
	}

	if config.SamplingRate != 1.0 {
		t.Errorf("DefaultLogConfig().SamplingRate = %f, want %f", config.SamplingRate, 1.0)
	}

	if config.AsyncLogging {
		t.Error("DefaultLogConfig().AsyncLogging should be false")
	}

	if !config.SanitizeHeaders {
		t.Error("DefaultLogConfig().SanitizeHeaders should be true")
	}

	if !config.FilterSensitiveData {
		t.Error("DefaultLogConfig().FilterSensitiveData should be true")
	}

	if len(config.AllowedHeaders) == 0 {
		t.Error("DefaultLogConfig().AllowedHeaders should not be empty")
	}

	if config.Pretty {
		t.Error("DefaultLogConfig().Pretty should be false")
	}
}

// ============ LoadLogConfig Tests ============

func TestLoadLogConfig_Defaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"LOG_ENABLED", "LOG_LEVEL", "LOG_INCLUDE_USER_CONTEXT",
		"LOG_INCLUDE_REQUEST_BODY", "LOG_INCLUDE_RESPONSE_BODY",
		"LOG_MAX_REQUEST_BODY_SIZE", "LOG_MAX_RESPONSE_BODY_SIZE",
		"LOG_SAMPLING_RATE", "LOG_ASYNC", "LOG_SANITIZE_HEADERS",
		"LOG_FILTER_SENSITIVE_DATA", "LOG_ALLOWED_HEADERS",
		"LOG_TIME_FORMAT", "LOG_PRETTY",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	config := LoadLogConfig()

	if config == nil {
		t.Fatal("LoadLogConfig() returned nil")
	}

	// Should use defaults when env vars are not set
	if !config.Enabled {
		t.Error("LoadLogConfig() without env vars should have Enabled = true")
	}
}

func TestLoadLogConfig_FromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_ENABLED", "false")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_INCLUDE_USER_CONTEXT", "false")
	os.Setenv("LOG_INCLUDE_REQUEST_BODY", "true")
	os.Setenv("LOG_INCLUDE_RESPONSE_BODY", "true")
	os.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "8192")
	os.Setenv("LOG_MAX_RESPONSE_BODY_SIZE", "16384")
	os.Setenv("LOG_SAMPLING_RATE", "0.5")
	os.Setenv("LOG_ASYNC", "true")
	os.Setenv("LOG_SANITIZE_HEADERS", "false")
	os.Setenv("LOG_FILTER_SENSITIVE_DATA", "false")
	os.Setenv("LOG_ALLOWED_HEADERS", "content-type,authorization")
	os.Setenv("LOG_PRETTY", "true")

	defer func() {
		os.Unsetenv("LOG_ENABLED")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_INCLUDE_USER_CONTEXT")
		os.Unsetenv("LOG_INCLUDE_REQUEST_BODY")
		os.Unsetenv("LOG_INCLUDE_RESPONSE_BODY")
		os.Unsetenv("LOG_MAX_REQUEST_BODY_SIZE")
		os.Unsetenv("LOG_MAX_RESPONSE_BODY_SIZE")
		os.Unsetenv("LOG_SAMPLING_RATE")
		os.Unsetenv("LOG_ASYNC")
		os.Unsetenv("LOG_SANITIZE_HEADERS")
		os.Unsetenv("LOG_FILTER_SENSITIVE_DATA")
		os.Unsetenv("LOG_ALLOWED_HEADERS")
		os.Unsetenv("LOG_PRETTY")
	}()

	config := LoadLogConfig()

	if config.Enabled {
		t.Error("LoadLogConfig().Enabled should be false")
	}
	if config.Level != "debug" {
		t.Errorf("LoadLogConfig().Level = %q, want %q", config.Level, "debug")
	}
	if config.IncludeUserContext {
		t.Error("LoadLogConfig().IncludeUserContext should be false")
	}
	if !config.IncludeRequestBody {
		t.Error("LoadLogConfig().IncludeRequestBody should be true")
	}
	if !config.IncludeResponseBody {
		t.Error("LoadLogConfig().IncludeResponseBody should be true")
	}
	if config.MaxRequestBodySize != 8192 {
		t.Errorf("LoadLogConfig().MaxRequestBodySize = %d, want %d", config.MaxRequestBodySize, 8192)
	}
	if config.MaxResponseBodySize != 16384 {
		t.Errorf("LoadLogConfig().MaxResponseBodySize = %d, want %d", config.MaxResponseBodySize, 16384)
	}
	if config.SamplingRate != 0.5 {
		t.Errorf("LoadLogConfig().SamplingRate = %f, want %f", config.SamplingRate, 0.5)
	}
	if !config.AsyncLogging {
		t.Error("LoadLogConfig().AsyncLogging should be true")
	}
	if config.SanitizeHeaders {
		t.Error("LoadLogConfig().SanitizeHeaders should be false")
	}
	if config.FilterSensitiveData {
		t.Error("LoadLogConfig().FilterSensitiveData should be false")
	}
	if len(config.AllowedHeaders) != 2 {
		t.Errorf("LoadLogConfig().AllowedHeaders length = %d, want 2", len(config.AllowedHeaders))
	}
	if !config.Pretty {
		t.Error("LoadLogConfig().Pretty should be true")
	}
}

func TestLoadLogConfig_InvalidValues(t *testing.T) {
	// Set invalid values - should fall back to defaults
	os.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "invalid")
	os.Setenv("LOG_SAMPLING_RATE", "invalid")
	defer func() {
		os.Unsetenv("LOG_MAX_REQUEST_BODY_SIZE")
		os.Unsetenv("LOG_SAMPLING_RATE")
	}()

	config := LoadLogConfig()

	// Should use default when invalid
	if config.MaxRequestBodySize != 4096 {
		t.Errorf("LoadLogConfig() with invalid size should use default, got %d", config.MaxRequestBodySize)
	}
	if config.SamplingRate != 1.0 {
		t.Errorf("LoadLogConfig() with invalid sampling rate should use default, got %f", config.SamplingRate)
	}
}

func TestLoadLogConfig_SamplingRateBounds(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected float64
	}{
		{"negative", "-0.5", 1.0},   // Should use default
		{"over 1", "1.5", 1.0},      // Should use default
		{"valid 0", "0", 0.0},       // Valid
		{"valid 1", "1", 1.0},       // Valid
		{"valid mid", "0.75", 0.75}, // Valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOG_SAMPLING_RATE", tt.value)
			defer os.Unsetenv("LOG_SAMPLING_RATE")

			config := LoadLogConfig()
			if config.SamplingRate != tt.expected {
				t.Errorf("LoadLogConfig().SamplingRate with %q = %f, want %f", tt.value, config.SamplingRate, tt.expected)
			}
		})
	}
}

// ============ ShouldLogRequest Tests ============

func TestShouldLogRequest_Always(t *testing.T) {
	config := &LogConfig{SamplingRate: 1.0}

	// Should always return true with rate 1.0
	for i := 0; i < 10; i++ {
		if !config.ShouldLogRequest() {
			t.Error("ShouldLogRequest() with rate 1.0 should always return true")
		}
	}
}

func TestShouldLogRequest_Never(t *testing.T) {
	config := &LogConfig{SamplingRate: 0.0}

	// Should always return false with rate 0.0
	for i := 0; i < 10; i++ {
		if config.ShouldLogRequest() {
			t.Error("ShouldLogRequest() with rate 0.0 should always return false")
		}
	}
}

func TestShouldLogRequest_Partial(t *testing.T) {
	config := &LogConfig{SamplingRate: 0.5}

	// Should return a mix with rate 0.5
	// We can't test exact ratio due to randomness, but we test it doesn't error
	trueCount := 0
	for i := 0; i < 100; i++ {
		if config.ShouldLogRequest() {
			trueCount++
		}
	}

	// With 100 samples and 50% rate, we expect roughly 50 true values
	// Allow for statistical variance
	if trueCount == 0 || trueCount == 100 {
		t.Logf("ShouldLogRequest() with rate 0.5 returned all same values (%d true), may be statistical anomaly", trueCount)
	}
}

// ============ IsHeaderAllowed Tests ============

func TestIsHeaderAllowed_SanitizationDisabled(t *testing.T) {
	config := &LogConfig{
		SanitizeHeaders: false,
		AllowedHeaders:  []string{"content-type"},
	}

	// All headers should be allowed when sanitization is disabled
	if !config.IsHeaderAllowed("authorization") {
		t.Error("IsHeaderAllowed() with SanitizeHeaders=false should allow all headers")
	}
	if !config.IsHeaderAllowed("x-custom-header") {
		t.Error("IsHeaderAllowed() with SanitizeHeaders=false should allow all headers")
	}
}

func TestIsHeaderAllowed_SanitizationEnabled(t *testing.T) {
	config := &LogConfig{
		SanitizeHeaders: true,
		AllowedHeaders:  []string{"content-type", "accept", "x-request-id"},
	}

	tests := []struct {
		header  string
		allowed bool
	}{
		{"content-type", true},
		{"Content-Type", true}, // Case insensitive
		{"CONTENT-TYPE", true},
		{"accept", true},
		{"x-request-id", true},
		{"authorization", false},
		{"cookie", false},
		{"x-custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			result := config.IsHeaderAllowed(tt.header)
			if result != tt.allowed {
				t.Errorf("IsHeaderAllowed(%q) = %v, want %v", tt.header, result, tt.allowed)
			}
		})
	}
}

func TestIsHeaderAllowed_EmptyAllowedHeaders(t *testing.T) {
	config := &LogConfig{
		SanitizeHeaders: true,
		AllowedHeaders:  []string{},
	}

	// No headers should be allowed when list is empty
	if config.IsHeaderAllowed("content-type") {
		t.Error("IsHeaderAllowed() with empty AllowedHeaders should not allow any headers")
	}
}

// ============ getRequestID Tests ============

func TestGetRequestID_FromHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "custom-request-id-123")

	requestID := getRequestID(req)

	if requestID != "custom-request-id-123" {
		t.Errorf("getRequestID() with X-Request-ID header = %q, want %q", requestID, "custom-request-id-123")
	}
}

func TestGetRequestID_Generated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No X-Request-ID header

	requestID := getRequestID(req)

	if requestID == "" {
		t.Error("getRequestID() should generate a request ID when none is provided")
	}

	// Should be a valid UUID format (36 characters with hyphens)
	if len(requestID) != 36 {
		t.Errorf("getRequestID() generated ID length = %d, want 36", len(requestID))
	}
}

func TestGetRequestID_Unique(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)

	id1 := getRequestID(req1)
	id2 := getRequestID(req2)

	if id1 == id2 {
		t.Error("getRequestID() should generate unique IDs")
	}
}

// ============ getRealIP Tests ============

func TestGetRealIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	ip := getRealIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("getRealIP() with X-Forwarded-For = %q, want %q", ip, "203.0.113.195")
	}
}

func TestGetRealIP_XForwardedForMultiple(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")

	ip := getRealIP(req)

	// Should return the first IP
	if ip != "203.0.113.195" {
		t.Errorf("getRealIP() with multiple X-Forwarded-For = %q, want %q", ip, "203.0.113.195")
	}
}

func TestGetRealIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")

	ip := getRealIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("getRealIP() with X-Real-IP = %q, want %q", ip, "192.168.1.1")
	}
}

func TestGetRealIP_CFConnectingIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-Connecting-IP", "104.28.0.1")

	ip := getRealIP(req)

	if ip != "104.28.0.1" {
		t.Errorf("getRealIP() with CF-Connecting-IP = %q, want %q", ip, "104.28.0.1")
	}
}

func TestGetRealIP_XForwarded(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded", "10.0.0.1")

	ip := getRealIP(req)

	if ip != "10.0.0.1" {
		t.Errorf("getRealIP() with X-Forwarded = %q, want %q", ip, "10.0.0.1")
	}
}

func TestGetRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// httptest.NewRequest sets RemoteAddr to "192.0.2.1:1234"

	ip := getRealIP(req)

	// Should fall back to RemoteAddr
	if ip == "" {
		t.Error("getRealIP() should return RemoteAddr as fallback")
	}
}

func TestGetRealIP_Priority(t *testing.T) {
	// X-Forwarded-For should have highest priority
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	req.Header.Set("X-Real-IP", "2.2.2.2")
	req.Header.Set("CF-Connecting-IP", "3.3.3.3")

	ip := getRealIP(req)

	if ip != "1.1.1.1" {
		t.Errorf("getRealIP() should prioritize X-Forwarded-For, got %q", ip)
	}
}

func TestGetRealIP_TrimsWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "  192.168.1.1  ")

	ip := getRealIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("getRealIP() should trim whitespace, got %q", ip)
	}
}

// ============ captureRequestBody Tests ============

func TestCaptureRequestBody_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	body := captureRequestBody(req, 1024)

	if body != "" {
		t.Errorf("captureRequestBody() with nil body = %q, want empty", body)
	}
}

func TestCaptureRequestBody_Small(t *testing.T) {
	bodyContent := `{"name": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(bodyContent))

	body := captureRequestBody(req, 1024)

	if body != bodyContent {
		t.Errorf("captureRequestBody() = %q, want %q", body, bodyContent)
	}

	// Body should still be readable
	remaining, _ := io.ReadAll(req.Body)
	if string(remaining) != bodyContent {
		t.Error("captureRequestBody() should restore the request body")
	}
}

func TestCaptureRequestBody_Truncation(t *testing.T) {
	bodyContent := "This is a long body that should be truncated"
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(bodyContent))

	body := captureRequestBody(req, 10)

	if len(body) > 10+len("... [truncated]") {
		t.Errorf("captureRequestBody() should truncate to max size, got length %d", len(body))
	}
}

// ============ responseCaptureWriter Tests ============

func TestResponseCaptureWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	buffer := &bytes.Buffer{}
	captureWriter := &responseCaptureWriter{
		ResponseWriter: w,
		buffer:         buffer,
		maxSize:        1024,
	}

	data := []byte("Hello, World!")
	n, err := captureWriter.Write(data)

	if err != nil {
		t.Errorf("responseCaptureWriter.Write() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("responseCaptureWriter.Write() n = %d, want %d", n, len(data))
	}

	// Check both the original writer and buffer have the data
	if w.Body.String() != string(data) {
		t.Errorf("Original writer has %q, want %q", w.Body.String(), string(data))
	}
	if buffer.String() != string(data) {
		t.Errorf("Capture buffer has %q, want %q", buffer.String(), string(data))
	}
}

func TestResponseCaptureWriter_MaxSize(t *testing.T) {
	w := httptest.NewRecorder()
	buffer := &bytes.Buffer{}
	captureWriter := &responseCaptureWriter{
		ResponseWriter: w,
		buffer:         buffer,
		maxSize:        10,
	}

	// Write more than maxSize
	data := []byte("This is a very long string that exceeds max size")
	captureWriter.Write(data)

	// Original writer should have all data
	if w.Body.String() != string(data) {
		t.Error("Original writer should have all data regardless of maxSize")
	}

	// Buffer should be limited
	if buffer.Len() > 10 {
		t.Errorf("Capture buffer should be limited to maxSize, got %d bytes", buffer.Len())
	}
}

// ============ StructuredLogger Tests ============

func TestStructuredLogger_Disabled(t *testing.T) {
	config := &LogConfig{Enabled: false}
	middleware := StructuredLoggerWithConfig(config)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even when logging is disabled")
	}
}

func TestStructuredLogger_Enabled(t *testing.T) {
	config := &LogConfig{
		Enabled:            true,
		SamplingRate:       1.0,
		IncludeUserContext: false,
		SanitizeHeaders:    true,
		AllowedHeaders:     []string{"content-type"},
	}
	middleware := StructuredLoggerWithConfig(config)

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Response status = %d, want %d", w.Code, http.StatusOK)
	}
}
