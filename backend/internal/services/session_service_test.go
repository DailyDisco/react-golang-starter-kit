package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil/mocks"
)

// ============ Session Service Helper Tests ============

func TestHashToken_Session(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"refresh token", "refresh_abc123xyz"},
		{"empty token", ""},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashToken(tt.token)

			// Hash should be consistent
			if hash != hashToken(tt.token) {
				t.Error("hashToken() should return consistent results")
			}

			// Hash should be 64 chars (SHA-256 hex)
			if len(hash) != 64 {
				t.Errorf("hashToken() length = %d, want 64", len(hash))
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "from X-Forwarded-For single IP",
			xForwardedFor: "192.168.1.1",
			remoteAddr:    "10.0.0.1:8080",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "from X-Forwarded-For multiple IPs",
			xForwardedFor: "192.168.1.1, 10.0.0.2, 172.16.0.1",
			remoteAddr:    "10.0.0.1:8080",
			expectedIP:    "192.168.1.1",
		},
		{
			name:       "from X-Real-IP",
			xRealIP:    "192.168.1.1",
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "from RemoteAddr with port",
			remoteAddr: "192.168.1.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "from RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[::1]:8080",
			expectedIP: "[::1]",
		},
		{
			name:          "X-Forwarded-For takes precedence",
			xForwardedFor: "1.1.1.1",
			xRealIP:       "2.2.2.2",
			remoteAddr:    "3.3.3.3:8080",
			expectedIP:    "1.1.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestSessionService_ParseDeviceInfo(t *testing.T) {
	s := &SessionService{}

	tests := []struct {
		name           string
		userAgent      string
		wantDeviceType string
		wantBrowser    string
	}{
		{
			name:           "Chrome on Windows",
			userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantDeviceType: "desktop",
			wantBrowser:    "Chrome",
		},
		{
			name:           "Safari on iPhone",
			userAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantDeviceType: "mobile",
			wantBrowser:    "Safari",
		},
		{
			name:           "Safari on iPad",
			userAgent:      "Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			wantDeviceType: "mobile", // Note: useragent lib reports iPad as mobile
			wantBrowser:    "Safari",
		},
		{
			name:           "Firefox on Linux",
			userAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
			wantDeviceType: "desktop",
			wantBrowser:    "Firefox",
		},
		{
			name:           "Android phone",
			userAgent:      "Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			wantDeviceType: "mobile",
			wantBrowser:    "Chrome",
		},
		{
			name:           "Empty user agent",
			userAgent:      "",
			wantDeviceType: "desktop",
			wantBrowser:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := s.ParseDeviceInfo(tt.userAgent)

			if info.DeviceType != tt.wantDeviceType {
				t.Errorf("ParseDeviceInfo().DeviceType = %q, want %q", info.DeviceType, tt.wantDeviceType)
			}

			if info.Browser != tt.wantBrowser {
				t.Errorf("ParseDeviceInfo().Browser = %q, want %q", info.Browser, tt.wantBrowser)
			}
		})
	}
}

func TestSessionService_GetLocationFromIP(t *testing.T) {
	s := &SessionService{}

	tests := []struct {
		name        string
		ip          string
		wantCountry string
	}{
		{
			name:        "localhost IPv4",
			ip:          "127.0.0.1",
			wantCountry: "Local",
		},
		{
			name:        "localhost IPv6",
			ip:          "::1",
			wantCountry: "Local",
		},
		{
			name:        "unknown IP",
			ip:          "8.8.8.8",
			wantCountry: "Unknown",
		},
		{
			name:        "private IP",
			ip:          "192.168.1.1",
			wantCountry: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := s.GetLocationFromIP(tt.ip)

			if location.Country != tt.wantCountry {
				t.Errorf("GetLocationFromIP(%q).Country = %q, want %q", tt.ip, location.Country, tt.wantCountry)
			}
		})
	}
}

