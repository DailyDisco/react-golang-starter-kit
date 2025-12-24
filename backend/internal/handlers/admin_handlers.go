package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"react-golang-starter/internal/audit"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// GetAdminStats returns admin dashboard statistics
// @Summary Get admin statistics
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} models.AdminStatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/stats [get]
func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	var stats models.AdminStatsResponse
	now := time.Now()
	today := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")

	// User stats
	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	database.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)
	database.DB.Model(&models.User{}).Where("email_verified = ?", true).Count(&stats.VerifiedUsers)
	database.DB.Model(&models.User{}).Where("DATE(created_at) = ?", today).Count(&stats.NewUsersToday)
	database.DB.Model(&models.User{}).Where("DATE(created_at) >= ?", weekAgo).Count(&stats.NewUsersThisWeek)
	database.DB.Model(&models.User{}).Where("DATE(created_at) >= ?", monthAgo).Count(&stats.NewUsersThisMonth)

	// Subscription stats
	database.DB.Model(&models.Subscription{}).Count(&stats.TotalSubscriptions)
	database.DB.Model(&models.Subscription{}).Where("status IN ?", []string{"active", "trialing"}).Count(&stats.ActiveSubscriptions)
	database.DB.Model(&models.Subscription{}).Where("status = ?", "canceled").Count(&stats.CanceledSubscriptions)

	// File stats
	database.DB.Model(&models.File{}).Count(&stats.TotalFiles)
	database.DB.Model(&models.File{}).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalFileSize)

	// Users by role
	stats.UsersByRole = make(map[string]int64)
	rows, err := database.DB.Model(&models.User{}).Select("role, COUNT(*) as count").Group("role").Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var role string
			var count int64
			if rows.Scan(&role, &count) == nil {
				stats.UsersByRole[role] = count
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetAuditLogs returns paginated audit logs
// @Summary Get audit logs
// @Tags Admin
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param target_type query string false "Filter by target type"
// @Param action query string false "Filter by action"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.AuditLogsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/audit-logs [get]
func GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	var filter models.AuditLogFilter

	// Parse query parameters
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			filter.UserID = &uid
		}
	}
	filter.TargetType = r.URL.Query().Get("target_type")
	if targetIDStr := r.URL.Query().Get("target_id"); targetIDStr != "" {
		if targetID, err := strconv.ParseUint(targetIDStr, 10, 32); err == nil {
			tid := uint(targetID)
			filter.TargetID = &tid
		}
	}
	filter.Action = r.URL.Query().Get("action")
	filter.StartDate = r.URL.Query().Get("start_date")
	filter.EndDate = r.URL.Query().Get("end_date")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	filter.Page = page

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	filter.Limit = limit

	logs, total, err := audit.GetAuditLogs(filter)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch audit logs"}`, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	logResponses := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToAuditLogResponse()
	}

	totalPages := (int(total) + limit - 1) / limit

	response := models.AuditLogsResponse{
		Logs:       logResponses,
		Count:      len(logResponses),
		Total:      int(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ImpersonateUser starts impersonating another user
// @Summary Start impersonating a user
// @Tags Admin
// @Security BearerAuth
// @Param body body models.ImpersonateRequest true "Impersonation request"
// @Success 200 {object} models.ImpersonateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/impersonate [post]
func ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only super_admin can impersonate
	if claims.Role != models.RoleSuperAdmin {
		http.Error(w, `{"error":"Only super admins can impersonate users"}`, http.StatusForbidden)
		return
	}

	// Already impersonating?
	if claims.OriginalUserID != 0 {
		http.Error(w, `{"error":"Already impersonating a user. Stop impersonation first."}`, http.StatusBadRequest)
		return
	}

	var req models.ImpersonateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Cannot impersonate yourself
	if req.UserID == claims.UserID {
		http.Error(w, `{"error":"Cannot impersonate yourself"}`, http.StatusBadRequest)
		return
	}

	// Find target user
	var targetUser models.User
	if err := database.DB.First(&targetUser, req.UserID).Error; err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Cannot impersonate another super_admin
	if targetUser.Role == models.RoleSuperAdmin {
		http.Error(w, `{"error":"Cannot impersonate other super admins"}`, http.StatusForbidden)
		return
	}

	// Generate impersonation token
	token, err := auth.GenerateImpersonationToken(&targetUser, claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Log the impersonation
	audit.LogImpersonate(claims.UserID, req.UserID, req.Reason, r)

	response := models.ImpersonateResponse{
		User:           targetUser.ToUserResponse(),
		Token:          token,
		OriginalUserID: claims.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopImpersonation stops impersonating and returns to original user
// @Summary Stop impersonating
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/admin/stop-impersonate [post]
func StopImpersonation(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Not impersonating?
	if claims.OriginalUserID == 0 {
		http.Error(w, `{"error":"Not currently impersonating anyone"}`, http.StatusBadRequest)
		return
	}

	// Find original admin user
	var adminUser models.User
	if err := database.DB.First(&adminUser, claims.OriginalUserID).Error; err != nil {
		http.Error(w, `{"error":"Original user not found"}`, http.StatusInternalServerError)
		return
	}

	// Generate new token for original user
	token, err := auth.GenerateToken(&adminUser)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Log stopping impersonation
	audit.LogStopImpersonate(claims.OriginalUserID, claims.UserID, r)

	response := models.AuthResponse{
		User:  adminUser.ToUserResponse(),
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminUpdateUserRole updates a user's role (admin version with audit logging)
// @Summary Update user role
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param body body object{role=string} true "New role"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/role [put]
func AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only super_admin can change roles
	if claims.Role != models.RoleSuperAdmin {
		http.Error(w, `{"error":"Only super admins can change user roles"}`, http.StatusForbidden)
		return
	}

	// Get user ID from path
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate role
	validRoles := map[string]bool{
		models.RoleSuperAdmin: true,
		models.RoleAdmin:      true,
		models.RolePremium:    true,
		models.RoleUser:       true,
	}
	if !validRoles[req.Role] {
		http.Error(w, `{"error":"Invalid role"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Cannot change own role
	if uint(userID) == claims.UserID {
		http.Error(w, `{"error":"Cannot change your own role"}`, http.StatusBadRequest)
		return
	}

	// Log the role change
	oldRole := user.Role
	if oldRole != req.Role {
		audit.LogRoleChange(claims.UserID, uint(userID), oldRole, req.Role, r)
	}

	// Update role
	user.Role = req.Role
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, `{"error":"Failed to update user role"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.ToUserResponse())
}

// DeactivateUser deactivates a user account
// @Summary Deactivate user
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/deactivate [post]
func DeactivateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only admin or super_admin can deactivate
	if claims.Role != models.RoleSuperAdmin && claims.Role != models.RoleAdmin {
		http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
		return
	}

	// Get user ID from path
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Cannot deactivate yourself
	if uint(userID) == claims.UserID {
		http.Error(w, `{"error":"Cannot deactivate yourself"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Admins cannot deactivate super_admins
	if claims.Role == models.RoleAdmin && user.Role == models.RoleSuperAdmin {
		http.Error(w, `{"error":"Cannot deactivate super admin"}`, http.StatusForbidden)
		return
	}

	// Deactivate
	user.IsActive = false
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, `{"error":"Failed to deactivate user"}`, http.StatusInternalServerError)
		return
	}

	// Log the action
	audit.LogUserUpdate(claims.UserID, uint(userID), map[string]interface{}{"is_active": false}, r)

	response := models.SuccessResponse{
		Success: true,
		Message: "User deactivated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReactivateUser reactivates a user account
// @Summary Reactivate user
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/reactivate [post]
func ReactivateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only admin or super_admin can reactivate
	if claims.Role != models.RoleSuperAdmin && claims.Role != models.RoleAdmin {
		http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
		return
	}

	// Get user ID from path
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Reactivate
	user.IsActive = true
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, `{"error":"Failed to reactivate user"}`, http.StatusInternalServerError)
		return
	}

	// Log the action
	audit.LogUserUpdate(claims.UserID, uint(userID), map[string]interface{}{"is_active": true}, r)

	response := models.SuccessResponse{
		Success: true,
		Message: "User reactivated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RestoreUser restores a soft-deleted user
// @Summary Restore deleted user
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/restore [post]
func RestoreUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only super_admin can restore deleted users
	if claims.Role != models.RoleSuperAdmin {
		http.Error(w, `{"error":"Only super admins can restore deleted users"}`, http.StatusForbidden)
		return
	}

	// Get user ID from path
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Find soft-deleted user (including deleted records)
	var user models.User
	if err := database.DB.Unscoped().First(&user, userID).Error; err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Check if user is actually deleted
	if !user.DeletedAt.Valid {
		http.Error(w, `{"error":"User is not deleted"}`, http.StatusBadRequest)
		return
	}

	// Restore user by setting DeletedAt to null
	if err := database.DB.Unscoped().Model(&user).Update("deleted_at", nil).Error; err != nil {
		http.Error(w, `{"error":"Failed to restore user"}`, http.StatusInternalServerError)
		return
	}

	// Log the action
	targetID := uint(userID)
	audit.LogEntry(&claims.UserID, models.AuditTargetUser, &targetID, "restore", nil, r)

	restoreResp := models.SuccessResponse{
		Success: true,
		Message: "User restored successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restoreResp)
}

// GetDeletedUsers returns a list of soft-deleted users
// @Summary Get deleted users
// @Tags Admin
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.UsersResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/users/deleted [get]
func GetDeletedUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	claims := userCtx.(*auth.Claims)

	// Only super_admin can view deleted users
	if claims.Role != models.RoleSuperAdmin {
		http.Error(w, `{"error":"Only super admins can view deleted users"}`, http.StatusForbidden)
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Count total deleted users
	var total int64
	database.DB.Unscoped().Model(&models.User{}).Where("deleted_at IS NOT NULL").Count(&total)

	// Get deleted users
	var users []models.User
	database.DB.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Offset(offset).Limit(limit).Find(&users)

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = u.ToUserResponse()
	}

	totalPages := (int(total) + limit - 1) / limit

	deletedResp := models.UsersResponse{
		Users:      userResponses,
		Count:      len(userResponses),
		Total:      int(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deletedResp)
}
