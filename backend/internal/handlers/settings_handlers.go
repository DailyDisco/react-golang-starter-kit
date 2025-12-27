package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/services"

	"github.com/go-chi/chi/v5"
)

var settingsService = services.NewSettingsService()
var healthService = services.NewHealthService()

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
	settings, err := settingsService.GetAllSettings()
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

	settings, err := settingsService.GetSettingsByCategory(category)
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

	if err := settingsService.UpdateSetting(key, req.Value); err != nil {
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
	settings, err := settingsService.GetEmailSettings()
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

	if err := settingsService.UpdateEmailSettings(&settings); err != nil {
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
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/settings/email/test [post]
func TestEmailSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement email sending test
	// This would use the email service to send a test email to the admin

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Test email sent successfully (not implemented)",
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
	settings, err := settingsService.GetSecuritySettings()
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

	if err := settingsService.UpdateSecuritySettings(&settings); err != nil {
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
	settings, err := settingsService.GetSiteSettings()
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

	if err := settingsService.UpdateSiteSettings(&settings); err != nil {
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
	blocks, err := settingsService.GetIPBlocklist()
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

	block, err := settingsService.BlockIP(&req, userID)
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

	if err := settingsService.UnblockIP(uint(id)); err != nil {
		if err.Error() == "IP block not found" {
			WriteNotFound(w, r, "IP block not found")
			return
		}
		WriteInternalError(w, r, err.Error())
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
	announcements, err := settingsService.GetAnnouncements()
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve announcements")
		return
	}

	responses := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = models.AnnouncementBannerResponse{
			ID:            a.ID,
			Title:         a.Title,
			Message:       a.Message,
			Type:          a.Type,
			LinkURL:       a.LinkURL,
			LinkText:      a.LinkText,
			IsDismissible: a.IsDismissible,
			Priority:      a.Priority,
			IsActive:      a.IsActive,
			TargetRoles:   a.TargetRoles,
		}
		if a.StartsAt != nil {
			responses[i].StartsAt = *a.StartsAt
		}
		if a.EndsAt != nil {
			responses[i].EndsAt = *a.EndsAt
		}
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

	announcements, err := settingsService.GetActiveAnnouncements(userID, userRole)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve announcements")
		return
	}

	responses := make([]models.AnnouncementBannerResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = models.AnnouncementBannerResponse{
			ID:            a.ID,
			Title:         a.Title,
			Message:       a.Message,
			Type:          a.Type,
			LinkURL:       a.LinkURL,
			LinkText:      a.LinkText,
			IsDismissible: a.IsDismissible,
			Priority:      a.Priority,
		}
	}

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
	announcement, err := settingsService.CreateAnnouncement(&req, userID)
	if err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	response := models.AnnouncementBannerResponse{
		ID:            announcement.ID,
		Title:         announcement.Title,
		Message:       announcement.Message,
		Type:          announcement.Type,
		LinkURL:       announcement.LinkURL,
		LinkText:      announcement.LinkText,
		IsDismissible: announcement.IsDismissible,
		Priority:      announcement.Priority,
		IsActive:      announcement.IsActive,
		TargetRoles:   announcement.TargetRoles,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
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

	announcement, err := settingsService.UpdateAnnouncement(uint(id), &req)
	if err != nil {
		if err.Error() == "announcement not found" {
			WriteNotFound(w, r, "Announcement not found")
			return
		}
		WriteInternalError(w, r, err.Error())
		return
	}

	response := models.AnnouncementBannerResponse{
		ID:            announcement.ID,
		Title:         announcement.Title,
		Message:       announcement.Message,
		Type:          announcement.Type,
		LinkURL:       announcement.LinkURL,
		LinkText:      announcement.LinkText,
		IsDismissible: announcement.IsDismissible,
		Priority:      announcement.Priority,
		IsActive:      announcement.IsActive,
		TargetRoles:   announcement.TargetRoles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	if err := settingsService.DeleteAnnouncement(uint(id)); err != nil {
		if err.Error() == "announcement not found" {
			WriteNotFound(w, r, "Announcement not found")
			return
		}
		WriteInternalError(w, r, err.Error())
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

	if err := settingsService.DismissAnnouncement(userID, uint(announcementID)); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Announcement dismissed",
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
	templates, err := settingsService.GetEmailTemplates()
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

	template, err := settingsService.GetEmailTemplate(uint(id))
	if err != nil {
		if err.Error() == "email template not found" {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, err.Error())
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
	template, err := settingsService.UpdateEmailTemplate(uint(id), &req, userID)
	if err != nil {
		if err.Error() == "email template not found" {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, err.Error())
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

	template, err := settingsService.GetEmailTemplate(uint(id))
	if err != nil {
		if err.Error() == "email template not found" {
			WriteNotFound(w, r, "Email template not found")
			return
		}
		WriteInternalError(w, r, err.Error())
		return
	}

	// Render template with provided variables
	subject := template.Subject
	bodyHTML := template.BodyHTML
	bodyText := template.BodyText

	for key, value := range req.Variables {
		placeholder := "{{" + key + "}}"
		subject = replaceAll(subject, placeholder, value)
		bodyHTML = replaceAll(bodyHTML, placeholder, value)
		bodyText = replaceAll(bodyText, placeholder, value)
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

// Helper functions

func replaceAll(s, old, new string) string {
	result := s
	for {
		idx := indexOf(result, old)
		if idx == -1 {
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
