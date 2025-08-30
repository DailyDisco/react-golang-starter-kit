package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// StructuredLogger returns a middleware that logs HTTP requests with structured logging
func StructuredLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := getRequestID(r)

			// Create a logger with request context
			logger := log.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Str("user_agent", r.UserAgent()).
				Str("ip", getRealIP(r)).
				Logger()

			// Log the incoming request
			logger.Info().
				Str("protocol", r.Proto).
				Str("host", r.Host).
				Str("referer", r.Referer()).
				Msg("request started")

			// Create a response writer wrapper to capture status code and size
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Defer logging the response
			defer func() {
				duration := time.Since(start)

				// Determine log level based on status code
				logEvent := logger.Info()
				if wrapped.Status() >= 400 && wrapped.Status() < 500 {
					logEvent = logger.Warn()
				} else if wrapped.Status() >= 500 {
					logEvent = logger.Error()
				}

				logEvent.
					Int("status", wrapped.Status()).
					Int("bytes_written", wrapped.BytesWritten()).
					Dur("duration_ms", duration).
					Msg("request completed")
			}()

			// Call the next handler
			next.ServeHTTP(wrapped, r)
		})
	}
}

// getRequestID extracts or generates a request ID
func getRequestID(r *http.Request) string {
	// Check for existing request ID in headers
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Check for Chi's request ID
	if requestID := middleware.GetReqID(r.Context()); requestID != "" {
		return requestID
	}

	// Generate a new UUID-based request ID
	return uuid.New().String()
}

// getRealIP extracts the real client IP address from various headers
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common with proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check X-Forwarded header
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		return strings.TrimSpace(xf)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
