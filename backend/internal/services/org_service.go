package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/websocket"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrOrgNotFound          = errors.New("organization not found")
	ErrOrgSlugTaken         = errors.New("organization slug is already taken")
	ErrInvalidSlug          = errors.New("invalid slug format")
	ErrNotMember            = errors.New("user is not a member of this organization")
	ErrInsufficientRole     = errors.New("insufficient role permissions")
	ErrCannotRemoveOwner    = errors.New("cannot remove the organization owner")
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationExpired    = errors.New("invitation has expired")
	ErrInvitationAccepted   = errors.New("invitation has already been accepted")
	ErrAlreadyMember        = errors.New("user is already a member")
	ErrCannotChangeOwnRole  = errors.New("cannot change your own role")
	ErrMustHaveOwner        = errors.New("organization must have at least one owner")
	ErrInvitationEmailTaken = errors.New("an invitation for this email already exists")
	ErrSeatLimitExceeded    = errors.New("organization has reached its seat limit")
)

// slugRegex validates organization slugs
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// OrgService handles organization business logic
type OrgService struct {
	db  *gorm.DB
	hub *websocket.Hub
}

// NewOrgService creates a new organization service
func NewOrgService(db *gorm.DB) *OrgService {
	return &OrgService{db: db}
}

// SetHub sets the WebSocket hub for broadcasting org/member updates
func (s *OrgService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// broadcastToOrgMembers sends a WebSocket message to all members of an organization
func (s *OrgService) broadcastToOrgMembers(ctx context.Context, orgID uint, msgType websocket.MessageType, payload interface{}) {
	if s.hub == nil {
		return
	}

	var members []models.OrganizationMember
	if err := s.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&members).Error; err != nil {
		log.Warn().Err(err).Uint("org_id", orgID).Msg("failed to get org members for WebSocket broadcast")
		return
	}

	userIDs := make([]uint, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}

	s.hub.SendToUsers(userIDs, msgType, payload)
	log.Debug().Uint("org_id", orgID).Int("member_count", len(userIDs)).Str("msg_type", string(msgType)).Msg("broadcasted to org members")
}

// CreateOrganization creates a new organization with the user as owner
func (s *OrgService) CreateOrganization(ctx context.Context, userID uint, name, slug string) (*models.Organization, error) {
	// Validate slug
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugRegex.MatchString(slug) {
		return nil, ErrInvalidSlug
	}
	// Validate slug length (2-63 chars, matching DNS label limits)
	if len(slug) < 2 || len(slug) > 63 {
		return nil, ErrInvalidSlug
	}

	// Check if slug is taken
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Organization{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOrgSlugTaken
	}

	// Create organization and owner membership in transaction
	var org models.Organization
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		org = models.Organization{
			Name:            strings.TrimSpace(name),
			Slug:            slug,
			Plan:            models.OrgPlanFree,
			CreatedByUserID: userID,
		}

		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		// Create owner membership
		now := time.Now()
		member := models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         userID,
			Role:           models.OrgRoleOwner,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		}

		return tx.Create(&member).Error
	})

	if err != nil {
		return nil, err
	}

	return &org, nil
}

