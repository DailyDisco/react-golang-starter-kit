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
	"encoding/json"
	"net/http"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// Service represents the application service with its dependencies
type Service struct {
	RedisClient *cache.Client
}

// NewService creates a new Service instance
func NewService(redisClient *cache.Client) *Service {
	return &Service{
		RedisClient: redisClient,
	}
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// HealthCheck godoc
// @Summary Check server health status
// @Description Get the health status of the server and its dependencies (database, redis)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthStatus "Server and its dependencies are healthy"
// @Failure 503 {object} models.HealthStatus "Server or one of its critical dependencies is unhealthy"
// @Router /health [get]
func (s *Service) HealthCheck(w http.ResponseWriter, r *http.Request) {
	overallStatus := "healthy"
	statusCode := http.StatusOK

	dbStatus := database.CheckDatabaseHealth()
	redisStatus := s.RedisClient.CheckRedisHealth()

	components := []models.ComponentStatus{
		dbStatus,
		redisStatus,
	}

	for _, comp := range components {
		if comp.Status == "unhealthy" {
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	healthResponse := models.HealthStatus{
		OverallStatus: overallStatus,
		Timestamp:     time.Now().Format(time.RFC3339),
		Components:    components,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthResponse)
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
func GetUsers(cacheService *cache.Service) http.HandlerFunc {
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
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Bad Request",
					Message: "Invalid page parameter",
					Code:    http.StatusBadRequest,
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			} else {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Bad Request",
					Message: "Invalid limit parameter (must be 1-100)",
					Code:    http.StatusBadRequest,
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		// Try to get from cache first
		if cachedUsers, err := cacheService.GetUserList(page, limit); err == nil {
			log.Debug().
				Int("page", page).
				Int("limit", limit).
				Msg("Retrieved users from cache")

			response := models.SuccessResponse{
				Success: true,
				Message: "Users retrieved successfully",
				Data:    cachedUsers,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				response := models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "Failed to encode response",
					Code:    http.StatusInternalServerError,
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
			}
			return
		}

		// Get total count (try cache first)
		var total int64
		if cachedCount, err := cacheService.GetUserCount(); err == nil {
			total = int64(cachedCount)
			log.Debug().Int64("total", total).Msg("Retrieved user count from cache")
		} else {
			if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "Failed to count users",
					Code:    http.StatusInternalServerError,
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
			// Cache the count
			cacheService.SetUserCount(int(total))
		}

		// Calculate offset and total pages
		offset := (page - 1) * limit
		totalPages := int((total + int64(limit) - 1) / int64(limit))

		// Get paginated users
		var users []models.User
		if err := database.DB.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to fetch users",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
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

		// Cache the result
		if err := cacheService.SetUserList(page, limit, &usersResponse); err != nil {
			log.Warn().Err(err).Msg("Failed to cache user list")
		} else {
			log.Debug().
				Int("page", page).
				Int("limit", limit).
				Int("count", len(userResponses)).
				Msg("Cached user list")
		}

		response := models.SuccessResponse{
			Success: true,
			Message: "Users retrieved successfully",
			Data:    usersResponse,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to encode response",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
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
func GetUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found in context",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Users can only view their own profile unless they're admins
		if currentUser.ID != uint(userID) {
			// TODO: Add role checking for admin users
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You can only view your own profile",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Try to get from cache first
		if cachedUser, err := cacheService.GetUser(uint(userID)); err == nil {
			log.Debug().
				Uint("userID", uint(userID)).
				Msg("Retrieved user from cache")

			w.Header().Set("Content-Type", "application/json")
			response := models.SuccessResponse{
				Success: true,
				Message: "User retrieved successfully",
				Data:    cachedUser,
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				response := models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "Failed to encode response",
					Code:    http.StatusInternalServerError,
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
			}
			return
		}

		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		userResponse := user.ToUserResponse()

		// Cache the result
		if err := cacheService.SetUser(&userResponse); err != nil {
			log.Warn().Err(err).Uint("userID", uint(userID)).Msg("Failed to cache user")
		} else {
			log.Debug().Uint("userID", uint(userID)).Msg("Cached user")
		}

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User retrieved successfully",
			Data:    userResponse,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to encode response",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}
}

// CreateUser godoc
// @Summary Create a new user (Admin endpoint)
// @Description Create a new user with the provided information. This endpoint is intended for administrative use and may require admin privileges in production.
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.SuccessResponse{data=models.UserResponse} "User created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid JSON or validation error"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden - admin privileges required"
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
func CreateUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated (optional for now, but could be made admin-only)
		_, _ = auth.GetUserFromContext(r.Context()) // authenticated status available for future admin-only logic

		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid JSON",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Validate email format
		if err := auth.ValidateEmail(req.Email); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Validate password strength
		if err := auth.ValidatePassword(req.Password); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Check if user already exists
		var existingUser models.User
		if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Conflict",
				Message: "User with this email already exists",
				Code:    http.StatusConflict,
			}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Hash password
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to hash password",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Generate verification token
		verificationToken, err := auth.GenerateVerificationToken()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to generate verification token",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Create user
		user := models.User{
			Name:                req.Name,
			Email:               req.Email,
			Password:            hashedPassword,
			VerificationToken:   verificationToken,
			VerificationExpires: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			EmailVerified:       true, // Admin-created users are pre-verified
			IsActive:            true,
			CreatedAt:           time.Now().Format(time.RFC3339),
			UpdatedAt:           time.Now().Format(time.RFC3339),
		}

		if err := database.DB.Create(&user).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to create user",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		userResponse := user.ToUserResponse()

		// Cache the new user
		if err := cacheService.SetUser(&userResponse); err != nil {
			log.Warn().Err(err).Uint("userID", user.ID).Msg("Failed to cache new user")
		} else {
			log.Debug().Uint("userID", user.ID).Msg("Cached new user")
		}

		// Invalidate user lists and count cache since we added a user
		if err := cacheService.InvalidateAllUsers(); err != nil {
			log.Warn().Err(err).Msg("Failed to invalidate user caches after creation")
		} else {
			log.Debug().Msg("Invalidated user caches after creation")
		}

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User created successfully",
			Data:    userResponse,
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to encode response",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
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
func UpdateUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found in context",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Users can only update their own profile unless they're admins
		if currentUser.ID != uint(userID) {
			// TODO: Add role checking for admin users
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You can only update your own profile",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		var updateData struct {
			Name  string `json:"name,omitempty"`
			Email string `json:"email,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid JSON",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Validate email if provided
		if updateData.Email != "" {
			if err := auth.ValidateEmail(updateData.Email); err != nil {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Bad Request",
					Message: err.Error(),
					Code:    http.StatusBadRequest,
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
				return
			}

			// Check if email is already taken by another user
			var existingUser models.User
			if err := database.DB.Where("email = ? AND id != ?", updateData.Email, userID).First(&existingUser).Error; err == nil {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Conflict",
					Message: "Email is already taken",
					Code:    http.StatusConflict,
				}
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
				return
			}

			user.Email = updateData.Email
		}

		if updateData.Name != "" {
			user.Name = updateData.Name
		}

		user.UpdatedAt = time.Now().Format(time.RFC3339)

		if err := database.DB.Save(&user).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to update user",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		userResponse := user.ToUserResponse()

		// Update cache with the new user data
		if err := cacheService.SetUser(&userResponse); err != nil {
			log.Warn().Err(err).Uint("userID", uint(userID)).Msg("Failed to cache updated user")
		} else {
			log.Debug().Uint("userID", uint(userID)).Msg("Cached updated user")
		}

		// Invalidate user lists since user data might appear in cached lists
		if err := cacheService.InvalidateUserList(); err != nil {
			log.Warn().Err(err).Msg("Failed to invalidate user list cache after update")
		} else {
			log.Debug().Msg("Invalidated user list cache after update")
		}

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User updated successfully",
			Data:    userResponse,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to encode response",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
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
func DeleteUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Get the authenticated user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found in context",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Check if user exists
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		// Users can only delete their own account unless they're admins
		if currentUser.ID != uint(userID) {
			// TODO: Add role checking for admin users
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You can only delete your own account",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
			return
		}

		if err := database.DB.Delete(&models.User{}, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to delete user",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Invalidate all user-related caches after deletion
		if err := cacheService.InvalidateUser(uint(userID)); err != nil {
			log.Warn().Err(err).Int("userID", userID).Msg("Failed to invalidate user cache after deletion")
		} else {
			log.Debug().Int("userID", userID).Msg("Invalidated user cache after deletion")
		}

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
func UpdateUserRole(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		var req UpdateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid request payload",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
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
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid role. Valid roles are: super_admin, admin, premium, user",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check if user exists
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Prevent users from modifying their own role (security measure)
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if ok && currentUser.ID == uint(userID) {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You cannot modify your own role",
				Code:    http.StatusForbidden,
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}

		oldRole := user.Role
		user.Role = req.Role
		user.UpdatedAt = time.Now().Format(time.RFC3339)

		if err := database.DB.Save(&user).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to update user role",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Invalidate user cache
		if err := cacheService.DeleteUser(uint(userID)); err != nil {
			log.Warn().Err(err).Int("userID", userID).Msg("Failed to invalidate user cache after role update")
		} else {
			log.Debug().Int("userID", userID).Str("oldRole", oldRole).Str("newRole", req.Role).Msg("Invalidated user cache after role update")
		}

		userResponse := user.ToUserResponse()

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User role updated successfully",
			Data:    userResponse,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
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
	response := models.SuccessResponse{
		Success: true,
		Message: "Welcome to the exclusive premium content!",
		Data: PremiumContentResponse{
			Content: "This is premium content only available to our valued subscribers and administrators.",
			Features: []string{
				"Exclusive articles and tutorials",
				"Priority customer support",
				"Advanced features and tools",
				"Early access to new features",
			},
			AccessLevel: "premium",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
func GetCurrentUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found in context",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		userResponse := currentUser.ToUserResponse()

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User profile retrieved successfully",
			Data:    userResponse,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
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
func UpdateCurrentUser(cacheService *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found in context",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		var req models.RegisterRequest // Reuse for update
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid JSON",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Validate email format if provided
		if req.Email != "" && req.Email != currentUser.Email {
			if err := auth.ValidateEmail(req.Email); err != nil {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Bad Request",
					Message: err.Error(),
					Code:    http.StatusBadRequest,
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}

			// Check if email is already taken by another user
			var existingUser models.User
			if err := database.DB.Where("email = ? AND id != ?", req.Email, currentUser.ID).First(&existingUser).Error; err == nil {
				w.Header().Set("Content-Type", "application/json")
				response := models.ErrorResponse{
					Error:   "Conflict",
					Message: "Email already taken",
					Code:    http.StatusConflict,
				}
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(response)
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
		currentUser.UpdatedAt = time.Now().Format(time.RFC3339)

		if err := database.DB.Save(&currentUser).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to update user",
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Invalidate user cache
		if err := cacheService.DeleteUser(currentUser.ID); err != nil {
			log.Warn().Err(err).Uint("userID", currentUser.ID).Msg("Failed to invalidate user cache after update")
		} else {
			log.Debug().Uint("userID", currentUser.ID).Msg("Invalidated user cache after update")
		}

		userResponse := currentUser.ToUserResponse()

		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "User profile updated successfully",
			Data:    userResponse,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
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
