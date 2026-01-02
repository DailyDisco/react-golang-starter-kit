package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// CacheRule defines a caching rule for a specific route pattern
type CacheRule struct {
	// PathPrefix is the URL path prefix to match (e.g., "/api/public/")
	PathPrefix string

	// MaxAge is the Cache-Control max-age in seconds
	MaxAge int

	// Public sets Cache-Control to "public" if true, "private" if false
	Public bool

	// NoCache sets Cache-Control: no-cache directive
	NoCache bool

	// NoStore sets Cache-Control: no-store directive
	NoStore bool

	// MustRevalidate adds must-revalidate directive
	MustRevalidate bool
}

// CacheHeadersConfig holds configuration for cache headers middleware
type CacheHeadersConfig struct {
	// Enabled controls whether cache headers are applied
	Enabled bool

	// Rules are evaluated in order; first match wins
	Rules []CacheRule
}

// DefaultCacheHeadersConfig returns sensible defaults for cache headers.
// Rules are ordered from most specific to most general.
func DefaultCacheHeadersConfig() *CacheHeadersConfig {
	return &CacheHeadersConfig{
		Enabled: true,
		Rules: []CacheRule{
			// Health check - very short cache for monitoring
			{PathPrefix: "/api/health", MaxAge: 5, Public: true},

			// Public changelog/announcements - CDN-friendly
			{PathPrefix: "/api/v1/changelog", MaxAge: 60, Public: true},

			// Admin routes - never cache (sensitive data)
			{PathPrefix: "/api/admin/", NoCache: true, NoStore: true, MustRevalidate: true},

			// Auth routes - never cache
			{PathPrefix: "/api/v1/auth/", NoCache: true, NoStore: true, MustRevalidate: true},

			// Billing routes - private, short cache
			{PathPrefix: "/api/v1/billing/", MaxAge: 60, Public: false},

			// Files routes - private, medium cache
			{PathPrefix: "/api/v1/files/", MaxAge: 300, Public: false},

			// Settings routes - private, short cache
			{PathPrefix: "/api/v1/settings/", MaxAge: 60, Public: false},

			// Organization routes - private, short cache
			{PathPrefix: "/api/v1/organizations/", MaxAge: 60, Public: false},

			// Default for all other API routes - private, short cache
			{PathPrefix: "/api/", MaxAge: 60, Public: false},
		},
	}
}

// LoadCacheHeadersConfig loads cache headers configuration from environment variables
func LoadCacheHeadersConfig() *CacheHeadersConfig {
	config := DefaultCacheHeadersConfig()

	// Check if cache headers should be enabled
	if enabled := os.Getenv("CACHE_HEADERS_ENABLED"); enabled != "" {
		config.Enabled = strings.ToLower(enabled) == "true"
	}

	return config
}

// CacheHeaders returns a middleware that applies Cache-Control headers
// based on route patterns. Rules are evaluated in order; first match wins.
//
// Only applies to GET and HEAD requests, as other methods should not be cached.
func CacheHeaders(config *CacheHeadersConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config == nil || !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Only apply cache headers to GET and HEAD requests
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Find matching rule (first match wins)
			for _, rule := range config.Rules {
				if strings.HasPrefix(r.URL.Path, rule.PathPrefix) {
					applyCacheRule(w, rule)
					break
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// applyCacheRule sets the appropriate cache headers based on the rule
func applyCacheRule(w http.ResponseWriter, rule CacheRule) {
	var directives []string

	// Handle no-cache/no-store first (overrides other settings)
	if rule.NoCache || rule.NoStore {
		if rule.NoCache {
			directives = append(directives, "no-cache")
		}
		if rule.NoStore {
			directives = append(directives, "no-store")
		}
		if rule.MustRevalidate {
			directives = append(directives, "must-revalidate")
		}

		w.Header().Set("Cache-Control", strings.Join(directives, ", "))
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		return
	}

	// Set visibility (public or private)
	if rule.Public {
		directives = append(directives, "public")
	} else {
		directives = append(directives, "private")
	}

	// Set max-age
	directives = append(directives, fmt.Sprintf("max-age=%d", rule.MaxAge))

	// Add must-revalidate if specified
	if rule.MustRevalidate {
		directives = append(directives, "must-revalidate")
	}

	w.Header().Set("Cache-Control", strings.Join(directives, ", "))
}

// WithCacheRule returns a middleware that applies a specific cache rule.
// Useful for overriding defaults on specific routes.
//
// Example:
//
//	r.With(middleware.WithCacheRule(middleware.CacheRule{
//	    MaxAge: 3600,
//	    Public: true,
//	})).Get("/static/config.json", handler)
func WithCacheRule(rule CacheRule) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to GET and HEAD
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				applyCacheRule(w, rule)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NoCacheHeaders is a convenience middleware that prevents caching.
// Use this for sensitive endpoints that should never be cached.
func NoCacheHeaders() func(next http.Handler) http.Handler {
	return WithCacheRule(CacheRule{
		NoCache:        true,
		NoStore:        true,
		MustRevalidate: true,
	})
}

// Common cache rules for reuse
var (
	// CacheRulePublic5Min is for public data that can be CDN-cached for 5 minutes
	CacheRulePublic5Min = CacheRule{MaxAge: 300, Public: true}

	// CacheRulePublic1Hour is for stable public data
	CacheRulePublic1Hour = CacheRule{MaxAge: 3600, Public: true}

	// CacheRulePrivate1Min is for user-specific data with short cache
	CacheRulePrivate1Min = CacheRule{MaxAge: 60, Public: false}

	// CacheRulePrivate5Min is for user-specific data with medium cache
	CacheRulePrivate5Min = CacheRule{MaxAge: 300, Public: false}

	// CacheRuleNoCache prevents any caching
	CacheRuleNoCache = CacheRule{NoCache: true, NoStore: true, MustRevalidate: true}
)
