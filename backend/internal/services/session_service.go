package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"strings"
	"time"

	"github.com/mssola/useragent"
)

// SessionService handles user session operations
type SessionService struct{}

// NewSessionService creates a new session service instance
func NewSessionService() *SessionService {
	return &SessionService{}
}

// CreateSession creates a new user session
func (s *SessionService) CreateSession(userID uint, refreshToken string, r *http.Request) (*models.UserSession, error) {
	// Hash the refresh token for secure storage
	tokenHash := hashToken(refreshToken)

	// Parse device info from user agent
	userAgent := r.UserAgent()
	deviceInfo := s.ParseDeviceInfo(userAgent)
	deviceInfoJSON, _ := json.Marshal(deviceInfo)

	// Get IP address
	ipAddress := getClientIP(r)

	// Get location from IP (placeholder - would use a geolocation service in production)
	location := s.GetLocationFromIP(ipAddress)
	locationJSON, _ := json.Marshal(location)

	// Calculate expiration (default 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	session := &models.UserSession{
		UserID:           userID,
		SessionTokenHash: tokenHash,
		DeviceInfo:       deviceInfoJSON,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Location:         locationJSON,
		IsCurrent:        false,
		LastActiveAt:     time.Now(),
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	}

	if err := database.DB.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(userID uint, currentTokenHash string) ([]models.UserSession, error) {
	var sessions []models.UserSession
	now := time.Now()

	if err := database.DB.Where("user_id = ? AND expires_at > ?", userID, now).
		Order("last_active_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	// Mark current session
	for i := range sessions {
		if sessions[i].SessionTokenHash == currentTokenHash {
			sessions[i].IsCurrent = true
		}
	}

	return sessions, nil
}

// RevokeSession revokes a specific session
func (s *SessionService) RevokeSession(userID, sessionID uint) error {
	result := database.DB.Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.UserSession{})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// RevokeAllSessions revokes all sessions for a user except the current one
func (s *SessionService) RevokeAllSessions(userID uint, exceptTokenHash string) error {
	query := database.DB.Where("user_id = ?", userID)
	if exceptTokenHash != "" {
		query = query.Where("session_token_hash != ?", exceptTokenHash)
	}

	if err := query.Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return nil
}

// RevokeSessionByTokenHash revokes a session by its token hash
func (s *SessionService) RevokeSessionByTokenHash(tokenHash string) error {
	result := database.DB.Where("session_token_hash = ?", tokenHash).
		Delete(&models.UserSession{})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke session: %w", result.Error)
	}
	return nil
}

// UpdateLastActive updates the last active timestamp for a session
func (s *SessionService) UpdateLastActive(tokenHash string) error {
	result := database.DB.Model(&models.UserSession{}).
		Where("session_token_hash = ?", tokenHash).
		Update("last_active_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to update last active: %w", result.Error)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions() (int64, error) {
	now := time.Now()
	result := database.DB.Where("expires_at < ?", now).Delete(&models.UserSession{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup sessions: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ParseDeviceInfo parses user agent string to extract device information
func (s *SessionService) ParseDeviceInfo(userAgentStr string) models.DeviceInfo {
	ua := useragent.New(userAgentStr)

	browserName, browserVersion := ua.Browser()
	osInfo := ua.OS()

	deviceType := "desktop"
	if ua.Mobile() {
		deviceType = "mobile"
	} else if strings.Contains(strings.ToLower(userAgentStr), "tablet") ||
		strings.Contains(strings.ToLower(userAgentStr), "ipad") {
		deviceType = "tablet"
	}

	return models.DeviceInfo{
		Browser:        browserName,
		BrowserVersion: browserVersion,
		OS:             osInfo,
		OSVersion:      "", // useragent package doesn't provide OS version separately
		DeviceType:     deviceType,
		DeviceName:     ua.Platform(),
	}
}

// GetLocationFromIP attempts to get location from IP address
// In production, this would use a geolocation service like MaxMind or ip-api.com
func (s *SessionService) GetLocationFromIP(ip string) models.LocationInfo {
	// Placeholder implementation
	// In production, integrate with a geolocation API
	if ip == "127.0.0.1" || ip == "::1" {
		return models.LocationInfo{
			Country:     "Local",
			CountryCode: "LO",
			City:        "Localhost",
			Region:      "Development",
		}
	}

	// Default unknown location
	return models.LocationInfo{
		Country:     "Unknown",
		CountryCode: "XX",
		City:        "Unknown",
		Region:      "Unknown",
	}
}

// ============ Login History Operations ============

// RecordLoginAttempt records a login attempt (success or failure)
func (s *SessionService) RecordLoginAttempt(userID uint, success bool, failureReason string, authMethod string, r *http.Request, sessionID *uint) error {
	userAgent := r.UserAgent()
	deviceInfo := s.ParseDeviceInfo(userAgent)
	deviceInfoJSON, _ := json.Marshal(deviceInfo)

	ipAddress := getClientIP(r)
	location := s.GetLocationFromIP(ipAddress)
	locationJSON, _ := json.Marshal(location)

	if authMethod == "" {
		authMethod = models.AuthMethodPassword
	}

	record := &models.LoginHistory{
		UserID:        userID,
		Success:       success,
		FailureReason: failureReason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		DeviceInfo:    deviceInfoJSON,
		Location:      locationJSON,
		AuthMethod:    authMethod,
		SessionID:     sessionID,
		CreatedAt:     time.Now(),
	}

	if err := database.DB.Create(record).Error; err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}
	return nil
}

// GetLoginHistory retrieves login history for a user
func (s *SessionService) GetLoginHistory(userID uint, limit, offset int) ([]models.LoginHistory, int64, error) {
	var history []models.LoginHistory
	var total int64

	// Get total count
	if err := database.DB.Model(&models.LoginHistory{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count login history: %w", err)
	}

	// Get paginated results
	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve login history: %w", err)
	}

	return history, total, nil
}

// Helper functions

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}
