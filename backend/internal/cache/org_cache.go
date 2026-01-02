package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

const (
	// OrgBySlugKeyPrefix is the cache key prefix for organizations by slug
	OrgBySlugKeyPrefix = "org:slug:"
	// OrgByIDKeyPrefix is the cache key prefix for organizations by ID
	OrgByIDKeyPrefix = "org:id:"
	// MembershipKeyPrefix is the cache key prefix for memberships
	MembershipKeyPrefix = "membership:"
)

// Default TTLs (used if config is not available)
var (
	defaultOrgTTL        = 5 * time.Minute
	defaultMembershipTTL = 5 * time.Minute
)

// orgCacheConfig holds the configured TTLs
var orgCacheConfig *Config

// SetOrgCacheConfig sets the cache configuration for org caching
func SetOrgCacheConfig(config *Config) {
	orgCacheConfig = config
}

// getOrgTTL returns the organization cache TTL
func getOrgTTL() time.Duration {
	if orgCacheConfig != nil && orgCacheConfig.OrganizationTTL > 0 {
		return orgCacheConfig.OrganizationTTL
	}
	return defaultOrgTTL
}

// getMembershipTTL returns the membership cache TTL
func getMembershipTTL() time.Duration {
	if orgCacheConfig != nil && orgCacheConfig.MembershipTTL > 0 {
		return orgCacheConfig.MembershipTTL
	}
	return defaultMembershipTTL
}

// orgSlugKey generates a cache key for an organization by slug
func orgSlugKey(slug string) string {
	return OrgBySlugKeyPrefix + slug
}

// orgIDKey generates a cache key for an organization by ID
func orgIDKey(id uint) string {
	return OrgByIDKeyPrefix + strconv.FormatUint(uint64(id), 10)
}

// membershipKey generates a cache key for a membership
func membershipKey(orgID, userID uint) string {
	return fmt.Sprintf("%s%d:%d", MembershipKeyPrefix, orgID, userID)
}

// GetOrganization retrieves an organization from the cache by slug
// Returns nil, nil if the organization is not in the cache (cache miss)
func GetOrganization(ctx context.Context, slug string) (*models.Organization, error) {
	if !IsAvailable() {
		return nil, nil
	}

	var org models.Organization
	err := GetJSON(ctx, orgSlugKey(slug), &org)
	if err != nil {
		// Cache miss or error - return nil to indicate caller should query DB
		return nil, nil
	}

	log.Debug().Str("slug", slug).Msg("organization cache hit")
	return &org, nil
}

// GetOrganizationByID retrieves an organization from the cache by ID
// Returns nil, nil if the organization is not in the cache (cache miss)
func GetOrganizationByID(ctx context.Context, id uint) (*models.Organization, error) {
	if !IsAvailable() {
		return nil, nil
	}

	var org models.Organization
	err := GetJSON(ctx, orgIDKey(id), &org)
	if err != nil {
		return nil, nil
	}

	log.Debug().Uint("id", id).Msg("organization cache hit by ID")
	return &org, nil
}

// SetOrganization caches an organization by both slug and ID
func SetOrganization(ctx context.Context, org *models.Organization) error {
	if !IsAvailable() || org == nil {
		return nil
	}

	ttl := getOrgTTL()

	// Cache by slug
	if err := SetJSON(ctx, orgSlugKey(org.Slug), org, ttl); err != nil {
		log.Warn().Err(err).Str("slug", org.Slug).Msg("failed to cache organization by slug")
		return err
	}

	// Also cache by ID for quick lookups
	if err := SetJSON(ctx, orgIDKey(org.ID), org, ttl); err != nil {
		log.Warn().Err(err).Uint("id", org.ID).Msg("failed to cache organization by ID")
		// Don't return error - slug cache succeeded
	}

	log.Debug().Str("slug", org.Slug).Uint("id", org.ID).Msg("organization cached")
	return nil
}

// GetMembership retrieves a membership from the cache
// Returns nil, nil if the membership is not in the cache (cache miss)
func GetMembership(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error) {
	if !IsAvailable() {
		return nil, nil
	}

	var membership models.OrganizationMember
	err := GetJSON(ctx, membershipKey(orgID, userID), &membership)
	if err != nil {
		return nil, nil
	}

	log.Debug().Uint("orgID", orgID).Uint("userID", userID).Msg("membership cache hit")
	return &membership, nil
}

// SetMembership caches a membership
func SetMembership(ctx context.Context, membership *models.OrganizationMember) error {
	if !IsAvailable() || membership == nil {
		return nil
	}

	ttl := getMembershipTTL()

	if err := SetJSON(ctx, membershipKey(membership.OrganizationID, membership.UserID), membership, ttl); err != nil {
		log.Warn().Err(err).
			Uint("orgID", membership.OrganizationID).
			Uint("userID", membership.UserID).
			Msg("failed to cache membership")
		return err
	}

	log.Debug().
		Uint("orgID", membership.OrganizationID).
		Uint("userID", membership.UserID).
		Msg("membership cached")
	return nil
}

// InvalidateOrganization removes an organization from the cache
func InvalidateOrganization(ctx context.Context, slug string, id uint) error {
	if !IsAvailable() {
		return nil
	}

	var lastErr error

	// Delete by slug
	if slug != "" {
		if err := Delete(ctx, orgSlugKey(slug)); err != nil {
			log.Warn().Err(err).Str("slug", slug).Msg("failed to invalidate org cache by slug")
			lastErr = err
		}
	}

	// Delete by ID
	if id > 0 {
		if err := Delete(ctx, orgIDKey(id)); err != nil {
			log.Warn().Err(err).Uint("id", id).Msg("failed to invalidate org cache by ID")
			lastErr = err
		}
	}

	log.Debug().Str("slug", slug).Uint("id", id).Msg("organization cache invalidated")
	return lastErr
}

// InvalidateMembership removes a membership from the cache
func InvalidateMembership(ctx context.Context, orgID, userID uint) error {
	if !IsAvailable() {
		return nil
	}

	if err := Delete(ctx, membershipKey(orgID, userID)); err != nil {
		log.Warn().Err(err).
			Uint("orgID", orgID).
			Uint("userID", userID).
			Msg("failed to invalidate membership cache")
		return err
	}

	log.Debug().Uint("orgID", orgID).Uint("userID", userID).Msg("membership cache invalidated")
	return nil
}

// InvalidateOrgMemberships invalidates all membership caches for an organization
// This uses pattern matching which may not be supported by all cache implementations
func InvalidateOrgMemberships(ctx context.Context, orgID uint) error {
	if !IsAvailable() {
		return nil
	}

	pattern := fmt.Sprintf("%s%d:*", MembershipKeyPrefix, orgID)
	if inst := Instance(); inst != nil {
		if err := inst.Clear(ctx, pattern); err != nil {
			log.Warn().Err(err).Uint("orgID", orgID).Msg("failed to invalidate org memberships")
			return err
		}
	}

	log.Debug().Uint("orgID", orgID).Msg("organization memberships cache invalidated")
	return nil
}
