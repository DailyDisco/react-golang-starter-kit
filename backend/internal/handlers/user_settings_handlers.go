package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/services"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ensure auth package is imported for context key access
var _ = auth.UserContextKey

var sessionService = services.NewSessionService()
var totpService = services.NewTOTPService()
var userPrefsService = services.NewUserPreferencesService()

// ============ User Preferences Handlers ============

// GetUserPreferences returns user preferences
// @Summary Get user preferences
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.UserPreferencesResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/preferences [get]
func GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	prefs, err := userPrefsService.GetPreferences(userID)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve preferences")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs.ToResponse())
}

// UpdateUserPreferences updates user preferences
// @Summary Update user preferences
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.UpdateUserPreferencesRequest true "Preferences to update"
// @Success 200 {object} models.UserPreferencesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/preferences [put]
func UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.UpdateUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	prefs, err := userPrefsService.UpdatePreferences(userID, &req)
	if err != nil {
		WriteBadRequest(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs.ToResponse())
}

// ============ Session Management Handlers ============

// GetUserSessions returns all active sessions for current user
// @Summary Get user sessions
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {array} models.UserSessionResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/sessions [get]
func GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// Get current token hash to mark current session
	currentTokenHash := ""
	if token := getRefreshTokenFromContext(r); token != "" {
		currentTokenHash = services.HashToken(token)
	}

	sessions, err := sessionService.GetUserSessions(userID, currentTokenHash)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve sessions")
		return
	}

	responses := make([]models.UserSessionResponse, len(sessions))
	for i, s := range sessions {
		responses[i] = s.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// RevokeSession revokes a specific session
// @Summary Revoke a session
// @Tags User Settings
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/users/me/sessions/{id} [delete]
func RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	sessionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid session ID")
		return
	}

	if err := sessionService.RevokeSession(userID, uint(sessionID)); err != nil {
		if err.Error() == "session not found" {
			WriteNotFound(w, r, "Session not found")
			return
		}
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions except current
// @Summary Revoke all other sessions
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/sessions [delete]
func RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// Keep current session
	currentTokenHash := ""
	if token := getRefreshTokenFromContext(r); token != "" {
		currentTokenHash = services.HashToken(token)
	}

	if err := sessionService.RevokeAllSessions(userID, currentTokenHash); err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "All other sessions revoked successfully",
	})
}

// ============ Login History Handlers ============

// GetLoginHistory returns login history for current user
// @Summary Get login history
// @Tags User Settings
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/login-history [get]
func GetLoginHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	history, total, err := sessionService.GetLoginHistory(userID, limit, offset)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve login history")
		return
	}

	responses := make([]models.LoginHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = h.ToResponse()
	}

	totalPages := (int(total) + limit - 1) / limit

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history":     responses,
		"count":       len(responses),
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ============ Password Change Handlers ============

// ChangePassword changes the user's password
// @Summary Change password
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.ChangePasswordRequest true "Password change request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/password [put]
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate passwords match
	if req.NewPassword != req.ConfirmPassword {
		WriteBadRequest(w, r, "Passwords do not match")
		return
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		WriteBadRequest(w, r, "Password must be at least 8 characters")
		return
	}

	// Get user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve user")
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		WriteBadRequest(w, r, "Current password is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		WriteInternalError(w, r, "Failed to hash password")
		return
	}

	// Update password
	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		WriteInternalError(w, r, "Failed to update password")
		return
	}

	// Revoke all other sessions (security best practice)
	currentTokenHash := ""
	if token := getRefreshTokenFromContext(r); token != "" {
		currentTokenHash = services.HashToken(token)
	}
	sessionService.RevokeAllSessions(userID, currentTokenHash)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Password changed successfully. Other sessions have been logged out.",
	})
}

// ============ Two-Factor Authentication Handlers ============

