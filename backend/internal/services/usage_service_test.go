package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ GetLimitsForPriceID Tests ============

func TestGetLimitsForPriceID(t *testing.T) {
	tests := []struct {
		name    string
		priceID string
		want    models.UsageLimits
	}{
		{
			name:    "empty price ID returns default limits",
			priceID: "",
			want:    DefaultUsageLimits,
		},
		{
			name:    "unknown price ID returns default limits",
			priceID: "price_unknown_123",
			want:    DefaultUsageLimits,
		},
		{
			name:    "pro monthly tier",
			priceID: "price_pro_monthly",
			want: models.UsageLimits{
				APICalls:     100000,
				StorageBytes: 10737418240,
				ComputeMS:    36000000,
				FileUploads:  1000,
			},
		},
		{
			name:    "pro yearly tier",
			priceID: "price_pro_yearly",
			want: models.UsageLimits{
				APICalls:     100000,
				StorageBytes: 10737418240,
				ComputeMS:    36000000,
				FileUploads:  1000,
			},
		},
		{
			name:    "enterprise monthly tier",
			priceID: "price_enterprise_monthly",
			want: models.UsageLimits{
				APICalls:     1000000,
				StorageBytes: 107374182400,
				ComputeMS:    360000000,
				FileUploads:  10000,
			},
		},
		{
			name:    "enterprise yearly tier",
			priceID: "price_enterprise_yearly",
			want: models.UsageLimits{
				APICalls:     1000000,
				StorageBytes: 107374182400,
				ComputeMS:    360000000,
				FileUploads:  10000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLimitsForPriceID(tt.priceID)
			if got.APICalls != tt.want.APICalls {
				t.Errorf("APICalls = %d, want %d", got.APICalls, tt.want.APICalls)
			}
			if got.StorageBytes != tt.want.StorageBytes {
				t.Errorf("StorageBytes = %d, want %d", got.StorageBytes, tt.want.StorageBytes)
			}
			if got.ComputeMS != tt.want.ComputeMS {
				t.Errorf("ComputeMS = %d, want %d", got.ComputeMS, tt.want.ComputeMS)
			}
			if got.FileUploads != tt.want.FileUploads {
				t.Errorf("FileUploads = %d, want %d", got.FileUploads, tt.want.FileUploads)
			}
		})
	}
}

// ============ DefaultUsageLimits Tests ============

func TestDefaultUsageLimits(t *testing.T) {
	// Verify default limits are set correctly
	if DefaultUsageLimits.APICalls != 10000 {
		t.Errorf("DefaultUsageLimits.APICalls = %d, want 10000", DefaultUsageLimits.APICalls)
	}
	if DefaultUsageLimits.StorageBytes != 1073741824 {
		t.Errorf("DefaultUsageLimits.StorageBytes = %d, want 1073741824 (1 GB)", DefaultUsageLimits.StorageBytes)
	}
	if DefaultUsageLimits.ComputeMS != 3600000 {
		t.Errorf("DefaultUsageLimits.ComputeMS = %d, want 3600000 (1 hour)", DefaultUsageLimits.ComputeMS)
	}
	if DefaultUsageLimits.FileUploads != 100 {
		t.Errorf("DefaultUsageLimits.FileUploads = %d, want 100", DefaultUsageLimits.FileUploads)
	}
}

// ============ determinePlanFromLimits Tests ============

func TestDeterminePlanFromLimits(t *testing.T) {
	tests := []struct {
		name   string
		limits models.UsageLimits
		want   string
	}{
		{
			name:   "free tier - low API calls",
			limits: models.UsageLimits{APICalls: 10000},
			want:   "free",
		},
		{
			name:   "free tier - just under pro threshold",
			limits: models.UsageLimits{APICalls: 99999},
			want:   "free",
		},
		{
			name:   "pro tier - at threshold",
			limits: models.UsageLimits{APICalls: 100000},
			want:   "pro",
		},
		{
			name:   "pro tier - between pro and enterprise",
			limits: models.UsageLimits{APICalls: 500000},
			want:   "pro",
		},
		{
			name:   "enterprise tier - at threshold",
			limits: models.UsageLimits{APICalls: 1000000},
			want:   "enterprise",
		},
		{
			name:   "enterprise tier - high API calls",
			limits: models.UsageLimits{APICalls: 10000000},
			want:   "enterprise",
		},
		{
			name:   "zero API calls",
			limits: models.UsageLimits{APICalls: 0},
			want:   "free",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determinePlanFromLimits(tt.limits)
			if got != tt.want {
				t.Errorf("determinePlanFromLimits(%+v) = %q, want %q", tt.limits, got, tt.want)
			}
		})
	}
}

