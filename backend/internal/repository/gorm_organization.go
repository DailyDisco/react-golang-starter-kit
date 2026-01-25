package repository

import (
	"context"
	"time"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// GormOrganizationRepository implements OrganizationRepository using GORM.
type GormOrganizationRepository struct {
	db *gorm.DB
}

// NewGormOrganizationRepository creates a new GORM-backed organization repository.
func NewGormOrganizationRepository(db *gorm.DB) *GormOrganizationRepository {
	return &GormOrganizationRepository{db: db}
}

// FindBySlug returns an organization by its slug.
func (r *GormOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindBySlugWithMembers returns an organization with preloaded members.
func (r *GormOrganizationRepository) FindBySlugWithMembers(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Preload("Members.User").Where("slug = ?", slug).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByID returns an organization by ID.
func (r *GormOrganizationRepository) FindByID(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByStripeCustomerID returns an organization by Stripe customer ID.
func (r *GormOrganizationRepository) FindByStripeCustomerID(ctx context.Context, customerID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("stripe_customer_id = ?", customerID).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByStripeSubscriptionID returns an organization by Stripe subscription ID.
func (r *GormOrganizationRepository) FindByStripeSubscriptionID(ctx context.Context, subID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("stripe_subscription_id = ?", subID).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// CountBySlug returns the count of organizations with the given slug.
func (r *GormOrganizationRepository) CountBySlug(ctx context.Context, slug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Organization{}).Where("slug = ?", slug).Count(&count).Error
	return count, err
}

// Create creates a new organization.
func (r *GormOrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// Update updates an existing organization.
func (r *GormOrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// UpdatePlan updates the organization's plan and optional subscription ID.
func (r *GormOrganizationRepository) UpdatePlan(ctx context.Context, orgID uint, plan models.OrganizationPlan, stripeSubID *string) error {
	updates := map[string]interface{}{
		"plan": plan,
	}
	if stripeSubID != nil {
		updates["stripe_subscription_id"] = *stripeSubID
	}
	return r.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", orgID).Updates(updates).Error
}

// UpdateStripeCustomer sets the Stripe customer ID for an organization.
func (r *GormOrganizationRepository) UpdateStripeCustomer(ctx context.Context, orgID uint, customerID string) error {
	return r.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", orgID).
		Update("stripe_customer_id", customerID).Error
}

// Delete soft-deletes an organization.
func (r *GormOrganizationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Organization{}, id).Error
}

// GormOrganizationMemberRepository implements OrganizationMemberRepository using GORM.
type GormOrganizationMemberRepository struct {
	db *gorm.DB
}

// NewGormOrganizationMemberRepository creates a new GORM-backed member repository.
func NewGormOrganizationMemberRepository(db *gorm.DB) *GormOrganizationMemberRepository {
	return &GormOrganizationMemberRepository{db: db}
}

// FindByOrgID returns all members of an organization with preloaded users.
func (r *GormOrganizationMemberRepository) FindByOrgID(ctx context.Context, orgID uint) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := r.db.WithContext(ctx).Preload("User").Where("organization_id = ?", orgID).Find(&members).Error
	return members, err
}

// FindByOrgIDAndUserID returns a specific membership.
func (r *GormOrganizationMemberRepository) FindByOrgIDAndUserID(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// FindOrgsByUserID returns all organizations a user is an active member of.
func (r *GormOrganizationMemberRepository) FindOrgsByUserID(ctx context.Context, userID uint) ([]models.Organization, error) {
	var orgs []models.Organization
	err := r.db.WithContext(ctx).
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, models.MemberStatusActive).
		Find(&orgs).Error
	return orgs, err
}

// FindOrgsWithRolesByUserID returns all organizations with the user's role.
func (r *GormOrganizationMemberRepository) FindOrgsWithRolesByUserID(ctx context.Context, userID uint) ([]OrgWithRole, error) {
	type result struct {
		models.Organization
		Role models.OrganizationRole `gorm:"column:member_role"`
	}

	var results []result
	err := r.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*, organization_members.role as member_role").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, models.MemberStatusActive).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	orgsWithRoles := make([]OrgWithRole, len(results))
	for i, r := range results {
		orgsWithRoles[i] = OrgWithRole{
			Organization: r.Organization,
			Role:         r.Role,
		}
	}
	return orgsWithRoles, nil
}

// CountByOrgIDAndRole counts members with a specific role.
func (r *GormOrganizationMemberRepository) CountByOrgIDAndRole(ctx context.Context, orgID uint, role models.OrganizationRole) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND role = ?", orgID, role).
		Count(&count).Error
	return count, err
}

// CountActiveByOrgID counts active members in an organization.
func (r *GormOrganizationMemberRepository) CountActiveByOrgID(ctx context.Context, orgID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND status = ?", orgID, models.MemberStatusActive).
		Count(&count).Error
	return count, err
}

// Create creates a new member.
func (r *GormOrganizationMemberRepository) Create(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// Update updates a member.
func (r *GormOrganizationMemberRepository) Update(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete removes a member.
func (r *GormOrganizationMemberRepository) Delete(ctx context.Context, member *models.OrganizationMember) error {
	return r.db.WithContext(ctx).Delete(member).Error
}

// DeleteByOrgID removes all members of an organization.
func (r *GormOrganizationMemberRepository) DeleteByOrgID(ctx context.Context, orgID uint) error {
	return r.db.WithContext(ctx).Where("organization_id = ?", orgID).Delete(&models.OrganizationMember{}).Error
}

// GormOrganizationInvitationRepository implements OrganizationInvitationRepository using GORM.
type GormOrganizationInvitationRepository struct {
	db *gorm.DB
}

// NewGormOrganizationInvitationRepository creates a new GORM-backed invitation repository.
func NewGormOrganizationInvitationRepository(db *gorm.DB) *GormOrganizationInvitationRepository {
	return &GormOrganizationInvitationRepository{db: db}
}

// FindByToken returns an invitation by token with preloaded organization.
func (r *GormOrganizationInvitationRepository) FindByToken(ctx context.Context, token string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).Preload("Organization").Where("token = ?", token).First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// FindPendingByOrgID returns all pending invitations for an organization.
func (r *GormOrganizationInvitationRepository) FindPendingByOrgID(ctx context.Context, orgID uint, now time.Time) ([]models.OrganizationInvitation, error) {
	var invitations []models.OrganizationInvitation
	err := r.db.WithContext(ctx).Preload("InvitedByUser").
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, now).
		Find(&invitations).Error
	return invitations, err
}

// CountPendingByOrgIDAndEmail counts pending invitations for an email in an org.
func (r *GormOrganizationInvitationRepository) CountPendingByOrgIDAndEmail(ctx context.Context, orgID uint, email string, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("organization_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > ?", orgID, email, now).
		Count(&count).Error
	return count, err
}

// CountPendingByOrgID counts all pending invitations for an organization.
func (r *GormOrganizationInvitationRepository) CountPendingByOrgID(ctx context.Context, orgID uint, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, now).
		Count(&count).Error
	return count, err
}

// Create creates a new invitation.
func (r *GormOrganizationInvitationRepository) Create(ctx context.Context, invitation *models.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// Update updates an invitation.
func (r *GormOrganizationInvitationRepository) Update(ctx context.Context, invitation *models.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// DeleteByIDAndOrgID cancels a pending invitation.
func (r *GormOrganizationInvitationRepository) DeleteByIDAndOrgID(ctx context.Context, id, orgID uint) (int64, error) {
	result := r.db.WithContext(ctx).Where("id = ? AND organization_id = ? AND accepted_at IS NULL", id, orgID).
		Delete(&models.OrganizationInvitation{})
	return result.RowsAffected, result.Error
}

// DeleteByOrgID removes all invitations for an organization.
func (r *GormOrganizationInvitationRepository) DeleteByOrgID(ctx context.Context, orgID uint) error {
	return r.db.WithContext(ctx).Where("organization_id = ?", orgID).Delete(&models.OrganizationInvitation{}).Error
}

// DeleteExpired removes all expired, unaccepted invitations.
func (r *GormOrganizationInvitationRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ? AND accepted_at IS NULL", now).
		Delete(&models.OrganizationInvitation{}).Error
}

// GormSubscriptionRepository implements SubscriptionRepository using GORM.
type GormSubscriptionRepository struct {
	db *gorm.DB
}

// NewGormSubscriptionRepository creates a new GORM-backed subscription repository.
func NewGormSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

// FindByOrgID returns the subscription for an organization.
func (r *GormSubscriptionRepository) FindByOrgID(ctx context.Context, orgID uint) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// Create creates a new subscription.
func (r *GormSubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// Update updates a subscription.
func (r *GormSubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// GormUserRepository implements UserRepository using GORM.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM-backed user repository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByID returns a user by ID.
func (r *GormUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail returns a user by email.
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user.
func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates an existing user.
func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft-deletes a user.
func (r *GormUserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
