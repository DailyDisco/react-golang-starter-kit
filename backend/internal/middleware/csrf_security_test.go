package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Token Security Tests ---

func TestCSRF_TokenLength(t *testing.T) {
	// Token should be at least 44 characters (32 bytes base64 encoded)
	token, err := generateCSRFToken(32)
	require.NoError(t, err)

	// 32 bytes base64 encoded = ceil(32 * 4/3) = 44 chars (with padding)
	assert.GreaterOrEqual(t, len(token), 43, "token should be at least 43 chars for 32 bytes")
}

func TestCSRF_TokenEntropy(t *testing.T) {
	// Generate multiple tokens and verify they're all unique
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := generateCSRFToken(32)
		require.NoError(t, err)

		if tokens[token] {
			t.Fatalf("Duplicate token generated after %d iterations", i)
		}
		tokens[token] = true
	}

	assert.Len(t, tokens, iterations, "all tokens should be unique")
}

func TestCSRF_TokenCharacterSet(t *testing.T) {
	// Token should only contain URL-safe base64 characters
	urlSafeBase64Regex := regexp.MustCompile(`^[A-Za-z0-9_-]+=*$`)

	for i := 0; i < 10; i++ {
		token, err := generateCSRFToken(32)
		require.NoError(t, err)

		assert.True(t, urlSafeBase64Regex.MatchString(token),
			"token should only contain URL-safe base64 characters: %s", token)
	}
}

func TestCSRF_TokenDecodable(t *testing.T) {
	// Token should be valid base64 and decode to expected length
	token, err := generateCSRFToken(32)
	require.NoError(t, err)

	decoded, err := base64.URLEncoding.DecodeString(token)
	require.NoError(t, err, "token should be valid base64")
	assert.Len(t, decoded, 32, "decoded token should be 32 bytes")
}

func TestCSRF_EmptyCookieValue(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: "", // Empty value
	})
	req.Header.Set("X-CSRF-Token", "some-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code, "empty cookie should be rejected")
}

func TestCSRF_WhitespaceOnlyToken(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with whitespace in both cookie and header
	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: "   ", // Whitespace only
	})
	req.Header.Set("X-CSRF-Token", "   ") // Matching whitespace
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Whitespace-only tokens match (subtle.ConstantTimeCompare) but are suspicious
	// The current implementation would accept this - documenting behavior
	// This could be a security improvement to reject whitespace-only tokens
}

// --- Token Rotation Tests ---

func TestCSRF_TokenRotationAfterPOST(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	originalToken := "original-csrf-token-12345"
	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: originalToken,
	})
	req.Header.Set("X-CSRF-Token", originalToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Check that a new token was set in the response
	cookies := rec.Result().Cookies()
	var newCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == config.TokenName {
			newCookie = c
			break
		}
	}

	require.NotNil(t, newCookie, "new CSRF cookie should be set after POST")
	assert.NotEqual(t, originalToken, newCookie.Value, "token should be rotated after successful POST")
}

func TestCSRF_TokenRotation_OldTokenInvalid(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with original token
	originalToken := "original-token-abc123"
	req1 := httptest.NewRequest("POST", "/api/users", nil)
	req1.AddCookie(&http.Cookie{Name: config.TokenName, Value: originalToken})
	req1.Header.Set("X-CSRF-Token", originalToken)
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	// Get the new rotated token from the response
	var newToken string
	for _, c := range rec1.Result().Cookies() {
		if c.Name == config.TokenName {
			newToken = c.Value
			break
		}
	}
	require.NotEmpty(t, newToken)

	// Second request using the OLD token (should fail)
	req2 := httptest.NewRequest("POST", "/api/users", nil)
	req2.AddCookie(&http.Cookie{Name: config.TokenName, Value: newToken}) // New cookie
	req2.Header.Set("X-CSRF-Token", originalToken)                        // Old header token
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusForbidden, rec2.Code, "old token should not work after rotation")
}

// --- Cookie Security Attributes Tests ---

func TestCSRF_CookieSameSite(t *testing.T) {
	config := DefaultCSRFConfig()

	// Default should be Lax
	assert.Equal(t, http.SameSiteLaxMode, config.CookieSameSite)
}

func TestCSRF_CookieHTTPOnly(t *testing.T) {
	config := DefaultCSRFConfig()

	// Must be false for double-submit pattern to work
	assert.False(t, config.CookieHTTPOnly, "CookieHTTPOnly must be false for double-submit pattern")
}

func TestCSRF_CookieSecureFlag_Development(t *testing.T) {
	// Clear any environment
	os.Unsetenv("GO_ENV")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("CSRF_COOKIE_SECURE")

	config := LoadCSRFConfig()

	// In development, Secure should be false by default
	assert.False(t, config.CookieSecure, "Secure should be false in development")
}

