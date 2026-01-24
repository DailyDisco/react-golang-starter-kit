package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/websocket"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UsageService handles usage metering operations
type UsageService struct {
	db         *gorm.DB
	hub        *websocket.Hub
	workQueue  chan *models.UsageEvent
	workerDone chan struct{}
}

const (
	// maxQueueSize limits buffered usage events to prevent unbounded memory growth
	maxQueueSize = 1000
	// numWorkers is the number of concurrent workers processing usage events
	numWorkers = 3
)

// NewUsageService creates a new usage service with bounded worker pool
func NewUsageService(db *gorm.DB) *UsageService {
	s := &UsageService{
		db:         db,
		workQueue:  make(chan *models.UsageEvent, maxQueueSize),
		workerDone: make(chan struct{}),
	}

	// Start worker pool for async usage processing
	for i := 0; i < numWorkers; i++ {
		go s.worker(i)
	}

	return s
}

// worker processes usage events from the work queue
func (s *UsageService) worker(id int) {
	log.Debug().Int("worker_id", id).Msg("usage worker started")
	for event := range s.workQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s.updatePeriodTotals(ctx, event)
		cancel()
	}
	log.Debug().Int("worker_id", id).Msg("usage worker stopped")
}

// Shutdown gracefully stops the worker pool
func (s *UsageService) Shutdown() {
	log.Info().Msg("shutting down usage service workers")
	close(s.workQueue)
}

