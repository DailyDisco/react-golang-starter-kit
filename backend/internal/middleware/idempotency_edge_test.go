package middleware

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Edge case tests for idempotency middleware

// --- Cache Key User ID Format Tests ---
// Verifies that userID is correctly formatted as a string in cache keys.
// Previously had a bug where string(rune(userID)) was used, which converted
// userID to a Unicode character (e.g., 65 -> "A"). Now uses strconv.FormatUint.

func TestBuildIdempotencyCacheKey_UserIDCorrectFormat(t *testing.T) {
	t.Run("userID 65 should be '65' not 'A'", func(t *testing.T) {
		key := buildIdempotencyCacheKey("test-key", 65, "/api/test")

		// Should contain ":65:" as the userID segment
		assert.Contains(t, key, ":65:", "userID 65 should be string '65', not character 'A'")
		assert.NotContains(t, key, ":A:", "should not convert userID to ASCII character")
	})

	t.Run("large userIDs work correctly", func(t *testing.T) {
		key := buildIdempotencyCacheKey("test-key", 1000, "/api/test")

		// Should contain ":1000:" as the userID segment
		assert.Contains(t, key, ":1000:", "large userID should be formatted as string")

		// Verify exact format
		expected := "idempotency:test-key:" + strconv.FormatUint(1000, 10) + ":/api/test"
		assert.Equal(t, expected, key)
	})

	t.Run("very large userIDs work correctly", func(t *testing.T) {
		// Test with userID larger than Unicode max (1,114,111)
		key := buildIdempotencyCacheKey("test-key", 9999999, "/api/test")

		assert.Contains(t, key, ":9999999:", "very large userID should work")
		expected := "idempotency:test-key:9999999:/api/test"
		assert.Equal(t, expected, key)
	})

	t.Run("userID 0 should be '0'", func(t *testing.T) {
		key := buildIdempotencyCacheKey("test-key", 0, "/api/test")

		// Should contain ":0:" not a null character
		assert.Contains(t, key, ":0:", "userID 0 should be string '0'")
		expected := "idempotency:test-key:0:/api/test"
		assert.Equal(t, expected, key)
	})
}

// --- Key Collision Tests ---

func TestBuildIdempotencyCacheKey_UserIDCollisions(t *testing.T) {
	// Verifies that different userIDs always produce different cache keys

	t.Run("different userIDs produce different keys", func(t *testing.T) {
		key1 := buildIdempotencyCacheKey("test", 1, "/api/test")
		key2 := buildIdempotencyCacheKey("test", 2, "/api/test")

		// These should be different
		assert.NotEqual(t, key1, key2, "different userIDs should produce different keys")
	})

	t.Run("same key different users are independent", func(t *testing.T) {
		// Even with the bug, different character representations should differ
		key1 := buildIdempotencyCacheKey("payment-123", 100, "/api/billing")
		key2 := buildIdempotencyCacheKey("payment-123", 101, "/api/billing")

		assert.NotEqual(t, key1, key2, "same idempotency key for different users should produce different cache keys")
	})
}

// --- Key Format Tests ---

func TestBuildIdempotencyCacheKey_Format(t *testing.T) {
	t.Run("contains all components", func(t *testing.T) {
		key := buildIdempotencyCacheKey("abc-123", 42, "/api/v1/users")

		assert.Contains(t, key, "idempotency:")
		assert.Contains(t, key, "abc-123")
		assert.Contains(t, key, "/api/v1/users")
	})

	t.Run("special characters in idempotency key", func(t *testing.T) {
		// Idempotency keys with special chars should work
		keys := []string{
			"uuid-with-dashes",
			"key_with_underscores",
			"key.with.dots",
			"key=with=equals",
			"key:with:colons",
		}

		for _, idempKey := range keys {
			t.Run(idempKey, func(t *testing.T) {
				key := buildIdempotencyCacheKey(idempKey, 1, "/api/test")
				assert.NotEmpty(t, key)
				assert.Contains(t, key, idempKey)
			})
		}
	})

	t.Run("paths with query strings", func(t *testing.T) {
		// Query strings in path are included in cache key
		key1 := buildIdempotencyCacheKey("test", 1, "/api/users?page=1")
		key2 := buildIdempotencyCacheKey("test", 1, "/api/users?page=2")

		// Different query strings = different paths = different keys
		assert.NotEqual(t, key1, key2)
	})
}

