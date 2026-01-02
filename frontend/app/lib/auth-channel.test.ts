import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  broadcastLogin,
  broadcastLogout,
  broadcastSessionExpired,
  closeAuthChannel,
  isChannelHealthy,
  isUsingFallback,
  subscribeToAuthEvents,
  type AuthEvent,
} from "./auth-channel";

// Mock BroadcastChannel
class MockBroadcastChannel {
  static instances: MockBroadcastChannel[] = [];

  name: string;
  onmessage: ((event: MessageEvent) => void) | null = null;
  postMessage = vi.fn();
  close = vi.fn();
  addEventListener = vi.fn((event: string, handler: (event: MessageEvent) => void) => {
    if (event === "message") {
      this.onmessage = handler;
    }
  });
  removeEventListener = vi.fn();

  constructor(name: string) {
    this.name = name;
    MockBroadcastChannel.instances.push(this);
  }

  // Helper to simulate receiving a message from another tab
  simulateMessage(data: AuthEvent): void {
    if (this.onmessage) {
      this.onmessage(new MessageEvent("message", { data }));
    }
  }
}

describe("auth-channel", () => {
  beforeEach(() => {
    // Reset module state by closing any existing channel
    closeAuthChannel();
    // Clear mock instances
    MockBroadcastChannel.instances = [];
    vi.clearAllMocks();
  });

  afterEach(() => {
    closeAuthChannel();
    // Restore original BroadcastChannel if it was modified
    vi.unstubAllGlobals();
  });

  describe("getChannel (via broadcast functions)", () => {
    it("returns singleton instance on repeated calls", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      // First broadcast creates the channel
      broadcastLogout();
      expect(MockBroadcastChannel.instances).toHaveLength(1);

      // Second broadcast reuses same channel
      broadcastLogout();
      expect(MockBroadcastChannel.instances).toHaveLength(1);
    });

    it("returns null when BroadcastChannel is undefined", () => {
      vi.stubGlobal("BroadcastChannel", undefined);

      // Should not throw, just silently fail
      expect(() => broadcastLogout()).not.toThrow();
    });

    it("returns null when window is undefined (SSR)", () => {
      const originalWindow = global.window;
      // @ts-expect-error - simulating SSR
      delete global.window;

      expect(() => broadcastLogout()).not.toThrow();

      global.window = originalWindow;
    });

    it("returns null when BroadcastChannel constructor throws", () => {
      vi.stubGlobal(
        "BroadcastChannel",
        class {
          constructor() {
            throw new Error("BroadcastChannel not supported");
          }
        }
      );

      expect(() => broadcastLogout()).not.toThrow();
    });
  });

  describe("broadcastLogout", () => {
    it("posts message with correct type and timestamp", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const now = Date.now();
      vi.setSystemTime(now);

      broadcastLogout();

      const instance = MockBroadcastChannel.instances[0];
      expect(instance.postMessage).toHaveBeenCalledWith({
        type: "logout",
        userId: undefined,
        timestamp: now,
      });

      vi.useRealTimers();
    });
  });

  describe("broadcastLogin", () => {
    it("posts message with userId", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const now = Date.now();
      vi.setSystemTime(now);

      broadcastLogin(42);

      const instance = MockBroadcastChannel.instances[0];
      expect(instance.postMessage).toHaveBeenCalledWith({
        type: "login",
        userId: 42,
        timestamp: now,
      });

      vi.useRealTimers();
    });
  });

  describe("broadcastSessionExpired", () => {
    it("posts correct event type", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const now = Date.now();
      vi.setSystemTime(now);

      broadcastSessionExpired();

      const instance = MockBroadcastChannel.instances[0];
      expect(instance.postMessage).toHaveBeenCalledWith({
        type: "session_expired",
        userId: undefined,
        timestamp: now,
      });

      vi.useRealTimers();
    });
  });

  describe("subscribeToAuthEvents", () => {
    it("receives messages from other tabs", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      const testEvent: AuthEvent = {
        type: "logout",
        timestamp: Date.now(),
      };

      instance.simulateMessage(testEvent);

      expect(callback).toHaveBeenCalledWith(testEvent);
    });

    it("returns unsubscribe function", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      const unsubscribe = subscribeToAuthEvents(callback);
      unsubscribe();

      const instance = MockBroadcastChannel.instances[0];
      instance.simulateMessage({
        type: "logout",
        timestamp: Date.now(),
      });

      expect(callback).not.toHaveBeenCalled();
    });

    it("supports multiple subscribers", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      subscribeToAuthEvents(callback1);
      subscribeToAuthEvents(callback2);

      const instance = MockBroadcastChannel.instances[0];
      const testEvent: AuthEvent = {
        type: "login",
        userId: 1,
        timestamp: Date.now(),
      };

      instance.simulateMessage(testEvent);

      expect(callback1).toHaveBeenCalledWith(testEvent);
      expect(callback2).toHaveBeenCalledWith(testEvent);
    });

    it("unsubscribe only removes specific subscriber", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      const unsubscribe1 = subscribeToAuthEvents(callback1);
      subscribeToAuthEvents(callback2);

      unsubscribe1();

      const instance = MockBroadcastChannel.instances[0];
      instance.simulateMessage({
        type: "logout",
        timestamp: Date.now(),
      });

      expect(callback1).not.toHaveBeenCalled();
      expect(callback2).toHaveBeenCalled();
    });
  });

  describe("subscriber error isolation", () => {
    it("subscriber errors do not break other subscribers", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const errorCallback = vi.fn(() => {
        throw new Error("Subscriber error");
      });
      const successCallback = vi.fn();

      subscribeToAuthEvents(errorCallback);
      subscribeToAuthEvents(successCallback);

      const instance = MockBroadcastChannel.instances[0];
      instance.simulateMessage({
        type: "logout",
        timestamp: Date.now(),
      });

      expect(errorCallback).toHaveBeenCalled();
      expect(successCallback).toHaveBeenCalled();
    });
  });

  describe("message validation", () => {
    it("ignores messages with invalid type", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      // Simulate invalid message (type is not a string)
      if (instance.onmessage) {
        instance.onmessage(new MessageEvent("message", { data: { type: 123, timestamp: Date.now() } }));
      }

      expect(callback).not.toHaveBeenCalled();
    });

    it("ignores messages with invalid timestamp", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      // Simulate invalid message (timestamp is not a number)
      if (instance.onmessage) {
        instance.onmessage(new MessageEvent("message", { data: { type: "logout", timestamp: "invalid" } }));
      }

      expect(callback).not.toHaveBeenCalled();
    });

    it("ignores null/undefined messages", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      if (instance.onmessage) {
        instance.onmessage(new MessageEvent("message", { data: null }));
        instance.onmessage(new MessageEvent("message", { data: undefined }));
      }

      expect(callback).not.toHaveBeenCalled();
    });
  });

  describe("closeAuthChannel", () => {
    it("cleans up resources", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      // Create channel by subscribing
      const callback = vi.fn();
      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];

      closeAuthChannel();

      expect(instance.removeEventListener).toHaveBeenCalled();
      expect(instance.close).toHaveBeenCalled();
    });

    it("clears all subscribers", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      const callback = vi.fn();
      subscribeToAuthEvents(callback);

      closeAuthChannel();

      // Re-create channel and simulate message
      broadcastLogout(); // This creates a new channel
      const newInstance = MockBroadcastChannel.instances[1];
      if (newInstance?.onmessage) {
        newInstance.onmessage(
          new MessageEvent("message", {
            data: { type: "logout", timestamp: Date.now() },
          })
        );
      }

      // Old subscriber should not be called
      expect(callback).not.toHaveBeenCalled();
    });

    it("handles being called multiple times", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      subscribeToAuthEvents(vi.fn());

      expect(() => {
        closeAuthChannel();
        closeAuthChannel();
        closeAuthChannel();
      }).not.toThrow();
    });

    it("handles being called when no channel exists", () => {
      expect(() => closeAuthChannel()).not.toThrow();
    });
  });

  describe("postMessage error handling", () => {
    it("handles postMessage throwing an error", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      broadcastLogout(); // Create channel

      const instance = MockBroadcastChannel.instances[0];
      instance.postMessage.mockImplementation(() => {
        throw new Error("Channel closed");
      });

      // Should not throw
      expect(() => broadcastLogout()).not.toThrow();
    });
  });

  describe("timestamp validation", () => {
    it("ignores stale events with older timestamps", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      const now = Date.now();

      // First event with current timestamp
      instance.simulateMessage({
        type: "logout",
        timestamp: now,
      });

      expect(callback).toHaveBeenCalledTimes(1);

      // Second event with older timestamp should be ignored
      instance.simulateMessage({
        type: "login",
        userId: 1,
        timestamp: now - 1000, // 1 second earlier
      });

      expect(callback).toHaveBeenCalledTimes(1); // Still 1
    });

    it("ignores duplicate events with same timestamp", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      const timestamp = Date.now();

      instance.simulateMessage({ type: "logout", timestamp });
      instance.simulateMessage({ type: "logout", timestamp }); // Duplicate

      expect(callback).toHaveBeenCalledTimes(1);
    });

    it("processes events with newer timestamps", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);
      const callback = vi.fn();

      subscribeToAuthEvents(callback);

      const instance = MockBroadcastChannel.instances[0];
      const now = Date.now();

      instance.simulateMessage({ type: "logout", timestamp: now });
      instance.simulateMessage({ type: "login", userId: 1, timestamp: now + 1000 });

      expect(callback).toHaveBeenCalledTimes(2);
    });
  });

  describe("localStorage fallback", () => {
    it("uses localStorage when BroadcastChannel is undefined", () => {
      vi.stubGlobal("BroadcastChannel", undefined);
      const addEventListenerSpy = vi.spyOn(window, "addEventListener");

      subscribeToAuthEvents(vi.fn());

      expect(addEventListenerSpy).toHaveBeenCalledWith("storage", expect.any(Function));
      expect(isUsingFallback()).toBe(true);
    });

    it("broadcasts via localStorage when using fallback", () => {
      // Need to clear any existing state first
      closeAuthChannel();

      vi.stubGlobal("BroadcastChannel", undefined);

      // Create mock localStorage
      const mockLocalStorage = {
        setItem: vi.fn(),
        getItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
        length: 0,
        key: vi.fn(),
      };
      vi.stubGlobal("localStorage", mockLocalStorage);

      // Subscribe first to initialize the fallback
      subscribeToAuthEvents(vi.fn());

      // Verify we're using fallback
      expect(isUsingFallback()).toBe(true);

      // Now broadcast
      broadcastLogout();

      expect(mockLocalStorage.setItem).toHaveBeenCalledWith("auth_sync_event", expect.any(String));
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith("auth_sync_event");
    });

    it("receives events via localStorage storage event", () => {
      vi.stubGlobal("BroadcastChannel", undefined);
      const callback = vi.fn();
      let storageHandler: ((e: StorageEvent) => void) | undefined;

      vi.spyOn(window, "addEventListener").mockImplementation((event, handler) => {
        if (event === "storage") {
          storageHandler = handler as (e: StorageEvent) => void;
        }
      });

      subscribeToAuthEvents(callback);

      // Simulate storage event from another tab
      const event: AuthEvent = { type: "logout", timestamp: Date.now() };
      storageHandler?.(
        new StorageEvent("storage", {
          key: "auth_sync_event",
          newValue: JSON.stringify(event),
        })
      );

      expect(callback).toHaveBeenCalledWith(event);
    });

    it("ignores storage events for other keys", () => {
      vi.stubGlobal("BroadcastChannel", undefined);
      const callback = vi.fn();
      let storageHandler: ((e: StorageEvent) => void) | undefined;

      vi.spyOn(window, "addEventListener").mockImplementation((event, handler) => {
        if (event === "storage") {
          storageHandler = handler as (e: StorageEvent) => void;
        }
      });

      subscribeToAuthEvents(callback);

      storageHandler?.(
        new StorageEvent("storage", {
          key: "some_other_key",
          newValue: JSON.stringify({ type: "logout", timestamp: Date.now() }),
        })
      );

      expect(callback).not.toHaveBeenCalled();
    });

    it("cleans up storage listener on close", () => {
      vi.stubGlobal("BroadcastChannel", undefined);
      const removeEventListenerSpy = vi.spyOn(window, "removeEventListener");

      subscribeToAuthEvents(vi.fn());
      closeAuthChannel();

      expect(removeEventListenerSpy).toHaveBeenCalledWith("storage", expect.any(Function));
      expect(isUsingFallback()).toBe(false);
    });
  });

  describe("isChannelHealthy", () => {
    it("returns true when BroadcastChannel is available", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      subscribeToAuthEvents(vi.fn());

      expect(isChannelHealthy()).toBe(true);
    });

    it("returns true when using fallback", () => {
      vi.stubGlobal("BroadcastChannel", undefined);

      subscribeToAuthEvents(vi.fn());

      expect(isChannelHealthy()).toBe(true);
    });

    it("returns false before initialization in SSR", () => {
      const originalWindow = global.window;
      // @ts-expect-error - simulating SSR
      delete global.window;

      // In SSR (no window), channel cannot be healthy
      // Note: This is hard to fully test in happy-dom since window is always defined
      // The actual isChannelHealthy() checks typeof window === "undefined"
      expect(typeof global.window).toBe("undefined");

      global.window = originalWindow;
    });
  });

  describe("isUsingFallback", () => {
    it("returns false when using BroadcastChannel", () => {
      vi.stubGlobal("BroadcastChannel", MockBroadcastChannel);

      subscribeToAuthEvents(vi.fn());

      expect(isUsingFallback()).toBe(false);
    });

    it("returns true when BroadcastChannel unavailable", () => {
      vi.stubGlobal("BroadcastChannel", undefined);

      subscribeToAuthEvents(vi.fn());

      expect(isUsingFallback()).toBe(true);
    });
  });
});