// GetOrganization retrieves an organization by slug
func (s *OrgService) GetOrganization(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.WithContext(ctx).Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// GetOrganizationByID retrieves an organization by ID
func (s *OrgService) GetOrganizationByID(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.WithContext(ctx).First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// GetOrganizationWithMembers retrieves an organization with its members
func (s *OrgService) GetOrganizationWithMembers(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.WithContext(ctx).Preload("Members.User").Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// OrgWithRole represents an organization with the user's role
type OrgWithRole struct {
	Organization models.Organization
	Role         models.OrganizationRole
}

// GetUserOrganizations returns all organizations a user is a member of
func (s *OrgService) GetUserOrganizations(ctx context.Context, userID uint) ([]models.Organization, error) {
	var orgs []models.Organization
	if err := s.db.WithContext(ctx).
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, models.MemberStatusActive).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetUserOrganizationsWithRoles returns all organizations with the user's role in each (single query, no N+1)
func (s *OrgService) GetUserOrganizationsWithRoles(ctx context.Context, userID uint) ([]OrgWithRole, error) {
	type result struct {
		models.Organization
		Role models.OrganizationRole `gorm:"column:member_role"`
	}

	var results []result
	if err := s.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*, organization_members.role as member_role").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, models.MemberStatusActive).
		Scan(&results).Error; err != nil {
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

// GetUserMembership returns a user's membership in an organization
func (s *OrgService) GetUserMembership(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	if err := s.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotMember
		}
		return nil, err
	}
	return &member, nil
}

// UpdateOrganization updates organization details
func (s *OrgService) UpdateOrganization(ctx context.Context, org *models.Organization, name string) error {
	org.Name = strings.TrimSpace(name)
	if err := s.db.WithContext(ctx).Save(org).Error; err != nil {
		return err
	}
	// Invalidate org cache after successful update
	_ = cache.InvalidateOrganization(ctx, org.Slug, org.ID)

	// Broadcast org update to all members
	s.broadcastToOrgMembers(ctx, org.ID, websocket.MessageTypeOrgUpdate, websocket.OrgUpdatePayload{
		OrgSlug: org.Slug,
		Event:   "settings_changed",
		Field:   "name",
	})

	return nil
}

// DeleteOrganization deletes an organization and all related data
func (s *OrgService) DeleteOrganization(ctx context.Context, org *models.Organization) error {
	// Broadcast deletion to all members before deleting (so they get notified)
	s.broadcastToOrgMembers(ctx, org.ID, websocket.MessageTypeOrgUpdate, websocket.OrgUpdatePayload{
		OrgSlug: org.Slug,
		Event:   "deleted",
	})

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete invitations
		if err := tx.Where("organization_id = ?", org.ID).Delete(&models.OrganizationInvitation{}).Error; err != nil {
			return err
		}

		// Delete members
		if err := tx.Where("organization_id = ?", org.ID).Delete(&models.OrganizationMember{}).Error; err != nil {
			return err
		}

		// Delete organization
		return tx.Delete(&models.Organization{}, org.ID).Error
	})
	if err != nil {
		return err
	}
	// Invalidate caches after successful deletion
	_ = cache.InvalidateOrganization(ctx, org.Slug, org.ID)
	_ = cache.InvalidateOrgMemberships(ctx, org.ID)
	return nil
}

// GetMembers returns all members of an organization
func (s *OrgService) GetMembers(ctx context.Context, orgID uint) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	if err := s.db.WithContext(ctx).Preload("User").Where("organization_id = ?", orgID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateMemberRole updates a member's role
func (s *OrgService) UpdateMemberRole(ctx context.Context, orgID, userID, actorUserID uint, newRole models.OrganizationRole) error {
	if userID == actorUserID {
		return ErrCannotChangeOwnRole
	}

	// Get target member
	var member models.OrganizationMember
	if err := s.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotMember
		}
		return err
	}

	// If demoting from owner, ensure there's another owner
	if member.Role == models.OrgRoleOwner && newRole != models.OrgRoleOwner {
		var ownerCount int64
		if err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND role = ?", orgID, models.OrgRoleOwner).
			Count(&ownerCount).Error; err != nil {
			return err
		}
		if ownerCount <= 1 {
			return ErrMustHaveOwner
		}
	}

	member.Role = newRole
	if err := s.db.WithContext(ctx).Save(&member).Error; err != nil {
		return err
	}

	// Invalidate membership cache after role update
	_ = cache.InvalidateMembership(ctx, orgID, userID)

	// Broadcast member update to all org members
	var org models.Organization
	if err := s.db.WithContext(ctx).First(&org, orgID).Error; err == nil {
		s.broadcastToOrgMembers(ctx, orgID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
			OrgSlug: org.Slug,
			Event:   "role_changed",
			UserID:  userID,
			Role:    string(newRole),
		})
	}

	return nil
}

