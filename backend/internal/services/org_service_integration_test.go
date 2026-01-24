package services

import (
	"context"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"

	"gorm.io/gorm"
)

// testOrgSetup creates the service and returns cleanup function
func testOrgSetup(t *testing.T) (*OrgService, *gorm.DB, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)
	svc := NewOrgService(tt.DB)

	return svc, tt.DB, func() {
		tt.Rollback()
	}
}

// createTestUser creates a user for testing
func createTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    email,
		Name:     "Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestOrgService_CreateOrganization_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("creates organization with owner membership", func(t *testing.T) {
		user := createTestUser(t, db, "owner@example.com")

		org, err := svc.CreateOrganization(context.Background(), user.ID, "Test Org", "test-org")
		if err != nil {
			t.Fatalf("CreateOrganization failed: %v", err)
		}

		if org.ID == 0 {
			t.Error("Expected org to have ID")
		}
		if org.Name != "Test Org" {
			t.Errorf("Expected name 'Test Org', got: %s", org.Name)
		}
		if org.Slug != "test-org" {
			t.Errorf("Expected slug 'test-org', got: %s", org.Slug)
		}
		if org.Plan != models.OrgPlanFree {
			t.Errorf("Expected plan 'free', got: %s", org.Plan)
		}
		if org.CreatedByUserID != user.ID {
			t.Errorf("Expected CreatedByUserID %d, got: %d", user.ID, org.CreatedByUserID)
		}

		// Verify owner membership was created
		var member models.OrganizationMember
		if err := db.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).First(&member).Error; err != nil {
			t.Fatalf("Failed to find owner membership: %v", err)
		}
		if member.Role != models.OrgRoleOwner {
			t.Errorf("Expected role 'owner', got: %s", member.Role)
		}
		if member.Status != models.MemberStatusActive {
			t.Errorf("Expected status 'active', got: %s", member.Status)
		}
		if member.AcceptedAt == nil {
			t.Error("Expected AcceptedAt to be set")
		}
	})

	t.Run("normalizes slug to lowercase", func(t *testing.T) {
		user := createTestUser(t, db, "user2@example.com")

		org, err := svc.CreateOrganization(context.Background(), user.ID, "My Org", "My-ORG")
		if err != nil {
			t.Fatalf("CreateOrganization failed: %v", err)
		}

		if org.Slug != "my-org" {
			t.Errorf("Expected slug to be lowercase 'my-org', got: %s", org.Slug)
		}
	})

	t.Run("rejects invalid slug format", func(t *testing.T) {
		user := createTestUser(t, db, "user3@example.com")

		tests := []struct {
			name string
			slug string
		}{
			{"starts with hyphen", "-invalid"},
			{"ends with hyphen", "invalid-"},
			{"has spaces", "invalid slug"},
			// Note: uppercase is normalized to lowercase, not rejected (tested in normalizes_slug_to_lowercase)
			{"has special chars", "invalid@slug"},
			{"double hyphen", "invalid--slug"},
			{"empty", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := svc.CreateOrganization(context.Background(), user.ID, "Test", tt.slug)
				if err != ErrInvalidSlug {
					t.Errorf("Expected ErrInvalidSlug for slug %q, got: %v", tt.slug, err)
				}
			})
		}
	})

	t.Run("rejects duplicate slug", func(t *testing.T) {
		user1 := createTestUser(t, db, "user4@example.com")
		user2 := createTestUser(t, db, "user5@example.com")

		_, err := svc.CreateOrganization(context.Background(), user1.ID, "First Org", "duplicate-slug")
		if err != nil {
			t.Fatalf("First CreateOrganization failed: %v", err)
		}

		_, err = svc.CreateOrganization(context.Background(), user2.ID, "Second Org", "duplicate-slug")
		if err != ErrOrgSlugTaken {
			t.Errorf("Expected ErrOrgSlugTaken, got: %v", err)
		}
	})

	t.Run("trims name whitespace", func(t *testing.T) {
		user := createTestUser(t, db, "user6@example.com")

		org, err := svc.CreateOrganization(context.Background(), user.ID, "  Trimmed Name  ", "trimmed-org")
		if err != nil {
			t.Fatalf("CreateOrganization failed: %v", err)
		}

		if org.Name != "Trimmed Name" {
			t.Errorf("Expected trimmed name, got: %q", org.Name)
		}
	})
}

