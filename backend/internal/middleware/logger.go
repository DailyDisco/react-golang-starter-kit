package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"react-golang-starter/internal/auth"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// sensitiveFields contains field names that should be redacted from logs
var sensitiveFields = []string{
	"password",
	"currentPassword",
	"current_password",
	"newPassword",
	"new_password",
	"confirmPassword",
	"confirm_password",
	"token",
	"accessToken",
	"access_token",
	"refreshToken",
	"refresh_token",
	"secret",
	"apiKey",
	"api_key",
	"apiSecret",
	"api_secret",
	"privateKey",
	"private_key",
	"credential",
	"authorization",
	"totp",
	"totpCode",
	"totp_code",
	"backupCode",
	"backup_code",
}

// StructuredLogger returns a middleware that logs HTTP requests with structured logging
func StructuredLogger() func(http.Handler) http.Handler {
	return StructuredLoggerWithConfig(LoadLogConfig())
}

// StructuredLoggerWithConfig returns a middleware that logs HTTP requests with structured logging using provided config
func StructuredLoggerWithConfig(config *LogConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check sampling rate
			if !config.ShouldLogRequest() {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			requestID := getRequestID(r)

			// Extract user context if enabled
			var userContext map[string]interface{}
			if config.IncludeUserContext {
				userContext = extractUserContext(r.Context())
			} else {
				userContext = make(map[string]interface{})
			}

			// Capture request body if enabled
			var requestBody string
			if config.IncludeRequestBody && r.Body != nil {
				requestBody = captureRequestBody(r, config.MaxRequestBodySize)
			}

			// Create a logger with request context
			logger := log.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Str("user_agent", r.UserAgent()).
				Str("ip", getRealIP(r)).
				Logger()

			// Log headers if sanitization is enabled
			headers := make(map[string]string)
			if config.SanitizeHeaders {
				for name, values := range r.Header {
					if config.IsHeaderAllowed(name) {
						headers[name] = strings.Join(values, ", ")
					}
				}
			}

			// Log the incoming request
			requestLog := logger.Info().
				Str("protocol", r.Proto).
				Str("host", r.Host).
				Str("referer", r.Referer()).
				Dict("headers", zerolog.Dict())

			// Add headers to log
			for name, value := range headers {
				requestLog = requestLog.Str(name, value)
			}

			// Add user context if available
			for key, value := range userContext {
				if strVal, ok := value.(string); ok {
					requestLog = requestLog.Str(key, strVal)
				} else if intVal, ok := value.(uint); ok {
					requestLog = requestLog.Uint64(key, uint64(intVal))
				}
			}

			// Add request body if captured
			if requestBody != "" {
				requestLog = requestLog.Str("request_body", requestBody)
			}

			requestLog.Msg("request started")

			// Create a response writer wrapper to capture status code and size
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Capture response body if enabled
			var responseBuffer *bytes.Buffer
			var responseWriter http.ResponseWriter
			if config.IncludeResponseBody {
				responseBuffer = &bytes.Buffer{}
				responseWriter = &responseCaptureWriter{
					ResponseWriter: wrapped,
					buffer:         responseBuffer,
					maxSize:        config.MaxResponseBodySize,
				}
			} else {
				responseWriter = wrapped
			}

			// Defer logging the response
			defer func() {
				duration := time.Since(start)

				// Capture response body if enabled
				var responseBody string
				if responseBuffer != nil && responseBuffer.Len() > 0 {
					responseBody = responseBuffer.String()
					if len(responseBody) > config.MaxResponseBodySize {
						responseBody = responseBody[:config.MaxResponseBodySize] + "... [truncated]"
					}
				}

				// Determine log level based on status code
				var logEvent *zerolog.Event
				if wrapped.Status() >= 500 {
					logEvent = logger.Error()
				} else if wrapped.Status() >= 400 {
					logEvent = logger.Warn()
				} else {
					logEvent = logger.Info()
				}

				logEvent.
					Int("status", wrapped.Status()).
					Int("bytes_written", wrapped.BytesWritten()).
					Dur("duration_ms", duration)

				// Add response body if captured
				if responseBody != "" {
					logEvent = logEvent.Str("response_body", responseBody)
				}

				logEvent.Msg("request completed")
			}()

			// Call the next handler
			next.ServeHTTP(responseWriter, r)
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

	// Check X-Real-IP header (set by reverse proxy)
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

// extractUserContext extracts user information from request context
func extractUserContext(ctx context.Context) map[string]interface{} {
	userContext := make(map[string]interface{})

	// Try to get user from context
	if user, ok := auth.GetUserFromContext(ctx); ok {
		userContext["user_id"] = user.ID
		userContext["user_email"] = user.Email
		userContext["user_name"] = user.Name
	} else {
		// Fallback to individual context values
		if userID, ok := auth.GetUserIDFromContext(ctx); ok {
			userContext["user_id"] = userID
		}
		if userEmail, ok := auth.GetUserEmailFromContext(ctx); ok {
			userContext["user_email"] = userEmail
		}
	}

	return userContext
}

// sanitizeRequestBody redacts sensitive fields from JSON request bodies
func sanitizeRequestBody(body string) string {
	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// Not JSON or invalid - use regex fallback for common patterns
		return sanitizeWithRegex(body)
	}

	// Recursively sanitize the JSON object
	sanitizeMap(data)

	// Re-encode to JSON
	sanitized, err := json.Marshal(data)
	if err != nil {
		return "[sanitization error]"
	}
	return string(sanitized)
}

// sanitizeMap recursively redacts sensitive fields from a map
func sanitizeMap(data map[string]interface{}) {
	for key, value := range data {
		// Check if this key is sensitive
		if isSensitiveField(key) {
			data[key] = "[REDACTED]"
			continue
		}

		// Recursively handle nested objects
		switch v := value.(type) {
		case map[string]interface{}:
			sanitizeMap(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					sanitizeMap(m)
				}
			}
		}
	}
}

// isSensitiveField checks if a field name is sensitive
func isSensitiveField(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.ToLower(sensitive) == lower {
			return true
		}
	}
	return false
}

// sensitiveFieldRegex matches common sensitive field patterns in non-JSON bodies
var sensitiveFieldRegex = regexp.MustCompile(`(?i)(password|token|secret|api_?key|credential|authorization)=([^&\s]+)`)

// sanitizeWithRegex redacts sensitive values in non-JSON request bodies
func sanitizeWithRegex(body string) string {
	return sensitiveFieldRegex.ReplaceAllString(body, "${1}=[REDACTED]")
}

// captureRequestBody captures the request body for logging
func captureRequestBody(r *http.Request, maxSize int) string {
	if r.Body == nil {
		return ""
	}

	// Read the body
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, int64(maxSize)))
	if err != nil {
		return ""
	}

	// Restore the body for the handler
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	bodyStr := string(bodyBytes)

	// Truncate if too long
	if len(bodyStr) > maxSize {
		bodyStr = bodyStr[:maxSize] + "... [truncated]"
	}

	// Sanitize sensitive fields before returning
	return sanitizeRequestBody(bodyStr)
}

// responseCaptureWriter captures response body for logging
type responseCaptureWriter struct {
	http.ResponseWriter
	buffer  *bytes.Buffer
	maxSize int
	written int
}

func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	// Write to the original response writer
	n, err := w.ResponseWriter.Write(data)

	// Also write to our buffer if we haven't exceeded max size
	if w.buffer != nil && w.written < w.maxSize {
		remaining := w.maxSize - w.written
		if len(data) > remaining {
			data = data[:remaining]
		}
		w.buffer.Write(data)
		w.written += len(data)
	}

	return n, err
}
