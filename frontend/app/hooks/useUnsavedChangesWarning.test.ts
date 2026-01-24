import { renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from "vitest";

import { useUnsavedChangesDialog, useUnsavedChangesWarning } from "./useUnsavedChangesWarning";

// Mock TanStack Router's useBlocker
const mockBlocker = {
  status: "idle",
  proceed: vi.fn(),
  reset: vi.fn(),
};

vi.mock("@tanstack/react-router", () => ({
  useBlocker: vi.fn(({ shouldBlockFn }: { shouldBlockFn: () => boolean }) => {
    // Store the shouldBlockFn for testing
    (global as any).__shouldBlockFn = shouldBlockFn;
    return mockBlocker;
  }),
}));

// Store mock reference for type-safe access
let confirmMock: Mock<(message?: string) => boolean>;

describe("useUnsavedChangesWarning", () => {
  let addEventListenerSpy: ReturnType<typeof vi.spyOn>;
  let removeEventListenerSpy: ReturnType<typeof vi.spyOn>;
  let originalConfirm: typeof window.confirm | undefined;

  beforeEach(() => {
    vi.clearAllMocks();
    mockBlocker.status = "idle";

    addEventListenerSpy = vi.spyOn(window, "addEventListener");
    removeEventListenerSpy = vi.spyOn(window, "removeEventListener");

    // Save and mock window.confirm (may not exist in jsdom)
    originalConfirm = window.confirm;
    confirmMock = vi.fn(() => false);
    window.confirm = confirmMock;
  });

  afterEach(() => {
    addEventListenerSpy.mockRestore();
    removeEventListenerSpy.mockRestore();
    // Restore original confirm (or undefined)
    if (originalConfirm !== undefined) {
      window.confirm = originalConfirm;
    }
  });

  it("adds beforeunload listener when dirty", () => {
    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
      })
    );

    expect(addEventListenerSpy).toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });

  it("does not add beforeunload listener when not dirty", () => {
    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: false,
      })
    );

    expect(addEventListenerSpy).not.toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });

  it("removes beforeunload listener on cleanup", () => {
    const { unmount } = renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
      })
    );

    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });

  it("removes and re-adds listener when isDirty changes", () => {
    const { rerender } = renderHook(({ isDirty }) => useUnsavedChangesWarning({ isDirty }), {
      initialProps: { isDirty: true },
    });

    expect(addEventListenerSpy).toHaveBeenCalledTimes(1);

    rerender({ isDirty: false });

    expect(removeEventListenerSpy).toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });

  it("blocks router navigation when dirty and user cancels", () => {
    confirmMock.mockReturnValue(false);

    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
      })
    );

    const shouldBlock = (global as any).__shouldBlockFn;
    expect(shouldBlock()).toBe(true); // Blocked because confirm returned false
    expect(window.confirm).toHaveBeenCalled();
  });

  it("allows router navigation when user confirms", () => {
    confirmMock.mockReturnValue(true);

    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
      })
    );

    const shouldBlock = (global as any).__shouldBlockFn;
    expect(shouldBlock()).toBe(false); // Not blocked because confirm returned true
  });

  it("allows navigation when not dirty", () => {
    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: false,
      })
    );

    const shouldBlock = (global as any).__shouldBlockFn;
    expect(shouldBlock()).toBe(false);
    expect(window.confirm).not.toHaveBeenCalled();
  });

  it("uses custom message", () => {
    confirmMock.mockReturnValue(false);
    const customMessage = "Custom unsaved changes message";

    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
        message: customMessage,
      })
    );

    const shouldBlock = (global as any).__shouldBlockFn;
    shouldBlock();

    expect(window.confirm).toHaveBeenCalledWith(customMessage);
  });

  it("uses custom shouldBlock function", () => {
    const customShouldBlock = vi.fn().mockReturnValue(false);

    renderHook(() =>
      useUnsavedChangesWarning({
        isDirty: true,
        shouldBlock: customShouldBlock,
      })
    );

    const shouldBlock = (global as any).__shouldBlockFn;
    expect(shouldBlock()).toBe(false);
    expect(customShouldBlock).toHaveBeenCalled();
    expect(window.confirm).not.toHaveBeenCalled();
  });
});

describe("useUnsavedChangesDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockBlocker.status = "idle";
  });

  it("returns isBlocked false when not blocked", () => {
    mockBlocker.status = "idle";

    const { result } = renderHook(() =>
      useUnsavedChangesDialog({
        isDirty: true,
      })
    );

    expect(result.current.isBlocked).toBe(false);
  });

  it("returns isBlocked true when blocked", () => {
    mockBlocker.status = "blocked";

    const { result } = renderHook(() =>
      useUnsavedChangesDialog({
        isDirty: true,
      })
    );

    expect(result.current.isBlocked).toBe(true);
  });

  it("calls blocker.proceed when proceed is called", () => {
    mockBlocker.status = "blocked";

    const { result } = renderHook(() =>
      useUnsavedChangesDialog({
        isDirty: true,
      })
    );

    result.current.proceed();

    expect(mockBlocker.proceed).toHaveBeenCalled();
  });

  it("calls blocker.reset when cancel is called", () => {
    mockBlocker.status = "blocked";

    const { result } = renderHook(() =>
      useUnsavedChangesDialog({
        isDirty: true,
      })
    );

    result.current.cancel();

    expect(mockBlocker.reset).toHaveBeenCalled();
  });

  it("does not call proceed when not blocked", () => {
    mockBlocker.status = "idle";

    const { result } = renderHook(() =>
      useUnsavedChangesDialog({
        isDirty: true,
      })
    );

    result.current.proceed();

    expect(mockBlocker.proceed).not.toHaveBeenCalled();
  });
});
