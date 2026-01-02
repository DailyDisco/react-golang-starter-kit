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
	Type      string   // Event type (e.g., EventFeatureFlagsUpdated)
	Keys      []string // Specific keys to invalidate
	Pattern   string   // Pattern to invalidate (e.g., "feature_flags:*")
	Broadcast bool     // Whether to broadcast via WebSocket to clients
	QueryKeys []string // Frontend TanStack Query keys to invalidate (for WebSocket broadcast)
}

// PublishInvalidation publishes a cache invalidation event.
// It clears specific keys and/or patterns from the server-side cache,
// and optionally broadcasts to connected clients via WebSocket.
func PublishInvalidation(ctx context.Context, event Event) {
	log.Info().
		Str("event", event.Type).
		Strs("keys", event.Keys).
		Str("pattern", event.Pattern).
		Bool("broadcast", event.Broadcast).
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

	// Broadcast via WebSocket if enabled
	if event.Broadcast && len(event.QueryKeys) > 0 {
		BroadcastInvalidation(ctx, event, event.QueryKeys)
	}
}

// InvalidateFeatureFlags invalidates all feature flag caches and broadcasts to clients.
// This is an admin action that should update all connected clients.
func InvalidateFeatureFlags(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:      EventFeatureFlagsUpdated,
		Pattern:   "feature_flags:*",
		Broadcast: true,
		QueryKeys: []string{QueryKeyFeatureFlags},
	})
}

// InvalidateSettings invalidates all settings caches and broadcasts to clients.
// This is an admin action that should update all connected clients.
func InvalidateSettings(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:      EventSettingsUpdated,
		Pattern:   "settings:*",
		Broadcast: true,
		QueryKeys: []string{QueryKeySettings},
	})
}

// InvalidateAnnouncements invalidates all announcement caches and broadcasts to clients.
// This is an admin action that should update all connected clients.
func InvalidateAnnouncements(ctx context.Context) {
	PublishInvalidation(ctx, Event{
		Type:      EventAnnouncementUpdated,
		Pattern:   "announcements:*",
		Broadcast: true,
		QueryKeys: []string{QueryKeyAnnouncements},
	})
}

// InvalidateUser invalidates a specific user's cache.
// This is user-specific and does NOT broadcast globally.
func InvalidateUser(ctx context.Context, userID uint) {
	PublishInvalidation(ctx, Event{
		Type:      EventUserUpdated,
		Keys:      []string{UserCacheKey(userID)},
		Broadcast: false, // User-specific, no global broadcast
	})
}
