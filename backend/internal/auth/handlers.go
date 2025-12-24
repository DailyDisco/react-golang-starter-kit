package auth

import (
	"encoding/json"
	"net/http"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/jobs"
	"react-golang-starter/internal/models"

	"time"

	"github.com/rs/zerolog/log"
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
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	// Validate email format
	if err := ValidateEmail(req.Email); err != nil {
		writeBadRequest(w, r, err.Error())
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		writeBadRequest(w, r, err.Error())
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		writeConflict(w, r, "User with this email already exists")
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		writeInternalError(w, r, "Failed to hash password")
		return
	}

	// Generate verification token
	verificationToken, err := GenerateVerificationToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate verification token")
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
		writeInternalError(w, r, "Failed to create user")
		return
	}

	// Queue verification email (async via job queue)
	if jobs.IsAvailable() {
		if err := jobs.EnqueueVerificationEmail(r.Context(), user.ID, user.Email, user.Name, verificationToken); err != nil {
			// Log but don't fail registration - email can be resent
			log.Warn().Err(err).Uint("user_id", user.ID).Msg("failed to queue verification email")
		}
	}

	// Generate JWT access token
	token, err := GenerateJWT(&user)
	if err != nil {
		writeInternalError(w, r, "Failed to generate token")
		return
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate refresh token")
		return
	}

	// Save refresh token to user
	user.RefreshToken = refreshToken
	user.RefreshTokenExpires = time.Now().Add(GetRefreshTokenExpirationTime()).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		log.Warn().Err(err).Msg("failed to save refresh token")
	}

	// Set auth cookie
	SetAuthCookie(w, token)

	// Return response with refresh token
	response := models.AuthResponse{
		User:         user.ToUserResponse(),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(GetAccessTokenExpirationTime().Seconds()),
	}

	writeJSON(w, http.StatusCreated, response)
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
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid credentials")
		return
	}

	// Check if account is active
	if !user.IsActive {
		writeAccountInactive(w, r, "Account is deactivated")
		return
	}

	// Check password
	if !CheckPassword(req.Password, user.Password) {
		writeUnauthorized(w, r, "Invalid credentials")
		return
	}

	// Generate JWT access token
	token, err := GenerateJWT(&user)
	if err != nil {
		writeInternalError(w, r, "Failed to generate token")
		return
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate refresh token")
		return
	}

	// Save refresh token to user
	user.RefreshToken = refreshToken
	user.RefreshTokenExpires = time.Now().Add(GetRefreshTokenExpirationTime()).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		log.Warn().Err(err).Msg("failed to save refresh token")
	}

	// Set auth cookie
	SetAuthCookie(w, token)

	// Return response with refresh token
	response := models.AuthResponse{
		User:         user.ToUserResponse(),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(GetAccessTokenExpirationTime().Seconds()),
	}

	writeJSON(w, http.StatusOK, response)
}

