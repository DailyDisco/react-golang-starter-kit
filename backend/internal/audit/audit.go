package audit

import (
	"encoding/json"
	"net/http"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

// LogEntry creates an audit log entry
func LogEntry(userID *uint, targetType string, targetID *uint, action string, changes interface{}, r *http.Request) {
	go logEntryAsync(userID, targetType, targetID, action, changes, r)
}

// logEntryAsync performs the actual logging asynchronously
func logEntryAsync(userID *uint, targetType string, targetID *uint, action string, changes interface{}, r *http.Request) {
	entry := models.AuditLog{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Action:     action,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// Serialize changes if provided
	if changes != nil {
		changesJSON, err := json.Marshal(changes)
		if err == nil {
			entry.Changes = string(changesJSON)
		}
	}

	// Extract request info if available
	if r != nil {
		entry.IPAddress = getClientIP(r)
		entry.UserAgent = r.UserAgent()
	}

	// Save to database
	if err := database.DB.Create(&entry).Error; err != nil {
		log.Error().Err(err).
			Interface("user_id", userID).
			Str("target_type", targetType).
			Str("action", action).
			Msg("Failed to create audit log entry")
	}
}

// LogWithMetadata creates an audit log entry with additional metadata
func LogWithMetadata(userID *uint, targetType string, targetID *uint, action string, changes interface{}, metadata interface{}, r *http.Request) {
	go logWithMetadataAsync(userID, targetType, targetID, action, changes, metadata, r)
}

func logWithMetadataAsync(userID *uint, targetType string, targetID *uint, action string, changes interface{}, metadata interface{}, r *http.Request) {
	entry := models.AuditLog{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Action:     action,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	// Serialize changes if provided
	if changes != nil {
		changesJSON, err := json.Marshal(changes)
		if err == nil {
			entry.Changes = string(changesJSON)
		}
	}

	// Serialize metadata if provided
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err == nil {
			entry.Metadata = string(metadataJSON)
		}
	}

	// Extract request info if available
	if r != nil {
		entry.IPAddress = getClientIP(r)
		entry.UserAgent = r.UserAgent()
	}

	// Save to database
	if err := database.DB.Create(&entry).Error; err != nil {
		log.Error().Err(err).
			Interface("user_id", userID).
			Str("target_type", targetType).
			Str("action", action).
			Msg("Failed to create audit log entry")
	}
}

// LogLogin creates an audit log entry for login
func LogLogin(userID uint, r *http.Request, metadata map[string]interface{}) {
	LogWithMetadata(&userID, models.AuditTargetUser, &userID, models.AuditActionLogin, nil, metadata, r)
}

// LogLogout creates an audit log entry for logout
func LogLogout(userID uint, r *http.Request) {
	LogEntry(&userID, models.AuditTargetUser, &userID, models.AuditActionLogout, nil, r)
}

// LogImpersonate creates an audit log entry for impersonation
func LogImpersonate(adminUserID uint, targetUserID uint, reason string, r *http.Request) {
	metadata := map[string]interface{}{
		"reason": reason,
	}
	LogWithMetadata(&adminUserID, models.AuditTargetUser, &targetUserID, models.AuditActionImpersonate, nil, metadata, r)
}

// LogStopImpersonate creates an audit log entry for stopping impersonation
func LogStopImpersonate(adminUserID uint, targetUserID uint, r *http.Request) {
	LogEntry(&adminUserID, models.AuditTargetUser, &targetUserID, models.AuditActionStopImpersonate, nil, r)
}

// LogUserCreate creates an audit log entry for user creation
func LogUserCreate(actorUserID *uint, newUserID uint, r *http.Request) {
	LogEntry(actorUserID, models.AuditTargetUser, &newUserID, models.AuditActionCreate, nil, r)
}

// LogUserUpdate creates an audit log entry for user update
func LogUserUpdate(actorUserID uint, targetUserID uint, changes map[string]interface{}, r *http.Request) {
	LogEntry(&actorUserID, models.AuditTargetUser, &targetUserID, models.AuditActionUpdate, changes, r)
}

// LogUserDelete creates an audit log entry for user deletion
func LogUserDelete(actorUserID uint, targetUserID uint, r *http.Request) {
	LogEntry(&actorUserID, models.AuditTargetUser, &targetUserID, models.AuditActionDelete, nil, r)
}

// LogRoleChange creates an audit log entry for role change
func LogRoleChange(actorUserID uint, targetUserID uint, oldRole string, newRole string, r *http.Request) {
	changes := map[string]interface{}{
		"old_role": oldRole,
		"new_role": newRole,
	}
	LogEntry(&actorUserID, models.AuditTargetUser, &targetUserID, models.AuditActionRoleChange, changes, r)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// GetAuditLogs retrieves audit logs with filtering and pagination
func GetAuditLogs(filter models.AuditLogFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.DB.Model(&models.AuditLog{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.TargetType != "" {
		query = query.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != nil {
		query = query.Where("target_id = ?", *filter.TargetID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	// Fetch logs with user preloaded
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetUserAuditLogs retrieves audit logs for a specific user
func GetUserAuditLogs(userID uint, page, limit int) ([]models.AuditLog, int64, error) {
	filter := models.AuditLogFilter{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	}
	return GetAuditLogs(filter)
}
