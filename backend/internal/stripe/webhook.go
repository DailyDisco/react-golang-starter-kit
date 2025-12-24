package stripe

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
)

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

	// Find user by Stripe customer ID
	var user models.User
	if err := database.DB.Where("stripe_customer_id = ?", sub.Customer.ID).First(&user).Error; err != nil {
		log.Error().Err(err).Str("customer_id", sub.Customer.ID).Msg("user not found for customer")
		return
	}

	// Get the price ID from the first item
	var priceID string
	if len(sub.Items.Data) > 0 {
		priceID = sub.Items.Data[0].Price.ID
	}

	// Create subscription record
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
		user.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := database.DB.Save(&user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update user role")
		}
	}

	log.Info().Uint("user_id", user.ID).Msg("subscription created for user")
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

	// Update subscription record
	subscription.Status = string(sub.Status)
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

	// Sync user role based on subscription status
	syncUserRole(subscription.UserID, sub.Status)

	log.Info().Uint("user_id", subscription.UserID).Msg("subscription updated for user")
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

	// Downgrade user role
	syncUserRole(subscription.UserID, stripe.SubscriptionStatusCanceled)

	log.Info().Uint("user_id", subscription.UserID).Msg("subscription deleted for user")
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
		user.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := database.DB.Save(&user).Error; err != nil {
			log.Error().Err(err).Uint("user_id", userID).Msg("failed to sync user role")
			return
		}
		log.Info().Uint("user_id", userID).Str("new_role", newRole).Msg("user role synced")
	}
}
