package repository

import (
	"context"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// GormUsageEventRepository implements UsageEventRepository using GORM.
type GormUsageEventRepository struct {
	db *gorm.DB
}

// NewGormUsageEventRepository creates a new GORM-backed usage event repository.
func NewGormUsageEventRepository(db *gorm.DB) *GormUsageEventRepository {
	return &GormUsageEventRepository{db: db}
}

// Create creates a new usage event.
func (r *GormUsageEventRepository) Create(ctx context.Context, event *models.UsageEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GormUsagePeriodRepository implements UsagePeriodRepository using GORM.
type GormUsagePeriodRepository struct {
	db *gorm.DB
}

// NewGormUsagePeriodRepository creates a new GORM-backed usage period repository.
func NewGormUsagePeriodRepository(db *gorm.DB) *GormUsagePeriodRepository {
	return &GormUsagePeriodRepository{db: db}
}

// FindByUserAndPeriod finds a usage period by user ID and period dates.
func (r *GormUsagePeriodRepository) FindByUserAndPeriod(ctx context.Context, userID uint, periodStart, periodEnd string) (*models.UsagePeriod, error) {
	var period models.UsagePeriod
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND period_start = ? AND period_end = ?", userID, periodStart, periodEnd).
		First(&period).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

// FindByOrgAndPeriod finds a usage period by organization ID and period dates.
func (r *GormUsagePeriodRepository) FindByOrgAndPeriod(ctx context.Context, orgID uint, periodStart, periodEnd string) (*models.UsagePeriod, error) {
	var period models.UsagePeriod
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND period_start = ? AND period_end = ?", orgID, periodStart, periodEnd).
		First(&period).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

// FindHistoryByUser returns usage history for a user.
func (r *GormUsagePeriodRepository) FindHistoryByUser(ctx context.Context, userID uint, limit int) ([]models.UsagePeriod, error) {
	var periods []models.UsagePeriod
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("period_start DESC").
		Limit(limit).
		Find(&periods).Error
	return periods, err
}

// FindHistoryByOrg returns usage history for an organization.
func (r *GormUsagePeriodRepository) FindHistoryByOrg(ctx context.Context, orgID uint, limit int) ([]models.UsagePeriod, error) {
	var periods []models.UsagePeriod
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("period_start DESC").
		Limit(limit).
		Find(&periods).Error
	return periods, err
}

// Create creates a new usage period.
func (r *GormUsagePeriodRepository) Create(ctx context.Context, period *models.UsagePeriod) error {
	return r.db.WithContext(ctx).Create(period).Error
}

// Update updates a usage period.
func (r *GormUsagePeriodRepository) Update(ctx context.Context, period *models.UsagePeriod, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(period).Updates(updates).Error
}

// Upsert creates or updates a usage period.
func (r *GormUsagePeriodRepository) Upsert(ctx context.Context, period *models.UsagePeriod) error {
	return r.db.WithContext(ctx).Save(period).Error
}

// GormUsageAlertRepository implements UsageAlertRepository using GORM.
type GormUsageAlertRepository struct {
	db *gorm.DB
}

// NewGormUsageAlertRepository creates a new GORM-backed usage alert repository.
func NewGormUsageAlertRepository(db *gorm.DB) *GormUsageAlertRepository {
	return &GormUsageAlertRepository{db: db}
}

// FindUnacknowledgedByUser returns unacknowledged alerts for a user.
func (r *GormUsageAlertRepository) FindUnacknowledgedByUser(ctx context.Context, userID uint) ([]models.UsageAlert, error) {
	var alerts []models.UsageAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND acknowledged = ?", userID, false).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// FindUnacknowledgedByOrg returns unacknowledged alerts for an organization.
func (r *GormUsageAlertRepository) FindUnacknowledgedByOrg(ctx context.Context, orgID uint) ([]models.UsageAlert, error) {
	var alerts []models.UsageAlert
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND acknowledged = ?", orgID, false).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// FindOrCreate finds an existing alert or creates a new one.
// Returns true if a new alert was created.
func (r *GormUsageAlertRepository) FindOrCreate(ctx context.Context, alert *models.UsageAlert) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND alert_type = ? AND usage_type = ? AND period_start = ?",
			alert.UserID, alert.AlertType, alert.UsageType, alert.PeriodStart).
		FirstOrCreate(alert)
	return result.RowsAffected > 0, result.Error
}

// Acknowledge marks an alert as acknowledged.
func (r *GormUsageAlertRepository) Acknowledge(ctx context.Context, alertID uint, acknowledgedBy uint, acknowledgedAt string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.UsageAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": acknowledgedAt,
			"acknowledged_by": acknowledgedBy,
		})
	return result.RowsAffected, result.Error
}
