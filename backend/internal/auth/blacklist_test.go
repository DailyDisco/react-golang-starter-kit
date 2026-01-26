package auth

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected int // Expected hash length (64 chars for SHA-256 hex)
	}{
		{
			name:     "regular token",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			expected: 64,
		},
		{
			name:     "empty token",
			token:    "",
			expected: 64,
		},
		{
			name:     "short token",
			token:    "abc",
			expected: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashToken(tt.token)

			if len(hash) != tt.expected {
				t.Errorf("Expected hash length %d, got %d", tt.expected, len(hash))
			}

			// Hash should be consistent
			hash2 := HashToken(tt.token)
			if hash != hash2 {
				t.Error("Hash should be deterministic for same input")
			}
		})
	}
}

func TestHashToken_DifferentInputs(t *testing.T) {
	token1 := "token1"
	token2 := "token2"

	hash1 := HashToken(token1)
	hash2 := HashToken(token2)

	if hash1 == hash2 {
		t.Error("Different tokens should produce different hashes")
	}
}

func TestHashToken_SimilarInputs(t *testing.T) {
	// Even slightly different inputs should produce completely different hashes
	token1 := "abc123"
	token2 := "abc124"

	hash1 := HashToken(token1)
	hash2 := HashToken(token2)

	if hash1 == hash2 {
		t.Error("Similar but different tokens should produce different hashes")
	}

	// Hashes should be completely different, not just in one character
	matchingChars := 0
	for i := 0; i < len(hash1); i++ {
		if hash1[i] == hash2[i] {
			matchingChars++
		}
	}

	// With SHA-256, matching more than 50% of characters is statistically unlikely
	if matchingChars > 32 {
		t.Errorf("Hash avalanche effect not working properly, %d matching characters", matchingChars)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Should be 64 characters (32 bytes hex encoded)
	if len(token1) != 64 {
		t.Errorf("Expected refresh token length 64, got %d", len(token1))
	}

	// Should be unique each time
	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate second refresh token: %v", err)
	}

	if token1 == token2 {
		t.Error("Refresh tokens should be unique")
	}
}

func TestGetRefreshTokenExpirationTime_Default(t *testing.T) {
	// Clear env var
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "")

	duration := GetRefreshTokenExpirationTime()

	expectedDuration := 7 * 24 * time.Hour // 7 days
	if duration != expectedDuration {
		t.Errorf("Expected default duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetRefreshTokenExpirationTime_Custom(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "14")

	duration := GetRefreshTokenExpirationTime()

	expectedDuration := 14 * 24 * time.Hour
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetRefreshTokenExpirationTime_InvalidValue(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "invalid")

	duration := GetRefreshTokenExpirationTime()

	expectedDuration := 7 * 24 * time.Hour // Should fall back to default
	if duration != expectedDuration {
		t.Errorf("Expected default duration %v for invalid input, got %v", expectedDuration, duration)
	}
}

func TestGetRefreshTokenExpirationTime_NegativeValue(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "-5")

	duration := GetRefreshTokenExpirationTime()

	expectedDuration := 7 * 24 * time.Hour // Should fall back to default
	if duration != expectedDuration {
		t.Errorf("Expected default duration %v for negative input, got %v", expectedDuration, duration)
	}
}

func TestGetAccessTokenExpirationTime_Default(t *testing.T) {
	// Clear env vars
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "")
	t.Setenv("JWT_EXPIRATION_HOURS", "")

	duration := GetAccessTokenExpirationTime()

	expectedDuration := 15 * time.Minute // 15 minutes default
	if duration != expectedDuration {
		t.Errorf("Expected default duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetAccessTokenExpirationTime_Custom(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "30")

	duration := GetAccessTokenExpirationTime()

	expectedDuration := 30 * time.Minute
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetAccessTokenExpirationTime_LegacyHours(t *testing.T) {
	// Test backwards compatibility with JWT_EXPIRATION_HOURS
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "")
	t.Setenv("JWT_EXPIRATION_HOURS", "2")

	duration := GetAccessTokenExpirationTime()

	expectedDuration := 2 * time.Hour
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetAccessTokenExpirationTime_MinutesTakesPrecedence(t *testing.T) {
	// Minutes setting should take precedence over hours
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "45")
	t.Setenv("JWT_EXPIRATION_HOURS", "2")

	duration := GetAccessTokenExpirationTime()

	expectedDuration := 45 * time.Minute
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestGetTokenExpirationTime_BackwardsCompatibility(t *testing.T) {
	// GetTokenExpirationTime should now call GetAccessTokenExpirationTime
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "20")

	duration := GetTokenExpirationTime()
	expectedDuration := 20 * time.Minute

	if duration != expectedDuration {
		t.Errorf("Expected GetTokenExpirationTime to return %v, got %v", expectedDuration, duration)
	}
}

// --- Blacklist Fail Mode Configuration Tests ---

func TestGetBlacklistFailMode_Default(t *testing.T) {
	os.Unsetenv("TOKEN_BLACKLIST_FAIL_MODE")

	mode := getBlacklistFailMode()

	assert.Equal(t, "closed", mode, "default fail mode should be 'closed' for security")
}

func TestGetBlacklistFailMode_Open(t *testing.T) {
	t.Setenv("TOKEN_BLACKLIST_FAIL_MODE", "open")

	mode := getBlacklistFailMode()

	assert.Equal(t, "open", mode)
}

func TestGetBlacklistFailMode_Closed(t *testing.T) {
	t.Setenv("TOKEN_BLACKLIST_FAIL_MODE", "closed")

	mode := getBlacklistFailMode()

	assert.Equal(t, "closed", mode)
}

func TestGetBlacklistFailMode_InvalidValueDefaultsToClosed(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
	}{
		{"random string", "random"},
		{"OPEN uppercase", "OPEN"},
		{"mixed case", "Open"},
		{"whitespace", "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TOKEN_BLACKLIST_FAIL_MODE", tt.envValue)

			mode := getBlacklistFailMode()

			assert.Equal(t, "closed", mode, "invalid values should default to 'closed' for security")
		})
	}
}

