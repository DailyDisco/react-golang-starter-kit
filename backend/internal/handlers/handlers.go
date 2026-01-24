// Package handlers contains HTTP request handlers for the API
//
// # React Go Starter Kit API
//
// A comprehensive REST API for the React Go Starter Kit application built with Fiber, GORM, and PostgreSQL.
// This API provides secure user authentication, user management, and system health monitoring.
//
// ## Features
//
// - **User Authentication**: JWT-based authentication with email verification
// - **User Management**: Complete CRUD operations for user accounts
// - **Password Security**: Secure password hashing and reset functionality
// - **Rate Limiting**: Built-in protection against abuse
// - **Health Monitoring**: System health checks and status endpoints
//
// ## Authentication
//
// Most endpoints require JWT Bearer token authentication. Obtain a token by logging in
// and include it in the Authorization header: `Authorization: Bearer {token}`
//
// ## Rate Limiting
//
// API endpoints are protected by rate limiting to prevent abuse. Different endpoints
// have different rate limits based on their sensitivity.
//
// Terms Of Service: https://github.com/your-org/react-golang-starter-kit
//
// Schemes: http, https
// Host: localhost:8080
// BasePath: /api
// Version: 1.0.0
// Contact:
//
//	name: API Support
//	url: https://github.com/your-org/react-golang-starter-kit/issues
//	email: support@example.com
//
// License:
//
//	name: MIT
//	url: https://opensource.org/licenses/MIT
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// SecurityDefinitions:
// BearerAuth:
//
//	type: apiKey
//	name: Authorization
//	in: header
//	description: |
//	  JWT Authorization header using the Bearer scheme.
//
//	  Format: `Authorization: Bearer {token}`
//
//	  To obtain a token:
//	  1. Register a new user account via POST /api/auth/register
//	  2. Login via POST /api/auth/login to receive a JWT token
//	  3. Include the token in all subsequent requests
//
// swagger:meta
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/jobs"
	"react-golang-starter/internal/models"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// Cache key prefixes
const (
	userCacheKeyPrefix = "user:"
	userCacheTTL       = 2 * time.Minute
)

// getUserCacheKey generates a cache key for a user by ID
func getUserCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", userCacheKeyPrefix, userID)
}

// getCachedUser attempts to retrieve a user from cache
func getCachedUser(ctx context.Context, userID uint) (*models.UserResponse, bool) {
	if !cache.IsAvailable() {
		return nil, false
	}

	var userResponse models.UserResponse
	err := cache.GetJSON(ctx, getUserCacheKey(userID), &userResponse)
	if err != nil {
		return nil, false
	}
	return &userResponse, true
}

// cacheUser stores a user response in the cache
func cacheUser(ctx context.Context, userID uint, userResponse *models.UserResponse) {
	if !cache.IsAvailable() {
		return
	}
	// Log cache errors for debugging - caching is best-effort
	if err := cache.SetJSON(ctx, getUserCacheKey(userID), userResponse, userCacheTTL); err != nil {
		log.Warn().Err(err).Uint("userID", userID).Msg("cache set failed")
	}
}

// invalidateUserCache removes a user from the cache
func invalidateUserCache(ctx context.Context, userID uint) {
	if !cache.IsAvailable() {
		return
	}
	if err := cache.Delete(ctx, getUserCacheKey(userID)); err != nil {
		log.Warn().Err(err).Uint("userID", userID).Msg("cache delete failed")
	}
}

// Build-time variables (set via ldflags)
var (
	Version   = "1.0.0"
	BuildTime = ""
	GitCommit = ""
)

// Application start time for uptime calculation
var startTime = time.Now()

// Service represents the application service with its dependencies
type Service struct {
}

