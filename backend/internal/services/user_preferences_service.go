package services

import (
	"encoding/json"
	"fmt"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"time"

	"gorm.io/gorm"
)

// UserPreferencesService handles user preferences operations
type UserPreferencesService struct {
	db *gorm.DB
}

// NewUserPreferencesService creates a new user preferences service instance
func NewUserPreferencesService() *UserPreferencesService {
	return &UserPreferencesService{
		db: database.DB,
	}
}

// GetPreferences retrieves user preferences, creating defaults if not exists
func (s *UserPreferencesService) GetPreferences(userID uint) (*models.UserPreferences, error) {
	var prefs models.UserPreferences
	result := s.db.Where("user_id = ?", userID).First(&prefs)

	if result.Error == gorm.ErrRecordNotFound {
		// Create default preferences
		prefs = s.createDefaultPreferences(userID)
		if err := s.db.Create(&prefs).Error; err != nil {
			return nil, fmt.Errorf("failed to create default preferences: %w", err)
		}
		return &prefs, nil
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve preferences: %w", result.Error)
	}

	return &prefs, nil
}

// UpdatePreferences updates user preferences
func (s *UserPreferencesService) UpdatePreferences(userID uint, req *models.UpdateUserPreferencesRequest) (*models.UserPreferences, error) {
	// Get or create preferences
	prefs, err := s.GetPreferences(userID)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	}

	if req.Theme != nil {
		if !isValidTheme(*req.Theme) {
			return nil, fmt.Errorf("invalid theme: must be 'light', 'dark', or 'system'")
		}
		updates["theme"] = *req.Theme
	}

	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}

	if req.Language != nil {
		if !isValidLanguage(*req.Language) {
			return nil, fmt.Errorf("invalid language: must be a supported language code")
		}
		updates["language"] = *req.Language
	}

	if req.DateFormat != nil {
		updates["date_format"] = *req.DateFormat
	}

	if req.TimeFormat != nil {
		if *req.TimeFormat != "12h" && *req.TimeFormat != "24h" {
			return nil, fmt.Errorf("invalid time format: must be '12h' or '24h'")
		}
		updates["time_format"] = *req.TimeFormat
	}

	if req.EmailNotifications != nil {
		emailNotifJSON, _ := json.Marshal(req.EmailNotifications)
		updates["email_notifications"] = emailNotifJSON
	}

	if err := s.db.Model(prefs).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	// Reload preferences
	return s.GetPreferences(userID)
}

// UpdateTheme updates just the theme preference
func (s *UserPreferencesService) UpdateTheme(userID uint, theme string) error {
	if !isValidTheme(theme) {
		return fmt.Errorf("invalid theme: must be 'light', 'dark', or 'system'")
	}

	// Ensure preferences exist
	_, err := s.GetPreferences(userID)
	if err != nil {
		return err
	}

	return s.db.Model(&models.UserPreferences{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"theme":      theme,
			"updated_at": time.Now().Format(time.RFC3339),
		}).Error
}

// UpdateEmailNotifications updates email notification preferences
func (s *UserPreferencesService) UpdateEmailNotifications(userID uint, notifications *models.EmailNotificationSettings) error {
	// Ensure preferences exist
	_, err := s.GetPreferences(userID)
	if err != nil {
		return err
	}

	notifJSON, _ := json.Marshal(notifications)

	return s.db.Model(&models.UserPreferences{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_notifications": notifJSON,
			"updated_at":          time.Now().Format(time.RFC3339),
		}).Error
}

// DeletePreferences deletes user preferences (for account deletion)
func (s *UserPreferencesService) DeletePreferences(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.UserPreferences{}).Error
}

// createDefaultPreferences creates default preferences for a new user
func (s *UserPreferencesService) createDefaultPreferences(userID uint) models.UserPreferences {
	defaultNotifications := models.EmailNotificationSettings{
		Marketing:    false,
		Security:     true,
		Updates:      true,
		WeeklyDigest: false,
	}
	notifJSON, _ := json.Marshal(defaultNotifications)

	return models.UserPreferences{
		UserID:             userID,
		Theme:              "system",
		Timezone:           "UTC",
		Language:           "en",
		DateFormat:         "MM/DD/YYYY",
		TimeFormat:         "12h",
		EmailNotifications: notifJSON,
		CreatedAt:          time.Now().Format(time.RFC3339),
		UpdatedAt:          time.Now().Format(time.RFC3339),
	}
}

func isValidTheme(theme string) bool {
	return theme == "light" || theme == "dark" || theme == "system"
}

func isValidLanguage(lang string) bool {
	_, exists := SupportedLanguages[lang]
	return exists
}

// Available timezones (subset of common ones)
var CommonTimezones = []string{
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Anchorage",
	"Pacific/Honolulu",
	"America/Phoenix",
	"America/Toronto",
	"America/Vancouver",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Madrid",
	"Europe/Rome",
	"Europe/Amsterdam",
	"Europe/Stockholm",
	"Europe/Moscow",
	"Asia/Dubai",
	"Asia/Kolkata",
	"Asia/Singapore",
	"Asia/Hong_Kong",
	"Asia/Tokyo",
	"Asia/Seoul",
	"Asia/Shanghai",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Pacific/Auckland",
}

// Available languages
var SupportedLanguages = map[string]string{
	"en": "English",
	"es": "Spanish",
	"fr": "French",
	"de": "German",
	"it": "Italian",
	"pt": "Portuguese",
	"ja": "Japanese",
	"ko": "Korean",
	"zh": "Chinese",
}

// Date formats
var SupportedDateFormats = []string{
	"MM/DD/YYYY",
	"DD/MM/YYYY",
	"YYYY-MM-DD",
	"DD.MM.YYYY",
	"YYYY/MM/DD",
}
