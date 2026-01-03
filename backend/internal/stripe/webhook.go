package stripe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"

	"gorm.io/gorm"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/services"
)

// customerOwner represents either a user or organization that owns a Stripe customer
type customerOwner struct {
	User *models.User
	Org  *models.Organization
}

// findCustomerOwner finds the user or organization that owns a Stripe customer ID
func findCustomerOwner(customerID string) (*customerOwner, error) {
	// Check for organization first (org billing takes precedence)
	var org models.Organization
	if err := database.DB.Where("stripe_customer_id = ?", customerID).First(&org).Error; err == nil {
		return &customerOwner{Org: &org}, nil
	}

	// Check for user
	var user models.User
	if err := database.DB.Where("stripe_customer_id = ?", customerID).First(&user).Error; err == nil {
		return &customerOwner{User: &user}, nil
	}

	return nil, gorm.ErrRecordNotFound
}

// HandleWebhook handles incoming Stripe webhook events
// @Summary Handle Stripe webhook
// @Description Receives and processes Stripe webhook events
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/webhooks/stripe [post]
func HandleWebhook(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("failed to read webhook body")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Failed to read request body",
				Code:    http.StatusBadRequest,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Verify webhook signature
		sigHeader := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sigHeader, config.WebhookSecret)
		if err != nil {
			log.Error().Err(err).Msg("webhook signature verification failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid webhook signature",
				Code:    http.StatusBadRequest,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		log.Info().Str("event_type", string(event.Type)).Str("event_id", event.ID).Msg("received stripe webhook")

		// Handle the event
		switch event.Type {
		case "checkout.session.completed":
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal checkout session")
				break
			}
			handleCheckoutSessionCompleted(&session)

		case "customer.subscription.created":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionCreated(&sub)

		case "customer.subscription.updated":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionUpdated(&sub)

		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionDeleted(&sub)

		case "invoice.payment_failed":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal invoice")
				break
			}
			handlePaymentFailed(&invoice)

		default:
			log.Debug().Str("event_type", string(event.Type)).Msg("unhandled webhook event type")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(models.SuccessResponse{
			Success: true,
			Message: "Webhook processed",
		}); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// handleCheckoutSessionCompleted processes successful checkout sessions
func handleCheckoutSessionCompleted(session *stripe.CheckoutSession) {
	log.Info().
		Str("session_id", session.ID).
		Str("customer_id", session.Customer.ID).
		Str("subscription_id", session.Subscription.ID).
		Msg("checkout session completed")

	// Find user by Stripe customer ID
	var user models.User
	if err := database.DB.Where("stripe_customer_id = ?", session.Customer.ID).First(&user).Error; err != nil {
		log.Error().Err(err).Str("customer_id", session.Customer.ID).Msg("user not found for customer")
		return
	}

	// The subscription will be created/updated via the subscription webhook events
	// Just log success here
	log.Info().Uint("user_id", user.ID).Msg("checkout completed for user")
}

// handleSubscriptionCreated processes new subscription creation
func handleSubscriptionCreated(sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Str("customer_id", sub.Customer.ID).
		Str("status", string(sub.Status)).
		Msg("subscription created")

	// Find owner (user or org) by Stripe customer ID
	owner, err := findCustomerOwner(sub.Customer.ID)
	if err != nil {
		log.Error().Err(err).Str("customer_id", sub.Customer.ID).Msg("customer owner not found")
		return
	}

	// Get the price ID from the first item
	var priceID string
	if len(sub.Items.Data) > 0 {
		priceID = sub.Items.Data[0].Price.ID
	}

	if owner.Org != nil {
		// Organization subscription
		handleOrgSubscriptionCreated(owner.Org, sub, priceID)
	} else {
		// User subscription
		handleUserSubscriptionCreated(owner.User, sub, priceID)
	}
}

// handleUserSubscriptionCreated processes user-level subscription creation
func handleUserSubscriptionCreated(user *models.User, sub *stripe.Subscription, priceID string) {
	subscription := models.Subscription{
		UserID:               user.ID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        priceID,
		Status:               string(sub.Status),
		CurrentPeriodStart:   time.Unix(sub.CurrentPeriodStart, 0).Format(time.RFC3339),
		CurrentPeriodEnd:     time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
		CreatedAt:            time.Now().Format(time.RFC3339),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to create subscription record")
		return
	}

	// Update user role to premium if subscription is active
	if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing {
		user.Role = models.RolePremium
		user.UpdatedAt = time.Now()
		if err := database.DB.Save(user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update user role")
		}
	}

	syncUsageLimits(user.ID, priceID)
	log.Info().Uint("user_id", user.ID).Msg("subscription created for user")
}

// handleOrgSubscriptionCreated processes organization-level subscription creation
func handleOrgSubscriptionCreated(org *models.Organization, sub *stripe.Subscription, priceID string) {
	// Get the org owner to set as billing contact
	var owner models.OrganizationMember
	if err := database.DB.Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner).Error; err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("org owner not found")
		return
	}

	orgID := org.ID
	subscription := models.Subscription{
		UserID:               owner.UserID,
		OrganizationID:       &orgID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        priceID,
		Status:               string(sub.Status),
		CurrentPeriodStart:   time.Unix(sub.CurrentPeriodStart, 0).Format(time.RFC3339),
		CurrentPeriodEnd:     time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
		CreatedAt:            time.Now().Format(time.RFC3339),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to create org subscription record")
		return
	}

	// Update org plan based on price ID
	newPlan := getPlanFromPriceID(priceID)
	if err := database.DB.Model(org).Updates(map[string]interface{}{
		"plan":                   newPlan,
		"stripe_subscription_id": sub.ID,
	}).Error; err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to update org plan")
	}

	log.Info().Uint("org_id", org.ID).Str("plan", string(newPlan)).Msg("subscription created for organization")
}