// Get2FAStatus returns 2FA status for current user
// @Summary Get 2FA status
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.TwoFactorStatusResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/2fa/status [get]
func Get2FAStatus(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	status, err := totpService.GetTwoFactorStatus(userID)
	if err != nil {
		WriteInternalError(w, r, "Failed to retrieve 2FA status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Setup2FA initiates 2FA setup
// @Summary Setup 2FA
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.TwoFactorSetupResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/2fa/setup [post]
func Setup2FA(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// Get user email
	email := getUserEmailFromContext(r)
	if email == "" {
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			WriteInternalError(w, r, "Failed to retrieve user")
			return
		}
		email = user.Email
	}

	setup, err := totpService.SetupTwoFactor(userID, email)
	if err != nil {
		WriteInternalError(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setup)
}

// Verify2FA verifies and enables 2FA
// @Summary Verify and enable 2FA
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.TwoFactorVerifyRequest true "Verification code"
// @Success 200 {object} models.TwoFactorBackupCodesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/2fa/verify [post]
func Verify2FA(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if len(req.Code) != 6 {
		WriteBadRequest(w, r, "Code must be 6 digits")
		return
	}

	backupCodes, err := totpService.VerifyAndEnable(userID, req.Code)
	if err != nil {
		WriteBadRequest(w, r, err.Error())
		return
	}

	// Format backup codes for display
	formattedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		formattedCodes[i] = services.FormatBackupCode(code)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TwoFactorBackupCodesResponse{
		BackupCodes: formattedCodes,
		Message:     "Two-factor authentication has been enabled. Please save these backup codes in a secure location.",
	})
}

// Disable2FA disables 2FA
// @Summary Disable 2FA
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.TwoFactorVerifyRequest true "Verification code"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/2fa/disable [post]
func Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Unformat backup code if needed
	code := services.UnformatBackupCode(req.Code)

	if err := totpService.DisableTwoFactor(userID, code); err != nil {
		WriteBadRequest(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Two-factor authentication has been disabled",
	})
}

// RegenerateBackupCodes generates new backup codes
// @Summary Regenerate backup codes
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.TwoFactorVerifyRequest true "Verification code"
// @Success 200 {object} models.TwoFactorBackupCodesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/2fa/backup-codes [post]
func RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	backupCodes, err := totpService.RegenerateBackupCodes(userID, req.Code)
	if err != nil {
		WriteBadRequest(w, r, err.Error())
		return
	}

	// Format backup codes for display
	formattedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		formattedCodes[i] = services.FormatBackupCode(code)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TwoFactorBackupCodesResponse{
		BackupCodes: formattedCodes,
		Message:     "New backup codes have been generated. Your old codes are no longer valid.",
	})
}

// ============ Account Deletion Handlers ============

// RequestAccountDeletion initiates account deletion
// @Summary Request account deletion
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.RequestAccountDeletionRequest true "Deletion request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/delete [post]
func RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.RequestAccountDeletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Get user and verify password
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		WriteBadRequest(w, r, "Incorrect password")
		return
	}

	// Schedule deletion for 14 days from now (grace period)
	// In production, you would also send a confirmation email
	// and create a background job for the actual deletion
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Account deletion requested. Your account will be deleted in 14 days. You can cancel this request by logging in before then.",
	})
}

// CancelAccountDeletion cancels a pending account deletion
// @Summary Cancel account deletion
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/delete/cancel [post]
func CancelAccountDeletion(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// Clear deletion request
	// In production, this would clear the deletion_scheduled_at field
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Account deletion has been cancelled",
	})
}

// ============ Data Export Handlers ============

// RequestDataExport initiates a data export
// @Summary Request data export
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/export [post]
func RequestDataExport(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// In production, this would create a background job to generate the export
	// and send an email when it's ready
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Data export requested. You will receive an email when your export is ready for download.",
	})
}

// ============ Helper Functions ============

func getUserIDFromContext(r *http.Request) uint {
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		return 0
	}
	claims, ok := userCtx.(*auth.Claims)
	if !ok {
		return 0
	}
	return claims.UserID
}

func getUserRoleFromContext(r *http.Request) string {
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		return ""
	}
	claims, ok := userCtx.(*auth.Claims)
	if !ok {
		return ""
	}
	return claims.Role
}

func getUserEmailFromContext(r *http.Request) string {
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		return ""
	}
	claims, ok := userCtx.(*auth.Claims)
	if !ok {
		return ""
	}
	return claims.Email
}

func getRefreshTokenFromContext(r *http.Request) string {
	// This would typically be stored in the context by the auth middleware
	// For now, return empty string
	return ""
}
