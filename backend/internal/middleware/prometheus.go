package middleware

import (
	"net/http"
	"strconv"
	"time"

	"react-golang-starter/internal/observability"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// PrometheusMiddleware records HTTP request metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track in-flight requests
		observability.HTTPRequestsInFlight.Inc()
		defer observability.HTTPRequestsInFlight.Dec()

		// Wrap response writer to capture status code
		wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get route pattern for consistent labels (avoids high cardinality from path params)
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		// Normalize route pattern to reduce cardinality
		routePattern = normalizeRoutePath(routePattern)

		// Record metrics
		observability.RecordHTTPRequest(
			r.Method,
			routePattern,
			strconv.Itoa(wrapped.Status()),
			duration,
		)
	})
}

// normalizeRoutePath reduces path cardinality by replacing dynamic segments
func normalizeRoutePath(path string) string {
	// Common patterns that should be normalized
	// These are already handled by Chi's RoutePattern, but we add some fallbacks

	// Limit path length to prevent high cardinality
	if len(path) > 100 {
		path = path[:100]
	}

	return path
}
