package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============ shouldSkipUsageTracking Tests ============

func TestShouldSkipUsageTracking(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "health endpoint",
			path: "/api/health",
			want: true,
		},
		{
			name: "health subpath",
			path: "/api/health/ready",
			want: true,
		},
		{
			name: "webhooks endpoint",
			path: "/api/webhooks",
			want: true,
		},
		{
			name: "webhooks stripe",
			path: "/api/webhooks/stripe",
			want: true,
		},
		{
			name: "usage endpoint",
			path: "/api/usage",
			want: true,
		},
		{
			name: "usage stats",
			path: "/api/usage/stats",
			want: true,
		},
		{
			name: "metrics endpoint",
			path: "/metrics",
			want: true,
		},
		{
			name: "debug endpoint",
			path: "/debug",
			want: true,
		},
		{
			name: "debug pprof",
			path: "/debug/pprof",
			want: true,
		},
		{
			name: "regular api endpoint",
			path: "/api/users",
			want: false,
		},
		{
			name: "organizations endpoint",
			path: "/api/organizations",
			want: false,
		},
		{
			name: "nested api path",
			path: "/api/organizations/test-org/members",
			want: false,
		},
		{
			name: "files endpoint",
			path: "/api/files",
			want: false,
		},
		{
			name: "root path",
			path: "/",
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
		{
			name: "health prefix also matches healthy",
			path: "/api/healthy",
			want: true, // HasPrefix matches - this is expected behavior
		},
		{
			name: "usage-like but different",
			path: "/api/user-stats",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipUsageTracking(tt.path)
			if got != tt.want {
				t.Errorf("shouldSkipUsageTracking(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ============ getClientIP Tests ============

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.2, 172.16.0.1"},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For with spaces",
			headers:    map[string]string{"X-Forwarded-For": "  192.168.1.1  ,  10.0.0.2  "},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "192.168.1.1"},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1", "X-Real-IP": "10.0.0.2"},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr IPv4 with port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr IPv4 without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr IPv6 bracket format",
			headers:    map[string]string{},
			remoteAddr: "[::1]:12345",
			expectedIP: "::1",
		},
		{
			name:       "RemoteAddr IPv6 full address bracket format",
			headers:    map[string]string{},
			remoteAddr: "[2001:db8::1]:12345",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "RemoteAddr localhost",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "127.0.0.1",
		},
		{
			name:       "Empty headers fall back to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:54321",
			expectedIP: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			got := getClientIP(req)
			if got != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", got, tt.expectedIP)
			}
		})
	}
}

// ============ UsageMiddleware Tests ============

func TestUsageMiddleware_PassesToNextHandler(t *testing.T) {
	// UsageMiddleware requires a UsageService, but we can test with nil
	// to verify it doesn't panic and still calls the next handler
	handlerCalled := false
	handler := UsageMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("UsageMiddleware did not call next handler")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("UsageMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestUsageMiddleware_PreservesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Not Found", http.StatusNotFound},
		{"Internal Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := UsageMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("UsageMiddleware() status = %d, want %d", rr.Code, tt.statusCode)
			}
		})
	}
}

func TestUsageMiddleware_WritesResponseBody(t *testing.T) {
	expectedBody := "test response body"
	handler := UsageMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Body.String() != expectedBody {
		t.Errorf("UsageMiddleware() body = %q, want %q", rr.Body.String(), expectedBody)
	}
}

func TestUsageMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled := false
			handler := UsageMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/api/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if !handlerCalled {
				t.Errorf("UsageMiddleware did not call handler for method %s", method)
			}
		})
	}
}

func TestUsageMiddleware_SkippedPaths(t *testing.T) {
	paths := []string{
		"/api/health",
		"/api/webhooks/stripe",
		"/api/usage/stats",
		"/metrics",
		"/debug/pprof",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			handlerCalled := false
			handler := UsageMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			// Handler should still be called even for skipped paths
			if !handlerCalled {
				t.Errorf("UsageMiddleware did not call handler for path %s", path)
			}

			if rr.Code != http.StatusOK {
				t.Errorf("UsageMiddleware() status = %d, want %d for path %s", rr.Code, http.StatusOK, path)
			}
		})
	}
}