func TestOrgService_GetOrganization_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns organization by slug", func(t *testing.T) {
		user := createTestUser(t, db, "owner@example.com")
		created, _ := svc.CreateOrganization(context.Background(), user.ID, "Test Org", "test-org")

		org, err := svc.GetOrganization(context.Background(), "test-org")
		if err != nil {
			t.Fatalf("GetOrganization failed: %v", err)
		}

		if org.ID != created.ID {
			t.Errorf("Expected ID %d, got: %d", created.ID, org.ID)
		}
	})

	t.Run("returns ErrOrgNotFound for non-existent slug", func(t *testing.T) {
		_, err := svc.GetOrganization(context.Background(), "non-existent")
		if err != ErrOrgNotFound {
			t.Errorf("Expected ErrOrgNotFound, got: %v", err)
		}
	})
}

func TestOrgService_GetOrganizationByID_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns organization by ID", func(t *testing.T) {
		user := createTestUser(t, db, "owner@example.com")
		created, _ := svc.CreateOrganization(context.Background(), user.ID, "Test Org", "test-org")

		org, err := svc.GetOrganizationByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetOrganizationByID failed: %v", err)
		}

		if org.Slug != "test-org" {
			t.Errorf("Expected slug 'test-org', got: %s", org.Slug)
		}
	})

	t.Run("returns ErrOrgNotFound for non-existent ID", func(t *testing.T) {
		_, err := svc.GetOrganizationByID(context.Background(), 99999)
		if err != ErrOrgNotFound {
			t.Errorf("Expected ErrOrgNotFound, got: %v", err)
		}
	})
}

func TestOrgService_GetOrganizationWithMembers_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns organization with preloaded members", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		member := createTestUser(t, db, "member@example.com")

		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		// Add another member
		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         member.ID,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		result, err := svc.GetOrganizationWithMembers(context.Background(), "test-org")
		if err != nil {
			t.Fatalf("GetOrganizationWithMembers failed: %v", err)
		}

		if len(result.Members) != 2 {
			t.Errorf("Expected 2 members, got: %d", len(result.Members))
		}

		// Verify users are preloaded
		for _, m := range result.Members {
			if m.User.Email == "" {
				t.Error("Expected User to be preloaded")
			}
		}
	})
}

func TestOrgService_GetUserOrganizations_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns all organizations user is member of", func(t *testing.T) {
		user := createTestUser(t, db, "user@example.com")
		otherUser := createTestUser(t, db, "other@example.com")

		// Create 3 orgs - user owns 2, other owns 1
		svc.CreateOrganization(context.Background(), user.ID, "Org 1", "org-1")
		svc.CreateOrganization(context.Background(), user.ID, "Org 2", "org-2")
		svc.CreateOrganization(context.Background(), otherUser.ID, "Org 3", "org-3")

		orgs, err := svc.GetUserOrganizations(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("GetUserOrganizations failed: %v", err)
		}

		if len(orgs) != 2 {
			t.Errorf("Expected 2 organizations, got: %d", len(orgs))
		}
	})

	t.Run("excludes inactive memberships", func(t *testing.T) {
		user := createTestUser(t, db, "inactive@example.com")
		org, _ := svc.CreateOrganization(context.Background(), user.ID, "Test Org", "inactive-test")

		// Set membership to inactive
		db.Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND user_id = ?", org.ID, user.ID).
			Update("status", models.MemberStatusInactive)

		orgs, err := svc.GetUserOrganizations(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("GetUserOrganizations failed: %v", err)
		}

		if len(orgs) != 0 {
			t.Errorf("Expected 0 organizations (inactive), got: %d", len(orgs))
		}
	})
}

