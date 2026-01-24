package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/email"
	"react-golang-starter/internal/jobs"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/response"
	"react-golang-starter/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var settingsService *services.SettingsService
var healthService = services.NewHealthService()

// InitSettingsHandlers initializes the settings handlers with their dependencies.
// This must be called after database.DB is initialized.
func InitSettingsHandlers(db *gorm.DB) {
	settingsService = services.NewSettingsService(db)
}

// ============ System Settings Handlers ============

// GetAllSettings returns all system settings
// @Summary Get all system settings
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {array} models.SystemSettingResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings [get]
func GetAllSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settingsService.GetAllSettings(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve settings")
		return
	}

	// Convert to response format (hide sensitive values)
	responses := make([]models.SystemSettingResponse, len(settings))
	for i, s := range settings {
		responses[i] = s.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetSettingsByCategory returns settings for a specific category
// @Summary Get settings by category
// @Tags Admin Settings
// @Security BearerAuth
// @Param category path string true "Settings category"
// @Success 200 {array} models.SystemSettingResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/{category} [get]
func GetSettingsByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		WriteBadRequest(w, r, "Category is required")
		return
	}

	settings, err := settingsService.GetSettingsByCategory(r.Context(), category)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve settings")
		return
	}

	responses := make([]models.SystemSettingResponse, len(settings))
	for i, s := range settings {
		responses[i] = s.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UpdateSetting updates a single setting
// @Summary Update a setting
// @Tags Admin Settings
// @Security BearerAuth
// @Param key path string true "Setting key"
// @Param body body models.UpdateSystemSettingRequest true "Setting value"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/{key} [put]
func UpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		WriteBadRequest(w, r, "Setting key is required")
		return
	}

	var req models.UpdateSystemSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := settingsService.UpdateSettingWithCache(r.Context(), key, req.Value); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Setting updated successfully",
	})
}

// ============ Email Settings Handlers ============

// GetEmailSettings returns email/SMTP configuration
// @Summary Get email settings
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} models.EmailSettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/email [get]
func GetEmailSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settingsService.GetEmailSettings(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve email settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateEmailSettings updates email/SMTP configuration
// @Summary Update email settings
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.EmailSettings true "Email settings"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/email [put]
func UpdateEmailSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.EmailSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := settingsService.UpdateEmailSettings(r.Context(), &settings); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Email settings updated successfully",
	})
}

// TestEmailSettings sends a test email
// @Summary Send test email
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.TestEmailRequest true "Test email request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/settings/email/test [post]
func TestEmailSettings(w http.ResponseWriter, r *http.Request) {
	// Check if email service is available
	if !email.IsAvailable() {
		WriteBadRequest(w, r, "Email service is not configured or disabled")
		return
	}

	// Parse request body for recipient email
	var req models.TestEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Default to requesting user's email if not specified
	recipientEmail := req.RecipientEmail
	if recipientEmail == "" {
		// Get user from context
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if ok && claims != nil && claims.Email != "" {
			recipientEmail = claims.Email
		} else {
			WriteBadRequest(w, r, "Recipient email is required")
			return
		}
	}

	// Send test email
	err := email.Send(r.Context(), email.SendParams{
		To:           recipientEmail,
		Subject:      "Test Email - Email Configuration Verified",
		TemplateName: "",
		PlainText:    "This is a test email to verify your email configuration is working correctly.\n\nIf you received this email, your SMTP settings are configured properly.",
		Data: map[string]interface{}{
			"recipient": recipientEmail,
			"timestamp": time.Now().Format(time.RFC1123),
		},
	})

	if err != nil {
		log.Error().Err(err).Str("recipient", recipientEmail).Msg("failed to send test email")
		WriteInternalError(w, r, "Failed to send test email. Please check your SMTP configuration.")
		return
	}

	log.Info().Str("recipient", recipientEmail).Msg("test email sent successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Test email sent successfully to " + recipientEmail,
	})
}

// ============ Security Settings Handlers ============