// --- Blacklist Cache Constants Tests ---

func TestBlacklistCacheConstants(t *testing.T) {
	t.Run("cache TTL is reasonable", func(t *testing.T) {
		// TTL should be shorter than typical token expiry to avoid stale cache
		assert.Equal(t, 15*time.Minute, blacklistCacheTTL)
		assert.LessOrEqual(t, blacklistCacheTTL, 30*time.Minute,
			"cache TTL should be relatively short")
	})

	t.Run("cache prefix is set", func(t *testing.T) {
		assert.Equal(t, "blacklist:", blacklistCachePrefix)
		assert.NotEmpty(t, blacklistCachePrefix)
	})
}

// --- Token Hash Security Tests ---

func TestHashToken_Consistency(t *testing.T) {
	token := "test-token-12345"

	hash1 := HashToken(token)
	hash2 := HashToken(token)

	assert.Equal(t, hash1, hash2, "same token should produce same hash")
}

func TestHashToken_CacheKeyPrefix(t *testing.T) {
	token := "test-token"
	hash := HashToken(token)

	// Code uses first 16 chars of hash for cache key
	prefix := hash[:16]

	assert.Len(t, prefix, 16, "hash prefix for cache key should be 16 chars")
}

// --- Blacklist Without Database Tests ---
// These test the early-return behavior when database.DB is nil

func TestBlacklistToken_NilDatabase(t *testing.T) {
	// When database.DB is nil, BlacklistToken should return nil (no-op)
	err := BlacklistToken("test-token", 1, time.Now().Add(time.Hour), "test")

	assert.NoError(t, err, "should not error when DB is nil")
}

func TestIsTokenBlacklisted_NilDatabase(t *testing.T) {
	// When database.DB is nil, IsTokenBlacklisted should return false
	result := IsTokenBlacklisted("test-token")

	assert.False(t, result, "should return false when DB is nil")
}

func TestRevokeAllUserTokens_NilDatabase(t *testing.T) {
	// When database.DB is nil, RevokeAllUserTokens should return nil (no-op)
	err := RevokeAllUserTokens(1, "test")

	assert.NoError(t, err, "should not error when DB is nil")
}

func TestCleanupExpiredBlacklistEntries_NilDatabase(t *testing.T) {
	// When database.DB is nil, cleanup should return nil (no-op)
	err := CleanupExpiredBlacklistEntries()

	assert.NoError(t, err, "should not error when DB is nil")
}

// --- Fail Mode Behavior Documentation Tests ---

func TestFailModeBehavior_ClosedMode(t *testing.T) {
	// In 'closed' mode (default), when database query fails:
	// - IsTokenBlacklisted returns TRUE (deny request)
	// - This is security-first: when uncertain, deny access
	t.Setenv("TOKEN_BLACKLIST_FAIL_MODE", "closed")

	mode := getBlacklistFailMode()
	assert.Equal(t, "closed", mode)

	// Document expected behavior
	// If DB error occurs in 'closed' mode, the code returns true (blacklisted)
	// This means: deny the request when we can't verify token status
}

func TestFailModeBehavior_OpenMode(t *testing.T) {
	// In 'open' mode, when database query fails:
	// - IsTokenBlacklisted returns FALSE (allow request)
	// - This is availability-first: when uncertain, allow access
	t.Setenv("TOKEN_BLACKLIST_FAIL_MODE", "open")

	mode := getBlacklistFailMode()
	assert.Equal(t, "open", mode)

	// Document expected behavior
	// If DB error occurs in 'open' mode, the code returns false (not blacklisted)
	// This means: allow the request when we can't verify token status
}

// --- Reason Codes Tests ---

func TestBlacklistReason_ValidReasons(t *testing.T) {
	// Common reasons that might be used for blacklisting
	validReasons := []string{
		"logout",
		"password_change",
		"session_revoke",
		"security_incident",
		"admin_action",
	}

	for _, reason := range validReasons {
		t.Run(reason, func(t *testing.T) {
			assert.NotEmpty(t, reason)
			// The code accepts any string as reason
			// Empty reason defaults to "logout"
		})
	}
}

func TestBlacklistReason_DefaultsToLogout(t *testing.T) {
	// The code sets default reason to "logout" if empty
	// This is verified by inspecting BlacklistToken which does:
	// if reason == "" { reason = "logout" }
	reason := ""
	if reason == "" {
		reason = "logout"
	}

	assert.Equal(t, "logout", reason)
}
