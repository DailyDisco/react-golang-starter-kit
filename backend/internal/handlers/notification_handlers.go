package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// GetNotifications returns paginated notifications for the current user
// @Summary Get user notifications
// @Description Returns paginated list of notifications for the authenticated user
// @Tags Notifications
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Items per page (default 20, max 100)"
// @Param unread query bool false "Filter to only unread notifications"
// @Success 200 {object} models.NotificationsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/notifications [get]
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	// Parse pagination
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"

	// Build query
	query := database.DB.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("read = ?", false)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("failed to count notifications")
		response.HandleErrorWithMessage(w, r, err, "failed to get notifications")
		return
	}

	// Get unread count
	var unread int64
	if err := database.DB.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&unread).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("failed to count unread notifications")
	}

	// Fetch notifications
	var notifications []models.Notification
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&notifications).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("failed to fetch notifications")
		response.HandleErrorWithMessage(w, r, err, "failed to get notifications")
		return
	}

	// Transform to response
	notificationResponses := make([]models.NotificationResponse, len(notifications))
	for i, n := range notifications {
		notificationResponses[i] = models.NotificationResponse{
			ID:        n.ID,
			CreatedAt: n.CreatedAt,
			Type:      n.Type,
			Title:     n.Title,
			Message:   n.Message,
			Link:      n.Link,
			Read:      n.Read,
			ReadAt:    n.ReadAt,
		}
	}

	response.JSON(w, http.StatusOK, models.NotificationsResponse{
		Notifications: notificationResponses,
		Total:         total,
		Unread:        unread,
		Page:          page,
		PerPage:       perPage,
	})
}

// MarkNotificationRead marks a single notification as read
// @Summary Mark notification as read
// @Description Marks a specific notification as read
// @Tags Notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/notifications/{id}/read [post]
func MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	notificationID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, r, "invalid notification ID")
		return
	}

	now := time.Now()
	result := database.DB.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		})

	if result.Error != nil {
		log.Error().Err(result.Error).Uint("user_id", userID).Uint64("notification_id", notificationID).Msg("failed to mark notification as read")
		response.HandleErrorWithMessage(w, r, result.Error, "failed to mark notification as read")
		return
	}

	if result.RowsAffected == 0 {
		response.NotFound(w, r, "notification not found")
		return
	}

	response.JSON(w, http.StatusOK, models.SuccessResponse{Message: "notification marked as read"})
}

// MarkAllNotificationsRead marks all notifications as read for the current user
// @Summary Mark all notifications as read
// @Description Marks all unread notifications as read for the authenticated user
// @Tags Notifications
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/notifications/read-all [post]
func MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	now := time.Now()
	result := database.DB.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		})

	if result.Error != nil {
		log.Error().Err(result.Error).Uint("user_id", userID).Msg("failed to mark all notifications as read")
		response.HandleErrorWithMessage(w, r, result.Error, "failed to mark notifications as read")
		return
	}

	response.JSON(w, http.StatusOK, models.SuccessResponse{
		Message: "all notifications marked as read",
	})
}

// DeleteNotification deletes a single notification
// @Summary Delete notification
// @Description Deletes a specific notification
// @Tags Notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/notifications/{id} [delete]
func DeleteNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok || userID == 0 {
		response.Unauthorized(w, r, "unauthorized")
		return
	}

	notificationID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, r, "invalid notification ID")
		return
	}

	result := database.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		log.Error().Err(result.Error).Uint("user_id", userID).Uint64("notification_id", notificationID).Msg("failed to delete notification")
		response.HandleErrorWithMessage(w, r, result.Error, "failed to delete notification")
		return
	}

	if result.RowsAffected == 0 {
		response.NotFound(w, r, "notification not found")
		return
	}

	response.JSON(w, http.StatusOK, models.SuccessResponse{Message: "notification deleted"})
}

// CreateNotificationRequest represents a request to create a notification (admin only)
type CreateNotificationRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Message string `json:"message"`
	Link    string `json:"link"`
}

// CreateNotification creates a notification for a user (admin only)
// @Summary Create notification (admin)
// @Description Creates a notification for a specific user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notification body CreateNotificationRequest true "Notification data"
// @Success 201 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/notifications [post]
func CreateNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "invalid request body")
		return
	}

	if req.Title == "" {
		response.BadRequest(w, r, "title is required")
		return
	}

	notification := models.Notification{
		UserID:  req.UserID,
		Type:    req.Type,
		Title:   req.Title,
		Message: req.Message,
		Link:    req.Link,
	}

	if err := database.DB.WithContext(ctx).Create(&notification).Error; err != nil {
		log.Error().Err(err).Uint("user_id", req.UserID).Msg("failed to create notification")
		response.HandleErrorWithMessage(w, r, err, "failed to create notification")
		return
	}

	response.JSON(w, http.StatusCreated, models.NotificationResponse{
		ID:        notification.ID,
		CreatedAt: notification.CreatedAt,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Link:      notification.Link,
		Read:      notification.Read,
		ReadAt:    notification.ReadAt,
	})
}

// CreateNotificationForUser is a helper function to create notifications from other parts of the app
func CreateNotificationForUser(userID uint, notificationType, title, message, link string) error {
	notification := models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Link:    link,
	}
	return database.DB.Create(&notification).Error
}
