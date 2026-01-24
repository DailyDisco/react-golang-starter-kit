package email

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/wneessen/go-mail"
	"gorm.io/gorm"
)

// SMTPProvider implements EmailProvider using SMTP
type SMTPProvider struct {
	config           *Config
	templates        *TemplateManager
	templateResolver *TemplateResolver
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(cfg *Config, db *gorm.DB) (*SMTPProvider, error) {
	templates, err := NewTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	provider := &SMTPProvider{
		config:    cfg,
		templates: templates,
	}

	// Initialize template resolver if database is available
	if db != nil {
		provider.templateResolver = NewTemplateResolver(db, templates)
	}

	return provider, nil
}

// Send sends an email via SMTP
func (p *SMTPProvider) Send(ctx context.Context, params SendParams) error {
	// Validate params
	if params.To == "" {
		return fmt.Errorf("%w: recipient email is required", ErrInvalidParams)
	}

	// In dev mode, just log the email
	if p.config.DevMode {
		log.Info().
			Str("to", params.To).
			Str("template", params.TemplateName).
			Str("subject", params.Subject).
			Interface("data", params.Data).
			Msg("DEV MODE: would send email")
		return nil
	}

	// Create message
	m := mail.NewMsg()

	// Set sender
	if err := m.FromFormat(p.config.SMTPFromName, p.config.SMTPFrom); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}

	// Set recipient
	if err := m.To(params.To); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}

	var subject, htmlBody, textBody string
	var err error
	var templateSource string

	// Render template if specified
	if params.TemplateName != "" {
		// Use template resolver if available (checks DB first, falls back to file)
		if p.templateResolver != nil {
			result, renderErr := p.templateResolver.Render(ctx, params.TemplateName, params.Data)
			if renderErr != nil {
				return fmt.Errorf("failed to render template: %w", renderErr)
			}
			subject = result.Subject
			htmlBody = result.BodyHTML
			textBody = result.BodyText
			templateSource = result.Source
		} else {
			// Fall back to direct file-based templates
			subject, htmlBody, textBody, err = p.templates.Render(params.TemplateName, params.Data)
			if err != nil {
				return fmt.Errorf("failed to render template: %w", err)
			}
			templateSource = "file"
		}
	}

	// Use provided subject if template doesn't specify one
	if subject == "" && params.Subject != "" {
		subject = params.Subject
	}
	if subject == "" {
		subject = "No Subject"
	}
	m.Subject(subject)

	// Set body
	if htmlBody != "" {
		m.SetBodyString(mail.TypeTextHTML, htmlBody)
	}
	if textBody != "" {
		m.AddAlternativeString(mail.TypeTextPlain, textBody)
	} else if params.PlainText != "" {
		m.AddAlternativeString(mail.TypeTextPlain, params.PlainText)
	}

	// Create SMTP client options
	clientOpts := []mail.Option{
		mail.WithPort(p.config.SMTPPort),
		mail.WithTLSPolicy(p.getTLSPolicy()),
	}

	// Add authentication if credentials are provided
	if p.config.SMTPUser != "" && p.config.SMTPPassword != "" {
		clientOpts = append(clientOpts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(p.config.SMTPUser),
			mail.WithPassword(p.config.SMTPPassword),
		)
	}

	// Add TLS config
	clientOpts = append(clientOpts,
		mail.WithTLSConfig(&tls.Config{
			ServerName: p.config.SMTPHost,
		}),
	)

	// Create SMTP client
	client, err := mail.NewClient(p.config.SMTPHost, clientOpts...)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Send email
	if err := client.DialAndSend(m); err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	log.Info().
		Str("to", params.To).
		Str("subject", subject).
		Str("template", params.TemplateName).
		Str("source", templateSource).
		Msg("email sent successfully")

	// Track send count for database templates
	if templateSource == "database" && p.templateResolver != nil {
		go func() {
			if err := p.templateResolver.IncrementSendCount(context.Background(), params.TemplateName); err != nil {
				log.Warn().Err(err).Str("template", params.TemplateName).Msg("failed to increment template send count")
			}
		}()
	}

	return nil
}

// SendBatch sends multiple emails
func (p *SMTPProvider) SendBatch(ctx context.Context, params []SendParams) error {
	var lastErr error
	for _, param := range params {
		if err := p.Send(ctx, param); err != nil {
			lastErr = err
			log.Error().Err(err).Str("to", param.To).Msg("failed to send email in batch")
		}
	}
	return lastErr
}

// IsAvailable returns whether SMTP is available
func (p *SMTPProvider) IsAvailable() bool {
	return p.config.SMTPHost != "" && p.config.SMTPPort > 0
}

// Close is a no-op for SMTP (connections are created per-send)
func (p *SMTPProvider) Close() error {
	return nil
}

func (p *SMTPProvider) getTLSPolicy() mail.TLSPolicy {
	switch p.config.TLSPolicy {
	case "mandatory":
		return mail.TLSMandatory
	case "none":
		return mail.NoTLS
	default:
		return mail.TLSOpportunistic
	}
}

// NoOpProvider implements EmailProvider but does nothing (for testing/disabled mode)
type NoOpProvider struct{}

// NewNoOpProvider creates a new no-op email provider
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{}
}

// Send logs the email but doesn't send it
func (p *NoOpProvider) Send(ctx context.Context, params SendParams) error {
	log.Debug().
		Str("to", params.To).
		Str("template", params.TemplateName).
		Msg("no-op: email not sent (provider disabled)")
	return nil
}

// SendBatch logs emails but doesn't send them
func (p *NoOpProvider) SendBatch(ctx context.Context, params []SendParams) error {
	for _, param := range params {
		log.Debug().
			Str("to", param.To).
			Str("template", param.TemplateName).
			Msg("no-op: email not sent (provider disabled)")
	}
	return nil
}

// IsAvailable returns false for no-op provider
func (p *NoOpProvider) IsAvailable() bool {
	return false
}

// Close is a no-op
func (p *NoOpProvider) Close() error {
	return nil
}