// NewService creates a new Service instance
func NewService() *Service {
	return &Service{}
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration to human readable format
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// HealthCheck godoc
// @Summary Check server health status
// @Description Get the health status of the server and its dependencies (database, cache, jobs)
// @Tags health
// @Accept json
// @Produce json
// @Param verbose query bool false "Include runtime information"
// @Success 200 {object} models.HealthStatus "Server and its dependencies are healthy"
// @Failure 503 {object} models.HealthStatus "Server or one of its critical dependencies is unhealthy"
// @Router /health [get]
func (s *Service) HealthCheck(w http.ResponseWriter, r *http.Request) {
	overallStatus := "healthy"
	statusCode := http.StatusOK

	// Check all components
	dbStatus := database.CheckDatabaseHealth()
	cacheStatus := cache.CheckCacheHealth()

	components := []models.ComponentStatus{
		dbStatus,
		cacheStatus,
	}

	// Check job queue status if available
	if jobs.IsAvailable() {
		jobStatus := models.ComponentStatus{
			Name:    "jobs",
			Status:  "healthy",
			Message: "Job queue is running",
		}
		components = append(components, jobStatus)
	}

	// Determine overall status based on component health
	for _, comp := range components {
		if comp.Status == "unhealthy" {
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			break
		}
		// Degraded doesn't make overall unhealthy, but note it
		if comp.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	// Build version info
	versionInfo := models.VersionInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	}

	// Calculate uptime
	uptime := formatDuration(time.Since(startTime))

	healthResponse := models.HealthStatus{
		OverallStatus: overallStatus,
		Timestamp:     time.Now().Format(time.RFC3339),
		Uptime:        uptime,
		Version:       versionInfo,
		Components:    components,
	}

	// Include runtime info if verbose query param is set
	verbose := r.URL.Query().Get("verbose") == "true"
	if verbose {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		healthResponse.Runtime = &models.RuntimeInfo{
			Goroutines:  runtime.NumGoroutine(),
			MemoryAlloc: formatBytes(memStats.Alloc),
			MemorySys:   formatBytes(memStats.Sys),
			NumGC:       memStats.NumGC,
			GoVersion:   runtime.Version(),
			NumCPU:      runtime.NumCPU(),
			GOOS:        runtime.GOOS,
			GOARCH:      runtime.GOARCH,
		}
	}

	// Add environment info header
	if env := os.Getenv("GO_ENV"); env != "" {
		w.Header().Set("X-Environment", env)
	}

	WriteJSON(w, statusCode, healthResponse)
}

// ReadinessCheck godoc
// @Summary Check if server is ready to receive traffic
// @Description Fast readiness check for deployment orchestration. Verifies database connectivity.
// @Description Cache is checked but treated as non-critical (degraded, not unhealthy).
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Server is ready"
// @Failure 503 {object} map[string]string "Server is not ready (database unavailable)"
// @Router /health/ready [get]
func (s *Service) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := map[string]string{
		"status":   "healthy",
		"database": "healthy",
		"cache":    "healthy",
	}
	statusCode := http.StatusOK

	// Check database connectivity (critical)
	if database.DB == nil {
		response["status"] = "unhealthy"
		response["database"] = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else {
		sqlDB, err := database.DB.DB()
		if err != nil {
			response["status"] = "unhealthy"
			response["database"] = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else if err := sqlDB.PingContext(ctx); err != nil {
			response["status"] = "unhealthy"
			response["database"] = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}
	}

	// Check cache connectivity (non-critical - degraded only)
	if cache.IsAvailable() {
		cacheInstance := cache.Instance()
		if cacheInstance != nil {
			if err := cacheInstance.Ping(ctx); err != nil {
				response["cache"] = "degraded"
				if response["status"] == "healthy" {
					response["status"] = "degraded"
				}
			}
		}
	} else {
		response["cache"] = "unavailable"
	}

	WriteJSON(w, statusCode, response)
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a paginated list of all users. This endpoint is public and does not require authentication.
// @Description This endpoint is rate limited to prevent abuse. Maximum 30 requests per minute per IP.
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} models.SuccessResponse{data=models.UsersResponse} "List of users retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid pagination parameters"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} models.ErrorResponse "Failed to fetch users"
// @Router /users [get]
// @Example
//
// Request:
// GET /api/users?page=1&limit=10
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "Users retrieved successfully",
//	  "data": {
//	    "users": [
//	      {
//	        "id": 1,
//	        "name": "John Doe",
//	        "email": "john.doe@example.com",
//	        "email_verified": true,
//	        "is_active": true,
//	        "created_at": "2023-08-27T12:00:00Z",
//	        "updated_at": "2023-08-27T12:00:00Z"
//	      }
//	    ],
//	    "count": 1,
//	    "total": 25,
//	    "page": 1,
//	    "limit": 10,
//	    "total_pages": 3
//	  }
//	}
func GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		page := 1
		limit := 10

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			} else {
				WriteBadRequest(w, r, "Invalid page parameter")
				return
			}
		}

		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			} else {
				WriteBadRequest(w, r, "Invalid limit parameter (must be 1-100)")
				return
			}
		}

		// Get total count
		var total int64
		if err := database.DB.WithContext(r.Context()).Model(&models.User{}).Count(&total).Error; err != nil {
			WriteInternalError(w, r, "Failed to count users")
			return
		}

		// Calculate offset and total pages
		offset := (page - 1) * limit
		totalPages := int((total + int64(limit) - 1) / int64(limit))

		// Get paginated users
		var users []models.User
		if err := database.DB.WithContext(r.Context()).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
			WriteInternalError(w, r, "Failed to fetch users")
			return
		}

		// Convert to UserResponse to hide sensitive fields
		var userResponses []models.UserResponse
		for _, user := range users {
			userResponses = append(userResponses, user.ToUserResponse())
		}

		usersResponse := models.UsersResponse{
			Users:      userResponses,
			Count:      len(userResponses),
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		}

		WriteSuccess(w, "Users retrieved successfully", usersResponse)
	}
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Retrieve a single user by their ID. Users can only access their own profile unless they have admin privileges.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} models.ErrorResponse "Forbidden - can only access own profile"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve user"
// @Router /users/{id} [get]
// @Example
//
// Request:
// GET /api/users/1
// Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "User retrieved successfully",
//	  "data": {
//	    "id": 1,
//	    "name": "John Doe",
//	    "email": "john.doe@example.com",
//	    "email_verified": true,
//	    "is_active": true,
//	    "created_at": "2023-08-27T12:00:00Z",
//	    "updated_at": "2023-08-27T12:00:00Z"
//	  }
//	}
func GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			WriteBadRequest(w, r, "Invalid user ID")
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, r, "User not found in context")
			return
		}

		// Users can only view their own profile unless they're admins
		isAdmin := auth.HasRole(currentUser.Role, models.RoleAdmin, models.RoleSuperAdmin)
		if currentUser.ID != uint(userID) && !isAdmin {
			WriteForbidden(w, r, "You can only view your own profile")
			return
		}

		// Try to get user from cache first
		if cachedUser, found := getCachedUser(r.Context(), uint(userID)); found {
			w.Header().Set("X-Cache", "HIT")
			WriteSuccess(w, "User retrieved successfully", cachedUser)
			return
		}

		var user models.User
		if err := database.DB.WithContext(r.Context()).First(&user, userID).Error; err != nil {
			WriteNotFound(w, r, "User not found")
			return
		}

		userResponse := user.ToUserResponse()

		// Cache the user response for future requests
		cacheUser(r.Context(), uint(userID), &userResponse)

		w.Header().Set("X-Cache", "MISS")
		WriteSuccess(w, "User retrieved successfully", userResponse)
	}
}

