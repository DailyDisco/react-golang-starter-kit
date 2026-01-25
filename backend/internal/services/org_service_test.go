package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ============ Org Service Error Tests ============

func TestOrgServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrOrgNotFound", ErrOrgNotFound, "organization not found"},
		{"ErrOrgSlugTaken", ErrOrgSlugTaken, "organization slug is already taken"},
		{"ErrInvalidSlug", ErrInvalidSlug, "invalid slug format"},
		{"ErrNotMember", ErrNotMember, "user is not a member of this organization"},
		{"ErrInsufficientRole", ErrInsufficientRole, "insufficient role permissions"},
		{"ErrCannotRemoveOwner", ErrCannotRemoveOwner, "cannot remove the organization owner"},
		{"ErrInvitationNotFound", ErrInvitationNotFound, "invitation not found"},
		{"ErrInvitationExpired", ErrInvitationExpired, "invitation has expired"},
		{"ErrInvitationAccepted", ErrInvitationAccepted, "invitation has already been accepted"},
		{"ErrAlreadyMember", ErrAlreadyMember, "user is already a member"},
		{"ErrCannotChangeOwnRole", ErrCannotChangeOwnRole, "cannot change your own role"},
		{"ErrMustHaveOwner", ErrMustHaveOwner, "organization must have at least one owner"},
		{"ErrInvitationEmailTaken", ErrInvitationEmailTaken, "an invitation for this email already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

// ============ Slug Validation Tests ============

func TestSlugValidation(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		isValid bool
	}{
		// Valid slugs
		{"lowercase letters", "myorg", true},
		{"with numbers", "myorg123", true},
		{"with hyphens", "my-org", true},
		{"multiple hyphens", "my-cool-org", true},
		{"starts with number", "123org", true},
		{"single letter", "a", true},
		{"single number", "1", true},

		// Invalid slugs
		{"uppercase letters", "MyOrg", false},
		{"with spaces", "my org", false},
		{"with underscores", "my_org", false},
		{"starts with hyphen", "-myorg", false},
		{"ends with hyphen", "myorg-", false},
		{"double hyphen", "my--org", false},
		{"special characters", "my@org", false},
		{"empty string", "", false},
		{"only hyphen", "-", false},
		{"unicode characters", "orgäöü", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := slugRegex.MatchString(tt.slug)
			if matched != tt.isValid {
				t.Errorf("slugRegex.MatchString(%q) = %v, want %v", tt.slug, matched, tt.isValid)
			}
		})
	}
}

// ============ Generate Invitation Token Tests ============

func TestGenerateInvitationToken(t *testing.T) {
	token1, err := generateInvitationToken()
	if err != nil {
		t.Fatalf("generateInvitationToken() error = %v", err)
	}

	// Token should be 64 characters (32 bytes hex encoded)
	if len(token1) != 64 {
		t.Errorf("generateInvitationToken() length = %d, want 64", len(token1))
	}

	// Tokens should be unique
	token2, err := generateInvitationToken()
	if err != nil {
		t.Fatalf("generateInvitationToken() error = %v", err)
	}

	if token1 == token2 {
		t.Error("generateInvitationToken() should generate unique tokens")
	}
}

func TestGenerateInvitationToken_ValidHex(t *testing.T) {
	token, err := generateInvitationToken()
	if err != nil {
		t.Fatalf("generateInvitationToken() error = %v", err)
	}

	// Verify token is valid hex
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("generateInvitationToken() contains non-hex character: %c", c)
		}
	}
}

// ============ Organization Role Tests ============

func TestOrganizationRoleConstants(t *testing.T) {
	// Verify role constants are defined correctly
	tests := []struct {
		name string
		role models.OrganizationRole
		want string
	}{
		{"owner role", models.OrgRoleOwner, "owner"},
		{"admin role", models.OrgRoleAdmin, "admin"},
		{"member role", models.OrgRoleMember, "member"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.role, tt.want)
			}
		})
	}
}

func TestMemberStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status models.MemberStatus
		want   string
	}{
		{"active status", models.MemberStatusActive, "active"},
		{"pending status", models.MemberStatusPending, "pending"},
		{"inactive status", models.MemberStatusInactive, "inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.status, tt.want)
			}
		})
	}
}

func TestOrgPlanConstants(t *testing.T) {
	tests := []struct {
		name string
		plan models.OrganizationPlan
		want string
	}{
		{"free plan", models.OrgPlanFree, "free"},
		{"pro plan", models.OrgPlanPro, "pro"},
		{"enterprise plan", models.OrgPlanEnterprise, "enterprise"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.plan) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.plan, tt.want)
			}
		})
	}
}

// ============ OrgService Constructor Tests ============

func TestNewOrgService(t *testing.T) {
	// Test with nil db (shouldn't panic)
	service := NewOrgService(nil)
	if service == nil {
		t.Error("NewOrgService(nil) should return non-nil service")
	}
}

func TestOrgService_SetHub(t *testing.T) {
	service := &OrgService{}

	// Setting nil hub should not panic
	service.SetHub(nil)
	if service.hub != nil {
		t.Error("SetHub(nil) should set hub to nil")
	}
}

// ============ OrgWithRole Structure Tests ============

