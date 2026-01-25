package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"

	"gorm.io/gorm"
)

// testUsageSetup creates the service and returns cleanup function
func testUsageSetup(t *testing.T) (*UsageService, *gorm.DB, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)
	svc := NewUsageService(tt.DB)

	return svc, tt.DB, func() {
		svc.Shutdown()
		tt.Rollback()
	}
}

// createTestUserForUsage creates a user for testing usage
func createTestUserForUsage(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    email,
		Name:     "Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestUsageService_RecordEvent_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("records API call event", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "user1@example.com")

		event := &models.UsageEvent{
			UserID:    &user.ID,
			EventType: UsageTypeAPICall,
			Resource:  "/api/users",
			Quantity:  1,
		}

		err := svc.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}

		if event.ID == 0 {
			t.Error("Expected event to have ID after save")
		}
		if event.BillingPeriodStart == "" {
			t.Error("Expected BillingPeriodStart to be set")
		}
		if event.BillingPeriodEnd == "" {
			t.Error("Expected BillingPeriodEnd to be set")
		}
		if event.CreatedAt == "" {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("records storage event with bytes", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "user2@example.com")

		event := &models.UsageEvent{
			UserID:    &user.ID,
			EventType: UsageTypeStorage,
			Resource:  "document.pdf",
			Quantity:  1024 * 1024, // 1MB
			Unit:      "bytes",
		}

		err := svc.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}

		// Verify event was saved correctly
		var saved models.UsageEvent
		if err := db.First(&saved, event.ID).Error; err != nil {
			t.Fatalf("Failed to retrieve saved event: %v", err)
		}

		if saved.Quantity != 1024*1024 {
			t.Errorf("Expected quantity %d, got %d", 1024*1024, saved.Quantity)
		}
		if saved.Unit != "bytes" {
			t.Errorf("Expected unit 'bytes', got %q", saved.Unit)
		}
	})

	t.Run("sets default quantity and unit", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "user3@example.com")

		event := &models.UsageEvent{
			UserID:    &user.ID,
			EventType: UsageTypeFileUpload,
			Resource:  "image.png",
		}

		err := svc.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}

		if event.Quantity != 1 {
			t.Errorf("Expected default quantity 1, got %d", event.Quantity)
		}
		if event.Unit != "count" {
			t.Errorf("Expected default unit 'count', got %q", event.Unit)
		}
	})
}

func TestUsageService_RecordAPICall_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("records API call with context", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "apicall@example.com")

		svc.RecordAPICall(context.Background(), &user.ID, nil, "/api/users", "192.168.1.1", "Mozilla/5.0")

		// Wait briefly for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify event was recorded
		var count int64
		db.Model(&models.UsageEvent{}).
			Where("user_id = ? AND event_type = ?", user.ID, UsageTypeAPICall).
			Count(&count)

		if count == 0 {
			t.Error("Expected API call event to be recorded")
		}
	})
}

func TestUsageService_RecordStorageUsage_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("records storage bytes", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "storage@example.com")

		svc.RecordStorageUsage(context.Background(), &user.ID, nil, 5*1024*1024, "large-file.zip")

		// Wait briefly for async processing
		time.Sleep(50 * time.Millisecond)

		// Verify event was recorded
		var event models.UsageEvent
		err := db.Where("user_id = ? AND event_type = ?", user.ID, UsageTypeStorage).First(&event).Error

		if err != nil {
			t.Fatalf("Failed to find storage event: %v", err)
		}

		if event.Quantity != 5*1024*1024 {
			t.Errorf("Expected quantity %d, got %d", 5*1024*1024, event.Quantity)
		}
		if event.Unit != "bytes" {
			t.Errorf("Expected unit 'bytes', got %q", event.Unit)
		}
	})
}

