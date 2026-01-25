// Package repository defines data access interfaces for testable services.
// Services depend on these interfaces, not concrete implementations.
// This enables unit testing with mocks and integration testing with real databases.
package repository

import (
	"context"
	"time"

	"react-golang-starter/internal/models"
)

// SessionRepository defines data access operations for user sessions.
// Implementations: GormSessionRepository (production), MockSessionRepository (testing)
type SessionRepository interface {
	// Create creates a new session record.
	Create(ctx context.Context, session *models.UserSession) error

	// FindByUserID returns all sessions for a user that haven't expired.
	FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error)

	// DeleteByID deletes a session by ID and user ID (for authorization).
	DeleteByID(ctx context.Context, sessionID, userID uint) (int64, error)

	// DeleteByUserID deletes all sessions for a user, optionally excluding a token hash.
	DeleteByUserID(ctx context.Context, userID uint, exceptTokenHash string) error

	// DeleteByTokenHash deletes a session by its token hash.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// UpdateLastActive updates the last_active_at timestamp for a session.
	UpdateLastActive(ctx context.Context, tokenHash string, lastActive time.Time) error

	// DeleteExpired removes all sessions that have expired before the given time.
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// LoginHistoryRepository defines data access operations for login history.
type LoginHistoryRepository interface {
	// Create records a login attempt.
	Create(ctx context.Context, record *models.LoginHistory) error

	// FindByUserID returns login history for a user with pagination.
	FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.LoginHistory, error)

	// CountByUserID returns the total number of login records for a user.
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}

// UserRepository defines data access operations for users.
type UserRepository interface {
	// FindByID returns a user by ID.
	FindByID(ctx context.Context, id uint) (*models.User, error)

	// FindByEmail returns a user by email.
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// Create creates a new user.
	Create(ctx context.Context, user *models.User) error

	// Update updates an existing user.
	Update(ctx context.Context, user *models.User) error

	// Delete soft-deletes a user.
	Delete(ctx context.Context, id uint) error
}

// OrganizationRepository defines data access operations for organizations.
type OrganizationRepository interface {
	// FindBySlug returns an organization by its slug.
	FindBySlug(ctx context.Context, slug string) (*models.Organization, error)

	// FindBySlugWithMembers returns an organization with preloaded members.
	FindBySlugWithMembers(ctx context.Context, slug string) (*models.Organization, error)

	// FindByID returns an organization by ID.
	FindByID(ctx context.Context, id uint) (*models.Organization, error)

	// FindByStripeCustomerID returns an organization by Stripe customer ID.
	FindByStripeCustomerID(ctx context.Context, customerID string) (*models.Organization, error)

	// FindByStripeSubscriptionID returns an organization by Stripe subscription ID.
	FindByStripeSubscriptionID(ctx context.Context, subID string) (*models.Organization, error)

	// CountBySlug returns the count of organizations with the given slug.
	CountBySlug(ctx context.Context, slug string) (int64, error)

	// Create creates a new organization.
	Create(ctx context.Context, org *models.Organization) error

	// Update updates an existing organization.
	Update(ctx context.Context, org *models.Organization) error

	// UpdatePlan updates the organization's plan and optional subscription ID.
	UpdatePlan(ctx context.Context, orgID uint, plan models.OrganizationPlan, stripeSubID *string) error

	// UpdateStripeCustomer sets the Stripe customer ID for an organization.
	UpdateStripeCustomer(ctx context.Context, orgID uint, customerID string) error

	// Delete soft-deletes an organization.
	Delete(ctx context.Context, id uint) error
}

