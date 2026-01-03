package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"
)

func testSessionSetup(t *testing.T) (*SessionService, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB for the session service
	oldDB := database.DB
	database.DB = tt.DB

	svc := NewSessionService()

	return svc, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func createTestUserForSession(t *testing.T, suffix string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    "session_test_" + suffix + "@example.com",
		Name:     "Session Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createMockRequest(userAgent string) *http.Request {
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	return req
}

func TestSessionService_CreateSession_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("creates session with device info", func(t *testing.T) {
		user := createTestUserForSession(t, "create1")
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		req := createMockRequest(userAgent)

		session, err := svc.CreateSession(user.ID, "test-refresh-token", req)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		if session.ID == 0 {
			t.Error("Expected session to have ID")
		}
		if session.UserID != user.ID {
			t.Errorf("Expected user ID %d, got: %d", user.ID, session.UserID)
		}
		if session.SessionTokenHash == "" {
			t.Error("Expected session token hash to be set")
		}
		if session.SessionTokenHash == "test-refresh-token" {
			t.Error("Expected token to be hashed, not stored as plaintext")
		}
		if session.IPAddress != "192.168.1.100" {
			t.Errorf("Expected IP '192.168.1.100', got: %s", session.IPAddress)
		}
		if session.UserAgent != userAgent {
			t.Error("Expected user agent to match")
		}
	})

	t.Run("parses desktop browser correctly", func(t *testing.T) {
		user := createTestUserForSession(t, "create2")
		userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0"
		req := createMockRequest(userAgent)

		session, _ := svc.CreateSession(user.ID, "token-desktop", req)

		deviceInfo := svc.ParseDeviceInfo(session.UserAgent)
		if deviceInfo.DeviceType != "desktop" {
			t.Errorf("Expected device type 'desktop', got: %s", deviceInfo.DeviceType)
		}
	})

	t.Run("parses mobile browser correctly", func(t *testing.T) {
		user := createTestUserForSession(t, "create3")
		userAgent := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148"
		req := createMockRequest(userAgent)

		session, _ := svc.CreateSession(user.ID, "token-mobile", req)

		deviceInfo := svc.ParseDeviceInfo(session.UserAgent)
		if deviceInfo.DeviceType != "mobile" {
			t.Errorf("Expected device type 'mobile', got: %s", deviceInfo.DeviceType)
		}
	})

	t.Run("sets expiration time in future", func(t *testing.T) {
		user := createTestUserForSession(t, "create4")
		req := createMockRequest("TestAgent/1.0")

		session, _ := svc.CreateSession(user.ID, "token-expires", req)

		if session.ExpiresAt.Before(time.Now()) {
			t.Error("Expected expiration to be in the future")
		}
		// Should be approximately 7 days
		expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
		if session.ExpiresAt.After(expectedExpiry.Add(time.Hour)) || session.ExpiresAt.Before(expectedExpiry.Add(-time.Hour)) {
			t.Error("Expected expiration to be around 7 days from now")
		}
	})
}

func TestSessionService_GetUserSessions_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("returns all active sessions", func(t *testing.T) {
		user := createTestUserForSession(t, "getsess1")

		// Create multiple sessions
		req1 := createMockRequest("Browser1/1.0")
		session1, _ := svc.CreateSession(user.ID, "token-1", req1)

		req2 := createMockRequest("Browser2/1.0")
		svc.CreateSession(user.ID, "token-2", req2)

		sessions, err := svc.GetUserSessions(user.ID, session1.SessionTokenHash)
		if err != nil {
			t.Fatalf("GetUserSessions failed: %v", err)
		}

		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got: %d", len(sessions))
		}
	})

	t.Run("marks current session", func(t *testing.T) {
		user := createTestUserForSession(t, "getsess2")

		req := createMockRequest("CurrentBrowser/1.0")
		session, _ := svc.CreateSession(user.ID, "current-token", req)

		sessions, err := svc.GetUserSessions(user.ID, session.SessionTokenHash)
		if err != nil {
			t.Fatalf("GetUserSessions failed: %v", err)
		}

		found := false
		for _, s := range sessions {
			if s.SessionTokenHash == session.SessionTokenHash && s.IsCurrent {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected current session to be marked as IsCurrent")
		}
	})

	t.Run("excludes expired sessions", func(t *testing.T) {
		user := createTestUserForSession(t, "getsess3")

		// Create active session
		req := createMockRequest("Active/1.0")
		activeSession, _ := svc.CreateSession(user.ID, "active-token", req)

		// Create expired session directly in DB
		expiredSession := &models.UserSession{
			UserID:           user.ID,
			SessionTokenHash: hashToken("expired-token"),
			IPAddress:        "127.0.0.1",
			UserAgent:        "Expired/1.0",
			LastActiveAt:     time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt:        time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt:        time.Now().Add(-8 * 24 * time.Hour),
		}
		database.DB.Create(expiredSession)

		sessions, _ := svc.GetUserSessions(user.ID, activeSession.SessionTokenHash)

		for _, s := range sessions {
			if s.ID == expiredSession.ID {
				t.Error("Expected expired session to be excluded")
			}
		}
	})

	t.Run("orders by last active descending", func(t *testing.T) {
		user := createTestUserForSession(t, "getsess4")

		// Create sessions with different last active times
		req1 := createMockRequest("Old/1.0")
		svc.CreateSession(user.ID, "old-token", req1)

		time.Sleep(1100 * time.Millisecond) // Sleep > 1 second for RFC3339 timestamp difference

		req2 := createMockRequest("New/1.0")
		newSession, _ := svc.CreateSession(user.ID, "new-token", req2)

		sessions, _ := svc.GetUserSessions(user.ID, "")

		if len(sessions) > 0 && sessions[0].ID != newSession.ID {
			t.Error("Expected most recent session first")
		}
	})
}

