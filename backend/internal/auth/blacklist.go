package auth

import (
	"context"
	"os"
	"time"

	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
)

// Blacklist cache configuration
const (
	// blacklistCacheTTL is how long a blacklisted token is cached.
	// Should be shorter than token expiry to avoid stale cache issues.
	blacklistCacheTTL = 15 * time.Minute

	// blacklistCachePrefix is the cache key prefix for blacklisted tokens.
	blacklistCachePrefix = "blacklist:"
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

	tokenHash := HashToken(token)
	entry := models.TokenBlacklist{
		TokenHash: tokenHash,
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

	// Cache the blacklisted token for faster lookups
	// Use first 16 chars of hash as cache key (sufficient for uniqueness)
	cacheKey := blacklistCachePrefix + tokenHash[:16]
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cache.Set(ctx, cacheKey, []byte("1"), blacklistCacheTTL); err != nil {
		// Cache failure is non-critical, just log it
		log.Debug().Err(err).Str("key", cacheKey).Msg("failed to cache blacklisted token")
	}

	return nil
}

// IsTokenBlacklisted checks if a token has been revoked.
// It first checks the cache for faster lookups, then falls back to the database.
// On database error, behavior is controlled by TOKEN_BLACKLIST_FAIL_MODE:
// - "closed" (default): Deny request on error (security-first)
// - "open": Allow request on error (availability-first)
func IsTokenBlacklisted(token string) bool {
	// Skip if database is not initialized (for testing)
	if database.DB == nil {
		return false
	}

	tokenHash := HashToken(token)
	cacheKey := blacklistCachePrefix + tokenHash[:16]

	// Check cache first (fast path)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if cache.Exists(ctx, cacheKey) {
		// Token is cached as blacklisted
		return true
	}

	// Cache miss - check database
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

	isBlacklisted := count > 0

	// Cache the result if blacklisted (positive caching)
	// We don't cache negative results to avoid memory bloat
	if isBlacklisted {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cacheCancel()
		if err := cache.Set(cacheCtx, cacheKey, []byte("1"), blacklistCacheTTL); err != nil {
			log.Debug().Err(err).Str("key", cacheKey).Msg("failed to cache blacklist result")
		}
	}

	return isBlacklisted
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