func TestOrgWithRole_Structure(t *testing.T) {
	orgWithRole := OrgWithRole{
		Organization: models.Organization{
			Name: "Test Org",
			Slug: "test-org",
			Plan: models.OrgPlanPro,
		},
		Role: models.OrgRoleAdmin,
	}

	if orgWithRole.Organization.Name != "Test Org" {
		t.Errorf("Organization.Name = %q, want %q", orgWithRole.Organization.Name, "Test Org")
	}

	if orgWithRole.Role != models.OrgRoleAdmin {
		t.Errorf("Role = %q, want %q", orgWithRole.Role, models.OrgRoleAdmin)
	}
}

// ============ Extended Slug Validation Tests ============

func TestSlugValidation_Length(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		isValid bool
	}{
		{"min length 1 char", "a", true},
		{"2 chars", "ab", true},
		{"typical slug", "my-organization", true},
		{"long valid slug", "this-is-a-very-long-slug-name-for-testing", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := slugRegex.MatchString(tt.slug)
			if matched != tt.isValid {
				t.Errorf("slugRegex.MatchString(%q) = %v, want %v", tt.slug, matched, tt.isValid)
			}
		})
	}
}

func TestSlugValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		isValid bool
	}{
		{"all numbers", "123456", true},
		{"numbers and hyphens", "123-456", true},
		{"mixed alphanumeric", "org123test456", true},
		{"hyphen in middle", "abc-def-ghi", true},
		{"single hyphen between chars", "a-b", true},
		{"single hyphen between long strings", "organization-name", true},

		// Invalid cases
		{"leading hyphen", "-abc", false},
		{"trailing hyphen", "abc-", false},
		{"double hyphen", "abc--def", false},
		{"triple hyphen", "abc---def", false},
		{"hyphen only", "-", false},
		{"multiple hyphens only", "---", false},
		{"space in middle", "abc def", false},
		{"tab character", "abc\tdef", false},
		{"newline", "abc\ndef", false},
		{"period", "abc.def", false},
		{"comma", "abc,def", false},
		{"colon", "abc:def", false},
		{"semicolon", "abc;def", false},
		{"exclamation", "abc!def", false},
		{"question mark", "abc?def", false},
		{"hash", "abc#def", false},
		{"dollar sign", "abc$def", false},
		{"percent", "abc%def", false},
		{"ampersand", "abc&def", false},
		{"asterisk", "abc*def", false},
		{"plus sign", "abc+def", false},
		{"equals sign", "abc=def", false},
		{"brackets", "abc[def]", false},
		{"braces", "abc{def}", false},
		{"parentheses", "abc(def)", false},
		{"forward slash", "abc/def", false},
		{"backslash", "abc\\def", false},
		{"pipe", "abc|def", false},
		{"tilde", "abc~def", false},
		{"backtick", "abc`def", false},
		{"single quote", "abc'def", false},
		{"double quote", "abc\"def", false},
		{"less than", "abc<def", false},
		{"greater than", "abc>def", false},
		{"caret", "abc^def", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := slugRegex.MatchString(tt.slug)
			if matched != tt.isValid {
				t.Errorf("slugRegex.MatchString(%q) = %v, want %v", tt.slug, matched, tt.isValid)
			}
		})
	}
}

// ============ Invitation Token Multiple Generation Tests ============

func TestGenerateInvitationToken_Multiple(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := generateInvitationToken()
		if err != nil {
			t.Fatalf("generateInvitationToken() error = %v on iteration %d", err, i)
		}

		if tokens[token] {
			t.Errorf("generateInvitationToken() generated duplicate token on iteration %d", i)
		}
		tokens[token] = true
	}
}

func TestGenerateInvitationToken_Format(t *testing.T) {
	for i := 0; i < 10; i++ {
		token, err := generateInvitationToken()
		if err != nil {
			t.Fatalf("generateInvitationToken() error = %v", err)
		}

		// Should be exactly 64 characters
		if len(token) != 64 {
			t.Errorf("Token length = %d, want 64", len(token))
		}

		// Should only contain lowercase hex characters
		for j, c := range token {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("Token contains invalid character %q at position %d", c, j)
			}
		}
	}
}

// ============ Error Equality Tests ============

func TestOrgServiceErrors_Identity(t *testing.T) {
	// Verify errors are distinct by pointer
	errors := []error{
		ErrOrgNotFound,
		ErrOrgSlugTaken,
		ErrInvalidSlug,
		ErrNotMember,
		ErrInsufficientRole,
		ErrCannotRemoveOwner,
		ErrInvitationNotFound,
		ErrInvitationExpired,
		ErrInvitationAccepted,
		ErrAlreadyMember,
		ErrCannotChangeOwnRole,
		ErrMustHaveOwner,
		ErrInvitationEmailTaken,
		ErrSeatLimitExceeded,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j && err1 == err2 {
				t.Errorf("Errors at index %d and %d are the same pointer", i, j)
			}
		}
	}
}

// ============ Organization Model Tests ============

func TestOrganization_Fields(t *testing.T) {
	org := models.Organization{
		Name:            "Test Organization",
		Slug:            "test-organization",
		Plan:            models.OrgPlanPro,
		CreatedByUserID: 123,
	}

	if org.Name != "Test Organization" {
		t.Errorf("Name = %q, want %q", org.Name, "Test Organization")
	}
	if org.Slug != "test-organization" {
		t.Errorf("Slug = %q, want %q", org.Slug, "test-organization")
	}
	if org.Plan != models.OrgPlanPro {
		t.Errorf("Plan = %q, want %q", org.Plan, models.OrgPlanPro)
	}
	if org.CreatedByUserID != 123 {
		t.Errorf("CreatedByUserID = %d, want %d", org.CreatedByUserID, 123)
	}
}