func TestSessionService_RevokeSession_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("revokes specific session", func(t *testing.T) {
		user := createTestUserForSession(t, "revoke1")

		req := createMockRequest("ToRevoke/1.0")
		session, _ := svc.CreateSession(user.ID, "to-revoke-token", req)

		err := svc.RevokeSession(user.ID, session.ID)
		if err != nil {
			t.Fatalf("RevokeSession failed: %v", err)
		}

		// Verify session is deleted
		var count int64
		database.DB.Model(&models.UserSession{}).Where("id = ?", session.ID).Count(&count)
		if count != 0 {
			t.Error("Expected session to be deleted")
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		user := createTestUserForSession(t, "revoke2")

		err := svc.RevokeSession(user.ID, 99999)
		if err == nil {
			t.Error("Expected error for non-existent session")
		}
	})

	t.Run("cannot revoke other user's session", func(t *testing.T) {
		user1 := createTestUserForSession(t, "revoke3")

		user2 := &models.User{
			Email:    "other_session@example.com",
			Name:     "Other User",
			Password: "hashedpassword",
			Role:     models.RoleUser,
		}
		database.DB.Create(user2)

		req := createMockRequest("User1Session/1.0")
		session, _ := svc.CreateSession(user1.ID, "user1-token", req)

		// User2 tries to revoke User1's session
		err := svc.RevokeSession(user2.ID, session.ID)
		if err == nil {
			t.Error("Expected error when revoking another user's session")
		}
	})
}

func TestSessionService_RevokeAllSessions_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("revokes all sessions except current", func(t *testing.T) {
		user := createTestUserForSession(t, "revokeall1")

		// Create multiple sessions
		req1 := createMockRequest("Session1/1.0")
		currentSession, _ := svc.CreateSession(user.ID, "current-token", req1)

		req2 := createMockRequest("Session2/1.0")
		svc.CreateSession(user.ID, "other-token-1", req2)

		req3 := createMockRequest("Session3/1.0")
		svc.CreateSession(user.ID, "other-token-2", req3)

		err := svc.RevokeAllSessions(user.ID, currentSession.SessionTokenHash)
		if err != nil {
			t.Fatalf("RevokeAllSessions failed: %v", err)
		}

		sessions, _ := svc.GetUserSessions(user.ID, currentSession.SessionTokenHash)
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session (current), got: %d", len(sessions))
		}
		if sessions[0].ID != currentSession.ID {
			t.Error("Expected only current session to remain")
		}
	})

	t.Run("revokes all sessions when no exception", func(t *testing.T) {
		user := createTestUserForSession(t, "revokeall2")

		req := createMockRequest("Session/1.0")
		svc.CreateSession(user.ID, "token", req)

		err := svc.RevokeAllSessions(user.ID, "")
		if err != nil {
			t.Fatalf("RevokeAllSessions failed: %v", err)
		}

		sessions, _ := svc.GetUserSessions(user.ID, "")
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got: %d", len(sessions))
		}
	})
}

