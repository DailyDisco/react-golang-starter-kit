package stripe

import (
	"context"
	"os"
	"sync"
	"testing"
)

// ============ Config Tests ============

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Enabled {
		t.Error("DefaultConfig().Enabled = true, want false")
	}

	if config.SecretKey != "" {
		t.Errorf("DefaultConfig().SecretKey = %q, want empty", config.SecretKey)
	}

	if config.PublishableKey != "" {
		t.Errorf("DefaultConfig().PublishableKey = %q, want empty", config.PublishableKey)
	}

	if config.SuccessURL != "http://localhost:5173/billing/success" {
		t.Errorf("DefaultConfig().SuccessURL = %q, want %q", config.SuccessURL, "http://localhost:5173/billing/success")
	}

	if config.CancelURL != "http://localhost:5173/billing/cancel" {
		t.Errorf("DefaultConfig().CancelURL = %q, want %q", config.CancelURL, "http://localhost:5173/billing/cancel")
	}

	if config.PortalReturnURL != "http://localhost:5173/billing" {
		t.Errorf("DefaultConfig().PortalReturnURL = %q, want %q", config.PortalReturnURL, "http://localhost:5173/billing")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all Stripe env vars
	envVars := []string{
		"STRIPE_SECRET_KEY",
		"STRIPE_PUBLISHABLE_KEY",
		"STRIPE_WEBHOOK_SECRET",
		"STRIPE_SUCCESS_URL",
		"STRIPE_CANCEL_URL",
		"STRIPE_PORTAL_RETURN_URL",
		"STRIPE_ENABLED",
		"STRIPE_PREMIUM_PRICE_ID",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	config := LoadConfig()

	if config.Enabled {
		t.Error("LoadConfig() should be disabled when no secret key is set")
	}

	if config.SecretKey != "" {
		t.Errorf("LoadConfig().SecretKey = %q, want empty", config.SecretKey)
	}
}

func TestLoadConfig_WithSecretKey(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_12345")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_test_12345")

	config := LoadConfig()

	if !config.Enabled {
		t.Error("LoadConfig() should be enabled when secret key is set")
	}

	if config.SecretKey != "sk_test_12345" {
		t.Errorf("LoadConfig().SecretKey = %q, want %q", config.SecretKey, "sk_test_12345")
	}

	if config.PublishableKey != "pk_test_12345" {
		t.Errorf("LoadConfig().PublishableKey = %q, want %q", config.PublishableKey, "pk_test_12345")
	}
}

func TestLoadConfig_ExplicitEnabled(t *testing.T) {
	t.Setenv("STRIPE_ENABLED", "true")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_12345")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_test_12345")

	config := LoadConfig()

	if !config.Enabled {
		t.Error("LoadConfig() should be enabled when STRIPE_ENABLED=true")
	}
}

func TestLoadConfig_ExplicitDisabled(t *testing.T) {
	t.Setenv("STRIPE_ENABLED", "false")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_12345")

	config := LoadConfig()

	if config.Enabled {
		t.Error("LoadConfig() should be disabled when STRIPE_ENABLED=false even with secret key")
	}
}

func TestLoadConfig_CustomURLs(t *testing.T) {
	t.Setenv("STRIPE_SUCCESS_URL", "https://example.com/success")
	t.Setenv("STRIPE_CANCEL_URL", "https://example.com/cancel")
	t.Setenv("STRIPE_PORTAL_RETURN_URL", "https://example.com/billing")

	config := LoadConfig()

	if config.SuccessURL != "https://example.com/success" {
		t.Errorf("LoadConfig().SuccessURL = %q, want %q", config.SuccessURL, "https://example.com/success")
	}

	if config.CancelURL != "https://example.com/cancel" {
		t.Errorf("LoadConfig().CancelURL = %q, want %q", config.CancelURL, "https://example.com/cancel")
	}

	if config.PortalReturnURL != "https://example.com/billing" {
		t.Errorf("LoadConfig().PortalReturnURL = %q, want %q", config.PortalReturnURL, "https://example.com/billing")
	}
}

func TestLoadConfig_PremiumPriceID(t *testing.T) {
	t.Setenv("STRIPE_PREMIUM_PRICE_ID", "price_12345")

	config := LoadConfig()

	if config.PremiumPriceID != "price_12345" {
		t.Errorf("LoadConfig().PremiumPriceID = %q, want %q", config.PremiumPriceID, "price_12345")
	}
}

func TestLoadConfig_WebhookSecret(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_12345")

	config := LoadConfig()

	if config.WebhookSecret != "whsec_12345" {
		t.Errorf("LoadConfig().WebhookSecret = %q, want %q", config.WebhookSecret, "whsec_12345")
	}
}

// ============ Config Validation Tests ============

func TestConfig_Validate_Disabled(t *testing.T) {
	config := &Config{
		Enabled: false,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Config.Validate() error = %v, want nil for disabled config", err)
	}
}

func TestConfig_Validate_MissingSecretKey(t *testing.T) {
	config := &Config{
		Enabled:        true,
		SecretKey:      "",
		PublishableKey: "pk_test_12345",
	}

	err := config.Validate()
	if err != ErrMissingSecretKey {
		t.Errorf("Config.Validate() error = %v, want %v", err, ErrMissingSecretKey)
	}
}

