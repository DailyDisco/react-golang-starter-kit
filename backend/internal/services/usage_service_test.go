package services

import (
	"testing"
	"time"

	"react-golang-starter/internal/models"
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
