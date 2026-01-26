package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Middleware chain integration tests.
// These verify that middleware work correctly together.

// --- Request ID Propagation ---

func TestMiddlewareChain_RequestIDPropagation(t *testing.T) {
	t.Run("request ID is set and available to downstream handlers", func(t *testing.T) {
		var capturedRequestID string

		// Handler that captures the request ID
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedRequestID = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		// Chain: RequestID -> Handler
		chain := RequestIDMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		assert.NotEmpty(t, capturedRequestID, "request ID should be set")
		assert.Equal(t, capturedRequestID, rec.Header().Get("X-Request-ID"),
			"request ID in header should match context")
	})

	t.Run("existing request ID is preserved", func(t *testing.T) {
		existingID := "existing-request-id-12345"
		var capturedRequestID string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedRequestID = GetRequestID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		chain := RequestIDMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		assert.Equal(t, existingID, capturedRequestID,
			"existing request ID should be preserved")
	})
}

// --- Security Headers Applied ---

func TestMiddlewareChain_SecurityHeadersApplied(t *testing.T) {
	t.Run("security headers are set on responses", func(t *testing.T) {
		config := &SecurityConfig{
			Enabled:            true,
			ContentTypeOptions: "nosniff",
			FrameOptions:       "DENY",
			XSSProtection:      "1; mode=block",
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		chain := SecurityHeaders(config)(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		// Security headers should be present
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	})
}

// --- CSRF Exempt Paths ---

func TestMiddlewareChain_CSRFExemptPaths(t *testing.T) {
	// Default exempt paths from CSRF config
	defaultExemptPaths := []string{
		"/api/webhooks",
		"/api/health",
		"/api/auth/oauth",
	}

	exemptTestPaths := []string{
		"/api/webhooks/stripe",
		"/api/health",
		"/api/auth/oauth/google/callback",
		"/api/auth/oauth/github/callback",
	}

	for _, path := range exemptTestPaths {
		t.Run("exempt: "+path, func(t *testing.T) {
			result := isExemptPath(path, defaultExemptPaths)
			assert.True(t, result, "path %s should be exempt from CSRF", path)
		})
	}

	nonExemptPaths := []string{
		"/api/users",
		"/api/settings",
		"/api/v1/organizations",
	}

	for _, path := range nonExemptPaths {
		t.Run("not exempt: "+path, func(t *testing.T) {
			result := isExemptPath(path, defaultExemptPaths)
			assert.False(t, result, "path %s should NOT be exempt from CSRF", path)
		})
	}
}

// --- Recovery Middleware ---

func TestMiddlewareChain_RecoveryPreventsPanicPropagation(t *testing.T) {
	t.Run("panic in handler is recovered", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong!")
		})

		chain := RecoveryMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Should not panic
		require.NotPanics(t, func() {
			chain.ServeHTTP(rec, req)
		})

		// Should return 500 error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

// --- Safe Methods Don't Require CSRF Token ---

func TestMiddlewareChain_SafeMethodsNoCSRF(t *testing.T) {
	safeMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
	}

	config := DefaultCSRFConfig()

	for _, method := range safeMethods {
		t.Run(method+" is safe", func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			chain := CSRFProtection(config)(handler)

			req := httptest.NewRequest(method, "/api/users", nil)
			rec := httptest.NewRecorder()

			chain.ServeHTTP(rec, req)

			// Safe methods should succeed without CSRF token
			assert.Equal(t, http.StatusOK, rec.Code,
				"%s should not require CSRF token", method)
		})
	}
}

// --- Unsafe Methods Require CSRF Token ---

func TestMiddlewareChain_UnsafeMethodsRequireCSRF(t *testing.T) {
	unsafeMethods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	config := DefaultCSRFConfig()

	for _, method := range unsafeMethods {
		t.Run(method+" requires CSRF", func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			chain := CSRFProtection(config)(handler)

			req := httptest.NewRequest(method, "/api/users", nil)
			// No CSRF token provided
			rec := httptest.NewRecorder()

			chain.ServeHTTP(rec, req)

			// Should be forbidden without CSRF token
			assert.Equal(t, http.StatusForbidden, rec.Code,
				"%s should require CSRF token", method)
		})
	}
}

// --- Multiple Middleware Layer Order ---

func TestMiddlewareChain_LayerOrder(t *testing.T) {
	t.Run("outer middleware runs before inner", func(t *testing.T) {
		var order []string

		outer := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "outer-before")
				next.ServeHTTP(w, r)
				order = append(order, "outer-after")
			})
		}

		inner := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "inner-before")
				next.ServeHTTP(w, r)
				order = append(order, "inner-after")
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		// Chain: outer -> inner -> handler
		chain := outer(inner(handler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		expected := []string{
			"outer-before",
			"inner-before",
			"handler",
			"inner-after",
			"outer-after",
		}

		assert.Equal(t, expected, order, "middleware should execute in correct order")
	})
}
