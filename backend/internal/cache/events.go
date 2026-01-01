package cache

import (
	"context"

	"github.com/rs/zerolog/log"
)

// Event types for cache invalidation
const (
	EventFeatureFlagsUpdated = "feature_flags:updated"
	EventSettingsUpdated     = "settings:updated"
	EventAnnouncementUpdated = "announcement:updated"
	EventUserUpdated         = "user:updated"
)

// Event represents a cache invalidation event
type Event struct {
	Type    string   // Event type (e.g., EventFeatureFlagsUpdated)
	Keys    []string // Specific keys to invalidate
	Pattern string   // Pattern to invalidate (e.g., "feature_flags:*")
}

// PublishInvalidation publishes a cache invalidation event
// It clears specific keys and/or patterns from the cache
func PublishInvalidation(ctx context.Context, event Event) {
	log.Info().
		Str("event", event.Type).
		Strs("keys", event.Keys).
		Str("pattern", event.Pattern).
		Msg("cache invalidation event")

	// Invalidate specific keys
	for _, key := range event.Keys {
		if err := Delete(ctx, key); err != nil {
			log.Warn().
				Err(err).
				Str("key", key).
				Msg("failed to invalidate cache key")
		}
	}

	// Invalidate by pattern
	if event.Pattern != "" {
		if instance != nil {
			if err := instance.Clear(ctx, event.Pattern); err != nil {
				log.Warn().
					Err(err).
					Str("pattern", event.Pattern).
					Msg("failed to invalidate cache pattern")
			}
		}
	}
}

// InvalidateFeatureFlags invalidates all feature flag caches
func InvalidateFeatureFlags(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:    EventFeatureFlagsUpdated,
		Pattern: "feature_flags:*",
	})
}

// InvalidateSettings invalidates all settings caches
func InvalidateSettings(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:    EventSettingsUpdated,
		Pattern: "settings:*",
	})
}

// InvalidateAnnouncements invalidates all announcement caches
func InvalidateAnnouncements(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:    EventAnnouncementUpdated,
		Pattern: "announcements:*",
	})
}

// InvalidateUser invalidates a specific user's cache
func InvalidateUser(ctx context.Context, userID uint) {
	PublishInvalidation(ctx, Event{
		Type: EventUserUpdated,
		Keys: []string{UserCacheKey(userID)},
	})
}