// ============ getUpgradeSuggestion Tests ============

func TestGetUpgradeSuggestion(t *testing.T) {
	tests := []struct {
		name           string
		currentPlan    string
		wantCanUpgrade bool
		wantSuggested  string
	}{
		{
			name:           "free plan can upgrade to Pro",
			currentPlan:    "free",
			wantCanUpgrade: true,
			wantSuggested:  "Pro",
		},
		{
			name:           "pro plan can upgrade to Enterprise",
			currentPlan:    "pro",
			wantCanUpgrade: true,
			wantSuggested:  "Enterprise",
		},
		{
			name:           "enterprise plan cannot upgrade",
			currentPlan:    "enterprise",
			wantCanUpgrade: false,
			wantSuggested:  "",
		},
		{
			name:           "unknown plan cannot upgrade",
			currentPlan:    "unknown",
			wantCanUpgrade: false,
			wantSuggested:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canUpgrade, suggested := getUpgradeSuggestion(tt.currentPlan)
			if canUpgrade != tt.wantCanUpgrade {
				t.Errorf("canUpgrade = %v, want %v", canUpgrade, tt.wantCanUpgrade)
			}
			if suggested != tt.wantSuggested {
				t.Errorf("suggested = %q, want %q", suggested, tt.wantSuggested)
			}
		})
	}
}

// ============ getCurrentBillingPeriod Tests ============

func TestGetCurrentBillingPeriod(t *testing.T) {
	start, end := getCurrentBillingPeriod()

	// Parse the dates
	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}

	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		t.Fatalf("failed to parse end date: %v", err)
	}

	now := time.Now()

	// Start should be the first day of the current month
	if startTime.Day() != 1 {
		t.Errorf("start day = %d, want 1", startTime.Day())
	}
	if startTime.Month() != now.Month() {
		t.Errorf("start month = %v, want %v", startTime.Month(), now.Month())
	}
	if startTime.Year() != now.Year() {
		t.Errorf("start year = %d, want %d", startTime.Year(), now.Year())
	}

	// End should be the last day of the current month
	// (first day of next month minus 1 second, but date only shows as last day)
	expectedEndMonth := now.Month()
	if now.Month() == time.December {
		expectedEndMonth = time.January
	} else {
		expectedEndMonth = now.Month()
	}
	// The end date should be in the same month as start
	if endTime.Month() != expectedEndMonth {
		t.Errorf("end month = %v, want %v", endTime.Month(), expectedEndMonth)
	}

	// Start should be before or equal to now
	if startTime.After(now) {
		t.Error("start date should not be after current time")
	}

	// End should be after start
	if !endTime.After(startTime) {
		t.Error("end date should be after start date")
	}
}

// ============ Usage Type Constants Tests ============

func TestUsageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"UsageTypeAPICall", UsageTypeAPICall, "api_call"},
		{"UsageTypeStorage", UsageTypeStorage, "storage"},
		{"UsageTypeCompute", UsageTypeCompute, "compute"},
		{"UsageTypeFileUpload", UsageTypeFileUpload, "file_upload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

// ============ TierLimits Map Tests ============

func TestTierLimitsContainsAllTiers(t *testing.T) {
	expectedTiers := []string{
		"",                         // Free tier
		"price_pro_monthly",        // Pro monthly
		"price_pro_yearly",         // Pro yearly
		"price_enterprise_monthly", // Enterprise monthly
		"price_enterprise_yearly",  // Enterprise yearly
	}

	for _, tier := range expectedTiers {
		t.Run("tier_"+tier, func(t *testing.T) {
			if _, ok := TierLimits[tier]; !ok {
				t.Errorf("TierLimits missing tier %q", tier)
			}
		})
	}
}

func TestTierLimitsProGreaterThanFree(t *testing.T) {
	free := TierLimits[""]
	proMonthly := TierLimits["price_pro_monthly"]
	proYearly := TierLimits["price_pro_yearly"]

	// Pro should have higher limits than free
	if proMonthly.APICalls <= free.APICalls {
		t.Error("Pro monthly APICalls should be greater than free")
	}
	if proMonthly.StorageBytes <= free.StorageBytes {
		t.Error("Pro monthly StorageBytes should be greater than free")
	}
	if proMonthly.ComputeMS <= free.ComputeMS {
		t.Error("Pro monthly ComputeMS should be greater than free")
	}
	if proMonthly.FileUploads <= free.FileUploads {
		t.Error("Pro monthly FileUploads should be greater than free")
	}

	// Pro yearly should equal pro monthly
	if proYearly.APICalls != proMonthly.APICalls {
		t.Error("Pro yearly APICalls should equal pro monthly")
	}
}

func TestTierLimitsEnterpriseGreaterThanPro(t *testing.T) {
	proMonthly := TierLimits["price_pro_monthly"]
	enterpriseMonthly := TierLimits["price_enterprise_monthly"]
	enterpriseYearly := TierLimits["price_enterprise_yearly"]

	// Enterprise should have higher limits than pro
	if enterpriseMonthly.APICalls <= proMonthly.APICalls {
		t.Error("Enterprise monthly APICalls should be greater than pro")
	}
	if enterpriseMonthly.StorageBytes <= proMonthly.StorageBytes {
		t.Error("Enterprise monthly StorageBytes should be greater than pro")
	}
	if enterpriseMonthly.ComputeMS <= proMonthly.ComputeMS {
		t.Error("Enterprise monthly ComputeMS should be greater than pro")
	}
	if enterpriseMonthly.FileUploads <= proMonthly.FileUploads {
		t.Error("Enterprise monthly FileUploads should be greater than pro")
	}

	// Enterprise yearly should equal enterprise monthly
	if enterpriseYearly.APICalls != enterpriseMonthly.APICalls {
		t.Error("Enterprise yearly APICalls should equal enterprise monthly")
	}
}

// ============ Service Worker Pool Constants Tests ============

func TestWorkerPoolConstants(t *testing.T) {
	// Verify constants are reasonable values
	if maxQueueSize < 100 {
		t.Errorf("maxQueueSize = %d, should be at least 100", maxQueueSize)
	}
	if maxQueueSize > 10000 {
		t.Errorf("maxQueueSize = %d, should not exceed 10000 to prevent memory issues", maxQueueSize)
	}

	if numWorkers < 1 {
		t.Errorf("numWorkers = %d, should be at least 1", numWorkers)
	}
	if numWorkers > 10 {
		t.Errorf("numWorkers = %d, should not exceed 10 to prevent resource exhaustion", numWorkers)
	}
}

// ============ NewUsageService Tests ============

func TestNewUsageService_NilDB(t *testing.T) {
	service := NewUsageService(nil)
	if service == nil {
		t.Fatal("NewUsageService(nil) returned nil, should return service instance")
	}

	// Clean up workers
	service.Shutdown()
}

func TestNewUsageService_ReturnsValidInstance(t *testing.T) {
	service := NewUsageService(nil)
	if service == nil {
		t.Fatal("NewUsageService() should return non-nil service")
	}

	// Verify service can accept SetHub calls without panic
	service.SetHub(nil)

	// Clean up
	service.Shutdown()
}

// ============ SetHub Tests ============

func TestUsageService_SetHub_Nil(t *testing.T) {
	service := NewUsageService(nil)
	defer service.Shutdown()

	// Should not panic when setting nil hub
	service.SetHub(nil)
}

func TestUsageService_SetHub_DoesNotPanic(t *testing.T) {
	service := NewUsageService(nil)
	defer service.Shutdown()

	// Setting hub multiple times should not panic
	service.SetHub(nil)
	service.SetHub(nil)
}

// ============ DefaultUsageLimits Structure Tests ============

func TestDefaultUsageLimits_AllFieldsSet(t *testing.T) {
	// Ensure all fields have non-zero values
	if DefaultUsageLimits.APICalls == 0 {
		t.Error("DefaultUsageLimits.APICalls should not be 0")
	}
	if DefaultUsageLimits.StorageBytes == 0 {
		t.Error("DefaultUsageLimits.StorageBytes should not be 0")
	}
	if DefaultUsageLimits.ComputeMS == 0 {
		t.Error("DefaultUsageLimits.ComputeMS should not be 0")
	}
	if DefaultUsageLimits.FileUploads == 0 {
		t.Error("DefaultUsageLimits.FileUploads should not be 0")
	}
}

func TestDefaultUsageLimits_ReasonableValues(t *testing.T) {
	// Free tier limits should be reasonable
	if DefaultUsageLimits.APICalls < 1000 {
		t.Error("APICalls should be at least 1000 for usability")
	}
	if DefaultUsageLimits.StorageBytes < 100*1024*1024 {
		t.Error("StorageBytes should be at least 100MB")
	}
	if DefaultUsageLimits.FileUploads < 10 {
		t.Error("FileUploads should be at least 10")
	}
}

// ============ determinePlanFromLimits Boundary Tests ============

func TestDeterminePlanFromLimits_NegativeValues(t *testing.T) {
	// Negative values should be treated as free tier
	limits := models.UsageLimits{APICalls: -1}
	got := determinePlanFromLimits(limits)
	if got != "free" {
		t.Errorf("determinePlanFromLimits with negative APICalls = %q, want %q", got, "free")
	}
}

func TestDeterminePlanFromLimits_ExactBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		apiCalls int64
		want     string
	}{
		{"just below pro", 99999, "free"},
		{"exactly pro", 100000, "pro"},
		{"just above pro", 100001, "pro"},
		{"just below enterprise", 999999, "pro"},
		{"exactly enterprise", 1000000, "enterprise"},
		{"just above enterprise", 1000001, "enterprise"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := models.UsageLimits{APICalls: tt.apiCalls}
			got := determinePlanFromLimits(limits)
			if got != tt.want {
				t.Errorf("determinePlanFromLimits(APICalls=%d) = %q, want %q", tt.apiCalls, got, tt.want)
			}
		})
	}
}

