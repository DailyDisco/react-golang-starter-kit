package stripe

import (
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	stripe "github.com/stripe/stripe-go/v76"
)

// Organization test helpers

func createTestOrganization(t *testing.T, slug, stripeCustomerID string) *models.Organization {
	t.Helper()

	// First create a user to be the org creator
	user := createTestUserForWebhook(t, slug+"@org.com", "cus_org_creator_"+slug)

	org := &models.Organization{
		Name:             "Test Org " + slug,
		Slug:             slug,
		Plan:             models.OrgPlanFree,
		StripeCustomerID: &stripeCustomerID,
		CreatedByUserID:  user.ID,
	}
	if err := database.DB.Create(org).Error; err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create owner membership
	createTestOrgMember(t, org.ID, user.ID, models.OrgRoleOwner)

	return org
}

func createTestOrgMember(t *testing.T, orgID, userID uint, role models.OrganizationRole) *models.OrganizationMember {
	t.Helper()
	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		Status:         models.MemberStatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := database.DB.Create(member).Error; err != nil {
		t.Fatalf("Failed to create org member: %v", err)
	}
	return member
}

func createTestOrgSubscription(t *testing.T, orgID, userID uint, stripeSubID, status string) *models.Subscription {
	t.Helper()
	sub := &models.Subscription{
		UserID:               userID,
		OrganizationID:       &orgID,
		StripeSubscriptionID: stripeSubID,
		StripePriceID:        "price_org_test_123",
		Status:               status,
		CurrentPeriodStart:   time.Now().Format(time.RFC3339),
		CurrentPeriodEnd:     time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		CreatedAt:            time.Now().Format(time.RFC3339),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}
	if err := database.DB.Create(sub).Error; err != nil {
		t.Fatalf("Failed to create org subscription: %v", err)
	}
	return sub
}

// Organization Billing Tests

func TestHandleOrgSubscriptionCreated_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("creates subscription record for organization", func(t *testing.T) {
		org := createTestOrganization(t, "newsub-org", "cus_org_new_123")

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_org_new",
			Customer: &stripe.Customer{
				ID: "cus_org_new_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_org_pro",
						},
					},
				},
			},
		}

		handleSubscriptionCreated(sub)

		// Verify subscription was created with org ID
		var dbSub models.Subscription
		err := database.DB.Where("stripe_subscription_id = ?", "sub_org_new").First(&dbSub).Error
		if err != nil {
			t.Fatalf("Expected subscription to be created: %v", err)
		}

		if dbSub.OrganizationID == nil || *dbSub.OrganizationID != org.ID {
			t.Errorf("Expected OrganizationID %d, got: %v", org.ID, dbSub.OrganizationID)
		}
		if dbSub.Status != "active" {
			t.Errorf("Expected status 'active', got: %s", dbSub.Status)
		}
	})

	t.Run("updates org plan via price ID mapping", func(t *testing.T) {
		org := createTestOrganization(t, "plan-update-org", "cus_org_plan_123")

		// Verify org starts on free plan
		if org.Plan != models.OrgPlanFree {
			t.Fatalf("Expected org to start on free plan, got: %s", org.Plan)
		}

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_org_plan_update",
			Customer: &stripe.Customer{
				ID: "cus_org_plan_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_org_pro_tier",
						},
					},
				},
			},
		}

		handleSubscriptionCreated(sub)

		// Verify org plan was updated
		var updatedOrg models.Organization
		database.DB.First(&updatedOrg, org.ID)
		if updatedOrg.Plan != models.OrgPlanPro {
			t.Errorf("Expected org plan '%s', got: %s", models.OrgPlanPro, updatedOrg.Plan)
		}

		// Verify stripe subscription ID was set
		if updatedOrg.StripeSubscriptionID == nil || *updatedOrg.StripeSubscriptionID != "sub_org_plan_update" {
			t.Errorf("Expected stripe subscription ID 'sub_org_plan_update', got: %v", updatedOrg.StripeSubscriptionID)
		}
	})

	t.Run("sets billing contact as org owner", func(t *testing.T) {
		org := createTestOrganization(t, "owner-billing-org", "cus_org_owner_123")

		// Get the org owner
		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_org_owner",
			Customer: &stripe.Customer{
				ID: "cus_org_owner_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_owner_test",
						},
					},
				},
			},
		}

		handleSubscriptionCreated(sub)

		// Verify subscription UserID is the org owner
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_org_owner").First(&dbSub)
		if dbSub.UserID != owner.UserID {
			t.Errorf("Expected subscription UserID to be org owner %d, got: %d", owner.UserID, dbSub.UserID)
		}
	})
}

