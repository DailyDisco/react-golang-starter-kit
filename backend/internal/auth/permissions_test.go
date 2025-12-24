package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ HasPermission Tests ============

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		perm     Permission
		expected bool
	}{
		// Super admin has all permissions
		{"super_admin has system_admin", models.RoleSuperAdmin, PermSystemAdmin, true},
		{"super_admin has view_users", models.RoleSuperAdmin, PermViewUsers, true},
		{"super_admin has create_users", models.RoleSuperAdmin, PermCreateUsers, true},
		{"super_admin has update_users", models.RoleSuperAdmin, PermUpdateUsers, true},
		{"super_admin has delete_users", models.RoleSuperAdmin, PermDeleteUsers, true},
		{"super_admin has manage_roles", models.RoleSuperAdmin, PermManageRoles, true},
		{"super_admin has view_premium", models.RoleSuperAdmin, PermViewPremium, true},
		{"super_admin has manage_content", models.RoleSuperAdmin, PermManageContent, true},

		// Admin permissions
		{"admin has view_users", models.RoleAdmin, PermViewUsers, true},
		{"admin has update_users", models.RoleAdmin, PermUpdateUsers, true},
		{"admin has view_premium", models.RoleAdmin, PermViewPremium, true},
		{"admin has manage_content", models.RoleAdmin, PermManageContent, true},
		{"admin lacks system_admin", models.RoleAdmin, PermSystemAdmin, false},
		{"admin lacks create_users", models.RoleAdmin, PermCreateUsers, false},
		{"admin lacks delete_users", models.RoleAdmin, PermDeleteUsers, false},
		{"admin lacks manage_roles", models.RoleAdmin, PermManageRoles, false},

		// Premium user permissions
		{"premium has view_premium", models.RolePremium, PermViewPremium, true},
		{"premium lacks view_users", models.RolePremium, PermViewUsers, false},
		{"premium lacks manage_content", models.RolePremium, PermManageContent, false},
		{"premium lacks system_admin", models.RolePremium, PermSystemAdmin, false},

		// Basic user permissions
		{"user lacks view_premium", models.RoleUser, PermViewPremium, false},
		{"user lacks view_users", models.RoleUser, PermViewUsers, false},
		{"user lacks system_admin", models.RoleUser, PermSystemAdmin, false},

		// Unknown role
		{"unknown role has no permissions", "unknown_role", PermViewUsers, false},
		{"empty role has no permissions", "", PermViewUsers, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(tt.role, tt.perm)
			if got != tt.expected {
				t.Errorf("HasPermission(%q, %q) = %v, want %v", tt.role, tt.perm, got, tt.expected)
			}
		})
	}
}

// ============ HasAnyPermission Tests ============

func TestHasAnyPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		perms    []Permission
		expected bool
	}{
		{
			"super_admin has any of multiple perms",
			models.RoleSuperAdmin,
			[]Permission{PermSystemAdmin, PermViewUsers},
			true,
		},
		{
			"admin has one of multiple perms",
			models.RoleAdmin,
			[]Permission{PermSystemAdmin, PermViewUsers},
			true, // has PermViewUsers
		},
		{
			"admin lacks all required perms",
			models.RoleAdmin,
			[]Permission{PermSystemAdmin, PermDeleteUsers},
			false,
		},
		{
			"premium has view_premium from list",
			models.RolePremium,
			[]Permission{PermViewPremium, PermManageContent},
			true,
		},
		{
			"user has none of the perms",
			models.RoleUser,
			[]Permission{PermViewPremium, PermViewUsers},
			false,
		},
		{
			"empty perms list returns false",
			models.RoleSuperAdmin,
			[]Permission{},
			false,
		},
		{
			"unknown role with valid perms returns false",
			"unknown",
			[]Permission{PermViewUsers},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasAnyPermission(tt.role, tt.perms...)
			if got != tt.expected {
				t.Errorf("HasAnyPermission(%q, %v) = %v, want %v", tt.role, tt.perms, got, tt.expected)
			}
		})
	}
}

// ============ HasRole Tests ============

func TestHasRole(t *testing.T) {
	tests := []struct {
		name          string
		userRole      string
		requiredRoles []string
		expected      bool
	}{
		{
			"exact role match",
			models.RoleAdmin,
			[]string{models.RoleAdmin},
			true,
		},
		{
			"role in list",
			models.RoleAdmin,
			[]string{models.RoleSuperAdmin, models.RoleAdmin},
			true,
		},
		{
			"role not in list",
			models.RoleUser,
			[]string{models.RoleSuperAdmin, models.RoleAdmin},
			false,
		},
		{
			"empty required roles",
			models.RoleAdmin,
			[]string{},
			false,
		},
		{
			"unknown role",
			"unknown",
			[]string{models.RoleAdmin},
			false,
		},
		{
			"empty user role",
			"",
			[]string{models.RoleAdmin},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRole(tt.userRole, tt.requiredRoles...)
			if got != tt.expected {
				t.Errorf("HasRole(%q, %v) = %v, want %v", tt.userRole, tt.requiredRoles, got, tt.expected)
			}
		})
	}
}

