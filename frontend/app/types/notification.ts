export type NotificationType = "info" | "success" | "warning" | "error" | "announcement" | "billing" | "security";

export interface Notification {
  id: number;
  created_at: string;
  type: NotificationType;
  title: string;
  message: string;
  link?: string;
  read: boolean;
  read_at?: string;
}

export interface NotificationsResponse {
  notifications: Notification[];
  total: number;
  unread: number;
  page: number;
  per_page: number;
}

export interface NotificationsParams {
  page?: number;
  per_page?: number;
  unread?: boolean;
}
