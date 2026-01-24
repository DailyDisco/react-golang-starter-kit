import { useCallback } from "react";

import { logger } from "../../lib/logger";
import { useNotificationStore } from "../../stores/notification-store";
import type { NotificationPayload, SubscriptionUpdatePayload, UsageAlertPayload, WebSocketMessage } from "./types";

interface UseWebSocketMessagesReturn {
  /** Handle notification messages */
  handleNotification: (payload: NotificationPayload) => void;
  /** Handle broadcast messages */
  handleBroadcast: (payload: NotificationPayload) => void;
  /** Handle usage alert messages */
  handleUsageAlert: (payload: UsageAlertPayload) => void;
  /** Handle subscription update messages */
  handleSubscriptionUpdate: (payload: SubscriptionUpdatePayload) => void;
}

/**
 * Hook for handling WebSocket message notifications and alerts.
 * Manages toast notifications and custom event dispatching.
 */
export function useWebSocketMessages(): UseWebSocketMessagesReturn {
  const addNotification = useNotificationStore((state) => state.addNotification);

  const handleNotification = useCallback(
    (payload: NotificationPayload) => {
      addNotification({
        title: payload.title,
        message: payload.message,
        type: payload.type,
        data: payload.data,
      });
    },
    [addNotification]
  );

  const handleBroadcast = useCallback(
    (payload: NotificationPayload) => {
      if (payload) {
        addNotification({
          title: payload.title || "System",
          message: payload.message,
          type: payload.type || "info",
          data: payload.data,
        });
      }
    },
    [addNotification]
  );

  const handleUsageAlert = useCallback(
    (payload: UsageAlertPayload) => {
      const notificationType = payload.alertType === "exceeded" ? "error" : "warning";
      const title = payload.alertType === "exceeded" ? "Usage Limit Exceeded" : "Usage Warning";

      addNotification({
        title,
        message: payload.message,
        type: notificationType,
        data: {
          usageType: payload.usageType,
          percentageUsed: payload.percentageUsed,
          canUpgrade: payload.canUpgrade,
          suggestedPlan: payload.suggestedPlan,
          upgradeUrl: payload.upgradeUrl,
        },
      });

      // For exceeded alerts with upgrade option, dispatch event for upgrade prompt
      if (payload.alertType === "exceeded" && payload.canUpgrade) {
        window.dispatchEvent(new CustomEvent("usage-limit-exceeded", { detail: payload }));
      }
    },
    [addNotification]
  );

  const handleSubscriptionUpdate = useCallback(
    (payload: SubscriptionUpdatePayload) => {
      let notificationType: "info" | "success" | "warning" | "error" = "info";
      let title = "Subscription Update";

      switch (payload.event) {
        case "created":
          notificationType = "success";
          title = "Subscription Activated";
          break;
        case "updated":
          notificationType = payload.cancelAtPeriodEnd ? "warning" : "info";
          title = payload.cancelAtPeriodEnd ? "Subscription Canceling" : "Subscription Updated";
          break;
        case "deleted":
          notificationType = "warning";
          title = "Subscription Ended";
          break;
        case "payment_failed":
          notificationType = "error";
          title = "Payment Failed";
          break;
      }

      addNotification({
        title,
        message: payload.message,
        type: notificationType,
        data: {
          event: payload.event,
          status: payload.status,
          plan: payload.plan,
        },
      });

      // Dispatch event for components that need to react to subscription changes
      window.dispatchEvent(new CustomEvent("subscription-update", { detail: payload }));
    },
    [addNotification]
  );

  return {
    handleNotification,
    handleBroadcast,
    handleUsageAlert,
    handleSubscriptionUpdate,
  };
}

/**
 * Dispatch a custom window event for org/member updates
 */
export function dispatchOrgEvent(eventName: string, payload: unknown): void {
  logger.debug(`${eventName} received`, { payload });
  window.dispatchEvent(new CustomEvent(eventName, { detail: payload }));
}
