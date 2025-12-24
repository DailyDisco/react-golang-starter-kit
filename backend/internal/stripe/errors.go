package stripe

import "errors"

var (
	// Configuration errors
	ErrMissingSecretKey      = errors.New("stripe: missing secret key")
	ErrMissingPublishableKey = errors.New("stripe: missing publishable key")
	ErrNotInitialized        = errors.New("stripe: service not initialized")
	ErrDisabled              = errors.New("stripe: service is disabled")

	// Customer errors
	ErrCustomerNotFound = errors.New("stripe: customer not found")
	ErrCustomerExists   = errors.New("stripe: customer already exists")

	// Subscription errors
	ErrSubscriptionNotFound = errors.New("stripe: subscription not found")
	ErrNoActiveSubscription = errors.New("stripe: no active subscription")

	// Webhook errors
	ErrInvalidSignature = errors.New("stripe: invalid webhook signature")
	ErrUnhandledEvent   = errors.New("stripe: unhandled event type")
)
