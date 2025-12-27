package services

import (
	"encoding/json"
	"fmt"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"time"

	"gorm.io/gorm"
)

// SettingsService handles system settings operations
type SettingsService struct{}

// NewSettingsService creates a new settings service instance
func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

// db returns the database connection - accessed at runtime to avoid nil issues
func (s *SettingsService) db() *gorm.DB {
	return database.DB
}

// GetAllSettings retrieves all system settings
func (s *SettingsService) GetAllSettings() ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	if err := s.db().Order("category, key").Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve settings: %w", err)
	}
	return settings, nil
}

// GetSettingsByCategory retrieves settings for a specific category
func (s *SettingsService) GetSettingsByCategory(category string) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	if err := s.db().Where("category = ?", category).Order("key").Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve settings for category %s: %w", category, err)
	}
	return settings, nil
}

// GetSetting retrieves a single setting by key
func (s *SettingsService) GetSetting(key string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	if err := s.db().Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("setting not found: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve setting: %w", err)
	}
	return &setting, nil
}

// GetSettingValue retrieves the value of a setting by key and unmarshals it into the provided interface
func (s *SettingsService) GetSettingValue(key string, dest interface{}) error {
	setting, err := s.GetSetting(key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(setting.Value, dest); err != nil {
		return fmt.Errorf("failed to unmarshal setting value: %w", err)
	}
	return nil
}

// UpdateSetting updates a setting value by key
func (s *SettingsService) UpdateSetting(key string, value interface{}) error {
	// Marshal the value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setting value: %w", err)
	}

	// Update the setting
	result := s.db().Model(&models.SystemSetting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      jsonValue,
			"updated_at": time.Now().Format(time.RFC3339),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update setting: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}
	return nil
}

// GetEmailSettings retrieves email/SMTP configuration
func (s *SettingsService) GetEmailSettings() (*models.EmailSettings, error) {
	settings := &models.EmailSettings{}

	var smtpHost, smtpUser, fromEmail, fromName string
	var smtpPort int
	var enabled bool

	if err := s.GetSettingValue("smtp_host", &smtpHost); err == nil {
		settings.SMTPHost = smtpHost
	}
	if err := s.GetSettingValue("smtp_port", &smtpPort); err == nil {
		settings.SMTPPort = smtpPort
	}
	if err := s.GetSettingValue("smtp_user", &smtpUser); err == nil {
		settings.SMTPUser = smtpUser
	}
	if err := s.GetSettingValue("smtp_from_email", &fromEmail); err == nil {
		settings.FromEmail = fromEmail
	}
	if err := s.GetSettingValue("smtp_from_name", &fromName); err == nil {
		settings.FromName = fromName
	}
	if err := s.GetSettingValue("smtp_enabled", &enabled); err == nil {
		settings.Enabled = enabled
	}

	// Note: Password is never returned
	return settings, nil
}

// UpdateEmailSettings updates email/SMTP configuration
func (s *SettingsService) UpdateEmailSettings(settings *models.EmailSettings) error {
	if settings.SMTPHost != "" {
		if err := s.UpdateSetting("smtp_host", settings.SMTPHost); err != nil {
			return err
		}
	}
	if settings.SMTPPort > 0 {
		if err := s.UpdateSetting("smtp_port", settings.SMTPPort); err != nil {
			return err
		}
	}
	if settings.SMTPUser != "" {
		if err := s.UpdateSetting("smtp_user", settings.SMTPUser); err != nil {
			return err
		}
	}
	if settings.SMTPPassword != "" {
		if err := s.UpdateSetting("smtp_password", settings.SMTPPassword); err != nil {
			return err
		}
	}
	if settings.FromEmail != "" {
		if err := s.UpdateSetting("smtp_from_email", settings.FromEmail); err != nil {
			return err
		}
	}
	if settings.FromName != "" {
		if err := s.UpdateSetting("smtp_from_name", settings.FromName); err != nil {
			return err
		}
	}
	if err := s.UpdateSetting("smtp_enabled", settings.Enabled); err != nil {
		return err
	}
	return nil
}