// RemoveMember removes a member from an organization
func (s *OrgService) RemoveMember(ctx context.Context, orgID, userID uint) error {
	// Get member
	var member models.OrganizationMember
	if err := s.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotMember
		}
		return err
	}

	// Cannot remove owner if they're the only one
	if member.Role == models.OrgRoleOwner {
		var ownerCount int64
		if err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND role = ?", orgID, models.OrgRoleOwner).
			Count(&ownerCount).Error; err != nil {
			return err
		}
		if ownerCount <= 1 {
			return ErrCannotRemoveOwner
		}
	}

	if err := s.db.WithContext(ctx).Delete(&member).Error; err != nil {
		return err
	}

	// Invalidate membership cache after removal
	_ = cache.InvalidateMembership(ctx, orgID, userID)

	// Broadcast member removal to all org members (including the removed user)
	var org models.Organization
	if err := s.db.WithContext(ctx).First(&org, orgID).Error; err == nil {
		// First notify the org (remaining members will see updated list)
		s.broadcastToOrgMembers(ctx, orgID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
			OrgSlug: org.Slug,
			Event:   "removed",
			UserID:  userID,
		})
		// Also notify the removed user directly so their UI updates
		if s.hub != nil {
			s.hub.SendToUser(userID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
				OrgSlug: org.Slug,
				Event:   "removed",
				UserID:  userID,
			})
		}
	}

	return nil
}

// CreateInvitation creates a new invitation to join an organization
func (s *OrgService) CreateInvitation(ctx context.Context, orgID, inviterID uint, email string, role models.OrganizationRole) (*models.OrganizationInvitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check seat limit before inviting
	canAdd, err := s.CanAddMember(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if !canAdd {
		return nil, ErrSeatLimitExceeded
	}

	// Check if user is already a member
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err == nil {
		var count int64
		if err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND user_id = ?", orgID, user.ID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrAlreadyMember
		}
	}

	// Check for existing pending invitation
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("organization_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > ?", orgID, email, time.Now()).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrInvitationEmailTaken
	}

	// Generate token
	token, err := generateInvitationToken()
	if err != nil {
		return nil, err
	}

	invitation := models.OrganizationInvitation{
		OrganizationID:  orgID,
		Email:           email,
		Role:            role,
		Token:           token,
		InvitedByUserID: inviterID,
		ExpiresAt:       time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.db.WithContext(ctx).Create(&invitation).Error; err != nil {
		return nil, err
	}

	// Broadcast invitation sent to all org members
	var org models.Organization
	if err := s.db.WithContext(ctx).First(&org, orgID).Error; err == nil {
		s.broadcastToOrgMembers(ctx, orgID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
			OrgSlug: org.Slug,
			Event:   "invitation_sent",
		})
	}

	return &invitation, nil
}

// GetInvitationByToken retrieves an invitation by its token
func (s *OrgService) GetInvitationByToken(ctx context.Context, token string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	if err := s.db.WithContext(ctx).Preload("Organization").Where("token = ?", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *OrgService) AcceptInvitation(ctx context.Context, token string, userID uint) (*models.OrganizationMember, error) {
	invitation, err := s.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if invitation.IsExpired() {
		return nil, ErrInvitationExpired
	}

	if invitation.IsAccepted() {
		return nil, ErrInvitationAccepted
	}

	// Verify user email matches invitation
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	if strings.ToLower(user.Email) != invitation.Email {
		return nil, errors.New("email does not match invitation")
	}

	// Check if already a member
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", invitation.OrganizationID, userID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAlreadyMember
	}

	var member models.OrganizationMember
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Mark invitation as accepted
		now := time.Now()
		invitation.AcceptedAt = &now
		if err := tx.Save(invitation).Error; err != nil {
			return err
		}

		// Create membership
		member = models.OrganizationMember{
			OrganizationID:  invitation.OrganizationID,
			UserID:          userID,
			Role:            invitation.Role,
			Status:          models.MemberStatusActive,
			InvitedByUserID: &invitation.InvitedByUserID,
			AcceptedAt:      &now,
		}

		return tx.Create(&member).Error
	})

	if err != nil {
		return nil, err
	}

	// Broadcast new member added to all org members
	if invitation.Organization.Slug != "" {
		s.broadcastToOrgMembers(ctx, invitation.OrganizationID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
			OrgSlug: invitation.Organization.Slug,
			Event:   "added",
			UserID:  userID,
			Role:    string(invitation.Role),
		})
	}

	return &member, nil
}

// GetPendingInvitations returns all pending invitations for an organization
func (s *OrgService) GetPendingInvitations(ctx context.Context, orgID uint) ([]models.OrganizationInvitation, error) {
	var invitations []models.OrganizationInvitation
	if err := s.db.WithContext(ctx).Preload("InvitedByUser").
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, time.Now()).
		Find(&invitations).Error; err != nil {
		return nil, err
	}
	return invitations, nil
}

