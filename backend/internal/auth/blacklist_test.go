package auth

import (
	"testing"
	"time"
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