// SetHub sets the WebSocket hub for broadcasting alerts
func (s *UsageService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// broadcastUsageAlert sends a usage alert to the user via WebSocket
func (s *UsageService) broadcastUsageAlert(userID uint, alertType, usageType string, current, limit int64, percentage int, limits models.UsageLimits) {
	if s.hub == nil {
		return
	}

	// Determine current plan and upgrade eligibility based on limits
	currentPlan := determinePlanFromLimits(limits)
	canUpgrade, suggestedPlan := getUpgradeSuggestion(currentPlan)

	var message string
	switch alertType {
	case "exceeded":
		if canUpgrade {
			message = fmt.Sprintf("You have exceeded your %s limit. Upgrade to %s for higher limits.", usageType, suggestedPlan)
		} else {
			message = fmt.Sprintf("You have exceeded your %s limit. Contact support for enterprise options.", usageType)
		}
	case "warning_90":
		message = fmt.Sprintf("You have used 90%% of your %s limit", usageType)
	case "warning_80":
		message = fmt.Sprintf("You have used 80%% of your %s limit", usageType)
	default:
		message = fmt.Sprintf("Usage alert for %s", usageType)
	}

	payload := websocket.UsageAlertPayload{
		AlertType:      alertType,
		UsageType:      usageType,
		CurrentUsage:   current,
		Limit:          limit,
		PercentageUsed: percentage,
		Message:        message,
		CanUpgrade:     canUpgrade,
		CurrentPlan:    currentPlan,
		SuggestedPlan:  suggestedPlan,
		UpgradeURL:     "/settings/billing",
	}

	s.hub.SendToUser(userID, websocket.MessageTypeUsageAlert, payload)
	log.Debug().Uint("user_id", userID).Str("alert_type", alertType).Str("usage_type", usageType).Msg("usage alert broadcasted")
}

// determinePlanFromLimits infers the subscription tier from usage limits
func determinePlanFromLimits(limits models.UsageLimits) string {
	switch {
	case limits.APICalls >= 1000000:
		return "enterprise"
	case limits.APICalls >= 100000:
		return "pro"
	default:
		return "free"
	}
}

// getUpgradeSuggestion returns whether the user can upgrade and the suggested plan
func getUpgradeSuggestion(currentPlan string) (bool, string) {
	switch currentPlan {
	case "free":
		return true, "Pro"
	case "pro":
		return true, "Enterprise"
	default:
		return false, ""
	}
}

// Common usage event types
const (
	UsageTypeAPICall    = "api_call"
	UsageTypeStorage    = "storage"
	UsageTypeCompute    = "compute"
	UsageTypeFileUpload = "file_upload"
)

// Default limits for free tier
var DefaultUsageLimits = models.UsageLimits{
	APICalls:     10000,      // 10k API calls per month
	StorageBytes: 1073741824, // 1 GB
	ComputeMS:    3600000,    // 1 hour of compute
	FileUploads:  100,        // 100 files per month
}

// TierLimits maps Stripe price IDs to usage limits
// Price IDs should match your Stripe dashboard configuration
var TierLimits = map[string]models.UsageLimits{
	"": DefaultUsageLimits, // Free tier (no subscription)
	// Pro tier - 10x free limits
	"price_pro_monthly": {
		APICalls:     100000,      // 100k API calls per month
		StorageBytes: 10737418240, // 10 GB
		ComputeMS:    36000000,    // 10 hours of compute
		FileUploads:  1000,        // 1000 files per month
	},
	"price_pro_yearly": {
		APICalls:     100000,
		StorageBytes: 10737418240,
		ComputeMS:    36000000,
		FileUploads:  1000,
	},
	// Enterprise tier - 100x free limits
	"price_enterprise_monthly": {
		APICalls:     1000000,      // 1M API calls per month
		StorageBytes: 107374182400, // 100 GB
		ComputeMS:    360000000,    // 100 hours of compute
		FileUploads:  10000,        // 10k files per month
	},
	"price_enterprise_yearly": {
		APICalls:     1000000,
		StorageBytes: 107374182400,
		ComputeMS:    360000000,
		FileUploads:  10000,
	},
}

// GetLimitsForPriceID returns usage limits for a Stripe price ID
func GetLimitsForPriceID(priceID string) models.UsageLimits {
	if limits, ok := TierLimits[priceID]; ok {
		return limits
	}
	return DefaultUsageLimits
}

// UpdateUserLimits updates usage limits for a user based on their subscription tier
func (s *UsageService) UpdateUserLimits(ctx context.Context, userID uint, priceID string) error {
	limits := GetLimitsForPriceID(priceID)
	limitsJSON, err := json.Marshal(limits)
	if err != nil {
		return fmt.Errorf("failed to marshal limits: %w", err)
	}

	periodStart, periodEnd := getCurrentBillingPeriod()
	now := time.Now().Format(time.RFC3339)

	// Update or create period with new limits
	result := s.db.WithContext(ctx).
		Where("user_id = ? AND period_start = ?", userID, periodStart).
		Assign(map[string]interface{}{
			"usage_limits": string(limitsJSON),
			"updated_at":   now,
		}).
		FirstOrCreate(&models.UsagePeriod{
			UserID:      &userID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			UsageTotals: "{}",
			UsageLimits: string(limitsJSON),
			CreatedAt:   now,
			UpdatedAt:   now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user limits: %w", result.Error)
	}

	log.Info().
		Uint("user_id", userID).
		Str("price_id", priceID).
		Msg("updated usage limits for user")

	return nil
}

// RecordEvent records a usage event
func (s *UsageService) RecordEvent(ctx context.Context, event *models.UsageEvent) error {
	// Set billing period if not provided
	if event.BillingPeriodStart == "" || event.BillingPeriodEnd == "" {
		start, end := getCurrentBillingPeriod()
		event.BillingPeriodStart = start
		event.BillingPeriodEnd = end
	}

	// Set default quantity
	if event.Quantity == 0 {
		event.Quantity = 1
	}

	// Set default unit
	if event.Unit == "" {
		event.Unit = "count"
	}

	// Set timestamp
	event.CreatedAt = time.Now().Format(time.RFC3339)

	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to record usage event: %w", err)
	}

	// Queue async update of aggregated totals (bounded worker pool)
	select {
	case s.workQueue <- event:
		// Event queued successfully
	default:
		// Queue full - log warning but don't block the request
		log.Warn().
			Str("event_type", event.EventType).
			Msg("usage event queue full, event will be aggregated on next period calculation")
	}

	return nil
}

// RecordAPICall is a convenience method for recording API calls
func (s *UsageService) RecordAPICall(ctx context.Context, userID *uint, orgID *uint, resource string, ipAddress string, userAgent string) {
	event := &models.UsageEvent{
		UserID:         userID,
		OrganizationID: orgID,
		EventType:      UsageTypeAPICall,
		Resource:       resource,
		Quantity:       1,
		Unit:           "count",
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

	if err := s.RecordEvent(ctx, event); err != nil {
		log.Warn().Err(err).Msg("failed to record API call usage")
	}
}

// RecordStorageUsage records storage usage
func (s *UsageService) RecordStorageUsage(ctx context.Context, userID *uint, orgID *uint, bytes int64, resource string) {
	event := &models.UsageEvent{
		UserID:         userID,
		OrganizationID: orgID,
		EventType:      UsageTypeStorage,
		Resource:       resource,
		Quantity:       bytes,
		Unit:           "bytes",
	}

	if err := s.RecordEvent(ctx, event); err != nil {
		log.Warn().Err(err).Msg("failed to record storage usage")
	}
}

// RecordFileUpload records a file upload
func (s *UsageService) RecordFileUpload(ctx context.Context, userID *uint, orgID *uint, fileName string, fileSize int64) {
	event := &models.UsageEvent{
		UserID:         userID,
		OrganizationID: orgID,
		EventType:      UsageTypeFileUpload,
		Resource:       fileName,
		Quantity:       1,
		Unit:           "count",
	}

	if err := s.RecordEvent(ctx, event); err != nil {
		log.Warn().Err(err).Msg("failed to record file upload usage")
	}

	// Also record storage bytes
	s.RecordStorageUsage(ctx, userID, orgID, fileSize, fileName)
}

// GetCurrentUsageSummary returns the current billing period's usage summary
func (s *UsageService) GetCurrentUsageSummary(ctx context.Context, userID *uint, orgID *uint) (*models.UsageSummaryResponse, error) {
	periodStart, periodEnd := getCurrentBillingPeriod()

	// Try to get existing period record
	var period models.UsagePeriod
	query := s.db.WithContext(ctx).Where("period_start = ? AND period_end = ?", periodStart, periodEnd)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		return nil, fmt.Errorf("either user_id or organization_id must be provided")
	}

	err := query.First(&period).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get usage period: %w", err)
	}

	// Parse usage totals
	var totals models.UsageTotals
	if period.UsageTotals != "" {
		if err := json.Unmarshal([]byte(period.UsageTotals), &totals); err != nil {
			log.Warn().Err(err).Msg("failed to parse usage totals")
		}
	}

	// Parse or use default limits
	var limits models.UsageLimits
	if period.UsageLimits != "" && period.UsageLimits != "{}" {
		if err := json.Unmarshal([]byte(period.UsageLimits), &limits); err != nil {
			log.Warn().Err(err).Msg("failed to parse usage limits")
			limits = DefaultUsageLimits
		}
	} else {
		limits = DefaultUsageLimits
	}

	// Calculate percentages
	response := &models.UsageSummaryResponse{
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Totals:         totals,
		Limits:         limits,
		LimitsExceeded: period.LimitsExceeded,
	}

	if limits.APICalls > 0 {
		response.Percentages.APICalls = int(float64(totals.APICalls) / float64(limits.APICalls) * 100)
	}
	if limits.StorageBytes > 0 {
		response.Percentages.StorageBytes = int(float64(totals.StorageBytes) / float64(limits.StorageBytes) * 100)
	}
	if limits.ComputeMS > 0 {
		response.Percentages.ComputeMS = int(float64(totals.ComputeMS) / float64(limits.ComputeMS) * 100)
	}
	if limits.FileUploads > 0 {
		response.Percentages.FileUploads = int(float64(totals.FileUploads) / float64(limits.FileUploads) * 100)
	}

	return response, nil
}