// CancelInvitation cancels a pending invitation
func (s *OrgService) CancelInvitation(ctx context.Context, invitationID, orgID uint) error {
	result := s.db.WithContext(ctx).Where("id = ? AND organization_id = ? AND accepted_at IS NULL", invitationID, orgID).
		Delete(&models.OrganizationInvitation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvitationNotFound
	}

	// Broadcast invitation revoked to all org members
	var org models.Organization
	if err := s.db.WithContext(ctx).First(&org, orgID).Error; err == nil {
		s.broadcastToOrgMembers(ctx, orgID, websocket.MessageTypeMemberUpdate, websocket.MemberUpdatePayload{
			OrgSlug: org.Slug,
			Event:   "invitation_revoked",
		})
	}

	return nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *OrgService) CleanupExpiredInvitations(ctx context.Context) error {
	return s.db.WithContext(ctx).Where("expires_at < ? AND accepted_at IS NULL", time.Now()).
		Delete(&models.OrganizationInvitation{}).Error
}

// generateInvitationToken generates a secure random token
func generateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// =====================
// Billing-related methods
// =====================

// GetOrganizationByStripeCustomerID retrieves an organization by its Stripe customer ID
func (s *OrgService) GetOrganizationByStripeCustomerID(ctx context.Context, customerID string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.WithContext(ctx).Where("stripe_customer_id = ?", customerID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// UpdateOrganizationPlan updates the organization's plan and Stripe subscription info
func (s *OrgService) UpdateOrganizationPlan(ctx context.Context, orgID uint, plan models.OrganizationPlan, stripeSubID *string) error {
	updates := map[string]interface{}{
		"plan": plan,
	}
	if stripeSubID != nil {
		updates["stripe_subscription_id"] = *stripeSubID
	}
	return s.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", orgID).Updates(updates).Error
}

// SetOrganizationStripeCustomer sets the Stripe customer ID for an organization
func (s *OrgService) SetOrganizationStripeCustomer(ctx context.Context, orgID uint, customerID string) error {
	return s.db.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", orgID).
		Update("stripe_customer_id", customerID).Error
}

// GetMemberCount returns the number of active members in an organization
func (s *OrgService) GetMemberCount(ctx context.Context, orgID uint) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND status = ?", orgID, models.MemberStatusActive).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CanAddMember checks if the organization can add another member based on seat limits
// Returns true if the organization has available seats or has unlimited seats
func (s *OrgService) CanAddMember(ctx context.Context, orgID uint) (bool, error) {
	org, err := s.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return false, err
	}

	seatLimit := org.GetSeatLimit()
	if seatLimit == 0 {
		// Unlimited seats (enterprise)
		return true, nil
	}

	memberCount, err := s.GetMemberCount(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Count pending invitations as well
	var inviteCount int64
	if err := s.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, time.Now()).
		Count(&inviteCount).Error; err != nil {
		return false, err
	}

	return int(memberCount)+int(inviteCount) < seatLimit, nil
}

// GetOrganizationSubscription retrieves the subscription for an organization
func (s *OrgService) GetOrganizationSubscription(ctx context.Context, orgID uint) (*models.Subscription, error) {
	var sub models.Subscription
	if err := s.db.WithContext(ctx).Where("organization_id = ?", orgID).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No subscription
		}
		return nil, err
	}
	return &sub, nil
}

// CreateOrganizationSubscription creates a subscription for an organization
func (s *OrgService) CreateOrganizationSubscription(ctx context.Context, sub *models.Subscription) error {
	return s.db.WithContext(ctx).Create(sub).Error
}

// UpdateOrganizationSubscription updates an organization's subscription
func (s *OrgService) UpdateOrganizationSubscription(ctx context.Context, sub *models.Subscription) error {
	return s.db.WithContext(ctx).Save(sub).Error
}

// GetOrganizationByStripeSubscriptionID retrieves an organization by its Stripe subscription ID
func (s *OrgService) GetOrganizationByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.WithContext(ctx).Where("stripe_subscription_id = ?", stripeSubID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}