// GetSecuritySettings returns security configuration
// @Summary Get security settings
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} models.SecuritySettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/security [get]
func GetSecuritySettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settingsService.GetSecuritySettings(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve security settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSecuritySettings updates security configuration
// @Summary Update security settings
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.SecuritySettings true "Security settings"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/security [put]
func UpdateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SecuritySettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := settingsService.UpdateSecuritySettings(r.Context(), &settings); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Security settings updated successfully",
	})
}

// ============ Site Settings Handlers ============

// GetSiteSettings returns site configuration
// @Summary Get site settings
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} models.SiteSettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/site [get]
func GetSiteSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settingsService.GetSiteSettings(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve site settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSiteSettings updates site configuration
// @Summary Update site settings
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.SiteSettings true "Site settings"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/site [put]
func UpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SiteSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := settingsService.UpdateSiteSettings(r.Context(), &settings); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Site settings updated successfully",
	})
}

// ============ IP Blocklist Handlers ============

// GetIPBlocklist returns the IP blocklist
// @Summary Get IP blocklist
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {array} models.IPBlocklistResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/ip-blocklist [get]
func GetIPBlocklist(w http.ResponseWriter, r *http.Request) {
	blocks, err := settingsService.GetIPBlocklist(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve IP blocklist")
		return
	}

	responses := make([]models.IPBlocklistResponse, len(blocks))
	for i, b := range blocks {
		responses[i] = models.IPBlocklistResponse{
			ID:        b.ID,
			IPAddress: b.IPAddress,
			IPRange:   b.IPRange,
			Reason:    b.Reason,
			BlockType: b.BlockType,
			HitCount:  b.HitCount,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
		}
		if b.ExpiresAt != nil {
			responses[i].ExpiresAt = *b.ExpiresAt
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// BlockIP adds an IP to the blocklist
// @Summary Block an IP
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.CreateIPBlockRequest true "IP block request"
// @Success 201 {object} models.IPBlocklistResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/ip-blocklist [post]
func BlockIP(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIPBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if req.IPAddress == "" {
		WriteBadRequest(w, r, "IP address is required")
		return
	}

	// Get admin user ID from context
	userID := getUserIDFromContext(r)

	block, err := settingsService.BlockIP(r.Context(), &req, userID)
	if err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	response := models.IPBlocklistResponse{
		ID:        block.ID,
		IPAddress: block.IPAddress,
		IPRange:   block.IPRange,
		Reason:    block.Reason,
		BlockType: block.BlockType,
		HitCount:  block.HitCount,
		IsActive:  block.IsActive,
		CreatedAt: block.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UnblockIP removes an IP from the blocklist
// @Summary Unblock an IP
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Block ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/ip-blocklist/{id} [delete]
func UnblockIP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	if err := settingsService.UnblockIP(r.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrIPBlockNotFound) {
			WriteNotFound(w, r, "IP block not found")
			return
		}
		WriteInternalError(w, r, "Failed to unblock IP")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "IP unblocked successfully",
	})
}

// ============ Announcement Handlers ============

// GetAnnouncements returns all announcements (for admin)
// @Summary Get all announcements
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {array} models.AnnouncementBannerResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/announcements [get]
func GetAnnouncements(w http.ResponseWriter, r *http.Request) {
	announcements, err := settingsService.GetAnnouncements(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve announcements")
		return
	}

	responses := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = a.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetActiveAnnouncements returns active announcements for current user
// @Summary Get active announcements
// @Tags Announcements
// @Success 200 {array} models.AnnouncementBannerResponse
// @Router /api/announcements [get]
func GetActiveAnnouncements(w http.ResponseWriter, r *http.Request) {
	var userID *uint
	userRole := ""

	// Get user info if authenticated
	if uid := getUserIDFromContext(r); uid != 0 {
		userID = &uid
		userRole = getUserRoleFromContext(r)
	}

	announcements, err := settingsService.GetActiveAnnouncements(r.Context(), userID, userRole)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve announcements")
		return
	}

	responses := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = a.ToResponse()
	}

	// Set cache headers - private since user-specific, 1 minute
	response.SetCachePrivate(w, 60)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// CreateAnnouncement creates a new announcement
// @Summary Create announcement
// @Tags Admin Settings
// @Security BearerAuth
// @Param body body models.CreateAnnouncementRequest true "Announcement data"
// @Success 201 {object} models.AnnouncementBannerResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/announcements [post]
func CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if req.Title == "" || req.Message == "" {
		WriteBadRequest(w, r, "Title and message are required")
		return
	}

	userID := getUserIDFromContext(r)
	announcement, err := settingsService.CreateAnnouncement(r.Context(), &req, userID)
	if err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	// Queue emails if send_email is true and the announcement is active
	if req.SendEmail && announcement.IsActive {
		go enqueueAnnouncementEmails(r.Context(), announcement)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(announcement.ToResponse())
}

// enqueueAnnouncementEmails queues announcement emails to eligible users
func enqueueAnnouncementEmails(ctx context.Context, announcement *models.AnnouncementBanner) {
	// Get eligible users (respects email preferences and email_verified)
	users, err := settingsService.GetUsersForAnnouncementEmail(ctx, announcement.ID, announcement.TargetRoles)
	if err != nil {
		log.Error().Err(err).
			Uint("announcement_id", announcement.ID).
			Msg("failed to get users for announcement email")
		return
	}

	if len(users) == 0 {
		log.Info().
			Uint("announcement_id", announcement.ID).
			Msg("no eligible users for announcement email")
		return
	}

	log.Info().
		Uint("announcement_id", announcement.ID).
		Int("user_count", len(users)).
		Msg("queueing announcement emails")

	// Queue email jobs for each user
	enqueued := 0
	for _, user := range users {
		err := jobs.EnqueueAnnouncementEmail(ctx, jobs.SendAnnouncementEmailArgs{
			AnnouncementID: announcement.ID,
			UserID:         user.ID,
			UserEmail:      user.Email,
			UserName:       user.Name,
			Title:          announcement.Title,
			Message:        announcement.Message,
			Category:       announcement.Category,
			LinkURL:        announcement.LinkURL,
			LinkText:       announcement.LinkText,
		})
		if err != nil {
			log.Error().Err(err).
				Uint("user_id", user.ID).
				Str("email", user.Email).
				Msg("failed to enqueue announcement email")
			continue
		}
		enqueued++
	}

	// Mark announcement as email sent
	if enqueued > 0 {
		if err := settingsService.MarkAnnouncementEmailSent(ctx, announcement.ID); err != nil {
			log.Error().Err(err).
				Uint("announcement_id", announcement.ID).
				Msg("failed to mark announcement email as sent")
		}
	}

	log.Info().
		Uint("announcement_id", announcement.ID).
		Int("enqueued", enqueued).
		Int("total_users", len(users)).
		Msg("announcement emails queued")
}

// UpdateAnnouncement updates an announcement
// @Summary Update announcement
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Param body body models.UpdateAnnouncementRequest true "Announcement data"
// @Success 200 {object} models.AnnouncementBannerResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/announcements/{id} [put]
func UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	var req models.UpdateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	announcement, err := settingsService.UpdateAnnouncement(r.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, services.ErrAnnouncementNotFound) {
			WriteNotFound(w, r, "Announcement not found")
			return
		}
		WriteInternalError(w, r, "Failed to update announcement")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcement.ToResponse())
}