// CheckLimits checks if usage limits are exceeded and creates alerts if needed
func (s *UsageService) CheckLimits(ctx context.Context, userID *uint, orgID *uint) (bool, error) {
	summary, err := s.GetCurrentUsageSummary(ctx, userID, orgID)
	if err != nil {
		return false, err
	}

	limitsExceeded := false
	warningThresholds := []int{80, 90, 100}

	// Check each usage type
	checkUsageType := func(usageType string, current int64, limit int64, percentage int) {
		if limit <= 0 {
			return
		}

		for _, threshold := range warningThresholds {
			if percentage >= threshold {
				alertType := fmt.Sprintf("warning_%d", threshold)
				if threshold == 100 {
					alertType = "exceeded"
					limitsExceeded = true
				}

				// Try to create alert (ignore if duplicate)
				alert := &models.UsageAlert{
					UserID:         userID,
					OrganizationID: orgID,
					AlertType:      alertType,
					UsageType:      usageType,
					CurrentUsage:   current,
					UsageLimit:     limit,
					PercentageUsed: percentage,
					PeriodStart:    summary.PeriodStart,
					PeriodEnd:      summary.PeriodEnd,
					CreatedAt:      time.Now().Format(time.RFC3339),
				}

				// Use ON CONFLICT to avoid duplicates - check if it's a new alert
				result := s.db.WithContext(ctx).
					Where("user_id = ? AND alert_type = ? AND usage_type = ? AND period_start = ?",
						userID, alertType, usageType, summary.PeriodStart).
					FirstOrCreate(alert)

				// If a new alert was created, broadcast via WebSocket
				if result.RowsAffected > 0 && s.hub != nil && userID != nil {
					s.broadcastUsageAlert(*userID, alertType, usageType, current, limit, percentage, summary.Limits)
				}
			}
		}
	}

	checkUsageType(UsageTypeAPICall, summary.Totals.APICalls, summary.Limits.APICalls, summary.Percentages.APICalls)
	checkUsageType(UsageTypeStorage, summary.Totals.StorageBytes, summary.Limits.StorageBytes, summary.Percentages.StorageBytes)
	checkUsageType(UsageTypeCompute, summary.Totals.ComputeMS, summary.Limits.ComputeMS, summary.Percentages.ComputeMS)
	checkUsageType(UsageTypeFileUpload, summary.Totals.FileUploads, summary.Limits.FileUploads, summary.Percentages.FileUploads)

	return limitsExceeded, nil
}