func TestSessionService_RevokeSessionByTokenHash_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("revokes session by token hash", func(t *testing.T) {
		user := createTestUserForSession(t, "revokehash1")

		req := createMockRequest("ToRevoke/1.0")
		session, _ := svc.CreateSession(user.ID, "token-to-revoke", req)

		err := svc.RevokeSessionByTokenHash(session.SessionTokenHash)
		if err != nil {
			t.Fatalf("RevokeSessionByTokenHash failed: %v", err)
		}

		var count int64
		database.DB.Model(&models.UserSession{}).Where("id = ?", session.ID).Count(&count)
		if count != 0 {
			t.Error("Expected session to be deleted")
		}
	})
}

func TestSessionService_UpdateLastActive_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("updates last active timestamp", func(t *testing.T) {
		user := createTestUserForSession(t, "update1")

		req := createMockRequest("Session/1.0")
		session, _ := svc.CreateSession(user.ID, "active-token", req)

		originalLastActive := session.LastActiveAt

		time.Sleep(1100 * time.Millisecond) // Sleep > 1 second for RFC3339 timestamp difference

		err := svc.UpdateLastActive(session.SessionTokenHash)
		if err != nil {
			t.Fatalf("UpdateLastActive failed: %v", err)
		}

		// Verify update
		var updated models.UserSession
		database.DB.First(&updated, session.ID)

		if updated.LastActiveAt == originalLastActive {
			t.Error("Expected last active timestamp to be updated")
		}
	})
}

func TestSessionService_CleanupExpiredSessions_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("removes expired sessions", func(t *testing.T) {
		user := createTestUserForSession(t, "cleanup1")

		// Create active session
		req := createMockRequest("Active/1.0")
		activeSession, _ := svc.CreateSession(user.ID, "active-token", req)

		// Create expired session
		expiredSession := &models.UserSession{
			UserID:           user.ID,
			SessionTokenHash: hashToken("expired-token"),
			IPAddress:        "127.0.0.1",
			UserAgent:        "Expired/1.0",
			LastActiveAt:     time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt:        time.Now().Add(-1 * time.Hour),
			CreatedAt:        time.Now().Add(-8 * 24 * time.Hour),
		}
		database.DB.Create(expiredSession)

		deleted, err := svc.CleanupExpiredSessions()
		if err != nil {
			t.Fatalf("CleanupExpiredSessions failed: %v", err)
		}

		if deleted < 1 {
			t.Errorf("Expected at least 1 deleted session, got: %d", deleted)
		}

		// Verify active session still exists
		var count int64
		database.DB.Model(&models.UserSession{}).Where("id = ?", activeSession.ID).Count(&count)
		if count != 1 {
			t.Error("Expected active session to remain")
		}

		// Verify expired session is deleted
		database.DB.Model(&models.UserSession{}).Where("id = ?", expiredSession.ID).Count(&count)
		if count != 0 {
			t.Error("Expected expired session to be deleted")
		}
	})
}

func TestSessionService_ParseDeviceInfo_Comprehensive(t *testing.T) {
	svc := &SessionService{}

	tests := []struct {
		name            string
		userAgent       string
		expectedType    string
		expectedBrowser string
	}{
		{
			name:            "Chrome on Windows",
			userAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expectedType:    "desktop",
			expectedBrowser: "Chrome",
		},
		{
			name:            "Safari on iPhone",
			userAgent:       "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			expectedType:    "mobile",
			expectedBrowser: "Safari",
		},
		{
			name:            "Safari on iPad",
			userAgent:       "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			expectedType:    "mobile", // Note: useragent library detects iPad as mobile, tablet detection is keyword-based
			expectedBrowser: "Safari",
		},
		{
			name:            "Firefox on Mac",
			userAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",
			expectedType:    "desktop",
			expectedBrowser: "Firefox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := svc.ParseDeviceInfo(tt.userAgent)

			if info.DeviceType != tt.expectedType {
				t.Errorf("Expected device type '%s', got: %s", tt.expectedType, info.DeviceType)
			}
			if info.Browser != tt.expectedBrowser {
				t.Errorf("Expected browser '%s', got: %s", tt.expectedBrowser, info.Browser)
			}
		})
	}
}

