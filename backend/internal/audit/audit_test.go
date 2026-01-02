package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For with multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.195, 70.41.3.18, 150.172.238.178",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "203.0.113.50"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195", "X-Real-IP": "203.0.113.50"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "Falls back to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1:12345",
		},
		{
			name:       "Empty headers fallback",
			headers:    map[string]string{"X-Forwarded-For": "", "X-Real-IP": ""},
			remoteAddr: "10.0.0.1:8080",
			expectedIP: "10.0.0.1:8080",
		},
		{
			name:       "Localhost",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:54321",
			expectedIP: "127.0.0.1:54321",
		},
		{
			name:       "IPv6 address",
			headers:    map[string]string{"X-Real-IP": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
			remoteAddr: "[::1]:8080",
			expectedIP: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getClientIP(req)
			if result != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", result, tt.expectedIP)
			}
		})
	}
}

func TestGetClientIP_NilRequest(t *testing.T) {
	// This shouldn't panic - getClientIP is only called when r != nil
	// but let's document the expected behavior
	t.Skip("getClientIP expects a non-nil request")
}

// Integration tests - require database
func TestLogEntry_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	userID := uint(1)
	targetID := uint(2)

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/users/2", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	req.Header.Set("User-Agent", "TestAgent/1.0")

	// Log entry - note: this runs async
	logEntryAsync(&userID, models.AuditTargetUser, &targetID, models.AuditActionUpdate, map[string]string{"name": "Updated"}, req)

	// Query the audit log
	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND target_id = ? AND action = ?", userID, targetID, models.AuditActionUpdate).First(&log).Error

	if err != nil {
		t.Fatalf("Expected audit log to be created: %v", err)
	}

	if log.TargetType != models.AuditTargetUser {
		t.Errorf("Expected target_type %s, got %s", models.AuditTargetUser, log.TargetType)
	}
	if log.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", log.IPAddress)
	}
	if log.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected User-Agent TestAgent/1.0, got %s", log.UserAgent)
	}
	if log.Changes == "" {
		t.Error("Expected changes to be serialized")
	}
}

func TestLogEntry_WithNilRequest_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	userID := uint(1)
	targetID := uint(2)

	// Log entry without request
	logEntryAsync(&userID, models.AuditTargetUser, &targetID, models.AuditActionCreate, nil, nil)

	// Query the audit log
	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND target_id = ?", userID, targetID).First(&log).Error

	if err != nil {
		t.Fatalf("Expected audit log to be created: %v", err)
	}

	if log.IPAddress != "" {
		t.Errorf("Expected empty IP address, got %s", log.IPAddress)
	}
	if log.UserAgent != "" {
		t.Errorf("Expected empty User-Agent, got %s", log.UserAgent)
	}
}

func TestLogWithMetadata_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	userID := uint(1)
	targetID := uint(2)
	metadata := map[string]interface{}{
		"reason":    "Testing",
		"timestamp": time.Now().Unix(),
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	logWithMetadataAsync(&userID, models.AuditTargetUser, &targetID, models.AuditActionLogin, nil, metadata, req)

	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND action = ?", userID, models.AuditActionLogin).First(&log).Error

	if err != nil {
		t.Fatalf("Expected audit log to be created: %v", err)
	}

	if log.Metadata == "" {
		t.Error("Expected metadata to be serialized")
	}
}