func TestOrganizationMember_Fields(t *testing.T) {
	member := models.OrganizationMember{
		OrganizationID: 1,
		UserID:         2,
		Role:           models.OrgRoleAdmin,
		Status:         models.MemberStatusActive,
	}

	if member.OrganizationID != 1 {
		t.Errorf("OrganizationID = %d, want %d", member.OrganizationID, 1)
	}
	if member.UserID != 2 {
		t.Errorf("UserID = %d, want %d", member.UserID, 2)
	}
	if member.Role != models.OrgRoleAdmin {
		t.Errorf("Role = %q, want %q", member.Role, models.OrgRoleAdmin)
	}
	if member.Status != models.MemberStatusActive {
		t.Errorf("Status = %q, want %q", member.Status, models.MemberStatusActive)
	}
}

func TestOrganizationInvitation_Fields(t *testing.T) {
	invitation := models.OrganizationInvitation{
		OrganizationID:  1,
		Email:           "test@example.com",
		Role:            models.OrgRoleMember,
		Token:           "abc123",
		InvitedByUserID: 2,
	}

	if invitation.OrganizationID != 1 {
		t.Errorf("OrganizationID = %d, want %d", invitation.OrganizationID, 1)
	}
	if invitation.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", invitation.Email, "test@example.com")
	}
	if invitation.Role != models.OrgRoleMember {
		t.Errorf("Role = %q, want %q", invitation.Role, models.OrgRoleMember)
	}
	if invitation.Token != "abc123" {
		t.Errorf("Token = %q, want %q", invitation.Token, "abc123")
	}
}

// ============ Mock-Based Tests ============

// newTestOrgService creates an OrgService with mock repositories for testing.
func newTestOrgService() (*OrgService, *mocks.MockOrganizationRepository, *mocks.MockOrganizationMemberRepository, *mocks.MockOrganizationInvitationRepository, *mocks.MockSubscriptionRepository, *mocks.MockUserRepository) {
	orgRepo := mocks.NewMockOrganizationRepository()
	memberRepo := mocks.NewMockOrganizationMemberRepository()
	invitationRepo := mocks.NewMockOrganizationInvitationRepository()
	subRepo := mocks.NewMockSubscriptionRepository()
	userRepo := mocks.NewMockUserRepository()

	svc := NewOrgServiceWithRepo(nil, orgRepo, memberRepo, invitationRepo, subRepo, userRepo)
	return svc, orgRepo, memberRepo, invitationRepo, subRepo, userRepo
}

func TestOrgService_GetOrganization(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		setupOrg *models.Organization
		setupErr error
		wantErr  error
		wantSlug string
	}{
		{
			name: "success",
			slug: "test-org",
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
				Plan: models.OrgPlanFree,
			},
			wantSlug: "test-org",
		},
		{
			name:     "not found",
			slug:     "nonexistent",
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrOrgNotFound,
		},
		{
			name:     "database error",
			slug:     "test-org",
			setupErr: errors.New("database error"),
			wantErr:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.FindBySlugErr = tt.setupErr
			}

			org, err := svc.GetOrganization(context.Background(), tt.slug)

			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.wantErr == ErrOrgNotFound {
					assert.ErrorIs(t, err, ErrOrgNotFound)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSlug, org.Slug)
		})
	}
}

func TestOrgService_GetOrganizationByID(t *testing.T) {
	tests := []struct {
		name     string
		orgID    uint
		setupOrg *models.Organization
		setupErr error
		wantErr  error
		wantName string
	}{
		{
			name:  "success",
			orgID: 1,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
			},
			wantName: "Test Org",
		},
		{
			name:     "not found",
			orgID:    999,
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrOrgNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.FindByIDErr = tt.setupErr
			}

			org, err := svc.GetOrganizationByID(context.Background(), tt.orgID)

			if tt.wantErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, org.Name)
		})
	}
}

func TestOrgService_GetUserOrganizations(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupOrgs []models.Organization
		setupErr  error
		wantCount int
	}{
		{
			name:   "returns user organizations",
			userID: 1,
			setupOrgs: []models.Organization{
				{ID: 1, Name: "Org 1", Slug: "org-1"},
				{ID: 2, Name: "Org 2", Slug: "org-2"},
			},
			wantCount: 2,
		},
		{
			name:      "no organizations",
			userID:    2,
			setupOrgs: []models.Organization{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, _, _, _ := newTestOrgService()

			memberRepo.SetUserOrgs(tt.userID, tt.setupOrgs)
			if tt.setupErr != nil {
				memberRepo.FindOrgsByUserIDErr = tt.setupErr
			}

			orgs, err := svc.GetUserOrganizations(context.Background(), tt.userID)

			if tt.setupErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, orgs, tt.wantCount)
		})
	}
}

