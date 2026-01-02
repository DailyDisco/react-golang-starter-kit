package stripe

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/subscription"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
)

// Service defines the Stripe service interface
type Service interface {
	// Customer operations
	CreateCustomer(ctx context.Context, user *models.User) (string, error)
	GetOrCreateCustomer(ctx context.Context, user *models.User) (string, error)

	// Checkout operations
	CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error)

	// Portal operations
	CreatePortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error)

	// Subscription operations
	GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error)

	// Price/Plan operations
	GetPrices(ctx context.Context) ([]*stripe.Price, error)

	// Configuration
	GetPublishableKey() string
	IsAvailable() bool
}

// stripeService implements the Service interface
type stripeService struct {
	config *Config
}

var (
	instance Service
	once     sync.Once
	mu       sync.RWMutex
)

// Initialize sets up the Stripe service
func Initialize(config *Config) error {
	var initErr error

	once.Do(func() {
		if err := config.Validate(); err != nil {
			initErr = err
			return
		}

		if !config.Enabled {
			log.Info().Msg("stripe service disabled")
			instance = &noOpService{}
			return
		}

		// Set the Stripe API key globally
		stripe.Key = config.SecretKey

		instance = &stripeService{
			config: config,
		}

		log.Info().Msg("stripe service initialized")
	})

	return initErr
}

// GetService returns the global Stripe service instance
func GetService() Service {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// IsAvailable returns true if the Stripe service is initialized and available
func IsAvailable() bool {
	svc := GetService()
	return svc != nil && svc.IsAvailable()
}

// CreateCustomer creates a new Stripe customer for a user
func (s *stripeService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(user.Email),
		Name:  stripe.String(user.Name),
		Metadata: map[string]string{
			"user_id": strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return cust.ID, nil
}

// GetOrCreateCustomer returns existing customer ID or creates a new one
func (s *stripeService) GetOrCreateCustomer(ctx context.Context, user *models.User) (string, error) {
	// Check if user already has a Stripe customer ID
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		return *user.StripeCustomerID, nil
	}

	// Create new customer
	customerID, err := s.CreateCustomer(ctx, user)
	if err != nil {
		return "", err
	}

	// Update user with customer ID
	user.StripeCustomerID = &customerID
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(user).Error; err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to save stripe customer ID")
		return customerID, nil // Return customer ID anyway, DB update is not critical
	}

	return customerID, nil
}

// CreateCheckoutSession creates a new Stripe checkout session
func (s *stripeService) CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(cancelURL),
	}

	return checkoutsession.New(params)
}

// CreatePortalSession creates a billing portal session for customer self-management
func (s *stripeService) CreatePortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	return session.New(params)
}

// GetSubscription retrieves a subscription by ID
func (s *stripeService) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	return subscription.Get(subscriptionID, nil)
}

// CancelSubscription cancels a subscription
func (s *stripeService) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error) {
	if cancelAtPeriodEnd {
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		return subscription.Update(subscriptionID, params)
	}

	return subscription.Cancel(subscriptionID, nil)
}

// GetPrices retrieves all active prices
func (s *stripeService) GetPrices(ctx context.Context) ([]*stripe.Price, error) {
	params := &stripe.PriceListParams{
		Active: stripe.Bool(true),
	}
	params.AddExpand("data.product")

	var prices []*stripe.Price
	iter := price.List(params)
	for iter.Next() {
		prices = append(prices, iter.Price())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}

// GetPublishableKey returns the Stripe publishable key
func (s *stripeService) GetPublishableKey() string {
	return s.config.PublishableKey
}

// IsAvailable returns true if the service is enabled
func (s *stripeService) IsAvailable() bool {
	return s.config.Enabled
}

// noOpService is a no-op implementation when Stripe is disabled
type noOpService struct{}

func (n *noOpService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	return "", ErrDisabled
}

func (n *noOpService) GetOrCreateCustomer(ctx context.Context, user *models.User) (string, error) {
	return "", ErrDisabled
}

func (n *noOpService) CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	return nil, ErrDisabled
}

func (n *noOpService) CreatePortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	return nil, ErrDisabled
}

func (n *noOpService) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GetPrices(ctx context.Context) ([]*stripe.Price, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GetPublishableKey() string {
	return ""
}

func (n *noOpService) IsAvailable() bool {
	return false
}