// GetSecuritySettings retrieves security configuration
func (s *SettingsService) GetSecuritySettings() (*models.SecuritySettings, error) {
	settings := &models.SecuritySettings{}

	var minLen, timeout, maxAttempts, lockoutDuration int
	var reqUpper, reqLower, reqNumber, reqSpecial, req2FA bool

	if err := s.GetSettingValue("password_min_length", &minLen); err == nil {
		settings.PasswordMinLength = minLen
	}
	if err := s.GetSettingValue("password_require_uppercase", &reqUpper); err == nil {
		settings.PasswordRequireUppercase = reqUpper
	}
	if err := s.GetSettingValue("password_require_lowercase", &reqLower); err == nil {
		settings.PasswordRequireLowercase = reqLower
	}
	if err := s.GetSettingValue("password_require_number", &reqNumber); err == nil {
		settings.PasswordRequireNumber = reqNumber
	}
	if err := s.GetSettingValue("password_require_special", &reqSpecial); err == nil {
		settings.PasswordRequireSpecial = reqSpecial
	}
	if err := s.GetSettingValue("session_timeout_minutes", &timeout); err == nil {
		settings.SessionTimeoutMinutes = timeout
	}
	if err := s.GetSettingValue("max_login_attempts", &maxAttempts); err == nil {
		settings.MaxLoginAttempts = maxAttempts
	}
	if err := s.GetSettingValue("lockout_duration_minutes", &lockoutDuration); err == nil {
		settings.LockoutDurationMinutes = lockoutDuration
	}
	if err := s.GetSettingValue("require_2fa_for_admins", &req2FA); err == nil {
		settings.Require2FAForAdmins = req2FA
	}

	return settings, nil
}

// UpdateSecuritySettings updates security configuration
func (s *SettingsService) UpdateSecuritySettings(settings *models.SecuritySettings) error {
	if settings.PasswordMinLength > 0 {
		if err := s.UpdateSetting("password_min_length", settings.PasswordMinLength); err != nil {
			return err
		}
	}
	if err := s.UpdateSetting("password_require_uppercase", settings.PasswordRequireUppercase); err != nil {
		return err
	}
	if err := s.UpdateSetting("password_require_lowercase", settings.PasswordRequireLowercase); err != nil {
		return err
	}
	if err := s.UpdateSetting("password_require_number", settings.PasswordRequireNumber); err != nil {
		return err
	}
	if err := s.UpdateSetting("password_require_special", settings.PasswordRequireSpecial); err != nil {
		return err
	}
	if settings.SessionTimeoutMinutes > 0 {
		if err := s.UpdateSetting("session_timeout_minutes", settings.SessionTimeoutMinutes); err != nil {
			return err
		}
	}
	if settings.MaxLoginAttempts > 0 {
		if err := s.UpdateSetting("max_login_attempts", settings.MaxLoginAttempts); err != nil {
			return err
		}
	}
	if settings.LockoutDurationMinutes > 0 {
		if err := s.UpdateSetting("lockout_duration_minutes", settings.LockoutDurationMinutes); err != nil {
			return err
		}
	}
	if err := s.UpdateSetting("require_2fa_for_admins", settings.Require2FAForAdmins); err != nil {
		return err
	}
	return nil
}

// GetSiteSettings retrieves site configuration
func (s *SettingsService) GetSiteSettings() (*models.SiteSettings, error) {
	settings := &models.SiteSettings{}

	var siteName, logoURL, maintenanceMsg string
	var maintenanceMode bool

	if err := s.GetSettingValue("site_name", &siteName); err == nil {
		settings.SiteName = siteName
	}
	if err := s.GetSettingValue("site_logo_url", &logoURL); err == nil {
		settings.SiteLogoURL = logoURL
	}
	if err := s.GetSettingValue("maintenance_mode", &maintenanceMode); err == nil {
		settings.MaintenanceMode = maintenanceMode
	}
	if err := s.GetSettingValue("maintenance_message", &maintenanceMsg); err == nil {
		settings.MaintenanceMessage = maintenanceMsg
	}

	return settings, nil
}

// UpdateSiteSettings updates site configuration
func (s *SettingsService) UpdateSiteSettings(settings *models.SiteSettings) error {
	if settings.SiteName != "" {
		if err := s.UpdateSetting("site_name", settings.SiteName); err != nil {
			return err
		}
	}
	if err := s.UpdateSetting("site_logo_url", settings.SiteLogoURL); err != nil {
		return err
	}
	if err := s.UpdateSetting("maintenance_mode", settings.MaintenanceMode); err != nil {
		return err
	}
	if settings.MaintenanceMessage != "" {
		if err := s.UpdateSetting("maintenance_message", settings.MaintenanceMessage); err != nil {
			return err
		}
	}
	return nil
}

