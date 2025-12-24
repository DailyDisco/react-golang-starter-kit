package ratelimit

import (
	"encoding/json"
	"net/http"
	"react-golang-starter/internal/auth"
	"strconv"
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

// rateLimitHandler creates a custom handler for rate limit exceeded responses
// that includes proper JSON response and ensures Retry-After header is set
func rateLimitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// httprate sets Retry-After header automatically
	w.WriteHeader(http.StatusTooManyRequests)

	response := rateLimitResponse{
		Error:      "RATE_LIMITED",
		Message:    "Too many requests. Please try again later.",
		Code:       http.StatusTooManyRequests,
		RetryAfter: 60, // Default window in seconds
	}

	json.NewEncoder(w).Encode(response)
}

// NewIPRateLimitMiddleware creates middleware for IP-based rate limiting
// Headers automatically set by httprate:
// - X-RateLimit-Limit: maximum requests per window
// - X-RateLimit-Remaining: requests remaining in current window
// - X-RateLimit-Reset: Unix timestamp when the window resets
// - Retry-After: seconds to wait (only when rate limited)
func NewIPRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.IPRequestsPerMinute,
		config.GetIPWindow(),
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(rateLimitHandler),
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
			// Fallback to IP if no user context
			ip, err := httprate.KeyByIP(r)
			if err != nil {
				return "", err
			}
			return "ip:" + ip, nil
		}),
		httprate.WithLimitHandler(rateLimitHandler),
	)
}

// NewAuthRateLimitMiddleware creates middleware for authentication endpoints
// More restrictive limits to prevent brute-force attacks
// Headers automatically set by httprate (same as NewIPRateLimitMiddleware)
func NewAuthRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.Limit(
		config.AuthRequestsPerMinute,
		config.GetAuthWindow(),
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(rateLimitHandler),
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
			// Use IP for unauthenticated requests
			ip, err := httprate.KeyByIP(r)
			if err != nil {
				return "", err
			}
			return "ip:" + ip, nil
		}),
		httprate.WithLimitHandler(rateLimitHandler),
	)
}