// ============ PermissionMiddleware Tests ============

func TestPermissionMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		hasRoleContext bool
		requiredPerms  []Permission
		expectedStatus int
	}{
		{
			"super_admin with system_admin perm passes",
			models.RoleSuperAdmin,
			true,
			[]Permission{PermSystemAdmin},
			http.StatusOK,
		},
		{
			"admin with view_users perm passes",
			models.RoleAdmin,
			true,
			[]Permission{PermViewUsers},
			http.StatusOK,
		},
		{
			"admin without system_admin perm forbidden",
			models.RoleAdmin,
			true,
			[]Permission{PermSystemAdmin},
			http.StatusForbidden,
		},
		{
			"user without any perms forbidden",
			models.RoleUser,
			true,
			[]Permission{PermViewPremium},
			http.StatusForbidden,
		},
		{
			"no role in context unauthorized",
			"",
			false,
			[]Permission{PermViewUsers},
			http.StatusUnauthorized,
		},
		{
			"multiple perms - has one passes",
			models.RoleAdmin,
			true,
			[]Permission{PermSystemAdmin, PermViewUsers},
			http.StatusOK,
		},
		{
			"multiple perms - has none forbidden",
			models.RoleUser,
			true,
			[]Permission{PermSystemAdmin, PermViewUsers},
			http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that returns 200
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with permission middleware
			middleware := PermissionMiddleware(tt.requiredPerms...)
			wrappedHandler := middleware(handler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Add role to context if specified
			if tt.hasRoleContext {
				ctx := context.WithValue(req.Context(), UserRoleContextKey, tt.userRole)
				req = req.WithContext(ctx)
			}

			// Execute
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("PermissionMiddleware returned status %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

// ============ RoleMiddleware Tests ============

func TestRoleMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		hasRoleContext bool
		requiredRoles  []string
		expectedStatus int
	}{
		{
			"admin with admin role passes",
			models.RoleAdmin,
			true,
			[]string{models.RoleAdmin},
			http.StatusOK,
		},
		{
			"super_admin with admin or super_admin passes",
			models.RoleSuperAdmin,
			true,
			[]string{models.RoleAdmin, models.RoleSuperAdmin},
			http.StatusOK,
		},
		{
			"user without admin role forbidden",
			models.RoleUser,
			true,
			[]string{models.RoleAdmin},
			http.StatusForbidden,
		},
		{
			"premium without admin role forbidden",
			models.RolePremium,
			true,
			[]string{models.RoleAdmin, models.RoleSuperAdmin},
			http.StatusForbidden,
		},
		{
			"no role in context unauthorized",
			"",
			false,
			[]string{models.RoleAdmin},
			http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that returns 200
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with role middleware
			middleware := RoleMiddleware(tt.requiredRoles...)
			wrappedHandler := middleware(handler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Add role to context if specified
			if tt.hasRoleContext {
				ctx := context.WithValue(req.Context(), UserRoleContextKey, tt.userRole)
				req = req.WithContext(ctx)
			}

			// Execute
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("RoleMiddleware returned status %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

// ============ Permission Constants Tests ============

func TestPermissionConstants(t *testing.T) {
	// Ensure permission constants are defined correctly
	permissions := []Permission{
		PermViewUsers,
		PermCreateUsers,
		PermUpdateUsers,
		PermDeleteUsers,
		PermManageRoles,
		PermViewPremium,
		PermManageContent,
		PermSystemAdmin,
	}

	for _, perm := range permissions {
		if perm == "" {
			t.Error("Found empty permission constant")
		}
	}
}

// ============ RolePermissions Map Tests ============

func TestRolePermissionsMap(t *testing.T) {
	// Ensure all roles are defined in the map
	requiredRoles := []string{
		models.RoleSuperAdmin,
		models.RoleAdmin,
		models.RolePremium,
		models.RoleUser,
	}

	for _, role := range requiredRoles {
		if _, exists := RolePermissions[role]; !exists {
			t.Errorf("Role %q not found in RolePermissions map", role)
		}
	}

	// Super admin should have all permissions
	superAdminPerms := RolePermissions[models.RoleSuperAdmin]
	if len(superAdminPerms) != 8 {
		t.Errorf("Super admin should have 8 permissions, got %d", len(superAdminPerms))
	}

	// User should have no special permissions
	userPerms := RolePermissions[models.RoleUser]
	if len(userPerms) != 0 {
		t.Errorf("Basic user should have 0 permissions, got %d", len(userPerms))
	}
}
