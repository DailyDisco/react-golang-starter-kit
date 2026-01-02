package models

import (
	"time"

	"gorm.io/datatypes"
)

// OrganizationRole represents the role of a user within an organization
type OrganizationRole string

const (
	OrgRoleOwner  OrganizationRole = "owner"
	OrgRoleAdmin  OrganizationRole = "admin"
	OrgRoleMember OrganizationRole = "member"
)

// OrganizationPlan represents the subscription plan of an organization
type OrganizationPlan string

const (
	OrgPlanFree       OrganizationPlan = "free"
	OrgPlanPro        OrganizationPlan = "pro"
	OrgPlanEnterprise OrganizationPlan = "enterprise"
)

// MemberStatus represents the status of an organization member
type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
	MemberStatusPending  MemberStatus = "pending"
)

// Organization represents a tenant/organization in the multi-tenant system
type Organization struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Basic info
	Name string `gorm:"not null;size:100" json:"name"`
	Slug string `gorm:"not null;uniqueIndex;size:100" json:"slug"`

	// Billing
	Plan                 OrganizationPlan `gorm:"type:varchar(20);default:'free'" json:"plan"`
	StripeCustomerID     *string          `gorm:"size:100" json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string          `gorm:"size:255" json:"stripe_subscription_id,omitempty"`
	PlanFeatures         datatypes.JSON   `gorm:"type:jsonb;default:'{}'" json:"plan_features"`

	// Settings stored as JSON
	Settings datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`

	// Creator
	CreatedByUserID uint  `gorm:"not null" json:"created_by_user_id"`
	CreatedByUser   *User `gorm:"foreignKey:CreatedByUserID" json:"created_by_user,omitempty"`

	// Relations
	Members     []OrganizationMember     `gorm:"foreignKey:OrganizationID" json:"members,omitempty"`
	Invitations []OrganizationInvitation `gorm:"foreignKey:OrganizationID" json:"invitations,omitempty"`
}

// TableName specifies the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	OrganizationID uint          `gorm:"not null;uniqueIndex:idx_org_member_unique" json:"organization_id"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	UserID         uint          `gorm:"not null;uniqueIndex:idx_org_member_unique" json:"user_id"`
	User           *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Role and status
	Role   OrganizationRole `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	Status MemberStatus     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`

	// Invitation tracking
	InvitedByUserID *uint `json:"invited_by_user_id,omitempty"`
	InvitedByUser   *User `gorm:"foreignKey:InvitedByUserID" json:"invited_by_user,omitempty"`

	// Timestamps
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// TableName specifies the table name for OrganizationMember
func (OrganizationMember) TableName() string {
	return "organization_members"
}

// OrganizationInvitation represents a pending invitation to join an organization
type OrganizationInvitation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Relationship
	OrganizationID uint          `gorm:"not null" json:"organization_id"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	// Invite details
	Email string           `gorm:"not null;size:255" json:"email"`
	Role  OrganizationRole `gorm:"type:varchar(50);not null;default:'member'" json:"role"`

	// Token for accepting invitation
	Token string `gorm:"not null;size:64;uniqueIndex" json:"-"` // Hidden from JSON

	// Who invited
	InvitedByUserID uint  `gorm:"not null" json:"invited_by_user_id"`
	InvitedByUser   *User `gorm:"foreignKey:InvitedByUserID" json:"invited_by_user,omitempty"`

	// Expiration and acceptance
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// TableName specifies the table name for OrganizationInvitation
func (OrganizationInvitation) TableName() string {
	return "organization_invitations"
}

// IsExpired checks if the invitation has expired
func (i *OrganizationInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (i *OrganizationInvitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// OrganizationSettings represents the JSON settings for an organization
type OrganizationSettings struct {
	AllowMemberInvites bool   `json:"allow_member_invites"`
	DefaultRole        string `json:"default_role"`
	MaxMembers         int    `json:"max_members"`
	Features           struct {
		AdvancedAnalytics bool `json:"advanced_analytics"`
		CustomBranding    bool `json:"custom_branding"`
		APIAccess         bool `json:"api_access"`
	} `json:"features"`
}

// Helper functions for role checking

// CanManageMembers returns true if the role can manage members
func (r OrganizationRole) CanManageMembers() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanManageSettings returns true if the role can manage organization settings
func (r OrganizationRole) CanManageSettings() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanDeleteOrganization returns true if the role can delete the organization
func (r OrganizationRole) CanDeleteOrganization() bool {
	return r == OrgRoleOwner
}

// CanTransferOwnership returns true if the role can transfer ownership
func (r OrganizationRole) CanTransferOwnership() bool {
	return r == OrgRoleOwner
}

// IsHigherOrEqualTo compares role hierarchy (owner > admin > member)
func (r OrganizationRole) IsHigherOrEqualTo(other OrganizationRole) bool {
	roleHierarchy := map[OrganizationRole]int{
		OrgRoleOwner:  3,
		OrgRoleAdmin:  2,
		OrgRoleMember: 1,
	}
	return roleHierarchy[r] >= roleHierarchy[other]
}

// PlanFeaturesData represents the plan-specific features for an organization
type PlanFeaturesData struct {
	SeatLimit    int `json:"seat_limit"`
	StorageLimit int `json:"storage_limit_mb"`
	APICallLimit int `json:"api_call_limit"`
}

// DefaultPlanFeatures returns default features for each plan
func DefaultPlanFeatures(plan OrganizationPlan) PlanFeaturesData {
	switch plan {
	case OrgPlanPro:
		return PlanFeaturesData{
			SeatLimit:    25,
			StorageLimit: 10240, // 10GB
			APICallLimit: 100000,
		}
	case OrgPlanEnterprise:
		return PlanFeaturesData{
			SeatLimit:    0, // unlimited
			StorageLimit: 0, // unlimited
			APICallLimit: 0, // unlimited
		}
	default: // Free
		return PlanFeaturesData{
			SeatLimit:    5,
			StorageLimit: 1024, // 1GB
			APICallLimit: 10000,
		}
	}
}

// GetSeatLimit returns the seat limit for this organization based on its plan
// Returns 0 for unlimited seats (enterprise)
func (o *Organization) GetSeatLimit() int {
	features := DefaultPlanFeatures(o.Plan)
	return features.SeatLimit
}

// HasSubscription returns true if the organization has an active subscription
func (o *Organization) HasSubscription() bool {
	return o.StripeSubscriptionID != nil && *o.StripeSubscriptionID != ""
}

// IsOrganizationSubscription checks if a subscription is org-level
func (s *Subscription) IsOrganizationSubscription() bool {
	return s.OrganizationID != nil && *s.OrganizationID > 0
}
