import { useQuery } from "@tanstack/react-query";

import { queryKeys } from "../../lib/query-keys";
import { NotificationService } from "../../services/notifications/notificationService";
import type { NotificationsParams } from "../../types/notification";

/**
 * Hook to fetch paginated notifications for the current user
 */
export function useNotifications(params?: NotificationsParams) {
  return useQuery({
    queryKey: queryKeys.notifications.list(params),
    queryFn: () => NotificationService.getNotifications(params),
    staleTime: 30 * 1000, // 30 seconds
  });
}

/**
 * Hook to fetch just the unread count (for badge display)
 */
export function useUnreadNotificationCount() {
  return useQuery({
    queryKey: queryKeys.notifications.unreadCount(),
    queryFn: async () => {
      const response = await NotificationService.getNotifications({ per_page: 1 });
      return response.unread;
    },
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refetch every minute
  });
}
