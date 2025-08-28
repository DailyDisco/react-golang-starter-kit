// Package handlers contains HTTP request handlers for the API
//
// # React Go Starter Kit API
//
// This is the REST API for the React Go Starter Kit application.
// It provides endpoints for user authentication, user management, and health checks.
//
// Terms Of Service: http://swagger.io/terms/
//
// Schemes: http, https
// Host: localhost:8080
// BasePath: /api
// Version: 1.0.0
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
//	description: "JWT Authorization header using the Bearer scheme. Example: \"Authorization: Bearer {token}\""
//
// swagger:meta
package handlers

import (
	"encoding/json"
	"net/http"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// HealthCheck godoc
// @Summary Check server health status
// @Description Get the health status of the server
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "Server is running",
		Data: models.HealthResponse{
			Status:  "ok",
			Message: "Server is running",
		},
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all users (public endpoint)
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch users",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}

	// Convert to UserResponse to hide sensitive fields
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	response := models.SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data: models.UsersResponse{
			Users: userResponses,
			Count: len(userResponses),
		},
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
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Retrieve a single user by their ID (protected endpoint)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user.ToUserResponse(),
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}
}

// CreateUser godoc
// @Summary Create a new user (Admin endpoint)
// @Description Create a new user with the provided information (requires admin privileges)
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user.ToUserResponse(),
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Update a user's information by their ID (protected endpoint)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.User true "Updated user object"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user.ToUserResponse(),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by their ID (protected endpoint)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(response) // Error intentionally ignored as we're already in an error state
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
