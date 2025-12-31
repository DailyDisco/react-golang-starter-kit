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
