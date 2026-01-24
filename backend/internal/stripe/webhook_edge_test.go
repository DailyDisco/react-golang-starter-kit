package stripe

import (
	"context"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	stripe "github.com/stripe/stripe-go/v76"
)

// Edge Case Tests for Stripe Webhooks

func TestWebhookEdgeCases_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("subscription with no items (empty price ID)", func(t *testing.T) {
		user := createTestUserForWebhook(t, "noitems@example.com", "cus_noitems_123")

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_noitems",
			Customer: &stripe.Customer{
				ID: "cus_noitems_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{}, // Empty items
			},
		}

		// Should not panic
		handleSubscriptionCreated(context.Background(), sub)

		// Verify subscription was created with empty price ID
		var dbSub models.Subscription
		err := database.DB.Where("stripe_subscription_id = ?", "sub_noitems").First(&dbSub).Error
		if err != nil {
			t.Fatalf("Expected subscription to be created: %v", err)
		}

		if dbSub.StripePriceID != "" {
			t.Errorf("Expected empty price ID, got: %s", dbSub.StripePriceID)
		}

		// User should still be upgraded for active subscription
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected premium role, got: %s", updatedUser.Role)
		}
	})

	t.Run("org with no owner - error logged, not panicked", func(t *testing.T) {
		// Create org without any members
		user := createTestUserForWebhook(t, "noowner-creator@example.com", "cus_creator_noowner")
		org := &models.Organization{
			Name:             "No Owner Org",
			Slug:             "no-owner-org",
			Plan:             models.OrgPlanFree,
			StripeCustomerID: func() *string { s := "cus_org_noowner_123"; return &s }(),
			CreatedByUserID:  user.ID,
		}
		if err := database.DB.Create(org).Error; err != nil {
			t.Fatalf("Failed to create org: %v", err)
		}
		// Note: We're NOT creating an owner membership

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_noowner",
			Customer: &stripe.Customer{
				ID: "cus_org_noowner_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_noowner",
						},
					},
				},
			},
		}

		// Should not panic, just log error
		handleSubscriptionCreated(context.Background(), sub)

		// Verify no subscription was created
		var count int64
		database.DB.Model(&models.Subscription{}).Where("stripe_subscription_id = ?", "sub_noowner").Count(&count)
		if count != 0 {
			t.Error("Expected no subscription to be created for org without owner")
		}
	})

	t.Run("unknown price ID defaults to free plan", func(t *testing.T) {
		// Test the getPlanFromPriceID function
		plan := getPlanFromPriceID("")
		if plan != models.OrgPlanFree {
			t.Errorf("Expected free plan for empty price ID, got: %s", plan)
		}

		// Any non-empty price ID should map to pro (current implementation)
		plan = getPlanFromPriceID("price_unknown_xyz")
		if plan != models.OrgPlanPro {
			t.Errorf("Expected pro plan for unknown price ID, got: %s", plan)
		}
	})

	t.Run("user deleted between webhook events", func(t *testing.T) {
		user := createTestUserForWebhook(t, "deleted@example.com", "cus_deleted_123")
		createTestSubscription(t, user.ID, "sub_deleted_user", models.SubscriptionStatusActive)

		// Delete the user
		database.DB.Delete(&user)

		// Simulate subscription update for deleted user
		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_deleted_user",
			Customer: &stripe.Customer{
				ID: "cus_deleted_123",
			},
			Status:             stripe.SubscriptionStatusCanceled,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
		}

		// Should not panic
		handleSubscriptionUpdated(context.Background(), sub)

		// Subscription status should still be updated (record exists)
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_deleted_user").First(&dbSub)
		if dbSub.Status != "canceled" {
			t.Errorf("Expected status 'canceled', got: %s", dbSub.Status)
		}
	})

	t.Run("organization deleted between webhook events", func(t *testing.T) {
		org := createTestOrganization(t, "deleted-org", "cus_deleted_org_123")

		var owner models.OrganizationMember
		database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner)

		createTestOrgSubscription(t, org.ID, owner.UserID, "sub_deleted_org", models.SubscriptionStatusActive)

		// Delete the organization (soft delete via GORM)
		database.DB.Delete(&org)

		// Simulate subscription update
		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_deleted_org",
			Customer: &stripe.Customer{
				ID: "cus_deleted_org_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
		}

		// Should not panic
		handleSubscriptionUpdated(context.Background(), sub)

		// Subscription should still be updated
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_deleted_org").First(&dbSub)
		if dbSub.Status != "active" {
			t.Errorf("Expected status 'active', got: %s", dbSub.Status)
		}
	})

	t.Run("duplicate webhook delivery (idempotency)", func(t *testing.T) {
		user := createTestUserForWebhook(t, "idempotent@example.com", "cus_idempotent_123")

		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_idempotent",
			Customer: &stripe.Customer{
				ID: "cus_idempotent_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_idempotent",
						},
					},
				},
			},
		}

		// First delivery
		handleSubscriptionCreated(context.Background(), sub)

		// Second delivery (duplicate)
		// Current implementation will try to create another record which will fail
		// or the handler should be idempotent
		handleSubscriptionCreated(context.Background(), sub)

		// Verify only one subscription exists
		var count int64
		database.DB.Model(&models.Subscription{}).Where("stripe_subscription_id = ?", "sub_idempotent").Count(&count)
		// Due to unique constraint or idempotency, should only have 1
		// Note: This depends on whether there's a unique index on stripe_subscription_id
		if count > 1 {
			t.Errorf("Expected 1 subscription record, got: %d (not idempotent)", count)
		}

		// User should still be premium
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected premium role, got: %s", updatedUser.Role)
		}
	})

	t.Run("subscription update without existing record", func(t *testing.T) {
		// This tests the case where we receive an update for a subscription
		// that we don't have in our database
		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_nonexistent_update",
			Customer: &stripe.Customer{
				ID: "cus_nonexistent",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
		}

		// Should not panic
		handleSubscriptionUpdated(context.Background(), sub)

		// Verify no subscription was created (update should fail silently)
		var count int64
		database.DB.Model(&models.Subscription{}).Where("stripe_subscription_id = ?", "sub_nonexistent_update").Count(&count)
		if count != 0 {
			t.Error("Expected no subscription to be created for update of non-existent subscription")
		}
	})

	t.Run("subscription deletion without existing record", func(t *testing.T) {
		sub := &stripe.Subscription{
			ID: "sub_nonexistent_delete",
			Customer: &stripe.Customer{
				ID: "cus_nonexistent",
			},
		}

		// Should not panic
		handleSubscriptionDeleted(context.Background(), sub)
	})

	t.Run("payment failed for non-existent subscription", func(t *testing.T) {
		invoice := &stripe.Invoice{
			ID: "inv_nonexistent",
			Customer: &stripe.Customer{
				ID: "cus_nonexistent",
			},
			Subscription: &stripe.Subscription{
				ID: "sub_nonexistent_payment",
			},
		}

		// Should not panic
		handlePaymentFailed(context.Background(), invoice)
	})

	t.Run("cancel at period end flag is preserved", func(t *testing.T) {
		user := createTestUserForWebhook(t, "cancelend@example.com", "cus_cancelend_123")
		createTestSubscription(t, user.ID, "sub_cancelend", models.SubscriptionStatusActive)

		now := time.Now().Unix()
		canceledAt := now + 86400 // Cancel scheduled for tomorrow
		sub := &stripe.Subscription{
			ID: "sub_cancelend",
			Customer: &stripe.Customer{
				ID: "cus_cancelend_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
			CancelAtPeriodEnd:  true,
			CanceledAt:         canceledAt,
		}

		handleSubscriptionUpdated(context.Background(), sub)

		// Verify CancelAtPeriodEnd was set
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_cancelend").First(&dbSub)
		if !dbSub.CancelAtPeriodEnd {
			t.Error("Expected CancelAtPeriodEnd to be true")
		}
		if dbSub.CanceledAt == "" {
			t.Error("Expected CanceledAt to be set")
		}
	})
}

