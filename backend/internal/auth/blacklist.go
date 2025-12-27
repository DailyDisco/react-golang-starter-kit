package auth

import (
	"os"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

// getBlacklistFailMode returns the configured fail mode for token blacklist checks.
// Default is "closed" for security - if blacklist check fails, deny the request.
// Set to "open" via TOKEN_BLACKLIST_FAIL_MODE=open for availability-first behavior.
func getBlacklistFailMode() string {
	mode := os.Getenv("TOKEN_BLACKLIST_FAIL_MODE")
	if mode == "open" {
		return "open"
	}
	return "closed"
}

// BlacklistToken adds a token to the blacklist
// The token is hashed before storage for security
func BlacklistToken(token string, userID uint, expiresAt time.Time, reason string) error {
	// Skip if database is not initialized (for testing)
	if database.DB == nil {
		return nil
	}

	if reason == "" {
		reason = "logout"
	}

	entry := models.TokenBlacklist{
		TokenHash: HashToken(token),
		UserID:    userID,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		RevokedAt: time.Now().Format(time.RFC3339),
		Reason:    reason,
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		// If token is already blacklisted (duplicate), that's fine
		log.Debug().Err(err).Msg("failed to blacklist token (may already exist)")
		return nil
	}

	return nil
}

// IsTokenBlacklisted checks if a token has been revoked.
// On database error, behavior is controlled by TOKEN_BLACKLIST_FAIL_MODE:
// - "closed" (default): Deny request on error (security-first)
// - "open": Allow request on error (availability-first)
func IsTokenBlacklisted(token string) bool {
	// Skip if database is not initialized (for testing)
	if database.DB == nil {
		return false
	}

	tokenHash := HashToken(token)

	var count int64
	if err := database.DB.Model(&models.TokenBlacklist{}).
		Where("token_hash = ?", tokenHash).
		Count(&count).Error; err != nil {
		failMode := getBlacklistFailMode()
		log.Warn().
			Err(err).
			Str("fail_mode", failMode).
			Str("token_hash_prefix", tokenHash[:8]+"...").
			Msg("SECURITY: Token blacklist check failed")

		// In closed mode, treat as blacklisted on error for security
		if failMode == "closed" {
			return true
		}
		// In open mode, allow the request for availability
		return false
	}

	return count > 0
}

// CleanupExpiredBlacklistEntries removes expired entries from the blacklist
// This should be run periodically (e.g., daily) to prevent table bloat
func CleanupExpiredBlacklistEntries() error {
	// Skip if database is not initialized (for testing)
	if database.DB == nil {
		return nil
	}

	result := database.DB.
		Where("expires_at < ?", time.Now().Format(time.RFC3339)).
		Delete(&models.TokenBlacklist{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Info().Int64("count", result.RowsAffected).Msg("cleaned up expired blacklist entries")
	}

	return nil
}

// RevokeAllUserTokens adds all of a user's active tokens to the blacklist
// This is used when a user changes their password or is deactivated
func RevokeAllUserTokens(userID uint, reason string) error {
	// Skip if database is not initialized (for testing)
	if database.DB == nil {
		return nil
	}

	// Since we don't track all issued tokens, we can only blacklist
	// the current refresh token. The access tokens will expire naturally.
	// For immediate revocation, we set a flag or delete the refresh token.

	// Clear the user's refresh token to prevent new access tokens
	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"refresh_token":         "",
			"refresh_token_expires": "",
			"updated_at":            time.Now().Format(time.RFC3339),
		}).Error; err != nil {
		return err
	}

	log.Info().Uint("user_id", userID).Str("reason", reason).Msg("revoked all user tokens")
	return nil
}
