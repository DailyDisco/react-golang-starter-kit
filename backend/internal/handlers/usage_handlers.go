package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/response"
	"react-golang-starter/internal/services"

	"github.com/rs/zerolog/log"
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

// RecordUsage records a usage event
// @Summary Record usage event
// @Description Records a usage event for metering purposes
// @Tags Usage
// @Accept json
// @Produce json
// @Param body body models.UsageEventRequest true "Usage event details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/usage/record [post]
func (h *UsageHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)

	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	var req models.UsageEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "invalid request body")
		return
	}

	// Validate required fields
	if req.EventType == "" {
		response.BadRequest(w, r, "event_type is required")
		return
	}
	if req.Resource == "" {
		response.BadRequest(w, r, "resource is required")
		return
	}

	// Set default quantity
	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}

	// Create usage event
	event := &models.UsageEvent{
		UserID:    &userID,
		EventType: req.EventType,
		Resource:  req.Resource,
		Quantity:  quantity,
		Unit:      req.Unit,
	}

	// Set metadata if provided
	if len(req.Metadata) > 0 {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err == nil {
			event.Metadata = string(metadataJSON)
		}
	}

	if err := h.usageService.RecordEvent(ctx, event); err != nil {
		response.HandleErrorWithMessage(w, r, err, "failed to record usage event")
		return
	}

	// Check limits after recording asynchronously
	// Use background context since request context will be canceled when response is sent
	go func(uid uint) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := h.usageService.CheckLimits(bgCtx, &uid, nil); err != nil {
			log.Error().Err(err).Uint("userID", uid).Msg("async limit check failed")
		}
	}(userID)

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "usage recorded",
		"event_id": event.ID,
	})
}