func TestHandleOrgSubscriptionUpdated_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("updates org plan on tier change", func(t *testing.T) {
		org := createTestOrganization(t, "tier-change-org", "cus_org_tier_123")

		// Get owner for subscription creation
		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_tier_update", models.SubscriptionStatusActive)

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_org_tier_update",
			Customer: &stripe.Customer{
				ID: "cus_org_tier_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_enterprise", // New price
						},
					},
				},
			},
		}

		handleSubscriptionUpdated(sub)

		// Verify org plan was updated
		var updatedOrg models.Organization
		database.DB.First(&updatedOrg, org.ID)
		if updatedOrg.Plan != models.OrgPlanPro {
			t.Errorf("Expected org plan '%s', got: %s", models.OrgPlanPro, updatedOrg.Plan)
		}
	})

	t.Run("syncs subscription status correctly", func(t *testing.T) {
		org := createTestOrganization(t, "status-sync-org", "cus_org_status_123")

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_status", models.SubscriptionStatusActive)

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_org_status",
			Customer: &stripe.Customer{
				ID: "cus_org_status_123",
			},
			Status:             stripe.SubscriptionStatusPastDue,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
		}

		handleSubscriptionUpdated(sub)

		// Verify subscription status
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_org_status").First(&dbSub)
		if dbSub.Status != "past_due" {
			t.Errorf("Expected status 'past_due', got: %s", dbSub.Status)
		}
	})

	t.Run("handles price ID changes mid-cycle", func(t *testing.T) {
		org := createTestOrganization(t, "price-change-org", "cus_org_price_123")

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		sub := createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_price_change", models.SubscriptionStatusActive)
		originalPriceID := sub.StripePriceID

		now := time.Now().Unix()
		stripeSub := &stripe.Subscription{
			ID: "sub_org_price_change",
			Customer: &stripe.Customer{
				ID: "cus_org_price_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_new_tier_xyz",
						},
					},
				},
			},
		}

		handleSubscriptionUpdated(stripeSub)

		// Verify price ID was updated
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_org_price_change").First(&dbSub)
		if dbSub.StripePriceID == originalPriceID {
			t.Error("Expected price ID to be updated")
		}
		if dbSub.StripePriceID != "price_new_tier_xyz" {
			t.Errorf("Expected price ID 'price_new_tier_xyz', got: %s", dbSub.StripePriceID)
		}
	})
}

