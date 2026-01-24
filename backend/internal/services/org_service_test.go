package services

import (
	"testing"

	"react-golang-starter/internal/models"
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
