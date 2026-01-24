import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useNotificationStore } from "./notification-store";

// Mock crypto.randomUUID
vi.stubGlobal("crypto", {
  randomUUID: vi.fn(() => "mock-uuid-" + Math.random().toString(36).substr(2, 9)),
});

describe("useNotificationStore", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    act(() => {
      useNotificationStore.getState().clearAll();
    });
  });

  describe("initial state", () => {
    it("has empty notifications array initially", () => {
      const state = useNotificationStore.getState();
      expect(state.notifications).toEqual([]);
    });

    it("has unreadCount set to 0 initially", () => {
      const state = useNotificationStore.getState();
      expect(state.unreadCount).toBe(0);
    });
  });

  describe("addNotification", () => {
    it("adds a notification with auto-generated id and timestamp", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test Notification",
          message: "This is a test message",
          type: "info",
        });
      });

      const state = useNotificationStore.getState();
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].title).toBe("Test Notification");
      expect(state.notifications[0].message).toBe("This is a test message");
      expect(state.notifications[0].type).toBe("info");
      expect(state.notifications[0].read).toBe(false);
      expect(state.notifications[0].id).toBeDefined();
      expect(typeof state.notifications[0].timestamp).toBe("string");
      expect(new Date(state.notifications[0].timestamp).toISOString()).toBe(state.notifications[0].timestamp);
    });

    it("increments unreadCount when adding notification", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "success",
        });
      });

      expect(useNotificationStore.getState().unreadCount).toBe(1);

      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test 2",
          message: "Message 2",
          type: "error",
        });
      });

      expect(useNotificationStore.getState().unreadCount).toBe(2);
    });

    it("adds new notifications at the beginning of the array", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "First",
          message: "First message",
          type: "info",
        });
      });

      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Second",
          message: "Second message",
          type: "warning",
        });
      });

      const notifications = useNotificationStore.getState().notifications;
      expect(notifications[0].title).toBe("Second");
      expect(notifications[1].title).toBe("First");
    });

    it("limits notifications to 50 maximum", () => {
      // Add 55 notifications
      for (let i = 0; i < 55; i++) {
        act(() => {
          useNotificationStore.getState().addNotification({
            title: `Notification ${i}`,
            message: `Message ${i}`,
            type: "info",
          });
        });
      }

      expect(useNotificationStore.getState().notifications).toHaveLength(50);
    });

    it("supports different notification types", () => {
      const types: Array<"info" | "success" | "warning" | "error"> = ["info", "success", "warning", "error"];

      for (const type of types) {
        act(() => {
          useNotificationStore.getState().addNotification({
            title: `${type} notification`,
            message: `This is a ${type} message`,
            type,
          });
        });
      }

      const notifications = useNotificationStore.getState().notifications;
      expect(notifications.map((n) => n.type)).toEqual(["error", "warning", "success", "info"]);
    });

    it("supports optional data field", () => {
      const customData = { userId: 123, action: "created" };

      act(() => {
        useNotificationStore.getState().addNotification({
          title: "User Action",
          message: "User performed an action",
          type: "info",
          data: customData,
        });
      });

      const notification = useNotificationStore.getState().notifications[0];
      expect(notification.data).toEqual(customData);
    });
  });

  describe("markAsRead", () => {
    it("marks a notification as read", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const notification = useNotificationStore.getState().notifications[0];

      act(() => {
        useNotificationStore.getState().markAsRead(notification.id);
      });

      const updatedNotification = useNotificationStore.getState().notifications[0];
      expect(updatedNotification.read).toBe(true);
    });

    it("decrements unreadCount when marking as read", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      expect(useNotificationStore.getState().unreadCount).toBe(1);

      const notification = useNotificationStore.getState().notifications[0];

      act(() => {
        useNotificationStore.getState().markAsRead(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });

    it("does not decrement unreadCount if notification already read", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const notification = useNotificationStore.getState().notifications[0];

      act(() => {
        useNotificationStore.getState().markAsRead(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);

      act(() => {
        useNotificationStore.getState().markAsRead(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });

    it("does nothing for non-existent notification id", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const initialState = useNotificationStore.getState();

      act(() => {
        useNotificationStore.getState().markAsRead("non-existent-id");
      });

      expect(useNotificationStore.getState().unreadCount).toBe(initialState.unreadCount);
    });
  });

  describe("markAllAsRead", () => {
    it("marks all notifications as read", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test 1",
          message: "Message 1",
          type: "info",
        });
        useNotificationStore.getState().addNotification({
          title: "Test 2",
          message: "Message 2",
          type: "success",
        });
      });

      expect(useNotificationStore.getState().notifications.every((n) => !n.read)).toBe(true);

      act(() => {
        useNotificationStore.getState().markAllAsRead();
      });

      expect(useNotificationStore.getState().notifications.every((n) => n.read)).toBe(true);
    });

    it("sets unreadCount to 0", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test 1",
          message: "Message 1",
          type: "info",
        });
        useNotificationStore.getState().addNotification({
          title: "Test 2",
          message: "Message 2",
          type: "success",
        });
      });

      expect(useNotificationStore.getState().unreadCount).toBe(2);

      act(() => {
        useNotificationStore.getState().markAllAsRead();
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });
  });

  describe("removeNotification", () => {
    it("removes a notification by id", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const notification = useNotificationStore.getState().notifications[0];
      expect(useNotificationStore.getState().notifications).toHaveLength(1);

      act(() => {
        useNotificationStore.getState().removeNotification(notification.id);
      });

      expect(useNotificationStore.getState().notifications).toHaveLength(0);
    });

    it("decrements unreadCount if removed notification was unread", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const notification = useNotificationStore.getState().notifications[0];
      expect(useNotificationStore.getState().unreadCount).toBe(1);

      act(() => {
        useNotificationStore.getState().removeNotification(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });

    it("does not decrement unreadCount if removed notification was already read", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test",
          message: "Message",
          type: "info",
        });
      });

      const notification = useNotificationStore.getState().notifications[0];

      act(() => {
        useNotificationStore.getState().markAsRead(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);

      act(() => {
        useNotificationStore.getState().removeNotification(notification.id);
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });
  });

  describe("clearAll", () => {
    it("clears all notifications", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test 1",
          message: "Message 1",
          type: "info",
        });
        useNotificationStore.getState().addNotification({
          title: "Test 2",
          message: "Message 2",
          type: "success",
        });
      });

      expect(useNotificationStore.getState().notifications).toHaveLength(2);

      act(() => {
        useNotificationStore.getState().clearAll();
      });

      expect(useNotificationStore.getState().notifications).toHaveLength(0);
    });

    it("resets unreadCount to 0", () => {
      act(() => {
        useNotificationStore.getState().addNotification({
          title: "Test 1",
          message: "Message 1",
          type: "info",
        });
        useNotificationStore.getState().addNotification({
          title: "Test 2",
          message: "Message 2",
          type: "success",
        });
      });

      expect(useNotificationStore.getState().unreadCount).toBe(2);

      act(() => {
        useNotificationStore.getState().clearAll();
      });

      expect(useNotificationStore.getState().unreadCount).toBe(0);
    });
  });
});
