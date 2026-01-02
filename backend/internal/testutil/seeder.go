package testutil

import (
	"fmt"
	"testing"
	"time"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// TestSeeder provides utilities for seeding test data.
type TestSeeder struct {
	db *gorm.DB
	t  *testing.T
}

// NewTestSeeder creates a new test seeder.
func NewTestSeeder(t *testing.T, db *gorm.DB) *TestSeeder {
	return &TestSeeder{db: db, t: t}
}

// SeedUser creates a user with the given options.
func (s *TestSeeder) SeedUser(opts ...UserOption) *models.User {
	s.t.Helper()

	factory := NewUserFactory()
	for _, opt := range opts {
		opt(factory)
	}

	user := factory.Build()

	if err := s.db.Create(user).Error; err != nil {
		s.t.Fatalf("Failed to seed user: %v", err)
	}

	return user
}

// SeedOrganization creates an organization with an owner member.
func (s *TestSeeder) SeedOrganization(name string, owner *models.User) *models.Organization {
	s.t.Helper()

	org := &models.Organization{
		Name:            name,
		Slug:            fmt.Sprintf("test-org-%d", time.Now().UnixNano()),
		Plan:            models.OrgPlanFree,
		CreatedByUserID: owner.ID,
	}

	if err := s.db.Create(org).Error; err != nil {
		s.t.Fatalf("Failed to seed organization: %v", err)
	}

	// Add owner as member
	s.SeedOrganizationMember(org, owner, models.OrgRoleOwner)

	return org
}

// SeedOrganizationMember adds a user to an organization.
func (s *TestSeeder) SeedOrganizationMember(org *models.Organization, user *models.User, role models.OrganizationRole) *models.OrganizationMember {
	s.t.Helper()

	now := time.Now()
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           role,
		Status:         models.MemberStatusActive,
		AcceptedAt:     &now,
	}

	if err := s.db.Create(member).Error; err != nil {
		s.t.Fatalf("Failed to seed organization member: %v", err)
	}

	return member
}

