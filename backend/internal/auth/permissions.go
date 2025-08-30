package auth

import (
	"net/http"

	"react-golang-starter/internal/models"
)

// Permission represents a specific action a user can perform
type Permission string

const (
	// User management permissions
	PermViewUsers   Permission = "users:view"
	PermCreateUsers Permission = "users:create"
	PermUpdateUsers Permission = "users:update"
	PermDeleteUsers Permission = "users:delete"
	PermManageRoles Permission = "users:manage_roles"

	// Content permissions
	PermViewPremium   Permission = "content:premium"
	PermManageContent Permission = "content:manage"

	// System permissions
	PermSystemAdmin Permission = "system:admin"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]Permission{
	models.RoleSuperAdmin: {
		PermViewUsers, PermCreateUsers, PermUpdateUsers, PermDeleteUsers,
		PermManageRoles, PermViewPremium, PermManageContent, PermSystemAdmin,
	},
	models.RoleAdmin: {
		PermViewUsers, PermUpdateUsers, PermViewPremium, PermManageContent,
	},
	models.RolePremium: {
		PermViewPremium,
	},
	models.RoleUser: {
		// Basic user permissions only
	},
}

// HasPermission checks if a user's role has the required permission
func HasPermission(userRole string, requiredPerm Permission) bool {
	permissions, exists := RolePermissions[userRole]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == requiredPerm {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if user has any of the required permissions
func HasAnyPermission(userRole string, requiredPerms ...Permission) bool {
	for _, perm := range requiredPerms {
		if HasPermission(userRole, perm) {
			return true
		}
	}
	return false
}

// HasRole checks if user has one of the required roles
func HasRole(userRole string, requiredRoles ...string) bool {
	for _, role := range requiredRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// PermissionMiddleware creates middleware for specific permissions
func PermissionMiddleware(requiredPerms ...Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized: User role not found", http.StatusUnauthorized)
				return
			}

			if !HasAnyPermission(userRole, requiredPerms...) {
				http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware creates middleware for specific roles
func RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized: User role not found", http.StatusUnauthorized)
				return
			}

			if !HasRole(userRole, requiredRoles...) {
				http.Error(w, "Forbidden: Insufficient role permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
