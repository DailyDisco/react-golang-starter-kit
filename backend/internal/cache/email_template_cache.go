package cache

import (
	"context"
	"time"

	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

const (
	// EmailTemplateKeyPrefix is the cache key prefix for email templates by key
	EmailTemplateKeyPrefix = "email_template:key:"
)

// Default TTL for email templates
var defaultEmailTemplateTTL = 5 * time.Minute

// emailTemplateKey generates a cache key for an email template by its key
func emailTemplateKey(key string) string {
	return EmailTemplateKeyPrefix + key
}

// GetEmailTemplate retrieves an email template from the cache by key
// Returns nil, nil if the template is not in the cache (cache miss)
func GetEmailTemplate(ctx context.Context, key string) (*models.EmailTemplate, error) {
	if !IsAvailable() {
		return nil, nil
	}

	var template models.EmailTemplate
	err := GetJSON(ctx, emailTemplateKey(key), &template)
	if err != nil {
		// Cache miss or error - return nil to indicate caller should query DB
		return nil, nil
	}

	log.Debug().Str("key", key).Msg("email template cache hit")
	return &template, nil
}

// SetEmailTemplate caches an email template
func SetEmailTemplate(ctx context.Context, template *models.EmailTemplate) error {
	if !IsAvailable() || template == nil {
		return nil
	}

	if err := SetJSON(ctx, emailTemplateKey(template.Key), template, defaultEmailTemplateTTL); err != nil {
		log.Warn().Err(err).Str("key", template.Key).Msg("failed to cache email template")
		return err
	}

	log.Debug().Str("key", template.Key).Msg("email template cached")
	return nil
}

// InvalidateEmailTemplate removes an email template from the cache by key
func InvalidateEmailTemplate(ctx context.Context, key string) error {
	if !IsAvailable() {
		return nil
	}

	if err := Delete(ctx, emailTemplateKey(key)); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("failed to invalidate email template cache")
		return err
	}

	log.Debug().Str("key", key).Msg("email template cache invalidated")
	return nil
}

// InvalidateAllEmailTemplates clears all email template caches
func InvalidateAllEmailTemplates(ctx context.Context) error {
	if !IsAvailable() {
		return nil
	}

	pattern := EmailTemplateKeyPrefix + "*"
	if inst := Instance(); inst != nil {
		if err := inst.Clear(ctx, pattern); err != nil {
			log.Warn().Err(err).Msg("failed to invalidate all email template caches")
			return err
		}
	}

	log.Debug().Msg("all email template caches invalidated")
	return nil
}