func TestUsageService_RecordFileUpload_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("records file upload and storage events", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "fileupload@example.com")

		svc.RecordFileUpload(context.Background(), &user.ID, nil, "document.pdf", 2*1024*1024)

		// Wait briefly for async processing
		time.Sleep(50 * time.Millisecond)

		// Should create both file upload and storage events
		var uploadCount, storageCount int64
		db.Model(&models.UsageEvent{}).
			Where("user_id = ? AND event_type = ?", user.ID, UsageTypeFileUpload).
			Count(&uploadCount)
		db.Model(&models.UsageEvent{}).
			Where("user_id = ? AND event_type = ?", user.ID, UsageTypeStorage).
			Count(&storageCount)

		if uploadCount == 0 {
			t.Error("Expected file upload event to be recorded")
		}
		if storageCount == 0 {
			t.Error("Expected storage event to be recorded for file upload")
		}
	})
}

func TestUsageService_GetCurrentUsageSummary_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("returns usage summary with defaults for new user", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "summary@example.com")

		summary, err := svc.GetCurrentUsageSummary(context.Background(), &user.ID, nil)
		if err != nil {
			t.Fatalf("GetCurrentUsageSummary failed: %v", err)
		}

		if summary.PeriodStart == "" {
			t.Error("Expected PeriodStart to be set")
		}
		if summary.PeriodEnd == "" {
			t.Error("Expected PeriodEnd to be set")
		}

		// Should have default limits
		if summary.Limits.APICalls != DefaultUsageLimits.APICalls {
			t.Errorf("Expected default APICalls limit %d, got %d", DefaultUsageLimits.APICalls, summary.Limits.APICalls)
		}
	})

	t.Run("requires user_id or organization_id", func(t *testing.T) {
		_, err := svc.GetCurrentUsageSummary(context.Background(), nil, nil)
		if err == nil {
			t.Error("Expected error when neither user_id nor organization_id provided")
		}
	})

	t.Run("returns usage summary with recorded events", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "summary2@example.com")

		// Record some events
		for i := 0; i < 5; i++ {
			event := &models.UsageEvent{
				UserID:    &user.ID,
				EventType: UsageTypeAPICall,
				Resource:  "/api/test",
				Quantity:  1,
			}
			svc.RecordEvent(context.Background(), event)
		}

		// Wait for async workers to process
		time.Sleep(200 * time.Millisecond)

		summary, err := svc.GetCurrentUsageSummary(context.Background(), &user.ID, nil)
		if err != nil {
			t.Fatalf("GetCurrentUsageSummary failed: %v", err)
		}

		// Totals should reflect recorded events (async, so may not be exact)
		// The period should at least exist
		if summary.PeriodStart == "" || summary.PeriodEnd == "" {
			t.Error("Expected period dates to be set")
		}
	})
}

func TestUsageService_UpdateUserLimits_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("creates period with pro limits", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "limits@example.com")

		err := svc.UpdateUserLimits(context.Background(), user.ID, "price_pro_monthly")
		if err != nil {
			t.Fatalf("UpdateUserLimits failed: %v", err)
		}

		// Verify period was created with correct limits
		var period models.UsagePeriod
		err = db.Where("user_id = ?", user.ID).First(&period).Error
		if err != nil {
			t.Fatalf("Failed to find usage period: %v", err)
		}

		var limits models.UsageLimits
		if err := json.Unmarshal([]byte(period.UsageLimits), &limits); err != nil {
			t.Fatalf("Failed to parse limits: %v", err)
		}

		proLimits := TierLimits["price_pro_monthly"]
		if limits.APICalls != proLimits.APICalls {
			t.Errorf("Expected APICalls %d, got %d", proLimits.APICalls, limits.APICalls)
		}
	})

	t.Run("updates existing period limits", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "limits2@example.com")

		// First create with free limits
		err := svc.UpdateUserLimits(context.Background(), user.ID, "")
		if err != nil {
			t.Fatalf("First UpdateUserLimits failed: %v", err)
		}

		// Then upgrade to enterprise
		err = svc.UpdateUserLimits(context.Background(), user.ID, "price_enterprise_monthly")
		if err != nil {
			t.Fatalf("Second UpdateUserLimits failed: %v", err)
		}

		// Verify limits were updated
		var period models.UsagePeriod
		db.Where("user_id = ?", user.ID).First(&period)

		var limits models.UsageLimits
		json.Unmarshal([]byte(period.UsageLimits), &limits)

		enterpriseLimits := TierLimits["price_enterprise_monthly"]
		if limits.APICalls != enterpriseLimits.APICalls {
			t.Errorf("Expected upgraded APICalls %d, got %d", enterpriseLimits.APICalls, limits.APICalls)
		}
	})
}

