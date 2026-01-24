import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useFormPersist } from "./useFormPersist";

// Mock localStorage
const createMockStorage = () => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: () => {
      store = {};
    },
  };
};

// Mock react-hook-form
const createMockForm = (initialValues: Record<string, unknown> = {}) => {
  let currentValues = { ...initialValues };
  const watchCallbacks: Array<(values: Record<string, unknown>) => void> = [];

  return {
    watch: vi.fn((callback?: (values: Record<string, unknown>) => void) => {
      if (callback) {
        watchCallbacks.push(callback);
      }
      return {
        unsubscribe: vi.fn(() => {
          const idx = callback ? watchCallbacks.indexOf(callback) : -1;
          if (idx > -1) watchCallbacks.splice(idx, 1);
        }),
      };
    }),
    reset: vi.fn((values: Record<string, unknown>) => {
      currentValues = { ...values };
    }),
    getValues: vi.fn(() => currentValues),
    // Helper to trigger watch callbacks
    _triggerWatch: (values: Record<string, unknown>) => {
      currentValues = values;
      for (const cb of watchCallbacks) {
        cb(values);
      }
    },
  };
};

describe("useFormPersist", () => {
  let mockStorage: ReturnType<typeof createMockStorage>;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    mockStorage = createMockStorage();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("restores saved draft on mount", () => {
    const savedData = { name: "John", email: "john@example.com" };
    mockStorage.setItem("form_draft_test", JSON.stringify(savedData));

    const mockForm = createMockForm();

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    expect(mockForm.reset).toHaveBeenCalledWith(savedData, { keepDefaultValues: true });
  });

  it("excludes specified fields when restoring", () => {
    const savedData = { name: "John", email: "john@example.com", password: "secret" };
    mockStorage.setItem("form_draft_test", JSON.stringify(savedData));

    const mockForm = createMockForm();

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        exclude: ["password"] as any,
        storage: mockStorage as unknown as Storage,
      })
    );

    // Password should be excluded from restored data
    expect(mockForm.reset).toHaveBeenCalledWith(
      { name: "John", email: "john@example.com" },
      { keepDefaultValues: true }
    );
  });

  it("saves form data with debounce", async () => {
    const mockForm = createMockForm({ name: "", email: "" });

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        debounceMs: 500,
        storage: mockStorage as unknown as Storage,
      })
    );

    // Trigger a form change
    act(() => {
      mockForm._triggerWatch({ name: "John", email: "john@example.com" });
    });

    // Should not save immediately
    expect(mockStorage.setItem).not.toHaveBeenCalledWith("form_draft_test", expect.any(String));

    // Advance past debounce
    await act(async () => {
      vi.advanceTimersByTime(600);
    });

    // Now it should save
    expect(mockStorage.setItem).toHaveBeenCalledWith(
      "form_draft_test",
      JSON.stringify({ name: "John", email: "john@example.com" })
    );
  });

  it("excludes fields when saving", async () => {
    const mockForm = createMockForm({ name: "", email: "", password: "" });

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        exclude: ["password"] as any,
        debounceMs: 100,
        storage: mockStorage as unknown as Storage,
      })
    );

    act(() => {
      mockForm._triggerWatch({ name: "John", email: "john@example.com", password: "secret" });
    });

    await act(async () => {
      vi.advanceTimersByTime(200);
    });

    // Password should be excluded
    expect(mockStorage.setItem).toHaveBeenCalledWith(
      "form_draft_test",
      JSON.stringify({ name: "John", email: "john@example.com" })
    );
  });

  it("clearDraft removes stored data", () => {
    mockStorage.setItem("form_draft_test", JSON.stringify({ name: "John" }));
    mockStorage.setItem("form_draft_test_timestamp", "12345");

    const mockForm = createMockForm();

    const { result } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    act(() => {
      result.current.clearDraft();
    });

    expect(mockStorage.removeItem).toHaveBeenCalledWith("form_draft_test");
    expect(mockStorage.removeItem).toHaveBeenCalledWith("form_draft_test_timestamp");
  });

  it("hasDraft returns true when draft exists", () => {
    mockStorage.setItem("form_draft_test", JSON.stringify({ name: "John" }));

    const mockForm = createMockForm();

    const { result } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    expect(result.current.hasDraft()).toBe(true);
  });

  it("hasDraft returns false when no draft exists", () => {
    const mockForm = createMockForm();

    const { result } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    expect(result.current.hasDraft()).toBe(false);
  });

  it("getDraftTimestamp returns saved timestamp", () => {
    mockStorage.setItem("form_draft_test_timestamp", "1700000000000");

    const mockForm = createMockForm();

    const { result } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    expect(result.current.getDraftTimestamp()).toBe(1700000000000);
  });

  it("getDraftTimestamp returns null when no draft", () => {
    const mockForm = createMockForm();

    const { result } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    expect(result.current.getDraftTimestamp()).toBeNull();
  });

  it("handles invalid JSON in storage gracefully", () => {
    mockStorage.setItem("form_draft_test", "invalid json");

    const mockForm = createMockForm();

    // Should not throw
    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    // Should clear invalid data
    expect(mockStorage.removeItem).toHaveBeenCalledWith("form_draft_test");
    expect(mockForm.reset).not.toHaveBeenCalled();
  });

  it("calls onError when restore fails with invalid JSON", () => {
    mockStorage.setItem("form_draft_test", "invalid json");

    const mockForm = createMockForm();
    const onError = vi.fn();

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
        onError,
      })
    );

    expect(onError).toHaveBeenCalledWith(expect.any(Error), "restore");
  });

  it("calls onError when save fails due to storage error", async () => {
    const mockForm = createMockForm({ name: "" });
    const onError = vi.fn();

    // Make setItem throw (e.g., quota exceeded)
    const errorStorage = {
      ...mockStorage,
      setItem: vi.fn(() => {
        throw new Error("QuotaExceededError");
      }),
    };

    renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        debounceMs: 100,
        storage: errorStorage as unknown as Storage,
        onError,
      })
    );

    act(() => {
      mockForm._triggerWatch({ name: "John" });
    });

    await act(async () => {
      vi.advanceTimersByTime(200);
    });

    expect(onError).toHaveBeenCalledWith(expect.any(Error), "save");
    expect(onError.mock.calls[0][0].message).toBe("QuotaExceededError");
  });

  it("only restores once", () => {
    mockStorage.setItem("form_draft_test", JSON.stringify({ name: "John" }));

    const mockForm = createMockForm();

    const { rerender } = renderHook(() =>
      useFormPersist(mockForm as any, {
        key: "test",
        storage: mockStorage as unknown as Storage,
      })
    );

    // First render restores
    expect(mockForm.reset).toHaveBeenCalledTimes(1);

    // Rerender should not restore again
    rerender();
    expect(mockForm.reset).toHaveBeenCalledTimes(1);
  });
});
