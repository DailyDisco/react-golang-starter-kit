package middleware

import (
	"context"
	"net/http"

	"react-golang-starter/internal/contextkeys"

	"github.com/google/uuid"
)

// RequestIDKey is re-exported from contextkeys for backward compatibility
var RequestIDKey = contextkeys.RequestIDKey

// RequestIDHeader is re-exported from contextkeys for backward compatibility
const RequestIDHeader = contextkeys.RequestIDHeader

// RequestIDMiddleware adds a unique request ID to each request.
// It checks for an existing X-Request-ID header first, and generates a new UUID if not present.
// The request ID is added to the request context and set as a response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID in header
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			// Generate a new UUID
			requestID = uuid.New().String()
		}

		// Add request ID to response header
		w.Header().Set(RequestIDHeader, requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), contextkeys.RequestIDKey, requestID)

		// Continue with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is found.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(contextkeys.RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetRequestIDFromRequest is a convenience function to get request ID from an HTTP request.
func GetRequestIDFromRequest(r *http.Request) string {
	return GetRequestID(r.Context())
}
