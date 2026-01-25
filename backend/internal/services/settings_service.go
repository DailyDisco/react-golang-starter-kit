package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/repository"
	"time"

	"gorm.io/gorm"
)

// Sentinel errors for settings service
var (
	ErrSettingNotFound       = errors.New("setting not found")
	ErrIPBlockNotFound       = errors.New("IP block not found")
	ErrAnnouncementNotFound  = errors.New("announcement not found")
	ErrEmailTemplateNotFound = errors.New("email template not found")
)

// SettingsService handles system settings operations
type SettingsService struct {
	db           *gorm.DB
	settingRepo  repository.SystemSettingRepository
	ipBlockRepo  repository.IPBlocklistRepository
	announceRepo repository.AnnouncementRepository
	templateRepo repository.EmailTemplateRepository
}

// NewSettingsService creates a new settings service instance
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		db:           db,
		settingRepo:  repository.NewGormSystemSettingRepository(db),
		ipBlockRepo:  repository.NewGormIPBlocklistRepository(db),
		announceRepo: repository.NewGormAnnouncementRepository(db),
		templateRepo: repository.NewGormEmailTemplateRepository(db),
	}
}

// NewSettingsServiceWithRepo creates a settings service with injected repositories for testing.
func NewSettingsServiceWithRepo(
	db *gorm.DB,
	settingRepo repository.SystemSettingRepository,
	ipBlockRepo repository.IPBlocklistRepository,
	announceRepo repository.AnnouncementRepository,
	templateRepo repository.EmailTemplateRepository,
) *SettingsService {
	return &SettingsService{
		db:           db,
		settingRepo:  settingRepo,
		ipBlockRepo:  ipBlockRepo,
		announceRepo: announceRepo,
		templateRepo: templateRepo,
	}
}

// GetAllSettings retrieves all system settings
func (s *SettingsService) GetAllSettings(ctx context.Context) ([]models.SystemSetting, error) {
	settings, err := s.settingRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve settings: %w", err)
	}
	return settings, nil
}

// GetSettingsByCategory retrieves settings for a specific category
func (s *SettingsService) GetSettingsByCategory(ctx context.Context, category string) ([]models.SystemSetting, error) {
	settings, err := s.settingRepo.FindByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve settings for category %s: %w", category, err)
	}
	return settings, nil
}

// GetSetting retrieves a single setting by key
func (s *SettingsService) GetSetting(ctx context.Context, key string) (*models.SystemSetting, error) {
	setting, err := s.settingRepo.FindByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingNotFound
		}
		return nil, fmt.Errorf("failed to retrieve setting: %w", err)
	}
	return setting, nil
}

// GetSettingValue retrieves the value of a setting by key and unmarshals it into the provided interface
func (s *SettingsService) GetSettingValue(ctx context.Context, key string, dest interface{}) error {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(setting.Value, dest); err != nil {
		return fmt.Errorf("failed to unmarshal setting value: %w", err)
	}
	return nil
}

// GetSettingsByKeys retrieves multiple settings by keys in a single query
func (s *SettingsService) GetSettingsByKeys(ctx context.Context, keys []string) (map[string]models.SystemSetting, error) {
	settings, err := s.settingRepo.FindByKeys(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve settings: %w", err)
	}

	result := make(map[string]models.SystemSetting, len(settings))
	for _, setting := range settings {
		result[setting.Key] = setting
	}
	return result, nil
}

// UpdateSetting updates a setting value by key
func (s *SettingsService) UpdateSetting(ctx context.Context, key string, value interface{}) error {
	// Marshal the value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setting value: %w", err)
	}

	// Update the setting
	rowsAffected, err := s.settingRepo.UpdateByKey(ctx, key, jsonValue, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSettingNotFound
	}
	return nil
}

// UpdateSettingWithCache updates a setting and invalidates the settings cache
func (s *SettingsService) UpdateSettingWithCache(ctx context.Context, key string, value interface{}) error {
	if err := s.UpdateSetting(ctx, key, value); err != nil {
		return err
	}
	cache.InvalidateSettings(ctx)
	return nil
}

// UpdateSettingsBatch updates multiple settings in a single transaction
// This reduces N database round-trips to 1 for bulk setting updates
func (s *SettingsService) UpdateSettingsBatch(ctx context.Context, settings map[string]interface{}) error {
	if len(settings) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Format(time.RFC3339)

		for key, value := range settings {
			// Marshal the value to JSON
			jsonValue, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal setting %s: %w", key, err)
			}

			result := tx.Model(&models.SystemSetting{}).
				Where("key = ?", key).
				Updates(map[string]interface{}{
					"value":      jsonValue,
					"updated_at": now,
				})

			if result.Error != nil {
				return fmt.Errorf("failed to update setting %s: %w", key, result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("setting %s: %w", key, ErrSettingNotFound)
			}
		}

		return nil
	})
}

