package stripe

import (
	"os"
	"strings"
)

// Config holds Stripe configuration
type Config struct {
	SecretKey       string
	PublishableKey  string
	WebhookSecret   string
	SuccessURL      string
	CancelURL       string
	PortalReturnURL string
	Enabled         bool
	PremiumPriceID  string // Stripe Price ID for premium subscription
}

// DefaultConfig returns the default Stripe configuration
func DefaultConfig() *Config {
	return &Config{
		SecretKey:       "",
		PublishableKey:  "",
		WebhookSecret:   "",
		SuccessURL:      "http://localhost:5173/billing/success",
		CancelURL:       "http://localhost:5173/billing/cancel",
		PortalReturnURL: "http://localhost:5173/billing",
		Enabled:         false,
		PremiumPriceID:  "",
	}
}

// LoadConfig loads Stripe configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	if val := os.Getenv("STRIPE_SECRET_KEY"); val != "" {
		config.SecretKey = val
	}
	if val := os.Getenv("STRIPE_PUBLISHABLE_KEY"); val != "" {
		config.PublishableKey = val
	}
	if val := os.Getenv("STRIPE_WEBHOOK_SECRET"); val != "" {
		config.WebhookSecret = val
	}
	if val := os.Getenv("STRIPE_SUCCESS_URL"); val != "" {
		config.SuccessURL = val
	}
	if val := os.Getenv("STRIPE_CANCEL_URL"); val != "" {
		config.CancelURL = val
	}
	if val := os.Getenv("STRIPE_PORTAL_RETURN_URL"); val != "" {
		config.PortalReturnURL = val
	}
	if val := os.Getenv("STRIPE_PREMIUM_PRICE_ID"); val != "" {
		config.PremiumPriceID = val
	}

	// Enable Stripe if secret key is provided
	if val := os.Getenv("STRIPE_ENABLED"); val != "" {
		config.Enabled = strings.ToLower(val) == "true" || val == "1"
	} else {
		config.Enabled = config.SecretKey != ""
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.SecretKey == "" {
		return ErrMissingSecretKey
	}
	if c.PublishableKey == "" {
		return ErrMissingPublishableKey
	}

	return nil
}