func TestOrgService_GetUserMembership(t *testing.T) {
	tests := []struct {
		name        string
		orgID       uint
		userID      uint
		setupMember *models.OrganizationMember
		setupErr    error
		wantErr     error
		wantRole    models.OrganizationRole
	}{
		{
			name:   "success",
			orgID:  1,
			userID: 1,
			setupMember: &models.OrganizationMember{
				OrganizationID: 1,
				UserID:         1,
				Role:           models.OrgRoleOwner,
				Status:         models.MemberStatusActive,
			},
			wantRole: models.OrgRoleOwner,
		},
		{
			name:     "not member",
			orgID:    1,
			userID:   999,
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrNotMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, _, _, _ := newTestOrgService()

			if tt.setupMember != nil {
				memberRepo.AddMember(*tt.setupMember)
			}
			if tt.setupErr != nil {
				memberRepo.FindByOrgIDAndUserIDErr = tt.setupErr
			}

			member, err := svc.GetUserMembership(context.Background(), tt.orgID, tt.userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRole, member.Role)
		})
	}
}

func TestOrgService_GetMembers(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		setupMembers []models.OrganizationMember
		setupErr     error
		wantCount    int
	}{
		{
			name:  "returns all members",
			orgID: 1,
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner},
				{OrganizationID: 1, UserID: 2, Role: models.OrgRoleAdmin},
				{OrganizationID: 1, UserID: 3, Role: models.OrgRoleMember},
			},
			wantCount: 3,
		},
		{
			name:         "no members",
			orgID:        2,
			setupMembers: []models.OrganizationMember{},
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, _, _, _ := newTestOrgService()

			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}
			if tt.setupErr != nil {
				memberRepo.FindByOrgIDErr = tt.setupErr
			}

			members, err := svc.GetMembers(context.Background(), tt.orgID)

			if tt.setupErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, members, tt.wantCount)
		})
	}
}

func TestOrgService_UpdateMemberRole(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		userID       uint
		actorUserID  uint
		newRole      models.OrganizationRole
		setupMembers []models.OrganizationMember
		setupErr     error
		wantErr      error
	}{
		{
			name:        "cannot change own role",
			orgID:       1,
			userID:      1,
			actorUserID: 1,
			newRole:     models.OrgRoleMember,
			wantErr:     ErrCannotChangeOwnRole,
		},
		{
			name:        "member not found",
			orgID:       1,
			userID:      999,
			actorUserID: 1,
			newRole:     models.OrgRoleMember,
			setupErr:    gorm.ErrRecordNotFound,
			wantErr:     ErrNotMember,
		},
		{
			name:        "demote last owner fails",
			orgID:       1,
			userID:      2,
			actorUserID: 1,
			newRole:     models.OrgRoleMember,
			setupMembers: []models.OrganizationMember{
				{ID: 1, OrganizationID: 1, UserID: 1, Role: models.OrgRoleAdmin},
				{ID: 2, OrganizationID: 1, UserID: 2, Role: models.OrgRoleOwner},
			},
			wantErr: ErrMustHaveOwner,
		},
		{
			name:        "success update role",
			orgID:       1,
			userID:      2,
			actorUserID: 1,
			newRole:     models.OrgRoleMember,
			setupMembers: []models.OrganizationMember{
				{ID: 1, OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner},
				{ID: 2, OrganizationID: 1, UserID: 2, Role: models.OrgRoleAdmin},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, memberRepo, _, _, _ := newTestOrgService()

			// Setup organization for broadcast
			orgRepo.AddOrganization(&models.Organization{ID: 1, Name: "Test", Slug: "test"})

			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}
			if tt.setupErr != nil {
				memberRepo.FindByOrgIDAndUserIDErr = tt.setupErr
			}

			err := svc.UpdateMemberRole(context.Background(), tt.orgID, tt.userID, tt.actorUserID, tt.newRole)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestOrgService_RemoveMember(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		userID       uint
		setupMembers []models.OrganizationMember
		setupErr     error
		wantErr      error
	}{
		{
			name:     "member not found",
			orgID:    1,
			userID:   999,
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrNotMember,
		},
		{
			name:   "cannot remove last owner",
			orgID:  1,
			userID: 1,
			setupMembers: []models.OrganizationMember{
				{ID: 1, OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner},
			},
			wantErr: ErrCannotRemoveOwner,
		},
		{
			name:   "success remove member",
			orgID:  1,
			userID: 2,
			setupMembers: []models.OrganizationMember{
				{ID: 1, OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner},
				{ID: 2, OrganizationID: 1, UserID: 2, Role: models.OrgRoleMember},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, memberRepo, _, _, _ := newTestOrgService()

			// Setup organization for broadcast
			orgRepo.AddOrganization(&models.Organization{ID: 1, Name: "Test", Slug: "test"})

			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}
			if tt.setupErr != nil {
				memberRepo.FindByOrgIDAndUserIDErr = tt.setupErr
			}

			err := svc.RemoveMember(context.Background(), tt.orgID, tt.userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestOrgService_GetInvitationByToken(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		setupInvitation *models.OrganizationInvitation
		setupErr        error
		wantErr         error
		wantEmail       string
	}{
		{
			name:  "success",
			token: "valid-token",
			setupInvitation: &models.OrganizationInvitation{
				OrganizationID: 1,
				Email:          "test@example.com",
				Token:          "valid-token",
				Role:           models.OrgRoleMember,
				ExpiresAt:      time.Now().Add(24 * time.Hour),
			},
			wantEmail: "test@example.com",
		},
		{
			name:     "not found",
			token:    "invalid-token",
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrInvitationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, invRepo, _, _ := newTestOrgService()

			if tt.setupInvitation != nil {
				invRepo.AddInvitation(*tt.setupInvitation)
			}
			if tt.setupErr != nil {
				invRepo.FindByTokenErr = tt.setupErr
			}

			inv, err := svc.GetInvitationByToken(context.Background(), tt.token)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEmail, inv.Email)
		})
	}
}

func TestOrgService_GetPendingInvitations(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		orgID            uint
		setupInvitations []models.OrganizationInvitation
		wantCount        int
	}{
		{
			name:  "returns pending invitations",
			orgID: 1,
			setupInvitations: []models.OrganizationInvitation{
				{OrganizationID: 1, Email: "a@test.com", Token: "t1", ExpiresAt: now.Add(24 * time.Hour)},
				{OrganizationID: 1, Email: "b@test.com", Token: "t2", ExpiresAt: now.Add(24 * time.Hour)},
			},
			wantCount: 2,
		},
		{
			name:             "no invitations",
			orgID:            2,
			setupInvitations: []models.OrganizationInvitation{},
			wantCount:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, invRepo, _, _ := newTestOrgService()

			for _, inv := range tt.setupInvitations {
				invRepo.AddInvitation(inv)
			}

			invitations, err := svc.GetPendingInvitations(context.Background(), tt.orgID)

			require.NoError(t, err)
			assert.Len(t, invitations, tt.wantCount)
		})
	}
}

