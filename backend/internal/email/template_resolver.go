package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TemplateResolver resolves email templates from database or file-based fallback
type TemplateResolver struct {
	db            *gorm.DB
	fileTemplates *TemplateManager
}

// RenderResult contains the rendered email content
type RenderResult struct {
	Subject  string
	BodyHTML string
	BodyText string
	Source   string // "database" or "file"
}

// NewTemplateResolver creates a new template resolver
func NewTemplateResolver(db *gorm.DB, fileTemplates *TemplateManager) *TemplateResolver {
	return &TemplateResolver{
		db:            db,
		fileTemplates: fileTemplates,
	}
}

// Render renders a template by key, checking database first then falling back to files
func (tr *TemplateResolver) Render(ctx context.Context, templateKey string, data map[string]interface{}) (*RenderResult, error) {
	// 1. Try to get from cache first
	dbTemplate, _ := cache.GetEmailTemplate(ctx, templateKey)

	// 2. If not in cache, try database
	if dbTemplate == nil {
		var err error
		dbTemplate, err = tr.getTemplateFromDB(ctx, templateKey)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Err(err).Str("key", templateKey).Msg("failed to query database template")
		}

		// Cache the result if found
		if dbTemplate != nil {
			cache.SetEmailTemplate(ctx, dbTemplate)
		}
	}

	// 3. Use database template if exists and is active
	if dbTemplate != nil && dbTemplate.IsActive {
		result, err := tr.renderDatabaseTemplate(dbTemplate, data)
		if err != nil {
			log.Warn().Err(err).Str("key", templateKey).Msg("database template render failed, falling back to file")
		} else {
			return result, nil
		}
	}

	// 4. Fall back to file-based template
	return tr.renderFileTemplate(templateKey, data)
}

// getTemplateFromDB retrieves a template from the database by key
func (tr *TemplateResolver) getTemplateFromDB(ctx context.Context, key string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := tr.db.WithContext(ctx).Where("key = ?", key).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// renderDatabaseTemplate renders a database template with simple variable substitution
func (tr *TemplateResolver) renderDatabaseTemplate(tmpl *models.EmailTemplate, data map[string]interface{}) (*RenderResult, error) {
	subject := tr.substituteVariables(tmpl.Subject, data)
	bodyHTML := tr.substituteVariables(tmpl.BodyHTML, data)

	// Generate plain text from provided or from HTML
	var bodyText string
	if tmpl.BodyText != "" {
		bodyText = tr.substituteVariables(tmpl.BodyText, data)
	} else {
		bodyText = generatePlainText(bodyHTML)
	}

	return &RenderResult{
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
		Source:   "database",
	}, nil
}

// substituteVariables performs {{variable}} substitution
// Maps worker data keys (PascalCase) to snake_case for database template compatibility
func (tr *TemplateResolver) substituteVariables(content string, data map[string]interface{}) string {
	result := content

	// Add common variables
	fullData := map[string]interface{}{
		"site_name":     getAppName(),
		"app_name":      getAppName(),
		"support_email": getSupportEmail(),
		"current_year":  time.Now().Year(),
		"frontend_url":  GetFrontendURL(),
	}

	// Merge in provided data with both original and snake_case keys
	for k, v := range data {
		fullData[k] = v
		// Also add snake_case version
		snakeKey := toSnakeCase(k)
		if snakeKey != k {
			fullData[snakeKey] = v
		}
	}

	// Substitute all {{key}} patterns
	for key, value := range fullData {
		placeholder := "{{" + key + "}}"
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result
}

// renderFileTemplate uses the existing file-based template system
func (tr *TemplateResolver) renderFileTemplate(templateKey string, data map[string]interface{}) (*RenderResult, error) {
	subject, bodyHTML, bodyText, err := tr.fileTemplates.Render(templateKey, data)
	if err != nil {
		return nil, err
	}

	return &RenderResult{
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
		Source:   "file",
	}, nil
}

// IncrementSendCount updates the send statistics for a template
func (tr *TemplateResolver) IncrementSendCount(ctx context.Context, key string) error {
	now := time.Now().Format(time.RFC3339)
	return tr.db.WithContext(ctx).Model(&models.EmailTemplate{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"send_count":   gorm.Expr("send_count + 1"),
			"last_sent_at": now,
		}).Error
}

// toSnakeCase converts PascalCase or camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