// OrganizationMemberRepository defines data access operations for organization members.
type OrganizationMemberRepository interface {
	// FindByOrgID returns all members of an organization with preloaded users.
	FindByOrgID(ctx context.Context, orgID uint) ([]models.OrganizationMember, error)

	// FindByOrgIDAndUserID returns a specific membership.
	FindByOrgIDAndUserID(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error)

	// FindOrgsByUserID returns all organizations a user is an active member of.
	FindOrgsByUserID(ctx context.Context, userID uint) ([]models.Organization, error)

	// FindOrgsWithRolesByUserID returns all organizations with the user's role.
	FindOrgsWithRolesByUserID(ctx context.Context, userID uint) ([]OrgWithRole, error)

	// CountByOrgIDAndRole counts members with a specific role.
	CountByOrgIDAndRole(ctx context.Context, orgID uint, role models.OrganizationRole) (int64, error)

	// CountActiveByOrgID counts active members in an organization.
	CountActiveByOrgID(ctx context.Context, orgID uint) (int64, error)

	// Create creates a new member.
	Create(ctx context.Context, member *models.OrganizationMember) error

	// Update updates a member's role.
	Update(ctx context.Context, member *models.OrganizationMember) error

	// Delete removes a member.
	Delete(ctx context.Context, member *models.OrganizationMember) error

	// DeleteByOrgID removes all members of an organization.
	DeleteByOrgID(ctx context.Context, orgID uint) error
}

// OrgWithRole represents an organization with the user's role.
type OrgWithRole struct {
	Organization models.Organization
	Role         models.OrganizationRole
}

// OrganizationInvitationRepository defines data access operations for invitations.
type OrganizationInvitationRepository interface {
	// FindByToken returns an invitation by token with preloaded organization.
	FindByToken(ctx context.Context, token string) (*models.OrganizationInvitation, error)

	// FindPendingByOrgID returns all pending invitations for an organization.
	FindPendingByOrgID(ctx context.Context, orgID uint, now time.Time) ([]models.OrganizationInvitation, error)

	// CountPendingByOrgIDAndEmail counts pending invitations for an email in an org.
	CountPendingByOrgIDAndEmail(ctx context.Context, orgID uint, email string, now time.Time) (int64, error)

	// CountPendingByOrgID counts all pending invitations for an organization.
	CountPendingByOrgID(ctx context.Context, orgID uint, now time.Time) (int64, error)

	// Create creates a new invitation.
	Create(ctx context.Context, invitation *models.OrganizationInvitation) error

	// Update updates an invitation (e.g., marking as accepted).
	Update(ctx context.Context, invitation *models.OrganizationInvitation) error

	// DeleteByIDAndOrgID cancels a pending invitation.
	DeleteByIDAndOrgID(ctx context.Context, id, orgID uint) (int64, error)

	// DeleteByOrgID removes all invitations for an organization.
	DeleteByOrgID(ctx context.Context, orgID uint) error

	// DeleteExpired removes all expired, unaccepted invitations.
	DeleteExpired(ctx context.Context, now time.Time) error
}

// SubscriptionRepository defines data access operations for subscriptions.
type SubscriptionRepository interface {
	// FindByOrgID returns the subscription for an organization.
	FindByOrgID(ctx context.Context, orgID uint) (*models.Subscription, error)

	// Create creates a new subscription.
	Create(ctx context.Context, sub *models.Subscription) error

	// Update updates a subscription.
	Update(ctx context.Context, sub *models.Subscription) error
}

// SystemSettingRepository defines data access operations for system settings.
type SystemSettingRepository interface {
	// FindAll returns all system settings.
	FindAll(ctx context.Context) ([]models.SystemSetting, error)

	// FindByCategory returns settings for a specific category.
	FindByCategory(ctx context.Context, category string) ([]models.SystemSetting, error)

	// FindByKey returns a single setting by key.
	FindByKey(ctx context.Context, key string) (*models.SystemSetting, error)

	// FindByKeys returns multiple settings by keys.
	FindByKeys(ctx context.Context, keys []string) ([]models.SystemSetting, error)

	// UpdateByKey updates a setting value by key.
	UpdateByKey(ctx context.Context, key string, value []byte, updatedAt string) (int64, error)
}