func TestOrgService_CancelInvitation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		invitationID     uint
		orgID            uint
		setupInvitations []models.OrganizationInvitation
		wantErr          error
	}{
		{
			name:         "success",
			invitationID: 1,
			orgID:        1,
			setupInvitations: []models.OrganizationInvitation{
				{ID: 1, OrganizationID: 1, Email: "test@test.com", Token: "t1", ExpiresAt: now.Add(24 * time.Hour)},
			},
		},
		{
			name:             "not found",
			invitationID:     999,
			orgID:            1,
			setupInvitations: []models.OrganizationInvitation{},
			wantErr:          ErrInvitationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, invRepo, _, _ := newTestOrgService()

			// Setup organization for broadcast
			orgRepo.AddOrganization(&models.Organization{ID: 1, Name: "Test", Slug: "test"})

			for _, inv := range tt.setupInvitations {
				invRepo.AddInvitation(inv)
			}

			err := svc.CancelInvitation(context.Background(), tt.invitationID, tt.orgID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestOrgService_GetMemberCount(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		setupMembers []models.OrganizationMember
		wantCount    int64
	}{
		{
			name:  "counts active members",
			orgID: 1,
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 3, Status: models.MemberStatusInactive},
			},
			wantCount: 2,
		},
		{
			name:         "zero members",
			orgID:        2,
			setupMembers: []models.OrganizationMember{},
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, _, _, _ := newTestOrgService()

			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}

			count, err := svc.GetMemberCount(context.Background(), tt.orgID)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestOrgService_GetOrganizationByStripeCustomerID(t *testing.T) {
	customerID := "cus_123"

	tests := []struct {
		name       string
		customerID string
		setupOrg   *models.Organization
		setupErr   error
		wantErr    error
	}{
		{
			name:       "success",
			customerID: customerID,
			setupOrg: &models.Organization{
				ID:               1,
				Name:             "Test",
				Slug:             "test",
				StripeCustomerID: &customerID,
			},
		},
		{
			name:       "not found",
			customerID: "cus_nonexistent",
			setupErr:   gorm.ErrRecordNotFound,
			wantErr:    ErrOrgNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.FindByStripeCustomerIDErr = tt.setupErr
			}

			org, err := svc.GetOrganizationByStripeCustomerID(context.Background(), tt.customerID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.customerID, *org.StripeCustomerID)
		})
	}
}

func TestOrgService_GetOrganizationSubscription(t *testing.T) {
	orgID := uint(1)

	tests := []struct {
		name     string
		orgID    uint
		setupSub *models.Subscription
		setupErr error
		wantNil  bool
	}{
		{
			name:  "success",
			orgID: orgID,
			setupSub: &models.Subscription{
				OrganizationID: &orgID,
				Status:         "active",
			},
		},
		{
			name:     "no subscription",
			orgID:    2,
			setupErr: gorm.ErrRecordNotFound,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, subRepo, _ := newTestOrgService()

			if tt.setupSub != nil {
				subRepo.AddSubscription(*tt.setupSub)
			}
			if tt.setupErr != nil {
				subRepo.FindByOrgIDErr = tt.setupErr
			}

			sub, err := svc.GetOrganizationSubscription(context.Background(), tt.orgID)

			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, sub)
			} else {
				assert.NotNil(t, sub)
			}
		})
	}
}

func TestOrgService_CleanupExpiredInvitations(t *testing.T) {
	svc, _, _, invRepo, _, _ := newTestOrgService()

	now := time.Now()

	// Add expired and valid invitations
	invRepo.AddInvitation(models.OrganizationInvitation{
		OrganizationID: 1,
		Email:          "expired@test.com",
		Token:          "expired-token",
		ExpiresAt:      now.Add(-24 * time.Hour), // Expired
	})
	invRepo.AddInvitation(models.OrganizationInvitation{
		OrganizationID: 1,
		Email:          "valid@test.com",
		Token:          "valid-token",
		ExpiresAt:      now.Add(24 * time.Hour), // Valid
	})

	err := svc.CleanupExpiredInvitations(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, invRepo.DeleteExpiredCalls)
}

