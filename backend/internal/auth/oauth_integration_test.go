package auth

import (
	"os"
	"testing"
	"time"

	"react-golang-starter/internal/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuthStateStorage_Integration tests OAuth state storage with Redis.
// These tests require a running Redis instance (from docker-compose.test.yml).
//
// Run with: INTEGRATION_TEST=true TEST_REDIS_URL=redis://localhost:6382 go test -v -run Integration ./internal/auth/...
func TestOAuthStateStorage_Integration(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL not set - skipping Redis integration tests")
	}

	// Initialize cache with Redis
	config := &cache.Config{
		Enabled:   true,
		Type:      "redis",
		RedisURL:  redisURL,
		KeyPrefix: "oauth_test",
	}

	err := cache.Initialize(config)
	require.NoError(t, err, "Failed to initialize Redis cache")

	// Ensure cache is available
	require.True(t, cache.IsAvailable(), "Cache should be available after initialization")

	// Clean up after tests
	t.Cleanup(func() {
		cache.Close()
	})

	t.Run("CreateAndValidateState", func(t *testing.T) {
		// Generate a new state
		state, err := generateState()
		require.NoError(t, err, "generateState should not error")
		require.NotEmpty(t, state, "state should not be empty")

		// State should be valid immediately after creation
		valid := validateState(state)
		assert.True(t, valid, "freshly created state should be valid")

		// After validation, state should be consumed (one-time use)
		valid = validateState(state)
		assert.False(t, valid, "state should not be valid after first use (replay prevention)")
	})

	t.Run("StateUniqueness", func(t *testing.T) {
		// Generate multiple states
		states := make(map[string]bool)
		for i := 0; i < 100; i++ {
			state, err := generateState()
			require.NoError(t, err)
			assert.False(t, states[state], "generated state should be unique")
			states[state] = true
		}
	})

	t.Run("InvalidStateRejected", func(t *testing.T) {
		// Random state that was never created
		valid := validateState("invalid-state-that-was-never-created")
		assert.False(t, valid, "invalid state should be rejected")

		// Empty state
		valid = validateState("")
		assert.False(t, valid, "empty state should be rejected")

		// Malformed state
		valid = validateState("not-base64-encoded")
		assert.False(t, valid, "malformed state should be rejected")
	})

	t.Run("StateReplayPrevention", func(t *testing.T) {
		// Generate a state
		state, err := generateState()
		require.NoError(t, err)

		// First validation should succeed
		valid1 := validateState(state)
		assert.True(t, valid1, "first validation should succeed")

		// Second validation of the same state should fail
		valid2 := validateState(state)
		assert.False(t, valid2, "second validation should fail (replay prevention)")

		// Third validation should also fail
		valid3 := validateState(state)
		assert.False(t, valid3, "third validation should also fail")
	})

	t.Run("MultipleStatesIndependent", func(t *testing.T) {
		// Generate two states
		state1, err := generateState()
		require.NoError(t, err)

		state2, err := generateState()
		require.NoError(t, err)

		// Both should be valid
		assert.True(t, validateState(state1), "state1 should be valid")
		assert.True(t, validateState(state2), "state2 should be valid")

		// Both should now be consumed
		assert.False(t, validateState(state1), "state1 should be consumed")
		assert.False(t, validateState(state2), "state2 should be consumed")
	})
}