// getPlanFromPriceID maps Stripe price IDs to organization plans
// Uses environment-configured price IDs to determine tier:
// - STRIPE_ENTERPRISE_PRICE_ID -> Enterprise
// - STRIPE_PREMIUM_PRICE_ID -> Pro
// - Any other non-empty price ID -> Pro (fallback for paid plans)
// - Empty price ID -> Free
func getPlanFromPriceID(priceID string) models.OrganizationPlan {
	if priceID == "" {
		return models.OrgPlanFree
	}

	config := LoadConfig()

	// Check for enterprise tier first
	if config.EnterprisePriceID != "" && priceID == config.EnterprisePriceID {
		return models.OrgPlanEnterprise
	}

	// Check for pro tier (premium price ID)
	if config.PremiumPriceID != "" && priceID == config.PremiumPriceID {
		return models.OrgPlanPro
	}

	// Default: any paid subscription without specific mapping is Pro
	return models.OrgPlanPro
}

// handleSubscriptionUpdated processes subscription updates
func handleSubscriptionUpdated(sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Str("status", string(sub.Status)).
		Msg("subscription updated")

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.Where("stripe_subscription_id = ?", sub.ID).First(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("subscription not found")
		return
	}

	// Get the price ID from the first item (may have changed on plan upgrade/downgrade)
	var priceID string
	if len(sub.Items.Data) > 0 {
		priceID = sub.Items.Data[0].Price.ID
	}

	// Update subscription record
	subscription.Status = string(sub.Status)
	subscription.StripePriceID = priceID
	subscription.CurrentPeriodStart = time.Unix(sub.CurrentPeriodStart, 0).Format(time.RFC3339)
	subscription.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339)
	subscription.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
	subscription.UpdatedAt = time.Now().Format(time.RFC3339)

	if sub.CanceledAt > 0 {
		subscription.CanceledAt = time.Unix(sub.CanceledAt, 0).Format(time.RFC3339)
	}

	if err := database.DB.Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to update subscription")
		return
	}

	// Handle org vs user subscription updates
	if subscription.OrganizationID != nil && *subscription.OrganizationID > 0 {
		// Organization subscription - update org plan
		newPlan := getPlanFromPriceID(priceID)
		if err := database.DB.Model(&models.Organization{}).Where("id = ?", *subscription.OrganizationID).
			Update("plan", newPlan).Error; err != nil {
			log.Error().Err(err).Uint("org_id", *subscription.OrganizationID).Msg("failed to update org plan")
		}
		log.Info().Uint("org_id", *subscription.OrganizationID).Str("plan", string(newPlan)).Msg("subscription updated for organization")
	} else {
		// User subscription
		syncUserRole(subscription.UserID, sub.Status)
		syncUsageLimits(subscription.UserID, priceID)
		log.Info().Uint("user_id", subscription.UserID).Msg("subscription updated for user")
	}
}

