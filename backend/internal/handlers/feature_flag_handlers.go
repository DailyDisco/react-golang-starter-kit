package handlers

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

// GetFeatureFlags returns all feature flags
// @Summary Get all feature flags
// @Tags Feature Flags
// @Security BearerAuth
// @Success 200 {object} models.FeatureFlagsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/feature-flags [get]
func GetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	var flags []models.FeatureFlag
	if err := database.DB.Order("key ASC").Find(&flags).Error; err != nil {
		WriteInternalError(w, r, "Failed to fetch feature flags")
		return
	}

	// Convert to response format
	flagResponses := make([]models.FeatureFlagResponse, len(flags))
	for i, flag := range flags {
		flagResponses[i] = toFeatureFlagResponse(flag)
	}

	response := models.FeatureFlagsResponse{
		Flags: flagResponses,
		Count: len(flagResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateFeatureFlag creates a new feature flag
// @Summary Create a feature flag
// @Tags Feature Flags
// @Security BearerAuth
// @Param body body models.CreateFeatureFlagRequest true "Feature flag data"
// @Success 201 {object} models.FeatureFlagResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/feature-flags [post]
func CreateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFeatureFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate key format (lowercase, underscores only)
	if !isValidFlagKey(req.Key) {
		WriteBadRequest(w, r, "Invalid key format. Use lowercase letters and underscores only.")
		return
	}

	// Validate rollout percentage
	if req.RolloutPercentage < 0 || req.RolloutPercentage > 100 {
		WriteBadRequest(w, r, "Rollout percentage must be between 0 and 100")
		return
	}

	now := time.Now().Format(time.RFC3339)
	flag := models.FeatureFlag{
		Key:               req.Key,
		Name:              req.Name,
		Description:       req.Description,
		Enabled:           req.Enabled,
		RolloutPercentage: req.RolloutPercentage,
		AllowedRoles:      pq.StringArray(req.AllowedRoles),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := database.DB.Create(&flag).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			WriteConflict(w, r, "Feature flag with this key already exists")
			return
		}
		WriteInternalError(w, r, "Failed to create feature flag")
		return
	}

	// Invalidate feature flags cache
	cache.InvalidateFeatureFlags(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toFeatureFlagResponse(flag))
}

// UpdateFeatureFlag updates a feature flag
// @Summary Update a feature flag
// @Tags Feature Flags
// @Security BearerAuth
// @Param key path string true "Feature flag key"
// @Param body body models.UpdateFeatureFlagRequest true "Feature flag data"
// @Success 200 {object} models.FeatureFlagResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/feature-flags/{key} [put]
func UpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	var flag models.FeatureFlag
	if err := database.DB.Where("key = ?", key).First(&flag).Error; err != nil {
		WriteNotFound(w, r, "Feature flag not found")
		return
	}

	var req models.UpdateFeatureFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Apply updates
	if req.Name != nil {
		flag.Name = *req.Name
	}
	if req.Description != nil {
		flag.Description = *req.Description
	}
	if req.Enabled != nil {
		flag.Enabled = *req.Enabled
	}
	if req.RolloutPercentage != nil {
		if *req.RolloutPercentage < 0 || *req.RolloutPercentage > 100 {
			WriteBadRequest(w, r, "Rollout percentage must be between 0 and 100")
			return
		}
		flag.RolloutPercentage = *req.RolloutPercentage
	}
	if req.AllowedRoles != nil {
		flag.AllowedRoles = pq.StringArray(*req.AllowedRoles)
	}
	flag.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&flag).Error; err != nil {
		WriteInternalError(w, r, "Failed to update feature flag")
		return
	}

	// Invalidate feature flags cache
	cache.InvalidateFeatureFlags(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toFeatureFlagResponse(flag))
}

