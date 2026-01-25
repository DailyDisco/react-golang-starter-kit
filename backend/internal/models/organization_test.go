package models

import (
	"testing"
	"time"
)

// ============ Table Name Tests ============

func TestOrganization_TableName(t *testing.T) {
	org := Organization{}
	if got := org.TableName(); got != "organizations" {
		t.Errorf("Organization.TableName() = %q, want %q", got, "organizations")
	}
}

func TestOrganizationMember_TableName(t *testing.T) {
	member := OrganizationMember{}
	if got := member.TableName(); got != "organization_members" {
		t.Errorf("OrganizationMember.TableName() = %q, want %q", got, "organization_members")
	}
}

func TestOrganizationInvitation_TableName(t *testing.T) {
	invitation := OrganizationInvitation{}
	if got := invitation.TableName(); got != "organization_invitations" {
		t.Errorf("OrganizationInvitation.TableName() = %q, want %q", got, "organization_invitations")
	}
}

// ============ Role Permission Tests ============

func TestOrganizationRole_CanManageMembers(t *testing.T) {
	tests := []struct {
		name string
		role OrganizationRole
		want bool
	}{
		{"owner can manage members", OrgRoleOwner, true},
		{"admin can manage members", OrgRoleAdmin, true},
		{"member cannot manage members", OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanManageMembers(); got != tt.want {
				t.Errorf("%s.CanManageMembers() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestOrganizationRole_CanManageSettings(t *testing.T) {
	tests := []struct {
		name string
		role OrganizationRole
		want bool
	}{
		{"owner can manage settings", OrgRoleOwner, true},
		{"admin can manage settings", OrgRoleAdmin, true},
		{"member cannot manage settings", OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanManageSettings(); got != tt.want {
				t.Errorf("%s.CanManageSettings() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestOrganizationRole_CanDeleteOrganization(t *testing.T) {
	tests := []struct {
		name string
		role OrganizationRole
		want bool
	}{
		{"owner can delete org", OrgRoleOwner, true},
		{"admin cannot delete org", OrgRoleAdmin, false},
		{"member cannot delete org", OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanDeleteOrganization(); got != tt.want {
				t.Errorf("%s.CanDeleteOrganization() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestOrganizationRole_CanTransferOwnership(t *testing.T) {
	tests := []struct {
		name string
		role OrganizationRole
		want bool
	}{
		{"owner can transfer ownership", OrgRoleOwner, true},
		{"admin cannot transfer ownership", OrgRoleAdmin, false},
		{"member cannot transfer ownership", OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanTransferOwnership(); got != tt.want {
				t.Errorf("%s.CanTransferOwnership() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestOrganizationRole_IsHigherOrEqualTo(t *testing.T) {
	tests := []struct {
		name  string
		role  OrganizationRole
		other OrganizationRole
		want  bool
	}{
		// Owner comparisons
		{"owner >= owner", OrgRoleOwner, OrgRoleOwner, true},
		{"owner >= admin", OrgRoleOwner, OrgRoleAdmin, true},
		{"owner >= member", OrgRoleOwner, OrgRoleMember, true},

		// Admin comparisons
		{"admin >= owner", OrgRoleAdmin, OrgRoleOwner, false},
		{"admin >= admin", OrgRoleAdmin, OrgRoleAdmin, true},
		{"admin >= member", OrgRoleAdmin, OrgRoleMember, true},

		// Member comparisons
		{"member >= owner", OrgRoleMember, OrgRoleOwner, false},
		{"member >= admin", OrgRoleMember, OrgRoleAdmin, false},
		{"member >= member", OrgRoleMember, OrgRoleMember, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsHigherOrEqualTo(tt.other); got != tt.want {
				t.Errorf("%s.IsHigherOrEqualTo(%s) = %v, want %v", tt.role, tt.other, got, tt.want)
			}
		})
	}
}

// ============ Invitation Tests ============

func TestOrganizationInvitation_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"expired yesterday", time.Now().Add(-24 * time.Hour), true},
		{"expired 1 minute ago", time.Now().Add(-1 * time.Minute), true},
		{"expires in 1 minute", time.Now().Add(1 * time.Minute), false},
		{"expires tomorrow", time.Now().Add(24 * time.Hour), false},
		{"expires in 7 days", time.Now().Add(7 * 24 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := &OrganizationInvitation{ExpiresAt: tt.expiresAt}
			if got := invitation.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrganizationInvitation_IsAccepted(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		acceptedAt *time.Time
		want       bool
	}{
		{"nil accepted_at", nil, false},
		{"non-nil accepted_at", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invitation := &OrganizationInvitation{AcceptedAt: tt.acceptedAt}
			if got := invitation.IsAccepted(); got != tt.want {
				t.Errorf("IsAccepted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============ Plan Features Tests ============

func TestDefaultPlanFeatures(t *testing.T) {
	tests := []struct {
		name             string
		plan             OrganizationPlan
		wantSeatLimit    int
		wantStorageLimit int
		wantAPILimit     int
	}{
		{
			name:             "free plan",
			plan:             OrgPlanFree,
			wantSeatLimit:    5,
			wantStorageLimit: 1024,
			wantAPILimit:     10000,
		},
		{
			name:             "pro plan",
			plan:             OrgPlanPro,
			wantSeatLimit:    25,
			wantStorageLimit: 10240,
			wantAPILimit:     100000,
		},
		{
			name:             "enterprise plan unlimited",
			plan:             OrgPlanEnterprise,
			wantSeatLimit:    0,
			wantStorageLimit: 0,
			wantAPILimit:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := DefaultPlanFeatures(tt.plan)

			if features.SeatLimit != tt.wantSeatLimit {
				t.Errorf("SeatLimit = %d, want %d", features.SeatLimit, tt.wantSeatLimit)
			}
			if features.StorageLimit != tt.wantStorageLimit {
				t.Errorf("StorageLimit = %d, want %d", features.StorageLimit, tt.wantStorageLimit)
			}
			if features.APICallLimit != tt.wantAPILimit {
				t.Errorf("APICallLimit = %d, want %d", features.APICallLimit, tt.wantAPILimit)
			}
		})
	}
}

func TestDefaultPlanFeatures_UnknownPlan(t *testing.T) {
	// Unknown plans should default to free tier
	features := DefaultPlanFeatures(OrganizationPlan("unknown"))

	if features.SeatLimit != 5 {
		t.Errorf("Unknown plan SeatLimit = %d, want 5 (free default)", features.SeatLimit)
	}
}

func TestPlanFeatures_Hierarchy(t *testing.T) {
	free := DefaultPlanFeatures(OrgPlanFree)
	pro := DefaultPlanFeatures(OrgPlanPro)

	if pro.SeatLimit <= free.SeatLimit {
		t.Errorf("Pro SeatLimit (%d) should be > Free (%d)", pro.SeatLimit, free.SeatLimit)
	}
	if pro.StorageLimit <= free.StorageLimit {
		t.Errorf("Pro StorageLimit (%d) should be > Free (%d)", pro.StorageLimit, free.StorageLimit)
	}
	if pro.APICallLimit <= free.APICallLimit {
		t.Errorf("Pro APICallLimit (%d) should be > Free (%d)", pro.APICallLimit, free.APICallLimit)
	}
}

// ============ Organization Methods Tests ============

func TestOrganization_GetSeatLimit(t *testing.T) {
	tests := []struct {
		name string
		plan OrganizationPlan
		want int
	}{
		{"free plan", OrgPlanFree, 5},
		{"pro plan", OrgPlanPro, 25},
		{"enterprise unlimited", OrgPlanEnterprise, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{Plan: tt.plan}
			if got := org.GetSeatLimit(); got != tt.want {
				t.Errorf("GetSeatLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOrganization_HasSubscription(t *testing.T) {
	subID := "sub_123456"
	emptySubID := ""

	tests := []struct {
		name  string
		subID *string
		want  bool
	}{
		{"nil subscription ID", nil, false},
		{"empty subscription ID", &emptySubID, false},
		{"valid subscription ID", &subID, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{StripeSubscriptionID: tt.subID}
			if got := org.HasSubscription(); got != tt.want {
				t.Errorf("HasSubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============ Role Constants Tests ============

func TestOrganizationRole_Constants(t *testing.T) {
	tests := []struct {
		role OrganizationRole
		want string
	}{
		{OrgRoleOwner, "owner"},
		{OrgRoleAdmin, "admin"},
		{OrgRoleMember, "member"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("role = %q, want %q", tt.role, tt.want)
			}
		})
	}
}

func TestMemberStatus_Constants(t *testing.T) {
	tests := []struct {
		status MemberStatus
		want   string
	}{
		{MemberStatusActive, "active"},
		{MemberStatusInactive, "inactive"},
		{MemberStatusPending, "pending"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("status = %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestOrganizationPlan_Constants(t *testing.T) {
	tests := []struct {
		plan OrganizationPlan
		want string
	}{
		{OrgPlanFree, "free"},
		{OrgPlanPro, "pro"},
		{OrgPlanEnterprise, "enterprise"},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			if string(tt.plan) != tt.want {
				t.Errorf("plan = %q, want %q", tt.plan, tt.want)
			}
		})
	}
}

// ============ Type Uniqueness Tests ============

func TestOrganizationRole_UniqueValues(t *testing.T) {
	roles := []OrganizationRole{OrgRoleOwner, OrgRoleAdmin, OrgRoleMember}
	seen := make(map[string]bool)

	for _, role := range roles {
		str := string(role)
		if seen[str] {
			t.Errorf("Duplicate role value: %q", str)
		}
		seen[str] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected 3 unique roles, got %d", len(seen))
	}
}

func TestMemberStatus_UniqueValues(t *testing.T) {
	statuses := []MemberStatus{MemberStatusActive, MemberStatusInactive, MemberStatusPending}
	seen := make(map[string]bool)

	for _, status := range statuses {
		str := string(status)
		if seen[str] {
			t.Errorf("Duplicate status value: %q", str)
		}
		seen[str] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected 3 unique statuses, got %d", len(seen))
	}
}

func TestOrganizationPlan_UniqueValues(t *testing.T) {
	plans := []OrganizationPlan{OrgPlanFree, OrgPlanPro, OrgPlanEnterprise}
	seen := make(map[string]bool)

	for _, plan := range plans {
		str := string(plan)
		if seen[str] {
			t.Errorf("Duplicate plan value: %q", str)
		}
		seen[str] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected 3 unique plans, got %d", len(seen))
	}
}

// ============ Struct Field Tests ============

func TestOrganization_Fields(t *testing.T) {
	org := Organization{
		Name:            "Test Organization",
		Slug:            "test-organization",
		Plan:            OrgPlanPro,
		CreatedByUserID: 123,
	}

	if org.Name != "Test Organization" {
		t.Errorf("Name = %q, want %q", org.Name, "Test Organization")
	}
	if org.Slug != "test-organization" {
		t.Errorf("Slug = %q, want %q", org.Slug, "test-organization")
	}
	if org.Plan != OrgPlanPro {
		t.Errorf("Plan = %q, want %q", org.Plan, OrgPlanPro)
	}
	if org.CreatedByUserID != 123 {
		t.Errorf("CreatedByUserID = %d, want %d", org.CreatedByUserID, 123)
	}
}

func TestOrganizationMember_Fields(t *testing.T) {
	member := OrganizationMember{
		OrganizationID: 1,
		UserID:         2,
		Role:           OrgRoleAdmin,
		Status:         MemberStatusActive,
	}

	if member.OrganizationID != 1 {
		t.Errorf("OrganizationID = %d, want %d", member.OrganizationID, 1)
	}
	if member.UserID != 2 {
		t.Errorf("UserID = %d, want %d", member.UserID, 2)
	}
	if member.Role != OrgRoleAdmin {
		t.Errorf("Role = %q, want %q", member.Role, OrgRoleAdmin)
	}
	if member.Status != MemberStatusActive {
		t.Errorf("Status = %q, want %q", member.Status, MemberStatusActive)
	}
}

func TestOrganizationInvitation_Fields(t *testing.T) {
	invitation := OrganizationInvitation{
		OrganizationID:  1,
		Email:           "test@example.com",
		Role:            OrgRoleMember,
		Token:           "abc123",
		InvitedByUserID: 2,
	}

	if invitation.OrganizationID != 1 {
		t.Errorf("OrganizationID = %d, want %d", invitation.OrganizationID, 1)
	}
	if invitation.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", invitation.Email, "test@example.com")
	}
	if invitation.Role != OrgRoleMember {
		t.Errorf("Role = %q, want %q", invitation.Role, OrgRoleMember)
	}
	if invitation.Token != "abc123" {
		t.Errorf("Token = %q, want %q", invitation.Token, "abc123")
	}
}

// ============ Settings Structure Tests ============

func TestOrganizationSettings_Structure(t *testing.T) {
	settings := OrganizationSettings{
		AllowMemberInvites: true,
		DefaultRole:        "member",
		MaxMembers:         100,
	}
	settings.Features.AdvancedAnalytics = true
	settings.Features.CustomBranding = false
	settings.Features.APIAccess = true

	if !settings.AllowMemberInvites {
		t.Error("AllowMemberInvites should be true")
	}
	if settings.DefaultRole != "member" {
		t.Errorf("DefaultRole = %q, want %q", settings.DefaultRole, "member")
	}
	if settings.MaxMembers != 100 {
		t.Errorf("MaxMembers = %d, want %d", settings.MaxMembers, 100)
	}
	if !settings.Features.AdvancedAnalytics {
		t.Error("Features.AdvancedAnalytics should be true")
	}
	if settings.Features.CustomBranding {
		t.Error("Features.CustomBranding should be false")
	}
	if !settings.Features.APIAccess {
		t.Error("Features.APIAccess should be true")
	}
}
