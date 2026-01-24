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
	ws "react-golang-starter/internal/websocket"
)

// wsHub holds the WebSocket hub for broadcasting subscription events
var wsHub *ws.Hub

// SetHub sets the WebSocket hub for broadcasting subscription events
func SetHub(hub *ws.Hub) {
	wsHub = hub
}

// broadcastSubscriptionEvent sends a subscription update to the user via WebSocket
func broadcastSubscriptionEvent(userID uint, event, status, plan, priceID string, cancelAtPeriodEnd bool, periodEnd, message string) {
	if wsHub == nil {
		return
	}

	payload := ws.SubscriptionUpdatePayload{
		Event:             event,
		Status:            status,
		Plan:              plan,
		PriceID:           priceID,
		CancelAtPeriodEnd: cancelAtPeriodEnd,
		CurrentPeriodEnd:  periodEnd,
		Message:           message,
		Timestamp:         time.Now().Unix(),
	}

	wsHub.SendToUser(userID, ws.MessageTypeSubscriptionUpdate, payload)
	log.Debug().
		Uint("user_id", userID).
		Str("event", event).
		Str("status", status).
		Msg("subscription update broadcast sent")
}

// broadcastOrgSubscriptionEvent sends a subscription update to all org members via WebSocket
func broadcastOrgSubscriptionEvent(ctx context.Context, orgID uint, event, status, plan, priceID string, cancelAtPeriodEnd bool, periodEnd, message string) {
	if wsHub == nil {
		return
	}

	// Get all org members to notify
	var members []models.OrganizationMember
	if err := database.DB.WithContext(ctx).Where("organization_id = ?", orgID).Find(&members).Error; err != nil {
		log.Error().Err(err).Uint("org_id", orgID).Msg("failed to get org members for broadcast")
		return
	}

	payload := ws.SubscriptionUpdatePayload{
		Event:             event,
		Status:            status,
		Plan:              plan,
		PriceID:           priceID,
		CancelAtPeriodEnd: cancelAtPeriodEnd,
		CurrentPeriodEnd:  periodEnd,
		Message:           message,
		Timestamp:         time.Now().Unix(),
	}

	for _, member := range members {
		wsHub.SendToUser(member.UserID, ws.MessageTypeSubscriptionUpdate, payload)
	}

	log.Debug().
		Uint("org_id", orgID).
		Int("member_count", len(members)).
		Str("event", event).
		Msg("org subscription update broadcast sent")
}

// customerOwner represents either a user or organization that owns a Stripe customer
type customerOwner struct {
	User *models.User
	Org  *models.Organization
}

// findCustomerOwner finds the user or organization that owns a Stripe customer ID
func findCustomerOwner(ctx context.Context, customerID string) (*customerOwner, error) {
	// Check for organization first (org billing takes precedence)
	var org models.Organization
	if err := database.DB.WithContext(ctx).Where("stripe_customer_id = ?", customerID).First(&org).Error; err == nil {
		return &customerOwner{Org: &org}, nil
	}

	// Check for user
	var user models.User
	if err := database.DB.WithContext(ctx).Where("stripe_customer_id = ?", customerID).First(&user).Error; err == nil {
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
			handleCheckoutSessionCompleted(r.Context(), &session)

		case "customer.subscription.created":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionCreated(r.Context(), &sub)

		case "customer.subscription.updated":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionUpdated(r.Context(), &sub)

		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal subscription")
				break
			}
			handleSubscriptionDeleted(r.Context(), &sub)

		case "invoice.payment_failed":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal invoice")
				break
			}
			handlePaymentFailed(r.Context(), &invoice)

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
func handleCheckoutSessionCompleted(ctx context.Context, session *stripe.CheckoutSession) {
	log.Info().
		Str("session_id", session.ID).
		Str("customer_id", session.Customer.ID).
		Str("subscription_id", session.Subscription.ID).
		Msg("checkout session completed")

	// Find user by Stripe customer ID
	var user models.User
	if err := database.DB.WithContext(ctx).Where("stripe_customer_id = ?", session.Customer.ID).First(&user).Error; err != nil {
		log.Error().Err(err).Str("customer_id", session.Customer.ID).Msg("user not found for customer")
		return
	}

	// The subscription will be created/updated via the subscription webhook events
	// Just log success here
	log.Info().Uint("user_id", user.ID).Msg("checkout completed for user")
}

