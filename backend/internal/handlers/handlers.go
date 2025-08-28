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
// @Success 200 {object} map[string]string "status: ok, message: Server is running"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "ok",
		"message": "Server is running",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all users (public endpoint)
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.UserResponse
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Convert to UserResponse to hide sensitive fields
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userResponses); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {object} models.UserResponse
// @Failure 400 {string} string "Invalid user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from context
	currentUser, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Users can only view their own profile unless they're admins
	if currentUser.ID != uint(userID) {
		// TODO: Add role checking for admin users
		http.Error(w, "Forbidden: You can only view your own profile", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user.ToUserResponse()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 201 {object} models.UserResponse
// @Failure 400 {string} string "Invalid JSON or validation error"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden: Admin privileges required"
// @Failure 409 {string} string "User already exists"
// @Failure 500 {string} string "Failed to create user"
// @Router /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated (optional for now, but could be made admin-only)
	_, _ = auth.GetUserFromContext(r.Context()) // authenticated status available for future admin-only logic

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate email format
	if err := auth.ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := auth.ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Generate verification token
	verificationToken, err := auth.GenerateVerificationToken()
	if err != nil {
		http.Error(w, "Failed to generate verification token", http.StatusInternalServerError)
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
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user.ToUserResponse()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {object} models.UserResponse
// @Failure 400 {string} string "Invalid user ID or JSON"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to update user"
// @Router /users/{id} [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from context
	currentUser, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Users can only update their own profile unless they're admins
	if currentUser.ID != uint(userID) {
		// TODO: Add role checking for admin users
		http.Error(w, "Forbidden: You can only update your own profile", http.StatusForbidden)
		return
	}

	var updateData struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate email if provided
	if updateData.Email != "" {
		if err := auth.ValidateEmail(updateData.Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if email is already taken by another user
		var existingUser models.User
		if err := database.DB.Where("email = ? AND id != ?", updateData.Email, userID).First(&existingUser).Error; err == nil {
			http.Error(w, "Email is already taken", http.StatusConflict)
			return
		}

		user.Email = updateData.Email
	}

	if updateData.Name != "" {
		user.Name = updateData.Name
	}

	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user.ToUserResponse()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Failure 400 {string} string "Invalid user ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to delete user"
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from context
	currentUser, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Users can only delete their own account unless they're admins
	if currentUser.ID != uint(userID) {
		// TODO: Add role checking for admin users
		http.Error(w, "Forbidden: You can only delete your own account", http.StatusForbidden)
		return
	}

	if err := database.DB.Delete(&models.User{}, userID).Error; err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