func TestHandleOrgSubscriptionDeleted_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("downgrades org to free plan", func(t *testing.T) {
		org := createTestOrganization(t, "downgrade-org", "cus_org_delete_123")

		// Set org to pro plan
		database.DB.Model(org).Update("plan", models.OrgPlanPro)

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_delete", models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: "sub_org_delete",
			Customer: &stripe.Customer{
				ID: "cus_org_delete_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify org was downgraded to free
		var updatedOrg models.Organization
		database.DB.First(&updatedOrg, org.ID)
		if updatedOrg.Plan != models.OrgPlanFree {
			t.Errorf("Expected org plan '%s', got: %s", models.OrgPlanFree, updatedOrg.Plan)
		}
	})

	t.Run("clears stripe_subscription_id", func(t *testing.T) {
		org := createTestOrganization(t, "clear-sub-org", "cus_org_clear_123")

		subID := "sub_org_clear"
		database.DB.Model(org).Update("stripe_subscription_id", subID)

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, subID, models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: subID,
			Customer: &stripe.Customer{
				ID: "cus_org_clear_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify stripe_subscription_id was cleared
		var updatedOrg models.Organization
		database.DB.First(&updatedOrg, org.ID)
		if updatedOrg.StripeSubscriptionID != nil {
			t.Errorf("Expected stripe_subscription_id to be nil, got: %v", updatedOrg.StripeSubscriptionID)
		}
	})

	t.Run("preserves org data (not deleted)", func(t *testing.T) {
		org := createTestOrganization(t, "preserve-org", "cus_org_preserve_123")
		originalName := org.Name
		originalSlug := org.Slug

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_preserve", models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: "sub_org_preserve",
			Customer: &stripe.Customer{
				ID: "cus_org_preserve_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify org still exists with original data
		var updatedOrg models.Organization
		err := database.DB.First(&updatedOrg, org.ID).Error
		if err != nil {
			t.Fatalf("Expected org to still exist: %v", err)
		}
		if updatedOrg.Name != originalName {
			t.Errorf("Expected org name '%s', got: %s", originalName, updatedOrg.Name)
		}
		if updatedOrg.Slug != originalSlug {
			t.Errorf("Expected org slug '%s', got: %s", originalSlug, updatedOrg.Slug)
		}
	})
}

func TestHandleOrgPaymentFailed_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("sets org subscription to past_due", func(t *testing.T) {
		org := createTestOrganization(t, "payment-fail-org", "cus_org_fail_123")

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_fail", models.SubscriptionStatusActive)

		invoice := &stripe.Invoice{
			ID: "inv_org_fail",
			Customer: &stripe.Customer{
				ID: "cus_org_fail_123",
			},
			Subscription: &stripe.Subscription{
				ID: "sub_org_fail",
			},
		}

		handlePaymentFailed(invoice)

		// Verify subscription status
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_org_fail").First(&dbSub)
		if dbSub.Status != models.SubscriptionStatusPastDue {
			t.Errorf("Expected status '%s', got: %s", models.SubscriptionStatusPastDue, dbSub.Status)
		}
	})

	t.Run("org plan remains until cancellation", func(t *testing.T) {
		org := createTestOrganization(t, "plan-remain-org", "cus_org_remain_123")
		database.DB.Model(org).Update("plan", models.OrgPlanPro)

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_org_remain", models.SubscriptionStatusActive)

		invoice := &stripe.Invoice{
			ID: "inv_org_remain",
			Customer: &stripe.Customer{
				ID: "cus_org_remain_123",
			},
			Subscription: &stripe.Subscription{
				ID: "sub_org_remain",
			},
		}

		handlePaymentFailed(invoice)

		// Verify org plan is still pro (grace period)
		var updatedOrg models.Organization
		database.DB.First(&updatedOrg, org.ID)
		if updatedOrg.Plan != models.OrgPlanPro {
			t.Errorf("Expected org plan '%s' during grace period, got: %s", models.OrgPlanPro, updatedOrg.Plan)
		}
	})
}

func TestFindCustomerOwner_OrgPrecedence(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("org billing takes precedence over user billing", func(t *testing.T) {
		// Create a user with a stripe customer ID
		user := createTestUserForWebhook(t, "precedence@example.com", "cus_shared_123")

		// Create an org with the same stripe customer ID
		org := &models.Organization{
			Name:             "Precedence Org",
			Slug:             "precedence-org",
			Plan:             models.OrgPlanFree,
			StripeCustomerID: func() *string { s := "cus_shared_123"; return &s }(),
			CreatedByUserID:  user.ID,
		}
		if err := database.DB.Create(org).Error; err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}

		owner, err := findCustomerOwner("cus_shared_123")
		if err != nil {
			t.Fatalf("Expected to find owner: %v", err)
		}

		// Should return org, not user
		if owner.Org == nil {
			t.Error("Expected org to be returned (org takes precedence)")
		}
		if owner.User != nil {
			t.Error("Expected user to be nil when org takes precedence")
		}
		if owner.Org.ID != org.ID {
			t.Errorf("Expected org ID %d, got: %d", org.ID, owner.Org.ID)
		}
	})

	t.Run("returns correct owner type for user-only customer", func(t *testing.T) {
		user := createTestUserForWebhook(t, "useronly@example.com", "cus_user_only_123")

		owner, err := findCustomerOwner("cus_user_only_123")
		if err != nil {
			t.Fatalf("Expected to find owner: %v", err)
		}

		if owner.User == nil {
			t.Error("Expected user to be returned")
		}
		if owner.Org != nil {
			t.Error("Expected org to be nil")
		}
		if owner.User.ID != user.ID {
			t.Errorf("Expected user ID %d, got: %d", user.ID, owner.User.ID)
		}
	})

	t.Run("returns error for unknown customer", func(t *testing.T) {
		_, err := findCustomerOwner("cus_unknown_xyz")
		if err == nil {
			t.Error("Expected error for unknown customer")
		}
	})
}