// SeedOrganizationInvitation creates an invitation to join an organization.
func (s *TestSeeder) SeedOrganizationInvitation(org *models.Organization, email string, inviter *models.User, role models.OrganizationRole) *models.OrganizationInvitation {
	s.t.Helper()

	invitation := &models.OrganizationInvitation{
		OrganizationID:  org.ID,
		Email:           email,
		Role:            role,
		Token:           fmt.Sprintf("test-invite-token-%d", time.Now().UnixNano()),
		InvitedByUserID: inviter.ID,
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(invitation).Error; err != nil {
		s.t.Fatalf("Failed to seed organization invitation: %v", err)
	}

	return invitation
}

// SeedSubscription creates a subscription for a user.
func (s *TestSeeder) SeedSubscription(user *models.User, status string) *models.Subscription {
	s.t.Helper()

	now := time.Now()
	subscription := &models.Subscription{
		UserID:               user.ID,
		StripeSubscriptionID: fmt.Sprintf("sub_test_%d", time.Now().UnixNano()),
		StripePriceID:        "price_test_premium",
		Status:               status,
		CurrentPeriodStart:   now.Format(time.RFC3339),
		CurrentPeriodEnd:     now.AddDate(0, 1, 0).Format(time.RFC3339),
		CreatedAt:            now.Format(time.RFC3339),
		UpdatedAt:            now.Format(time.RFC3339),
	}

	if err := s.db.Create(subscription).Error; err != nil {
		s.t.Fatalf("Failed to seed subscription: %v", err)
	}

	return subscription
}

// SeedFeatureFlag creates a feature flag.
func (s *TestSeeder) SeedFeatureFlag(key, name string, enabled bool) *models.FeatureFlag {
	s.t.Helper()

	now := time.Now().Format(time.RFC3339)
	flag := &models.FeatureFlag{
		Key:         key,
		Name:        name,
		Description: fmt.Sprintf("Test feature flag: %s", name),
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.db.Create(flag).Error; err != nil {
		s.t.Fatalf("Failed to seed feature flag: %v", err)
	}

	return flag
}

// SeedUserFeatureFlag creates a user-specific feature flag override.
func (s *TestSeeder) SeedUserFeatureFlag(user *models.User, flag *models.FeatureFlag, enabled bool) *models.UserFeatureFlag {
	s.t.Helper()

	now := time.Now().Format(time.RFC3339)
	userFlag := &models.UserFeatureFlag{
		UserID:        user.ID,
		FeatureFlagID: flag.ID,
		Enabled:       enabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.db.Create(userFlag).Error; err != nil {
		s.t.Fatalf("Failed to seed user feature flag: %v", err)
	}

	return userFlag
}

// SeedFile creates a file for a user.
func (s *TestSeeder) SeedFile(user *models.User, opts ...FileOption) *models.File {
	s.t.Helper()

	factory := NewFileFactory().WithUserID(user.ID)
	for _, opt := range opts {
		opt(factory)
	}

	file := factory.Build()

	if err := s.db.Create(file).Error; err != nil {
		s.t.Fatalf("Failed to seed file: %v", err)
	}

	return file
}

// SeedAuditLog creates an audit log entry.
func (s *TestSeeder) SeedAuditLog(user *models.User, action, targetType string, targetID uint) *models.AuditLog {
	s.t.Helper()

	factory := NewAuditLogFactory().
		WithUserID(user.ID).
		WithAction(action).
		WithTargetType(targetType).
		WithTargetID(targetID)

	auditLog := factory.Build()

	if err := s.db.Create(auditLog).Error; err != nil {
		s.t.Fatalf("Failed to seed audit log: %v", err)
	}

	return auditLog
}

// SeedAPIKey creates an API key for a user.
func (s *TestSeeder) SeedAPIKey(user *models.User, provider, name string) *models.UserAPIKey {
	s.t.Helper()

	now := time.Now().Format(time.RFC3339)
	apiKey := &models.UserAPIKey{
		UserID:       user.ID,
		Provider:     provider,
		Name:         name,
		KeyHash:      fmt.Sprintf("hash_%d", time.Now().UnixNano()),
		KeyEncrypted: "encrypted_key_data",
		KeyPreview:   "****abcd",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.db.Create(apiKey).Error; err != nil {
		s.t.Fatalf("Failed to seed API key: %v", err)
	}

	return apiKey
}

// UserOption is a function that modifies a UserFactory.
type UserOption func(*UserFactory)

// WithUserEmail sets the user's email.
func WithUserEmail(email string) UserOption {
	return func(f *UserFactory) {
		f.WithEmail(email)
	}
}

// WithUserName sets the user's name.
func WithUserName(name string) UserOption {
	return func(f *UserFactory) {
		f.WithName(name)
	}
}

// WithUserRole sets the user's role.
func WithUserRole(role string) UserOption {
	return func(f *UserFactory) {
		f.WithRole(role)
	}
}

// WithUserAdmin makes the user an admin.
func WithUserAdmin() UserOption {
	return func(f *UserFactory) {
		f.AsAdmin()
	}
}

// WithUserSuperAdmin makes the user a super admin.
func WithUserSuperAdmin() UserOption {
	return func(f *UserFactory) {
		f.AsSuperAdmin()
	}
}

// WithUserInactive makes the user inactive.
func WithUserInactive() UserOption {
	return func(f *UserFactory) {
		f.Inactive()
	}
}

// WithUserUnverified makes the user email unverified.
func WithUserUnverified() UserOption {
	return func(f *UserFactory) {
		f.UnverifiedEmail()
	}
}

// FileOption is a function that modifies a FileFactory.
type FileOption func(*FileFactory)

// WithFileName sets the file name.
func WithFileName(name string) FileOption {
	return func(f *FileFactory) {
		f.WithFileName(name)
	}
}

// WithFileAsImage sets the file as an image.
func WithFileAsImage() FileOption {
	return func(f *FileFactory) {
		f.AsImage()
	}
}

// WithFileAsPDF sets the file as a PDF.
func WithFileAsPDF() FileOption {
	return func(f *FileFactory) {
		f.AsPDF()
	}
}

// TestSuiteData holds commonly needed test data.
type TestSuiteData struct {
	Admin        *models.User
	User         *models.User
	Organization *models.Organization
	FeatureFlags map[string]*models.FeatureFlag
}

// SeedTestSuite creates a complete test data set.
func (s *TestSeeder) SeedTestSuite() *TestSuiteData {
	s.t.Helper()

	// Create admin user
	admin := s.SeedUser(WithUserSuperAdmin(), WithUserEmail("admin@test.local"))

	// Create regular user
	user := s.SeedUser(WithUserEmail("user@test.local"))

	// Create organization with admin as owner
	org := s.SeedOrganization("Test Organization", admin)

	// Add regular user as member
	s.SeedOrganizationMember(org, user, models.OrgRoleMember)

	// Create feature flags
	flags := make(map[string]*models.FeatureFlag)
	flags["enabled"] = s.SeedFeatureFlag("test_feature_enabled", "Enabled Feature", true)
	flags["disabled"] = s.SeedFeatureFlag("test_feature_disabled", "Disabled Feature", false)

	return &TestSuiteData{
		Admin:        admin,
		User:         user,
		Organization: org,
		FeatureFlags: flags,
	}
}

// SeedMultipleUsers creates multiple test users.
func (s *TestSeeder) SeedMultipleUsers(count int) []*models.User {
	s.t.Helper()

	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		users[i] = s.SeedUser(WithUserEmail(fmt.Sprintf("testuser%d@test.local", i)))
	}
	return users
}
