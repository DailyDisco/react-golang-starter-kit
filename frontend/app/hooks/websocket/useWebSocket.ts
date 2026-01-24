import { useCallback } from "react";

import { handleCacheInvalidation, type CacheInvalidatePayload } from "../../lib/cache-invalidation";
import { logger } from "../../lib/logger";
import type {
  MemberUpdatePayload,
  NotificationPayload,
  OrgUpdatePayload,
  SubscriptionUpdatePayload,
  UsageAlertPayload,
  UserUpdatePayload,
  UseWebSocketOptions,
  UseWebSocketReturn,
  WebSocketMessage,
} from "./types";
import { useWebSocketConnection } from "./useWebSocketConnection";
import { useWebSocketInvalidation } from "./useWebSocketInvalidation";
import { dispatchOrgEvent, useWebSocketMessages } from "./useWebSocketMessages";

/**
 * Main WebSocket hook that composes connection, message handling, and cache invalidation.
 * This is the primary hook to use in your application.
 *
 * @example
 * ```tsx
 * function App() {
 *   const { isConnected, sendMessage } = useWebSocket({
 *     onMessage: (msg) => console.log('Custom handler:', msg),
 *   });
 *
 *   return <div>Connected: {isConnected ? 'Yes' : 'No'}</div>;
 * }
 * ```
 */
export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const { reconnectInterval, maxRetries, autoConnect, onMessage } = options;

  // Message handling (notifications, toasts)
  const { handleNotification, handleBroadcast, handleUsageAlert, handleSubscriptionUpdate } = useWebSocketMessages();

  // Cache invalidation
  const {
    invalidateUserUpdate,
    invalidateCacheMessage,
    invalidateUsageQueries,
    invalidateSubscriptionQueries,
    invalidateOrgQueries,
    invalidateMemberQueries,
    invalidateFeatureFlagQueries,
    invalidateNotificationQueries,
  } = useWebSocketInvalidation();

  /**
   * Unified message handler that routes to appropriate sub-handlers
   */
  const handleMessage = useCallback(
    (message: WebSocketMessage) => {
      switch (message.type) {
        case "notification":
          handleNotification(message.payload as NotificationPayload);
          break;

        case "user_update":
          invalidateUserUpdate(message.payload as UserUpdatePayload);
          break;

        case "broadcast":
          handleBroadcast(message.payload as NotificationPayload);
          break;

        case "pong":
          // Server responded to our ping - no action needed
          break;

        case "cache_invalidate":
          invalidateCacheMessage(message.payload as CacheInvalidatePayload);
          break;

        case "usage_alert": {
          const payload = message.payload as UsageAlertPayload;
          handleUsageAlert(payload);
          invalidateUsageQueries();
          break;
        }

        case "subscription_update": {
          const payload = message.payload as SubscriptionUpdatePayload;
          handleSubscriptionUpdate(payload);
          invalidateSubscriptionQueries();
          break;
        }

        case "org_update": {
          const payload = message.payload as OrgUpdatePayload;
          invalidateOrgQueries(payload);
          dispatchOrgEvent("org-update", payload);
          break;
        }

        case "member_update": {
          const payload = message.payload as MemberUpdatePayload;
          invalidateMemberQueries(payload);
          dispatchOrgEvent("member-update", payload);
          break;
        }

        case "feature_flag_update":
          invalidateFeatureFlagQueries();
          break;

        case "notification_new":
          invalidateNotificationQueries();
          break;

        default:
          logger.debug("Unknown WebSocket message type", { type: message.type });
      }

      // Call custom handler if provided
      onMessage?.(message);
    },
    [
      handleNotification,
      handleBroadcast,
      handleUsageAlert,
      handleSubscriptionUpdate,
      invalidateUserUpdate,
      invalidateCacheMessage,
      invalidateUsageQueries,
      invalidateSubscriptionQueries,
      invalidateOrgQueries,
      invalidateMemberQueries,
      invalidateFeatureFlagQueries,
      invalidateNotificationQueries,
      onMessage,
    ]
  );

  // Connection management
  const { isConnected, connect, disconnect, sendMessage } = useWebSocketConnection({
    reconnectInterval,
    maxRetries,
    autoConnect,
    onMessage: handleMessage,
  });

  return {
    isConnected,
    connect,
    disconnect,
    sendMessage,
  };
}
