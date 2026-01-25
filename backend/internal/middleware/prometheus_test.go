package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ============ normalizeRoutePath Tests ============

func TestNormalizeRoutePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short path unchanged",
			input: "/api/users",
			want:  "/api/users",
		},
		{
			name:  "path with params",
			input: "/api/users/{id}",
			want:  "/api/users/{id}",
		},
		{
			name:  "nested path",
			input: "/api/organizations/{orgSlug}/members/{userId}",
			want:  "/api/organizations/{orgSlug}/members/{userId}",
		},
		{
			name:  "path exactly 100 chars stays same",
			input: "/api/users/12345678901234567890123456789012345678901234567890123456789012345678901234567890",
			want:  "/api/users/12345678901234567890123456789012345678901234567890123456789012345678901234567890",
		},
		{
			name:  "path over 100 chars truncated",
			input: "/api/users/1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678extra",
			want:  "/api/users/1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678e",
		},
		{
			name:  "empty path",
			input: "",
			want:  "",
		},
		{
			name:  "root path",
			input: "/",
			want:  "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRoutePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRoutePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// helper to create request with chi route context
func newRequestWithChiContext(method, path string, routePattern string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	rctx := chi.NewRouteContext()
	if routePattern != "" {
		rctx.RoutePatterns = []string{routePattern}
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ============ PrometheusMiddleware Tests ============

func TestPrometheusMiddleware_PassesToNextHandler(t *testing.T) {
	handlerCalled := false
	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := newRequestWithChiContext(http.MethodGet, "/test", "/test")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("PrometheusMiddleware did not call next handler")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("PrometheusMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestPrometheusMiddleware_PreservesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := newRequestWithChiContext(http.MethodGet, "/test", "")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("PrometheusMiddleware() status = %d, want %d", rr.Code, tt.statusCode)
			}
		})
	}
}

func TestPrometheusMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled := false
			handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if r.Method != method {
					t.Errorf("Request method = %q, want %q", r.Method, method)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := newRequestWithChiContext(method, "/test", "")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if !handlerCalled {
				t.Errorf("PrometheusMiddleware did not call handler for method %s", method)
			}
		})
	}
}

func TestPrometheusMiddleware_WithRoutePattern(t *testing.T) {
	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := newRequestWithChiContext(http.MethodGet, "/api/users/123", "/api/users/{id}")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PrometheusMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestPrometheusMiddleware_WithoutRoutePattern(t *testing.T) {
	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := newRequestWithChiContext(http.MethodGet, "/unknown/path", "")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should still work and use URL path as fallback
	if rr.Code != http.StatusOK {
		t.Errorf("PrometheusMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestPrometheusMiddleware_WritesResponseBody(t *testing.T) {
	expectedBody := "test response"
	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))

	req := newRequestWithChiContext(http.MethodGet, "/test", "")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Body.String() != expectedBody {
		t.Errorf("PrometheusMiddleware() body = %q, want %q", rr.Body.String(), expectedBody)
	}
}

func TestPrometheusMiddleware_WithLongPath(t *testing.T) {
	// Create a path longer than 100 characters
	longPath := "/api/very/long/path/that/exceeds/one/hundred/characters/and/should/be/truncated/by/the/normalize/function/12345"

	handler := PrometheusMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := newRequestWithChiContext(http.MethodGet, longPath, "")
	rr := httptest.NewRecorder()

	// Should not panic with long path
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PrometheusMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}
