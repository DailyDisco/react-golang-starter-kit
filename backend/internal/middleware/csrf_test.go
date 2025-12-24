package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()

	if !config.Enabled {
		t.Error("Expected CSRF to be enabled by default")
	}

	if config.TokenLength != 32 {
		t.Errorf("Expected token length 32, got %d", config.TokenLength)
	}

	if config.TokenName != "csrf_token" {
		t.Errorf("Expected token name 'csrf_token', got '%s'", config.TokenName)
	}

	if config.CookieHTTPOnly {
		t.Error("CookieHTTPOnly should be false for double-submit pattern")
	}

	if len(config.SafeMethods) == 0 {
		t.Error("Expected safe methods to be defined")
	}

	if len(config.ExemptPaths) == 0 {
		t.Error("Expected exempt paths to be defined")
	}
}

func TestIsSafeMethod(t *testing.T) {
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}

	tests := []struct {
		method   string
		expected bool
	}{
		{"GET", true},
		{"get", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"TRACE", true},
		{"POST", false},
		{"PUT", false},
		{"DELETE", false},
		{"PATCH", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := isSafeMethod(tt.method, safeMethods)
			if result != tt.expected {
				t.Errorf("isSafeMethod(%s) = %v, want %v", tt.method, result, tt.expected)
			}
		})
	}
}

func TestIsExemptPath(t *testing.T) {
	exemptPaths := []string{"/api/webhooks/", "/health"}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/webhooks/stripe", true},
		{"/api/webhooks/", true},
		{"/health", true},
		{"/api/users", false},
		{"/api/auth/login", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isExemptPath(tt.path, exemptPaths)
			if result != tt.expected {
				t.Errorf("isExemptPath(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGenerateCSRFToken(t *testing.T) {
	token1, err := generateCSRFToken(32)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if len(token1) == 0 {
		t.Error("Expected non-empty token")
	}

	// Generate another token and verify they're different
	token2, err := generateCSRFToken(32)
	if err != nil {
		t.Fatalf("Failed to generate second token: %v", err)
	}

	if token1 == token2 {
		t.Error("Expected tokens to be unique")
	}
}

func TestCSRFProtection_DisabledConfig(t *testing.T) {
	config := DefaultCSRFConfig()
	config.Enabled = false

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 when CSRF disabled, got %d", rec.Code)
	}
}

func TestCSRFProtection_ExemptPath(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/webhooks/stripe", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 for exempt path, got %d", rec.Code)
	}
}

func TestCSRFProtection_SafeMethod_SetsCookie(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that a CSRF cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == config.TokenName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Error("Expected CSRF cookie to be set on GET request")
	}
}

func TestCSRFProtection_MissingCookie(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 when CSRF cookie missing, got %d", rec.Code)
	}

	var response csrfErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "CSRF_ERROR" {
		t.Errorf("Expected error 'CSRF_ERROR', got '%s'", response.Error)
	}
}

func TestCSRFProtection_MissingHeader(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: "test-token",
	})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 when header missing, got %d", rec.Code)
	}
}

func TestCSRFProtection_TokenMismatch(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: "cookie-token",
	})
	req.Header.Set("X-CSRF-Token", "different-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 when tokens mismatch, got %d", rec.Code)
	}
}

func TestCSRFProtection_ValidToken(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "valid-csrf-token-12345"
	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: token,
	})
	req.Header.Set("X-CSRF-Token", token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 with valid token, got %d", rec.Code)
	}
}

func TestGetCSRFToken(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := GetCSRFToken(config)

	req := httptest.NewRequest("GET", "/api/csrf-token", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["csrf_token"] == "" {
		t.Error("Expected csrf_token in response")
	}

	// Check cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == config.TokenName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Error("Expected CSRF cookie to be set")
	}

	if csrfCookie.Value != response["csrf_token"] {
		t.Error("Cookie value should match response token")
	}
}
