import { create } from "zustand";
import { devtools, persist, type PersistStorage } from "zustand/middleware";

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  timestamp: string; // ISO string for serialization
  read: boolean;
  data?: unknown;
}

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;

  // Actions
  addNotification: (notification: Omit<Notification, "id" | "timestamp" | "read">) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

// Max notifications to keep (prevents unbounded storage growth)
const MAX_NOTIFICATIONS = 50;

// Max age for notifications (7 days)
const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

// BroadcastChannel for cross-tab sync
const CHANNEL_NAME = "notification-sync";
let broadcastChannel: BroadcastChannel | null = null;

function getBroadcastChannel(): BroadcastChannel | null {
  if (typeof window === "undefined" || !("BroadcastChannel" in window)) {
    return null;
  }
  if (!broadcastChannel) {
    broadcastChannel = new BroadcastChannel(CHANNEL_NAME);
  }
  return broadcastChannel;
}

// Broadcast state changes to other tabs
function broadcastChange(notifications: Notification[], unreadCount: number) {
  const channel = getBroadcastChannel();
  if (channel) {
    try {
      channel.postMessage({ type: "sync", notifications, unreadCount });
    } catch {
      // Ignore errors (e.g., if channel is closed)
    }
  }
}

// Type for persisted state (subset of full state)
type PersistedNotificationState = Pick<NotificationState, "notifications" | "unreadCount">;

// Custom storage that filters old notifications on read
const customStorage: PersistStorage<PersistedNotificationState> = {
  getItem: (name) => {
    const str = localStorage.getItem(name);
    if (!str) return null;

    try {
      const parsed = JSON.parse(str);
      if (parsed.state?.notifications) {
        const now = Date.now();
        // Filter out old notifications on rehydration
        parsed.state.notifications = parsed.state.notifications.filter((n: Notification) => {
          const age = now - new Date(n.timestamp).getTime();
          return age < MAX_AGE_MS;
        });
        // Recalculate unread count
        parsed.state.unreadCount = parsed.state.notifications.filter((n: Notification) => !n.read).length;
      }
      return parsed;
    } catch {
      return null;
    }
  },
  setItem: (name, value) => {
    localStorage.setItem(name, JSON.stringify(value));
  },
  removeItem: (name) => {
    localStorage.removeItem(name);
  },
};

export const useNotificationStore = create<NotificationState>()(
  devtools(
    persist(
      (set, get) => ({
        notifications: [],
        unreadCount: 0,

        addNotification: (notification) => {
          set((state) => {
            const newNotification: Notification = {
              ...notification,
              id: crypto.randomUUID(),
              timestamp: new Date().toISOString(),
              read: false,
            };
            const newNotifications = [newNotification, ...state.notifications].slice(0, MAX_NOTIFICATIONS);
            const newUnreadCount = newNotifications.filter((n) => !n.read).length;

            // Broadcast to other tabs
            broadcastChange(newNotifications, newUnreadCount);

            return {
              notifications: newNotifications,
              unreadCount: newUnreadCount,
            };
          });
        },

        markAsRead: (id) => {
          set((state) => {
            const notification = state.notifications.find((n) => n.id === id);
            if (!notification || notification.read) return state;

            const newNotifications = state.notifications.map((n) => (n.id === id ? { ...n, read: true } : n));
            const newUnreadCount = Math.max(0, state.unreadCount - 1);

            // Broadcast to other tabs
            broadcastChange(newNotifications, newUnreadCount);

            return {
              notifications: newNotifications,
              unreadCount: newUnreadCount,
            };
          });
        },

        markAllAsRead: () => {
          set((state) => {
            const newNotifications = state.notifications.map((n) => ({ ...n, read: true }));

            // Broadcast to other tabs
            broadcastChange(newNotifications, 0);

            return {
              notifications: newNotifications,
              unreadCount: 0,
            };
          });
        },

        removeNotification: (id) => {
          set((state) => {
            const notification = state.notifications.find((n) => n.id === id);
            const wasUnread = notification && !notification.read;

            const newNotifications = state.notifications.filter((n) => n.id !== id);
            const newUnreadCount = wasUnread ? Math.max(0, state.unreadCount - 1) : state.unreadCount;

            // Broadcast to other tabs
            broadcastChange(newNotifications, newUnreadCount);

            return {
              notifications: newNotifications,
              unreadCount: newUnreadCount,
            };
          });
        },

        clearAll: () => {
          // Broadcast to other tabs
          broadcastChange([], 0);

          set({
            notifications: [],
            unreadCount: 0,
          });
        },
      }),
      {
        name: "notification-store",
        storage: customStorage,
        partialize: (state) => ({
          notifications: state.notifications,
          unreadCount: state.unreadCount,
        }),
      }
    ),
    { name: "notification-store" }
  )
);

// Set up cross-tab sync listener
if (typeof window !== "undefined") {
  const channel = getBroadcastChannel();
  if (channel) {
    channel.onmessage = (event) => {
      if (event.data?.type === "sync") {
        const { notifications, unreadCount } = event.data;
        // Update store without triggering another broadcast
        useNotificationStore.setState({ notifications, unreadCount }, false);
      }
    };
  }
}
