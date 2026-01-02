package handlers

import (
	"net/http"
	"strconv"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/response"
	"react-golang-starter/internal/services"
)

// UsageHandler handles usage metering endpoints
type UsageHandler struct {
	usageService *services.UsageService
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(usageService *services.UsageService) *UsageHandler {
	return &UsageHandler{usageService: usageService}
}

// GetCurrentUsage returns the current billing period's usage summary
// @Summary Get current usage summary
// @Description Returns usage metrics for the current billing period
// @Tags Usage
// @Produce json
// @Success 200 {object} models.UsageSummaryResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/usage [get]
func (h *UsageHandler) GetCurrentUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)

	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	summary, err := h.usageService.GetCurrentUsageSummary(ctx, &userID, nil)
	if err != nil {
		response.HandleErrorWithMessage(w, r, err, "failed to get usage summary")
		return
	}

	response.JSON(w, http.StatusOK, summary)
}

// GetUsageHistory returns usage history for past billing periods
// @Summary Get usage history
// @Description Returns usage metrics for past billing periods
// @Tags Usage
// @Produce json
// @Param months query int false "Number of months to retrieve (default 6)"
// @Success 200 {array} models.UsageSummaryResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/usage/history [get]
func (h *UsageHandler) GetUsageHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)

	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	months := 6 // default
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 24 {
			months = parsed
		}
	}

	history, err := h.usageService.GetUsageHistory(ctx, &userID, nil, months)
	if err != nil {
		response.HandleErrorWithMessage(w, r, err, "failed to get usage history")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// GetAlerts returns unacknowledged usage alerts
// @Summary Get usage alerts
// @Description Returns unacknowledged usage alerts
// @Tags Usage
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/usage/alerts [get]
func (h *UsageHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)

	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	alerts, err := h.usageService.GetUnacknowledgedAlerts(ctx, &userID, nil)
	if err != nil {
		response.HandleErrorWithMessage(w, r, err, "failed to get alerts")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// AcknowledgeAlert marks an alert as acknowledged
// @Summary Acknowledge usage alert
// @Description Marks a usage alert as acknowledged
// @Tags Usage
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/usage/alerts/{id}/acknowledge [post]
func (h *UsageHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)

	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	// Get alert ID from URL
	alertIDStr := r.PathValue("id")
	if alertIDStr == "" {
		response.BadRequest(w, r, "alert ID required")
		return
	}

	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		response.BadRequest(w, r, "invalid alert ID")
		return
	}

	if err := h.usageService.AcknowledgeAlert(ctx, uint(alertID), userID); err != nil {
		response.NotFound(w, r, "alert not found")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "alert acknowledged",
	})
}