func TestOrgService_GetUserOrganizationsWithRoles_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns organizations with user roles", func(t *testing.T) {
		user := createTestUser(t, db, "user@example.com")
		otherOwner := createTestUser(t, db, "owner@example.com")

		// User owns org1
		svc.CreateOrganization(context.Background(), user.ID, "Owned Org", "owned-org")

		// User is member of org2
		org2, _ := svc.CreateOrganization(context.Background(), otherOwner.ID, "Member Org", "member-org")
		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org2.ID,
			UserID:         user.ID,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		orgsWithRoles, err := svc.GetUserOrganizationsWithRoles(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("GetUserOrganizationsWithRoles failed: %v", err)
		}

		if len(orgsWithRoles) != 2 {
			t.Errorf("Expected 2 organizations, got: %d", len(orgsWithRoles))
		}

		// Verify roles are correct
		roleMap := make(map[string]models.OrganizationRole)
		for _, owr := range orgsWithRoles {
			roleMap[owr.Organization.Slug] = owr.Role
		}

		if roleMap["owned-org"] != models.OrgRoleOwner {
			t.Errorf("Expected owner role for owned-org, got: %s", roleMap["owned-org"])
		}
		if roleMap["member-org"] != models.OrgRoleMember {
			t.Errorf("Expected member role for member-org, got: %s", roleMap["member-org"])
		}
	})
}

func TestOrgService_GetUserMembership_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns membership for member", func(t *testing.T) {
		user := createTestUser(t, db, "user@example.com")
		org, _ := svc.CreateOrganization(context.Background(), user.ID, "Test Org", "test-org")

		membership, err := svc.GetUserMembership(context.Background(), org.ID, user.ID)
		if err != nil {
			t.Fatalf("GetUserMembership failed: %v", err)
		}

		if membership.Role != models.OrgRoleOwner {
			t.Errorf("Expected owner role, got: %s", membership.Role)
		}
	})

	t.Run("returns ErrNotMember for non-member", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		nonMember := createTestUser(t, db, "nonmember@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-nm")

		_, err := svc.GetUserMembership(context.Background(), org.ID, nonMember.ID)
		if err != ErrNotMember {
			t.Errorf("Expected ErrNotMember, got: %v", err)
		}
	})
}

func TestOrgService_UpdateOrganization_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("updates organization name", func(t *testing.T) {
		user := createTestUser(t, db, "user@example.com")
		org, _ := svc.CreateOrganization(context.Background(), user.ID, "Old Name", "test-org")

		err := svc.UpdateOrganization(context.Background(), org, "New Name")
		if err != nil {
			t.Fatalf("UpdateOrganization failed: %v", err)
		}

		// Verify in database
		var updated models.Organization
		db.First(&updated, org.ID)
		if updated.Name != "New Name" {
			t.Errorf("Expected 'New Name', got: %s", updated.Name)
		}
	})

	t.Run("trims name whitespace", func(t *testing.T) {
		user := createTestUser(t, db, "user2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), user.ID, "Test", "test-org-2")

		svc.UpdateOrganization(context.Background(), org, "  Trimmed  ")

		var updated models.Organization
		db.First(&updated, org.ID)
		if updated.Name != "Trimmed" {
			t.Errorf("Expected trimmed name, got: %q", updated.Name)
		}
	})
}

func TestOrgService_DeleteOrganization_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("deletes organization and related data", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		member := createTestUser(t, db, "member@example.com")

		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		// Add a member
		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         member.ID,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		// Add an invitation
		db.Create(&models.OrganizationInvitation{
			OrganizationID:  org.ID,
			Email:           "invited@example.com",
			Role:            models.OrgRoleMember,
			Token:           "test-token",
			InvitedByUserID: owner.ID,
			ExpiresAt:       time.Now().Add(7 * 24 * time.Hour),
		})

		err := svc.DeleteOrganization(context.Background(), org)
		if err != nil {
			t.Fatalf("DeleteOrganization failed: %v", err)
		}

		// Verify org is deleted
		var count int64
		db.Model(&models.Organization{}).Where("id = ?", org.ID).Count(&count)
		if count != 0 {
			t.Error("Expected organization to be deleted")
		}

		// Verify members are deleted
		db.Model(&models.OrganizationMember{}).Where("organization_id = ?", org.ID).Count(&count)
		if count != 0 {
			t.Error("Expected members to be deleted")
		}

		// Verify invitations are deleted
		db.Model(&models.OrganizationInvitation{}).Where("organization_id = ?", org.ID).Count(&count)
		if count != 0 {
			t.Error("Expected invitations to be deleted")
		}
	})
}