// ============ getCurrentBillingPeriod Format Tests ============

func TestGetCurrentBillingPeriod_Format(t *testing.T) {
	start, end := getCurrentBillingPeriod()

	// Verify date format is YYYY-MM-DD
	if len(start) != 10 {
		t.Errorf("start date length = %d, want 10 (YYYY-MM-DD format)", len(start))
	}
	if len(end) != 10 {
		t.Errorf("end date length = %d, want 10 (YYYY-MM-DD format)", len(end))
	}

	// Verify dashes are in correct positions
	if start[4] != '-' || start[7] != '-' {
		t.Errorf("start date %q does not match YYYY-MM-DD format", start)
	}
	if end[4] != '-' || end[7] != '-' {
		t.Errorf("end date %q does not match YYYY-MM-DD format", end)
	}
}

func TestGetCurrentBillingPeriod_ConsistentAcrossCalls(t *testing.T) {
	// Multiple calls in same test should return same values
	start1, end1 := getCurrentBillingPeriod()
	start2, end2 := getCurrentBillingPeriod()

	if start1 != start2 {
		t.Errorf("start dates differ across calls: %q vs %q", start1, start2)
	}
	if end1 != end2 {
		t.Errorf("end dates differ across calls: %q vs %q", end1, end2)
	}
}

// ============ getUpgradeSuggestion Edge Cases ============

func TestGetUpgradeSuggestion_CaseInsensitivity(t *testing.T) {
	// Test that the function handles exact case matching
	tests := []struct {
		name           string
		plan           string
		wantCanUpgrade bool
	}{
		{"lowercase free", "free", true},
		{"uppercase FREE", "FREE", false},
		{"mixed case Free", "Free", false},
		{"lowercase pro", "pro", true},
		{"uppercase PRO", "PRO", false},
		{"lowercase enterprise", "enterprise", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canUpgrade, _ := getUpgradeSuggestion(tt.plan)
			if canUpgrade != tt.wantCanUpgrade {
				t.Errorf("getUpgradeSuggestion(%q) canUpgrade = %v, want %v", tt.plan, canUpgrade, tt.wantCanUpgrade)
			}
		})
	}
}

