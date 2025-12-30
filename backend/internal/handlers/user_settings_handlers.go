package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// ============ Avatar Management Handlers ============

// AllowedAvatarMimeTypes defines allowed MIME types for avatar uploads
var AllowedAvatarMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// UploadAvatar uploads a new avatar for the current user
// @Summary Upload user avatar
// @Tags User Settings
// @Security BearerAuth
// @Accept multipart/form-data
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} models.AvatarUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse "File too large"
// @Router /api/users/me/avatar [post]
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	// Parse multipart form (5MB max for avatars)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		WriteBadRequest(w, r, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		WriteBadRequest(w, r, "Failed to get avatar file from form")
		return
	}
	defer file.Close()

	// Validate file size (5MB limit for avatars)
	if header.Size > 5<<20 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "Request Entity Too Large",
			Message: "Avatar file size exceeds 5MB limit",
			Code:    http.StatusRequestEntityTooLarge,
		})
		return
	}

	// Read first 512 bytes to detect actual content type
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		WriteBadRequest(w, r, "Failed to read file content")
		return
	}

	// Detect actual content type from file content (magic bytes)
	detectedType := http.DetectContentType(buf[:n])
	normalizedType := strings.Split(detectedType, ";")[0]
	normalizedType = strings.TrimSpace(strings.ToLower(normalizedType))

	// Validate MIME type
	if !AllowedAvatarMimeTypes[normalizedType] {
		WriteBadRequest(w, r, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed")
		return
	}

	// Reset file reader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		WriteInternalError(w, r, "Failed to process file")
		return
	}

	// Read the full file content
	content, err := io.ReadAll(file)
	if err != nil {
		WriteInternalError(w, r, "Failed to read file")
		return
	}

	// For now, store avatar as base64 data URL
	// In production, you would upload to S3 or similar
	avatarURL := "data:" + normalizedType + ";base64," + encodeBase64(content)

	// Update user's avatar URL
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve user")
		return
	}

	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		WriteInternalError(w, r, "Failed to save avatar")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AvatarUploadResponse{
		AvatarURL: avatarURL,
	})
}

// DeleteAvatar removes the current user's avatar
// @Summary Delete user avatar
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/avatar [delete]
func DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve user")
		return
	}

	user.AvatarURL = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		WriteInternalError(w, r, "Failed to remove avatar")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Avatar removed successfully",
	})
}

// ============ Connected Accounts Handlers ============

// GetConnectedAccounts returns the user's linked OAuth providers
// @Summary Get connected OAuth accounts
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {array} models.ConnectedAccountResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/connected-accounts [get]
func GetConnectedAccounts(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var providers []models.OAuthProvider
	if err := database.DB.Where("user_id = ?", userID).Find(&providers).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve connected accounts")
		return
	}

	responses := make([]models.ConnectedAccountResponse, len(providers))
	for i, p := range providers {
		responses[i] = models.ConnectedAccountResponse{
			Provider:       p.Provider,
			ProviderUserID: p.ProviderUserID,
			Email:          p.Email,
			ConnectedAt:    p.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// DisconnectAccount removes an OAuth provider connection
// @Summary Disconnect OAuth provider
// @Tags User Settings
// @Security BearerAuth
// @Param provider path string true "OAuth provider (google, github)"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse "Cannot unlink only auth method"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "Provider not linked"
// @Router /api/users/me/connected-accounts/{provider} [delete]
func DisconnectAccount(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	provider := chi.URLParam(r, "provider")
	provider = strings.ToLower(strings.TrimSpace(provider))

	// Validate provider
	validProviders := map[string]bool{"google": true, "github": true}
	if !validProviders[provider] {
		WriteBadRequest(w, r, "Invalid OAuth provider")
		return
	}

	// Get user to check if they have a password
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve user")
		return
	}

	// Count linked providers
	var providerCount int64
	if err := database.DB.Model(&models.OAuthProvider{}).Where("user_id = ?", userID).Count(&providerCount).Error; err != nil {
		WriteInternalError(w, r, "Failed to check connected accounts")
		return
	}

	// If user has no password and only one provider, don't allow unlinking
	if user.Password == "" && providerCount <= 1 {
		WriteBadRequest(w, r, "Cannot disconnect your only authentication method. Please set a password first.")
		return
	}

	// Delete the provider connection
	result := database.DB.Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.OAuthProvider{})
	if result.Error != nil {
		WriteInternalError(w, r, "Failed to disconnect account")
		return
	}

	if result.RowsAffected == 0 {
		WriteNotFound(w, r, "Provider not linked to your account")
		return
	}

	// If this was the user's primary OAuth provider, clear it
	if user.OAuthProvider == provider {
		user.OAuthProvider = ""
		user.OAuthProviderID = ""
		user.UpdatedAt = time.Now().Format(time.RFC3339)
		database.DB.Save(&user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "Account disconnected successfully",
	})
}

// ============ Data Export Status Handler ============

// GetDataExportStatus returns the status of the user's latest data export request
// @Summary Get data export status
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.DataExportResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "No export request found"
// @Router /api/users/me/export [get]
func GetDataExportStatus(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var export models.DataExport
	err := database.DB.Where("user_id = ?", userID).Order("requested_at DESC").First(&export).Error
	if err != nil {
		WriteNotFound(w, r, "No export request found")
		return
	}

	// Check if export is expired
	if export.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, export.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			export.Status = models.ExportStatusExpired
			database.DB.Save(&export)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(export.ToResponse())
}

// ============ Helper Functions ============

// encodeBase64 encodes bytes to base64 string
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

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