// DeleteAnnouncement deletes an announcement
// @Summary Delete announcement
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/announcements/{id} [delete]
func DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	if err := settingsService.DeleteAnnouncement(r.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrAnnouncementNotFound) {
			WriteNotFound(w, r, "Announcement not found")
			return
		}
		WriteInternalError(w, r, "Failed to delete announcement")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Announcement deleted successfully",
	})
}

// DismissAnnouncement marks an announcement as dismissed for current user
// @Summary Dismiss announcement
// @Tags Announcements
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/announcements/{id}/dismiss [post]
func DismissAnnouncement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	announcementID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	if err := settingsService.DismissAnnouncement(r.Context(), userID, uint(announcementID)); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Announcement dismissed",
	})
}

// GetChangelog returns paginated changelog entries (public)
// @Summary Get changelog
// @Tags Announcements
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param category query string false "Filter by category (update, feature, bugfix)"
// @Success 200 {object} models.ChangelogResponse
// @Router /api/v1/changelog [get]
func GetChangelog(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	category := r.URL.Query().Get("category")

	changelog, err := settingsService.GetChangelog(r.Context(), page, limit, category)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve changelog")
		return
	}

	// Set cache headers for public caching
	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(changelog)
}

// GetUnreadModalAnnouncements returns unread modal announcements for current user
// @Summary Get unread modal announcements
// @Tags Announcements
// @Security BearerAuth
// @Success 200 {array} models.AnnouncementBannerResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/announcements/unread-modals [get]
func GetUnreadModalAnnouncements(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	userRole := getUserRoleFromContext(r)

	announcements, err := settingsService.GetUnreadModalAnnouncements(r.Context(), userID, userRole)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve unread announcements")
		return
	}

	responses := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = a.ToResponse()
	}

	// Set cache headers - private since user-specific, 1 minute
	response.SetCachePrivate(w, 60)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// MarkAnnouncementRead marks a modal announcement as read for current user