func TestOrgService_GetMembers_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns all members with preloaded users", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		member1 := createTestUser(t, db, "member1@example.com")
		member2 := createTestUser(t, db, "member2@example.com")

		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		now := time.Now()
		for _, m := range []*models.User{member1, member2} {
			db.Create(&models.OrganizationMember{
				OrganizationID: org.ID,
				UserID:         m.ID,
				Role:           models.OrgRoleMember,
				Status:         models.MemberStatusActive,
				AcceptedAt:     &now,
			})
		}

		members, err := svc.GetMembers(context.Background(), org.ID)
		if err != nil {
			t.Fatalf("GetMembers failed: %v", err)
		}

		if len(members) != 3 {
			t.Errorf("Expected 3 members, got: %d", len(members))
		}

		// Verify users are preloaded
		for _, m := range members {
			if m.User.Email == "" {
				t.Error("Expected User to be preloaded")
			}
		}
	})
}

func TestOrgService_UpdateMemberRole_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("updates member role", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		member := createTestUser(t, db, "member@example.com")

		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         member.ID,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		err := svc.UpdateMemberRole(context.Background(), org.ID, member.ID, owner.ID, models.OrgRoleAdmin)
		if err != nil {
			t.Fatalf("UpdateMemberRole failed: %v", err)
		}

		var updated models.OrganizationMember
		db.Where("organization_id = ? AND user_id = ?", org.ID, member.ID).First(&updated)
		if updated.Role != models.OrgRoleAdmin {
			t.Errorf("Expected admin role, got: %s", updated.Role)
		}
	})

	t.Run("cannot change own role", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		err := svc.UpdateMemberRole(context.Background(), org.ID, owner.ID, owner.ID, models.OrgRoleMember)
		if err != ErrCannotChangeOwnRole {
			t.Errorf("Expected ErrCannotChangeOwnRole, got: %v", err)
		}
	})

	t.Run("cannot demote last owner", func(t *testing.T) {
		owner := createTestUser(t, db, "owner3@example.com")
		admin := createTestUser(t, db, "admin@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-3")

		// Add admin to demote owner
		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         admin.ID,
			Role:           models.OrgRoleAdmin,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		err := svc.UpdateMemberRole(context.Background(), org.ID, owner.ID, admin.ID, models.OrgRoleMember)
		if err != ErrMustHaveOwner {
			t.Errorf("Expected ErrMustHaveOwner, got: %v", err)
		}
	})

	t.Run("allows demoting owner when another owner exists", func(t *testing.T) {
		owner1 := createTestUser(t, db, "owner4@example.com")
		owner2 := createTestUser(t, db, "owner5@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner1.ID, "Test Org", "test-org-4")

		// Add second owner
		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         owner2.ID,
			Role:           models.OrgRoleOwner,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		err := svc.UpdateMemberRole(context.Background(), org.ID, owner1.ID, owner2.ID, models.OrgRoleMember)
		if err != nil {
			t.Fatalf("Expected success when demoting with another owner, got: %v", err)
		}
	})

	t.Run("returns ErrNotMember for non-member", func(t *testing.T) {
		owner := createTestUser(t, db, "owner6@example.com")
		nonMember := createTestUser(t, db, "nonmember@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-5")

		err := svc.UpdateMemberRole(context.Background(), org.ID, nonMember.ID, owner.ID, models.OrgRoleAdmin)
		if err != ErrNotMember {
			t.Errorf("Expected ErrNotMember, got: %v", err)
		}
	})
}

func TestOrgService_RemoveMember_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("removes member from organization", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		member := createTestUser(t, db, "member@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         member.ID,
			Role:           models.OrgRoleMember,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		err := svc.RemoveMember(context.Background(), org.ID, member.ID)
		if err != nil {
			t.Fatalf("RemoveMember failed: %v", err)
		}

		var count int64
		db.Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND user_id = ?", org.ID, member.ID).
			Count(&count)
		if count != 0 {
			t.Error("Expected member to be removed")
		}
	})

	t.Run("cannot remove last owner", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		err := svc.RemoveMember(context.Background(), org.ID, owner.ID)
		if err != ErrCannotRemoveOwner {
			t.Errorf("Expected ErrCannotRemoveOwner, got: %v", err)
		}
	})

	t.Run("allows removing owner when another owner exists", func(t *testing.T) {
		owner1 := createTestUser(t, db, "owner3@example.com")
		owner2 := createTestUser(t, db, "owner4@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner1.ID, "Test Org", "test-org-3")

		now := time.Now()
		db.Create(&models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         owner2.ID,
			Role:           models.OrgRoleOwner,
			Status:         models.MemberStatusActive,
			AcceptedAt:     &now,
		})

		err := svc.RemoveMember(context.Background(), org.ID, owner1.ID)
		if err != nil {
			t.Fatalf("Expected success removing owner with another owner, got: %v", err)
		}
	})

	t.Run("returns ErrNotMember for non-member", func(t *testing.T) {
		owner := createTestUser(t, db, "owner5@example.com")
		nonMember := createTestUser(t, db, "nonmember@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-4")

		err := svc.RemoveMember(context.Background(), org.ID, nonMember.ID)
		if err != ErrNotMember {
			t.Errorf("Expected ErrNotMember, got: %v", err)
		}
	})
}