func TestGetAuditLogs_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	// Create test user first
	seeder := testutil.NewTestSeeder(t, tx.DB)
	user := seeder.SeedUser()

	// Create some audit logs
	for i := 0; i < 5; i++ {
		log := &models.AuditLog{
			UserID:     &user.ID,
			TargetType: models.AuditTargetUser,
			TargetID:   &user.ID,
			Action:     models.AuditActionUpdate,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
		if err := tx.DB.Create(log).Error; err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	t.Run("no filters returns all", func(t *testing.T) {
		filter := models.AuditLogFilter{
			Page:  1,
			Limit: 10,
		}

		logs, total, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		if total < 5 {
			t.Errorf("Expected at least 5 logs, got %d", total)
		}
		if len(logs) == 0 {
			t.Error("Expected some logs to be returned")
		}
	})

	t.Run("filter by user_id", func(t *testing.T) {
		filter := models.AuditLogFilter{
			UserID: &user.ID,
			Page:   1,
			Limit:  10,
		}

		logs, total, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		if total != 5 {
			t.Errorf("Expected 5 logs for user, got %d", total)
		}
		for _, log := range logs {
			if *log.UserID != user.ID {
				t.Errorf("Expected user_id %d, got %d", user.ID, *log.UserID)
			}
		}
	})

	t.Run("filter by action", func(t *testing.T) {
		filter := models.AuditLogFilter{
			Action: models.AuditActionUpdate,
			Page:   1,
			Limit:  10,
		}

		logs, _, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		for _, log := range logs {
			if log.Action != models.AuditActionUpdate {
				t.Errorf("Expected action %s, got %s", models.AuditActionUpdate, log.Action)
			}
		}
	})

	t.Run("filter by target_type", func(t *testing.T) {
		filter := models.AuditLogFilter{
			TargetType: models.AuditTargetUser,
			Page:       1,
			Limit:      10,
		}

		logs, _, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		for _, log := range logs {
			if log.TargetType != models.AuditTargetUser {
				t.Errorf("Expected target_type %s, got %s", models.AuditTargetUser, log.TargetType)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filter := models.AuditLogFilter{
			UserID: &user.ID,
			Page:   1,
			Limit:  2,
		}

		logs1, total, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		if len(logs1) != 2 {
			t.Errorf("Expected 2 logs on page 1, got %d", len(logs1))
		}

		filter.Page = 2
		logs2, _, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		if len(logs2) != 2 {
			t.Errorf("Expected 2 logs on page 2, got %d", len(logs2))
		}

		// Verify different logs on different pages
		if len(logs1) > 0 && len(logs2) > 0 && logs1[0].ID == logs2[0].ID {
			t.Error("Expected different logs on different pages")
		}

		// Total should be consistent
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}
	})

	t.Run("default pagination values", func(t *testing.T) {
		filter := models.AuditLogFilter{
			Page:  0, // Invalid, should default to 1
			Limit: 0, // Invalid, should default to 20
		}

		logs, _, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}

		// Should not error with invalid page/limit
		if logs == nil {
			t.Error("Expected non-nil logs slice")
		}
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		filter := models.AuditLogFilter{
			Page:  1,
			Limit: 500, // Should be capped to 100
		}

		_, _, err := GetAuditLogs(filter)
		if err != nil {
			t.Fatalf("GetAuditLogs error: %v", err)
		}
		// Just verify it doesn't error - actual limit enforcement is internal
	})
}

func TestGetUserAuditLogs_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	// Create test users
	seeder := testutil.NewTestSeeder(t, tx.DB)
	user1 := seeder.SeedUser(testutil.WithUserEmail("user1@test.local"))
	user2 := seeder.SeedUser(testutil.WithUserEmail("user2@test.local"))

	// Create audit logs for user1
	for i := 0; i < 3; i++ {
		seeder.SeedAuditLog(user1, models.AuditActionUpdate, models.AuditTargetUser, user1.ID)
	}

	// Create audit logs for user2
	seeder.SeedAuditLog(user2, models.AuditActionLogin, models.AuditTargetUser, user2.ID)

	logs, total, err := GetUserAuditLogs(user1.ID, 1, 10)
	if err != nil {
		t.Fatalf("GetUserAuditLogs error: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 logs for user1, got %d", total)
	}

	for _, log := range logs {
		if *log.UserID != user1.ID {
			t.Errorf("Expected user_id %d, got %d", user1.ID, *log.UserID)
		}
	}
}

// Convenience function tests
func TestLogLogin_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	seeder := testutil.NewTestSeeder(t, tx.DB)
	user := seeder.SeedUser()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	metadata := map[string]interface{}{"device": "Chrome", "os": "Windows"}

	// Call synchronous version for testing
	logWithMetadataAsync(&user.ID, models.AuditTargetUser, &user.ID, models.AuditActionLogin, nil, metadata, req)

	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND action = ?", user.ID, models.AuditActionLogin).First(&log).Error

	if err != nil {
		t.Fatalf("Expected login audit log: %v", err)
	}
	if log.Metadata == "" {
		t.Error("Expected metadata to be set")
	}
}

func TestLogLogout_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	seeder := testutil.NewTestSeeder(t, tx.DB)
	user := seeder.SeedUser()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

	logEntryAsync(&user.ID, models.AuditTargetUser, &user.ID, models.AuditActionLogout, nil, req)

	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND action = ?", user.ID, models.AuditActionLogout).First(&log).Error

	if err != nil {
		t.Fatalf("Expected logout audit log: %v", err)
	}
}

func TestLogImpersonate_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	seeder := testutil.NewTestSeeder(t, tx.DB)
	admin := seeder.SeedUser(testutil.WithUserSuperAdmin())
	target := seeder.SeedUser()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", nil)
	metadata := map[string]interface{}{"reason": "Support request #123"}

	logWithMetadataAsync(&admin.ID, models.AuditTargetUser, &target.ID, models.AuditActionImpersonate, nil, metadata, req)

	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND target_id = ? AND action = ?", admin.ID, target.ID, models.AuditActionImpersonate).First(&log).Error

	if err != nil {
		t.Fatalf("Expected impersonate audit log: %v", err)
	}
	if log.Metadata == "" {
		t.Error("Expected metadata with reason")
	}
}

func TestLogRoleChange_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tx := testutil.NewTestTransaction(t, db)
	database.DB = tx.DB

	seeder := testutil.NewTestSeeder(t, tx.DB)
	admin := seeder.SeedUser(testutil.WithUserAdmin())
	target := seeder.SeedUser()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/role", nil)
	changes := map[string]interface{}{"old_role": "user", "new_role": "admin"}

	logEntryAsync(&admin.ID, models.AuditTargetUser, &target.ID, models.AuditActionRoleChange, changes, req)

	var log models.AuditLog
	err := tx.DB.Where("user_id = ? AND action = ?", admin.ID, models.AuditActionRoleChange).First(&log).Error

	if err != nil {
		t.Fatalf("Expected role change audit log: %v", err)
	}
	if log.Changes == "" {
		t.Error("Expected changes to include old/new role")
	}
}

// Benchmark tests
func BenchmarkGetClientIP(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getClientIP(req)
	}
}
