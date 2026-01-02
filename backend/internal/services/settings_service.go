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

// GetSettingsByKeys retrieves multiple settings by keys in a single query
func (s *SettingsService) GetSettingsByKeys(keys []string) (map[string]models.SystemSetting, error) {
	var settings []models.SystemSetting
	if err := s.db().Where("key IN ?", keys).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve settings: %w", err)
	}

	result := make(map[string]models.SystemSetting, len(settings))
	for _, setting := range settings {
		result[setting.Key] = setting
	}
	return result, nil
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

// GetEmailSettings retrieves email/SMTP configuration (single query)
func (s *SettingsService) GetEmailSettings() (*models.EmailSettings, error) {
	settings := &models.EmailSettings{}

	// Fetch all email settings in a single query
	keys := []string{"smtp_host", "smtp_port", "smtp_user", "smtp_from_email", "smtp_from_name", "smtp_enabled"}
	settingsMap, err := s.GetSettingsByKeys(keys)
	if err != nil {
		return nil, err
	}

	// Unmarshal each setting if present
	if setting, ok := settingsMap["smtp_host"]; ok {
		var val string
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.SMTPHost = val
		}
	}
	if setting, ok := settingsMap["smtp_port"]; ok {
		var val int
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.SMTPPort = val
		}
	}
	if setting, ok := settingsMap["smtp_user"]; ok {
		var val string
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.SMTPUser = val
		}
	}
	if setting, ok := settingsMap["smtp_from_email"]; ok {
		var val string
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.FromEmail = val
		}
	}
	if setting, ok := settingsMap["smtp_from_name"]; ok {
		var val string
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.FromName = val
		}
	}
	if setting, ok := settingsMap["smtp_enabled"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.Enabled = val
		}
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