// handleSubscriptionCreated processes new subscription creation
func handleSubscriptionCreated(ctx context.Context, sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Str("customer_id", sub.Customer.ID).
		Str("status", string(sub.Status)).
		Msg("subscription created")

	// Find owner (user or org) by Stripe customer ID
	owner, err := findCustomerOwner(ctx, sub.Customer.ID)
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
		handleOrgSubscriptionCreated(ctx, owner.Org, sub, priceID)
	} else {
		// User subscription
		handleUserSubscriptionCreated(ctx, owner.User, sub, priceID)
	}
}

// handleUserSubscriptionCreated processes user-level subscription creation
func handleUserSubscriptionCreated(ctx context.Context, user *models.User, sub *stripe.Subscription, priceID string) {
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

	if err := database.DB.WithContext(ctx).Create(&subscription).Error; err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to create subscription record")
		return
	}

	// Update user role to premium if subscription is active
	if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing {
		user.Role = models.RolePremium
		user.UpdatedAt = time.Now()
		if err := database.DB.WithContext(ctx).Save(user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update user role")
		}
	}

	syncUsageLimits(ctx, user.ID, priceID)
	log.Info().Uint("user_id", user.ID).Msg("subscription created for user")

	// Broadcast subscription created event
	broadcastSubscriptionEvent(
		user.ID,
		"created",
		string(sub.Status),
		"premium",
		priceID,
		sub.CancelAtPeriodEnd,
		time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		"Your subscription is now active!",
	)
}

// handleOrgSubscriptionCreated processes organization-level subscription creation
func handleOrgSubscriptionCreated(ctx context.Context, org *models.Organization, sub *stripe.Subscription, priceID string) {
	// Get the org owner to set as billing contact
	var owner models.OrganizationMember
	if err := database.DB.WithContext(ctx).Where("organization_id = ? AND role = ?", org.ID, models.OrgRoleOwner).First(&owner).Error; err != nil {
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

	if err := database.DB.WithContext(ctx).Create(&subscription).Error; err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to create org subscription record")
		return
	}

	// Update org plan based on price ID
	newPlan := getPlanFromPriceID(priceID)
	if err := database.DB.WithContext(ctx).Model(org).Updates(map[string]interface{}{
		"plan":                   newPlan,
		"stripe_subscription_id": sub.ID,
	}).Error; err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to update org plan")
	}

	log.Info().Uint("org_id", org.ID).Str("plan", string(newPlan)).Msg("subscription created for organization")

	// Broadcast subscription created event to all org members
	broadcastOrgSubscriptionEvent(
		ctx,
		org.ID,
		"created",
		string(sub.Status),
		string(newPlan),
		priceID,
		sub.CancelAtPeriodEnd,
		time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		"Organization subscription is now active!",
	)
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
func handleSubscriptionUpdated(ctx context.Context, sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Str("status", string(sub.Status)).
		Msg("subscription updated")

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.WithContext(ctx).Where("stripe_subscription_id = ?", sub.ID).First(&subscription).Error; err != nil {
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

	if err := database.DB.WithContext(ctx).Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to update subscription")
		return
	}

	// Build status message
	var message string
	if sub.CancelAtPeriodEnd {
		message = "Subscription will be canceled at the end of the billing period"
	} else if sub.Status == stripe.SubscriptionStatusPastDue {
		message = "Payment failed - please update your payment method"
	} else {
		message = "Subscription has been updated"
	}

	// Handle org vs user subscription updates
	if subscription.OrganizationID != nil && *subscription.OrganizationID > 0 {
		// Organization subscription - update org plan
		newPlan := getPlanFromPriceID(priceID)
		if err := database.DB.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", *subscription.OrganizationID).
			Update("plan", newPlan).Error; err != nil {
			log.Error().Err(err).Uint("org_id", *subscription.OrganizationID).Msg("failed to update org plan")
		}
		log.Info().Uint("org_id", *subscription.OrganizationID).Str("plan", string(newPlan)).Msg("subscription updated for organization")

		// Broadcast to org members
		broadcastOrgSubscriptionEvent(
			ctx,
			*subscription.OrganizationID,
			"updated",
			string(sub.Status),
			string(newPlan),
			priceID,
			sub.CancelAtPeriodEnd,
			time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
			message,
		)
	} else {
		// User subscription
		syncUserRole(ctx, subscription.UserID, sub.Status)
		syncUsageLimits(ctx, subscription.UserID, priceID)
		log.Info().Uint("user_id", subscription.UserID).Msg("subscription updated for user")

		// Broadcast to user
		broadcastSubscriptionEvent(
			subscription.UserID,
			"updated",
			string(sub.Status),
			"premium",
			priceID,
			sub.CancelAtPeriodEnd,
			time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
			message,
		)
	}
}

