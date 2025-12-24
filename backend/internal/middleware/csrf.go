package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	// Enable/disable CSRF protection
	Enabled bool

	// Token settings
	TokenLength int
	TokenName   string // Cookie and header name

	// Cookie settings
	CookiePath     string
	CookieDomain   string
	CookieSecure   bool // Set to true in production with HTTPS
	CookieHTTPOnly bool // Should be false so JS can read it
	CookieSameSite http.SameSite
	CookieMaxAge   int // In seconds

	// Exempt paths (e.g., webhooks that use signature verification)
	ExemptPaths []string

	// Exempt methods (safe methods don't need CSRF)
	SafeMethods []string
}

// DefaultCSRFConfig returns secure default configuration
//
// SECURITY WARNING: In production with HTTPS, you MUST set:
//   - CSRF_COOKIE_SECURE=true (prevents cookie theft over HTTP)
//   - Use SameSite=Strict mode (default)
//
// The CookieHTTPOnly is intentionally false for the double-submit pattern,
// which requires JavaScript to read the CSRF token from the cookie.
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		Enabled:        true,
		TokenLength:    32,
		TokenName:      "csrf_token",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   false, // PRODUCTION: Set CSRF_COOKIE_SECURE=true with HTTPS
		CookieHTTPOnly: false, // Must be false for double-submit pattern (by design)
		CookieSameSite: http.SameSiteStrictMode,
		CookieMaxAge:   86400, // 24 hours
		ExemptPaths: []string{
			"/api/webhooks/",
			"/api/v1/webhooks/",
			"/health",
			"/test",
		},
		SafeMethods: []string{"GET", "HEAD", "OPTIONS", "TRACE"},
	}
}

// LoadCSRFConfig loads CSRF configuration from environment variables
func LoadCSRFConfig() *CSRFConfig {
	config := DefaultCSRFConfig()

	// Check if CSRF should be enabled
	if enabled := os.Getenv("CSRF_ENABLED"); enabled != "" {
		config.Enabled = strings.ToLower(enabled) == "true"
	}

	// Cookie secure flag
	if secure := os.Getenv("CSRF_COOKIE_SECURE"); secure != "" {
		config.CookieSecure = strings.ToLower(secure) == "true"
	}

	// Cookie domain
	if domain := os.Getenv("CSRF_COOKIE_DOMAIN"); domain != "" {
		config.CookieDomain = domain
	}

	// SameSite setting
	if sameSite := os.Getenv("CSRF_COOKIE_SAMESITE"); sameSite != "" {
		switch strings.ToLower(sameSite) {
		case "strict":
			config.CookieSameSite = http.SameSiteStrictMode
		case "lax":
			config.CookieSameSite = http.SameSiteLaxMode
		case "none":
			config.CookieSameSite = http.SameSiteNoneMode
		}
	}

	return config
}

// csrfErrorResponse is the JSON response for CSRF errors
type csrfErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// isSafeMethod checks if the HTTP method is safe (doesn't modify state)
func isSafeMethod(method string, safeMethods []string) bool {
	method = strings.ToUpper(method)
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// isExemptPath checks if the request path is exempt from CSRF protection
func isExemptPath(path string, exemptPaths []string) bool {
	for _, exempt := range exemptPaths {
		if strings.HasPrefix(path, exempt) {
			return true
		}
	}
	return false
}

// CSRFProtection returns a middleware that provides CSRF protection
// using the double-submit cookie pattern
func CSRFProtection(config *CSRFConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if CSRF is disabled
			if config == nil || !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip for exempt paths (webhooks, health checks)
			if isExemptPath(r.URL.Path, config.ExemptPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// For safe methods, just ensure a token exists in the cookie
			if isSafeMethod(r.Method, config.SafeMethods) {
				ensureCSRFToken(w, r, config)
				next.ServeHTTP(w, r)
				return
			}

			// For state-changing methods, validate the token
			cookie, err := r.Cookie(config.TokenName)
			if err != nil || cookie.Value == "" {
				writeCSRFError(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			// Get token from header
			headerToken := r.Header.Get("X-CSRF-Token")
			if headerToken == "" {
				// Also check form value as fallback
				headerToken = r.FormValue(config.TokenName)
			}

			if headerToken == "" {
				writeCSRFError(w, "CSRF token not provided in header", http.StatusForbidden)
				return
			}

			// Constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
				writeCSRFError(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}

			// Token is valid, proceed with request
			// Optionally rotate the token after validation for extra security
			ensureCSRFToken(w, r, config)
			next.ServeHTTP(w, r)
		})
	}
}

// ensureCSRFToken ensures a CSRF token exists in the response cookie
func ensureCSRFToken(w http.ResponseWriter, r *http.Request, config *CSRFConfig) {
	// Check if token already exists
	if cookie, err := r.Cookie(config.TokenName); err == nil && cookie.Value != "" {
		// Token exists, no need to set a new one
		return
	}

	// Generate new token
	token, err := generateCSRFToken(config.TokenLength)
	if err != nil {
		// Log error but don't fail the request
		return
	}

	// Set the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     config.TokenName,
		Value:    token,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		MaxAge:   config.CookieMaxAge,
		Secure:   config.CookieSecure,
		HttpOnly: config.CookieHTTPOnly,
		SameSite: config.CookieSameSite,
	})
}

// writeCSRFError writes a JSON error response for CSRF failures
func writeCSRFError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := csrfErrorResponse{
		Error:   "CSRF_ERROR",
		Message: message,
		Code:    code,
	}

	json.NewEncoder(w).Encode(response)
}

// GetCSRFToken is a handler that returns the current CSRF token
// Useful for SPAs to get the token on initial load
func GetCSRFToken(config *CSRFConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate a new token
		token, err := generateCSRFToken(config.TokenLength)
		if err != nil {
			http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
			return
		}

		// Set the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     config.TokenName,
			Value:    token,
			Path:     config.CookiePath,
			Domain:   config.CookieDomain,
			MaxAge:   config.CookieMaxAge,
			Secure:   config.CookieSecure,
			HttpOnly: config.CookieHTTPOnly,
			SameSite: config.CookieSameSite,
		})

		// Return the token in response body
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"csrf_token": token,
			"expires_in": "86400",
		})
	}
}