// GetEmailSettings retrieves email/SMTP configuration (single query)
func (s *SettingsService) GetEmailSettings(ctx context.Context) (*models.EmailSettings, error) {
	settings := &models.EmailSettings{}

	// Fetch all email settings in a single query
	keys := []string{"smtp_host", "smtp_port", "smtp_user", "smtp_from_email", "smtp_from_name", "smtp_enabled"}
	settingsMap, err := s.GetSettingsByKeys(ctx, keys)
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

// UpdateEmailSettings updates email/SMTP configuration (batched - 1 DB round-trip)
func (s *SettingsService) UpdateEmailSettings(ctx context.Context, settings *models.EmailSettings) error {
	updates := make(map[string]interface{})

	if settings.SMTPHost != "" {
		updates["smtp_host"] = settings.SMTPHost
	}
	if settings.SMTPPort > 0 {
		updates["smtp_port"] = settings.SMTPPort
	}
	if settings.SMTPUser != "" {
		updates["smtp_user"] = settings.SMTPUser
	}
	if settings.SMTPPassword != "" {
		updates["smtp_password"] = settings.SMTPPassword
	}
	if settings.FromEmail != "" {
		updates["smtp_from_email"] = settings.FromEmail
	}
	if settings.FromName != "" {
		updates["smtp_from_name"] = settings.FromName
	}
	updates["smtp_enabled"] = settings.Enabled

	if err := s.UpdateSettingsBatch(ctx, updates); err != nil {
		return err
	}

	cache.InvalidateSettings(ctx)
	return nil
}

// GetSecuritySettings retrieves security configuration (single query)
func (s *SettingsService) GetSecuritySettings(ctx context.Context) (*models.SecuritySettings, error) {
	settings := &models.SecuritySettings{}

	// Fetch all security settings in a single query
	keys := []string{
		"password_min_length", "password_require_uppercase", "password_require_lowercase",
		"password_require_number", "password_require_special", "session_timeout_minutes",
		"max_login_attempts", "lockout_duration_minutes", "require_2fa_for_admins",
	}
	settingsMap, err := s.GetSettingsByKeys(ctx, keys)
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
func (s *SettingsService) UpdateSecuritySettings(ctx context.Context, settings *models.SecuritySettings) error {
	updates := make(map[string]interface{})

	if settings.PasswordMinLength > 0 {
		updates["password_min_length"] = settings.PasswordMinLength
	}
	updates["password_require_uppercase"] = settings.PasswordRequireUppercase
	updates["password_require_lowercase"] = settings.PasswordRequireLowercase
	updates["password_require_number"] = settings.PasswordRequireNumber
	updates["password_require_special"] = settings.PasswordRequireSpecial
	if settings.SessionTimeoutMinutes > 0 {
		updates["session_timeout_minutes"] = settings.SessionTimeoutMinutes
	}
	if settings.MaxLoginAttempts > 0 {
		updates["max_login_attempts"] = settings.MaxLoginAttempts
	}
	if settings.LockoutDurationMinutes > 0 {
		updates["lockout_duration_minutes"] = settings.LockoutDurationMinutes
	}
	updates["require_2fa_for_admins"] = settings.Require2FAForAdmins

	if err := s.UpdateSettingsBatch(ctx, updates); err != nil {
		return err
	}

	cache.InvalidateSettings(ctx)
	return nil
}

// GetSiteSettings retrieves site configuration
func (s *SettingsService) GetSiteSettings(ctx context.Context) (*models.SiteSettings, error) {
	settings := &models.SiteSettings{}

	var siteName, logoURL, maintenanceMsg string
	var maintenanceMode bool

	if err := s.GetSettingValue(ctx, "site_name", &siteName); err == nil {
		settings.SiteName = siteName
	}
	if err := s.GetSettingValue(ctx, "site_logo_url", &logoURL); err == nil {
		settings.SiteLogoURL = logoURL
	}
	if err := s.GetSettingValue(ctx, "maintenance_mode", &maintenanceMode); err == nil {
		settings.MaintenanceMode = maintenanceMode
	}
	if err := s.GetSettingValue(ctx, "maintenance_message", &maintenanceMsg); err == nil {
		settings.MaintenanceMessage = maintenanceMsg
	}

	return settings, nil
}

// UpdateSiteSettings updates site configuration
func (s *SettingsService) UpdateSiteSettings(ctx context.Context, settings *models.SiteSettings) error {
	updates := make(map[string]interface{})

	if settings.SiteName != "" {
		updates["site_name"] = settings.SiteName
	}
	updates["site_logo_url"] = settings.SiteLogoURL
	updates["maintenance_mode"] = settings.MaintenanceMode
	if settings.MaintenanceMessage != "" {
		updates["maintenance_message"] = settings.MaintenanceMessage
	}

	if err := s.UpdateSettingsBatch(ctx, updates); err != nil {
		return err
	}

	cache.InvalidateSettings(ctx)
	return nil
}

// ============ IP Blocklist Operations ============

// GetIPBlocklist retrieves all blocked IPs
func (s *SettingsService) GetIPBlocklist(ctx context.Context) ([]models.IPBlocklist, error) {
	blocks, err := s.ipBlockRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve IP blocklist: %w", err)
	}
	return blocks, nil
}

// BlockIP adds an IP to the blocklist
func (s *SettingsService) BlockIP(ctx context.Context, req *models.CreateIPBlockRequest, blockedBy uint) (*models.IPBlocklist, error) {
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

	if err := s.ipBlockRepo.Create(ctx, block); err != nil {
		return nil, fmt.Errorf("failed to block IP: %w", err)
	}
	return block, nil
}

// UnblockIP removes an IP from the blocklist
func (s *SettingsService) UnblockIP(ctx context.Context, id uint) error {
	rowsAffected, err := s.ipBlockRepo.Deactivate(ctx, id, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to unblock IP: %w", err)
	}
	if rowsAffected == 0 {
		return ErrIPBlockNotFound
	}
	return nil
}

// IsIPBlocked checks if an IP is blocked
func (s *SettingsService) IsIPBlocked(ctx context.Context, ip string) (bool, error) {
	blocked, err := s.ipBlockRepo.IsBlocked(ctx, ip, time.Now().Format(time.RFC3339))
	if err != nil {
		return false, fmt.Errorf("failed to check IP blocklist: %w", err)
	}
	return blocked, nil
}

// ============ Announcement Banner Operations ============

// GetAnnouncements retrieves all announcements (for admin)
func (s *SettingsService) GetAnnouncements(ctx context.Context) ([]models.AnnouncementBanner, error) {
	announcements, err := s.announceRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve announcements: %w", err)
	}
	return announcements, nil
}