func TestGetPlanFromPriceID(t *testing.T) {
	// Test without env vars set (default behavior)
	t.Run("without env vars", func(t *testing.T) {
		tests := []struct {
			name     string
			priceID  string
			expected models.OrganizationPlan
		}{
			{"empty price ID returns free", "", models.OrgPlanFree},
			{"any non-empty price ID returns pro", "price_abc123", models.OrgPlanPro},
			{"unknown enterprise price ID returns pro", "price_enterprise_xyz", models.OrgPlanPro},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getPlanFromPriceID(tt.priceID)
				if result != tt.expected {
					t.Errorf("getPlanFromPriceID(%q) = %s, expected %s", tt.priceID, result, tt.expected)
				}
			})
		}
	})

	// Test with env vars set for tier mapping
	t.Run("with env vars", func(t *testing.T) {
		// Set up test price IDs
		t.Setenv("STRIPE_PREMIUM_PRICE_ID", "price_pro_test")
		t.Setenv("STRIPE_ENTERPRISE_PRICE_ID", "price_enterprise_test")

		tests := []struct {
			name     string
			priceID  string
			expected models.OrganizationPlan
		}{
			{"empty price ID returns free", "", models.OrgPlanFree},
			{"configured pro price ID returns pro", "price_pro_test", models.OrgPlanPro},
			{"configured enterprise price ID returns enterprise", "price_enterprise_test", models.OrgPlanEnterprise},
			{"unknown price ID returns pro (paid fallback)", "price_unknown", models.OrgPlanPro},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getPlanFromPriceID(tt.priceID)
				if result != tt.expected {
					t.Errorf("getPlanFromPriceID(%q) = %s, expected %s", tt.priceID, result, tt.expected)
				}
			})
		}
	})
}