func TestGetUpgradeSuggestion_EmptyString(t *testing.T) {
	canUpgrade, suggested := getUpgradeSuggestion("")
	// Empty string is not a valid plan, should not be upgradeable
	if canUpgrade {
		t.Error("empty string plan should not be upgradeable")
	}
	if suggested != "" {
		t.Errorf("empty string plan should have empty suggestion, got %q", suggested)
	}
}

// ============ TierLimits Validation Tests ============

func TestTierLimits_AllFieldsPositive(t *testing.T) {
	for tier, limits := range TierLimits {
		t.Run("tier_"+tier, func(t *testing.T) {
			if limits.APICalls <= 0 {
				t.Errorf("APICalls for tier %q should be positive", tier)
			}
			if limits.StorageBytes <= 0 {
				t.Errorf("StorageBytes for tier %q should be positive", tier)
			}
			if limits.ComputeMS <= 0 {
				t.Errorf("ComputeMS for tier %q should be positive", tier)
			}
			if limits.FileUploads <= 0 {
				t.Errorf("FileUploads for tier %q should be positive", tier)
			}
		})
	}
}

func TestTierLimits_MonthlyEqualsYearly(t *testing.T) {
	// Pro monthly should equal pro yearly
	proMonthly := TierLimits["price_pro_monthly"]
	proYearly := TierLimits["price_pro_yearly"]

	if proMonthly.APICalls != proYearly.APICalls {
		t.Error("Pro monthly APICalls should equal yearly")
	}
	if proMonthly.StorageBytes != proYearly.StorageBytes {
		t.Error("Pro monthly StorageBytes should equal yearly")
	}
	if proMonthly.ComputeMS != proYearly.ComputeMS {
		t.Error("Pro monthly ComputeMS should equal yearly")
	}
	if proMonthly.FileUploads != proYearly.FileUploads {
		t.Error("Pro monthly FileUploads should equal yearly")
	}

	// Enterprise monthly should equal enterprise yearly
	entMonthly := TierLimits["price_enterprise_monthly"]
	entYearly := TierLimits["price_enterprise_yearly"]

	if entMonthly.APICalls != entYearly.APICalls {
		t.Error("Enterprise monthly APICalls should equal yearly")
	}
	if entMonthly.StorageBytes != entYearly.StorageBytes {
		t.Error("Enterprise monthly StorageBytes should equal yearly")
	}
	if entMonthly.ComputeMS != entYearly.ComputeMS {
		t.Error("Enterprise monthly ComputeMS should equal yearly")
	}
	if entMonthly.FileUploads != entYearly.FileUploads {
		t.Error("Enterprise monthly FileUploads should equal yearly")
	}
}

// ============ Mock-based Tests ============

func TestUsageService_RecordEvent(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	event := &models.UsageEvent{
		EventType: UsageTypeAPICall,
		Resource:  "/api/users",
		Quantity:  1,
	}

	err := service.RecordEvent(ctx, event)
	require.NoError(t, err)
	assert.Equal(t, 1, eventRepo.CreateCalls)
	assert.NotEmpty(t, event.BillingPeriodStart)
	assert.NotEmpty(t, event.BillingPeriodEnd)
	assert.NotEmpty(t, event.CreatedAt)
}

func TestUsageService_RecordEvent_SetsDefaults(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	event := &models.UsageEvent{
		EventType: UsageTypeStorage,
		Resource:  "file.txt",
	}

	err := service.RecordEvent(ctx, event)
	require.NoError(t, err)

	// Check defaults were set
	assert.Equal(t, int64(1), event.Quantity)
	assert.Equal(t, "count", event.Unit)
}

func TestUsageService_RecordEvent_Error(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	eventRepo.CreateErr = errors.New("database error")
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	event := &models.UsageEvent{
		EventType: UsageTypeAPICall,
		Resource:  "/api/users",
	}

	err := service.RecordEvent(ctx, event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record usage event")
}

func TestUsageService_GetUnacknowledgedAlerts_ByUser(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	userID := uint(1)
	alertRepo.AddAlert(models.UsageAlert{
		UserID:    &userID,
		AlertType: "warning_80",
		UsageType: UsageTypeAPICall,
	})
	alertRepo.AddAlert(models.UsageAlert{
		UserID:    &userID,
		AlertType: "warning_90",
		UsageType: UsageTypeStorage,
	})

	alerts, err := service.GetUnacknowledgedAlerts(ctx, &userID, nil)
	require.NoError(t, err)
	assert.Len(t, alerts, 2)
	assert.Equal(t, 1, alertRepo.FindUnacknowledgedByUserCalls)
}

func TestUsageService_GetUnacknowledgedAlerts_ByOrg(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	orgID := uint(1)
	alertRepo.AddAlert(models.UsageAlert{
		OrganizationID: &orgID,
		AlertType:      "exceeded",
		UsageType:      UsageTypeAPICall,
	})

	alerts, err := service.GetUnacknowledgedAlerts(ctx, nil, &orgID)
	require.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, 1, alertRepo.FindUnacknowledgedByOrgCalls)
}

