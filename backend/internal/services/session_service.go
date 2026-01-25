package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/repository"
	"strings"
	"time"

	"github.com/mssola/useragent"
)

// Sentinel errors for session operations
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

// SessionService handles user session operations
type SessionService struct {
	sessionRepo repository.SessionRepository
	historyRepo repository.LoginHistoryRepository
}

// NewSessionService creates a new session service instance using global DB.
// Deprecated: Use NewSessionServiceWithRepo for better testability.
func NewSessionService() *SessionService {
	return &SessionService{
		sessionRepo: repository.NewGormSessionRepository(database.DB),
		historyRepo: repository.NewGormLoginHistoryRepository(database.DB),
	}
}

// NewSessionServiceWithRepo creates a session service with injected repositories.
// Use this constructor for testing with mock repositories.
func NewSessionServiceWithRepo(sessionRepo repository.SessionRepository, historyRepo repository.LoginHistoryRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		historyRepo: historyRepo,
	}
}

// CreateSession creates a new user session
func (s *SessionService) CreateSession(userID uint, refreshToken string, r *http.Request) (*models.UserSession, error) {
	return s.CreateSessionWithContext(r.Context(), userID, refreshToken, r)
}

// CreateSessionWithContext creates a new user session with explicit context.
func (s *SessionService) CreateSessionWithContext(ctx context.Context, userID uint, refreshToken string, r *http.Request) (*models.UserSession, error) {
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
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	session := &models.UserSession{
		UserID:           userID,
		SessionTokenHash: tokenHash,
		DeviceInfo:       deviceInfoJSON,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Location:         locationJSON,
		IsCurrent:        false,
		LastActiveAt:     now,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(userID uint, currentTokenHash string) ([]models.UserSession, error) {
	return s.GetUserSessionsWithContext(context.Background(), userID, currentTokenHash)
}

// GetUserSessionsWithContext retrieves all active sessions for a user with explicit context.
func (s *SessionService) GetUserSessionsWithContext(ctx context.Context, userID uint, currentTokenHash string) ([]models.UserSession, error) {
	now := time.Now()

	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, now)
	if err != nil {
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
	return s.RevokeSessionWithContext(context.Background(), userID, sessionID)
}

// RevokeSessionWithContext revokes a specific session with explicit context.
func (s *SessionService) RevokeSessionWithContext(ctx context.Context, userID, sessionID uint) error {
	rowsAffected, err := s.sessionRepo.DeleteByID(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// RevokeAllSessions revokes all sessions for a user except the current one
func (s *SessionService) RevokeAllSessions(userID uint, exceptTokenHash string) error {
	return s.RevokeAllSessionsWithContext(context.Background(), userID, exceptTokenHash)
}

// RevokeAllSessionsWithContext revokes all sessions for a user with explicit context.
func (s *SessionService) RevokeAllSessionsWithContext(ctx context.Context, userID uint, exceptTokenHash string) error {
	if err := s.sessionRepo.DeleteByUserID(ctx, userID, exceptTokenHash); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return nil
}

// RevokeSessionByTokenHash revokes a session by its token hash
func (s *SessionService) RevokeSessionByTokenHash(tokenHash string) error {
	return s.RevokeSessionByTokenHashWithContext(context.Background(), tokenHash)
}

// RevokeSessionByTokenHashWithContext revokes a session by its token hash with explicit context.
func (s *SessionService) RevokeSessionByTokenHashWithContext(ctx context.Context, tokenHash string) error {
	if err := s.sessionRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// UpdateLastActive updates the last active timestamp for a session
func (s *SessionService) UpdateLastActive(tokenHash string) error {
	return s.UpdateLastActiveWithContext(context.Background(), tokenHash)
}

// UpdateLastActiveWithContext updates the last active timestamp with explicit context.
func (s *SessionService) UpdateLastActiveWithContext(ctx context.Context, tokenHash string) error {
	if err := s.sessionRepo.UpdateLastActive(ctx, tokenHash, time.Now()); err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions() (int64, error) {
	return s.CleanupExpiredSessionsWithContext(context.Background())
}

// CleanupExpiredSessionsWithContext removes expired sessions with explicit context.
func (s *SessionService) CleanupExpiredSessionsWithContext(ctx context.Context) (int64, error) {
	count, err := s.sessionRepo.DeleteExpired(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup sessions: %w", err)
	}
	return count, nil
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
	return s.RecordLoginAttemptWithContext(r.Context(), userID, success, failureReason, authMethod, r, sessionID)
}

// RecordLoginAttemptWithContext records a login attempt with explicit context.
func (s *SessionService) RecordLoginAttemptWithContext(ctx context.Context, userID uint, success bool, failureReason string, authMethod string, r *http.Request, sessionID *uint) error {
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

	if err := s.historyRepo.Create(ctx, record); err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}
	return nil
}

// GetLoginHistory retrieves login history for a user
func (s *SessionService) GetLoginHistory(userID uint, limit, offset int) ([]models.LoginHistory, int64, error) {
	return s.GetLoginHistoryWithContext(context.Background(), userID, limit, offset)
}

// GetLoginHistoryWithContext retrieves login history with explicit context.
func (s *SessionService) GetLoginHistoryWithContext(ctx context.Context, userID uint, limit, offset int) ([]models.LoginHistory, int64, error) {
	// Get total count
	total, err := s.historyRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count login history: %w", err)
	}

	// Get paginated results
	history, err := s.historyRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
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
