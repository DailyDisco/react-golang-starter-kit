package services

import (
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
