import type { QueryClient } from "@tanstack/react-query";

import { logger } from "./logger";
import { queryKeys } from "./query-keys";

/**
 * Payload received from backend when cache is invalidated via WebSocket.
 * This matches the CacheInvalidatePayload struct in backend/internal/cache/broadcaster.go
 */
export interface CacheInvalidatePayload {
  /** TanStack Query keys to invalidate (e.g., ["featureFlags"], ["settings"]) */
  queryKeys: string[];

  /** Cache event type (e.g., "feature_flags:updated") */
  event?: string;

  /** Unix timestamp when invalidation occurred */
  timestamp: number;
}

/**
 * Map of backend event types to frontend query key hierarchies.
 * When a backend event is received, we invalidate all queries under the corresponding key.
 */
const EVENT_TYPE_TO_QUERY_KEY: Record<string, readonly unknown[]> = {
  "feature_flags:updated": queryKeys.featureFlags.all,
  "settings:updated": queryKeys.settings.all,
  "announcement:updated": queryKeys.changelog.all,
  "user:updated": queryKeys.auth.user,
};

/**
 * Map of query key strings to frontend query key hierarchies.
 * These match the QueryKey* constants in backend/internal/cache/broadcaster.go
 */
const QUERY_KEY_STRING_TO_KEY: Record<string, readonly unknown[]> = {
  featureFlags: queryKeys.featureFlags.all,
  settings: queryKeys.settings.all,
  changelog: queryKeys.changelog.all,
  users: queryKeys.users.all,
  adminStats: ["admin", "stats"],
  health: queryKeys.health.status,
  billing: queryKeys.billing.all,
};

/**
 * Handle a cache invalidation message from the WebSocket.
 * Invalidates the appropriate TanStack Query cache entries.
 *
 * @param queryClient - The TanStack Query client
 * @param payload - The cache invalidation payload from the server
 */
export function handleCacheInvalidation(queryClient: QueryClient, payload: CacheInvalidatePayload): void {
  const { queryKeys: keys, event } = payload;

  // Process explicit query keys first
  if (keys && keys.length > 0) {
    for (const key of keys) {
      const targetQueryKey = QUERY_KEY_STRING_TO_KEY[key];
      if (targetQueryKey) {
        queryClient.invalidateQueries({ queryKey: targetQueryKey });
        logger.debug("Cache invalidated via WebSocket", { key, event });
      } else {
        // If we don't have a mapping, try to invalidate using the key as-is
        queryClient.invalidateQueries({ queryKey: [key] });
        logger.debug("Cache invalidated via WebSocket (raw key)", { key, event });
      }
    }
    return;
  }

  // Fall back to event type mapping if no explicit keys
  if (event && EVENT_TYPE_TO_QUERY_KEY[event]) {
    const targetQueryKey = EVENT_TYPE_TO_QUERY_KEY[event];
    queryClient.invalidateQueries({ queryKey: targetQueryKey });
    logger.debug("Cache invalidated via WebSocket (event)", { event });
    return;
  }

  logger.warn("Cache invalidation received without recognized queryKeys or event", {
    queryKeys: payload.queryKeys,
    event: payload.event,
    timestamp: payload.timestamp,
  });
}

/**
 * Invalidate cache by event type.
 * Useful for triggering invalidation from frontend code.
 *
 * @param queryClient - The TanStack Query client
 * @param eventType - The event type (e.g., "feature_flags:updated")
 */
export function invalidateByEventType(queryClient: QueryClient, eventType: string): void {
  const queryKey = EVENT_TYPE_TO_QUERY_KEY[eventType];
  if (queryKey) {
    queryClient.invalidateQueries({ queryKey });
  }
}