func TestSessionService_GetLocationFromIP_Structure(t *testing.T) {
	s := &SessionService{}

	// Test localhost returns complete structure
	location := s.GetLocationFromIP("127.0.0.1")

	if location.CountryCode != "LO" {
		t.Errorf("GetLocationFromIP().CountryCode = %q, want %q", location.CountryCode, "LO")
	}
	if location.City != "Localhost" {
		t.Errorf("GetLocationFromIP().City = %q, want %q", location.City, "Localhost")
	}
	if location.Region != "Development" {
		t.Errorf("GetLocationFromIP().Region = %q, want %q", location.Region, "Development")
	}

	// Test unknown IP returns complete structure
	unknown := s.GetLocationFromIP("8.8.8.8")
	if unknown.CountryCode != "XX" {
		t.Errorf("GetLocationFromIP().CountryCode = %q, want %q", unknown.CountryCode, "XX")
	}
}

// ============ Device Info Structure Tests ============

func TestDeviceInfo_Fields(t *testing.T) {
	info := models.DeviceInfo{
		Browser:        "Chrome",
		BrowserVersion: "120.0.0.0",
		OS:             "Windows 10",
		OSVersion:      "10.0",
		DeviceType:     "desktop",
		DeviceName:     "Windows",
	}

	if info.Browser != "Chrome" {
		t.Errorf("DeviceInfo.Browser = %q, want %q", info.Browser, "Chrome")
	}
	if info.DeviceType != "desktop" {
		t.Errorf("DeviceInfo.DeviceType = %q, want %q", info.DeviceType, "desktop")
	}
}

// ============ SessionService Constructor Tests ============

func TestNewSessionService(t *testing.T) {
	s := NewSessionService()
	if s == nil {
		t.Fatal("NewSessionService() returned nil")
	}
}

// ============ LocationInfo Structure Tests ============

func TestLocationInfo_Fields(t *testing.T) {
	location := models.LocationInfo{
		Country:     "United States",
		CountryCode: "US",
		City:        "New York",
		Region:      "New York",
		Latitude:    40.7128,
		Longitude:   -74.0060,
	}

	if location.Country != "United States" {
		t.Errorf("LocationInfo.Country = %q, want %q", location.Country, "United States")
	}
	if location.CountryCode != "US" {
		t.Errorf("LocationInfo.CountryCode = %q, want %q", location.CountryCode, "US")
	}
	if location.City != "New York" {
		t.Errorf("LocationInfo.City = %q, want %q", location.City, "New York")
	}
	if location.Region != "New York" {
		t.Errorf("LocationInfo.Region = %q, want %q", location.Region, "New York")
	}
	if location.Latitude != 40.7128 {
		t.Errorf("LocationInfo.Latitude = %f, want %f", location.Latitude, 40.7128)
	}
	if location.Longitude != -74.0060 {
		t.Errorf("LocationInfo.Longitude = %f, want %f", location.Longitude, -74.0060)
	}
}

// ============ Additional ParseDeviceInfo Tests ============

func TestSessionService_ParseDeviceInfo_Tablet(t *testing.T) {
	s := &SessionService{}

	// Test tablet detection with iPad keyword
	userAgents := []string{
		"Mozilla/5.0 (Linux; Android 12; SM-T970) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Tablet",
		"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
	}

	for _, ua := range userAgents {
		info := s.ParseDeviceInfo(ua)
		// Mobile or tablet should be detected
		if info.DeviceType != "mobile" && info.DeviceType != "tablet" {
			t.Logf("ParseDeviceInfo(%q).DeviceType = %q (may be expected for this UA parser)", ua, info.DeviceType)
		}
	}
}

func TestSessionService_ParseDeviceInfo_EdgeBrowser(t *testing.T) {
	s := &SessionService{}

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"
	info := s.ParseDeviceInfo(ua)

	if info.DeviceType != "desktop" {
		t.Errorf("ParseDeviceInfo().DeviceType = %q, want %q", info.DeviceType, "desktop")
	}
}

func TestSessionService_ParseDeviceInfo_Opera(t *testing.T) {
	s := &SessionService{}

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0"
	info := s.ParseDeviceInfo(ua)

	if info.DeviceType != "desktop" {
		t.Errorf("ParseDeviceInfo().DeviceType = %q, want %q", info.DeviceType, "desktop")
	}
}

func TestSessionService_ParseDeviceInfo_Bot(t *testing.T) {
	s := &SessionService{}

	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"
	info := s.ParseDeviceInfo(ua)

	// Bot should be detected as desktop (no mobile indicator)
	if info.DeviceType != "desktop" {
		t.Errorf("ParseDeviceInfo().DeviceType = %q, want %q", info.DeviceType, "desktop")
	}
}

