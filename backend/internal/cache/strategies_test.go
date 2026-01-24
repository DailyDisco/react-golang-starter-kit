package cache

import (
	"context"
	"testing"
)

// ============ Cache Key Generation Tests ============

func TestBlacklistCacheKey(t *testing.T) {
	tests := []struct {
		name        string
		tokenPrefix string
		expectedKey string
	}{
		{"simple prefix", "abc123", "blacklist:abc123"},
		{"long prefix", "a1b2c3d4e5f6", "blacklist:a1b2c3d4e5f6"},
		{"empty prefix", "", "blacklist:"},
		{"special chars", "token-with-dash", "blacklist:token-with-dash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := BlacklistCacheKey(tt.tokenPrefix)
			if key != tt.expectedKey {
				t.Errorf("BlacklistCacheKey(%q) = %q, want %q", tt.tokenPrefix, key, tt.expectedKey)
			}
		})
	}
}

func TestUserCacheKey(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		expectedKey string
	}{
		{"user 1", 1, "user:1"},
		{"user 42", 42, "user:42"},
		{"user 0", 0, "user:0"},
		{"large user ID", 999999, "user:999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := UserCacheKey(tt.userID)
			if key != tt.expectedKey {
				t.Errorf("UserCacheKey(%d) = %q, want %q", tt.userID, key, tt.expectedKey)
			}
		})
	}
}

func TestSessionCacheKey(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		expectedKey string
	}{
		{"simple session", "sess123", "session:sess123"},
		{"uuid session", "550e8400-e29b-41d4-a716-446655440000", "session:550e8400-e29b-41d4-a716-446655440000"},
		{"empty session", "", "session:"},
		{"complex session ID", "user_123_session_456", "session:user_123_session_456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := SessionCacheKey(tt.sessionID)
			if key != tt.expectedKey {
				t.Errorf("SessionCacheKey(%q) = %q, want %q", tt.sessionID, key, tt.expectedKey)
			}
		})
	}
}

func TestFeatureFlagsCacheKey(t *testing.T) {
	if FeatureFlagsCacheKey != "feature_flags:all" {
		t.Errorf("FeatureFlagsCacheKey = %q, want 'feature_flags:all'", FeatureFlagsCacheKey)
	}
}

// ============ Exists Tests ============

func TestExists_NilInstance(t *testing.T) {
	// Save and restore the instance
	originalInstance := instance
	instance = nil
	defer func() { instance = originalInstance }()

	ctx := context.Background()
	exists := Exists(ctx, "test-key")

	if exists {
		t.Error("Exists() should return false when instance is nil")
	}
}

// ============ Invalidate Tests ============

func TestInvalidate_NilInstance(t *testing.T) {
	// Save and restore the instance
	originalInstance := instance
	instance = nil
	defer func() { instance = originalInstance }()

	ctx := context.Background()

	// Should not panic with nil instance
	Invalidate(ctx, "test-key")
}

// ============ InvalidatePattern Tests ============

func TestInvalidatePattern_NilInstance(t *testing.T) {
	// Save and restore the instance
	originalInstance := instance
	instance = nil
	defer func() { instance = originalInstance }()

	ctx := context.Background()

	// Should not panic with nil instance
	InvalidatePattern(ctx, "pattern:*")
}

// ============ SetIfNotExists Behavior Tests ============

func TestSetIfNotExists_KeyAlreadyExists(t *testing.T) {
	// This test documents expected behavior:
	// When a key already exists, SetIfNotExists should return false
	// Detailed integration tests for this are in cache_test.go
}

// ============ Key Format Tests ============

func TestCacheKeyFormats(t *testing.T) {
	// Test that all key formats follow expected patterns
	tests := []struct {
		name    string
		key     string
		pattern string
	}{
		{"blacklist key", BlacklistCacheKey("abc"), "blacklist:"},
		{"user key", UserCacheKey(1), "user:"},
		{"session key", SessionCacheKey("sess1"), "session:"},
		{"feature flags key", FeatureFlagsCacheKey, "feature_flags:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.key) < len(tt.pattern) {
				t.Errorf("%s is shorter than expected pattern %q", tt.name, tt.pattern)
			}
			if tt.key[:len(tt.pattern)] != tt.pattern {
				t.Errorf("%s = %q, doesn't start with %q", tt.name, tt.key, tt.pattern)
			}
		})
	}
}

// ============ UserCacheKey Uniqueness Tests ============

func TestUserCacheKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	userIDs := []uint{1, 2, 10, 100, 1000, 10000}

	for _, id := range userIDs {
		key := UserCacheKey(id)
		if keys[key] {
			t.Errorf("UserCacheKey(%d) produced duplicate key %q", id, key)
		}
		keys[key] = true
	}
}

// ============ SessionCacheKey Uniqueness Tests ============

func TestSessionCacheKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	sessionIDs := []string{"a", "b", "abc", "def", "session1", "session2"}

	for _, id := range sessionIDs {
		key := SessionCacheKey(id)
		if keys[key] {
			t.Errorf("SessionCacheKey(%q) produced duplicate key %q", id, key)
		}
		keys[key] = true
	}
}
