package services

import (
	"encoding/json"
	"os"
	"testing"
)

// ============ TOTP Helper Function Tests ============

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"simple token", "abc123"},
		{"empty token", ""},
		{"long token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"},
		{"special characters", "test@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := HashToken(tt.token)
			hash2 := HashToken(tt.token)

			// Hash should be consistent
			if hash1 != hash2 {
				t.Errorf("HashToken() returned different hashes for same input: %q != %q", hash1, hash2)
			}

			// Hash should be 64 chars (SHA-256 hex)
			if len(hash1) != 64 {
				t.Errorf("HashToken() length = %d, want 64", len(hash1))
			}
		})
	}
}

func TestHashToken_Uniqueness(t *testing.T) {
	hash1 := HashToken("token1")
	hash2 := HashToken("token2")

	if hash1 == hash2 {
		t.Error("HashToken() should return different hashes for different inputs")
	}
}

func TestFormatBackupCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"valid 8-char code", "ABCD1234", "ABCD-1234"},
		{"already formatted", "ABCD-1234", "ABCD-1234"}, // Will return as-is since len != 8
		{"too short", "ABC", "ABC"},
		{"too long", "ABCD12345", "ABCD12345"},
		{"empty", "", ""},
		{"exactly 8 with lowercase", "abcd1234", "abcd-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBackupCode(tt.code)
			if result != tt.expected {
				t.Errorf("FormatBackupCode(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestUnformatBackupCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"formatted code", "ABCD-1234", "ABCD1234"},
		{"unformatted code", "ABCD1234", "ABCD1234"},
		{"multiple hyphens", "AB-CD-12-34", "ABCD1234"},
		{"no hyphens", "ABCDEFGH", "ABCDEFGH"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnformatBackupCode(tt.code)
			if result != tt.expected {
				t.Errorf("UnformatBackupCode(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestFormatUnformatBackupCode_RoundTrip(t *testing.T) {
	original := "ABCD1234"

	formatted := FormatBackupCode(original)
	if formatted != "ABCD-1234" {
		t.Errorf("FormatBackupCode(%q) = %q, want %q", original, formatted, "ABCD-1234")
	}

	unformatted := UnformatBackupCode(formatted)
	if unformatted != original {
		t.Errorf("UnformatBackupCode(%q) = %q, want %q", formatted, unformatted, original)
	}
}

func TestGenerateRandomCode(t *testing.T) {
	lengths := []int{4, 8, 16, 32}

	for _, length := range lengths {
		t.Run("length_"+string(rune(length+'0')), func(t *testing.T) {
			code := generateRandomCode(length)

			if len(code) != length {
				t.Errorf("generateRandomCode(%d) length = %d, want %d", length, len(code), length)
			}

			// Verify only contains valid characters (A-Z, 0-9)
			for _, c := range code {
				if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
					t.Errorf("generateRandomCode(%d) contains invalid character: %c", length, c)
				}
			}
		})
	}
}

func TestGenerateRandomCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)

	// Generate 100 codes and check for uniqueness
	for i := 0; i < 100; i++ {
		code := generateRandomCode(8)
		if codes[code] {
			t.Errorf("generateRandomCode(8) generated duplicate code: %s", code)
		}
		codes[code] = true
	}
}

// ============ TOTPService Constructor Tests ============

func TestNewTOTPService(t *testing.T) {
	service := NewTOTPService()
	if service == nil {
		t.Fatal("NewTOTPService() returned nil")
	}

	// Should have an issuer
	if service.issuer == "" {
		t.Error("NewTOTPService() should set issuer")
	}

	// Should have encryption key (32 bytes)
	if len(service.encryptionKey) != 32 {
		t.Errorf("NewTOTPService() encryption key length = %d, want 32", len(service.encryptionKey))
	}
}

func TestNewTOTPService_DefaultIssuer(t *testing.T) {
	// Save and clear environment variables
	oldTOTPIssuer := os.Getenv("TOTP_ISSUER")
	oldSiteName := os.Getenv("SITE_NAME")
	os.Setenv("TOTP_ISSUER", "")
	os.Setenv("SITE_NAME", "")
	defer func() {
		os.Setenv("TOTP_ISSUER", oldTOTPIssuer)
		os.Setenv("SITE_NAME", oldSiteName)
	}()

	service := NewTOTPService()

	// Should use default issuer "MyApp"
	if service.issuer != "MyApp" {
		t.Errorf("NewTOTPService() issuer = %q, want %q", service.issuer, "MyApp")
	}
}

// ============ Encryption/Decryption Round-Trip Tests ============

func TestTOTPService_EncryptDecrypt_RoundTrip(t *testing.T) {
	service := NewTOTPService()

	secrets := []string{
		"JBSWY3DPEHPK3PXP",           // Typical TOTP secret
		"GEZDGNBVGY3TQOJQ",           // Another base32 secret
		"short",                      // Short string
		"a-very-long-secret-key-123", // Longer string
	}

	for _, secret := range secrets {
		t.Run(secret, func(t *testing.T) {
			encrypted, err := service.encryptSecret(secret)
			if err != nil {
				t.Fatalf("encryptSecret() error = %v", err)
			}

			// Encrypted should be different from original
			if encrypted == secret {
				t.Error("encryptSecret() should produce different output")
			}

			decrypted, err := service.decryptSecret(encrypted)
			if err != nil {
				t.Fatalf("decryptSecret() error = %v", err)
			}

			if decrypted != secret {
				t.Errorf("decryptSecret() = %q, want %q", decrypted, secret)
			}
		})
	}
}

func TestTOTPService_DecryptSecret_InvalidData(t *testing.T) {
	service := NewTOTPService()

	tests := []struct {
		name      string
		encrypted string
	}{
		{"not base64", "not-valid-base64!!!"},
		{"too short", "YWJj"}, // "abc" in base64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.decryptSecret(tt.encrypted)
			if err == nil {
				t.Error("decryptSecret() should return error for invalid data")
			}
		})
	}
}