// ============ IP Blocklist Operations ============

// GetIPBlocklist retrieves all blocked IPs
func (s *SettingsService) GetIPBlocklist() ([]models.IPBlocklist, error) {
	var blocks []models.IPBlocklist
	if err := s.db().Where("is_active = ?", true).Order("created_at DESC").Find(&blocks).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve IP blocklist: %w", err)
	}
	return blocks, nil
}

// BlockIP adds an IP to the blocklist
func (s *SettingsService) BlockIP(req *models.CreateIPBlockRequest, blockedBy uint) (*models.IPBlocklist, error) {
	block := &models.IPBlocklist{
		IPAddress: req.IPAddress,
		IPRange:   req.IPRange,
		Reason:    req.Reason,
		BlockType: "manual",
		BlockedBy: &blockedBy,
		IsActive:  true,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	if req.ExpiresAt != "" {
		block.ExpiresAt = &req.ExpiresAt
	}

	if err := s.db().Create(block).Error; err != nil {
		return nil, fmt.Errorf("failed to block IP: %w", err)
	}
	return block, nil
}

// UnblockIP removes an IP from the blocklist
func (s *SettingsService) UnblockIP(id uint) error {
	result := s.db().Model(&models.IPBlocklist{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now().Format(time.RFC3339),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to unblock IP: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("IP block not found")
	}
	return nil
}

// IsIPBlocked checks if an IP is blocked
func (s *SettingsService) IsIPBlocked(ip string) (bool, error) {
	var count int64
	err := s.db().Model(&models.IPBlocklist{}).
		Where("is_active = ? AND ip_address = ?", true, ip).
		Where("expires_at IS NULL OR expires_at > ?", time.Now().Format(time.RFC3339)).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check IP blocklist: %w", err)
	}
	return count > 0, nil
}

// ============ Announcement Banner Operations ============

// GetAnnouncements retrieves all announcements (for admin)
func (s *SettingsService) GetAnnouncements() ([]models.AnnouncementBanner, error) {
	var announcements []models.AnnouncementBanner
	if err := s.db().Order("priority DESC, created_at DESC").Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve announcements: %w", err)
	}
	return announcements, nil
}

// GetActiveAnnouncements retrieves active announcements for a user
func (s *SettingsService) GetActiveAnnouncements(userID *uint, userRole string) ([]models.AnnouncementBanner, error) {
	now := time.Now().Format(time.RFC3339)

	query := s.db().Where("is_active = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at > ?)", now).
		Order("priority DESC, created_at DESC")

	var announcements []models.AnnouncementBanner
	if err := query.Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve active announcements: %w", err)
	}

	// Filter by target roles
	var filtered []models.AnnouncementBanner
	for _, a := range announcements {
		if len(a.TargetRoles) == 0 {
			// No target roles means all users
			filtered = append(filtered, a)
		} else {
			// Check if user's role is in target roles
			for _, role := range a.TargetRoles {
				if role == userRole {
					filtered = append(filtered, a)
					break
				}
			}
		}
	}

	// If user is logged in, filter out dismissed announcements
	if userID != nil {
		var dismissedIDs []uint
		s.db().Model(&models.UserDismissedAnnouncement{}).
			Where("user_id = ?", *userID).
			Pluck("announcement_id", &dismissedIDs)

		dismissedMap := make(map[uint]bool)
		for _, id := range dismissedIDs {
			dismissedMap[id] = true
		}

		var result []models.AnnouncementBanner
		for _, a := range filtered {
			if !dismissedMap[a.ID] {
				result = append(result, a)
			}
		}
		return result, nil
	}

	return filtered, nil
}

// CreateAnnouncement creates a new announcement
func (s *SettingsService) CreateAnnouncement(req *models.CreateAnnouncementRequest, createdBy uint) (*models.AnnouncementBanner, error) {
	showOnPages, _ := json.Marshal(req.ShowOnPages)
	if len(req.ShowOnPages) == 0 {
		showOnPages = []byte(`["*"]`)
	}

	announcement := &models.AnnouncementBanner{
		Title:         req.Title,
		Message:       req.Message,
		Type:          req.Type,
		LinkURL:       req.LinkURL,
		LinkText:      req.LinkText,
		IsActive:      true,
		IsDismissible: req.IsDismissible,
		ShowOnPages:   showOnPages,
		TargetRoles:   req.TargetRoles,
		Priority:      req.Priority,
		CreatedBy:     &createdBy,
		CreatedAt:     time.Now().Format(time.RFC3339),
		UpdatedAt:     time.Now().Format(time.RFC3339),
	}

	if req.Type == "" {
		announcement.Type = "info"
	}
	if req.StartsAt != "" {
		announcement.StartsAt = &req.StartsAt
	}
	if req.EndsAt != "" {
		announcement.EndsAt = &req.EndsAt
	}

	if err := s.db().Create(announcement).Error; err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}
	return announcement, nil
}