// handleSubscriptionDeleted processes subscription cancellation/deletion
func handleSubscriptionDeleted(ctx context.Context, sub *stripe.Subscription) {
	log.Info().
		Str("subscription_id", sub.ID).
		Msg("subscription deleted")

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.WithContext(ctx).Where("stripe_subscription_id = ?", sub.ID).First(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("subscription not found")
		return
	}

	// Update subscription record
	subscription.Status = models.SubscriptionStatusCanceled
	subscription.CanceledAt = time.Now().Format(time.RFC3339)
	subscription.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.WithContext(ctx).Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to update subscription")
		return
	}

	// Handle org vs user subscription deletion
	if subscription.OrganizationID != nil && *subscription.OrganizationID > 0 {
		// Organization subscription - downgrade to free plan
		if err := database.DB.WithContext(ctx).Model(&models.Organization{}).Where("id = ?", *subscription.OrganizationID).
			Updates(map[string]any{
				"plan":                   models.OrgPlanFree,
				"stripe_subscription_id": nil,
			}).Error; err != nil {
			log.Error().Err(err).Uint("org_id", *subscription.OrganizationID).Msg("failed to downgrade org plan")
		}
		log.Info().Uint("org_id", *subscription.OrganizationID).Msg("subscription deleted for organization")

		// Broadcast to org members
		broadcastOrgSubscriptionEvent(
			ctx,
			*subscription.OrganizationID,
			"deleted",
			models.SubscriptionStatusCanceled,
			string(models.OrgPlanFree),
			"",
			false,
			"",
			"Subscription has been canceled. You are now on the free plan.",
		)
	} else {
		// User subscription
		syncUserRole(ctx, subscription.UserID, stripe.SubscriptionStatusCanceled)
		syncUsageLimits(ctx, subscription.UserID, "")
		log.Info().Uint("user_id", subscription.UserID).Msg("subscription deleted for user")

		// Broadcast to user
		broadcastSubscriptionEvent(
			subscription.UserID,
			"deleted",
			models.SubscriptionStatusCanceled,
			"free",
			"",
			false,
			"",
			"Your subscription has been canceled.",
		)
	}
}

// handlePaymentFailed processes failed payment events
func handlePaymentFailed(ctx context.Context, invoice *stripe.Invoice) {
	log.Warn().
		Str("invoice_id", invoice.ID).
		Str("customer_id", invoice.Customer.ID).
		Msg("payment failed")

	if invoice.Subscription == nil {
		return
	}

	// Find subscription by Stripe subscription ID
	var subscription models.Subscription
	if err := database.DB.WithContext(ctx).Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", invoice.Subscription.ID).Msg("subscription not found")
		return
	}

	// Update subscription status to past_due
	subscription.Status = models.SubscriptionStatusPastDue
	subscription.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.WithContext(ctx).Save(&subscription).Error; err != nil {
		log.Error().Err(err).Str("subscription_id", invoice.Subscription.ID).Msg("failed to update subscription")
		return
	}

	log.Warn().Uint("user_id", subscription.UserID).Msg("subscription marked as past_due")

	// Broadcast payment failed event
	if subscription.OrganizationID != nil && *subscription.OrganizationID > 0 {
		broadcastOrgSubscriptionEvent(
			ctx,
			*subscription.OrganizationID,
			"payment_failed",
			models.SubscriptionStatusPastDue,
			"",
			subscription.StripePriceID,
			false,
			"",
			"Payment failed. Please update your payment method to avoid service interruption.",
		)
	} else {
		broadcastSubscriptionEvent(
			subscription.UserID,
			"payment_failed",
			models.SubscriptionStatusPastDue,
			"premium",
			subscription.StripePriceID,
			false,
			"",
			"Payment failed. Please update your payment method to avoid service interruption.",
		)
	}
}

// syncUserRole updates the user's role based on their subscription status
func syncUserRole(ctx context.Context, userID uint, status stripe.SubscriptionStatus) {
	var user models.User
	if err := database.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
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
		if err := database.DB.WithContext(ctx).Save(&user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", userID).Msg("failed to sync user role")
			return
		}
		log.Info().Uint("user_id", userID).Str("new_role", newRole).Msg("user role synced")
	}
}

// syncUsageLimits updates the user's usage limits based on their subscription tier
func syncUsageLimits(ctx context.Context, userID uint, priceID string) {
	usageService := services.NewUsageService(database.DB)
	if err := usageService.UpdateUserLimits(ctx, userID, priceID); err != nil {
		log.Error().Err(err).Uint("user_id", userID).Str("price_id", priceID).Msg("failed to sync usage limits")
		return
	}
	log.Info().Uint("user_id", userID).Str("price_id", priceID).Msg("usage limits synced")
}
