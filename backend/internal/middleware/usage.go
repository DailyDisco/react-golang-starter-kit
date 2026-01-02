package middleware

import (
	"net/http"
	"strings"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/services"

	"github.com/go-chi/chi/v5"
)

// UsageMiddleware records API call usage for authenticated requests
func UsageMiddleware(usageService *services.UsageService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Process request first
			next.ServeHTTP(w, r)

			// Only record usage for authenticated users
			userID, ok := auth.GetUserIDFromContext(r.Context())
			if !ok || userID == 0 {
				return
			}

			// Get route pattern for consistent resource identification
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			// Skip certain paths from usage tracking
			if shouldSkipUsageTracking(routePattern) {
				return
			}

			// Get client info
			ip := getClientIP(r)
			ua := r.Header.Get("User-Agent")

			// Record the API call asynchronously to not block the response
			go func() {
				usageService.RecordAPICall(r.Context(), &userID, nil, routePattern, ip, ua)
			}()
		})
	}
}

// shouldSkipUsageTracking returns true for paths that shouldn't be counted
func shouldSkipUsageTracking(path string) bool {
	skipPrefixes := []string{
		"/api/health",
		"/api/webhooks",
		"/api/usage", // Don't count usage API calls themselves
		"/metrics",
		"/debug",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// Check if this is IPv6 (has multiple colons)
		if strings.Count(ip, ":") > 1 {
			// IPv6 address - check if it's in [ip]:port format
			if strings.HasPrefix(ip, "[") {
				if bracketIdx := strings.Index(ip, "]"); bracketIdx != -1 {
					ip = ip[1:bracketIdx]
				}
			}
		} else {
			// IPv4 address
			ip = ip[:idx]
		}
	}

	return ip
}
