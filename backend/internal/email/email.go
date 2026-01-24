package email

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SendParams contains parameters for sending an email
type SendParams struct {
	To           string                 // Recipient email address
	Subject      string                 // Email subject (optional if using template)
	TemplateName string                 // Template name (without .html)
	Data         map[string]interface{} // Template data
	PlainText    string                 // Plain text fallback (optional)
}

// EmailProvider defines the interface for email operations
type EmailProvider interface {
	// Send sends an email
	Send(ctx context.Context, params SendParams) error

	// SendBatch sends multiple emails
	SendBatch(ctx context.Context, params []SendParams) error

	// IsAvailable returns whether the provider is available
	IsAvailable() bool

	// Close closes any connections
	Close() error
}

// Sentinel errors
var (
	ErrNotInitialized = errors.New("email service not initialized")
	ErrInvalidParams  = errors.New("invalid email parameters")
	ErrSendFailed     = errors.New("failed to send email")
)

// instance holds the global email provider
var instance EmailProvider
var config *Config

// Initialize sets up the email provider based on configuration
// The db parameter is optional - if provided, enables database template resolution
func Initialize(cfg *Config, db *gorm.DB) error {
	config = cfg

	if !cfg.Enabled {
		log.Info().Msg("email disabled, using no-op provider")
		instance = NewNoOpProvider()
		return nil
	}

	// Initialize SMTP provider with optional database for template resolution
	smtpProvider, err := NewSMTPProvider(cfg, db)
	if err != nil {
		log.Warn().Err(err).Msg("failed to initialize SMTP provider, using no-op")
		instance = NewNoOpProvider()
		return nil
	}

	log.Info().
		Str("host", cfg.SMTPHost).
		Int("port", cfg.SMTPPort).
		Str("from", cfg.SMTPFrom).
		Bool("dev_mode", cfg.DevMode).
		Bool("db_templates", db != nil).
		Msg("SMTP email provider initialized")

	instance = smtpProvider
	return nil
}

// Send sends an email using the global provider
func Send(ctx context.Context, params SendParams) error {
	if instance == nil {
		return ErrNotInitialized
	}
	return instance.Send(ctx, params)
}

// SendBatch sends multiple emails
func SendBatch(ctx context.Context, params []SendParams) error {
	if instance == nil {
		return ErrNotInitialized
	}
	return instance.SendBatch(ctx, params)
}

// IsAvailable returns whether email is available
func IsAvailable() bool {
	if instance == nil {
		return false
	}
	return instance.IsAvailable()
}

// GetFrontendURL returns the configured frontend URL
func GetFrontendURL() string {
	if config == nil {
		return "http://localhost:5173"
	}
	return config.FrontendURL
}

// Close closes the email provider
func Close() error {
	if instance == nil {
		return nil
	}
	return instance.Close()
}

// GetConfig returns the current email configuration (for testing)
func GetConfig() *Config {
	return config
}
