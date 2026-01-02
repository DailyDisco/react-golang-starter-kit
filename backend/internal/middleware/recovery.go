package middleware

import (
	"net/http"
	"runtime/debug"

	"react-golang-starter/internal/response"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware recovers from panics and returns a proper error response.
// It logs the panic with stack trace and reports to Sentry if configured.
// Should be placed early in the middleware chain (after RequestID) to catch all panics.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for correlation
				requestID := GetRequestID(r.Context())

				// Capture stack trace
				stack := string(debug.Stack())

				// Log the panic with full context
				log.Error().
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("panic", err).
					Str("stack", stack).
					Msg("Recovered from panic")

				// Report to Sentry if configured
				if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
					hub.RecoverWithContext(r.Context(), err)
				} else {
					// Fallback to global hub
					sentry.CurrentHub().RecoverWithContext(r.Context(), err)
				}

				// Return a generic error response (don't expose internal details)
				response.InternalError(w, r, "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddlewareWithCallback is like RecoveryMiddleware but accepts a callback
// for custom panic handling (e.g., custom metrics, alerting).
func RecoveryMiddlewareWithCallback(onPanic func(r *http.Request, err interface{}, stack []byte)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					requestID := GetRequestID(r.Context())

					// Log the panic
					log.Error().
						Str("request_id", requestID).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Interface("panic", err).
						Str("stack", string(stack)).
						Msg("Recovered from panic")

					// Call custom handler
					if onPanic != nil {
						onPanic(r, err, stack)
					}

					// Report to Sentry
					if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
						hub.RecoverWithContext(r.Context(), err)
					} else {
						sentry.CurrentHub().RecoverWithContext(r.Context(), err)
					}

					// Return error response
					response.InternalError(w, r, "An unexpected error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