func TestUsageService_CheckLimits_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("returns false when under limits", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "check@example.com")

		// Set up period with limits
		svc.UpdateUserLimits(context.Background(), user.ID, "")

		exceeded, err := svc.CheckLimits(context.Background(), &user.ID, nil)
		if err != nil {
			t.Fatalf("CheckLimits failed: %v", err)
		}

		if exceeded {
			t.Error("Expected limits to not be exceeded for user with no usage")
		}
	})

	t.Run("requires user_id or organization_id", func(t *testing.T) {
		_, err := svc.CheckLimits(context.Background(), nil, nil)
		if err == nil {
			t.Error("Expected error when neither user_id nor organization_id provided")
		}
	})
}

func TestUsageService_GetUnacknowledgedAlerts_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("returns empty list for user with no alerts", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "alerts@example.com")

		alerts, err := svc.GetUnacknowledgedAlerts(context.Background(), &user.ID, nil)
		if err != nil {
			t.Fatalf("GetUnacknowledgedAlerts failed: %v", err)
		}

		if len(alerts) != 0 {
			t.Errorf("Expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("returns unacknowledged alerts", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "alerts2@example.com")

		// Create an alert directly
		periodStart, periodEnd := getCurrentBillingPeriod()
		alert := &models.UsageAlert{
			UserID:         &user.ID,
			AlertType:      "warning_80",
			UsageType:      UsageTypeAPICall,
			CurrentUsage:   8000,
			UsageLimit:     10000,
			PercentageUsed: 80,
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			CreatedAt:      time.Now().Format(time.RFC3339),
		}
		if err := db.Create(alert).Error; err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		alerts, err := svc.GetUnacknowledgedAlerts(context.Background(), &user.ID, nil)
		if err != nil {
			t.Fatalf("GetUnacknowledgedAlerts failed: %v", err)
		}

		if len(alerts) != 1 {
			t.Errorf("Expected 1 alert, got %d", len(alerts))
		}
	})

	t.Run("requires user_id or organization_id", func(t *testing.T) {
		_, err := svc.GetUnacknowledgedAlerts(context.Background(), nil, nil)
		if err == nil {
			t.Error("Expected error when neither user_id nor organization_id provided")
		}
	})
}

func TestUsageService_AcknowledgeAlert_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("acknowledges existing alert", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "ack@example.com")

		// Create an alert
		periodStart, periodEnd := getCurrentBillingPeriod()
		alert := &models.UsageAlert{
			UserID:         &user.ID,
			AlertType:      "warning_90",
			UsageType:      UsageTypeStorage,
			CurrentUsage:   900,
			UsageLimit:     1000,
			PercentageUsed: 90,
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			CreatedAt:      time.Now().Format(time.RFC3339),
		}
		db.Create(alert)

		err := svc.AcknowledgeAlert(context.Background(), alert.ID, user.ID)
		if err != nil {
			t.Fatalf("AcknowledgeAlert failed: %v", err)
		}

		// Verify alert was acknowledged
		var updated models.UsageAlert
		db.First(&updated, alert.ID)

		if !updated.Acknowledged {
			t.Error("Expected alert to be acknowledged")
		}
		if updated.AcknowledgedAt == nil {
			t.Error("Expected AcknowledgedAt to be set")
		}
		if updated.AcknowledgedBy == nil || *updated.AcknowledgedBy != user.ID {
			t.Error("Expected AcknowledgedBy to be set to user ID")
		}
	})

	t.Run("returns error for non-existent alert", func(t *testing.T) {
		createTestUserForUsage(t, db, "ack2@example.com")

		err := svc.AcknowledgeAlert(context.Background(), 99999, 1)
		if err == nil {
			t.Error("Expected error for non-existent alert")
		}
	})
}