func TestOrgService_CreateInvitation_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("creates invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		invitation, err := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invite@example.com", models.OrgRoleMember)
		if err != nil {
			t.Fatalf("CreateInvitation failed: %v", err)
		}

		if invitation.ID == 0 {
			t.Error("Expected invitation to have ID")
		}
		if invitation.Email != "invite@example.com" {
			t.Errorf("Expected email 'invite@example.com', got: %s", invitation.Email)
		}
		if invitation.Role != models.OrgRoleMember {
			t.Errorf("Expected role 'member', got: %s", invitation.Role)
		}
		if invitation.Token == "" {
			t.Error("Expected token to be generated")
		}
		if invitation.ExpiresAt.Before(time.Now()) {
			t.Error("Expected expiration to be in the future")
		}
	})

	t.Run("normalizes email to lowercase", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		invitation, err := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "UPPERCASE@EXAMPLE.COM", models.OrgRoleMember)
		if err != nil {
			t.Fatalf("CreateInvitation failed: %v", err)
		}

		if invitation.Email != "uppercase@example.com" {
			t.Errorf("Expected lowercase email, got: %s", invitation.Email)
		}
	})

	t.Run("rejects invitation for existing member", func(t *testing.T) {
		owner := createTestUser(t, db, "owner3@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-3")

		_, err := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "owner3@example.com", models.OrgRoleMember)
		if err != ErrAlreadyMember {
			t.Errorf("Expected ErrAlreadyMember, got: %v", err)
		}
	})

	t.Run("rejects duplicate pending invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner4@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-4")

		_, err := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "duplicate@example.com", models.OrgRoleMember)
		if err != nil {
			t.Fatalf("First invitation failed: %v", err)
		}

		_, err = svc.CreateInvitation(context.Background(), org.ID, owner.ID, "duplicate@example.com", models.OrgRoleAdmin)
		if err != ErrInvitationEmailTaken {
			t.Errorf("Expected ErrInvitationEmailTaken, got: %v", err)
		}
	})
}

func TestOrgService_GetInvitationByToken_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns invitation by token", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		created, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invite@example.com", models.OrgRoleMember)

		invitation, err := svc.GetInvitationByToken(context.Background(), created.Token)
		if err != nil {
			t.Fatalf("GetInvitationByToken failed: %v", err)
		}

		if invitation.ID != created.ID {
			t.Errorf("Expected ID %d, got: %d", created.ID, invitation.ID)
		}

		// Verify organization is preloaded
		if invitation.Organization.Slug == "" {
			t.Error("Expected Organization to be preloaded")
		}
	})

	t.Run("returns ErrInvitationNotFound for invalid token", func(t *testing.T) {
		_, err := svc.GetInvitationByToken(context.Background(), "invalid-token")
		if err != ErrInvitationNotFound {
			t.Errorf("Expected ErrInvitationNotFound, got: %v", err)
		}
	})
}

