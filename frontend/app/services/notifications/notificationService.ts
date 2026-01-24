import type { NotificationsParams, NotificationsResponse } from "../../types/notification";
import { apiClient } from "../api/client";

export const NotificationService = {
  /**
   * Get paginated notifications for the current user
   */
  getNotifications: async (params?: NotificationsParams): Promise<NotificationsResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.per_page) searchParams.set("per_page", params.per_page.toString());
    if (params?.unread) searchParams.set("unread", "true");

    const query = searchParams.toString();
    return apiClient.get<NotificationsResponse>(`/notifications${query ? `?${query}` : ""}`);
  },

  /**
   * Mark a single notification as read
   */
  markAsRead: async (id: number): Promise<void> => {
    await apiClient.post(`/notifications/${id}/read`, {});
  },

  /**
   * Mark all notifications as read
   */
  markAllAsRead: async (): Promise<void> => {
    await apiClient.post("/notifications/read-all", {});
  },

  /**
   * Delete a notification
   */
  deleteNotification: async (id: number): Promise<void> => {
    await apiClient.delete(`/notifications/${id}`);
  },
};