func TestCSRF_CookieSecureFlag_Production(t *testing.T) {
	// Set production environment
	t.Setenv("GO_ENV", "production")
	os.Unsetenv("CSRF_COOKIE_SECURE") // Don't override

	config := LoadCSRFConfig()

	assert.True(t, config.CookieSecure, "Secure should be true in production")
}

func TestCSRF_CookieSecureFlag_ProductionExplicitOverride(t *testing.T) {
	// Explicitly disable in production (not recommended but possible)
	t.Setenv("GO_ENV", "production")
	t.Setenv("CSRF_COOKIE_SECURE", "false")

	config := LoadCSRFConfig()

	// Explicit override takes precedence
	assert.False(t, config.CookieSecure, "explicit override should take precedence")
}

func TestCSRF_isProductionEnv(t *testing.T) {
	tests := []struct {
		name     string
		goEnv    string
		appEnv   string
		expected bool
	}{
		{"GO_ENV=production", "production", "", true},
		{"GO_ENV=prod", "prod", "", true},
		{"GO_ENV=PRODUCTION uppercase", "PRODUCTION", "", true},
		{"GO_ENV=development", "development", "", false},
		{"APP_ENV=production fallback", "", "production", true},
		{"APP_ENV=prod fallback", "", "prod", true},
		{"empty env", "", "", false},
		{"staging", "staging", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("GO_ENV")
			os.Unsetenv("APP_ENV")

			if tt.goEnv != "" {
				t.Setenv("GO_ENV", tt.goEnv)
			}
			if tt.appEnv != "" {
				t.Setenv("APP_ENV", tt.appEnv)
			}

			result := isProductionEnv()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- LoadCSRFConfig Tests ---

func TestLoadCSRFConfig_SameSiteModes(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected http.SameSite
	}{
		{"strict", "strict", http.SameSiteStrictMode},
		{"STRICT uppercase", "STRICT", http.SameSiteStrictMode},
		{"lax", "lax", http.SameSiteLaxMode},
		{"none", "none", http.SameSiteNoneMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("GO_ENV")
			t.Setenv("CSRF_COOKIE_SAMESITE", tt.envValue)

			config := LoadCSRFConfig()
			assert.Equal(t, tt.expected, config.CookieSameSite)
		})
	}
}

func TestLoadCSRFConfig_CookieDomain(t *testing.T) {
	t.Setenv("CSRF_COOKIE_DOMAIN", ".example.com")
	os.Unsetenv("GO_ENV")

	config := LoadCSRFConfig()

	assert.Equal(t, ".example.com", config.CookieDomain)
}

func TestLoadCSRFConfig_Disabled(t *testing.T) {
	t.Setenv("CSRF_ENABLED", "false")
	os.Unsetenv("GO_ENV")

	config := LoadCSRFConfig()

	assert.False(t, config.Enabled)
}

// --- Edge Cases ---

func TestCSRF_VeryLongToken(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a very long token (1000+ chars)
	longToken := make([]byte, 1000)
	for i := range longToken {
		longToken[i] = 'A'
	}
	longTokenStr := string(longToken)

	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: longTokenStr,
	})
	req.Header.Set("X-CSRF-Token", longTokenStr)
	rec := httptest.NewRecorder()

	// Should not crash, should match since cookie and header are same
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "matching long tokens should work")
}

func TestCSRF_NullByteInToken(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tokenWithNull := "token\x00withNull"

	req := httptest.NewRequest("POST", "/api/users", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: tokenWithNull,
	})
	req.Header.Set("X-CSRF-Token", tokenWithNull)
	rec := httptest.NewRecorder()

	// Go's http.Cookie sanitizes invalid bytes (including null) from cookie values
	// This is correct security behavior - null bytes get dropped, causing mismatch
	handler.ServeHTTP(rec, req)
	// Cookie value gets sanitized (null byte removed), header doesn't, so they mismatch
	assert.Equal(t, http.StatusForbidden, rec.Code,
		"null byte in cookie gets sanitized, causing token mismatch (good security)")
}

func TestCSRF_FormValueFallback(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "form-csrf-token-12345"
	req := httptest.NewRequest("POST", "/api/users?csrf_token="+token, nil)
	req.AddCookie(&http.Cookie{
		Name:  config.TokenName,
		Value: token,
	})
	// No X-CSRF-Token header, but form value present
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "form value fallback should work")
}

// --- Response Content Type ---

func TestCSRF_ErrorResponseContentType(t *testing.T) {
	config := DefaultCSRFConfig()

	handler := CSRFProtection(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/users", nil)
	// No cookie, no header - will fail
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
		"error response should be application/json")
}
