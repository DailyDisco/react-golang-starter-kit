package auth

import (
	"os"
	"testing"
	"time"

	"react-golang-starter/internal/models"
)

// ensureJWTSecret sets JWT_SECRET if not already set (for tests that need it)
func ensureJWTSecret(t *testing.T) {
	t.Helper()
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-integration-tests-only")
	}
}

// TestTokenGeneration verifies that access and refresh tokens are generated correctly
func TestTokenGeneration(t *testing.T) {
	ensureJWTSecret(t)

	// Test access token generation
	t.Run("GenerateToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  "user",
		}
		token, err := GenerateToken(user)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}

		// Verify token can be parsed
		claims, err := ValidateJWT(token)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}
		if claims.UserID != 1 {
			t.Errorf("expected user ID 1, got %d", claims.UserID)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", claims.Email)
		}
		if claims.Role != "user" {
			t.Errorf("expected role user, got %s", claims.Role)
		}
	})

	// Test refresh token generation
	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty refresh token")
		}
		// Refresh token should be 32 bytes base64 encoded (44 chars)
		if len(token) < 40 {
			t.Errorf("refresh token too short: %d chars", len(token))
		}
	})

	// Test that multiple refresh tokens are unique
	t.Run("RefreshTokensAreUnique", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := GenerateRefreshToken()
			if err != nil {
				t.Fatalf("failed to generate refresh token: %v", err)
			}
			if tokens[token] {
				t.Error("duplicate refresh token generated")
			}
			tokens[token] = true
		}
	})
}

// TestTokenExpiration verifies token expiration settings
func TestTokenExpiration(t *testing.T) {
	t.Run("AccessTokenExpiration", func(t *testing.T) {
		expiration := GetAccessTokenExpirationTime()
		// Default should be 15 minutes
		if expiration < 15*time.Minute || expiration > 60*time.Minute {
			t.Errorf("unexpected access token expiration: %v", expiration)
		}
	})

	t.Run("RefreshTokenExpiration", func(t *testing.T) {
		expiration := GetRefreshTokenExpirationTime()
		// Default should be 7 days
		if expiration < 24*time.Hour || expiration > 30*24*time.Hour {
			t.Errorf("unexpected refresh token expiration: %v", expiration)
		}
	})
}

// TestImpersonationToken verifies impersonation token generation
func TestImpersonationToken(t *testing.T) {
	ensureJWTSecret(t)

	targetUser := &models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  "user",
	}

	token, err := GenerateImpersonationToken(targetUser, 999)
	if err != nil {
		t.Fatalf("failed to generate impersonation token: %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate impersonation token: %v", err)
	}

	// Should have original user ID set
	if claims.OriginalUserID != 999 {
		t.Errorf("expected original user ID 999, got %d", claims.OriginalUserID)
	}

	// Should be impersonating user 1
	if claims.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", claims.UserID)
	}
}

// TestPasswordValidation verifies password requirements
func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "SecurePass123", false},
		{"too short", "Pass1", true},
		{"no uppercase", "password123", true},
		{"no lowercase", "PASSWORD123", true},
		{"no digit", "PasswordOnly", true},
		{"just meets requirements", "Aa123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr = %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// TestEmailValidation verifies email format requirements
func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"empty", "", true},
		{"too long", "a" + string(make([]byte, 300)) + "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr = %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

// TestPasswordHashing verifies bcrypt hashing and verification
func TestPasswordHashing(t *testing.T) {
	password := "SecurePassword123"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Verify correct password
	if !CheckPassword(password, hash) {
		t.Error("password verification failed for correct password")
	}

	// Verify incorrect password
	if CheckPassword("WrongPassword123", hash) {
		t.Error("password verification succeeded for incorrect password")
	}

	// Verify that hashing is not deterministic (different salts)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password second time: %v", err)
	}
	if hash == hash2 {
		t.Error("expected different hashes for same password")
	}
}

// TestTokenHashing verifies token hashing for blacklist
func TestTokenHashing(t *testing.T) {
	token := "test-token-12345"

	hash1 := HashToken(token)
	hash2 := HashToken(token)

	// Same token should produce same hash
	if hash1 != hash2 {
		t.Error("expected same hash for same token")
	}

	// Different tokens should produce different hashes
	differentHash := HashToken("different-token")
	if hash1 == differentHash {
		t.Error("expected different hash for different token")
	}

	// Hash should not be the token itself
	if hash1 == token {
		t.Error("hash should not equal original token")
	}
}