// handleSubscriptionDeleted processes subscription cancellation/deletion
func handleSubscriptionDeleted(sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Msg("subscription deleted")

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.Where("stripe_subscription_id = ?", sub.ID).First(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("subscription not found")
		return
	}

	// Update subscription record
	subscription.Status = models.SubscriptionStatusCanceled
	subscription.CanceledAt = time.Now().Format(time.RFC3339)
	subscription.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to update subscription")
		return
	}

	// Handle org vs user subscription deletion
	if subscription.OrganizationID != nil && *subscription.OrganizationID > 0 {
		// Organization subscription - downgrade to free plan
		if err := database.DB.Model(&models.Organization{}).Where("id = ?", *subscription.OrganizationID).
			Updates(map[string]any{
				"plan":                   models.OrgPlanFree,
				"stripe_subscription_id": nil,
			}).Error; err != nil {
			log.Error().Err(err).Uint("org_id", *subscription.OrganizationID).Msg("failed to downgrade org plan")
		}
		log.Info().Uint("org_id", *subscription.OrganizationID).Msg("subscription deleted for organization")
	} else {
		// User subscription
		syncUserRole(subscription.UserID, stripe.SubscriptionStatusCanceled)
		syncUsageLimits(subscription.UserID, "")
		log.Info().Uint("user_id", subscription.UserID).Msg("subscription deleted for user")
	}
}

// handlePaymentFailed processes failed payment events
func handlePaymentFailed(invoice *stripe.Invoice) {
	log.Warn().
		Str("invoice_id", invoice.ID).
		Str("customer_id", invoice.Customer.ID).
		Msg("payment failed")

	if invoice.Subscription == nil {
		return
	}

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", invoice.Subscription.ID).Msg("subscription not found")
		return
	}

	// Update subscription status to past_due
	subscription.Status = models.SubscriptionStatusPastDue
	subscription.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", invoice.Subscription.ID).Msg("failed to update subscription")
		return
	}

	log.Warn().Uint("user_id", subscription.UserID).Msg("subscription marked as past_due")
}

// syncUserRole updates the user's role based on their subscription status
func syncUserRole(userID uint, status stripe.SubscriptionStatus) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("user not found for role sync")
		return
	}

	// Don't downgrade admin roles
	if user.Role == models.RoleSuperAdmin || user.Role == models.RoleAdmin {
		return
	}

	var newRole string
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		newRole = models.RolePremium
	case stripe.SubscriptionStatusPastDue:
		// Keep premium during grace period
		newRole = models.RolePremium
	default:
		// Canceled, unpaid, etc. - downgrade to user
		newRole = models.RoleUser
	}

	if user.Role != newRole {
		user.Role = newRole
		user.UpdatedAt = time.Now()
		if err := database.DB.Save(&user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", userID).Msg("failed to sync user role")
			return
		}
		log.Info().Uint("user_id", userID).Str("new_role", newRole).Msg("user role synced")
	}
}

// syncUsageLimits updates the user's usage limits based on their subscription tier
func syncUsageLimits(userID uint, priceID string) {
	ctx := context.Background()
	usageService := services.NewUsageService(database.DB)
	if err := usageService.UpdateUserLimits(ctx, userID, priceID); err != nil {
		log.Error().Err(err).Uint("user_id", userID).Str("price_id", priceID).Msg("failed to sync usage limits")
		return
	}
	log.Info().Uint("user_id", userID).Str("price_id", priceID).Msg("usage limits synced")
}
