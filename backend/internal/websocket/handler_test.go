package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============ extractTokenFromRequest Tests ============

func TestExtractTokenFromRequest_FromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "cookie-token-123",
	})

	token, err := extractTokenFromRequest(req)

	if err != nil {
		t.Fatalf("extractTokenFromRequest() error = %v", err)
	}

	if token != "cookie-token-123" {
		t.Errorf("token = %q, want %q", token, "cookie-token-123")
	}
}

func TestExtractTokenFromRequest_FromAuthHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Authorization", "Bearer header-token-456")

	token, err := extractTokenFromRequest(req)

	if err != nil {
		t.Fatalf("extractTokenFromRequest() error = %v", err)
	}

	if token != "header-token-456" {
		t.Errorf("token = %q, want %q", token, "header-token-456")
	}
}

func TestExtractTokenFromRequest_FromQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws?token=query-token-789", nil)

	token, err := extractTokenFromRequest(req)

	if err != nil {
		t.Fatalf("extractTokenFromRequest() error = %v", err)
	}

	if token != "query-token-789" {
		t.Errorf("token = %q, want %q", token, "query-token-789")
	}
}

func TestExtractTokenFromRequest_Priority(t *testing.T) {
	// Cookie should take priority over header
	req := httptest.NewRequest(http.MethodGet, "/ws?token=query", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "cookie",
	})
	req.Header.Set("Authorization", "Bearer header")

	token, _ := extractTokenFromRequest(req)

	if token != "cookie" {
		t.Errorf("Cookie should take priority, got token = %q, want %q", token, "cookie")
	}
}

func TestExtractTokenFromRequest_HeaderPriorityOverQuery(t *testing.T) {
	// Header should take priority over query param
	req := httptest.NewRequest(http.MethodGet, "/ws?token=query", nil)
	req.Header.Set("Authorization", "Bearer header")

	token, _ := extractTokenFromRequest(req)

	if token != "header" {
		t.Errorf("Header should take priority over query, got token = %q, want %q", token, "header")
	}
}

func TestExtractTokenFromRequest_NoToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	_, err := extractTokenFromRequest(req)

	if err == nil {
		t.Error("extractTokenFromRequest() should return error when no token")
	}

	if err != http.ErrNoCookie {
		t.Errorf("error = %v, want %v", err, http.ErrNoCookie)
	}
}

func TestExtractTokenFromRequest_EmptyCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "",
	})

	_, err := extractTokenFromRequest(req)

	if err == nil {
		t.Error("extractTokenFromRequest() should return error for empty cookie")
	}
}

func TestExtractTokenFromRequest_InvalidAuthHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{"no bearer prefix", "token-only"},
		{"wrong prefix", "Basic some-token"},
		{"empty", ""},
		{"just bearer", "Bearer"},
		{"bearer no space", "Bearertoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			_, err := extractTokenFromRequest(req)

			if err == nil {
				t.Error("extractTokenFromRequest() should return error for invalid header")
			}
		})
	}
}

func TestExtractTokenFromRequest_CaseInsensitiveBearer(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"lowercase", "bearer token123"},
		{"uppercase", "BEARER token123"},
		{"mixed", "BeArEr token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			req.Header.Set("Authorization", tt.header)

			token, err := extractTokenFromRequest(req)

			if err != nil {
				t.Fatalf("extractTokenFromRequest() error = %v", err)
			}

			if token != "token123" {
				t.Errorf("token = %q, want %q", token, "token123")
			}
		})
	}
}

// ============ getAllowedOrigins Tests ============

func TestGetAllowedOrigins_RemovesProtocol(t *testing.T) {
	// Note: This tests the function behavior, but actual origins come from config
	// We can't easily mock the config, so we test the logic patterns

	origins := getAllowedOrigins()

	// Origins should not contain http:// or https:// prefixes
	for _, origin := range origins {
		if len(origin) > 7 && origin[:7] == "http://" {
			t.Errorf("origin contains http:// prefix: %q", origin)
		}
		if len(origin) > 8 && origin[:8] == "https://" {
			t.Errorf("origin contains https:// prefix: %q", origin)
		}
	}
}

func TestGetAllowedOrigins_ReturnsSlice(t *testing.T) {
	origins := getAllowedOrigins()

	if origins == nil {
		t.Error("getAllowedOrigins() returned nil, want non-nil slice")
	}
}

// ============ Handler Tests ============

func TestHandler_ReturnsHandler(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestHandler_UnauthorizedWithoutToken(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_InvalidToken(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "invalid-token",
	})
	w := httptest.NewRecorder()

	handler(w, req)

	// Should return 401 for invalid token
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