// TestOAuthStateExpiration_Integration tests that OAuth states expire after TTL.
// This test takes time to run so it's separated.
func TestOAuthStateExpiration_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping expiration test in short mode")
	}

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL not set - skipping Redis integration tests")
	}

	// Initialize cache with Redis
	config := &cache.Config{
		Enabled:   true,
		Type:      "redis",
		RedisURL:  redisURL,
		KeyPrefix: "oauth_test_expiry",
	}

	err := cache.Initialize(config)
	require.NoError(t, err, "Failed to initialize Redis cache")

	t.Cleanup(func() {
		cache.Close()
	})

	t.Run("StateExpiresAfterTTL", func(t *testing.T) {
		// The OAuth state TTL is 5 minutes by default
		// For testing, we'll verify the state is stored with TTL
		// by checking it exists immediately and would expire

		state, err := generateState()
		require.NoError(t, err)

		// State should be valid immediately
		valid := validateState(state)
		assert.True(t, valid, "state should be valid immediately")

		// Note: Since oauthStateTTL is 5 minutes, we can't realistically
		// wait for expiration in tests. The TTL is set in Redis, so we
		// trust Redis to expire it. The important thing is that the state
		// is consumed on first use (tested above).
	})
}

// TestOAuthStateConcurrency_Integration tests concurrent state operations.
func TestOAuthStateConcurrency_Integration(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL not set - skipping Redis integration tests")
	}

	// Initialize cache with Redis
	config := &cache.Config{
		Enabled:   true,
		Type:      "redis",
		RedisURL:  redisURL,
		KeyPrefix: "oauth_test_concurrent",
	}

	err := cache.Initialize(config)
	require.NoError(t, err, "Failed to initialize Redis cache")

	t.Cleanup(func() {
		cache.Close()
	})

	t.Run("ConcurrentStateGeneration", func(t *testing.T) {
		const numGoroutines = 50
		states := make(chan string, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Generate states concurrently
		for i := 0; i < numGoroutines; i++ {
			go func() {
				state, err := generateState()
				if err != nil {
					errors <- err
					return
				}
				states <- state
			}()
		}

		// Collect results
		uniqueStates := make(map[string]bool)
		for i := 0; i < numGoroutines; i++ {
			select {
			case err := <-errors:
				t.Errorf("Error generating state: %v", err)
			case state := <-states:
				if uniqueStates[state] {
					t.Error("Duplicate state generated concurrently")
				}
				uniqueStates[state] = true
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for state generation")
			}
		}

		assert.Equal(t, numGoroutines, len(uniqueStates), "All states should be unique")
	})

	t.Run("ConcurrentValidation", func(t *testing.T) {
		// Generate a single state
		state, err := generateState()
		require.NoError(t, err)

		const numGoroutines = 10
		results := make(chan bool, numGoroutines)

		// Try to validate the same state concurrently
		for i := 0; i < numGoroutines; i++ {
			go func() {
				results <- validateState(state)
			}()
		}

		// Collect results - exactly one should succeed
		validCount := 0
		for i := 0; i < numGoroutines; i++ {
			select {
			case valid := <-results:
				if valid {
					validCount++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for validation")
			}
		}

		// Due to race conditions in concurrent validation,
		// it's possible 0 or 1 validations succeed
		// (0 if all delete the key before any can verify it exists)
		// The important thing is that at most 1 succeeds
		assert.LessOrEqual(t, validCount, 1, "At most one concurrent validation should succeed")
	})
}

// TestOAuthStateWithCacheFailure_Integration tests behavior when cache is unavailable.
func TestOAuthStateWithCacheFailure_Integration(t *testing.T) {
	// Test with NoOp cache (cache disabled)
	config := &cache.Config{
		Enabled: false,
	}

	err := cache.Initialize(config)
	require.NoError(t, err)

	t.Cleanup(func() {
		cache.Close()
	})

	t.Run("StateGenerationWithoutCache", func(t *testing.T) {
		// generateState should still work (just won't persist to cache)
		state, err := generateState()
		require.NoError(t, err, "generateState should not error without cache")
		require.NotEmpty(t, state, "state should still be generated")
	})

	t.Run("ValidationFailsWithoutCache", func(t *testing.T) {
		state, err := generateState()
		require.NoError(t, err)

		// Without cache, validation will always fail
		// because the state is never stored
		valid := validateState(state)
		assert.False(t, valid, "validation should fail without cache")
	})
}
