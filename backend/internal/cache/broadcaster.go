package cache

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// CacheInvalidatePayload is sent to clients when server-side cache is invalidated.
// Clients should invalidate their corresponding TanStack Query cache entries.
type CacheInvalidatePayload struct {
	// QueryKeys are the TanStack Query keys to invalidate (e.g., ["featureFlags"], ["settings"])
	QueryKeys []string `json:"queryKeys"`

	// Event is the cache event type (e.g., "feature_flags:updated")
	Event string `json:"event,omitempty"`

	// Timestamp is the Unix timestamp when invalidation occurred
	Timestamp int64 `json:"timestamp"`
}

// CacheBroadcaster is an interface for broadcasting cache invalidation messages.
// This allows the cache package to broadcast without importing the websocket package,
// avoiding import cycles.
type CacheBroadcaster interface {
	BroadcastCacheInvalidation(payload CacheInvalidatePayload)
}

// broadcaster is the package-level instance
var broadcaster CacheBroadcaster

// InitBroadcaster initializes the cache broadcaster.
// This should be called during application startup after the WebSocket hub is created.
// Pass a CacheBroadcaster implementation (e.g., HubBroadcaster wrapper around websocket.Hub).
func InitBroadcaster(b CacheBroadcaster) {
	if b == nil {
		log.Warn().Msg("cache broadcaster initialized with nil - broadcasts will be disabled")
		return
	}
	broadcaster = b
	log.Info().Msg("cache broadcaster initialized")
}

// BroadcastInvalidation sends a cache invalidation message to all connected clients.
// This should be called when admin actions modify shared data (feature flags, settings, etc.).
func BroadcastInvalidation(ctx context.Context, event Event, queryKeys []string) {
	if broadcaster == nil {
		log.Debug().Msg("cache broadcaster not initialized, skipping WebSocket broadcast")
		return
	}

	if len(queryKeys) == 0 {
		log.Debug().Str("event", event.Type).Msg("no query keys to broadcast, skipping")
		return
	}

	payload := CacheInvalidatePayload{
		QueryKeys: queryKeys,
		Event:     event.Type,
		Timestamp: time.Now().Unix(),
	}

	broadcaster.BroadcastCacheInvalidation(payload)

	log.Debug().
		Str("event", event.Type).
		Strs("query_keys", queryKeys).
		Msg("cache invalidation broadcasted via WebSocket")
}

// Query key constants matching frontend TanStack Query keys.
// These should match the keys defined in frontend/app/lib/query-keys.ts
const (
	QueryKeyFeatureFlags  = "featureFlags"
	QueryKeySettings      = "settings"
	QueryKeyAnnouncements = "changelog"
	QueryKeyUsers         = "users"
	QueryKeyAdminStats    = "adminStats"
	QueryKeyHealth        = "health"
	QueryKeyBilling       = "billing"
)