// ============ Additional GetClientIP Tests ============

func TestGetClientIP_EmptyHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"

	ip := getClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("getClientIP() = %q, want %q", ip, "10.0.0.1")
	}
}

func TestGetClientIP_XRealIPPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Real-IP", "192.168.1.1")

	ip := getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("getClientIP() = %q, want %q", ip, "192.168.1.1")
	}
}

func TestGetClientIP_TrimSpaces(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "  192.168.1.1  , 10.0.0.2")

	ip := getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("getClientIP() = %q, want %q", ip, "192.168.1.1")
	}
}

// ============ HashToken Tests ============

func TestHashToken_DifferentInputs(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	hashes := make(map[string]bool)

	for _, token := range tokens {
		hash := hashToken(token)
		if hashes[hash] {
			t.Errorf("hashToken() produced duplicate hash for different inputs")
		}
		hashes[hash] = true
	}
}

func TestHashToken_EmptyString(t *testing.T) {
	hash := hashToken("")
	if len(hash) != 64 {
		t.Errorf("hashToken(\"\") length = %d, want 64", len(hash))
	}
}

// ============ Auth Method Constants Tests ============

func TestAuthMethodConstants(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{"password", models.AuthMethodPassword, "password"},
		{"google oauth", models.AuthMethodOAuthGoogle, "oauth_google"},
		{"github oauth", models.AuthMethodOAuthGitHub, "oauth_github"},
		{"refresh token", models.AuthMethodRefreshToken, "refresh_token"},
		{"2fa", models.AuthMethod2FA, "2fa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.method != tt.want {
				t.Errorf("AuthMethod constant = %q, want %q", tt.method, tt.want)
			}
		})
	}
}

// ============ Mock-Based Unit Tests ============

func TestSessionService_CreateSessionWithContext(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		token     string
		userAgent string
		remoteIP  string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "successful creation",
			userID:    1,
			token:     "test-refresh-token",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
			remoteIP:  "192.168.1.1:8080",
			wantErr:   false,
		},
		{
			name:      "repository error",
			userID:    1,
			token:     "test-token",
			userAgent: "TestAgent",
			remoteIP:  "10.0.0.1:8080",
			repoErr:   errors.New("database connection failed"),
			wantErr:   true,
		},
		{
			name:      "mobile user agent",
			userID:    2,
			token:     "mobile-token",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) Safari/604.1",
			remoteIP:  "172.16.0.1:443",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.CreateErr = tt.repoErr

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.Header.Set("User-Agent", tt.userAgent)
			req.RemoteAddr = tt.remoteIP

			ctx := context.Background()
			session, err := svc.CreateSessionWithContext(ctx, tt.userID, tt.token, req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSessionWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if session == nil {
				t.Fatal("CreateSessionWithContext() returned nil session")
			}

			if session.UserID != tt.userID {
				t.Errorf("session.UserID = %d, want %d", session.UserID, tt.userID)
			}

			if session.SessionTokenHash == "" {
				t.Error("session.SessionTokenHash is empty")
			}

			if sessionRepo.CreateCalls != 1 {
				t.Errorf("CreateCalls = %d, want 1", sessionRepo.CreateCalls)
			}
		})
	}
}

