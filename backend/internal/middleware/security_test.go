package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// ============ DefaultSecurityConfig Tests ============

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if !config.Enabled {
		t.Error("DefaultSecurityConfig().Enabled = false, want true")
	}

	if config.ContentSecurityPolicy == "" {
		t.Error("DefaultSecurityConfig().ContentSecurityPolicy is empty")
	}

	if config.FrameOptions != "DENY" {
		t.Errorf("DefaultSecurityConfig().FrameOptions = %q, want %q", config.FrameOptions, "DENY")
	}

	if config.ContentTypeOptions != "nosniff" {
		t.Errorf("DefaultSecurityConfig().ContentTypeOptions = %q, want %q", config.ContentTypeOptions, "nosniff")
	}

	if config.XSSProtection != "1; mode=block" {
		t.Errorf("DefaultSecurityConfig().XSSProtection = %q, want %q", config.XSSProtection, "1; mode=block")
	}

	if config.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("DefaultSecurityConfig().ReferrerPolicy = %q, want %q", config.ReferrerPolicy, "strict-origin-when-cross-origin")
	}

	if config.HSTSMaxAge != 31536000 {
		t.Errorf("DefaultSecurityConfig().HSTSMaxAge = %d, want %d", config.HSTSMaxAge, 31536000)
	}
}

// ============ LoadSecurityConfig Tests ============

func TestLoadSecurityConfig_Defaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"SECURITY_HEADERS_ENABLED",
		"SECURITY_CSP",
		"SECURITY_FRAME_OPTIONS",
		"SECURITY_HSTS_ENABLED",
		"SECURITY_REFERRER_POLICY",
		"SECURITY_PERMISSIONS_POLICY",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	config := LoadSecurityConfig()

	if !config.Enabled {
		t.Error("LoadSecurityConfig() should be enabled by default")
	}

	if config.FrameOptions != "DENY" {
		t.Errorf("LoadSecurityConfig().FrameOptions = %q, want %q", config.FrameOptions, "DENY")
	}
}

func TestLoadSecurityConfig_Disabled(t *testing.T) {
	t.Setenv("SECURITY_HEADERS_ENABLED", "false")

	config := LoadSecurityConfig()

	if config.Enabled {
		t.Error("LoadSecurityConfig() should be disabled when SECURITY_HEADERS_ENABLED=false")
	}
}

func TestLoadSecurityConfig_CustomCSP(t *testing.T) {
	customCSP := "default-src 'self'; script-src 'self' 'unsafe-inline'"
	t.Setenv("SECURITY_CSP", customCSP)

	config := LoadSecurityConfig()

	if config.ContentSecurityPolicy != customCSP {
		t.Errorf("LoadSecurityConfig().ContentSecurityPolicy = %q, want %q", config.ContentSecurityPolicy, customCSP)
	}
}

func TestLoadSecurityConfig_CustomFrameOptions(t *testing.T) {
	t.Setenv("SECURITY_FRAME_OPTIONS", "SAMEORIGIN")

	config := LoadSecurityConfig()

	if config.FrameOptions != "SAMEORIGIN" {
		t.Errorf("LoadSecurityConfig().FrameOptions = %q, want %q", config.FrameOptions, "SAMEORIGIN")
	}
}

func TestLoadSecurityConfig_HSTSEnabled(t *testing.T) {
	t.Setenv("SECURITY_HSTS_ENABLED", "true")

	config := LoadSecurityConfig()

	if !config.HSTSEnabled {
		t.Error("LoadSecurityConfig().HSTSEnabled = false, want true")
	}
}

func TestLoadSecurityConfig_CustomReferrerPolicy(t *testing.T) {
	t.Setenv("SECURITY_REFERRER_POLICY", "no-referrer")

	config := LoadSecurityConfig()

	if config.ReferrerPolicy != "no-referrer" {
		t.Errorf("LoadSecurityConfig().ReferrerPolicy = %q, want %q", config.ReferrerPolicy, "no-referrer")
	}
}

func TestLoadSecurityConfig_CustomPermissionsPolicy(t *testing.T) {
	customPolicy := "geolocation=()"
	t.Setenv("SECURITY_PERMISSIONS_POLICY", customPolicy)

	config := LoadSecurityConfig()

	if config.PermissionsPolicy != customPolicy {
		t.Errorf("LoadSecurityConfig().PermissionsPolicy = %q, want %q", config.PermissionsPolicy, customPolicy)
	}
}

// ============ SecurityHeaders Middleware Tests ============

