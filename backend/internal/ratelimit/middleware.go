package ratelimit

import (
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

// NewIPRateLimitMiddleware creates middleware for IP-based rate limiting
func NewIPRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.LimitByIP(config.IPRequestsPerMinute, config.GetIPWindow())
}

// NewUserRateLimitMiddleware creates middleware for user-based rate limiting
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
	)
}

// NewAuthRateLimitMiddleware creates middleware for authentication endpoints
func NewAuthRateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return httprate.LimitByIP(config.AuthRequestsPerMinute, config.GetAuthWindow())
}

// NewAPIRateLimitMiddleware creates middleware for general API endpoints
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
	)
}

// RateLimitByIP creates IP-based rate limiting middleware with custom limits
func RateLimitByIP(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitByIP(requests, window)
}

// RateLimitByUser creates user-based rate limiting middleware with custom limits
func RateLimitByUser(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(
		requests,
		window,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			if userID, ok := auth.GetUserIDFromContext(r.Context()); ok {
				return strconv.FormatUint(uint64(userID), 10), nil
			}
			ip, err := httprate.KeyByIP(r)
			if err != nil {
				return "", err
			}
			return ip, nil
		}),
	)
}
