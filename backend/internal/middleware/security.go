package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SecurityConfig holds security headers configuration
type SecurityConfig struct {
	// Enable/disable security headers
	Enabled bool

	// Content-Security-Policy settings
	ContentSecurityPolicy string

	// X-Frame-Options: DENY, SAMEORIGIN, or ALLOW-FROM uri
	FrameOptions string

	// X-Content-Type-Options: nosniff
	ContentTypeOptions string

	// Strict-Transport-Security settings
	HSTSEnabled           bool
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// X-XSS-Protection (legacy, but still useful for older browsers)
	XSSProtection string

	// Referrer-Policy
	ReferrerPolicy string

	// Permissions-Policy (formerly Feature-Policy)
	PermissionsPolicy string

	// Cross-Origin policies
	CrossOriginOpenerPolicy   string
	CrossOriginEmbedderPolicy string
	CrossOriginResourcePolicy string
}

// DefaultSecurityConfig returns secure default configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Enabled: true,

		// Strict CSP - customize based on your frontend needs
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",

		// Prevent clickjacking
		FrameOptions: "DENY",

		// Prevent MIME type sniffing
		ContentTypeOptions: "nosniff",

		// HSTS settings - SECURITY: Enable in production with HTTPS
		// Set SECURITY_HSTS_ENABLED=true to enable (requires HTTPS)
		HSTSEnabled:           false,    // PRODUCTION: Set SECURITY_HSTS_ENABLED=true
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,

		// XSS protection for legacy browsers
		XSSProtection: "1; mode=block",

		// Referrer policy - strict but allows same-origin referrers
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Permissions policy - restrict sensitive APIs
		PermissionsPolicy: "camera=(), microphone=(), geolocation=(), payment=()",

		// Cross-origin policies
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginResourcePolicy: "cross-origin",
	}
}

// isProductionEnvironment checks if the application is running in production mode
func isProductionEnvironment() bool {
	env := strings.ToLower(os.Getenv("GO_ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("APP_ENV"))
	}
	return env == "production" || env == "prod"
}

// LoadSecurityConfig loads security configuration from environment variables
func LoadSecurityConfig() *SecurityConfig {
	config := DefaultSecurityConfig()

	// Check if security headers should be enabled
	if enabled := os.Getenv("SECURITY_HEADERS_ENABLED"); enabled != "" {
		config.Enabled = strings.ToLower(enabled) == "true"
	}

	// Custom CSP from environment
	if csp := os.Getenv("SECURITY_CSP"); csp != "" {
		config.ContentSecurityPolicy = csp
	}

	// Frame options
	if frameOptions := os.Getenv("SECURITY_FRAME_OPTIONS"); frameOptions != "" {
		config.FrameOptions = frameOptions
	}

	// Auto-enable HSTS in production (unless explicitly disabled)
	isProduction := isProductionEnvironment()
	if isProduction {
		config.HSTSEnabled = true // Default to enabled in production
	}

	// HSTS settings - explicit setting overrides auto-detection
	if hstsEnabled := os.Getenv("SECURITY_HSTS_ENABLED"); hstsEnabled != "" {
		config.HSTSEnabled = strings.ToLower(hstsEnabled) == "true"
	}

	// Referrer policy
	if referrer := os.Getenv("SECURITY_REFERRER_POLICY"); referrer != "" {
		config.ReferrerPolicy = referrer
	}

	// Permissions policy
	if permissions := os.Getenv("SECURITY_PERMISSIONS_POLICY"); permissions != "" {
		config.PermissionsPolicy = permissions
	}

	// Log production security status
	if isProduction {
		if config.HSTSEnabled {
			fmt.Println("[INFO] Security: HSTS enabled (production mode)")
		} else {
			fmt.Println("[WARN] SECURITY WARNING: HSTS is explicitly disabled in production. This allows HTTP downgrade attacks.")
		}
	}

	return config
}

// SecurityHeaders returns a middleware that adds security headers to responses
func SecurityHeaders(config *SecurityConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config == nil || !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// X-Content-Type-Options
			if config.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			}

			// X-Frame-Options
			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-XSS-Protection (legacy but still useful)
			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}

			// Content-Security-Policy
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			// Permissions-Policy
			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			// Strict-Transport-Security (HSTS) - only set over HTTPS
			if config.HSTSEnabled && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
				hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
				if config.HSTSIncludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if config.HSTSPreload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Cross-Origin-Opener-Policy
			if config.CrossOriginOpenerPolicy != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
			}

			// Cross-Origin-Embedder-Policy
			if config.CrossOriginEmbedderPolicy != "" {
				w.Header().Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
			}

			// Cross-Origin-Resource-Policy
			if config.CrossOriginResourcePolicy != "" {
				w.Header().Set("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Default body size limits
const (
	// DefaultMaxBodySize is the default maximum request body size (1MB)
	DefaultMaxBodySize = 1 << 20 // 1MB

	// MaxBodySizeSmall is for endpoints with small payloads like auth (16KB)
	MaxBodySizeSmall = 16 << 10 // 16KB

	// MaxBodySizeLarge is for file uploads (10MB)
	MaxBodySizeLarge = 10 << 20 // 10MB
)

// MaxBodySize returns a middleware that limits the size of request bodies
// to prevent memory exhaustion attacks. When the limit is exceeded, the
// request body becomes unusable and subsequent reads will return an error.
//
// Usage:
//
//	r.Use(middleware.MaxBodySize(middleware.DefaultMaxBodySize))
//
// For different limits per route:
//
//	r.With(middleware.MaxBodySize(middleware.MaxBodySizeSmall)).Post("/auth/login", handler)
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for GET, HEAD, OPTIONS which don't have request bodies
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check Content-Length header first for early rejection
			if r.ContentLength > maxBytes {
				http.Error(w, `{"error":"REQUEST_TOO_LARGE","message":"Request body exceeds maximum allowed size"}`, http.StatusRequestEntityTooLarge)
				return
			}

			// Wrap the body with MaxBytesReader for streaming protection
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
