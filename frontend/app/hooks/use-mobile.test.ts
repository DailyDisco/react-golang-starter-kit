import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useIsMobile } from "./use-mobile";

describe("useIsMobile", () => {
  let originalInnerWidth: number;
  let matchMediaListeners: Array<(e: { matches: boolean }) => void> = [];

  const mockMatchMedia = (matches: boolean) => {
    return vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn((event, handler) => {
        if (event === "change") {
          matchMediaListeners.push(handler);
        }
      }),
      removeEventListener: vi.fn((event, handler) => {
        if (event === "change") {
          matchMediaListeners = matchMediaListeners.filter((h) => h !== handler);
        }
      }),
      dispatchEvent: vi.fn(),
    }));
  };

  beforeEach(() => {
    originalInnerWidth = window.innerWidth;
    matchMediaListeners = [];
  });

  afterEach(() => {
    Object.defineProperty(window, "innerWidth", {
      value: originalInnerWidth,
      writable: true,
      configurable: true,
    });
    vi.restoreAllMocks();
  });

  it("returns false on desktop viewport (>= 768px)", () => {
    Object.defineProperty(window, "innerWidth", {
      value: 1024,
      writable: true,
      configurable: true,
    });
    window.matchMedia = mockMatchMedia(false);

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(false);
  });

  it("returns true on mobile viewport (< 768px)", () => {
    Object.defineProperty(window, "innerWidth", {
      value: 500,
      writable: true,
      configurable: true,
    });
    window.matchMedia = mockMatchMedia(true);

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(true);
  });

  it("returns true at exactly 767px (boundary)", () => {
    Object.defineProperty(window, "innerWidth", {
      value: 767,
      writable: true,
      configurable: true,
    });
    window.matchMedia = mockMatchMedia(true);

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(true);
  });

  it("returns false at exactly 768px (boundary)", () => {
    Object.defineProperty(window, "innerWidth", {
      value: 768,
      writable: true,
      configurable: true,
    });
    window.matchMedia = mockMatchMedia(false);

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(false);
  });

  it("uses correct media query", () => {
    window.matchMedia = mockMatchMedia(false);

    renderHook(() => useIsMobile());

    expect(window.matchMedia).toHaveBeenCalledWith("(max-width: 767px)");
  });

  it("adds event listener for changes", () => {
    const mockMql = {
      matches: false,
      media: "(max-width: 767px)",
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    };

    window.matchMedia = vi.fn().mockReturnValue(mockMql);

    renderHook(() => useIsMobile());

    expect(mockMql.addEventListener).toHaveBeenCalledWith("change", expect.any(Function));
  });

  it("removes event listener on unmount", () => {
    const mockMql = {
      matches: false,
      media: "(max-width: 767px)",
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    };

    window.matchMedia = vi.fn().mockReturnValue(mockMql);

    const { unmount } = renderHook(() => useIsMobile());
    unmount();

    expect(mockMql.removeEventListener).toHaveBeenCalledWith("change", expect.any(Function));
  });

  it("updates value when viewport changes", () => {
    Object.defineProperty(window, "innerWidth", {
      value: 1024,
      writable: true,
      configurable: true,
    });

    const mockMql = {
      matches: false,
      media: "(max-width: 767px)",
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn((event: string, handler: (e: { matches: boolean }) => void) => {
        if (event === "change") {
          matchMediaListeners.push(handler);
        }
      }),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    };

    window.matchMedia = vi.fn().mockReturnValue(mockMql);

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(false);

    // Simulate viewport change to mobile
    act(() => {
      Object.defineProperty(window, "innerWidth", {
        value: 500,
        writable: true,
        configurable: true,
      });
      for (const handler of matchMediaListeners) handler({ matches: true });
    });

    expect(result.current).toBe(true);
  });
});