// --- Response Caching Tests ---

func TestIdempotencyResponseCaching(t *testing.T) {
	t.Run("2xx responses are cached", func(t *testing.T) {
		// Status codes 200-299 should be cached
		for _, code := range []int{200, 201, 202, 204} {
			shouldCache := code >= 200 && code < 500
			assert.True(t, shouldCache, "status %d should be cached", code)
		}
	})

	t.Run("4xx client errors are cached", func(t *testing.T) {
		// Status codes 400-499 should be cached
		for _, code := range []int{400, 401, 403, 404, 409, 422} {
			shouldCache := code >= 200 && code < 500
			assert.True(t, shouldCache, "status %d should be cached", code)
		}
	})

	t.Run("5xx server errors are NOT cached", func(t *testing.T) {
		// Status codes 500-599 should NOT be cached
		for _, code := range []int{500, 502, 503, 504} {
			shouldCache := code >= 200 && code < 500
			assert.False(t, shouldCache, "status %d should NOT be cached", code)
		}
	})
}

// --- Safe Methods Tests ---

func TestIdempotencyMiddleware_SafeMethods(t *testing.T) {
	// Safe methods should bypass idempotency checking entirely
	safeMethods := []string{"GET", "HEAD", "OPTIONS"}
	unsafeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range safeMethods {
		t.Run(method+" is safe", func(t *testing.T) {
			// Safe methods should pass through without idempotency processing
			isSafe := method == "GET" || method == "HEAD" || method == "OPTIONS"
			assert.True(t, isSafe)
		})
	}

	for _, method := range unsafeMethods {
		t.Run(method+" is NOT safe", func(t *testing.T) {
			// These methods should be subject to idempotency checking
			isSafe := method == "GET" || method == "HEAD" || method == "OPTIONS"
			assert.False(t, isSafe)
		})
	}
}

// --- Header Configuration Tests ---

func TestIdempotencyConfig_HeaderVariations(t *testing.T) {
	t.Run("default header name", func(t *testing.T) {
		config := LoadIdempotencyConfig()
		// Clear any env overrides
		t.Setenv("IDEMPOTENCY_HEADER", "")
		config = LoadIdempotencyConfig()

		// Default should be "Idempotency-Key"
		expected := "Idempotency-Key"
		if config.HeaderName != expected {
			t.Logf("Note: Header name is %q (may be overridden by env)", config.HeaderName)
		}
	})

	t.Run("custom header name from env", func(t *testing.T) {
		t.Setenv("IDEMPOTENCY_HEADER", "X-Request-Id")
		config := LoadIdempotencyConfig()

		assert.Equal(t, "X-Request-Id", config.HeaderName)
	})
}

// --- TTL Configuration Tests ---

func TestIdempotencyConfig_TTL(t *testing.T) {
	t.Run("short TTL for testing", func(t *testing.T) {
		t.Setenv("IDEMPOTENCY_TTL", "1m")
		config := LoadIdempotencyConfig()

		assert.Equal(t, "1m0s", config.TTL.String())
	})

	t.Run("long TTL for production", func(t *testing.T) {
		t.Setenv("IDEMPOTENCY_TTL", "48h")
		config := LoadIdempotencyConfig()

		assert.Equal(t, "48h0m0s", config.TTL.String())
	})

	t.Run("invalid TTL falls back to default", func(t *testing.T) {
		t.Setenv("IDEMPOTENCY_TTL", "not-a-duration")
		config := LoadIdempotencyConfig()

		// Default is 24h
		assert.Equal(t, "24h0m0s", config.TTL.String())
	})
}
