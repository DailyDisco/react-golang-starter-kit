package auth

import (
	"encoding/json"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"time"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Create a new user account with email and password. The user will receive an email verification link.
// @Description This endpoint is rate limited to prevent abuse. Maximum 5 requests per minute per IP.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.AuthResponse "User registered successfully with JWT token"
// @Failure 400 {object} models.ErrorResponse "Invalid input or validation error"
// @Failure 409 {object} models.ErrorResponse "User with this email already exists"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} models.ErrorResponse "Failed to create user"
// @Router /auth/register [post]
// @Example
//
// Request:
// POST /api/auth/register
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
//	  "user": {
//	    "id": 1,
//	    "name": "John Doe",
//	    "email": "john.doe@example.com",
//	    "email_verified": false,
//	    "is_active": true,
//	    "created_at": "2023-08-27T12:00:00Z",
//	    "updated_at": "2023-08-27T12:00:00Z"
//	  },
//	  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	}
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid JSON",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Validate email format
	if err := ValidateEmail(req.Email); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
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
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to hash password",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Generate verification token
	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate verification token",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Create user
	user := models.User{
		Name:                req.Name,
		Email:               req.Email,
		Password:            hashedPassword,
		VerificationToken:   verificationToken,
		VerificationExpires: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		EmailVerified:       false, // In production, you might want to send verification email
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
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(&user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate token",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Return response
	response := models.AuthResponse{
		User:  user.ToUserResponse(),
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// LoginUser godoc
// @Summary Login user
// @Description Authenticate user with email and password. Returns JWT token for subsequent requests.
// @Description This endpoint is rate limited to prevent brute force attacks. Maximum 10 requests per minute per IP.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.AuthResponse "Login successful with JWT token"
// @Failure 400 {object} models.ErrorResponse "Invalid JSON or missing fields"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials or account deactivated"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} models.ErrorResponse "Failed to generate token"
// @Router /auth/login [post]
// @Example
//
// Request:
// POST /api/auth/login
//
//	{
//	  "email": "john.doe@example.com",
//	  "password": "SecurePass123!"
//	}
//
// Response:
//
//	{
//	  "user": {
//	    "id": 1,
//	    "name": "John Doe",
//	    "email": "john.doe@example.com",
//	    "email_verified": true,
//	    "is_active": true,
//	    "created_at": "2023-08-27T12:00:00Z",
//	    "updated_at": "2023-08-27T12:00:00Z"
//	  },
//	  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	}
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid JSON",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid credentials",
			Code:    http.StatusUnauthorized,
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Check if account is active
	if !user.IsActive {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Account is deactivated",
			Code:    http.StatusUnauthorized,
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Check password
	if !CheckPassword(req.Password, user.Password) {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid credentials",
			Code:    http.StatusUnauthorized,
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(&user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate token",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Return response
	response := models.AuthResponse{
		User:  user.ToUserResponse(),
		Token: token,
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
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Retrieve detailed information about the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse{data=models.UserResponse} "User information retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve user information"
// @Router /auth/me [get]
// @Example
//
// Request:
// GET /api/auth/me
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
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
			Code:    http.StatusUnauthorized,
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
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
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user's email address using the verification token sent via email
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token from email"
// @Success 200 {object} models.SuccessResponse "Email verified successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid or expired verification token"
// @Failure 500 {object} models.ErrorResponse "Failed to verify email"
// @Router /auth/verify-email [get]
// @Example
//
// Request:
// GET /api/auth/verify-email?token=abc123def456
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "Email verified successfully"
//	}
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Verification token is required",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid verification token",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Verification token has expired",
				Code:    http.StatusBadRequest,
			}
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	}

	// Update user verification status
	user.EmailVerified = true
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to verify email",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send password reset email to user. For security, always returns success message even if email doesn't exist.
// @Description This endpoint is rate limited to prevent abuse. Maximum 3 requests per minute per IP.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request with email"
// @Success 200 {object} models.SuccessResponse "Reset email sent (if email exists)"
// @Failure 400 {object} models.ErrorResponse "Invalid email format"
// @Failure 429 {object} models.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} models.ErrorResponse "Failed to process reset request"
// @Router /auth/reset-password [post]
// @Example
//
// Request:
// POST /api/auth/reset-password
//
//	{
//	  "email": "john.doe@example.com"
//	}
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "If the email exists, a password reset link has been sent"
//	}
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid JSON",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		w.Header().Set("Content-Type", "application/json")
		response := models.SuccessResponse{
			Success: true,
			Message: "If the email exists, a password reset link has been sent",
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Generate reset token
	resetToken, err := GenerateVerificationToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate reset token",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Update user with reset token
	user.VerificationToken = resetToken
	user.VerificationExpires = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update user",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// TODO: Send email with reset link
	// For now, just return success message

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using the reset token from email and new password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirm true "Password reset confirmation with token and new password"
// @Success 200 {object} models.SuccessResponse "Password reset successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid input or password validation failed"
// @Failure 401 {object} models.ErrorResponse "Invalid or expired reset token"
// @Failure 500 {object} models.ErrorResponse "Failed to reset password"
// @Router /auth/reset-password/confirm [post]
// @Example
//
// Request:
// POST /api/auth/reset-password/confirm
//
//	{
//	  "token": "reset_token_123",
//	  "password": "NewSecurePass123!"
//	}
//
// Response:
//
//	{
//	  "success": true,
//	  "message": "Password reset successfully"
//	}
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid JSON",
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", req.Token).First(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid reset token",
			Code:    http.StatusUnauthorized,
		}
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			w.Header().Set("Content-Type", "application/json")
			response := models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Reset token has expired",
				Code:    http.StatusUnauthorized,
			}
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	}

	// Hash new password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to hash password",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Update user password and clear reset token
	user.Password = hashedPassword
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to reset password",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := models.SuccessResponse{
		Success: true,
		Message: "Password reset successfully",
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		response := models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to encode response",
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}