func TestNewOrgServiceWithRepo(t *testing.T) {
	orgRepo := mocks.NewMockOrganizationRepository()
	memberRepo := mocks.NewMockOrganizationMemberRepository()
	invRepo := mocks.NewMockOrganizationInvitationRepository()
	subRepo := mocks.NewMockSubscriptionRepository()
	userRepo := mocks.NewMockUserRepository()

	svc := NewOrgServiceWithRepo(nil, orgRepo, memberRepo, invRepo, subRepo, userRepo)

	assert.NotNil(t, svc)
	assert.Equal(t, orgRepo, svc.orgRepo)
	assert.Equal(t, memberRepo, svc.memberRepo)
	assert.Equal(t, invRepo, svc.invitationRepo)
	assert.Equal(t, subRepo, svc.subRepo)
	assert.Equal(t, userRepo, svc.userRepo)
}

// ============ CreateOrganization Validation Tests ============
// Note: CreateOrganization uses s.db.Transaction() which requires a real DB.
// We can only test validation that fails BEFORE the transaction is started.

func TestOrgService_CreateOrganization_ValidationBeforeDBCall(t *testing.T) {
	// These tests validate slugs that fail regex validation (before DB calls)
	// Note: uppercase is converted to lowercase before validation, so it's valid
	tests := []struct {
		name    string
		userID  uint
		orgName string
		slug    string
		wantErr error
	}{
		{
			name:    "invalid slug - spaces",
			userID:  1,
			orgName: "Test Org",
			slug:    "test org",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - special chars",
			userID:  1,
			orgName: "Test Org",
			slug:    "test@org",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - leading hyphen",
			userID:  1,
			orgName: "Test Org",
			slug:    "-test-org",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - trailing hyphen",
			userID:  1,
			orgName: "Test Org",
			slug:    "test-org-",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - double hyphen",
			userID:  1,
			orgName: "Test Org",
			slug:    "test--org",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - too short",
			userID:  1,
			orgName: "Test Org",
			slug:    "a",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - too long (64 chars)",
			userID:  1,
			orgName: "Test Org",
			slug:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "invalid slug - empty",
			userID:  1,
			orgName: "Test Org",
			slug:    "",
			wantErr: ErrInvalidSlug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, _, _ := newTestOrgService()

			_, err := svc.CreateOrganization(context.Background(), tt.userID, tt.orgName, tt.slug)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrgService_CreateOrganization_SlugTaken(t *testing.T) {
	svc, orgRepo, _, _, _, _ := newTestOrgService()

	// Add existing organization - this causes CountBySlug to return 1
	orgRepo.AddOrganization(&models.Organization{
		ID:   1,
		Name: "Existing Org",
		Slug: "existing-org",
	})

	_, err := svc.CreateOrganization(context.Background(), 1, "New Org", "existing-org")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOrgSlugTaken)
	// Verify CountBySlug was called
	assert.Equal(t, 1, orgRepo.CountBySlugCalls)
}

func TestOrgService_CreateOrganization_CountSlugError(t *testing.T) {
	svc, orgRepo, _, _, _, _ := newTestOrgService()

	orgRepo.CountBySlugErr = errors.New("database error")

	_, err := svc.CreateOrganization(context.Background(), 1, "Test Org", "test-org")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	// Verify CountBySlug was called
	assert.Equal(t, 1, orgRepo.CountBySlugCalls)
}

// ============ GetOrganizationWithMembers Tests ============

func TestOrgService_GetOrganizationWithMembers(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		setupOrg *models.Organization
		setupErr error
		wantErr  error
		wantName string
	}{
		{
			name: "success",
			slug: "test-org",
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
			},
			wantName: "Test Org",
		},
		{
			name:     "not found",
			slug:     "nonexistent",
			setupErr: gorm.ErrRecordNotFound,
			wantErr:  ErrOrgNotFound,
		},
		{
			name:     "database error",
			slug:     "test-org",
			setupErr: errors.New("database error"),
			wantErr:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.FindBySlugWithMembersErr = tt.setupErr
			}

			org, err := svc.GetOrganizationWithMembers(context.Background(), tt.slug)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrOrgNotFound) {
					assert.ErrorIs(t, err, ErrOrgNotFound)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, org.Name)
		})
	}
}

// ============ GetUserOrganizationsWithRoles Tests ============

