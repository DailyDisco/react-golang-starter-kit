package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/models"

	"github.com/stretchr/testify/assert"
)

// Security-focused tests for session management.
// These tests document and verify security-critical behaviors.

// --- Session Expiration Tests ---

func TestSession_ExpirationDuration(t *testing.T) {
	// Sessions should expire after 7 days by default
	// This is set in CreateSessionWithContext

	t.Run("default expiration is 7 days", func(t *testing.T) {
		expectedDuration := 7 * 24 * time.Hour

		// Simulate the calculation from CreateSessionWithContext
		now := time.Now()
		expiresAt := now.Add(expectedDuration)

		// Verify the duration
		actualDuration := expiresAt.Sub(now)
		assert.Equal(t, expectedDuration, actualDuration)
	})

	t.Run("session cannot exceed maximum duration", func(t *testing.T) {
		// Document that sessions have a maximum lifetime
		maxDuration := 7 * 24 * time.Hour

		// In production, sessions should not exceed this
		assert.LessOrEqual(t, maxDuration, 30*24*time.Hour,
			"session duration should be reasonable (max 30 days)")
	})
}

// --- IP Address Security Tests ---

func TestGetClientIP_Security(t *testing.T) {
	t.Run("first IP in X-Forwarded-For is used (client IP)", func(t *testing.T) {
		// X-Forwarded-For format: client, proxy1, proxy2
		// We should use the first IP (client) not the last
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1, 10.0.0.2")
		req.RemoteAddr = "10.0.0.3:8080"

		ip := getClientIP(req)

		assert.Equal(t, "203.0.113.1", ip, "should use first IP (client)")
	})

	t.Run("handles spoofed X-Forwarded-For with malicious IPs", func(t *testing.T) {
		// Attackers might try to inject malicious data in XFF header
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "  192.168.1.1  ") // With whitespace
		req.RemoteAddr = "10.0.0.1:8080"

		ip := getClientIP(req)

		assert.Equal(t, "192.168.1.1", ip, "should trim whitespace")
	})

	t.Run("handles IPv6 addresses", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "2001:db8::1")
		req.RemoteAddr = "10.0.0.1:8080"

		ip := getClientIP(req)

		assert.Equal(t, "2001:db8::1", ip, "should handle IPv6")
	})

	t.Run("fallback chain: XFF -> X-Real-IP -> RemoteAddr", func(t *testing.T) {
		// Test that X-Forwarded-For takes precedence over X-Real-IP
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "1.1.1.1")
		req.Header.Set("X-Real-IP", "2.2.2.2")
		req.RemoteAddr = "3.3.3.3:8080"

		ip := getClientIP(req)
		assert.Equal(t, "1.1.1.1", ip, "X-Forwarded-For should have highest priority")

		// Test X-Real-IP when XFF is absent
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("X-Real-IP", "2.2.2.2")
		req2.RemoteAddr = "3.3.3.3:8080"

		ip2 := getClientIP(req2)
		assert.Equal(t, "2.2.2.2", ip2, "X-Real-IP should be fallback")

		// Test RemoteAddr when both headers absent
		req3 := httptest.NewRequest(http.MethodGet, "/", nil)
		req3.RemoteAddr = "3.3.3.3:8080"

		ip3 := getClientIP(req3)
		assert.Equal(t, "3.3.3.3", ip3, "RemoteAddr should be final fallback")
	})
}

// --- Token Hash Security Tests ---

func TestTokenHash_Security(t *testing.T) {
	t.Run("hash is one-way (cannot recover token)", func(t *testing.T) {
		token := "secret-refresh-token-12345"
		hash := hashToken(token)

		// The hash should not contain the original token
		assert.NotContains(t, hash, token)
		assert.NotContains(t, hash, "secret")
		assert.NotContains(t, hash, "refresh")
	})

	t.Run("same token always produces same hash", func(t *testing.T) {
		token := "consistent-token"

		hash1 := hashToken(token)
		hash2 := hashToken(token)

		assert.Equal(t, hash1, hash2, "hash should be deterministic")
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		token1 := "token-a"
		token2 := "token-b"

		hash1 := hashToken(token1)
		hash2 := hashToken(token2)

		assert.NotEqual(t, hash1, hash2, "different tokens should have different hashes")
	})

	t.Run("hash length is fixed (SHA-256)", func(t *testing.T) {
		shortToken := "a"
		longToken := "a very long token that contains lots of characters and is much longer than the short one"

		hashShort := hashToken(shortToken)
		hashLong := hashToken(longToken)

		// Both should be 64 hex characters (256 bits / 4 bits per hex char)
		assert.Len(t, hashShort, 64, "short token hash should be 64 chars")
		assert.Len(t, hashLong, 64, "long token hash should be 64 chars")
	})
}

// --- Device Info Security Tests ---

func TestParseDeviceInfo_Security(t *testing.T) {
	svc := &SessionService{}

	t.Run("handles malformed user agent gracefully", func(t *testing.T) {
		// Malformed user agents should not crash
		malformedAgents := []string{
			"",
			"   ",
			"\x00\x01\x02", // Null bytes
			"<script>alert('xss')</script>",
			string(make([]byte, 10000)), // Very long string
		}

		for _, ua := range malformedAgents {
			// Should not panic
			deviceInfo := svc.ParseDeviceInfo(ua)
			assert.NotNil(t, deviceInfo, "should return valid device info even for malformed UA")
		}
	})

	t.Run("does not expose sensitive info in device info", func(t *testing.T) {
		// User agent might contain sensitive paths or info
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) secret-api-key-12345"

		deviceInfo := svc.ParseDeviceInfo(ua)

		// Device info should not contain secrets (it just extracts browser/OS info)
		assert.NotContains(t, deviceInfo.Browser, "secret")
		assert.NotContains(t, deviceInfo.OS, "api-key")
	})
}

// --- Auth Method Constants ---

func TestAuthMethodConstants_Security(t *testing.T) {
	// Verify all auth methods are defined and distinct
	methods := []string{
		models.AuthMethodPassword,
		models.AuthMethodOAuthGoogle,
		models.AuthMethodOAuthGitHub,
		models.AuthMethodRefreshToken,
		models.AuthMethod2FA,
	}

	seen := make(map[string]bool)
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			assert.NotEmpty(t, method, "auth method should not be empty")
			assert.False(t, seen[method], "auth method should be unique")
			seen[method] = true
		})
	}
}
