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
		CrossOriginResourcePolicy: "same-origin",
	}
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

	// HSTS settings
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
