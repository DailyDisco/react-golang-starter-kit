package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"react-golang-starter/internal/audit"
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

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", normalizedEmail).First(&existingUser).Error; err == nil {
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
		Email:               normalizedEmail,
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

	// Audit log user registration
	audit.LogUserCreate(nil, user.ID, r)

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
	refreshExpires := time.Now().Add(GetRefreshTokenExpirationTime())
	user.RefreshTokenExpires = &refreshExpires
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		log.Warn().Err(err).Msg("failed to save refresh token")
	}

	// Set auth cookie
	SetAuthCookie(w, token)

	// Set refresh token as httpOnly cookie (more secure than localStorage)
	SetRefreshCookie(w, refreshToken)

	// Return response (refresh token now in httpOnly cookie, not exposed in response)
	response := models.AuthResponse{
		User:      user.ToUserResponse(),
		Token:     token,
		ExpiresIn: int64(GetAccessTokenExpirationTime().Seconds()),
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
//
// Account lockout configuration
const (
	MaxFailedLoginAttempts = 5                // Lock after this many failed attempts
	LockoutDuration        = 30 * time.Minute // Lock account for this duration
	FailedLoginWindow      = 15 * time.Minute // Reset counter if last failure was longer ago
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, r, "Invalid JSON")
		return
	}

	// Normalize email to lowercase for case-insensitive comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid credentials")
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remainingLockTime := time.Until(*user.LockedUntil).Round(time.Minute)
		log.Warn().
			Uint("user_id", user.ID).
			Str("email", user.Email).
			Time("locked_until", *user.LockedUntil).
			Msg("login attempt on locked account")
		writeJSON(w, http.StatusTooManyRequests, models.ErrorResponse{
			Error:   "Account Locked",
			Message: "Account is temporarily locked due to too many failed login attempts. Try again in " + remainingLockTime.String(),
			Code:    http.StatusTooManyRequests,
		})
		return
	}

	// Check if account is active
	if !user.IsActive {
		writeAccountInactive(w, r, "Account is deactivated")
		return
	}

	// Check password
	if !CheckPassword(req.Password, user.Password) {
		// Track failed login attempt
		handleFailedLogin(&user, r)
		writeUnauthorized(w, r, "Invalid credentials")
		return
	}

	// Successful login - reset failed login counter
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		user.LastFailedLogin = nil
		if err := database.DB.Model(&user).Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"locked_until":          nil,
			"last_failed_login":     nil,
		}).Error; err != nil {
			log.Warn().Err(err).Uint("user_id", user.ID).Msg("failed to reset login attempts")
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
	refreshExpires := time.Now().Add(GetRefreshTokenExpirationTime())
	user.RefreshTokenExpires = &refreshExpires
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		log.Warn().Err(err).Msg("failed to save refresh token")
	}

	// Set auth and refresh cookies
	SetAuthCookie(w, token)
	SetRefreshCookie(w, refreshToken)

	// Audit log successful login
	audit.LogLogin(user.ID, r, nil)

	// Return response (refresh token now in httpOnly cookie, not exposed in response)
	response := models.AuthResponse{
		User:      user.ToUserResponse(),
		Token:     token,
		ExpiresIn: int64(GetAccessTokenExpirationTime().Seconds()),
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

			// Audit log logout
			audit.LogLogout(claims.UserID, r)
		}
	}

	ClearAuthCookie(w)
	ClearRefreshCookie(w)
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

	// Generate reset token (using dedicated password reset token field)
	resetToken, err := GenerateVerificationToken()
	if err != nil {
		writeInternalError(w, r, "Failed to generate reset token")
		return
	}

	// Hash the token before storing for security (plaintext token is sent via email)
	hashedResetToken := HashToken(resetToken)

	// Update user with hashed reset token (separate from email verification token)
	user.PasswordResetToken = hashedResetToken
	user.PasswordResetExpires = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
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

	// Hash the submitted token to compare with stored hash
	hashedToken := HashToken(req.Token)

	var user models.User
	if err := database.DB.Where("password_reset_token = ?", hashedToken).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid reset token")
		return
	}

	// Check if token is expired
	if user.PasswordResetExpires != "" {
		expiresTime, err := time.Parse(time.RFC3339, user.PasswordResetExpires)
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
	user.PasswordResetToken = ""
	user.PasswordResetExpires = ""
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to reset password")
		return
	}

	// Audit log password reset
	audit.LogUserUpdate(user.ID, user.ID, map[string]interface{}{"password": "changed"}, r)

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
	// Extract refresh token from httpOnly cookie
	refreshToken, err := ExtractRefreshTokenFromCookie(r)
	if err != nil {
		writeUnauthorized(w, r, "Refresh token not found")
		return
	}

	// Find user by refresh token
	var user models.User
	if err := database.DB.Where("refresh_token = ?", refreshToken).First(&user).Error; err != nil {
		writeUnauthorized(w, r, "Invalid refresh token")
		return
	}

	// Check if refresh token is expired
	if user.RefreshTokenExpires != nil {
		if time.Now().After(*user.RefreshTokenExpires) {
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
	newRefreshExpires := time.Now().Add(GetRefreshTokenExpirationTime())
	user.RefreshTokenExpires = &newRefreshExpires
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := database.DB.Save(&user).Error; err != nil {
		writeInternalError(w, r, "Failed to save refresh token")
		return
	}

	// Set new auth and refresh cookies
	SetAuthCookie(w, token)
	SetRefreshCookie(w, newRefreshToken)

	// Return response (refresh token now in httpOnly cookie, not exposed in response)
	response := models.AuthResponse{
		User:      user.ToUserResponse(),
		Token:     token,
		ExpiresIn: int64(GetAccessTokenExpirationTime().Seconds()),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleFailedLogin tracks failed login attempts and locks accounts after too many failures.
// This provides brute-force protection at the account level (in addition to IP-based rate limiting).
func handleFailedLogin(user *models.User, r *http.Request) {
	now := time.Now()

	// Check if we should reset the counter (last failure was outside the window)
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}

	// Increment failed attempts
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	// Lock account if threshold exceeded
	if user.FailedLoginAttempts >= MaxFailedLoginAttempts {
		lockUntil := now.Add(LockoutDuration)
		user.LockedUntil = &lockUntil

		log.Warn().
			Uint("user_id", user.ID).
			Str("email", user.Email).
			Int("failed_attempts", user.FailedLoginAttempts).
			Time("locked_until", lockUntil).
			Str("ip", getClientIP(r)).
			Msg("account locked due to too many failed login attempts")

		// TODO: Queue email notification about account lockout
		// if jobs.IsAvailable() {
		//     jobs.EnqueueAccountLockoutNotification(r.Context(), user.ID, user.Email, user.Name, lockUntil)
		// }
	}

	// Save the updated user
	updates := map[string]interface{}{
		"failed_login_attempts": user.FailedLoginAttempts,
		"last_failed_login":     user.LastFailedLogin,
		"locked_until":          user.LockedUntil,
	}
	if err := database.DB.Model(user).Updates(updates).Error; err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update login attempt tracking")
	}

	// Audit log failed login attempt
	audit.LogLogin(user.ID, r, map[string]interface{}{
		"success":         false,
		"failed_attempts": user.FailedLoginAttempts,
		"locked":          user.LockedUntil != nil,
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}