func TestSessionService_GetLocationFromIP_Comprehensive(t *testing.T) {
	svc := &SessionService{}

	t.Run("returns localhost for local IP", func(t *testing.T) {
		location := svc.GetLocationFromIP("127.0.0.1")
		if location.Country != "Local" {
			t.Errorf("Expected country 'Local', got: %s", location.Country)
		}
	})

	t.Run("returns unknown for other IPs", func(t *testing.T) {
		location := svc.GetLocationFromIP("8.8.8.8")
		if location.Country != "Unknown" {
			t.Errorf("Expected country 'Unknown', got: %s", location.Country)
		}
	})
}

func TestSessionService_LoginHistory_Integration(t *testing.T) {
	svc, cleanup := testSessionSetup(t)
	defer cleanup()

	t.Run("records successful login", func(t *testing.T) {
		user := createTestUserForSession(t, "login1")
		req := createMockRequest("Browser/1.0")

		err := svc.RecordLoginAttempt(user.ID, true, "", models.AuthMethodPassword, req, nil)
		if err != nil {
			t.Fatalf("RecordLoginAttempt failed: %v", err)
		}

		history, total, err := svc.GetLoginHistory(user.ID, 10, 0)
		if err != nil {
			t.Fatalf("GetLoginHistory failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected 1 login record, got: %d", total)
		}
		if !history[0].Success {
			t.Error("Expected login to be marked as successful")
		}
	})

	t.Run("records failed login with reason", func(t *testing.T) {
		user := createTestUserForSession(t, "login2")
		req := createMockRequest("Browser/1.0")

		err := svc.RecordLoginAttempt(user.ID, false, "invalid_password", models.AuthMethodPassword, req, nil)
		if err != nil {
			t.Fatalf("RecordLoginAttempt failed: %v", err)
		}

		history, _, _ := svc.GetLoginHistory(user.ID, 10, 0)

		failedFound := false
		for _, h := range history {
			if !h.Success && h.FailureReason == "invalid_password" {
				failedFound = true
				break
			}
		}
		if !failedFound {
			t.Error("Expected failed login with reason")
		}
	})

	t.Run("records different auth methods", func(t *testing.T) {
		user := createTestUserForSession(t, "login3")
		req := createMockRequest("Browser/1.0")

		svc.RecordLoginAttempt(user.ID, true, "", models.AuthMethodOAuthGoogle, req, nil)

		history, _, _ := svc.GetLoginHistory(user.ID, 10, 0)

		oauthFound := false
		for _, h := range history {
			if h.AuthMethod == models.AuthMethodOAuthGoogle {
				oauthFound = true
				break
			}
		}
		if !oauthFound {
			t.Error("Expected OAuth auth method in history")
		}
	})

	t.Run("paginates login history", func(t *testing.T) {
		user := createTestUserForSession(t, "login4")
		req := createMockRequest("Browser/1.0")

		// Create 15 login records
		for i := 0; i < 15; i++ {
			svc.RecordLoginAttempt(user.ID, true, "", models.AuthMethodPassword, req, nil)
		}

		// Get first page
		page1, total, _ := svc.GetLoginHistory(user.ID, 10, 0)
		if len(page1) != 10 {
			t.Errorf("Expected 10 records on page 1, got: %d", len(page1))
		}
		if total != 15 {
			t.Errorf("Expected total 15, got: %d", total)
		}

		// Get second page
		page2, _, _ := svc.GetLoginHistory(user.ID, 10, 10)
		if len(page2) != 5 {
			t.Errorf("Expected 5 records on page 2, got: %d", len(page2))
		}
	})

	t.Run("orders by created_at descending", func(t *testing.T) {
		user := createTestUserForSession(t, "login5")
		req := createMockRequest("Browser/1.0")

		svc.RecordLoginAttempt(user.ID, true, "", models.AuthMethodPassword, req, nil)
		time.Sleep(1100 * time.Millisecond) // Sleep > 1 second for RFC3339 timestamp difference
		svc.RecordLoginAttempt(user.ID, false, "failed", models.AuthMethodPassword, req, nil)

		history, _, _ := svc.GetLoginHistory(user.ID, 10, 0)

		if len(history) < 2 {
			t.Fatal("Expected at least 2 records")
		}

		// Most recent should be first (failed login)
		if history[0].Success {
			t.Error("Expected most recent (failed) login first")
		}
	})
}

func TestSessionService_getClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1, 172.16.0.1"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "192.168.1.1",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "10.0.0.100"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "10.0.0.100",
		},
		{
			name:       "Falls back to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.50:12345",
			expected:   "192.168.1.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("Expected IP '%s', got: %s", tt.expected, ip)
			}
		})
	}
}
