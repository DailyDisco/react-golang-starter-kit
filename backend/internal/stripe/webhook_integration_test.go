package stripe

import (
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"

	stripe "github.com/stripe/stripe-go/v76"
)

func testWebhookSetup(t *testing.T) func() {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB
	oldDB := database.DB
	database.DB = tt.DB

	return func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func createTestUserForWebhook(t *testing.T, email, stripeCustomerID string) *models.User {
	t.Helper()
	user := &models.User{
		Email:            email,
		Name:             "Webhook Test User",
		Password:         "hashedpassword",
		Role:             models.RoleUser,
		StripeCustomerID: &stripeCustomerID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestSubscription(t *testing.T, userID uint, stripeSubID string, status string) *models.Subscription {
	t.Helper()
	sub := &models.Subscription{
		UserID:               userID,
		StripeSubscriptionID: stripeSubID,
		StripePriceID:        "price_test_123",
		Status:               status,
		CurrentPeriodStart:   time.Now().Format(time.RFC3339),
		CurrentPeriodEnd:     time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		CreatedAt:            time.Now().Format(time.RFC3339),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}
	if err := database.DB.Create(sub).Error; err != nil {
		t.Fatalf("Failed to create test subscription: %v", err)
	}
	return sub
}

func TestHandleSubscriptionCreated_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("creates subscription record and upgrades user to premium", func(t *testing.T) {
		user := createTestUserForWebhook(t, "newsub@example.com", "cus_test_123")

		// Create a mock Stripe subscription
		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_test_new",
			Customer: &stripe.Customer{
				ID: "cus_test_123",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000, // 30 days
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID: "price_test_abc",
						},
					},
				},
			},
		}

		handleSubscriptionCreated(sub)

		// Verify subscription was created
		var dbSub models.Subscription
		err := database.DB.Where("stripe_subscription_id = ?", "sub_test_new").First(&dbSub).Error
		if err != nil {
			t.Fatalf("Expected subscription to be created: %v", err)
		}

		if dbSub.UserID != user.ID {
			t.Errorf("Expected UserID %d, got: %d", user.ID, dbSub.UserID)
		}
		if dbSub.Status != "active" {
			t.Errorf("Expected status 'active', got: %s", dbSub.Status)
		}
		if dbSub.StripePriceID != "price_test_abc" {
			t.Errorf("Expected price ID 'price_test_abc', got: %s", dbSub.StripePriceID)
		}

		// Verify user was upgraded to premium
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected user role '%s', got: %s", models.RolePremium, updatedUser.Role)
		}
	})

	t.Run("handles orphaned customer gracefully", func(t *testing.T) {
		// Create subscription with non-existent customer
		sub := &stripe.Subscription{
			ID: "sub_orphan",
			Customer: &stripe.Customer{
				ID: "cus_nonexistent",
			},
			Status:             stripe.SubscriptionStatusActive,
			CurrentPeriodStart: time.Now().Unix(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		// Should not panic
		handleSubscriptionCreated(sub)

		// Verify no subscription was created
		var count int64
		database.DB.Model(&models.Subscription{}).Where("stripe_subscription_id = ?", "sub_orphan").Count(&count)
		if count != 0 {
			t.Error("Expected no subscription to be created for orphaned customer")
		}
	})
}

func TestHandleSubscriptionUpdated_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("updates subscription status", func(t *testing.T) {
		user := createTestUserForWebhook(t, "update@example.com", "cus_update_123")
		createTestSubscription(t, user.ID, "sub_update_test", models.SubscriptionStatusActive)

		// Simulate status change to past_due
		now := time.Now().Unix()
		sub := &stripe.Subscription{
			ID: "sub_update_test",
			Customer: &stripe.Customer{
				ID: "cus_update_123",
			},
			Status:             stripe.SubscriptionStatusPastDue,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now + 2592000,
		}

		handleSubscriptionUpdated(sub)

		// Verify subscription was updated
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_update_test").First(&dbSub)
		if dbSub.Status != "past_due" {
			t.Errorf("Expected status 'past_due', got: %s", dbSub.Status)
		}
	})

	t.Run("syncs user role on status change", func(t *testing.T) {
		user := createTestUserForWebhook(t, "sync@example.com", "cus_sync_123")
		user.Role = models.RolePremium
		database.DB.Save(user)

		createTestSubscription(t, user.ID, "sub_sync_test", models.SubscriptionStatusActive)

		// Simulate cancellation
		sub := &stripe.Subscription{
			ID: "sub_sync_test",
			Customer: &stripe.Customer{
				ID: "cus_sync_123",
			},
			Status:             stripe.SubscriptionStatusCanceled,
			CurrentPeriodStart: time.Now().Unix(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		handleSubscriptionUpdated(sub)

		// Verify user role was downgraded
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleUser {
			t.Errorf("Expected user role '%s', got: %s", models.RoleUser, updatedUser.Role)
		}
	})
}

func TestHandleSubscriptionDeleted_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("sets subscription to canceled and downgrades user", func(t *testing.T) {
		user := createTestUserForWebhook(t, "delete@example.com", "cus_delete_123")
		user.Role = models.RolePremium
		database.DB.Save(user)

		createTestSubscription(t, user.ID, "sub_delete_test", models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: "sub_delete_test",
			Customer: &stripe.Customer{
				ID: "cus_delete_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify subscription status
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_delete_test").First(&dbSub)
		if dbSub.Status != models.SubscriptionStatusCanceled {
			t.Errorf("Expected status '%s', got: %s", models.SubscriptionStatusCanceled, dbSub.Status)
		}
		if dbSub.CanceledAt == "" {
			t.Error("Expected CanceledAt to be set")
		}

		// Verify user role was downgraded
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleUser {
			t.Errorf("Expected user role '%s', got: %s", models.RoleUser, updatedUser.Role)
		}
	})

	t.Run("protects admin role from downgrade", func(t *testing.T) {
		user := createTestUserForWebhook(t, "admin@example.com", "cus_admin_123")
		user.Role = models.RoleAdmin
		database.DB.Save(user)

		createTestSubscription(t, user.ID, "sub_admin_test", models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: "sub_admin_test",
			Customer: &stripe.Customer{
				ID: "cus_admin_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify admin role is preserved
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleAdmin {
			t.Errorf("Expected admin role to be preserved, got: %s", updatedUser.Role)
		}
	})

	t.Run("protects superadmin role from downgrade", func(t *testing.T) {
		user := createTestUserForWebhook(t, "superadmin@example.com", "cus_superadmin_123")
		user.Role = models.RoleSuperAdmin
		database.DB.Save(user)

		createTestSubscription(t, user.ID, "sub_superadmin_test", models.SubscriptionStatusActive)

		sub := &stripe.Subscription{
			ID: "sub_superadmin_test",
			Customer: &stripe.Customer{
				ID: "cus_superadmin_123",
			},
		}

		handleSubscriptionDeleted(sub)

		// Verify superadmin role is preserved
		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleSuperAdmin {
			t.Errorf("Expected superadmin role to be preserved, got: %s", updatedUser.Role)
		}
	})
}

func TestHandlePaymentFailed_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("sets subscription status to past_due", func(t *testing.T) {
		user := createTestUserForWebhook(t, "failed@example.com", "cus_failed_123")
		createTestSubscription(t, user.ID, "sub_failed_test", models.SubscriptionStatusActive)

		invoice := &stripe.Invoice{
			ID: "inv_test_123",
			Customer: &stripe.Customer{
				ID: "cus_failed_123",
			},
			Subscription: &stripe.Subscription{
				ID: "sub_failed_test",
			},
		}

		handlePaymentFailed(invoice)

		// Verify subscription status
		var dbSub models.Subscription
		database.DB.Where("stripe_subscription_id = ?", "sub_failed_test").First(&dbSub)
		if dbSub.Status != models.SubscriptionStatusPastDue {
			t.Errorf("Expected status '%s', got: %s", models.SubscriptionStatusPastDue, dbSub.Status)
		}
	})

	t.Run("handles invoice without subscription", func(t *testing.T) {
		invoice := &stripe.Invoice{
			ID: "inv_no_sub",
			Customer: &stripe.Customer{
				ID: "cus_no_sub",
			},
			// No Subscription field
		}

		// Should not panic
		handlePaymentFailed(invoice)
	})
}

func TestCheckoutSessionCompleted_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("logs successful checkout for known user", func(t *testing.T) {
		user := createTestUserForWebhook(t, "checkout@example.com", "cus_checkout_123")

		session := &stripe.CheckoutSession{
			ID: "cs_test_123",
			Customer: &stripe.Customer{
				ID: "cus_checkout_123",
			},
			Subscription: &stripe.Subscription{
				ID: "sub_checkout_123",
			},
		}

		// Should complete without error
		handleCheckoutSessionCompleted(session)

		// User should still exist (function just logs)
		var updatedUser models.User
		err := database.DB.First(&updatedUser, user.ID).Error
		if err != nil {
			t.Errorf("User should still exist: %v", err)
		}
	})
}

func TestSyncUserRole_Integration(t *testing.T) {
	cleanup := testWebhookSetup(t)
	defer cleanup()

	t.Run("upgrades to premium for active subscription", func(t *testing.T) {
		user := createTestUserForWebhook(t, "active@example.com", "cus_active")

		syncUserRole(user.ID, stripe.SubscriptionStatusActive)

		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected premium role, got: %s", updatedUser.Role)
		}
	})

	t.Run("upgrades to premium for trialing subscription", func(t *testing.T) {
		user := createTestUserForWebhook(t, "trialing@example.com", "cus_trialing")

		syncUserRole(user.ID, stripe.SubscriptionStatusTrialing)

		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected premium role, got: %s", updatedUser.Role)
		}
	})

	t.Run("keeps premium for past_due (grace period)", func(t *testing.T) {
		user := createTestUserForWebhook(t, "pastdue@example.com", "cus_pastdue")
		user.Role = models.RolePremium
		database.DB.Save(user)

		syncUserRole(user.ID, stripe.SubscriptionStatusPastDue)

		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RolePremium {
			t.Errorf("Expected premium role during grace period, got: %s", updatedUser.Role)
		}
	})

	t.Run("downgrades to user for canceled", func(t *testing.T) {
		user := createTestUserForWebhook(t, "canceled@example.com", "cus_canceled")
		user.Role = models.RolePremium
		database.DB.Save(user)

		syncUserRole(user.ID, stripe.SubscriptionStatusCanceled)

		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleUser {
			t.Errorf("Expected user role after cancellation, got: %s", updatedUser.Role)
		}
	})

	t.Run("does not downgrade admin role", func(t *testing.T) {
		user := createTestUserForWebhook(t, "adminsync@example.com", "cus_adminsync")
		user.Role = models.RoleAdmin
		database.DB.Save(user)

		syncUserRole(user.ID, stripe.SubscriptionStatusCanceled)

		var updatedUser models.User
		database.DB.First(&updatedUser, user.ID)
		if updatedUser.Role != models.RoleAdmin {
			t.Errorf("Expected admin role to be preserved, got: %s", updatedUser.Role)
		}
	})
}