// GetSecuritySettings retrieves security configuration (single query)
func (s *SettingsService) GetSecuritySettings() (*models.SecuritySettings, error) {
	settings := &models.SecuritySettings{}

	// Fetch all security settings in a single query
	keys := []string{
		"password_min_length", "password_require_uppercase", "password_require_lowercase",
		"password_require_number", "password_require_special", "session_timeout_minutes",
		"max_login_attempts", "lockout_duration_minutes", "require_2fa_for_admins",
	}
	settingsMap, err := s.GetSettingsByKeys(keys)
	if err != nil {
		return nil, err
	}

	// Unmarshal each setting if present
	if setting, ok := settingsMap["password_min_length"]; ok {
		var val int
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.PasswordMinLength = val
		}
	}
	if setting, ok := settingsMap["password_require_uppercase"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.PasswordRequireUppercase = val
		}
	}
	if setting, ok := settingsMap["password_require_lowercase"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.PasswordRequireLowercase = val
		}
	}
	if setting, ok := settingsMap["password_require_number"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.PasswordRequireNumber = val
		}
	}
	if setting, ok := settingsMap["password_require_special"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.PasswordRequireSpecial = val
		}
	}
	if setting, ok := settingsMap["session_timeout_minutes"]; ok {
		var val int
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.SessionTimeoutMinutes = val
		}
	}
	if setting, ok := settingsMap["max_login_attempts"]; ok {
		var val int
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.MaxLoginAttempts = val
		}
	}
	if setting, ok := settingsMap["lockout_duration_minutes"]; ok {
		var val int
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.LockoutDurationMinutes = val
		}
	}
	if setting, ok := settingsMap["require_2fa_for_admins"]; ok {
		var val bool
		if json.Unmarshal(setting.Value, &val) == nil {
			settings.Require2FAForAdmins = val
		}
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

	now := time.Now().Format(time.RFC3339)
	announcement := &models.AnnouncementBanner{
		Title:         req.Title,
		Message:       req.Message,
		Type:          req.Type,
		DisplayType:   req.DisplayType,
		Category:      req.Category,
		LinkURL:       req.LinkURL,
		LinkText:      req.LinkText,
		IsActive:      req.IsActive,
		IsDismissible: req.IsDismissible,
		ShowOnPages:   showOnPages,
		TargetRoles:   req.TargetRoles,
		Priority:      req.Priority,
		PublishedAt:   &now,
		CreatedBy:     &createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if req.Type == "" {
		announcement.Type = "info"
	}
	if req.DisplayType == "" {
		announcement.DisplayType = "banner"
	}
	if req.Category == "" {
		announcement.Category = "update"
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
	if req.DisplayType != nil {
		updates["display_type"] = *req.DisplayType
	}
	if req.Category != nil {
		updates["category"] = *req.Category
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

// GetChangelog retrieves paginated changelog entries (public)
func (s *SettingsService) GetChangelog(page, limit int, category string) (*models.ChangelogResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := s.db().Model(&models.AnnouncementBanner{}).
		Where("is_active = ?", true).
		Where("published_at IS NOT NULL")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count changelog entries: %w", err)
	}

	var announcements []models.AnnouncementBanner
	if err := query.Order("published_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve changelog: %w", err)
	}

	// Convert to response format
	data := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		data[i] = a.ToResponse()
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &models.ChangelogResponse{
		Data: data,
		Meta: models.ChangelogMeta{
			Page:       page,
			PerPage:    limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// GetUnreadModalAnnouncements retrieves modal announcements user hasn't seen
func (s *SettingsService) GetUnreadModalAnnouncements(userID uint, userRole string) ([]models.AnnouncementBanner, error) {
	now := time.Now().Format(time.RFC3339)

	// Get IDs of announcements user has already read
	var readIDs []uint
	s.db().Model(&models.UserAnnouncementRead{}).
		Where("user_id = ?", userID).
		Pluck("announcement_id", &readIDs)

	query := s.db().Where("is_active = ?", true).
		Where("display_type = ?", "modal").
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at > ?)", now)

	if len(readIDs) > 0 {
		query = query.Where("id NOT IN ?", readIDs)
	}

	var announcements []models.AnnouncementBanner
	if err := query.Order("priority DESC, published_at DESC").Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve unread modal announcements: %w", err)
	}

	// Filter by target roles
	var filtered []models.AnnouncementBanner
	for _, a := range announcements {
		if len(a.TargetRoles) == 0 {
			filtered = append(filtered, a)
		} else {
			for _, role := range a.TargetRoles {
				if role == userRole {
					filtered = append(filtered, a)
					break
				}
			}
		}
	}

	return filtered, nil
}

// MarkAnnouncementRead records that a user has viewed a modal announcement
func (s *SettingsService) MarkAnnouncementRead(userID, announcementID uint) error {
	read := &models.UserAnnouncementRead{
		UserID:         userID,
		AnnouncementID: announcementID,
		ReadAt:         time.Now().Format(time.RFC3339),
	}

	// Use upsert to handle duplicate reads
	if err := s.db().Save(read).Error; err != nil {
		return fmt.Errorf("failed to mark announcement as read: %w", err)
	}

	// Increment view count
	s.db().Model(&models.AnnouncementBanner{}).
		Where("id = ?", announcementID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	return nil
}

// GetUsersForAnnouncementEmail returns users who should receive announcement emails
func (s *SettingsService) GetUsersForAnnouncementEmail(announcementID uint, targetRoles []string) ([]models.User, error) {
	// Get users who:
	// 1. Have email_verified = true
	// 2. Have is_active = true
	// 3. Have updates notification preference enabled
	// 4. Match target roles (if specified)

	query := s.db().Model(&models.User{}).
		Where("email_verified = ?", true).
		Where("is_active = ?", true)

	if len(targetRoles) > 0 {
		query = query.Where("role IN ?", targetRoles)
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users for announcement email: %w", err)
	}

	// Filter by email notification preferences
	var eligibleUsers []models.User
	for _, user := range users {
		// Get user preferences
		var prefs models.UserPreferences
		if err := s.db().Where("user_id = ?", user.ID).First(&prefs).Error; err != nil {
			// No preferences found, skip this user (conservative approach)
			continue
		}

		// Parse email notifications
		var emailNotifs models.EmailNotificationSettings
		if prefs.EmailNotifications != nil {
			if err := json.Unmarshal(prefs.EmailNotifications, &emailNotifs); err != nil {
				continue
			}
		}

		// Check if updates notifications are enabled
		if emailNotifs.Updates {
			eligibleUsers = append(eligibleUsers, user)
		}
	}

	return eligibleUsers, nil
}

// MarkAnnouncementEmailSent marks an announcement as having had emails sent
func (s *SettingsService) MarkAnnouncementEmailSent(announcementID uint) error {
	now := time.Now().Format(time.RFC3339)
	result := s.db().Model(&models.AnnouncementBanner{}).
		Where("id = ?", announcementID).
		Updates(map[string]interface{}{
			"email_sent":    true,
			"email_sent_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark announcement email sent: %w", result.Error)
	}
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