// LogoutUser godoc
// @Summary Logout user
// @Description Logout user by clearing the authentication cookie and revoking the token
// @Tags auth
// @Produce json
// @Success 200 {object} models.SuccessResponse "Logout successful"
// @Router /auth/logout [post]
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	// Try to get the current token to blacklist it
	tokenString, err := ExtractTokenFromCookie(r)
	if err != nil {
		// Try Authorization header as fallback
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString, _ = ExtractTokenFromHeader(authHeader)
		}
	}

	// If we have a token, blacklist it
	if tokenString != "" {
		claims, err := ValidateJWT(tokenString)
		if err == nil && claims != nil {
			// Blacklist the access token
			if claims.ExpiresAt != nil {
				_ = BlacklistToken(tokenString, claims.UserID, claims.ExpiresAt.Time, "logout")
			}

			// Clear the user's refresh token
			_ = RevokeAllUserTokens(claims.UserID, "logout")

			// Invalidate the user cache
			_ = InvalidateUserCache(r.Context(), claims.UserID)
		}
	}

	ClearAuthCookie(w)
	writeSuccess(w, "Logout successful", nil)
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
		writeUnauthorized(w, r, "User not found in context")
		return
	}

	writeSuccess(w, "User retrieved successfully", user.ToUserResponse())
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
		writeBadRequest(w, r, "Verification token is required")
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		writeBadRequest(w, r, "Invalid verification token")
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			writeTokenExpired(w, r, "Verification token has expired")
			return
		}
	}

	// Update user verification status
	user.EmailVerified = true
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to verify email")
		return
	}

	writeSuccess(w, "Email verified successfully", nil)
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
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		writeSuccess(w, "If the email exists, a password reset link has been sent", nil)
		return
	}

	// Generate reset token
	resetToken, err := GenerateVerificationToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate reset token")
		return
	}

	// Update user with reset token
	user.VerificationToken = resetToken
	user.VerificationExpires = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to update user")
		return
	}

	// Queue password reset email (async via job queue)
	if jobs.IsAvailable() {
		if err := jobs.EnqueuePasswordResetEmail(r.Context(), user.ID, user.Email, user.Name, resetToken); err != nil {
			log.Warn().Err(err).Uint("user_id", user.ID).Msg("failed to queue password reset email")
		}
	}

	writeSuccess(w, "If the email exists, a password reset link has been sent", nil)
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
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		writeBadRequest(w, r, err.Error())
		return
	}

	var user models.User
	if err := database.DB.Where("verification_token = ?", req.Token).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid reset token")
		return
	}

	// Check if token is expired
	if user.VerificationExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.VerificationExpires)
		if err != nil || time.Now().After(expiresTime) {
			writeTokenExpired(w, r, "Reset token has expired")
			return
		}
	}

	// Hash new password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		writeInternalError(w, r, "Failed to hash password")
		return
	}

	// Update user password and clear reset token
	user.Password = hashedPassword
	user.VerificationToken = ""
	user.VerificationExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to reset password")
		return
	}

	// Invalidate user cache after password change
	_ = InvalidateUserCache(r.Context(), user.ID)

	// Revoke all existing tokens to force re-login
	_ = RevokeAllUserTokens(user.ID, "password_reset")

	writeSuccess(w, "Password reset successfully", nil)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token. This allows maintaining sessions without re-authentication.
// @Description The refresh token is long-lived (7 days by default) while access tokens are short-lived (15 minutes by default).
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.AuthResponse "New access token generated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} models.ErrorResponse "Failed to generate new token"
// @Router /auth/refresh [post]
func RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	if req.RefreshToken == "" {
		writeBadRequest(w, r, "Refresh token is required")
		return
	}

	// Find user by refresh token
	var user models.User
	if err := database.DB.Where("refresh_token = ?", req.RefreshToken).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid refresh token")
		return
	}

	// Check if refresh token is expired
	if user.RefreshTokenExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.RefreshTokenExpires)
		if err != nil || time.Now().After(expiresTime) {
			writeTokenExpired(w, r, "Refresh token has expired")
			return
		}
	}

	// Check if account is active
	if !user.IsActive {
		writeAccountInactive(w, r, "Account is deactivated")
		return
	}

	// Generate new access token
	token, err := GenerateJWT(&user)
	if err != nil {
		writeInternalError(w, r, "Failed to generate token")
		return
	}

	// Rotate refresh token for additional security
	// This prevents stolen refresh tokens from being reused indefinitely
	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate refresh token")
		return
	}
	user.RefreshToken = newRefreshToken
	user.RefreshTokenExpires = time.Now().Add(GetRefreshTokenExpirationTime()).Format(time.RFC3339)
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to save refresh token")
		return
	}

	// Set new auth cookie
	SetAuthCookie(w, token)

	// Return response with rotated refresh token
	response := models.AuthResponse{
		User:         user.ToUserResponse(),
		Token:        token,
		RefreshToken: newRefreshToken, // Return the new rotated refresh token
		ExpiresIn:    int64(GetAccessTokenExpirationTime().Seconds()),
	}

	writeJSON(w, http.StatusOK, response)
}