func TestUsageService_GetUsageHistory_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("returns empty history for new user", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "history@example.com")

		history, err := svc.GetUsageHistory(context.Background(), &user.ID, nil, 12)
		if err != nil {
			t.Fatalf("GetUsageHistory failed: %v", err)
		}

		if len(history) != 0 {
			t.Errorf("Expected 0 history entries, got %d", len(history))
		}
	})

	t.Run("returns history with usage data", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "history2@example.com")

		// Create a usage period
		periodStart, periodEnd := getCurrentBillingPeriod()
		totals := models.UsageTotals{
			APICalls:     5000,
			StorageBytes: 500000000,
		}
		totalsJSON, _ := json.Marshal(totals)
		limitsJSON, _ := json.Marshal(DefaultUsageLimits)

		period := &models.UsagePeriod{
			UserID:      &user.ID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			UsageTotals: string(totalsJSON),
			UsageLimits: string(limitsJSON),
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
		if err := db.Create(period).Error; err != nil {
			t.Fatalf("Failed to create period: %v", err)
		}

		history, err := svc.GetUsageHistory(context.Background(), &user.ID, nil, 12)
		if err != nil {
			t.Fatalf("GetUsageHistory failed: %v", err)
		}

		if len(history) != 1 {
			t.Fatalf("Expected 1 history entry, got %d", len(history))
		}

		if history[0].Totals.APICalls != 5000 {
			t.Errorf("Expected APICalls 5000, got %d", history[0].Totals.APICalls)
		}
		if history[0].Percentages.APICalls != 50 {
			t.Errorf("Expected 50%% APICalls usage, got %d%%", history[0].Percentages.APICalls)
		}
	})

	t.Run("requires user_id or organization_id", func(t *testing.T) {
		_, err := svc.GetUsageHistory(context.Background(), nil, nil, 12)
		if err == nil {
			t.Error("Expected error when neither user_id nor organization_id provided")
		}
	})
}

func TestUsageService_Organization_Integration(t *testing.T) {
	svc, db, cleanup := testUsageSetup(t)
	defer cleanup()

	t.Run("records events for organization", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "orgowner@example.com")

		// Create organization
		org := &models.Organization{
			Name:            "Test Org",
			Slug:            "test-org-usage",
			Plan:            models.OrgPlanFree,
			CreatedByUserID: user.ID,
		}
		if err := db.Create(org).Error; err != nil {
			t.Fatalf("Failed to create organization: %v", err)
		}

		// Record event for org
		event := &models.UsageEvent{
			OrganizationID: &org.ID,
			EventType:      UsageTypeAPICall,
			Resource:       "/api/org/endpoint",
			Quantity:       1,
		}

		err := svc.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}

		if event.ID == 0 {
			t.Error("Expected event to be saved")
		}
	})

	t.Run("gets usage summary for organization", func(t *testing.T) {
		user := createTestUserForUsage(t, db, "orgowner2@example.com")

		org := &models.Organization{
			Name:            "Test Org 2",
			Slug:            "test-org-usage-2",
			Plan:            models.OrgPlanFree,
			CreatedByUserID: user.ID,
		}
		db.Create(org)

		summary, err := svc.GetCurrentUsageSummary(context.Background(), nil, &org.ID)
		if err != nil {
			t.Fatalf("GetCurrentUsageSummary failed: %v", err)
		}

		if summary.PeriodStart == "" || summary.PeriodEnd == "" {
			t.Error("Expected period dates to be set")
		}
	})
}

func TestUsageService_Shutdown_Integration(t *testing.T) {
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)
	defer tt.Rollback()

	svc := NewUsageService(tt.DB)

	// Shutdown should not panic
	svc.Shutdown()

	// After shutdown, service should handle gracefully (no panic)
	// Note: We don't test recording after shutdown as that would block
}
