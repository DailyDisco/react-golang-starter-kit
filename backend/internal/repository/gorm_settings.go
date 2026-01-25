package repository

import (
	"context"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// GormSystemSettingRepository implements SystemSettingRepository using GORM.
type GormSystemSettingRepository struct {
	db *gorm.DB
}

// NewGormSystemSettingRepository creates a new GORM-backed system setting repository.
func NewGormSystemSettingRepository(db *gorm.DB) *GormSystemSettingRepository {
	return &GormSystemSettingRepository{db: db}
}

// FindAll returns all system settings.
func (r *GormSystemSettingRepository) FindAll(ctx context.Context) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	err := r.db.WithContext(ctx).Order("category, key").Find(&settings).Error
	return settings, err
}

// FindByCategory returns settings for a specific category.
func (r *GormSystemSettingRepository) FindByCategory(ctx context.Context, category string) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	err := r.db.WithContext(ctx).Where("category = ?", category).Order("key").Find(&settings).Error
	return settings, err
}

// FindByKey returns a single setting by key.
func (r *GormSystemSettingRepository) FindByKey(ctx context.Context, key string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// FindByKeys returns multiple settings by keys.
func (r *GormSystemSettingRepository) FindByKeys(ctx context.Context, keys []string) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	err := r.db.WithContext(ctx).Where("key IN ?", keys).Find(&settings).Error
	return settings, err
}

// UpdateByKey updates a setting value by key.
func (r *GormSystemSettingRepository) UpdateByKey(ctx context.Context, key string, value []byte, updatedAt string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.SystemSetting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      value,
			"updated_at": updatedAt,
		})
	return result.RowsAffected, result.Error
}

// GormIPBlocklistRepository implements IPBlocklistRepository using GORM.
type GormIPBlocklistRepository struct {
	db *gorm.DB
}

// NewGormIPBlocklistRepository creates a new GORM-backed IP blocklist repository.
func NewGormIPBlocklistRepository(db *gorm.DB) *GormIPBlocklistRepository {
	return &GormIPBlocklistRepository{db: db}
}

// FindActive returns all active IP blocks.
func (r *GormIPBlocklistRepository) FindActive(ctx context.Context) ([]models.IPBlocklist, error) {
	var blocks []models.IPBlocklist
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at DESC").Find(&blocks).Error
	return blocks, err
}

// Create creates a new IP block entry.
func (r *GormIPBlocklistRepository) Create(ctx context.Context, block *models.IPBlocklist) error {
	return r.db.WithContext(ctx).Create(block).Error
}

// Deactivate marks an IP block as inactive.
func (r *GormIPBlocklistRepository) Deactivate(ctx context.Context, id uint, updatedAt string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.IPBlocklist{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": updatedAt,
		})
	return result.RowsAffected, result.Error
}

// IsBlocked checks if an IP is currently blocked.
func (r *GormIPBlocklistRepository) IsBlocked(ctx context.Context, ip string, now string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.IPBlocklist{}).
		Where("is_active = ? AND ip_address = ?", true, ip).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GormAnnouncementRepository implements AnnouncementRepository using GORM.
type GormAnnouncementRepository struct {
	db *gorm.DB
}

// NewGormAnnouncementRepository creates a new GORM-backed announcement repository.
func NewGormAnnouncementRepository(db *gorm.DB) *GormAnnouncementRepository {
	return &GormAnnouncementRepository{db: db}
}

// FindAll returns all announcements.
func (r *GormAnnouncementRepository) FindAll(ctx context.Context) ([]models.AnnouncementBanner, error) {
	var announcements []models.AnnouncementBanner
	err := r.db.WithContext(ctx).Order("priority DESC, created_at DESC").Find(&announcements).Error
	return announcements, err
}

// FindByID returns an announcement by ID.
func (r *GormAnnouncementRepository) FindByID(ctx context.Context, id uint) (*models.AnnouncementBanner, error) {
	var announcement models.AnnouncementBanner
	err := r.db.WithContext(ctx).First(&announcement, id).Error
	if err != nil {
		return nil, err
	}
	return &announcement, nil
}

// FindActive returns active announcements for display.
func (r *GormAnnouncementRepository) FindActive(ctx context.Context, now string) ([]models.AnnouncementBanner, error) {
	var announcements []models.AnnouncementBanner
	err := r.db.WithContext(ctx).Where("is_active = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at > ?)", now).
		Order("priority DESC, created_at DESC").
		Find(&announcements).Error
	return announcements, err
}

// Create creates a new announcement.
func (r *GormAnnouncementRepository) Create(ctx context.Context, announcement *models.AnnouncementBanner) error {
	return r.db.WithContext(ctx).Create(announcement).Error
}

// Update updates an announcement.
func (r *GormAnnouncementRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.AnnouncementBanner{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes an announcement.
func (r *GormAnnouncementRepository) Delete(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.AnnouncementBanner{}, id)
	return result.RowsAffected, result.Error
}

// IncrementDismissCount increments the dismiss count for an announcement.
func (r *GormAnnouncementRepository) IncrementDismissCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.AnnouncementBanner{}).
		Where("id = ?", id).
		UpdateColumn("dismiss_count", gorm.Expr("dismiss_count + 1")).Error
}

// IncrementViewCount increments the view count for an announcement.
func (r *GormAnnouncementRepository) IncrementViewCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.AnnouncementBanner{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// GormEmailTemplateRepository implements EmailTemplateRepository using GORM.
type GormEmailTemplateRepository struct {
	db *gorm.DB
}

// NewGormEmailTemplateRepository creates a new GORM-backed email template repository.
func NewGormEmailTemplateRepository(db *gorm.DB) *GormEmailTemplateRepository {
	return &GormEmailTemplateRepository{db: db}
}

// FindAll returns all email templates.
func (r *GormEmailTemplateRepository) FindAll(ctx context.Context) ([]models.EmailTemplate, error) {
	var templates []models.EmailTemplate
	err := r.db.WithContext(ctx).Order("key").Find(&templates).Error
	return templates, err
}

// FindByID returns an email template by ID.
func (r *GormEmailTemplateRepository) FindByID(ctx context.Context, id uint) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := r.db.WithContext(ctx).First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByKey returns an email template by key.
func (r *GormEmailTemplateRepository) FindByKey(ctx context.Context, key string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// Update updates an email template.
func (r *GormEmailTemplateRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.EmailTemplate{}).Where("id = ?", id).Updates(updates).Error
}