func TestSecurityHeaders_NilConfig(t *testing.T) {
	handler := SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should pass through without adding headers
	if rr.Code != http.StatusOK {
		t.Errorf("SecurityHeaders() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSecurityHeaders_Disabled(t *testing.T) {
	config := &SecurityConfig{
		Enabled:      false,
		FrameOptions: "DENY",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should NOT add headers when disabled
	if rr.Header().Get("X-Frame-Options") != "" {
		t.Error("SecurityHeaders() should not add headers when disabled")
	}
}

func TestSecurityHeaders_ContentTypeOptions(t *testing.T) {
	config := &SecurityConfig{
		Enabled:            true,
		ContentTypeOptions: "nosniff",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-Content-Type-Options")
	if got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
}

func TestSecurityHeaders_FrameOptions(t *testing.T) {
	config := &SecurityConfig{
		Enabled:      true,
		FrameOptions: "DENY",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-Frame-Options")
	if got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want %q", got, "DENY")
	}
}

func TestSecurityHeaders_XSSProtection(t *testing.T) {
	config := &SecurityConfig{
		Enabled:       true,
		XSSProtection: "1; mode=block",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-XSS-Protection")
	if got != "1; mode=block" {
		t.Errorf("X-XSS-Protection = %q, want %q", got, "1; mode=block")
	}
}

func TestSecurityHeaders_CSP(t *testing.T) {
	config := &SecurityConfig{
		Enabled:               true,
		ContentSecurityPolicy: "default-src 'self'",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Content-Security-Policy")
	if got != "default-src 'self'" {
		t.Errorf("Content-Security-Policy = %q, want %q", got, "default-src 'self'")
	}
}

func TestSecurityHeaders_ReferrerPolicy(t *testing.T) {
	config := &SecurityConfig{
		Enabled:        true,
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Referrer-Policy")
	if got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy = %q, want %q", got, "strict-origin-when-cross-origin")
	}
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	config := &SecurityConfig{
		Enabled:           true,
		PermissionsPolicy: "camera=(), microphone=()",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Permissions-Policy")
	if got != "camera=(), microphone=()" {
		t.Errorf("Permissions-Policy = %q, want %q", got, "camera=(), microphone=()")
	}
}

func TestSecurityHeaders_HSTS_WithHTTPS(t *testing.T) {
	config := &SecurityConfig{
		Enabled:               true,
		HSTSEnabled:           true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Simulate HTTPS via X-Forwarded-Proto header (common in reverse proxy setups)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Strict-Transport-Security")
	expected := "max-age=31536000; includeSubDomains"
	if got != expected {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, expected)
	}
}

func TestSecurityHeaders_HSTS_WithPreload(t *testing.T) {
	config := &SecurityConfig{
		Enabled:               true,
		HSTSEnabled:           true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Strict-Transport-Security")
	expected := "max-age=31536000; includeSubDomains; preload"
	if got != expected {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, expected)
	}
}

func TestSecurityHeaders_HSTS_NoHTTPS(t *testing.T) {
	config := &SecurityConfig{
		Enabled:     true,
		HSTSEnabled: true,
		HSTSMaxAge:  31536000,
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// HTTP request (no TLS, no X-Forwarded-Proto)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// HSTS should NOT be set for HTTP
	got := rr.Header().Get("Strict-Transport-Security")
	if got != "" {
		t.Errorf("Strict-Transport-Security should not be set for HTTP, got %q", got)
	}
}

func TestSecurityHeaders_CrossOriginPolicies(t *testing.T) {
	config := &SecurityConfig{
		Enabled:                   true,
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginResourcePolicy: "same-origin",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	tests := []struct {
		header string
		want   string
	}{
		{"Cross-Origin-Opener-Policy", "same-origin"},
		{"Cross-Origin-Embedder-Policy", "require-corp"},
		{"Cross-Origin-Resource-Policy", "same-origin"},
	}

	for _, tt := range tests {
		got := rr.Header().Get(tt.header)
		if got != tt.want {
			t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestSecurityHeaders_AllHeaders(t *testing.T) {
	config := DefaultSecurityConfig()

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check all expected headers are present
	expectedHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Content-Security-Policy",
		"Referrer-Policy",
		"Permissions-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Resource-Policy",
	}

	for _, header := range expectedHeaders {
		if rr.Header().Get(header) == "" {
			t.Errorf("Expected header %s to be set", header)
		}
	}
}

func TestSecurityHeaders_EmptyValues(t *testing.T) {
	config := &SecurityConfig{
		Enabled:      true,
		FrameOptions: "", // Explicitly empty
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Empty values should NOT set headers
	if rr.Header().Get("X-Frame-Options") != "" {
		t.Error("X-Frame-Options should not be set when config value is empty")
	}
}

// ============ Middleware Chain Tests ============

func TestSecurityHeaders_PassesToNextHandler(t *testing.T) {
	config := DefaultSecurityConfig()
	handlerCalled := false

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("SecurityHeaders middleware did not call next handler")
	}

	if rr.Code != http.StatusTeapot {
		t.Errorf("SecurityHeaders() status = %d, want %d", rr.Code, http.StatusTeapot)
	}
}
