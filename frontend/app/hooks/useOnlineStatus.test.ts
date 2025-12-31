import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useOnlineStatus } from "./useOnlineStatus";

describe("useOnlineStatus", () => {
  let originalNavigator: typeof navigator.onLine;
  let onlineListeners: Array<() => void> = [];
  let offlineListeners: Array<() => void> = [];

  beforeEach(() => {
    originalNavigator = navigator.onLine;

    // Mock navigator.onLine
    Object.defineProperty(navigator, "onLine", {
      value: true,
      writable: true,
      configurable: true,
    });

    // Capture event listeners
    onlineListeners = [];
    offlineListeners = [];

    vi.spyOn(window, "addEventListener").mockImplementation((event, handler) => {
      if (event === "online") {
        onlineListeners.push(handler as () => void);
      } else if (event === "offline") {
        offlineListeners.push(handler as () => void);
      }
    });

    vi.spyOn(window, "removeEventListener").mockImplementation((event, handler) => {
      if (event === "online") {
        onlineListeners = onlineListeners.filter((h) => h !== handler);
      } else if (event === "offline") {
        offlineListeners = offlineListeners.filter((h) => h !== handler);
      }
    });

    vi.useFakeTimers();
  });

  afterEach(() => {
    Object.defineProperty(navigator, "onLine", {
      value: originalNavigator,
      writable: true,
      configurable: true,
    });
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("returns initial online state based on navigator.onLine", () => {
    const { result } = renderHook(() => useOnlineStatus());

    expect(result.current.isOnline).toBe(true);
    expect(result.current.wasOffline).toBe(false);
  });

  it("returns offline state when navigator.onLine is false", () => {
    Object.defineProperty(navigator, "onLine", {
      value: false,
      writable: true,
      configurable: true,
    });

    const { result } = renderHook(() => useOnlineStatus());

    expect(result.current.isOnline).toBe(false);
  });

  it("adds event listeners on mount", () => {
    renderHook(() => useOnlineStatus());

    expect(window.addEventListener).toHaveBeenCalledWith("online", expect.any(Function));
    expect(window.addEventListener).toHaveBeenCalledWith("offline", expect.any(Function));
  });

  it("removes event listeners on unmount", () => {
    const { unmount } = renderHook(() => useOnlineStatus());

    unmount();

    expect(window.removeEventListener).toHaveBeenCalledWith("online", expect.any(Function));
    expect(window.removeEventListener).toHaveBeenCalledWith("offline", expect.any(Function));
  });

  it("updates isOnline to false when offline event fires", () => {
    const { result } = renderHook(() => useOnlineStatus());

    expect(result.current.isOnline).toBe(true);

    act(() => {
      // Trigger offline event
      for (const handler of offlineListeners) handler();
    });

    expect(result.current.isOnline).toBe(false);
  });

  it("updates isOnline to true and sets wasOffline when online event fires", () => {
    Object.defineProperty(navigator, "onLine", {
      value: false,
      writable: true,
      configurable: true,
    });

    const { result } = renderHook(() => useOnlineStatus());

    expect(result.current.isOnline).toBe(false);
    expect(result.current.wasOffline).toBe(false);

    act(() => {
      // Trigger online event
      for (const handler of onlineListeners) handler();
    });

    expect(result.current.isOnline).toBe(true);
    expect(result.current.wasOffline).toBe(true);
  });

  it("resets wasOffline to false after 5 seconds", () => {
    Object.defineProperty(navigator, "onLine", {
      value: false,
      writable: true,
      configurable: true,
    });

    const { result } = renderHook(() => useOnlineStatus());

    act(() => {
      // Trigger online event
      for (const handler of onlineListeners) handler();
    });

    expect(result.current.wasOffline).toBe(true);

    // Fast-forward 5 seconds
    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(result.current.wasOffline).toBe(false);
  });

  it("handles multiple offline/online transitions", () => {
    const { result } = renderHook(() => useOnlineStatus());

    // Initially online
    expect(result.current.isOnline).toBe(true);

    // Go offline
    act(() => {
      for (const handler of offlineListeners) handler();
    });
    expect(result.current.isOnline).toBe(false);

    // Come back online
    act(() => {
      for (const handler of onlineListeners) handler();
    });
    expect(result.current.isOnline).toBe(true);
    expect(result.current.wasOffline).toBe(true);

    // Go offline again
    act(() => {
      for (const handler of offlineListeners) handler();
    });
    expect(result.current.isOnline).toBe(false);
  });
});