// GetUnacknowledgedAlerts returns all unacknowledged alerts for a user or org
func (s *UsageService) GetUnacknowledgedAlerts(ctx context.Context, userID *uint, orgID *uint) ([]models.UsageAlert, error) {
	var alerts []models.UsageAlert
	query := s.db.WithContext(ctx).Where("acknowledged = ?", false)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	}

	if err := query.Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *UsageService) AcknowledgeAlert(ctx context.Context, alertID uint, acknowledgedBy uint) error {
	now := time.Now().Format(time.RFC3339)
	result := s.db.WithContext(ctx).Model(&models.UsageAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": now,
			"acknowledged_by": acknowledgedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alert not found: %d", alertID)
	}

	return nil
}

// updatePeriodTotals updates the aggregated totals for a usage period
func (s *UsageService) updatePeriodTotals(ctx context.Context, event *models.UsageEvent) {
	// Get or create period record
	var period models.UsagePeriod
	query := s.db.WithContext(ctx).
		Where("period_start = ? AND period_end = ?", event.BillingPeriodStart, event.BillingPeriodEnd)

	if event.UserID != nil {
		query = query.Where("user_id = ?", *event.UserID)
	} else if event.OrganizationID != nil {
		query = query.Where("organization_id = ?", *event.OrganizationID)
	} else {
		return
	}

	err := query.First(&period).Error
	if err == gorm.ErrRecordNotFound {
		// Create new period
		limitsJSON, _ := json.Marshal(DefaultUsageLimits)
		period = models.UsagePeriod{
			UserID:         event.UserID,
			OrganizationID: event.OrganizationID,
			PeriodStart:    event.BillingPeriodStart,
			PeriodEnd:      event.BillingPeriodEnd,
			UsageTotals:    "{}",
			UsageLimits:    string(limitsJSON),
			CreatedAt:      time.Now().Format(time.RFC3339),
			UpdatedAt:      time.Now().Format(time.RFC3339),
		}
		if err := s.db.WithContext(ctx).Create(&period).Error; err != nil {
			log.Warn().Err(err).Msg("failed to create usage period")
			return
		}
	} else if err != nil {
		log.Warn().Err(err).Msg("failed to get usage period")
		return
	}

	// Parse current totals
	var totals models.UsageTotals
	if period.UsageTotals != "" && period.UsageTotals != "{}" {
		if err := json.Unmarshal([]byte(period.UsageTotals), &totals); err != nil {
			log.Warn().Err(err).Msg("failed to parse usage totals")
		}
	}

	// Update totals based on event type
	switch event.EventType {
	case UsageTypeAPICall:
		totals.APICalls += event.Quantity
	case UsageTypeStorage:
		totals.StorageBytes += event.Quantity
	case UsageTypeCompute:
		totals.ComputeMS += event.Quantity
	case UsageTypeFileUpload:
		totals.FileUploads += event.Quantity
	}

	// Save updated totals
	totalsJSON, err := json.Marshal(totals)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal usage totals")
		return
	}

	now := time.Now().Format(time.RFC3339)
	if err := s.db.WithContext(ctx).Model(&period).Updates(map[string]interface{}{
		"usage_totals":       string(totalsJSON),
		"last_aggregated_at": now,
		"updated_at":         now,
	}).Error; err != nil {
		log.Warn().Err(err).Msg("failed to update usage period")
	}
}

