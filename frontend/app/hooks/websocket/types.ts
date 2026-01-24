// WebSocket message types (must match backend)
export type MessageType =
  | "notification"
  | "notification_new"
  | "user_update"
  | "broadcast"
  | "ping"
  | "pong"
  | "cache_invalidate"
  | "usage_alert"
  | "subscription_update"
  | "org_update"
  | "member_update"
  | "feature_flag_update";

export interface WebSocketMessage {
  type: MessageType;
  payload?: unknown;
}

export interface NotificationPayload {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  timestamp: string;
  data?: unknown;
}

export interface UserUpdatePayload {
  field: string;
  value?: unknown;
}

export interface UsageAlertPayload {
  alertType: string;
  usageType: string;
  currentUsage: number;
  limit: number;
  percentageUsed: number;
  message: string;
  canUpgrade: boolean;
  currentPlan?: string;
  suggestedPlan?: string;
  upgradeUrl?: string;
}

export interface SubscriptionUpdatePayload {
  event: "created" | "updated" | "deleted" | "payment_failed";
  status: string;
  plan?: string;
  priceId?: string;
  cancelAtPeriodEnd: boolean;
  currentPeriodEnd?: string;
  message: string;
  timestamp: number;
}

export interface OrgUpdatePayload {
  orgSlug: string;
  event: "settings_changed" | "billing_changed" | "deleted";
  field?: string;
}

export interface MemberUpdatePayload {
  orgSlug: string;
  event: "added" | "removed" | "role_changed" | "invitation_sent" | "invitation_revoked";
  userId?: number;
  role?: string;
}

export interface FeatureFlagUpdatePayload {
  event: "enabled" | "disabled" | "updated";
  flagKey?: string;
}

export interface UseWebSocketOptions {
  /** Custom message handler */
  onMessage?: (message: WebSocketMessage) => void;
  /** Reconnection interval in ms (default: 3000) */
  reconnectInterval?: number;
  /** Maximum reconnection attempts (default: 5) */
  maxRetries?: number;
  /** Enable auto-connect when authenticated (default: true) */
  autoConnect?: boolean;
}

export interface UseWebSocketReturn {
  /** Whether the WebSocket is connected */
  isConnected: boolean;
  /** Manually connect to WebSocket */
  connect: () => void;
  /** Manually disconnect from WebSocket */
  disconnect: () => void;
  /** Send a message through WebSocket */
  sendMessage: (type: MessageType, payload?: unknown) => void;
}

export interface ConnectionState {
  isConnected: boolean;
  retryCount: number;
}
