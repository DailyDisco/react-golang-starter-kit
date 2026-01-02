// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"react-golang-starter/internal/models"
)

// MockStripeService provides a test double for Stripe operations.
type MockStripeService struct {
	mu               sync.RWMutex
	customers        map[string]*MockStripeCustomer
	subscriptions    map[string]*MockStripeSubscription
	checkoutSessions map[string]*MockCheckoutSession

	// Configurable behaviors
	CreateCustomerError        error
	CreateCheckoutSessionError error
	GetSubscriptionError       error
	CancelSubscriptionError    error
	CreatePortalSessionError   error

	// Track calls for assertions
	CreateCustomerCalls        int
	CreateCheckoutSessionCalls int
	GetSubscriptionCalls       int
	CancelSubscriptionCalls    int
}

// MockStripeCustomer represents a mock Stripe customer.
type MockStripeCustomer struct {
	ID       string
	Email    string
	Name     string
	Metadata map[string]string
}

// MockStripeSubscription represents a mock Stripe subscription.
type MockStripeSubscription struct {
	ID                string
	CustomerID        string
	Status            string
	PriceID           string
	CurrentPeriodEnd  int64
	CancelAtPeriodEnd bool
}

// MockCheckoutSession represents a mock Stripe checkout session.
type MockCheckoutSession struct {
	ID         string
	URL        string
	CustomerID string
	Mode       string
	PriceID    string
}

// NewMockStripeService creates a new mock Stripe service.
func NewMockStripeService() *MockStripeService {
	return &MockStripeService{
		customers:        make(map[string]*MockStripeCustomer),
		subscriptions:    make(map[string]*MockStripeSubscription),
		checkoutSessions: make(map[string]*MockCheckoutSession),
	}
}

// IsAvailable returns true for the mock.
func (m *MockStripeService) IsAvailable() bool {
	return true
}

// GetPublishableKey returns a mock publishable key.
func (m *MockStripeService) GetPublishableKey() string {
	return "pk_test_mock_key"
}

// CreateCustomer creates a mock customer.
func (m *MockStripeService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCustomerCalls++

	if m.CreateCustomerError != nil {
		return "", m.CreateCustomerError
	}

	customerID := fmt.Sprintf("cus_mock_%d", user.ID)
	m.customers[customerID] = &MockStripeCustomer{
		ID:    customerID,
		Email: user.Email,
		Name:  user.Name,
	}

	return customerID, nil
}

// GetOrCreateCustomer gets or creates a mock customer.
func (m *MockStripeService) GetOrCreateCustomer(ctx context.Context, user *models.User) (string, error) {
	m.mu.RLock()
	for _, c := range m.customers {
		if c.Email == user.Email {
			m.mu.RUnlock()
			return c.ID, nil
		}
	}
	m.mu.RUnlock()

	return m.CreateCustomer(ctx, user)
}

// CreateCheckoutSession creates a mock checkout session.
func (m *MockStripeService) CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCheckoutSessionCalls++

	if m.CreateCheckoutSessionError != nil {
		return "", m.CreateCheckoutSessionError
	}

	sessionID := fmt.Sprintf("cs_mock_%s_%d", customerID, time.Now().UnixNano())
	m.checkoutSessions[sessionID] = &MockCheckoutSession{
		ID:         sessionID,
		URL:        fmt.Sprintf("https://checkout.stripe.com/mock/%s", sessionID),
		CustomerID: customerID,
		Mode:       "subscription",
		PriceID:    priceID,
	}

	return m.checkoutSessions[sessionID].URL, nil
}

// CreatePortalSession creates a mock billing portal session.
func (m *MockStripeService) CreatePortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	if m.CreatePortalSessionError != nil {
		return "", m.CreatePortalSessionError
	}
	return "https://billing.stripe.com/mock/portal", nil
}

// GetSubscription retrieves a mock subscription.
func (m *MockStripeService) GetSubscription(ctx context.Context, subscriptionID string) (*MockStripeSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.GetSubscriptionCalls++

	if m.GetSubscriptionError != nil {
		return nil, m.GetSubscriptionError
	}

	if sub, ok := m.subscriptions[subscriptionID]; ok {
		return sub, nil
	}

	return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
}

// CancelSubscription cancels a mock subscription.
func (m *MockStripeService) CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) (*MockStripeSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CancelSubscriptionCalls++

	if m.CancelSubscriptionError != nil {
		return nil, m.CancelSubscriptionError
	}

	if sub, ok := m.subscriptions[subscriptionID]; ok {
		if immediate {
			sub.Status = "canceled"
		} else {
			sub.CancelAtPeriodEnd = true
		}
		return sub, nil
	}

	return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
}

// AddCustomer adds a customer to the mock store.
func (m *MockStripeService) AddCustomer(customer *MockStripeCustomer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customers[customer.ID] = customer
}

// AddSubscription adds a subscription to the mock store.
func (m *MockStripeService) AddSubscription(sub *MockStripeSubscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[sub.ID] = sub
}

// GetCustomer retrieves a customer from the mock store.
func (m *MockStripeService) GetCustomer(customerID string) *MockStripeCustomer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.customers[customerID]
}

// Reset clears all mock data and resets counters.
func (m *MockStripeService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.customers = make(map[string]*MockStripeCustomer)
	m.subscriptions = make(map[string]*MockStripeSubscription)
	m.checkoutSessions = make(map[string]*MockCheckoutSession)

	m.CreateCustomerError = nil
	m.CreateCheckoutSessionError = nil
	m.GetSubscriptionError = nil
	m.CancelSubscriptionError = nil
	m.CreatePortalSessionError = nil

	m.CreateCustomerCalls = 0
	m.CreateCheckoutSessionCalls = 0
	m.GetSubscriptionCalls = 0
	m.CancelSubscriptionCalls = 0
}