// getCurrentBillingPeriod returns the start and end dates of the current billing period
func getCurrentBillingPeriod() (string, string) {
	now := time.Now()
	// Use the first day of the current month as billing period start
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// End is the first day of next month
	end := start.AddDate(0, 1, 0).Add(-time.Second)

	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

// GetUsageHistory returns usage history for past periods
func (s *UsageService) GetUsageHistory(ctx context.Context, userID *uint, orgID *uint, months int) ([]models.UsageSummaryResponse, error) {
	var periods []models.UsagePeriod
	query := s.db.WithContext(ctx).Order("period_start DESC").Limit(months)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		return nil, fmt.Errorf("either user_id or organization_id must be provided")
	}

	if err := query.Find(&periods).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}

	var history []models.UsageSummaryResponse
	for _, period := range periods {
		var totals models.UsageTotals
		var limits models.UsageLimits

		if period.UsageTotals != "" {
			if err := json.Unmarshal([]byte(period.UsageTotals), &totals); err != nil {
				log.Warn().
					Err(err).
					Str("period_start", period.PeriodStart).
					Str("raw_totals", period.UsageTotals).
					Msg("failed to parse usage totals in history, using zero values")
			}
		}
		if period.UsageLimits != "" {
			if err := json.Unmarshal([]byte(period.UsageLimits), &limits); err != nil {
				log.Warn().
					Err(err).
					Str("period_start", period.PeriodStart).
					Str("raw_limits", period.UsageLimits).
					Msg("failed to parse usage limits in history, using defaults")
				limits = DefaultUsageLimits
			}
		} else {
			limits = DefaultUsageLimits
		}

		summary := models.UsageSummaryResponse{
			PeriodStart:    period.PeriodStart,
			PeriodEnd:      period.PeriodEnd,
			Totals:         totals,
			Limits:         limits,
			LimitsExceeded: period.LimitsExceeded,
		}

		if limits.APICalls > 0 {
			summary.Percentages.APICalls = int(float64(totals.APICalls) / float64(limits.APICalls) * 100)
		}
		if limits.StorageBytes > 0 {
			summary.Percentages.StorageBytes = int(float64(totals.StorageBytes) / float64(limits.StorageBytes) * 100)
		}
		if limits.ComputeMS > 0 {
			summary.Percentages.ComputeMS = int(float64(totals.ComputeMS) / float64(limits.ComputeMS) * 100)
		}
		if limits.FileUploads > 0 {
			summary.Percentages.FileUploads = int(float64(totals.FileUploads) / float64(limits.FileUploads) * 100)
		}

		history = append(history, summary)
	}

	return history, nil
}