func TestOrgService_AcceptInvitation_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("accepts invitation and creates membership", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		invitee := createTestUser(t, db, "invitee@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invitee@example.com", models.OrgRoleAdmin)

		membership, err := svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)
		if err != nil {
			t.Fatalf("AcceptInvitation failed: %v", err)
		}

		if membership.Role != models.OrgRoleAdmin {
			t.Errorf("Expected role 'admin', got: %s", membership.Role)
		}
		if membership.Status != models.MemberStatusActive {
			t.Errorf("Expected status 'active', got: %s", membership.Status)
		}
		if membership.AcceptedAt == nil {
			t.Error("Expected AcceptedAt to be set")
		}

		// Verify invitation is marked as accepted
		var updated models.OrganizationInvitation
		db.First(&updated, invitation.ID)
		if updated.AcceptedAt == nil {
			t.Error("Expected invitation AcceptedAt to be set")
		}
	})

	t.Run("rejects expired invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		invitee := createTestUser(t, db, "invitee2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		// Create expired invitation directly
		invitation := &models.OrganizationInvitation{
			OrganizationID:  org.ID,
			Email:           "invitee2@example.com",
			Role:            models.OrgRoleMember,
			Token:           "expired-token",
			InvitedByUserID: owner.ID,
			ExpiresAt:       time.Now().Add(-1 * time.Hour), // Expired
		}
		db.Create(invitation)

		_, err := svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)
		if err != ErrInvitationExpired {
			t.Errorf("Expected ErrInvitationExpired, got: %v", err)
		}
	})

	t.Run("rejects already accepted invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner3@example.com")
		invitee := createTestUser(t, db, "invitee3@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-3")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invitee3@example.com", models.OrgRoleMember)

		// Accept first time
		svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)

		// Try to accept again
		_, err := svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)
		if err != ErrInvitationAccepted {
			t.Errorf("Expected ErrInvitationAccepted, got: %v", err)
		}
	})

	t.Run("rejects if email doesn't match", func(t *testing.T) {
		owner := createTestUser(t, db, "owner4@example.com")
		wrongUser := createTestUser(t, db, "wrong@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-4")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "correct@example.com", models.OrgRoleMember)

		_, err := svc.AcceptInvitation(context.Background(), invitation.Token, wrongUser.ID)
		if err == nil || err.Error() != "email does not match invitation" {
			t.Errorf("Expected email mismatch error, got: %v", err)
		}
	})

	t.Run("rejects if already a member", func(t *testing.T) {
		owner := createTestUser(t, db, "owner5@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-5")

		// Create invitation for owner (who is already a member)
		invitation := &models.OrganizationInvitation{
			OrganizationID:  org.ID,
			Email:           "owner5@example.com",
			Role:            models.OrgRoleMember,
			Token:           "owner-token",
			InvitedByUserID: owner.ID,
			ExpiresAt:       time.Now().Add(7 * 24 * time.Hour),
		}
		db.Create(invitation)

		_, err := svc.AcceptInvitation(context.Background(), invitation.Token, owner.ID)
		if err != ErrAlreadyMember {
			t.Errorf("Expected ErrAlreadyMember, got: %v", err)
		}
	})
}

func TestOrgService_GetPendingInvitations_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("returns pending invitations", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invite1@example.com", models.OrgRoleMember)
		svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invite2@example.com", models.OrgRoleAdmin)

		invitations, err := svc.GetPendingInvitations(context.Background(), org.ID)
		if err != nil {
			t.Fatalf("GetPendingInvitations failed: %v", err)
		}

		if len(invitations) != 2 {
			t.Errorf("Expected 2 pending invitations, got: %d", len(invitations))
		}

		// Verify InvitedByUser is preloaded
		for _, inv := range invitations {
			if inv.InvitedByUser.Email == "" {
				t.Error("Expected InvitedByUser to be preloaded")
			}
		}
	})

	t.Run("excludes accepted invitations", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		invitee := createTestUser(t, db, "invitee@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invitee@example.com", models.OrgRoleMember)
		svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)

		invitations, _ := svc.GetPendingInvitations(context.Background(), org.ID)
		if len(invitations) != 0 {
			t.Errorf("Expected 0 pending invitations, got: %d", len(invitations))
		}
	})

	t.Run("excludes expired invitations", func(t *testing.T) {
		owner := createTestUser(t, db, "owner3@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-3")

		// Create expired invitation
		db.Create(&models.OrganizationInvitation{
			OrganizationID:  org.ID,
			Email:           "expired@example.com",
			Role:            models.OrgRoleMember,
			Token:           "expired-token",
			InvitedByUserID: owner.ID,
			ExpiresAt:       time.Now().Add(-1 * time.Hour),
		})

		invitations, _ := svc.GetPendingInvitations(context.Background(), org.ID)
		if len(invitations) != 0 {
			t.Errorf("Expected 0 pending invitations (expired), got: %d", len(invitations))
		}
	})
}

