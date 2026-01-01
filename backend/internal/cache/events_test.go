package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishInvalidation_SpecificKeys(t *testing.T) {
	// Setup cache with test data
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	// Add some test data
	err = Set(ctx, "test:key1", []byte("value1"), time.Minute)
	require.NoError(t, err)
	err = Set(ctx, "test:key2", []byte("value2"), time.Minute)
	require.NoError(t, err)

	// Verify data exists
	_, err = Get(ctx, "test:key1")
	require.NoError(t, err)

	// Publish invalidation for specific key
	PublishInvalidation(ctx, Event{
		Type: "test:invalidation",
		Keys: []string{"test:key1"},
	})

	// Key1 should be gone
	_, err = Get(ctx, "test:key1")
	assert.Error(t, err)

	// Key2 should still exist
	_, err = Get(ctx, "test:key2")
	assert.NoError(t, err)
}

func TestPublishInvalidation_Pattern(t *testing.T) {
	// Setup cache with test data
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	// Add some test data with a common prefix
	err = Set(ctx, "pattern:a", []byte("value1"), time.Minute)
	require.NoError(t, err)
	err = Set(ctx, "pattern:b", []byte("value2"), time.Minute)
	require.NoError(t, err)
	err = Set(ctx, "other:c", []byte("value3"), time.Minute)
	require.NoError(t, err)

	// Publish invalidation for pattern
	PublishInvalidation(ctx, Event{
		Type:    "test:invalidation",
		Pattern: "pattern:*",
	})

	// Pattern keys should be gone
	_, err = Get(ctx, "pattern:a")
	assert.Error(t, err)
	_, err = Get(ctx, "pattern:b")
	assert.Error(t, err)

	// Other key should still exist
	_, err = Get(ctx, "other:c")
	assert.NoError(t, err)
}

func TestInvalidateFeatureFlags(t *testing.T) {
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	// Add feature flag data
	err = Set(ctx, "feature_flags:user:123", []byte("flags"), time.Minute)
	require.NoError(t, err)

	// Invalidate
	InvalidateFeatureFlags(ctx)

	// Should be gone
	_, err = Get(ctx, "feature_flags:user:123")
	assert.Error(t, err)
}

func TestInvalidateSettings(t *testing.T) {
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	// Add settings data
	err = Set(ctx, "settings:site", []byte("settings"), time.Minute)
	require.NoError(t, err)

	// Invalidate
	InvalidateSettings(ctx)

	// Should be gone
	_, err = Get(ctx, "settings:site")
	assert.Error(t, err)
}

func TestInvalidateAnnouncements(t *testing.T) {
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	// Add announcement data
	err = Set(ctx, "announcements:active", []byte("announcements"), time.Minute)
	require.NoError(t, err)

	// Invalidate
	InvalidateAnnouncements(ctx)

	// Should be gone
	_, err = Get(ctx, "announcements:active")
	assert.Error(t, err)
}

func TestInvalidateUser(t *testing.T) {
	config := &Config{Enabled: true, MemoryCleanupInterval: time.Minute}
	err := Initialize(config)
	require.NoError(t, err)
	defer Close()

	ctx := context.Background()

	userID := uint(42)
	userKey := UserCacheKey(userID)

	// Add user data
	err = Set(ctx, userKey, []byte("user data"), time.Minute)
	require.NoError(t, err)

	// Invalidate
	InvalidateUser(ctx, userID)

	// Should be gone
	_, err = Get(ctx, userKey)
	assert.Error(t, err)
}