func TestOrgService_GetUserOrganizationsWithRoles(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupOrgs []struct {
			org  models.Organization
			role models.OrganizationRole
		}
		setupErr  error
		wantCount int
	}{
		{
			name:   "returns organizations with roles",
			userID: 1,
			setupOrgs: []struct {
				org  models.Organization
				role models.OrganizationRole
			}{
				{
					org:  models.Organization{ID: 1, Name: "Org 1", Slug: "org-1"},
					role: models.OrgRoleOwner,
				},
				{
					org:  models.Organization{ID: 2, Name: "Org 2", Slug: "org-2"},
					role: models.OrgRoleMember,
				},
			},
			wantCount: 2,
		},
		{
			name:      "no organizations",
			userID:    2,
			setupOrgs: nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, memberRepo, _, _, _ := newTestOrgService()

			if tt.setupOrgs != nil {
				var repoOrgs []struct {
					Org  models.Organization
					Role models.OrganizationRole
				}
				for _, o := range tt.setupOrgs {
					repoOrgs = append(repoOrgs, struct {
						Org  models.Organization
						Role models.OrganizationRole
					}{Org: o.org, Role: o.role})
				}
				memberRepo.SetUserOrgsWithRoles(tt.userID, repoOrgs)
			}
			if tt.setupErr != nil {
				memberRepo.FindOrgsWithRolesByUserIDErr = tt.setupErr
			}

			orgs, err := svc.GetUserOrganizationsWithRoles(context.Background(), tt.userID)

			if tt.setupErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, orgs, tt.wantCount)
		})
	}
}

// ============ CreateInvitation Tests ============

func TestOrgService_CreateInvitation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		orgID        uint
		inviterID    uint
		email        string
		role         models.OrganizationRole
		setupOrg     *models.Organization
		setupMembers []models.OrganizationMember
		setupUser    *models.User
		setupInvites []models.OrganizationInvitation
		wantErr      error
	}{
		{
			name:      "success - new user",
			orgID:     1,
			inviterID: 1,
			email:     "newuser@example.com",
			role:      models.OrgRoleMember,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
				Plan: models.OrgPlanPro, // Pro plan has more seats
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner, Status: models.MemberStatusActive},
			},
		},
		{
			name:      "error - already a member",
			orgID:     1,
			inviterID: 1,
			email:     "existing@example.com",
			role:      models.OrgRoleMember,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
				Plan: models.OrgPlanPro,
			},
			setupUser: &models.User{
				ID:    2,
				Email: "existing@example.com",
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Role: models.OrgRoleMember, Status: models.MemberStatusActive},
			},
			wantErr: ErrAlreadyMember,
		},
		{
			name:      "error - pending invitation exists",
			orgID:     1,
			inviterID: 1,
			email:     "pending@example.com",
			role:      models.OrgRoleMember,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test Org",
				Slug: "test-org",
				Plan: models.OrgPlanPro,
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Role: models.OrgRoleOwner, Status: models.MemberStatusActive},
			},
			setupInvites: []models.OrganizationInvitation{
				{
					OrganizationID: 1,
					Email:          "pending@example.com",
					Token:          "existing-token",
					ExpiresAt:      now.Add(24 * time.Hour),
				},
			},
			wantErr: ErrInvitationEmailTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, memberRepo, invRepo, _, userRepo := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}
			if tt.setupUser != nil {
				userRepo.AddUser(tt.setupUser)
			}
			for _, inv := range tt.setupInvites {
				invRepo.AddInvitation(inv)
			}

			invitation, err := svc.CreateInvitation(context.Background(), tt.orgID, tt.inviterID, tt.email, tt.role)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, invitation)
			assert.Equal(t, tt.email, invitation.Email)
			assert.Equal(t, tt.role, invitation.Role)
			assert.NotEmpty(t, invitation.Token)
		})
	}
}

func TestOrgService_CreateInvitation_SeatLimitExceeded(t *testing.T) {
	svc, orgRepo, memberRepo, _, _, _ := newTestOrgService()

	// Free plan has seat limit of 5
	orgRepo.AddOrganization(&models.Organization{
		ID:   1,
		Name: "Test Org",
		Slug: "test-org",
		Plan: models.OrgPlanFree,
	})

	// Add 5 active members to hit the limit
	for i := uint(1); i <= 5; i++ {
		memberRepo.AddMember(models.OrganizationMember{
			OrganizationID: 1,
			UserID:         i,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
		})
	}

	_, err := svc.CreateInvitation(context.Background(), 1, 1, "newuser@example.com", models.OrgRoleMember)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSeatLimitExceeded)
}

// ============ UpdateOrganizationPlan Tests ============

