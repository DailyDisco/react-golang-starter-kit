package ratelimit

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"react-golang-starter/internal/auth"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitError represents a rate limit error
type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
}

func (e RateLimitError) Error() string {
	return e.Message
}

// rateLimitResponse is the JSON response for rate limited requests
type rateLimitResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Code       int    `json:"code"`
	RetryAfter int    `json:"retry_after"`
}

// createRateLimitHandler creates a custom handler for rate limit exceeded responses
// that includes proper JSON response with the actual retry-after duration.
func createRateLimitHandler(windowSeconds int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers so the browser can read the error response
		setCORSErrorHeaders(w, r)
		w.Header().Set("Content-Type", "application/json")
		// httprate sets Retry-After header automatically
		w.WriteHeader(http.StatusTooManyRequests)

		response := rateLimitResponse{
			Error:      "RATE_LIMITED",
			Message:    "Too many requests. Please try again later.",
			Code:       http.StatusTooManyRequests,
			RetryAfter: windowSeconds,
		}

		json.NewEncoder(w).Encode(response)
	}
}

// setCORSErrorHeaders sets CORS headers on error responses.
// This is needed because rate limit error responses exit early
// before the CORS middleware can add headers to the response.
func setCORSErrorHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Check if origin is allowed
	allowedOrigins := getAllowedOriginsForRateLimit()
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			return
		}
	}
}

// getAllowedOriginsForRateLimit returns the allowed CORS origins from environment variables
func getAllowedOriginsForRateLimit() []string {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsEnv != "" {
		return strings.Split(originsEnv, ",")
	}

	// Default development origins
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://localhost:5193",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
}

// getClientIP extracts the real client IP from the request, respecting trusted proxies.
// If the request comes from a trusted proxy, it uses X-Forwarded-For header.
// Otherwise, it uses the RemoteAddr directly to prevent IP spoofing.
func getClientIP(r *http.Request, config *Config) string {
	// Extract the remote address (without port)
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		remoteIP = r.RemoteAddr
	}

	// Only trust X-Forwarded-For if request comes from a trusted proxy
	if len(config.parsedTrustedProxies) > 0 && config.IsTrustedProxy(remoteIP) {
		// Parse X-Forwarded-For header (can contain comma-separated IPs)
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			// X-Forwarded-For format: client, proxy1, proxy2, ...
			// We want the leftmost IP that is not a trusted proxy
			ips := strings.Split(xff, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if ip != "" && !config.IsTrustedProxy(ip) {
					return ip
				}
			}
		}

		// Also check X-Real-IP header
		xri := r.Header.Get("X-Real-IP")
		if xri != "" {
			xri = strings.TrimSpace(xri)
			if !config.IsTrustedProxy(xri) {
				return xri
			}
		}
	}

	// Fall back to RemoteAddr
	return remoteIP
}

// keyByTrustedIP creates a key function that respects trusted proxy configuration
func keyByTrustedIP(config *Config) httprate.KeyFunc {
	return func(r *http.Request) (string, error) {
		return getClientIP(r, config), nil
	}
}

// NewIPRateLimitMiddleware creates middleware for IP-based rate limiting
// Headers automatically set by httprate:
// - X-RateLimit-Limit: maximum requests per window
// - X-RateLimit-Remaining: requests remaining in current window
// - X-RateLimit-Reset: Unix timestamp when the window resets
// - Retry-After: seconds to wait (only when rate limited)
//
// SECURITY: This middleware uses trusted proxy configuration to prevent IP spoofing.
// Set RATE_LIMIT_TRUSTED_PROXIES to configure trusted proxy IPs/CIDRs.
func NewIPRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.IPRequestsPerMinute,
		config.GetIPWindow(),
		httprate.WithKeyFuncs(keyByTrustedIP(config)),
		httprate.WithLimitHandler(createRateLimitHandler(int(config.GetIPWindow().Seconds()))),
	)
}

// NewUserRateLimitMiddleware creates middleware for user-based rate limiting
// Headers automatically set by httprate (same as NewIPRateLimitMiddleware)
func NewUserRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.UserRequestsPerMinute,
		config.GetUserWindow(),
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			// Try to get user ID from context (set by auth middleware)
			if userID, ok := auth.GetUserIDFromContext(r.Context()); ok {
				return "user:" + strconv.FormatUint(uint64(userID), 10), nil
			}
			// Fallback to trusted IP if no user context
			return "ip:" + getClientIP(r, config), nil
		}),
		httprate.WithLimitHandler(createRateLimitHandler(int(config.GetUserWindow().Seconds()))),
	)
}

// NewAuthRateLimitMiddleware creates middleware for authentication endpoints
// More restrictive limits to prevent brute-force attacks
// Headers automatically set by httprate (same as NewIPRateLimitMiddleware)
//
// SECURITY: Uses trusted proxy configuration to prevent IP spoofing on auth endpoints.
func NewAuthRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.AuthRequestsPerMinute,
		config.GetAuthWindow(),
		httprate.WithKeyFuncs(keyByTrustedIP(config)),
		httprate.WithLimitHandler(createRateLimitHandler(int(config.GetAuthWindow().Seconds()))),
	)
}

// NewAPIRateLimitMiddleware creates middleware for general API endpoints
// Headers automatically set by httprate (same as NewIPRateLimitMiddleware)
func NewAPIRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.APIRequestsPerMinute,
		config.GetAPIWindow(),
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			// Check if user is authenticated
			if userID, ok := auth.GetUserIDFromContext(r.Context()); ok {
				// Use user ID for authenticated requests
				return "user:" + strconv.FormatUint(uint64(userID), 10), nil
			}
			// Use trusted IP for unauthenticated requests
			return "ip:" + getClientIP(r, config), nil
		}),
		httprate.WithLimitHandler(createRateLimitHandler(int(config.GetAPIWindow().Seconds()))),
	)
}
