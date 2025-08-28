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
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/register [post]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}

// LoginUser godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/login [post]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}

// GetCurrentUser godoc
// @Summary Get current user information
// @Description Retrieve information about the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/me [get]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user's email address using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/verify-email [get]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
			json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/reset-password [post]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirm true "Password reset confirmation"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/reset-password/confirm [post]
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
			json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
		return
	}
}
