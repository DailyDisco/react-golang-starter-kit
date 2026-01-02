package auth

import (
	"context"
	"net/http"

	"react-golang-starter/internal/cache"
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
// Uses cache-first lookup to reduce database queries
func (m *TenantMiddleware) RequireOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get authenticated user
		user, ok := GetUserFromContext(ctx)
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

		// Try cache first for organization
		var org *models.Organization
		cachedOrg, _ := cache.GetOrganization(ctx, orgSlug)
		if cachedOrg != nil {
			org = cachedOrg
		} else {
			// Cache miss - query database
			var dbOrg models.Organization
			if err := m.db.Where("slug = ?", orgSlug).First(&dbOrg).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "Organization not found", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			org = &dbOrg
			// Cache the result
			cache.SetOrganization(ctx, org)
		}

		// Try cache first for membership
		var membership *models.OrganizationMember
		cachedMembership, _ := cache.GetMembership(ctx, org.ID, user.ID)
		if cachedMembership != nil && cachedMembership.Status == models.MemberStatusActive {
			membership = cachedMembership
		} else {
			// Cache miss or stale - query database
			var dbMembership models.OrganizationMember
			if err := m.db.Where("organization_id = ? AND user_id = ? AND status = ?",
				org.ID, user.ID, models.MemberStatusActive).First(&dbMembership).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "Not a member of this organization", http.StatusForbidden)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			membership = &dbMembership
			// Cache the result
			cache.SetMembership(ctx, membership)
		}

		// Add organization and membership to context
		ctx = context.WithValue(ctx, OrganizationContextKey, org)
		ctx = context.WithValue(ctx, MembershipContextKey, membership)

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
// Uses cache-first lookup to reduce database queries
func (m *TenantMiddleware) OptionalOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := GetUserFromContext(ctx)
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

		// Try cache first for organization
		var org *models.Organization
		cachedOrg, _ := cache.GetOrganization(ctx, orgSlug)
		if cachedOrg != nil {
			org = cachedOrg
		} else {
			// Cache miss - query database
			var dbOrg models.Organization
			if err := m.db.Where("slug = ?", orgSlug).First(&dbOrg).Error; err != nil {
				// Continue without org context
				next.ServeHTTP(w, r)
				return
			}
			org = &dbOrg
			// Cache the result
			cache.SetOrganization(ctx, org)
		}

		// Try cache first for membership
		var membership *models.OrganizationMember
		cachedMembership, _ := cache.GetMembership(ctx, org.ID, user.ID)
		if cachedMembership != nil && cachedMembership.Status == models.MemberStatusActive {
			membership = cachedMembership
		} else {
			// Cache miss - query database
			var dbMembership models.OrganizationMember
			if err := m.db.Where("organization_id = ? AND user_id = ? AND status = ?",
				org.ID, user.ID, models.MemberStatusActive).First(&dbMembership).Error; err != nil {
				// Continue without org context
				next.ServeHTTP(w, r)
				return
			}
			membership = &dbMembership
			// Cache the result
			cache.SetMembership(ctx, membership)
		}

		// Add organization and membership to context
		ctx = context.WithValue(ctx, OrganizationContextKey, org)
		ctx = context.WithValue(ctx, MembershipContextKey, membership)

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
