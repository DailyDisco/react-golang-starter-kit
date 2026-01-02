package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"react-golang-starter/internal/config"
)

// ============ config.GetAllowedOrigins Tests ============

func TestGetAllowedOrigins_Default(t *testing.T) {
	// Clear env var to get defaults
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := config.GetAllowedOrigins()

	if len(origins) == 0 {
		t.Error("GetAllowedOrigins() should return default origins when CORS_ALLOWED_ORIGINS is not set")
	}

	// Check that default includes localhost:3000
	found := false
	for _, origin := range origins {
		if origin == "http://localhost:3000" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetAllowedOrigins() default should include http://localhost:3000")
	}
}

func TestGetAllowedOrigins_FromEnv(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := config.GetAllowedOrigins()

	if len(origins) != 2 {
		t.Errorf("GetAllowedOrigins() with env var returned %d origins, want 2", len(origins))
	}

	expected := map[string]bool{
		"https://example.com":     true,
		"https://api.example.com": true,
	}

	for _, origin := range origins {
		if !expected[origin] {
			t.Errorf("GetAllowedOrigins() unexpected origin: %s", origin)
		}
	}
}

func TestGetAllowedOrigins_SingleOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://myapp.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := config.GetAllowedOrigins()

	if len(origins) != 1 {
		t.Errorf("GetAllowedOrigins() returned %d origins, want 1", len(origins))
	}

	if origins[0] != "https://myapp.com" {
		t.Errorf("GetAllowedOrigins() = %v, want ['https://myapp.com']", origins)
	}
}

// ============ SetCORSErrorHeaders Tests ============

func TestSetCORSErrorHeaders_NoOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No Origin header set
	w := httptest.NewRecorder()

	SetCORSErrorHeaders(w, req)

	// Should not set any CORS headers when no origin is provided
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("SetCORSErrorHeaders() should not set CORS headers without Origin header")
	}
}

func TestSetCORSErrorHeaders_AllowedOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	SetCORSErrorHeaders(w, req)

	// Should set CORS headers for allowed origin
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("SetCORSErrorHeaders() Access-Control-Allow-Origin = %q, want %q", allowOrigin, "http://localhost:3000")
	}

	allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")
	if allowCredentials != "true" {
		t.Errorf("SetCORSErrorHeaders() Access-Control-Allow-Credentials = %q, want %q", allowCredentials, "true")
	}
}

func TestSetCORSErrorHeaders_DisallowedOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "https://malicious-site.com")
	w := httptest.NewRecorder()

	SetCORSErrorHeaders(w, req)

	// Should not set CORS headers for disallowed origin
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("SetCORSErrorHeaders() should not set CORS headers for disallowed origin")
	}
}

func TestSetCORSErrorHeaders_CustomAllowedOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://myapp.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "https://myapp.com")
	w := httptest.NewRecorder()

	SetCORSErrorHeaders(w, req)

	// Should set CORS headers for custom allowed origin
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "https://myapp.com" {
		t.Errorf("SetCORSErrorHeaders() Access-Control-Allow-Origin = %q, want %q", allowOrigin, "https://myapp.com")
	}
}

func TestSetCORSErrorHeaders_DefaultOriginsNotAllowed(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://production.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000") // Default origin, but not in custom list
	w := httptest.NewRecorder()

	SetCORSErrorHeaders(w, req)

	// Should not allow localhost when custom origins are set
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("SetCORSErrorHeaders() should not allow default origins when custom CORS_ALLOWED_ORIGINS is set")
	}
}

// ============ Multiple Origins Tests ============

func TestSetCORSErrorHeaders_MultipleAllowedOrigins(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://app1.com,https://app2.com,https://app3.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	tests := []struct {
		name    string
		origin  string
		allowed bool
	}{
		{"first origin", "https://app1.com", true},
		{"middle origin", "https://app2.com", true},
		{"last origin", "https://app3.com", true},
		{"not in list", "https://app4.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			SetCORSErrorHeaders(w, req)

			hasHeader := w.Header().Get("Access-Control-Allow-Origin") != ""
			if hasHeader != tt.allowed {
				t.Errorf("SetCORSErrorHeaders() for origin %q: hasHeader = %v, want %v", tt.origin, hasHeader, tt.allowed)
			}
		})
	}
}
