package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"react-golang-starter/internal/models"

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
)

// slugRegex validates organization slugs
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// OrgService handles organization business logic
type OrgService struct {
	db *gorm.DB
}

// NewOrgService creates a new organization service
func NewOrgService(db *gorm.DB) *OrgService {
	return &OrgService{db: db}
}

// CreateOrganization creates a new organization with the user as owner
func (s *OrgService) CreateOrganization(userID uint, name, slug string) (*models.Organization, error) {
	// Validate slug
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugRegex.MatchString(slug) {
		return nil, ErrInvalidSlug
	}

	// Check if slug is taken
	var count int64
	if err := s.db.Model(&models.Organization{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOrgSlugTaken
	}

	// Create organization and owner membership in transaction
	var org models.Organization
	err := s.db.Transaction(func(tx *gorm.DB) error {
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
func (s *OrgService) GetOrganization(slug string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// GetOrganizationWithMembers retrieves an organization with its members
func (s *OrgService) GetOrganizationWithMembers(slug string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.Preload("Members.User").Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// GetUserOrganizations returns all organizations a user is a member of
func (s *OrgService) GetUserOrganizations(userID uint) ([]models.Organization, error) {
	var orgs []models.Organization
	if err := s.db.
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, models.MemberStatusActive).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetUserMembership returns a user's membership in an organization
func (s *OrgService) GetUserMembership(orgID, userID uint) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotMember
		}
		return nil, err
	}
	return &member, nil
}

// UpdateOrganization updates organization details
func (s *OrgService) UpdateOrganization(org *models.Organization, name string) error {
	org.Name = strings.TrimSpace(name)
	return s.db.Save(org).Error
}

// DeleteOrganization deletes an organization and all related data
func (s *OrgService) DeleteOrganization(orgID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete invitations
		if err := tx.Where("organization_id = ?", orgID).Delete(&models.OrganizationInvitation{}).Error; err != nil {
			return err
		}

		// Delete members
		if err := tx.Where("organization_id = ?", orgID).Delete(&models.OrganizationMember{}).Error; err != nil {
			return err
		}

		// Delete organization
		return tx.Delete(&models.Organization{}, orgID).Error
	})
}

// GetMembers returns all members of an organization
func (s *OrgService) GetMembers(orgID uint) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	if err := s.db.Preload("User").Where("organization_id = ?", orgID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateMemberRole updates a member's role
func (s *OrgService) UpdateMemberRole(orgID, userID, actorUserID uint, newRole models.OrganizationRole) error {
	if userID == actorUserID {
		return ErrCannotChangeOwnRole
	}

	// Get target member
	var member models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotMember
		}
		return err
	}

	// If demoting from owner, ensure there's another owner
	if member.Role == models.OrgRoleOwner && newRole != models.OrgRoleOwner {
		var ownerCount int64
		if err := s.db.Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND role = ?", orgID, models.OrgRoleOwner).
			Count(&ownerCount).Error; err != nil {
			return err
		}
		if ownerCount <= 1 {
			return ErrMustHaveOwner
		}
	}

	member.Role = newRole
	return s.db.Save(&member).Error
}

// RemoveMember removes a member from an organization
func (s *OrgService) RemoveMember(orgID, userID uint) error {
	// Get member
	var member models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotMember
		}
		return err
	}

	// Cannot remove owner if they're the only one
	if member.Role == models.OrgRoleOwner {
		var ownerCount int64
		if err := s.db.Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND role = ?", orgID, models.OrgRoleOwner).
			Count(&ownerCount).Error; err != nil {
			return err
		}
		if ownerCount <= 1 {
			return ErrCannotRemoveOwner
		}
	}

	return s.db.Delete(&member).Error
}

// CreateInvitation creates a new invitation to join an organization
func (s *OrgService) CreateInvitation(orgID, inviterID uint, email string, role models.OrganizationRole) (*models.OrganizationInvitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if user is already a member
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err == nil {
		var count int64
		if err := s.db.Model(&models.OrganizationMember{}).
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
	if err := s.db.Model(&models.OrganizationInvitation{}).
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

	if err := s.db.Create(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

// GetInvitationByToken retrieves an invitation by its token
func (s *OrgService) GetInvitationByToken(token string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	if err := s.db.Preload("Organization").Where("token = ?", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *OrgService) AcceptInvitation(token string, userID uint) (*models.OrganizationMember, error) {
	invitation, err := s.GetInvitationByToken(token)
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
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if strings.ToLower(user.Email) != invitation.Email {
		return nil, errors.New("email does not match invitation")
	}

	// Check if already a member
	var count int64
	if err := s.db.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", invitation.OrganizationID, userID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAlreadyMember
	}

	var member models.OrganizationMember
	err = s.db.Transaction(func(tx *gorm.DB) error {
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

	return &member, nil
}

// GetPendingInvitations returns all pending invitations for an organization
func (s *OrgService) GetPendingInvitations(orgID uint) ([]models.OrganizationInvitation, error) {
	var invitations []models.OrganizationInvitation
	if err := s.db.Preload("InvitedByUser").
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, time.Now()).
		Find(&invitations).Error; err != nil {
		return nil, err
	}
	return invitations, nil
}

// CancelInvitation cancels a pending invitation
func (s *OrgService) CancelInvitation(invitationID, orgID uint) error {
	result := s.db.Where("id = ? AND organization_id = ? AND accepted_at IS NULL", invitationID, orgID).
		Delete(&models.OrganizationInvitation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *OrgService) CleanupExpiredInvitations() error {
	return s.db.Where("expires_at < ? AND accepted_at IS NULL", time.Now()).
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
