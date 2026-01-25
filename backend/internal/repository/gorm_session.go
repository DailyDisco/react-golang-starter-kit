package repository

import (
	"context"
	"time"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// GormSessionRepository implements SessionRepository using GORM.
type GormSessionRepository struct {
	db *gorm.DB
}

// NewGormSessionRepository creates a new GORM-backed session repository.
func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

// Create creates a new session record.
func (r *GormSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByUserID returns all sessions for a user that haven't expired.
func (r *GormSessionRepository) FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Order("last_active_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// DeleteByID deletes a session by ID and user ID.
func (r *GormSessionRepository) DeleteByID(ctx context.Context, sessionID, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.UserSession{})
	return result.RowsAffected, result.Error
}

// DeleteByUserID deletes all sessions for a user, optionally excluding a token hash.
func (r *GormSessionRepository) DeleteByUserID(ctx context.Context, userID uint, exceptTokenHash string) error {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if exceptTokenHash != "" {
		query = query.Where("session_token_hash != ?", exceptTokenHash)
	}
	return query.Delete(&models.UserSession{}).Error
}

// DeleteByTokenHash deletes a session by its token hash.
func (r *GormSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Where("session_token_hash = ?", tokenHash).
		Delete(&models.UserSession{}).Error
}

// UpdateLastActive updates the last_active_at timestamp for a session.
func (r *GormSessionRepository) UpdateLastActive(ctx context.Context, tokenHash string, lastActive time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("session_token_hash = ?", tokenHash).
		Update("last_active_at", lastActive).Error
}

// DeleteExpired removes all sessions that have expired before the given time.
func (r *GormSessionRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&models.UserSession{})
	return result.RowsAffected, result.Error
}

// GormLoginHistoryRepository implements LoginHistoryRepository using GORM.
type GormLoginHistoryRepository struct {
	db *gorm.DB
}

// NewGormLoginHistoryRepository creates a new GORM-backed login history repository.
func NewGormLoginHistoryRepository(db *gorm.DB) *GormLoginHistoryRepository {
	return &GormLoginHistoryRepository{db: db}
}

// Create records a login attempt.
func (r *GormLoginHistoryRepository) Create(ctx context.Context, record *models.LoginHistory) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// FindByUserID returns login history for a user with pagination.
func (r *GormLoginHistoryRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.LoginHistory, error) {
	var history []models.LoginHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error
	return history, err
}

// CountByUserID returns the total number of login records for a user.
func (r *GormLoginHistoryRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.LoginHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
