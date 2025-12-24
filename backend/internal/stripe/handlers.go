package stripe

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
)

// GetBillingConfig returns the public Stripe configuration
// @Summary Get billing configuration
// @Description Returns the Stripe publishable key for frontend integration
// @Tags billing
// @Produce json
// @Success 200 {object} models.BillingConfigResponse
// @Router /api/billing/config [get]
func GetBillingConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := GetService()

		response := models.BillingConfigResponse{
			PublishableKey: svc.GetPublishableKey(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// GetPlans returns available subscription plans
// @Summary Get subscription plans
// @Description Returns all available subscription plans with pricing
// @Tags billing
// @Produce json
// @Success 200 {array} models.BillingPlan
// @Failure 500 {object} models.ErrorResponse
// @Router /api/billing/plans [get]
func GetPlans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := GetService()

		if !svc.IsAvailable() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Service Unavailable",
				Message: "Billing is not configured",
				Code:    http.StatusServiceUnavailable,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		prices, err := svc.GetPrices(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("failed to fetch stripe prices")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to fetch plans",
				Code:    http.StatusInternalServerError,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Convert Stripe prices to BillingPlan format
		var plans []models.BillingPlan
		for _, p := range prices {
			plan := models.BillingPlan{
				ID:       p.ID,
				PriceID:  p.ID,
				Amount:   p.UnitAmount,
				Currency: string(p.Currency),
			}

			if p.Recurring != nil {
				plan.Interval = string(p.Recurring.Interval)
			}

			if p.Product != nil {
				plan.Name = p.Product.Name
				plan.Description = p.Product.Description
			}

			plans = append(plans, plan)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(plans); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// CreateCheckoutSession creates a new checkout session for subscription
// @Summary Create checkout session
// @Description Creates a Stripe checkout session for subscription purchase
// @Tags billing
// @Accept json
// @Produce json
// @Param request body models.CreateCheckoutRequest true "Checkout request"
// @Success 200 {object} models.CheckoutSessionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/billing/checkout [post]
func CreateCheckoutSession(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := GetService()

		if !svc.IsAvailable() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Service Unavailable",
				Message: "Billing is not configured",
				Code:    http.StatusServiceUnavailable,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get user from context
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Parse request
		var req models.CreateCheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid request body",
				Code:    http.StatusBadRequest,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if req.PriceID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Price ID is required",
				Code:    http.StatusBadRequest,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get user from database
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get or create Stripe customer
		customerID, err := svc.GetOrCreateCustomer(r.Context(), &user)
		if err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to get or create stripe customer")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to create customer",
				Code:    http.StatusInternalServerError,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Create checkout session
		session, err := svc.CreateCheckoutSession(r.Context(), customerID, req.PriceID, config.SuccessURL, config.CancelURL)
		if err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to create checkout session")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to create checkout session",
				Code:    http.StatusInternalServerError,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		response := models.CheckoutSessionResponse{
			SessionID: session.ID,
			URL:       session.URL,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// CreatePortalSession creates a billing portal session
// @Summary Create billing portal session
// @Description Creates a Stripe billing portal session for subscription management
// @Tags billing
// @Produce json
// @Success 200 {object} models.PortalSessionResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/billing/portal [post]
func CreatePortalSession(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := GetService()

		if !svc.IsAvailable() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Service Unavailable",
				Message: "Billing is not configured",
				Code:    http.StatusServiceUnavailable,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get user from context
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get user from database
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if user.StripeCustomerID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "No billing account found",
				Code:    http.StatusBadRequest,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Create portal session
		session, err := svc.CreatePortalSession(r.Context(), user.StripeCustomerID, config.PortalReturnURL)
		if err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to create portal session")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to create portal session",
				Code:    http.StatusInternalServerError,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		response := models.PortalSessionResponse{
			URL: session.URL,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// GetSubscription returns the current user's subscription
// @Summary Get current subscription
// @Description Returns the current user's subscription details
// @Tags billing
// @Produce json
// @Success 200 {object} models.SubscriptionResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/billing/subscription [get]
func GetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get subscription from database
		var subscription models.Subscription
		if err := database.DB.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Not Found",
				Message: "No subscription found",
				Code:    http.StatusNotFound,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(subscription.ToSubscriptionResponse()); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