// ============ Backup Code Generation Tests ============

func TestTOTPService_GenerateBackupCodes(t *testing.T) {
	service := NewTOTPService()

	counts := []int{5, 10, 15}

	for _, count := range counts {
		t.Run("count_"+string(rune(count+'0')), func(t *testing.T) {
			codes, hashedCodes, err := service.generateBackupCodes(count)
			if err != nil {
				t.Fatalf("generateBackupCodes(%d) error = %v", count, err)
			}

			if len(codes) != count {
				t.Errorf("generateBackupCodes(%d) returned %d codes, want %d", count, len(codes), count)
			}

			if len(hashedCodes) != count {
				t.Errorf("generateBackupCodes(%d) returned %d hashed codes, want %d", count, len(hashedCodes), count)
			}

			// Each code should be 8 characters
			for i, code := range codes {
				if len(code) != 8 {
					t.Errorf("codes[%d] length = %d, want 8", i, len(code))
				}
			}

			// Each hashed code should be bcrypt hash (starts with $2)
			for i, hash := range hashedCodes {
				if len(hash) < 50 || hash[:2] != "$2" {
					t.Errorf("hashedCodes[%d] doesn't appear to be bcrypt hash", i)
				}
			}
		})
	}
}

func TestTOTPService_GenerateBackupCodes_Unique(t *testing.T) {
	service := NewTOTPService()

	codes, _, err := service.generateBackupCodes(10)
	if err != nil {
		t.Fatalf("generateBackupCodes() error = %v", err)
	}

	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Error("generateBackupCodes() generated duplicate codes")
		}
		seen[code] = true
	}
}

// ============ Backup Code Validation Tests ============

func TestTOTPService_ValidateBackupCode(t *testing.T) {
	service := NewTOTPService()

	// Generate backup codes
	codes, hashedCodes, err := service.generateBackupCodes(5)
	if err != nil {
		t.Fatalf("generateBackupCodes() error = %v", err)
	}

	// Convert to JSON for validation
	hashedJSON, _ := json.Marshal(hashedCodes)

	// Test valid code
	valid, remaining := service.validateBackupCode(hashedJSON, codes[0])
	if !valid {
		t.Error("validateBackupCode() should return true for valid code")
	}
	if len(remaining) != 4 {
		t.Errorf("validateBackupCode() remaining codes = %d, want 4", len(remaining))
	}

	// Test invalid code
	valid, remaining = service.validateBackupCode(hashedJSON, "INVALID1")
	if valid {
		t.Error("validateBackupCode() should return false for invalid code")
	}
	if len(remaining) != 5 {
		t.Errorf("validateBackupCode() remaining codes = %d, want 5", len(remaining))
	}
}

func TestTOTPService_ValidateBackupCode_EmptyHashes(t *testing.T) {
	service := NewTOTPService()

	emptyJSON := []byte("[]")
	valid, remaining := service.validateBackupCode(emptyJSON, "ABCD1234")

	if valid {
		t.Error("validateBackupCode() should return false for empty hashes")
	}
	if len(remaining) != 0 {
		t.Errorf("validateBackupCode() remaining = %d, want 0", len(remaining))
	}
}

func TestTOTPService_ValidateBackupCode_InvalidJSON(t *testing.T) {
	service := NewTOTPService()

	invalidJSON := []byte("not valid json")
	valid, remaining := service.validateBackupCode(invalidJSON, "ABCD1234")

	if valid {
		t.Error("validateBackupCode() should return false for invalid JSON")
	}
	if remaining != nil {
		t.Error("validateBackupCode() should return nil remaining for invalid JSON")
	}
}
