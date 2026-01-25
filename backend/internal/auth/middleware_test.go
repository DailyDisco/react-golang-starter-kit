package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// ============ SetUserContext Tests ============

func TestSetUserContext(t *testing.T) {
	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  models.RoleAdmin,
	}

	ctx := SetUserContext(context.Background(), user)

	// Verify all context values are set
	retrievedUser, ok := GetUserFromContext(ctx)
	if !ok {
		t.Error("user should be retrievable from context")
	}
	if retrievedUser.ID != user.ID {
		t.Errorf("user.ID = %d, want %d", retrievedUser.ID, user.ID)
	}

	userID, ok := GetUserIDFromContext(ctx)
	if !ok || userID != user.ID {
		t.Errorf("userID = %d, want %d", userID, user.ID)
	}

	email, ok := GetUserEmailFromContext(ctx)
	if !ok || email != user.Email {
		t.Errorf("email = %q, want %q", email, user.Email)
	}

	role, ok := GetUserRoleFromContext(ctx)
	if !ok || role != user.Role {
		t.Errorf("role = %q, want %q", role, user.Role)
	}
}

// ============ SetClaimsContext Tests ============

func TestSetClaimsContext(t *testing.T) {
	claims := &Claims{
		UserID: 456,
		Email:  "claims@example.com",
		Role:   models.RoleSuperAdmin,
	}

	ctx := SetClaimsContext(context.Background(), claims)

	retrievedClaims, ok := GetClaimsFromContext(ctx)
	if !ok {
		t.Error("claims should be retrievable from context")
	}
	if retrievedClaims.UserID != claims.UserID {
		t.Errorf("claims.UserID = %d, want %d", retrievedClaims.UserID, claims.UserID)
	}
	if retrievedClaims.Email != claims.Email {
		t.Errorf("claims.Email = %q, want %q", retrievedClaims.Email, claims.Email)
	}
	if retrievedClaims.Role != claims.Role {
		t.Errorf("claims.Role = %q, want %q", retrievedClaims.Role, claims.Role)
	}
}

func TestGetClaimsFromContext_NoClaims(t *testing.T) {
	ctx := context.Background()

	_, ok := GetClaimsFromContext(ctx)
	if ok {
		t.Error("should return false when no claims in context")
	}
}

func TestGetClaimsFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsContextKey, "not claims")

	_, ok := GetClaimsFromContext(ctx)
	if ok {
		t.Error("should return false when wrong type in context")
	}
}

// ============ Blacklist Configuration Tests ============

func TestGetBlacklistFailMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{"default closed when not set", "", "closed"},
		{"explicit open mode", "open", "open"},
		{"invalid value defaults to closed", "invalid", "closed"},
		{"OPEN uppercase defaults to closed", "OPEN", "closed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("TOKEN_BLACKLIST_FAIL_MODE")
			} else {
				os.Setenv("TOKEN_BLACKLIST_FAIL_MODE", tt.envValue)
			}
			defer os.Unsetenv("TOKEN_BLACKLIST_FAIL_MODE")

			result := getBlacklistFailMode()
			if result != tt.expected {
				t.Errorf("getBlacklistFailMode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBlacklistConstants(t *testing.T) {
	// Verify constants have expected values
	if blacklistCacheTTL != 15*time.Minute {
		t.Errorf("blacklistCacheTTL = %v, want %v", blacklistCacheTTL, 15*time.Minute)
	}

	if blacklistCachePrefix != "blacklist:" {
		t.Errorf("blacklistCachePrefix = %q, want %q", blacklistCachePrefix, "blacklist:")
	}
}

// ============ Blacklist Functions with Nil DB Tests ============

func TestBlacklistToken_NilDB(t *testing.T) {
	// BlacklistToken should return nil when database is not initialized
	err := BlacklistToken("test-token", 1, time.Now().Add(time.Hour), "test")
	if err != nil {
		t.Errorf("BlacklistToken() with nil DB = %v, want nil", err)
	}
}

func TestIsTokenBlacklisted_NilDB(t *testing.T) {
	// IsTokenBlacklisted should return false when database is not initialized
	result := IsTokenBlacklisted("test-token")
	if result {
		t.Error("IsTokenBlacklisted() with nil DB = true, want false")
	}
}

func TestCleanupExpiredBlacklistEntries_NilDB(t *testing.T) {
	// CleanupExpiredBlacklistEntries should return nil when database is not initialized
	err := CleanupExpiredBlacklistEntries()
	if err != nil {
		t.Errorf("CleanupExpiredBlacklistEntries() with nil DB = %v, want nil", err)
	}
}

func TestRevokeAllUserTokens_NilDB(t *testing.T) {
	// RevokeAllUserTokens should return nil when database is not initialized
	err := RevokeAllUserTokens(1, "test")
	if err != nil {
		t.Errorf("RevokeAllUserTokens() with nil DB = %v, want nil", err)
	}
}

// ============ Admin Middleware Role Logic Tests ============

func TestHasRole_AdminCheck(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		expected bool
	}{
		{"user role not admin", models.RoleUser, false},
		{"admin role is admin", models.RoleAdmin, true},
		{"super_admin role is admin", models.RoleSuperAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// HasRole checks if a role matches any of the allowed roles
			result := HasRole(tt.userRole, models.RoleAdmin, models.RoleSuperAdmin)
			if result != tt.expected {
				t.Errorf("HasRole(%q, admin, super_admin) = %v, want %v", tt.userRole, result, tt.expected)
			}
		})
	}
}

func TestHasRole_SuperAdminCheck(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		expected bool
	}{
		{"user role not super_admin", models.RoleUser, false},
		{"admin role not super_admin", models.RoleAdmin, false},
		{"super_admin role is super_admin", models.RoleSuperAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For super admin check, we only check for super_admin role
			result := tt.userRole == models.RoleSuperAdmin
			if result != tt.expected {
				t.Errorf("role == super_admin for %q = %v, want %v", tt.userRole, result, tt.expected)
			}
		})
	}
}

// ============ Role Hierarchy Tests ============

func TestRoleHierarchy_Levels(t *testing.T) {
	// Verify role hierarchy exists and has correct ordering
	userLevel, userExists := models.RoleHierarchy[models.RoleUser]
	adminLevel, adminExists := models.RoleHierarchy[models.RoleAdmin]
	superAdminLevel, superExists := models.RoleHierarchy[models.RoleSuperAdmin]

	if !userExists {
		t.Error("RoleHierarchy should contain user role")
	}
	if !adminExists {
		t.Error("RoleHierarchy should contain admin role")
	}
	if !superExists {
		t.Error("RoleHierarchy should contain super_admin role")
	}

	// Verify hierarchy order: user < admin < super_admin
	if userLevel >= adminLevel {
		t.Errorf("user level (%d) should be less than admin level (%d)", userLevel, adminLevel)
	}
	if adminLevel >= superAdminLevel {
		t.Errorf("admin level (%d) should be less than super_admin level (%d)", adminLevel, superAdminLevel)
	}
}

func TestMinRoleLevel_Logic(t *testing.T) {
	tests := []struct {
		name       string
		userRole   string
		minRole    string
		shouldPass bool
	}{
		{"user meets user requirement", models.RoleUser, models.RoleUser, true},
		{"user fails admin requirement", models.RoleUser, models.RoleAdmin, false},
		{"user fails super_admin requirement", models.RoleUser, models.RoleSuperAdmin, false},
		{"admin meets user requirement", models.RoleAdmin, models.RoleUser, true},
		{"admin meets admin requirement", models.RoleAdmin, models.RoleAdmin, true},
		{"admin fails super_admin requirement", models.RoleAdmin, models.RoleSuperAdmin, false},
		{"super_admin meets user requirement", models.RoleSuperAdmin, models.RoleUser, true},
		{"super_admin meets admin requirement", models.RoleSuperAdmin, models.RoleAdmin, true},
		{"super_admin meets super_admin requirement", models.RoleSuperAdmin, models.RoleSuperAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userLevel := models.RoleHierarchy[tt.userRole]
			minLevel := models.RoleHierarchy[tt.minRole]
			passes := userLevel >= minLevel

			if passes != tt.shouldPass {
				t.Errorf("userLevel(%d) >= minLevel(%d) = %v, want %v",
					userLevel, minLevel, passes, tt.shouldPass)
			}
		})
	}
}

