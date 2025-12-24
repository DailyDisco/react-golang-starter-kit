package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ Context Helper Function Tests ============

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		wantUser   bool
		wantUserID uint
	}{
		{
			"user present in context",
			func() context.Context {
				user := &models.User{ID: 123, Email: "test@example.com", Role: "user"}
				return context.WithValue(context.Background(), UserContextKey, user)
			},
			true,
			123,
		},
		{
			"no user in context",
			func() context.Context {
				return context.Background()
			},
			false,
			0,
		},
		{
			"wrong type in context",
			func() context.Context {
				return context.WithValue(context.Background(), UserContextKey, "not a user")
			},
			false,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			user, ok := GetUserFromContext(ctx)

			if ok != tt.wantUser {
				t.Errorf("GetUserFromContext() ok = %v, want %v", ok, tt.wantUser)
			}

			if tt.wantUser && user.ID != tt.wantUserID {
				t.Errorf("GetUserFromContext() user.ID = %v, want %v", user.ID, tt.wantUserID)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID uint
		wantOK bool
	}{
		{
			"user ID present",
			context.WithValue(context.Background(), UserIDContextKey, uint(42)),
			42,
			true,
		},
		{
			"no user ID in context",
			context.Background(),
			0,
			false,
		},
		{
			"wrong type in context",
			context.WithValue(context.Background(), UserIDContextKey, "42"),
			0,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := GetUserIDFromContext(tt.ctx)

			if ok != tt.wantOK {
				t.Errorf("GetUserIDFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			if id != tt.wantID {
				t.Errorf("GetUserIDFromContext() id = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestGetUserEmailFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantEmail string
		wantOK    bool
	}{
		{
			"email present",
			context.WithValue(context.Background(), UserEmailContextKey, "test@example.com"),
			"test@example.com",
			true,
		},
		{
			"no email in context",
			context.Background(),
			"",
			false,
		},
		{
			"wrong type in context",
			context.WithValue(context.Background(), UserEmailContextKey, 123),
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, ok := GetUserEmailFromContext(tt.ctx)

			if ok != tt.wantOK {
				t.Errorf("GetUserEmailFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			if email != tt.wantEmail {
				t.Errorf("GetUserEmailFromContext() email = %v, want %v", email, tt.wantEmail)
			}
		})
	}
}

func TestGetUserRoleFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantRole string
		wantOK   bool
	}{
		{
			"role present",
			context.WithValue(context.Background(), UserRoleContextKey, models.RoleAdmin),
			models.RoleAdmin,
			true,
		},
		{
			"super_admin role",
			context.WithValue(context.Background(), UserRoleContextKey, models.RoleSuperAdmin),
			models.RoleSuperAdmin,
			true,
		},
		{
			"no role in context",
			context.Background(),
			"",
			false,
		},
		{
			"wrong type in context",
			context.WithValue(context.Background(), UserRoleContextKey, 123),
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, ok := GetUserRoleFromContext(tt.ctx)

			if ok != tt.wantOK {
				t.Errorf("GetUserRoleFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			if role != tt.wantRole {
				t.Errorf("GetUserRoleFromContext() role = %v, want %v", role, tt.wantRole)
			}
		})
	}
}

// ============ ContextKey Type Tests ============

func TestContextKeyType(t *testing.T) {
	// Ensure context keys are unique and don't collide with string keys
	ctx := context.WithValue(context.Background(), "user", "string-user")
	ctx = context.WithValue(ctx, UserContextKey, &models.User{ID: 1})

	// The ContextKey type should not collide with plain string key
	stringValue := ctx.Value("user")
	if stringValue != "string-user" {
		t.Errorf("String key value changed unexpectedly")
	}

	user, ok := GetUserFromContext(ctx)
	if !ok || user.ID != 1 {
		t.Error("ContextKey type should not collide with string keys")
	}
}

// ============ AuthMiddleware Tests (Token Extraction) ============

func TestAuthMiddleware_NoCookieNoHeader(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddleware_InvalidAuthHeader(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"missing Bearer prefix", "some-token"},
		{"empty Bearer", "Bearer "},
		{"invalid format", "Basic user:pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// Note: TestAuthMiddleware_ExpiredToken and TestAuthMiddleware_ValidToken
// require a database connection and are covered in integration tests.
// These tests focus on token extraction and validation logic that doesn't
// require database access.

// ============ OptionalAuthMiddleware Tests ============

func TestOptionalAuthMiddleware_NoToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	var userFromContext *models.User
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFromContext, _ = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if userFromContext != nil {
		t.Error("Expected no user in context when no token provided")
	}
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	var userFromContext *models.User
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFromContext, _ = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	// Should still pass through but without user context
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if userFromContext != nil {
		t.Error("Expected no user in context when invalid token provided")
	}
}

// ============ Context Key Constants Tests ============

func TestContextKeyConstants(t *testing.T) {
	// Ensure context key constants are defined and unique
	keys := map[ContextKey]string{
		UserContextKey:      "user",
		UserIDContextKey:    "user_id",
		UserEmailContextKey: "user_email",
		UserRoleContextKey:  "user_role",
	}

	for key, expectedValue := range keys {
		if string(key) != expectedValue {
			t.Errorf("ContextKey %v should be %q, got %q", key, expectedValue, string(key))
		}
	}

	// Ensure all keys are unique
	seen := make(map[ContextKey]bool)
	for key := range keys {
		if seen[key] {
			t.Errorf("Duplicate context key: %v", key)
		}
		seen[key] = true
	}
}