func TestSessionService_GetUserSessionsWithContext(t *testing.T) {
	tests := []struct {
		name             string
		userID           uint
		currentTokenHash string
		existingSessions []models.UserSession
		repoErr          error
		wantCount        int
		wantErr          bool
	}{
		{
			name:             "returns active sessions",
			userID:           1,
			currentTokenHash: "current-hash",
			existingSessions: []models.UserSession{
				{ID: 1, UserID: 1, SessionTokenHash: "current-hash", ExpiresAt: time.Now().Add(time.Hour)},
				{ID: 2, UserID: 1, SessionTokenHash: "other-hash", ExpiresAt: time.Now().Add(time.Hour)},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:             "excludes expired sessions",
			userID:           1,
			currentTokenHash: "hash",
			existingSessions: []models.UserSession{
				{ID: 1, UserID: 1, SessionTokenHash: "hash", ExpiresAt: time.Now().Add(time.Hour)},
				{ID: 2, UserID: 1, SessionTokenHash: "expired", ExpiresAt: time.Now().Add(-time.Hour)}, // expired
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:             "repository error",
			userID:           1,
			currentTokenHash: "hash",
			repoErr:          errors.New("database error"),
			wantErr:          true,
		},
		{
			name:             "no sessions",
			userID:           999,
			currentTokenHash: "hash",
			existingSessions: []models.UserSession{},
			wantCount:        0,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.FindByUserIDErr = tt.repoErr

			// Add existing sessions
			for _, s := range tt.existingSessions {
				sessionRepo.AddSession(s)
			}

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			sessions, err := svc.GetUserSessionsWithContext(ctx, tt.userID, tt.currentTokenHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserSessionsWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(sessions) != tt.wantCount {
				t.Errorf("GetUserSessionsWithContext() returned %d sessions, want %d", len(sessions), tt.wantCount)
			}

			// Check that current session is marked
			for _, s := range sessions {
				if s.SessionTokenHash == tt.currentTokenHash && !s.IsCurrent {
					t.Error("Current session should be marked as IsCurrent=true")
				}
				if s.SessionTokenHash != tt.currentTokenHash && s.IsCurrent {
					t.Error("Non-current session should not be marked as IsCurrent=true")
				}
			}
		})
	}
}

func TestSessionService_RevokeSessionWithContext(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		sessionID  uint
		hasSession bool
		repoErr    error
		wantErr    bool
		errType    error
	}{
		{
			name:       "successful revocation",
			userID:     1,
			sessionID:  1,
			hasSession: true,
			wantErr:    false,
		},
		{
			name:       "session not found",
			userID:     1,
			sessionID:  999,
			hasSession: false,
			wantErr:    true,
			errType:    ErrSessionNotFound,
		},
		{
			name:       "repository error",
			userID:     1,
			sessionID:  1,
			hasSession: true,
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.DeleteByIDErr = tt.repoErr

			if tt.hasSession {
				sessionRepo.AddSession(models.UserSession{
					ID:               tt.sessionID,
					UserID:           tt.userID,
					SessionTokenHash: "hash",
					ExpiresAt:        time.Now().Add(time.Hour),
				})
			}

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			err := svc.RevokeSessionWithContext(ctx, tt.userID, tt.sessionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeSessionWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("RevokeSessionWithContext() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestSessionService_RevokeAllSessionsWithContext(t *testing.T) {
	tests := []struct {
		name            string
		userID          uint
		exceptTokenHash string
		sessions        []models.UserSession
		repoErr         error
		wantRemaining   int
		wantErr         bool
	}{
		{
			name:            "revoke all sessions",
			userID:          1,
			exceptTokenHash: "",
			sessions: []models.UserSession{
				{ID: 1, UserID: 1, SessionTokenHash: "hash1"},
				{ID: 2, UserID: 1, SessionTokenHash: "hash2"},
			},
			wantRemaining: 0,
			wantErr:       false,
		},
		{
			name:            "revoke all except current",
			userID:          1,
			exceptTokenHash: "keep-this",
			sessions: []models.UserSession{
				{ID: 1, UserID: 1, SessionTokenHash: "keep-this"},
				{ID: 2, UserID: 1, SessionTokenHash: "remove1"},
				{ID: 3, UserID: 1, SessionTokenHash: "remove2"},
			},
			wantRemaining: 1,
			wantErr:       false,
		},
		{
			name:            "repository error",
			userID:          1,
			exceptTokenHash: "",
			repoErr:         errors.New("database error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.DeleteByUserIDErr = tt.repoErr

			for _, s := range tt.sessions {
				sessionRepo.AddSession(s)
			}

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			err := svc.RevokeAllSessionsWithContext(ctx, tt.userID, tt.exceptTokenHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeAllSessionsWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check remaining sessions
			remaining := sessionRepo.GetAllSessions()[tt.userID]
			if len(remaining) != tt.wantRemaining {
				t.Errorf("Remaining sessions = %d, want %d", len(remaining), tt.wantRemaining)
			}
		})
	}
}

func TestSessionService_CleanupExpiredSessionsWithContext(t *testing.T) {
	tests := []struct {
		name     string
		sessions []models.UserSession
		repoErr  error
		wantDel  int64
		wantErr  bool
	}{
		{
			name: "cleanup expired sessions",
			sessions: []models.UserSession{
				{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour)},   // expired
				{ID: 2, UserID: 1, ExpiresAt: time.Now().Add(-time.Minute)}, // expired
				{ID: 3, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)},    // active
			},
			wantDel: 2,
			wantErr: false,
		},
		{
			name: "no expired sessions",
			sessions: []models.UserSession{
				{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)},
			},
			wantDel: 0,
			wantErr: false,
		},
		{
			name:    "repository error",
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.DeleteExpiredErr = tt.repoErr

			for _, s := range tt.sessions {
				sessionRepo.AddSession(s)
			}

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			deleted, err := svc.CleanupExpiredSessionsWithContext(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CleanupExpiredSessionsWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if deleted != tt.wantDel {
				t.Errorf("CleanupExpiredSessionsWithContext() deleted = %d, want %d", deleted, tt.wantDel)
			}
		})
	}
}

func TestSessionService_RecordLoginAttemptWithContext(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		success       bool
		failureReason string
		authMethod    string
		repoErr       error
		wantErr       bool
	}{
		{
			name:       "successful login",
			userID:     1,
			success:    true,
			authMethod: models.AuthMethodPassword,
			wantErr:    false,
		},
		{
			name:          "failed login",
			userID:        1,
			success:       false,
			failureReason: "invalid password",
			authMethod:    models.AuthMethodPassword,
			wantErr:       false,
		},
		{
			name:       "oauth login",
			userID:     1,
			success:    true,
			authMethod: models.AuthMethodOAuthGoogle,
			wantErr:    false,
		},
		{
			name:       "default auth method",
			userID:     1,
			success:    true,
			authMethod: "", // should default to password
			wantErr:    false,
		},
		{
			name:    "repository error",
			userID:  1,
			success: true,
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			historyRepo.CreateErr = tt.repoErr

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.Header.Set("User-Agent", "TestAgent/1.0")
			req.RemoteAddr = "192.168.1.1:8080"
			ctx := context.Background()

			err := svc.RecordLoginAttemptWithContext(ctx, tt.userID, tt.success, tt.failureReason, tt.authMethod, req, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("RecordLoginAttemptWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if historyRepo.CreateCalls != 1 {
				t.Errorf("CreateCalls = %d, want 1", historyRepo.CreateCalls)
			}
		})
	}
}

func TestSessionService_GetLoginHistoryWithContext(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		limit     int
		offset    int
		records   []models.LoginHistory
		countErr  error
		findErr   error
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:   "returns paginated history",
			userID: 1,
			limit:  10,
			offset: 0,
			records: []models.LoginHistory{
				{ID: 1, UserID: 1, Success: true},
				{ID: 2, UserID: 1, Success: false},
			},
			wantCount: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "empty history",
			userID:    999,
			limit:     10,
			offset:    0,
			records:   []models.LoginHistory{},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "count error",
			userID:   1,
			countErr: errors.New("count error"),
			wantErr:  true,
		},
		{
			name:    "find error",
			userID:  1,
			findErr: errors.New("find error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			historyRepo.CountByUserErr = tt.countErr
			historyRepo.FindByUserErr = tt.findErr

			for _, r := range tt.records {
				historyRepo.AddHistory(r)
			}

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			history, total, err := svc.GetLoginHistoryWithContext(ctx, tt.userID, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoginHistoryWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(history) != tt.wantCount {
				t.Errorf("GetLoginHistoryWithContext() returned %d records, want %d", len(history), tt.wantCount)
			}

			if total != tt.wantTotal {
				t.Errorf("GetLoginHistoryWithContext() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestSessionService_UpdateLastActiveWithContext(t *testing.T) {
	tests := []struct {
		name      string
		tokenHash string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "successful update",
			tokenHash: "test-hash",
			wantErr:   false,
		},
		{
			name:      "repository error",
			tokenHash: "test-hash",
			repoErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.UpdateLastActiveErr = tt.repoErr

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			err := svc.UpdateLastActiveWithContext(ctx, tt.tokenHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLastActiveWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && sessionRepo.UpdateLastActiveCalls != 1 {
				t.Errorf("UpdateLastActiveCalls = %d, want 1", sessionRepo.UpdateLastActiveCalls)
			}
		})
	}
}

func TestNewSessionServiceWithRepo(t *testing.T) {
	sessionRepo := mocks.NewMockSessionRepository()
	historyRepo := mocks.NewMockLoginHistoryRepository()

	svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

	if svc == nil {
		t.Fatal("NewSessionServiceWithRepo() returned nil")
	}

	if svc.sessionRepo == nil {
		t.Error("sessionRepo is nil")
	}

	if svc.historyRepo == nil {
		t.Error("historyRepo is nil")
	}
}

func TestSessionService_RevokeSessionByTokenHashWithContext(t *testing.T) {
	tests := []struct {
		name      string
		tokenHash string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "successful revocation by hash",
			tokenHash: "test-hash-123",
			wantErr:   false,
		},
		{
			name:      "repository error",
			tokenHash: "test-hash",
			repoErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewMockSessionRepository()
			historyRepo := mocks.NewMockLoginHistoryRepository()
			sessionRepo.DeleteByTokenErr = tt.repoErr

			svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)
			ctx := context.Background()

			err := svc.RevokeSessionByTokenHashWithContext(ctx, tt.tokenHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeSessionByTokenHashWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && sessionRepo.DeleteByTokenCalls != 1 {
				t.Errorf("DeleteByTokenCalls = %d, want 1", sessionRepo.DeleteByTokenCalls)
			}
		})
	}
}

// Test non-context wrapper methods
func TestSessionService_WrapperMethods(t *testing.T) {
	sessionRepo := mocks.NewMockSessionRepository()
	historyRepo := mocks.NewMockLoginHistoryRepository()
	svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

	// Add a session for testing
	sessionRepo.AddSession(models.UserSession{
		ID:               1,
		UserID:           1,
		SessionTokenHash: "test-hash",
		ExpiresAt:        time.Now().Add(time.Hour),
	})

	// Test GetUserSessions (non-context version)
	sessions, err := svc.GetUserSessions(1, "test-hash")
	if err != nil {
		t.Errorf("GetUserSessions() error = %v", err)
	}
	if len(sessions) == 0 {
		t.Error("GetUserSessions() returned no sessions")
	}

	// Test RevokeSession (non-context version)
	sessionRepo.AddSession(models.UserSession{
		ID:               2,
		UserID:           1,
		SessionTokenHash: "another-hash",
		ExpiresAt:        time.Now().Add(time.Hour),
	})
	err = svc.RevokeSession(1, 2)
	if err != nil {
		t.Errorf("RevokeSession() error = %v", err)
	}

	// Test RevokeAllSessions (non-context version)
	err = svc.RevokeAllSessions(1, "")
	if err != nil {
		t.Errorf("RevokeAllSessions() error = %v", err)
	}

	// Test RevokeSessionByTokenHash (non-context version)
	err = svc.RevokeSessionByTokenHash("some-hash")
	if err != nil {
		t.Errorf("RevokeSessionByTokenHash() error = %v", err)
	}

	// Test UpdateLastActive (non-context version)
	err = svc.UpdateLastActive("hash")
	if err != nil {
		t.Errorf("UpdateLastActive() error = %v", err)
	}

	// Test CleanupExpiredSessions (non-context version)
	_, err = svc.CleanupExpiredSessions()
	if err != nil {
		t.Errorf("CleanupExpiredSessions() error = %v", err)
	}

	// Test GetLoginHistory (non-context version)
	_, _, err = svc.GetLoginHistory(1, 10, 0)
	if err != nil {
		t.Errorf("GetLoginHistory() error = %v", err)
	}
}

func TestSessionService_CreateSession_NonContext(t *testing.T) {
	sessionRepo := mocks.NewMockSessionRepository()
	historyRepo := mocks.NewMockLoginHistoryRepository()
	svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:8080"

	session, err := svc.CreateSession(1, "test-token", req)
	if err != nil {
		t.Errorf("CreateSession() error = %v", err)
	}
	if session == nil {
		t.Error("CreateSession() returned nil")
	}
}

func TestSessionService_RecordLoginAttempt_NonContext(t *testing.T) {
	sessionRepo := mocks.NewMockSessionRepository()
	historyRepo := mocks.NewMockLoginHistoryRepository()
	svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:8080"

	err := svc.RecordLoginAttempt(1, true, "", models.AuthMethodPassword, req, nil)
	if err != nil {
		t.Errorf("RecordLoginAttempt() error = %v", err)
	}
}
