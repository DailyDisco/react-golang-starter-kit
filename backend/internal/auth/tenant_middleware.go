package auth

import (
	"context"
	"net/http"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

type contextKey string

const (
	// OrganizationContextKey is the context key for the current organization
	OrganizationContextKey contextKey = "organization"
	// MembershipContextKey is the context key for the current user's membership
	MembershipContextKey contextKey = "membership"
)

// TenantMiddleware extracts organization context from the request
// It looks for the organization slug in URL path or X-Organization-Slug header
type TenantMiddleware struct {
	db *gorm.DB
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(db *gorm.DB) *TenantMiddleware {
	return &TenantMiddleware{db: db}
}

// RequireOrganization middleware requires a valid organization context
// It extracts the org slug from chi URL param "orgSlug" or X-Organization-Slug header
func (m *TenantMiddleware) RequireOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get organization slug from URL param or header
		orgSlug := r.PathValue("orgSlug")
		if orgSlug == "" {
			orgSlug = r.Header.Get("X-Organization-Slug")
		}

		if orgSlug == "" {
			http.Error(w, "Organization slug required", http.StatusBadRequest)
			return
		}

		// Find organization
		var org models.Organization
		if err := m.db.Where("slug = ?", orgSlug).First(&org).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check membership
		var membership models.OrganizationMember
		if err := m.db.Where("organization_id = ? AND user_id = ? AND status = ?",
			org.ID, user.ID, models.MemberStatusActive).First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Not a member of this organization", http.StatusForbidden)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Add organization and membership to context
		ctx := context.WithValue(r.Context(), OrganizationContextKey, &org)
		ctx = context.WithValue(ctx, MembershipContextKey, &membership)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireOrgRole middleware requires a minimum role within the organization
func (m *TenantMiddleware) RequireOrgRole(minRole models.OrganizationRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			membership := GetMembershipFromContext(r.Context())
			if membership == nil {
				http.Error(w, "Organization context required", http.StatusBadRequest)
				return
			}

			if !membership.Role.IsHigherOrEqualTo(minRole) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalOrganization middleware optionally extracts organization context
// Use this when org context is optional (e.g., listing all user's organizations)
func (m *TenantMiddleware) OptionalOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get organization slug from URL param or header
		orgSlug := r.PathValue("orgSlug")
		if orgSlug == "" {
			orgSlug = r.Header.Get("X-Organization-Slug")
		}

		if orgSlug == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Find organization
		var org models.Organization
		if err := m.db.Where("slug = ?", orgSlug).First(&org).Error; err != nil {
			// Continue without org context
			next.ServeHTTP(w, r)
			return
		}

		// Check membership
		var membership models.OrganizationMember
		if err := m.db.Where("organization_id = ? AND user_id = ? AND status = ?",
			org.ID, user.ID, models.MemberStatusActive).First(&membership).Error; err != nil {
			// Continue without org context
			next.ServeHTTP(w, r)
			return
		}

		// Add organization and membership to context
		ctx := context.WithValue(r.Context(), OrganizationContextKey, &org)
		ctx = context.WithValue(ctx, MembershipContextKey, &membership)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetOrganizationFromContext extracts the organization from the request context
func GetOrganizationFromContext(ctx context.Context) *models.Organization {
	org, ok := ctx.Value(OrganizationContextKey).(*models.Organization)
	if !ok {
		return nil
	}
	return org
}

// GetMembershipFromContext extracts the membership from the request context
func GetMembershipFromContext(ctx context.Context) *models.OrganizationMember {
	membership, ok := ctx.Value(MembershipContextKey).(*models.OrganizationMember)
	if !ok {
		return nil
	}
	return membership
}

// OrgScope returns a GORM scope that filters by organization ID
// Use this to ensure queries are scoped to the current organization
func OrgScope(orgID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("organization_id = ?", orgID)
	}
}