// UpdateAnnouncement updates an existing announcement
func (s *SettingsService) UpdateAnnouncement(id uint, req *models.UpdateAnnouncementRequest) (*models.AnnouncementBanner, error) {
	var announcement models.AnnouncementBanner
	if err := s.db().First(&announcement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("announcement not found")
		}
		return nil, fmt.Errorf("failed to find announcement: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	}

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Message != nil {
		updates["message"] = *req.Message
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.LinkURL != nil {
		updates["link_url"] = *req.LinkURL
	}
	if req.LinkText != nil {
		updates["link_text"] = *req.LinkText
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsDismissible != nil {
		updates["is_dismissible"] = *req.IsDismissible
	}
	if req.ShowOnPages != nil {
		showOnPages, _ := json.Marshal(*req.ShowOnPages)
		updates["show_on_pages"] = showOnPages
	}
	if req.TargetRoles != nil {
		updates["target_roles"] = *req.TargetRoles
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.StartsAt != nil {
		updates["starts_at"] = *req.StartsAt
	}
	if req.EndsAt != nil {
		updates["ends_at"] = *req.EndsAt
	}

	if err := s.db().Model(&announcement).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}

	// Reload to get updated values
	s.db().First(&announcement, id)
	return &announcement, nil
}

// DeleteAnnouncement deletes an announcement
func (s *SettingsService) DeleteAnnouncement(id uint) error {
	result := s.db().Delete(&models.AnnouncementBanner{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete announcement: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("announcement not found")
	}
	return nil
}

// DismissAnnouncement marks an announcement as dismissed for a user
func (s *SettingsService) DismissAnnouncement(userID, announcementID uint) error {
	dismissed := &models.UserDismissedAnnouncement{
		UserID:         userID,
		AnnouncementID: announcementID,
		DismissedAt:    time.Now().Format(time.RFC3339),
	}

	// Use upsert to handle duplicate dismissals
	if err := s.db().Save(dismissed).Error; err != nil {
		return fmt.Errorf("failed to dismiss announcement: %w", err)
	}

	// Increment dismiss count
	s.db().Model(&models.AnnouncementBanner{}).
		Where("id = ?", announcementID).
		UpdateColumn("dismiss_count", gorm.Expr("dismiss_count + 1"))

	return nil
}

// ============ Email Template Operations ============

// GetEmailTemplates retrieves all email templates
func (s *SettingsService) GetEmailTemplates() ([]models.EmailTemplate, error) {
	var templates []models.EmailTemplate
	if err := s.db().Order("key").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve email templates: %w", err)
	}
	return templates, nil
}

// GetEmailTemplate retrieves a single email template by ID
func (s *SettingsService) GetEmailTemplate(id uint) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := s.db().First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("email template not found")
		}
		return nil, fmt.Errorf("failed to retrieve email template: %w", err)
	}
	return &template, nil
}

// GetEmailTemplateByKey retrieves a single email template by key
func (s *SettingsService) GetEmailTemplateByKey(key string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := s.db().Where("key = ?", key).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("email template not found: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve email template: %w", err)
	}
	return &template, nil
}

// UpdateEmailTemplate updates an email template
func (s *SettingsService) UpdateEmailTemplate(id uint, req *models.UpdateEmailTemplateRequest, updatedBy uint) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := s.db().First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("email template not found")
		}
		return nil, fmt.Errorf("failed to find email template: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
		"updated_by": updatedBy,
	}

	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	if req.BodyHTML != nil {
		updates["body_html"] = *req.BodyHTML
	}
	if req.BodyText != nil {
		updates["body_text"] = *req.BodyText
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.db().Model(&template).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update email template: %w", err)
	}

	// Reload to get updated values
	s.db().First(&template, id)
	return &template, nil
}