func TestOrgService_UpdateOrganizationPlan(t *testing.T) {
	stripeSubID := "sub_123"

	tests := []struct {
		name        string
		orgID       uint
		plan        models.OrganizationPlan
		stripeSubID *string
		setupOrg    *models.Organization
		setupErr    error
		wantErr     bool
	}{
		{
			name:        "success",
			orgID:       1,
			plan:        models.OrgPlanPro,
			stripeSubID: &stripeSubID,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
				Plan: models.OrgPlanFree,
			},
		},
		{
			name:        "error - not found",
			orgID:       999,
			plan:        models.OrgPlanPro,
			stripeSubID: nil,
			setupErr:    errors.New("not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.UpdatePlanErr = tt.setupErr
			}

			err := svc.UpdateOrganizationPlan(context.Background(), tt.orgID, tt.plan, tt.stripeSubID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// ============ SetOrganizationStripeCustomer Tests ============

func TestOrgService_SetOrganizationStripeCustomer(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint
		customerID string
		setupOrg   *models.Organization
		setupErr   error
		wantErr    bool
	}{
		{
			name:       "success",
			orgID:      1,
			customerID: "cus_123",
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
			},
		},
		{
			name:       "error - not found",
			orgID:      999,
			customerID: "cus_123",
			setupErr:   errors.New("not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.UpdateStripeCustomerErr = tt.setupErr
			}

			err := svc.SetOrganizationStripeCustomer(context.Background(), tt.orgID, tt.customerID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// ============ CanAddMember Tests ============

func TestOrgService_CanAddMember(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		setupOrg     *models.Organization
		setupMembers []models.OrganizationMember
		setupInvites []models.OrganizationInvitation
		wantCanAdd   bool
		wantErr      bool
	}{
		{
			name:  "can add - under limit",
			orgID: 1,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
				Plan: models.OrgPlanFree, // 5 seat limit
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Status: models.MemberStatusActive},
			},
			wantCanAdd: true,
		},
		{
			name:  "cannot add - at limit",
			orgID: 1,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
				Plan: models.OrgPlanFree, // 5 seat limit
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 3, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 4, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 5, Status: models.MemberStatusActive},
			},
			wantCanAdd: false,
		},
		{
			name:  "can add - enterprise unlimited",
			orgID: 1,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
				Plan: models.OrgPlanEnterprise, // unlimited seats
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Status: models.MemberStatusActive},
			},
			wantCanAdd: true,
		},
		{
			name:  "counts pending invitations",
			orgID: 1,
			setupOrg: &models.Organization{
				ID:   1,
				Name: "Test",
				Slug: "test",
				Plan: models.OrgPlanFree, // 5 seat limit
			},
			setupMembers: []models.OrganizationMember{
				{OrganizationID: 1, UserID: 1, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 2, Status: models.MemberStatusActive},
				{OrganizationID: 1, UserID: 3, Status: models.MemberStatusActive},
			},
			setupInvites: []models.OrganizationInvitation{
				{OrganizationID: 1, Email: "a@test.com", Token: "t1", ExpiresAt: time.Now().Add(24 * time.Hour)},
				{OrganizationID: 1, Email: "b@test.com", Token: "t2", ExpiresAt: time.Now().Add(24 * time.Hour)},
			},
			wantCanAdd: false, // 3 members + 2 invites = 5 (at limit)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, memberRepo, invRepo, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			for _, m := range tt.setupMembers {
				memberRepo.AddMember(m)
			}
			for _, inv := range tt.setupInvites {
				invRepo.AddInvitation(inv)
			}

			canAdd, err := svc.CanAddMember(context.Background(), tt.orgID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCanAdd, canAdd)
		})
	}
}

// ============ Subscription Tests ============

func TestOrgService_CreateOrganizationSubscription(t *testing.T) {
	orgID := uint(1)

	tests := []struct {
		name     string
		sub      *models.Subscription
		setupErr error
		wantErr  bool
	}{
		{
			name: "success",
			sub: &models.Subscription{
				OrganizationID: &orgID,
				Status:         "active",
			},
		},
		{
			name: "error",
			sub: &models.Subscription{
				OrganizationID: &orgID,
				Status:         "active",
			},
			setupErr: errors.New("database error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, subRepo, _ := newTestOrgService()

			if tt.setupErr != nil {
				subRepo.CreateErr = tt.setupErr
			}

			err := svc.CreateOrganizationSubscription(context.Background(), tt.sub)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestOrgService_UpdateOrganizationSubscription(t *testing.T) {
	orgID := uint(1)

	tests := []struct {
		name     string
		sub      *models.Subscription
		setupSub *models.Subscription
		setupErr error
		wantErr  bool
	}{
		{
			name: "success",
			sub: &models.Subscription{
				ID:             1,
				OrganizationID: &orgID,
				Status:         "canceled",
			},
			setupSub: &models.Subscription{
				ID:             1,
				OrganizationID: &orgID,
				Status:         "active",
			},
		},
		{
			name: "error",
			sub: &models.Subscription{
				ID:             999,
				OrganizationID: &orgID,
				Status:         "canceled",
			},
			setupErr: errors.New("not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, subRepo, _ := newTestOrgService()

			if tt.setupSub != nil {
				subRepo.AddSubscription(*tt.setupSub)
			}
			if tt.setupErr != nil {
				subRepo.UpdateErr = tt.setupErr
			}

			err := svc.UpdateOrganizationSubscription(context.Background(), tt.sub)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestOrgService_GetOrganizationByStripeSubscriptionID(t *testing.T) {
	stripeSubID := "sub_123"

	tests := []struct {
		name        string
		stripeSubID string
		setupOrg    *models.Organization
		setupErr    error
		wantErr     error
	}{
		{
			name:        "success",
			stripeSubID: stripeSubID,
			setupOrg: &models.Organization{
				ID:                   1,
				Name:                 "Test",
				Slug:                 "test",
				StripeSubscriptionID: &stripeSubID,
			},
		},
		{
			name:        "not found",
			stripeSubID: "sub_nonexistent",
			setupErr:    gorm.ErrRecordNotFound,
			wantErr:     ErrOrgNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, orgRepo, _, _, _, _ := newTestOrgService()

			if tt.setupOrg != nil {
				orgRepo.AddOrganization(tt.setupOrg)
			}
			if tt.setupErr != nil {
				orgRepo.FindByStripeSubscriptionIDErr = tt.setupErr
			}

			org, err := svc.GetOrganizationByStripeSubscriptionID(context.Background(), tt.stripeSubID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.stripeSubID, *org.StripeSubscriptionID)
		})
	}
}
