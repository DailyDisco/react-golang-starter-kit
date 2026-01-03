// Import after mocks are set up
import { useQueryClient } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { handleCacheInvalidation } from "../lib/cache-invalidation";
import { AuthService } from "../services/auth/authService";
import { useAuthStore } from "../stores/auth-store";
import { useNotificationStore } from "../stores/notification-store";
import { useWebSocket } from "./useWebSocket";

// Mock dependencies before importing the hook
vi.mock("@tanstack/react-query", () => ({
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));

vi.mock("../stores/auth-store", () => ({
  useAuthStore: vi.fn((selector) => {
    const state = { isAuthenticated: true };
    return selector(state);
  }),
}));

vi.mock("../stores/notification-store", () => ({
  useNotificationStore: vi.fn((selector) => {
    const state = { addNotification: vi.fn() };
    return selector(state);
  }),
}));

vi.mock("../services/auth/authService", () => ({
  AuthService: {
    validateSession: vi.fn(() => Promise.resolve(true)),
  },
}));

vi.mock("../lib/logger", () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("../lib/cache-invalidation", () => ({
  handleCacheInvalidation: vi.fn(),
}));

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;
  static instances: MockWebSocket[] = [];

  url: string;
  readyState: number = MockWebSocket.CONNECTING;
  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  send = vi.fn();
  close = vi.fn();

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  // Helper to simulate connection opening
  simulateOpen(): void {
    this.readyState = MockWebSocket.OPEN;
    this.onopen?.(new Event("open"));
  }

  // Helper to simulate connection closing
  simulateClose(code = 1000): void {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.(new CloseEvent("close", { code }));
  }

  // Helper to simulate receiving a message
  simulateMessage(data: unknown): void {
    this.onmessage?.(new MessageEvent("message", { data: JSON.stringify(data) }));
  }

  // Helper to simulate an error
  simulateError(): void {
    this.onerror?.(new Event("error"));
  }
}

describe("useWebSocket", () => {
  let mockInvalidateQueries: ReturnType<typeof vi.fn>;
  let mockAddNotification: ReturnType<typeof vi.fn>;
  let mockQueryClient: ReturnType<typeof useQueryClient>;

  beforeEach(() => {
    vi.useFakeTimers();
    MockWebSocket.instances = [];
    vi.stubGlobal("WebSocket", MockWebSocket);

    // Set up mock return values
    mockInvalidateQueries = vi.fn();
    mockQueryClient = {
      invalidateQueries: mockInvalidateQueries,
    } as unknown as ReturnType<typeof useQueryClient>;
    vi.mocked(useQueryClient).mockReturnValue(mockQueryClient);

    mockAddNotification = vi.fn();
    vi.mocked(useNotificationStore).mockImplementation((selector) => {
      const state = { addNotification: mockAddNotification } as unknown;
      return selector(state as Parameters<typeof selector>[0]);
    });

    vi.mocked(useAuthStore).mockImplementation((selector) => {
      const state = { isAuthenticated: true } as unknown;
      return selector(state as Parameters<typeof selector>[0]);
    });

    vi.mocked(AuthService.validateSession).mockResolvedValue(true);

    // Mock window.location
    vi.stubGlobal("location", {
      protocol: "https:",
      host: "localhost:3000",
    });

    // Mock import.meta.env
    vi.stubGlobal("import.meta", { env: {} });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  describe("connection", () => {
    it("connects when authenticated", async () => {
      const { result } = renderHook(() => useWebSocket());

      // Wait for effect to run
      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(MockWebSocket.instances).toHaveLength(1);
      expect(MockWebSocket.instances[0].url).toBe("wss://localhost:3000/ws");
    });

    it("does not connect when not authenticated", async () => {
      vi.mocked(useAuthStore).mockImplementation((selector) => {
        const state = { isAuthenticated: false } as unknown;
        return selector(state as Parameters<typeof selector>[0]);
      });

      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(MockWebSocket.instances).toHaveLength(0);
    });

    it("disconnects when authentication is lost", async () => {
      const { rerender } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      const ws = MockWebSocket.instances[0];
      ws.simulateOpen();

      // Change auth state to false
      vi.mocked(useAuthStore).mockImplementation((selector) => {
        const state = { isAuthenticated: false } as unknown;
        return selector(state as Parameters<typeof selector>[0]);
      });

      rerender();

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(ws.close).toHaveBeenCalled();
    });

    it("sets isConnected to true when connection opens", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(result.current.isConnected).toBe(false);

      await act(async () => {
        MockWebSocket.instances[0].simulateOpen();
      });

      expect(result.current.isConnected).toBe(true);
    });

    it("sets isConnected to false when connection closes", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      expect(result.current.isConnected).toBe(true);

      await act(async () => {
        MockWebSocket.instances[0].simulateClose();
      });

      expect(result.current.isConnected).toBe(false);
    });
  });

  describe("message handling", () => {
    it("handles user_update message - invalidates auth queries", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      await act(async () => {
        MockWebSocket.instances[0].simulateMessage({
          type: "user_update",
          payload: { field: "profile" },
        });
      });

      expect(mockInvalidateQueries).toHaveBeenCalledWith({
        queryKey: ["auth", "user"],
      });
    });

    it("handles user_update for preferences", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "user_update",
          payload: { field: "preferences" },
        });
      });

      expect(mockInvalidateQueries).toHaveBeenCalledWith({
        queryKey: ["preferences"],
      });
    });

    it("handles usage_alert message - shows notification", async () => {
      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "usage_alert",
          payload: {
            alertType: "warning",
            usageType: "api_calls",
            currentUsage: 800,
            limit: 1000,
            percentageUsed: 80,
            message: "You've used 80% of your API calls",
          },
        });
      });

      expect(mockAddNotification).toHaveBeenCalledWith({
        title: "Usage Warning",
        message: "You've used 80% of your API calls",
        type: "warning",
        data: { usageType: "api_calls", percentageUsed: 80 },
      });

      expect(mockInvalidateQueries).toHaveBeenCalledWith({
        queryKey: ["usage"],
      });
    });

    it("handles usage_alert exceeded - shows error notification", async () => {
      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "usage_alert",
          payload: {
            alertType: "exceeded",
            usageType: "storage",
            currentUsage: 1100,
            limit: 1000,
            percentageUsed: 110,
            message: "Storage limit exceeded",
          },
        });
      });

      expect(mockAddNotification).toHaveBeenCalledWith({
        title: "Usage Limit Exceeded",
        message: "Storage limit exceeded",
        type: "error",
        data: { usageType: "storage", percentageUsed: 110 },
      });
    });

    it("handles cache_invalidate message", async () => {
      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "cache_invalidate",
          payload: { queryKey: ["feature-flags"] },
        });
      });

      expect(handleCacheInvalidation).toHaveBeenCalledWith(mockQueryClient, {
        queryKey: ["feature-flags"],
      });
    });

    it("handles notification message", async () => {
      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "notification",
          payload: {
            id: "notif-1",
            title: "Welcome",
            message: "Welcome to the app",
            type: "success",
            timestamp: "2024-01-01T00:00:00Z",
          },
        });
      });

      expect(mockAddNotification).toHaveBeenCalledWith({
        title: "Welcome",
        message: "Welcome to the app",
        type: "success",
        data: undefined,
      });
    });

    it("handles malformed messages gracefully", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      // Send invalid JSON directly
      await act(async () => {
        MockWebSocket.instances[0].onmessage?.(new MessageEvent("message", { data: "not valid json" }));
      });

      // Hook should not throw
      expect(result.current.isConnected).toBe(true);
    });

    it("calls custom onMessage handler", async () => {
      const onMessage = vi.fn();
      renderHook(() => useWebSocket({ onMessage }));

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateMessage({
          type: "pong",
        });
      });

      expect(onMessage).toHaveBeenCalledWith({ type: "pong" });
    });
  });

  describe("reconnection", () => {
    it("reconnects with exponential backoff on connection loss", async () => {
      renderHook(() => useWebSocket({ reconnectInterval: 1000, maxRetries: 3 }));

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      expect(MockWebSocket.instances).toHaveLength(1);

      // Close the connection
      await act(async () => {
        MockWebSocket.instances[0].simulateClose();
      });

      // Wait for first reconnect (1000ms * 1 = 1000ms)
      // The reconnect callback is async and calls validateSession, so we need to flush promises
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000);
        await vi.runAllTimersAsync(); // Flush any remaining async work
      });

      expect(MockWebSocket.instances).toHaveLength(2);
    });

    it("stops reconnecting after max retries", async () => {
      renderHook(() => useWebSocket({ reconnectInterval: 100, maxRetries: 2 }));

      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
        MockWebSocket.instances[0].simulateOpen();
      });

      // First close triggers reconnect attempt 1 (retry counter: 0 -> 1)
      await act(async () => {
        MockWebSocket.instances[0].simulateClose();
        await vi.advanceTimersByTimeAsync(100);
        // Flush the validateSession Promise
        await Promise.resolve();
        await Promise.resolve();
      });

      expect(MockWebSocket.instances).toHaveLength(2);

      // Second close triggers reconnect attempt 2 (retry counter: 1 -> 2)
      // Don't call simulateOpen() - this simulates a failed reconnection
      await act(async () => {
        MockWebSocket.instances[1].simulateClose();
        await vi.advanceTimersByTimeAsync(200);
        // Flush the validateSession Promise
        await Promise.resolve();
        await Promise.resolve();
      });

      expect(MockWebSocket.instances).toHaveLength(3);

      // Third close - retry counter is 2, which is NOT < maxRetries (2)
      // So no reconnect should be scheduled
      const countBeforeThirdClose = MockWebSocket.instances.length;
      await act(async () => {
        MockWebSocket.instances[2].simulateClose();
        await vi.advanceTimersByTimeAsync(300);
        await Promise.resolve();
        await Promise.resolve();
      });

      expect(MockWebSocket.instances).toHaveLength(countBeforeThirdClose);
    });

    it("validates session before reconnecting", async () => {
      renderHook(() => useWebSocket({ reconnectInterval: 100 }));

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateClose();
        await vi.advanceTimersByTimeAsync(100);
        await vi.runAllTimersAsync(); // Flush the async validateSession call
      });

      expect(AuthService.validateSession).toHaveBeenCalled();
    });

    it("does not reconnect if session is invalid", async () => {
      vi.mocked(AuthService.validateSession).mockResolvedValue(false);

      renderHook(() => useWebSocket({ reconnectInterval: 100 }));

      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
        MockWebSocket.instances[0].simulateOpen();
      });

      const initialCount = MockWebSocket.instances.length;

      await act(async () => {
        MockWebSocket.instances[0].simulateClose();
        await vi.advanceTimersByTimeAsync(100);
        // Flush the validateSession Promise chain
        await Promise.resolve();
        await Promise.resolve();
      });

      // Should not create new connection (validateSession returned false)
      expect(MockWebSocket.instances).toHaveLength(initialCount);
    });
  });

  describe("ping/pong", () => {
    it("sends ping every 30 seconds when connected", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      const ws = MockWebSocket.instances[0];
      ws.send.mockClear();

      // Advance 30 seconds
      await act(async () => {
        await vi.advanceTimersByTimeAsync(30000);
      });

      expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "ping" }));
    });

    it("does not send ping when disconnected", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      const ws = MockWebSocket.instances[0];
      // Don't call simulateOpen - connection stays in CONNECTING state

      await act(async () => {
        await vi.advanceTimersByTimeAsync(30000);
      });

      // send should not be called for ping (might be called for other reasons)
      expect(ws.send).not.toHaveBeenCalledWith(JSON.stringify({ type: "ping" }));
    });
  });

  describe("manual control", () => {
    it("disconnect clears reconnect timeout", async () => {
      const { result } = renderHook(() => useWebSocket({ reconnectInterval: 1000 }));

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
        MockWebSocket.instances[0].simulateClose();
      });

      // Disconnect before reconnect timer fires
      await act(async () => {
        result.current.disconnect();
      });

      // Advance past reconnect time
      await act(async () => {
        await vi.advanceTimersByTimeAsync(5000);
      });

      // Should not have reconnected
      expect(MockWebSocket.instances).toHaveLength(1);
    });

    it("sendMessage sends when connected", async () => {
      const { result } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      await act(async () => {
        result.current.sendMessage("ping", { test: true });
      });

      expect(MockWebSocket.instances[0].send).toHaveBeenCalledWith(
        JSON.stringify({ type: "ping", payload: { test: true } })
      );
    });

    it("sendMessage does not throw when disconnected", async () => {
      const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

      expect(() => {
        result.current.sendMessage("ping");
      }).not.toThrow();
    });
  });

  describe("URL generation", () => {
    it("uses wss: for https: pages", async () => {
      vi.stubGlobal("location", {
        protocol: "https:",
        host: "example.com",
      });

      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(MockWebSocket.instances[0].url).toBe("wss://example.com/ws");
    });

    it("uses ws: for http: pages", async () => {
      vi.stubGlobal("location", {
        protocol: "http:",
        host: "localhost:3000",
      });

      renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(MockWebSocket.instances[0].url).toBe("ws://localhost:3000/ws");
    });
  });

  describe("cleanup", () => {
    it("cleans up on unmount", async () => {
      const { unmount } = renderHook(() => useWebSocket());

      await act(async () => {
        await vi.runAllTimersAsync();
        MockWebSocket.instances[0].simulateOpen();
      });

      const ws = MockWebSocket.instances[0];

      unmount();

      expect(ws.close).toHaveBeenCalled();
    });
  });
});