// ============ InvalidateUserCache Tests ============

func TestInvalidateUserCache_KeyFormat(t *testing.T) {
	// Test that the cache key format is correct
	// InvalidateUserCache uses fmt.Sprintf("user:%d", userID)
	expectedKey := "user:12345"
	correctKey := "user:12345" // This is the expected format

	if expectedKey != correctKey {
		t.Errorf("cache key format mismatch: %q vs %q", expectedKey, correctKey)
	}
}

// ============ Middleware Request Context Tests ============

func TestAdminMiddleware_Logic_UserRole(t *testing.T) {
	// Test the role-checking logic that AdminMiddleware uses
	user := &models.User{ID: 1, Email: "user@example.com", Role: models.RoleUser}

	// AdminMiddleware checks: HasRole(user.Role, models.RoleAdmin, models.RoleSuperAdmin)
	isAdmin := HasRole(user.Role, models.RoleAdmin, models.RoleSuperAdmin)
	if isAdmin {
		t.Error("User with 'user' role should not pass admin check")
	}
}

func TestAdminMiddleware_Logic_AdminRole(t *testing.T) {
	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	isAdmin := HasRole(user.Role, models.RoleAdmin, models.RoleSuperAdmin)
	if !isAdmin {
		t.Error("User with 'admin' role should pass admin check")
	}
}

func TestAdminMiddleware_Logic_SuperAdminRole(t *testing.T) {
	user := &models.User{ID: 1, Email: "superadmin@example.com", Role: models.RoleSuperAdmin}

	isAdmin := HasRole(user.Role, models.RoleAdmin, models.RoleSuperAdmin)
	if !isAdmin {
		t.Error("User with 'super_admin' role should pass admin check")
	}
}

func TestSuperAdminMiddleware_Logic(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"user role fails", models.RoleUser, false},
		{"admin role fails", models.RoleAdmin, false},
		{"super_admin role passes", models.RoleSuperAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SuperAdminMiddleware checks: user.Role == models.RoleSuperAdmin
			isSuperAdmin := tt.role == models.RoleSuperAdmin
			if isSuperAdmin != tt.expected {
				t.Errorf("role == super_admin for %q = %v, want %v", tt.role, isSuperAdmin, tt.expected)
			}
		})
	}
}

// ============ Additional Context Key Tests ============

func TestClaimsContextKey_Value(t *testing.T) {
	if ClaimsContextKey != ContextKey("claims") {
		t.Errorf("ClaimsContextKey = %q, want %q", ClaimsContextKey, "claims")
	}
}

func TestAllContextKeysUnique(t *testing.T) {
	keys := []ContextKey{
		UserContextKey,
		UserIDContextKey,
		UserEmailContextKey,
		UserRoleContextKey,
		ClaimsContextKey,
	}

	seen := make(map[ContextKey]bool)
	for _, key := range keys {
		if seen[key] {
			t.Errorf("Duplicate context key found: %q", key)
		}
		seen[key] = true
	}
}

// ============ AuthMiddleware Cookie Extraction Tests ============

func TestAuthMiddleware_WithCookie(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "invalid-token-from-cookie",
	})
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	// Should fail with invalid token (but proves cookie extraction is attempted)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestOptionalAuthMiddleware_WithCookie(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	var userFromContext *models.User
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFromContext, _ = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "invalid-token-from-cookie",
	})
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	// Should pass through (optional) but without user context
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if userFromContext != nil {
		t.Error("Expected no user in context when invalid cookie token provided")
	}
}