// GetActiveAnnouncements retrieves active announcements for a user
func (s *SettingsService) GetActiveAnnouncements(ctx context.Context, userID *uint, userRole string) ([]models.AnnouncementBanner, error) {
	now := time.Now().Format(time.RFC3339)

	query := s.db.WithContext(ctx).Where("is_active = ?", true).
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
		if err := s.db.WithContext(ctx).Model(&models.UserDismissedAnnouncement{}).
			Where("user_id = ?", *userID).
			Pluck("announcement_id", &dismissedIDs).Error; err != nil {
			return nil, fmt.Errorf("failed to get dismissed announcements: %w", err)
		}

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
func (s *SettingsService) CreateAnnouncement(ctx context.Context, req *models.CreateAnnouncementRequest, createdBy uint) (*models.AnnouncementBanner, error) {
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

	if err := s.db.WithContext(ctx).Create(announcement).Error; err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}
	cache.InvalidateAnnouncements(ctx)
	return announcement, nil
}

// UpdateAnnouncement updates an existing announcement
func (s *SettingsService) UpdateAnnouncement(ctx context.Context, id uint, req *models.UpdateAnnouncementRequest) (*models.AnnouncementBanner, error) {
	// Check if announcement exists
	_, err := s.announceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAnnouncementNotFound
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

	if err := s.announceRepo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}

	// Reload to get updated values
	announcement, err := s.announceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload announcement: %w", err)
	}
	cache.InvalidateAnnouncements(ctx)
	return announcement, nil
}

// DeleteAnnouncement deletes an announcement
func (s *SettingsService) DeleteAnnouncement(ctx context.Context, id uint) error {
	rowsAffected, err := s.announceRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAnnouncementNotFound
	}
	cache.InvalidateAnnouncements(ctx)
	return nil
}