// @Summary Mark announcement as read
// @Tags Announcements
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/announcements/{id}/read [post]
func MarkAnnouncementRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	announcementID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	if err := settingsService.MarkAnnouncementRead(r.Context(), userID, uint(announcementID)); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Announcement marked as read",
	})
}

// ============ Email Template Handlers ============

// GetEmailTemplates returns all email templates
// @Summary Get all email templates
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {array} models.EmailTemplateResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/email-templates [get]
func GetEmailTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := settingsService.GetEmailTemplates(r.Context())
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve email templates")
		return
	}

	responses := make([]models.EmailTemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = t.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetEmailTemplate returns a single email template
// @Summary Get email template
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Success 200 {object} models.EmailTemplateResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/email-templates/{id} [get]
func GetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	template, err := settingsService.GetEmailTemplate(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrEmailTemplateNotFound) {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, "Failed to retrieve email template")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template.ToResponse())
}

// UpdateEmailTemplate updates an email template
// @Summary Update email template
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Param body body models.UpdateEmailTemplateRequest true "Template data"
// @Success 200 {object} models.EmailTemplateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/email-templates/{id} [put]
func UpdateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	var req models.UpdateEmailTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	userID := getUserIDFromContext(r)
	template, err := settingsService.UpdateEmailTemplate(r.Context(), uint(id), &req, userID)
	if err != nil {
		if errors.Is(err, services.ErrEmailTemplateNotFound) {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, "Failed to update email template")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template.ToResponse())
}

// PreviewEmailTemplate renders a preview of an email template
// @Summary Preview email template
// @Tags Admin Settings
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Param body body models.PreviewEmailTemplateRequest true "Preview variables"
// @Success 200 {object} models.PreviewEmailTemplateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/email-templates/{id}/preview [post]
func PreviewEmailTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid ID")
		return
	}

	var req models.PreviewEmailTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	template, err := settingsService.GetEmailTemplate(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrEmailTemplateNotFound) {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, "Failed to retrieve email template")
		return
	}

	// Render template with provided variables
	subject := template.Subject
	bodyHTML := template.BodyHTML
	bodyText := template.BodyText

	for key, value := range req.Variables {
		placeholder := "{{" + key + "}}"
		subject = strings.ReplaceAll(subject, placeholder, value)
		bodyHTML = strings.ReplaceAll(bodyHTML, placeholder, value)
		bodyText = strings.ReplaceAll(bodyText, placeholder, value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.PreviewEmailTemplateResponse{
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
	})
}

// ============ System Health Handlers ============

// GetSystemHealth returns overall system health
// @Summary Get system health
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} models.SystemHealthResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/health [get]
func GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health, err := healthService.GetSystemHealth()
	if err != nil {
		WriteInternalError(w, r, "Failed to get system health")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetDatabaseHealth returns detailed database health
// @Summary Get database health
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/health/database [get]
func GetDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	health, err := healthService.GetDetailedDatabaseHealth()
	if err != nil {
		WriteInternalError(w, r, "Failed to get database health")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetCacheHealth returns detailed cache health
// @Summary Get cache health
// @Tags Admin Settings
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/health/cache [get]
func GetCacheHealth(w http.ResponseWriter, r *http.Request) {
	health, err := healthService.GetDetailedCacheHealth()
	if err != nil {
		WriteInternalError(w, r, "Failed to get cache health")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
