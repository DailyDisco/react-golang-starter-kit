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
// @Failure 400 {string} string "Invalid JSON or validation error"
// @Failure 409 {string} string "User already exists"
// @Failure 500 {string} string "Failed to create user"
// @Router /auth/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate email format
	if err := ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
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
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Generate verification token
	verificationToken, err := GenerateVerificationToken()
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
		EmailVerified:       false, // In production, you might want to send verification email
		IsActive:            true,
		CreatedAt:           time.Now().Format(time.RFC3339),
		UpdatedAt:           time.Now().Format(time.RFC3339),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(&user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
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
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Failure 400 {string} string "Invalid JSON"
// @Failure 401 {string} string "Invalid credentials or account inactive"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if account is active
	if !user.IsActive {
		http.Error(w, "Account is deactivated", http.StatusUnauthorized)
		return
	}

	// Check password
	if !CheckPassword(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(&user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := models.AuthResponse{
		User:  user.ToUserResponse(),
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {object} models.UserResponse
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/me [get]
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user.ToUserResponse()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {string} string "Email verified successfully"
// @Failure 400 {string} string "Invalid or expired token"
// @Failure 500 {string} string "Failed to verify email"
// @Router /auth/verify-email [get]
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		http.Error(w, "Invalid verification token", http.StatusBadRequest)
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			http.Error(w, "Verification token has expired", http.StatusBadRequest)
			return
		}
	}

	// Update user verification status
	user.EmailVerified = true
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "Email verified successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {string} string "Password reset email sent"
// @Failure 400 {string} string "Invalid JSON"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to send reset email"
// @Router /auth/reset-password [post]
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate reset token
	resetToken, err := GenerateVerificationToken()
	if err != nil {
		http.Error(w, "Failed to generate reset token", http.StatusInternalServerError)
		return
	}

	// Update user with reset token
	user.VerificationToken = resetToken
	user.VerificationExpires = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// TODO: Send email with reset link
	// For now, just return success message

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
// @Success 200 {string} string "Password reset successfully"
// @Failure 400 {string} string "Invalid JSON or password"
// @Failure 401 {string} string "Invalid or expired token"
// @Failure 500 {string} string "Failed to reset password"
// @Router /auth/reset-password/confirm [post]
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", req.Token).First(&user).Error; err != nil {
		http.Error(w, "Invalid reset token", http.StatusUnauthorized)
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			http.Error(w, "Reset token has expired", http.StatusUnauthorized)
			return
		}
	}

	// Hash new password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update user password and clear reset token
	user.Password = hashedPassword
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "Password reset successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