// DismissAnnouncement marks an announcement as dismissed for a user
func (s *SettingsService) DismissAnnouncement(ctx context.Context, userID, announcementID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dismissed := &models.UserDismissedAnnouncement{
			UserID:         userID,
			AnnouncementID: announcementID,
			DismissedAt:    time.Now().Format(time.RFC3339),
		}

		// Use upsert to handle duplicate dismissals
		if err := tx.Save(dismissed).Error; err != nil {
			return fmt.Errorf("failed to dismiss announcement: %w", err)
		}

		// Increment dismiss count atomically within transaction
		if err := tx.Model(&models.AnnouncementBanner{}).
			Where("id = ?", announcementID).
			UpdateColumn("dismiss_count", gorm.Expr("dismiss_count + 1")).Error; err != nil {
			return fmt.Errorf("failed to increment dismiss count: %w", err)
		}

		return nil
	})
}

// GetChangelog retrieves paginated changelog entries (public)
func (s *SettingsService) GetChangelog(ctx context.Context, page, limit int, category string) (*models.ChangelogResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := s.db.WithContext(ctx).Model(&models.AnnouncementBanner{}).
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
func (s *SettingsService) GetUnreadModalAnnouncements(ctx context.Context, userID uint, userRole string) ([]models.AnnouncementBanner, error) {
	now := time.Now().Format(time.RFC3339)

	// Get IDs of announcements user has already read
	var readIDs []uint
	if err := s.db.WithContext(ctx).Model(&models.UserAnnouncementRead{}).
		Where("user_id = ?", userID).
		Pluck("announcement_id", &readIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get read announcements: %w", err)
	}

	query := s.db.WithContext(ctx).Where("is_active = ?", true).
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
func (s *SettingsService) MarkAnnouncementRead(ctx context.Context, userID, announcementID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		read := &models.UserAnnouncementRead{
			UserID:         userID,
			AnnouncementID: announcementID,
			ReadAt:         time.Now().Format(time.RFC3339),
		}

		// Use upsert to handle duplicate reads
		if err := tx.Save(read).Error; err != nil {
			return fmt.Errorf("failed to mark announcement as read: %w", err)
		}

		// Increment view count atomically within transaction
		if err := tx.Model(&models.AnnouncementBanner{}).
			Where("id = ?", announcementID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			return fmt.Errorf("failed to increment view count: %w", err)
		}

		return nil
	})
}

// GetUsersForAnnouncementEmail returns users who should receive announcement emails
func (s *SettingsService) GetUsersForAnnouncementEmail(ctx context.Context, announcementID uint, targetRoles []string) ([]models.User, error) {
	// Get users who:
	// 1. Have email_verified = true
	// 2. Have is_active = true
	// 3. Have updates notification preference enabled
	// 4. Match target roles (if specified)

	query := s.db.WithContext(ctx).Model(&models.User{}).
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
		if err := s.db.WithContext(ctx).Where("user_id = ?", user.ID).First(&prefs).Error; err != nil {
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
func (s *SettingsService) MarkAnnouncementEmailSent(ctx context.Context, announcementID uint) error {
	now := time.Now().Format(time.RFC3339)
	result := s.db.WithContext(ctx).Model(&models.AnnouncementBanner{}).
		Where("id = ?", announcementID).
		Updates(map[string]interface{}{
			"email_sent":    true,
			"email_sent_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark announcement email sent: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAnnouncementNotFound
	}
	return nil
}

// ============ Email Template Operations ============

// GetEmailTemplates retrieves all email templates
func (s *SettingsService) GetEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	templates, err := s.templateRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve email templates: %w", err)
	}
	return templates, nil
}

// GetEmailTemplate retrieves a single email template by ID
func (s *SettingsService) GetEmailTemplate(ctx context.Context, id uint) (*models.EmailTemplate, error) {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
		}
		return nil, fmt.Errorf("failed to retrieve email template: %w", err)
	}
	return template, nil
}

// GetEmailTemplateByKey retrieves a single email template by key
func (s *SettingsService) GetEmailTemplateByKey(ctx context.Context, key string) (*models.EmailTemplate, error) {
	template, err := s.templateRepo.FindByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
		}
		return nil, fmt.Errorf("failed to retrieve email template: %w", err)
	}
	return template, nil
}

// UpdateEmailTemplate updates an email template
func (s *SettingsService) UpdateEmailTemplate(ctx context.Context, id uint, req *models.UpdateEmailTemplateRequest, updatedBy uint) (*models.EmailTemplate, error) {
	// Check if template exists
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
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

	if err := s.templateRepo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update email template: %w", err)
	}

	// Reload to get updated values
	template, err = s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload email template: %w", err)
	}

	// Invalidate cache for this template
	cache.InvalidateEmailTemplate(ctx, template.Key)

	return template, nil
}