// IPBlocklistRepository defines data access operations for IP blocklist.
type IPBlocklistRepository interface {
	// FindActive returns all active IP blocks.
	FindActive(ctx context.Context) ([]models.IPBlocklist, error)

	// Create creates a new IP block entry.
	Create(ctx context.Context, block *models.IPBlocklist) error

	// Deactivate marks an IP block as inactive.
	Deactivate(ctx context.Context, id uint, updatedAt string) (int64, error)

	// IsBlocked checks if an IP is currently blocked.
	IsBlocked(ctx context.Context, ip string, now string) (bool, error)
}

// AnnouncementRepository defines data access operations for announcements.
type AnnouncementRepository interface {
	// FindAll returns all announcements.
	FindAll(ctx context.Context) ([]models.AnnouncementBanner, error)

	// FindByID returns an announcement by ID.
	FindByID(ctx context.Context, id uint) (*models.AnnouncementBanner, error)

	// FindActive returns active announcements for display.
	FindActive(ctx context.Context, now string) ([]models.AnnouncementBanner, error)

	// Create creates a new announcement.
	Create(ctx context.Context, announcement *models.AnnouncementBanner) error

	// Update updates an announcement.
	Update(ctx context.Context, id uint, updates map[string]interface{}) error

	// Delete deletes an announcement.
	Delete(ctx context.Context, id uint) (int64, error)

	// IncrementDismissCount increments the dismiss count for an announcement.
	IncrementDismissCount(ctx context.Context, id uint) error

	// IncrementViewCount increments the view count for an announcement.
	IncrementViewCount(ctx context.Context, id uint) error
}

// EmailTemplateRepository defines data access operations for email templates.
type EmailTemplateRepository interface {
	// FindAll returns all email templates.
	FindAll(ctx context.Context) ([]models.EmailTemplate, error)

	// FindByID returns an email template by ID.
	FindByID(ctx context.Context, id uint) (*models.EmailTemplate, error)

	// FindByKey returns an email template by key.
	FindByKey(ctx context.Context, key string) (*models.EmailTemplate, error)

	// Update updates an email template.
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
}

// UsageEventRepository defines data access operations for usage events.
type UsageEventRepository interface {
	// Create creates a new usage event.
	Create(ctx context.Context, event *models.UsageEvent) error
}

// UsagePeriodRepository defines data access operations for usage periods.
type UsagePeriodRepository interface {
	// FindByUserAndPeriod finds a usage period by user ID and period dates.
	FindByUserAndPeriod(ctx context.Context, userID uint, periodStart, periodEnd string) (*models.UsagePeriod, error)

	// FindByOrgAndPeriod finds a usage period by organization ID and period dates.
	FindByOrgAndPeriod(ctx context.Context, orgID uint, periodStart, periodEnd string) (*models.UsagePeriod, error)

	// FindHistoryByUser returns usage history for a user.
	FindHistoryByUser(ctx context.Context, userID uint, limit int) ([]models.UsagePeriod, error)

	// FindHistoryByOrg returns usage history for an organization.
	FindHistoryByOrg(ctx context.Context, orgID uint, limit int) ([]models.UsagePeriod, error)

	// Create creates a new usage period.
	Create(ctx context.Context, period *models.UsagePeriod) error

	// Update updates a usage period.
	Update(ctx context.Context, period *models.UsagePeriod, updates map[string]interface{}) error

	// Upsert creates or updates a usage period.
	Upsert(ctx context.Context, period *models.UsagePeriod) error
}

// UsageAlertRepository defines data access operations for usage alerts.
type UsageAlertRepository interface {
	// FindUnacknowledgedByUser returns unacknowledged alerts for a user.
	FindUnacknowledgedByUser(ctx context.Context, userID uint) ([]models.UsageAlert, error)

	// FindUnacknowledgedByOrg returns unacknowledged alerts for an organization.
	FindUnacknowledgedByOrg(ctx context.Context, orgID uint) ([]models.UsageAlert, error)

	// FindOrCreate finds an existing alert or creates a new one.
	FindOrCreate(ctx context.Context, alert *models.UsageAlert) (bool, error)

	// Acknowledge marks an alert as acknowledged.
	Acknowledge(ctx context.Context, alertID uint, acknowledgedBy uint, acknowledgedAt string) (int64, error)
}