func TestUsageService_GetUnacknowledgedAlerts_NoIDs(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	_, err := service.GetUnacknowledgedAlerts(ctx, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either user_id or organization_id must be provided")
}

func TestUsageService_GetUnacknowledgedAlerts_Error(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	alertRepo.FindUnacknowledgedByUserErr = errors.New("database error")
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	userID := uint(1)
	_, err := service.GetUnacknowledgedAlerts(ctx, &userID, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get alerts")
}

func TestUsageService_AcknowledgeAlert(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	userID := uint(1)
	alertRepo.AddAlert(models.UsageAlert{
		ID:        1,
		UserID:    &userID,
		AlertType: "warning_80",
	})

	err := service.AcknowledgeAlert(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, alertRepo.AcknowledgeCalls)
}

func TestUsageService_AcknowledgeAlert_NotFound(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	err := service.AcknowledgeAlert(ctx, 999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alert not found")
}

func TestUsageService_AcknowledgeAlert_Error(t *testing.T) {
	ctx := context.Background()
	alertRepo := mocks.NewMockUsageAlertRepository()
	alertRepo.AcknowledgeErr = errors.New("database error")
	service := NewUsageServiceWithRepo(nil, nil, nil, alertRepo)
	defer service.Shutdown()

	err := service.AcknowledgeAlert(ctx, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acknowledge alert")
}

func TestUsageService_GetUsageHistory_ByUser(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	userID := uint(1)
	periodRepo.AddPeriod(models.UsagePeriod{
		UserID:      &userID,
		PeriodStart: "2025-01-01",
		PeriodEnd:   "2025-01-31",
		UsageTotals: `{"api_calls": 5000, "storage_bytes": 1000000}`,
		UsageLimits: `{"api_calls": 10000, "storage_bytes": 1073741824}`,
	})

	history, err := service.GetUsageHistory(ctx, &userID, nil, 12)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, 1, periodRepo.FindHistoryByUserCalls)
}

func TestUsageService_GetUsageHistory_ByOrg(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	orgID := uint(1)
	periodRepo.AddPeriod(models.UsagePeriod{
		OrganizationID: &orgID,
		PeriodStart:    "2025-01-01",
		PeriodEnd:      "2025-01-31",
		UsageTotals:    `{}`,
		UsageLimits:    `{}`,
	})

	history, err := service.GetUsageHistory(ctx, nil, &orgID, 12)
	require.NoError(t, err)
	assert.NotNil(t, history)
	assert.Equal(t, 1, periodRepo.FindHistoryByOrgCalls)
}

func TestUsageService_GetUsageHistory_NoIDs(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	_, err := service.GetUsageHistory(ctx, nil, nil, 12)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either user_id or organization_id must be provided")
}

func TestUsageService_GetUsageHistory_Error(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	periodRepo.FindHistoryByUserErr = errors.New("database error")
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	userID := uint(1)
	_, err := service.GetUsageHistory(ctx, &userID, nil, 12)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get usage history")
}

func TestUsageService_GetUsageHistory_CalculatesPercentages(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	userID := uint(1)
	periodRepo.AddPeriod(models.UsagePeriod{
		UserID:      &userID,
		PeriodStart: "2025-01-01",
		PeriodEnd:   "2025-01-31",
		UsageTotals: `{"api_calls": 5000, "storage_bytes": 536870912, "compute_ms": 1800000, "file_uploads": 50}`,
		UsageLimits: `{"api_calls": 10000, "storage_bytes": 1073741824, "compute_ms": 3600000, "file_uploads": 100}`,
	})

	history, err := service.GetUsageHistory(ctx, &userID, nil, 12)
	require.NoError(t, err)
	require.Len(t, history, 1)

	// 5000/10000 = 50%
	assert.Equal(t, 50, history[0].Percentages.APICalls)
	// 536870912/1073741824 = 50%
	assert.Equal(t, 50, history[0].Percentages.StorageBytes)
	// 1800000/3600000 = 50%
	assert.Equal(t, 50, history[0].Percentages.ComputeMS)
	// 50/100 = 50%
	assert.Equal(t, 50, history[0].Percentages.FileUploads)
}

func TestUsageService_GetUsageHistory_DefaultsLimits(t *testing.T) {
	ctx := context.Background()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	service := NewUsageServiceWithRepo(nil, nil, periodRepo, nil)
	defer service.Shutdown()

	userID := uint(1)
	periodRepo.AddPeriod(models.UsagePeriod{
		UserID:      &userID,
		PeriodStart: "2025-01-01",
		PeriodEnd:   "2025-01-31",
		UsageTotals: `{}`,
		UsageLimits: "", // Empty limits should use defaults
	})

	history, err := service.GetUsageHistory(ctx, &userID, nil, 12)
	require.NoError(t, err)
	require.Len(t, history, 1)

	// Should use DefaultUsageLimits
	assert.Equal(t, DefaultUsageLimits.APICalls, history[0].Limits.APICalls)
	assert.Equal(t, DefaultUsageLimits.StorageBytes, history[0].Limits.StorageBytes)
}

// ============ NewUsageServiceWithRepo Tests ============

func TestNewUsageServiceWithRepo(t *testing.T) {
	eventRepo := mocks.NewMockUsageEventRepository()
	periodRepo := mocks.NewMockUsagePeriodRepository()
	alertRepo := mocks.NewMockUsageAlertRepository()

	service := NewUsageServiceWithRepo(nil, eventRepo, periodRepo, alertRepo)
	defer service.Shutdown()

	require.NotNil(t, service)
	assert.NotNil(t, service.eventRepo)
	assert.NotNil(t, service.periodRepo)
	assert.NotNil(t, service.alertRepo)
}

func TestNewUsageServiceWithRepo_WorkerPoolStarts(t *testing.T) {
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)

	// Service should have work queue initialized
	require.NotNil(t, service.workQueue)

	// Clean shutdown
	service.Shutdown()
}

// ============ RecordAPICall Tests ============

func TestUsageService_RecordAPICall(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	userID := uint(1)
	service.RecordAPICall(ctx, &userID, nil, "/api/users", "192.168.1.1", "Mozilla/5.0")

	assert.Equal(t, 1, eventRepo.CreateCalls)
	events := eventRepo.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, UsageTypeAPICall, events[0].EventType)
	assert.Equal(t, "/api/users", events[0].Resource)
	assert.Equal(t, int64(1), events[0].Quantity)
}

// ============ RecordStorageUsage Tests ============

func TestUsageService_RecordStorageUsage(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	userID := uint(1)
	service.RecordStorageUsage(ctx, &userID, nil, 1024*1024, "document.pdf")

	assert.Equal(t, 1, eventRepo.CreateCalls)
	events := eventRepo.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, UsageTypeStorage, events[0].EventType)
	assert.Equal(t, int64(1024*1024), events[0].Quantity)
	assert.Equal(t, "bytes", events[0].Unit)
}

// ============ RecordFileUpload Tests ============

func TestUsageService_RecordFileUpload(t *testing.T) {
	ctx := context.Background()
	eventRepo := mocks.NewMockUsageEventRepository()
	service := NewUsageServiceWithRepo(nil, eventRepo, nil, nil)
	defer service.Shutdown()

	userID := uint(1)
	service.RecordFileUpload(ctx, &userID, nil, "photo.jpg", 2*1024*1024)

	// Should create 2 events: file upload + storage
	assert.Equal(t, 2, eventRepo.CreateCalls)
	events := eventRepo.GetEvents()
	require.Len(t, events, 2)

	// First should be file upload
	assert.Equal(t, UsageTypeFileUpload, events[0].EventType)
	assert.Equal(t, int64(1), events[0].Quantity)

	// Second should be storage
	assert.Equal(t, UsageTypeStorage, events[1].EventType)
	assert.Equal(t, int64(2*1024*1024), events[1].Quantity)
}