func TestConfig_Validate_MissingPublishableKey(t *testing.T) {
	config := &Config{
		Enabled:        true,
		SecretKey:      "sk_test_12345",
		PublishableKey: "",
	}

	err := config.Validate()
	if err != ErrMissingPublishableKey {
		t.Errorf("Config.Validate() error = %v, want %v", err, ErrMissingPublishableKey)
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	config := &Config{
		Enabled:        true,
		SecretKey:      "sk_test_12345",
		PublishableKey: "pk_test_12345",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Config.Validate() error = %v, want nil", err)
	}
}

// ============ Error Types Tests ============

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrMissingSecretKey",
			err:  ErrMissingSecretKey,
			want: "stripe: missing secret key",
		},
		{
			name: "ErrMissingPublishableKey",
			err:  ErrMissingPublishableKey,
			want: "stripe: missing publishable key",
		},
		{
			name: "ErrNotInitialized",
			err:  ErrNotInitialized,
			want: "stripe: service not initialized",
		},
		{
			name: "ErrDisabled",
			err:  ErrDisabled,
			want: "stripe: service is disabled",
		},
		{
			name: "ErrCustomerNotFound",
			err:  ErrCustomerNotFound,
			want: "stripe: customer not found",
		},
		{
			name: "ErrCustomerExists",
			err:  ErrCustomerExists,
			want: "stripe: customer already exists",
		},
		{
			name: "ErrSubscriptionNotFound",
			err:  ErrSubscriptionNotFound,
			want: "stripe: subscription not found",
		},
		{
			name: "ErrNoActiveSubscription",
			err:  ErrNoActiveSubscription,
			want: "stripe: no active subscription",
		},
		{
			name: "ErrInvalidSignature",
			err:  ErrInvalidSignature,
			want: "stripe: invalid webhook signature",
		},
		{
			name: "ErrUnhandledEvent",
			err:  ErrUnhandledEvent,
			want: "stripe: unhandled event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

// ============ NoOp Service Tests ============

func TestNoOpService_IsAvailable(t *testing.T) {
	svc := &noOpService{}

	if svc.IsAvailable() {
		t.Error("noOpService.IsAvailable() = true, want false")
	}
}

func TestNoOpService_GetPublishableKey(t *testing.T) {
	svc := &noOpService{}

	if key := svc.GetPublishableKey(); key != "" {
		t.Errorf("noOpService.GetPublishableKey() = %q, want empty", key)
	}
}

func TestNoOpService_CreateCustomer(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.CreateCustomer(context.Background(), nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.CreateCustomer() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GetOrCreateCustomer(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GetOrCreateCustomer(context.Background(), nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.GetOrCreateCustomer() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_CreateCheckoutSession(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.CreateCheckoutSession(context.Background(), "", "", "", "")
	if err != ErrDisabled {
		t.Errorf("noOpService.CreateCheckoutSession() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_CreatePortalSession(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.CreatePortalSession(context.Background(), "", "")
	if err != ErrDisabled {
		t.Errorf("noOpService.CreatePortalSession() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GetSubscription(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GetSubscription(context.Background(), "")
	if err != ErrDisabled {
		t.Errorf("noOpService.GetSubscription() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_CancelSubscription(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.CancelSubscription(context.Background(), "", true)
	if err != ErrDisabled {
		t.Errorf("noOpService.CancelSubscription() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GetPrices(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GetPrices(context.Background())
	if err != ErrDisabled {
		t.Errorf("noOpService.GetPrices() error = %v, want %v", err, ErrDisabled)
	}
}

// ============ Initialize Tests ============

func TestInitialize_Disabled(t *testing.T) {
	// Reset singleton for testing
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled: false,
	}

	err := Initialize(config)
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}

	svc := GetService()
	if svc == nil {
		t.Fatal("GetService() returned nil after Initialize")
	}

	if svc.IsAvailable() {
		t.Error("Service should not be available when disabled")
	}
}

func TestInitialize_MissingSecretKey(t *testing.T) {
	// Reset singleton for testing
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled:        true,
		SecretKey:      "",
		PublishableKey: "pk_test_12345",
	}

	err := Initialize(config)
	if err != ErrMissingSecretKey {
		t.Errorf("Initialize() error = %v, want %v", err, ErrMissingSecretKey)
	}
}

func TestInitialize_MissingPublishableKey(t *testing.T) {
	// Reset singleton for testing
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled:        true,
		SecretKey:      "sk_test_12345",
		PublishableKey: "",
	}

	err := Initialize(config)
	if err != ErrMissingPublishableKey {
		t.Errorf("Initialize() error = %v, want %v", err, ErrMissingPublishableKey)
	}
}

// ============ GetService Tests ============

func TestGetService_BeforeInit(t *testing.T) {
	// Reset singleton
	once = sync.Once{}
	instance = nil

	svc := GetService()
	if svc != nil {
		t.Error("GetService() should return nil before initialization")
	}
}

// ============ IsAvailable Package Function Tests ============

func TestIsAvailable_NotInitialized(t *testing.T) {
	// Reset singleton
	once = sync.Once{}
	instance = nil

	if IsAvailable() {
		t.Error("IsAvailable() = true when not initialized, want false")
	}
}

func TestIsAvailable_Disabled(t *testing.T) {
	// Reset singleton and initialize with disabled config
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled: false,
	}
	Initialize(config)

	if IsAvailable() {
		t.Error("IsAvailable() = true when disabled, want false")
	}
}

// ============ StripeService Tests ============

func TestStripeService_GetPublishableKey(t *testing.T) {
	svc := &stripeService{
		config: &Config{
			PublishableKey: "pk_test_key",
		},
	}

	if key := svc.GetPublishableKey(); key != "pk_test_key" {
		t.Errorf("stripeService.GetPublishableKey() = %q, want %q", key, "pk_test_key")
	}
}

func TestStripeService_IsAvailable(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{
			name:    "enabled",
			enabled: true,
			want:    true,
		},
		{
			name:    "disabled",
			enabled: false,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stripeService{
				config: &Config{
					Enabled: tt.enabled,
				},
			}

			if got := svc.IsAvailable(); got != tt.want {
				t.Errorf("stripeService.IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