func TestOrgService_CancelInvitation_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("cancels pending invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invite@example.com", models.OrgRoleMember)

		err := svc.CancelInvitation(context.Background(), invitation.ID, org.ID)
		if err != nil {
			t.Fatalf("CancelInvitation failed: %v", err)
		}

		var count int64
		db.Model(&models.OrganizationInvitation{}).Where("id = ?", invitation.ID).Count(&count)
		if count != 0 {
			t.Error("Expected invitation to be deleted")
		}
	})

	t.Run("returns error for non-existent invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		err := svc.CancelInvitation(context.Background(), 99999, org.ID)
		if err != ErrInvitationNotFound {
			t.Errorf("Expected ErrInvitationNotFound, got: %v", err)
		}
	})

	t.Run("cannot cancel accepted invitation", func(t *testing.T) {
		owner := createTestUser(t, db, "owner3@example.com")
		invitee := createTestUser(t, db, "invitee@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-3")

		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invitee@example.com", models.OrgRoleMember)
		svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)

		err := svc.CancelInvitation(context.Background(), invitation.ID, org.ID)
		if err != ErrInvitationNotFound {
			t.Errorf("Expected ErrInvitationNotFound for accepted invitation, got: %v", err)
		}
	})

	t.Run("cannot cancel invitation from different org", func(t *testing.T) {
		owner := createTestUser(t, db, "owner4@example.com")
		org1, _ := svc.CreateOrganization(context.Background(), owner.ID, "Org 1", "org-1")
		org2, _ := svc.CreateOrganization(context.Background(), owner.ID, "Org 2", "org-2")

		invitation, _ := svc.CreateInvitation(context.Background(), org1.ID, owner.ID, "invite@example.com", models.OrgRoleMember)

		err := svc.CancelInvitation(context.Background(), invitation.ID, org2.ID)
		if err != ErrInvitationNotFound {
			t.Errorf("Expected ErrInvitationNotFound for wrong org, got: %v", err)
		}
	})
}

func TestOrgService_CleanupExpiredInvitations_Integration(t *testing.T) {
	svc, db, cleanup := testOrgSetup(t)
	defer cleanup()

	t.Run("removes expired invitations", func(t *testing.T) {
		owner := createTestUser(t, db, "owner@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org")

		// Create expired invitation
		db.Create(&models.OrganizationInvitation{
			OrganizationID:  org.ID,
			Email:           "expired@example.com",
			Role:            models.OrgRoleMember,
			Token:           "expired-token",
			InvitedByUserID: owner.ID,
			ExpiresAt:       time.Now().Add(-1 * time.Hour),
		})

		// Create valid invitation
		svc.CreateInvitation(context.Background(), org.ID, owner.ID, "valid@example.com", models.OrgRoleMember)

		err := svc.CleanupExpiredInvitations(context.Background())
		if err != nil {
			t.Fatalf("CleanupExpiredInvitations failed: %v", err)
		}

		var invitations []models.OrganizationInvitation
		db.Where("organization_id = ?", org.ID).Find(&invitations)

		if len(invitations) != 1 {
			t.Errorf("Expected 1 invitation (valid), got: %d", len(invitations))
		}

		if len(invitations) > 0 && invitations[0].Email != "valid@example.com" {
			t.Error("Expected valid invitation to remain")
		}
	})

	t.Run("does not remove accepted invitations", func(t *testing.T) {
		owner := createTestUser(t, db, "owner2@example.com")
		invitee := createTestUser(t, db, "invitee@example.com")
		org, _ := svc.CreateOrganization(context.Background(), owner.ID, "Test Org", "test-org-2")

		// Create and accept invitation, then manually set it to expired
		invitation, _ := svc.CreateInvitation(context.Background(), org.ID, owner.ID, "invitee@example.com", models.OrgRoleMember)
		svc.AcceptInvitation(context.Background(), invitation.Token, invitee.ID)

		// Update to expired (shouldn't matter since it's accepted)
		db.Model(&invitation).Update("expires_at", time.Now().Add(-1*time.Hour))

		svc.CleanupExpiredInvitations(context.Background())

		var count int64
		db.Model(&models.OrganizationInvitation{}).Where("id = ?", invitation.ID).Count(&count)
		if count != 1 {
			t.Error("Expected accepted invitation to remain")
		}
	})
}