// DeleteFeatureFlag deletes a feature flag
// @Summary Delete a feature flag
// @Tags Feature Flags
// @Security BearerAuth
// @Param key path string true "Feature flag key"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/feature-flags/{key} [delete]
func DeleteFeatureFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	result := database.DB.Where("key = ?", key).Delete(&models.FeatureFlag{})
	if result.Error != nil {
		WriteInternalError(w, r, "Failed to delete feature flag")
		return
	}
	if result.RowsAffected == 0 {
		WriteNotFound(w, r, "Feature flag not found")
		return
	}

	// Also delete user overrides
	database.DB.Where("feature_flag_id IN (SELECT id FROM feature_flags WHERE key = ?)", key).Delete(&models.UserFeatureFlag{})

	// Invalidate feature flags cache
	cache.InvalidateFeatureFlags(r.Context())

	resp := models.SuccessResponse{
		Success: true,
		Message: "Feature flag deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetFeatureFlagsForUser returns feature flags with their enabled status for the current user
// @Summary Get feature flags for current user
// @Tags Feature Flags
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 401 {object} models.ErrorResponse
// @Router /api/feature-flags [get]
func GetFeatureFlagsForUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		WriteUnauthorized(w, r, "Unauthorized")
		return
	}
	claims := userCtx.(*auth.Claims)

	// Get all feature flags
	var flags []models.FeatureFlag
	if err := database.DB.Find(&flags).Error; err != nil {
		WriteInternalError(w, r, "Failed to fetch feature flags")
		return
	}

	// Get user overrides
	var overrides []models.UserFeatureFlag
	database.DB.Where("user_id = ?", claims.UserID).Find(&overrides)

	// Build override map
	overrideMap := make(map[uint]bool)
	for _, override := range overrides {
		overrideMap[override.FeatureFlagID] = override.Enabled
	}

	// Evaluate flags for user
	result := make(map[string]bool)
	for _, flag := range flags {
		// Check for user override first
		if override, hasOverride := overrideMap[flag.ID]; hasOverride {
			result[flag.Key] = override
			continue
		}

		result[flag.Key] = isFeatureEnabledForUser(flag, claims.UserID, claims.Role)
	}

	// Set cache headers - private since user-specific, 5 minutes
	response.SetCachePrivate(w, 300)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// SetUserFeatureFlagOverride sets a feature flag override for a specific user
// @Summary Set user feature flag override
// @Tags Feature Flags
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Param key path string true "Feature flag key"
// @Param body body object{enabled=bool} true "Override value"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{userId}/feature-flags/{key} [put]
func SetUserFeatureFlagOverride(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid user ID")
		return
	}

	key := chi.URLParam(r, "key")

	// Find the feature flag
	var flag models.FeatureFlag
	if err := database.DB.Where("key = ?", key).First(&flag).Error; err != nil {
		WriteNotFound(w, r, "Feature flag not found")
		return
	}

	// Verify user exists
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		WriteNotFound(w, r, "User not found")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	now := time.Now().Format(time.RFC3339)
	override := models.UserFeatureFlag{
		UserID:        uint(userID),
		FeatureFlagID: flag.ID,
		Enabled:       req.Enabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Upsert the override
	if err := database.DB.Where("user_id = ? AND feature_flag_id = ?", userID, flag.ID).
		Assign(models.UserFeatureFlag{Enabled: req.Enabled, UpdatedAt: now}).
		FirstOrCreate(&override).Error; err != nil {
		WriteInternalError(w, r, "Failed to set override")
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Feature flag override set successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUserFeatureFlagOverride removes a feature flag override for a specific user
// @Summary Delete user feature flag override
// @Tags Feature Flags
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Param key path string true "Feature flag key"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/users/{userId}/feature-flags/{key} [delete]
func DeleteUserFeatureFlagOverride(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid user ID")
		return
	}

	key := chi.URLParam(r, "key")

	// Find the feature flag
	var flag models.FeatureFlag
	if err := database.DB.Where("key = ?", key).First(&flag).Error; err != nil {
		WriteNotFound(w, r, "Feature flag not found")
		return
	}

	result := database.DB.Where("user_id = ? AND feature_flag_id = ?", userID, flag.ID).Delete(&models.UserFeatureFlag{})
	if result.Error != nil {
		WriteInternalError(w, r, "Failed to delete override")
		return
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Feature flag override removed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func toFeatureFlagResponse(flag models.FeatureFlag) models.FeatureFlagResponse {
	return models.FeatureFlagResponse{
		ID:                flag.ID,
		Key:               flag.Key,
		Name:              flag.Name,
		Description:       flag.Description,
		Enabled:           flag.Enabled,
		RolloutPercentage: flag.RolloutPercentage,
		AllowedRoles:      []string(flag.AllowedRoles),
		CreatedAt:         flag.CreatedAt,
		UpdatedAt:         flag.UpdatedAt,
	}
}

func isValidFlagKey(key string) bool {
	if len(key) == 0 || len(key) > 100 {
		return false
	}
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || c == '_' || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func isFeatureEnabledForUser(flag models.FeatureFlag, userID uint, userRole string) bool {
	// If flag is disabled globally, return false
	if !flag.Enabled {
		return false
	}

	// Check if user's role is in allowed roles
	if len(flag.AllowedRoles) > 0 {
		for _, role := range flag.AllowedRoles {
			if role == userRole {
				return true
			}
		}
	}

	// Check rollout percentage
	if flag.RolloutPercentage >= 100 {
		return true
	}
	if flag.RolloutPercentage <= 0 {
		return false
	}

	// Use consistent hash based on user ID and flag key
	h := fnv.New32a()
	h.Write([]byte(flag.Key))
	h.Write([]byte(strconv.Itoa(int(userID))))
	hash := h.Sum32()

	return (hash % 100) < uint32(flag.RolloutPercentage)
}
