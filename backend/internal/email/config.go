package email

import (
	"os"
	"strconv"
	"time"
)

// Config holds email configuration
type Config struct {
	// General settings
	Enabled bool

	// SMTP settings
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// TLS settings
	UseTLS    bool
	TLSPolicy string // "mandatory", "opportunistic", "none"

	// Timeouts
	ConnectTimeout time.Duration
	SendTimeout    time.Duration

	// Frontend URL for email links
	FrontendURL string

	// Development mode (log emails instead of sending)
	DevMode bool
}

// DefaultConfig returns sensible default email configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		SMTPHost:       "localhost",
		SMTPPort:       587,
		SMTPFrom:       "noreply@example.com",
		SMTPFromName:   "App",
		UseTLS:         true,
		TLSPolicy:      "opportunistic",
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    30 * time.Second,
		FrontendURL:    "http://localhost:5173",
		DevMode:        true,
	}
}

// LoadConfig loads email configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	// Check if email is enabled
	if enabled := os.Getenv("EMAIL_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true"
	}

	// SMTP_HOST enables email if set
	if host := os.Getenv("SMTP_HOST"); host != "" {
		config.SMTPHost = host
		config.Enabled = true
	}

	if port := os.Getenv("SMTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.SMTPPort = p
		}
	}

	if user := os.Getenv("SMTP_USER"); user != "" {
		config.SMTPUser = user
	}

	if pass := os.Getenv("SMTP_PASSWORD"); pass != "" {
		config.SMTPPassword = pass
	}

	if from := os.Getenv("SMTP_FROM"); from != "" {
		config.SMTPFrom = from
	}
	if from := os.Getenv("EMAIL_FROM"); from != "" {
		config.SMTPFrom = from // Alias
	}

	if name := os.Getenv("SMTP_FROM_NAME"); name != "" {
		config.SMTPFromName = name
	}

	if url := os.Getenv("FRONTEND_URL"); url != "" {
		config.FrontendURL = url
	}

	if policy := os.Getenv("SMTP_TLS_POLICY"); policy != "" {
		config.TLSPolicy = policy
	}

	// Dev mode based on environment
	env := os.Getenv("GO_ENV")
	config.DevMode = env == "" || env == "development"

	// Override dev mode if explicitly set
	if devMode := os.Getenv("EMAIL_DEV_MODE"); devMode != "" {
		config.DevMode = devMode == "true"
	}

	return config
}