// CreateUser godoc
// @Summary Create a new user (Super Admin only)
// @Description Create a new user with the provided information. Requires super_admin role (users:create permission).
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.SuccessResponse{data=models.UserResponse} "User created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid JSON or validation error"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - super_admin role required"
// @Failure 409 {object} models.ErrorResponse "User already exists"
// @Failure 500 {object} models.ErrorResponse "Failed to create user"
// @Router /users [post]
// @Example
//
// Request:
// POST /api/users
//
//	{
//	  "name": "John Doe",
//	  "email": "john.doe@example.com",
//	  "password": "SecurePass123!"
//	}
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "User created successfully",
//	  "data": {
//	    "id": 1,
//	    "name": "John Doe",
//	    "email": "john.doe@example.com",
//	    "email_verified": true,
//	    "is_active": true,
//	    "created_at": "2023-08-27T12:00:00Z",
//	    "updated_at": "2023-08-27T12:00:00Z"
//	  }
//	}
func CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Route is protected by PermissionMiddleware(PermCreateUsers) - only super_admin can access
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, r, "Invalid JSON")
			return
		}

		// Validate email format
		if err := auth.ValidateEmail(req.Email); err != nil {
			WriteBadRequest(w, r, err.Error())
			return
		}

		// Validate password strength
		if err := auth.ValidatePassword(req.Password); err != nil {
			WriteBadRequest(w, r, err.Error())
			return
		}

		// Check if user already exists
		var existingUser models.User
		if err := database.DB.WithContext(r.Context()).Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			WriteConflict(w, r, "User with this email already exists")
			return
		}

		// Hash password
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			WriteInternalError(w, r, "Failed to hash password")
			return
		}

		// Generate verification token
		verificationToken, err := auth.GenerateVerificationToken()
		if err != nil {
			WriteInternalError(w, r, "Failed to generate verification token")
			return
		}

		// Create user
		user := models.User{
			Name:                req.Name,
			Email:               req.Email,
			Password:            hashedPassword,
			VerificationToken:   &verificationToken,
			VerificationExpires: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			EmailVerified:       true, // Admin-created users are pre-verified
			IsActive:            true,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if err := database.DB.WithContext(r.Context()).Create(&user).Error; err != nil {
			WriteInternalError(w, r, "Failed to create user")
			return
		}

		userResponse := user.ToUserResponse()
		WriteJSON(w, http.StatusCreated, models.SuccessResponse{
			Success: true,
			Message: "User created successfully",
			Data:    userResponse,
		})
	}
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Update a user's information by their ID. Users can only update their own profile unless they have admin privileges.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Param user body object true "Updated user data (name and/or email)" '{"name":"string","email":"string"}'
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID or JSON"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - can only update own profile"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 409 {object} models.ErrorResponse "Email already taken"
// @Failure 500 {object} models.ErrorResponse "Failed to update user"
// @Router /users/{id} [put]
// @Example
//
// Request:
// PUT /api/users/1
// Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
//	{
//	  "name": "John Smith",
//	  "email": "john.smith@example.com"
//	}
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "User updated successfully",
//	  "data": {
//	    "id": 1,
//	    "name": "John Smith",
//	    "email": "john.smith@example.com",
//	    "email_verified": true,
//	    "is_active": true,
//	    "created_at": "2023-08-27T12:00:00Z",
//	    "updated_at": "2023-08-27T14:30:00Z"
//	  }
//	}
func UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			WriteBadRequest(w, r, "Invalid user ID")
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, r, "User not found in context")
			return
		}

		var user models.User
		if err := database.DB.WithContext(r.Context()).First(&user, userID).Error; err != nil {
			WriteNotFound(w, r, "User not found")
			return
		}

		// Users can only update their own profile unless they're admins
		isAdmin := auth.HasRole(currentUser.Role, models.RoleAdmin, models.RoleSuperAdmin)
		if currentUser.ID != uint(userID) && !isAdmin {
			WriteForbidden(w, r, "You can only update your own profile")
			return
		}

		var updateData struct {
			Name  string `json:"name,omitempty"`
			Email string `json:"email,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			WriteBadRequest(w, r, "Invalid JSON")
			return
		}

		// Validate email if provided
		if updateData.Email != "" {
			if err := auth.ValidateEmail(updateData.Email); err != nil {
				WriteBadRequest(w, r, err.Error())
				return
			}

			// Check if email is already taken by another user
			var existingUser models.User
			if err := database.DB.WithContext(r.Context()).Where("email = ? AND id != ?", updateData.Email, userID).First(&existingUser).Error; err == nil {
				WriteConflict(w, r, "Email is already taken")
				return
			}

			user.Email = updateData.Email
		}

		if updateData.Name != "" {
			// Validate name length
			trimmedName := strings.TrimSpace(updateData.Name)
			if len(trimmedName) == 0 {
				WriteBadRequest(w, r, "Name cannot be empty or whitespace only")
				return
			}
			if len(updateData.Name) > 255 {
				WriteBadRequest(w, r, "Name exceeds maximum length of 255 characters")
				return
			}
			user.Name = trimmedName
		}

		user.UpdatedAt = time.Now()

		if err := database.DB.WithContext(r.Context()).Save(&user).Error; err != nil {
			WriteInternalError(w, r, "Failed to update user")
			return
		}

		// Invalidate user cache after update
		invalidateUserCache(r.Context(), uint(userID))

		userResponse := user.ToUserResponse()
		WriteSuccess(w, "User updated successfully", userResponse)
	}
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by their ID. Users can only delete their own account unless they have admin privileges. This action is irreversible.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Security BearerAuth
// @Success 204 "No Content - User deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - can only delete own account"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse "Failed to delete user"
// @Router /users/{id} [delete]
// @Example
//
// Request:
// DELETE /api/users/1
// Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
// Response:
// HTTP/1.1 204 No Content
func DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			WriteBadRequest(w, r, "Invalid user ID")
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, r, "User not found in context")
			return
		}

		// Check if user exists
		var user models.User
		if err := database.DB.WithContext(r.Context()).First(&user, userID).Error; err != nil {
			WriteNotFound(w, r, "User not found")
			return
		}

		// Users can only delete their own account unless they're admins
		isAdmin := auth.HasRole(currentUser.Role, models.RoleAdmin, models.RoleSuperAdmin)
		if currentUser.ID != uint(userID) && !isAdmin {
			WriteForbidden(w, r, "You can only delete your own account")
			return
		}

		if err := database.DB.WithContext(r.Context()).Delete(&models.User{}, userID).Error; err != nil {
			WriteInternalError(w, r, "Failed to delete user")
			return
		}

		// Invalidate user cache after deletion
		invalidateUserCache(r.Context(), uint(userID))

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateUserRole godoc
// @Summary Update a user's role (Admin only)
// @Description Update the role of a specific user by their ID. Requires admin or super_admin privileges.
// @Tags admin, users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Param role body UpdateRoleRequest true "New role for the user"
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User role updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID or role"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse "Failed to update user role"
// @Router /admin/users/{id}/role [put]
func UpdateUserRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			WriteBadRequest(w, r, "Invalid user ID")
			return
		}

		var req UpdateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, r, "Invalid request payload")
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
			WriteBadRequest(w, r, "Invalid role. Valid roles are: super_admin, admin, premium, user")
			return
		}

		// Check if user exists
		var user models.User
		if err := database.DB.WithContext(r.Context()).First(&user, userID).Error; err != nil {
			WriteNotFound(w, r, "User not found")
			return
		}

		// Prevent users from modifying their own role (security measure)
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if ok && currentUser.ID == uint(userID) {
			WriteForbidden(w, r, "You cannot modify your own role")
			return
		}

		user.Role = req.Role
		user.UpdatedAt = time.Now()

		if err := database.DB.WithContext(r.Context()).Save(&user).Error; err != nil {
			WriteInternalError(w, r, "Failed to update user role")
			return
		}

		// Invalidate user cache after role change
		invalidateUserCache(r.Context(), uint(userID))

		userResponse := user.ToUserResponse()
		WriteSuccess(w, "User role updated successfully", userResponse)
	}
}

// GetPremiumContent godoc
// @Summary Get premium content
// @Description Retrieve exclusive content available only to premium and admin users
// @Tags premium
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=PremiumContentResponse} "Premium content retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - premium subscription required"
// @Router /premium/content [get]
func GetPremiumContent(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, "Welcome to the exclusive premium content!", PremiumContentResponse{
		Content: "This is premium content only available to our valued subscribers and administrators.",
		Features: []string{
			"Exclusive articles and tutorials",
			"Priority customer support",
			"Advanced features and tools",
			"Early access to new features",
		},
		AccessLevel: "premium",
	})
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User profile retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve user"
// @Router /users/me [get]
func GetCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, r, "User not found in context")
			return
		}

		userResponse := currentUser.ToUserResponse()
		WriteSuccess(w, "User profile retrieved successfully", userResponse)
	}
}

// UpdateCurrentUser godoc
// @Summary Update current user profile
// @Description Update the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param user body object true "Updated user data (name and/or email)"
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User profile updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid JSON"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 409 {object} models.ErrorResponse "Email already taken"
// @Failure 500 {object} models.ErrorResponse "Failed to update user"
// @Router /users/me [put]
func UpdateCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			WriteUnauthorized(w, r, "User not found in context")
			return
		}

		var req models.RegisterRequest // Reuse for update
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, r, "Invalid JSON")
			return
		}

		// Validate email format if provided
		if req.Email != "" && req.Email != currentUser.Email {
			if err := auth.ValidateEmail(req.Email); err != nil {
				WriteBadRequest(w, r, err.Error())
				return
			}

			// Check if email is already taken by another user
			var existingUser models.User
			if err := database.DB.WithContext(r.Context()).Where("email = ? AND id != ?", req.Email, currentUser.ID).First(&existingUser).Error; err == nil {
				WriteConflict(w, r, "Email already taken")
				return
			}
		}

		// Update user fields
		if req.Name != "" {
			currentUser.Name = req.Name
		}
		if req.Email != "" {
			currentUser.Email = req.Email
		}
		currentUser.UpdatedAt = time.Now()

		if err := database.DB.WithContext(r.Context()).Save(&currentUser).Error; err != nil {
			WriteInternalError(w, r, "Failed to update user")
			return
		}

		// Invalidate user cache after profile update
		invalidateUserCache(r.Context(), currentUser.ID)

		userResponse := currentUser.ToUserResponse()
		WriteSuccess(w, "User profile updated successfully", userResponse)
	}
}

// UpdateRoleRequest represents the request payload for updating a user role
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required" example:"premium"`
}

// PremiumContentResponse represents the premium content response
type PremiumContentResponse struct {
	Content     string   `json:"content"`
	Features    []string `json:"features"`
	AccessLevel string   `json:"access_level"`
}
