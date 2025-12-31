package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ GetOrganizationFromContext Tests ============

func TestGetOrganizationFromContext_Valid(t *testing.T) {
	org := &models.Organization{ID: 1, Name: "Test Org", Slug: "test-org"}
	ctx := context.WithValue(context.Background(), OrganizationContextKey, org)

	result := GetOrganizationFromContext(ctx)
	if result == nil {
		t.Error("GetOrganizationFromContext() returned nil for valid context")
	}
	if result.ID != org.ID {
		t.Errorf("GetOrganizationFromContext() ID = %d, want %d", result.ID, org.ID)
	}
}

func TestGetOrganizationFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	result := GetOrganizationFromContext(ctx)
	if result != nil {
		t.Error("GetOrganizationFromContext() should return nil for missing context")
	}
}

func TestGetOrganizationFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), OrganizationContextKey, "not an org")

	result := GetOrganizationFromContext(ctx)
	if result != nil {
		t.Error("GetOrganizationFromContext() should return nil for wrong type")
	}
}

// ============ GetMembershipFromContext Tests ============

func TestGetMembershipFromContext_Valid(t *testing.T) {
	membership := &models.OrganizationMember{
		ID:             1,
		OrganizationID: 1,
		UserID:         1,
		Role:           models.OrgRoleMember,
	}
	ctx := context.WithValue(context.Background(), MembershipContextKey, membership)

	result := GetMembershipFromContext(ctx)
	if result == nil {
		t.Error("GetMembershipFromContext() returned nil for valid context")
	}
	if result.ID != membership.ID {
		t.Errorf("GetMembershipFromContext() ID = %d, want %d", result.ID, membership.ID)
	}
}

func TestGetMembershipFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	result := GetMembershipFromContext(ctx)
	if result != nil {
		t.Error("GetMembershipFromContext() should return nil for missing context")
	}
}

func TestGetMembershipFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), MembershipContextKey, "not a membership")

	result := GetMembershipFromContext(ctx)
	if result != nil {
		t.Error("GetMembershipFromContext() should return nil for wrong type")
	}
}

// ============ NewTenantMiddleware Tests ============

func TestNewTenantMiddleware(t *testing.T) {
	// Can create with nil db (we just test initialization)
	middleware := NewTenantMiddleware(nil)
	if middleware == nil {
		t.Error("NewTenantMiddleware() returned nil")
	}
}

// ============ RequireOrganization Middleware Tests ============

func TestRequireOrganization_Unauthorized(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	req := httptest.NewRequest(http.MethodGet, "/orgs/test/resource", nil)
	w := httptest.NewRecorder()

	handler := middleware.RequireOrganization(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when unauthorized")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RequireOrganization() without user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireOrganization_MissingSlug(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	// Create request with user context but no org slug
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := context.WithValue(context.Background(), UserContextKey, user)
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler := middleware.RequireOrganization(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when org slug is missing")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RequireOrganization() without slug status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ RequireOrgRole Middleware Tests ============

func TestRequireOrgRole_NoMembershipContext(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()

	roleMiddleware := middleware.RequireOrgRole(models.OrgRoleAdmin)
	handler := roleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without membership context")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RequireOrgRole() without membership status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRequireOrgRole_InsufficientRole(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	// Create request with membership that has lower role
	membership := &models.OrganizationMember{
		ID:   1,
		Role: models.OrgRoleMember, // Member role
	}
	ctx := context.WithValue(context.Background(), MembershipContextKey, membership)
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Require Admin role (higher than Member)
	roleMiddleware := middleware.RequireOrgRole(models.OrgRoleAdmin)
	handler := roleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with insufficient role")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("RequireOrgRole() with insufficient role status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireOrgRole_SufficientRole(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	// Create request with membership that has sufficient role
	membership := &models.OrganizationMember{
		ID:   1,
		Role: models.OrgRoleOwner, // Owner role (highest)
	}
	ctx := context.WithValue(context.Background(), MembershipContextKey, membership)
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlerCalled := false
	roleMiddleware := middleware.RequireOrgRole(models.OrgRoleAdmin)
	handler := roleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called with sufficient role")
	}
	if w.Code != http.StatusOK {
		t.Errorf("RequireOrgRole() with sufficient role status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ============ OptionalOrganization Middleware Tests ============

func TestOptionalOrganization_NoUser(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := middleware.OptionalOrganization(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Should have no org context
		if GetOrganizationFromContext(r.Context()) != nil {
			t.Error("Should not have organization context without user")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even without user")
	}
}

func TestOptionalOrganization_NoSlug(t *testing.T) {
	middleware := NewTenantMiddleware(nil)

	// Create request with user but no org slug
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := context.WithValue(context.Background(), UserContextKey, user)
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := middleware.OptionalOrganization(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Should have no org context
		if GetOrganizationFromContext(r.Context()) != nil {
			t.Error("Should not have organization context without slug")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even without org slug")
	}
}

// ============ OrgScope Tests ============

func TestOrgScope(t *testing.T) {
	scope := OrgScope(123)
	if scope == nil {
		t.Error("OrgScope() returned nil")
	}
	// We can't fully test the scope without a real DB, but we verify it returns a function
}

// ============ Context Key Tests ============

func TestContextKeys(t *testing.T) {
	// Ensure context keys are unique
	if OrganizationContextKey == MembershipContextKey {
		t.Error("OrganizationContextKey and MembershipContextKey should be different")
	}

	// Ensure context keys have expected values
	if string(OrganizationContextKey) != "organization" {
		t.Errorf("OrganizationContextKey = %s, want 'organization'", OrganizationContextKey)
	}
	if string(MembershipContextKey) != "membership" {
		t.Errorf("MembershipContextKey = %s, want 'membership'", MembershipContextKey)
	}
}
