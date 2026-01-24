import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  createMutationErrorHandler,
  createMutationHandlers,
  showMutationError,
  showMutationSuccess,
} from "./mutation-toast";

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

// Mock error-utils
vi.mock("./error-utils", () => ({
  categorizeError: vi.fn((error) => {
    if (error?.message === "Rate limited") {
      return {
        message: "Rate limited",
        category: "network",
        retryable: true,
        retryAfter: 5,
      };
    }
    if (error?.message === "Network error") {
      return {
        message: "Network error",
        category: "network",
        retryable: true,
        details: "Check your connection",
      };
    }
    return {
      message: error?.message || "Unknown error",
      category: "unknown",
      retryable: false,
    };
  }),
}));

describe("mutation-toast", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("showMutationError", () => {
    it("should show error toast with categorized message", async () => {
      const { toast } = await import("sonner");

      const result = showMutationError({
        error: new Error("Something went wrong"),
      });

      expect(toast.error).toHaveBeenCalledWith(
        "Something went wrong",
        expect.objectContaining({
          duration: 5000,
        })
      );
      expect(result).toEqual({
        message: "Something went wrong",
        category: "unknown",
        retryable: false,
      });
    });

    it("should add context to error message", async () => {
      const { toast } = await import("sonner");

      showMutationError({
        error: new Error("Failed"),
        context: "Saving user",
      });

      expect(toast.error).toHaveBeenCalledWith("Saving user: Failed", expect.any(Object));
    });

    it("should show retry button for retryable errors", async () => {
      const { toast } = await import("sonner");
      const onRetry = vi.fn();

      showMutationError({
        error: new Error("Network error"),
        onRetry,
      });

      expect(toast.error).toHaveBeenCalledWith(
        "Network error",
        expect.objectContaining({
          duration: 8000,
          action: expect.objectContaining({
            label: "Retry",
          }),
        })
      );
    });

    it("should not show retry button for non-retryable errors", async () => {
      const { toast } = await import("sonner");

      showMutationError({
        error: new Error("Validation error"),
      });

      expect(toast.error).toHaveBeenCalledWith(
        "Validation error",
        expect.objectContaining({
          action: undefined,
        })
      );
    });

    it("should use custom duration when provided", async () => {
      const { toast } = await import("sonner");

      showMutationError({
        error: new Error("Error"),
        duration: 10000,
      });

      expect(toast.error).toHaveBeenCalledWith(
        "Error",
        expect.objectContaining({
          duration: 10000,
        })
      );
    });

    it("should handle rate limit errors with countdown", async () => {
      const { toast } = await import("sonner");

      const result = showMutationError({
        error: new Error("Rate limited"),
      });

      // Initial toast should be shown
      expect(toast.error).toHaveBeenCalled();
      expect(result.retryAfter).toBe(5);
    });
  });

  describe("showMutationSuccess", () => {
    it("should show success toast with message", async () => {
      const { toast } = await import("sonner");

      showMutationSuccess({
        message: "User created successfully",
      });

      expect(toast.success).toHaveBeenCalledWith("User created successfully", {
        description: undefined,
        duration: 3000,
      });
    });

    it("should show success toast with description", async () => {
      const { toast } = await import("sonner");

      showMutationSuccess({
        message: "User created",
        description: "They will receive a welcome email",
      });

      expect(toast.success).toHaveBeenCalledWith("User created", {
        description: "They will receive a welcome email",
        duration: 3000,
      });
    });

    it("should use custom duration", async () => {
      const { toast } = await import("sonner");

      showMutationSuccess({
        message: "Done",
        duration: 5000,
      });

      expect(toast.success).toHaveBeenCalledWith("Done", {
        description: undefined,
        duration: 5000,
      });
    });
  });

  describe("createMutationErrorHandler", () => {
    it("should create an error handler function", () => {
      const handler = createMutationErrorHandler();

      expect(typeof handler).toBe("function");
    });

    it("should call showMutationError when invoked", async () => {
      const { toast } = await import("sonner");
      const handler = createMutationErrorHandler();

      handler(new Error("Test error"), undefined);

      expect(toast.error).toHaveBeenCalled();
    });

    it("should pass retry function to showMutationError", async () => {
      const { toast } = await import("sonner");
      const retryFn = vi.fn();
      const handler = createMutationErrorHandler<number>(retryFn);

      handler(new Error("Network error"), 123);

      expect(toast.error).toHaveBeenCalledWith(
        "Network error",
        expect.objectContaining({
          action: expect.objectContaining({
            label: "Retry",
          }),
        })
      );
    });

    it("should return categorized error", () => {
      const handler = createMutationErrorHandler();

      const result = handler(new Error("Test error"), undefined);

      expect(result).toEqual({
        message: "Test error",
        category: "unknown",
        retryable: false,
      });
    });
  });

  describe("createMutationHandlers", () => {
    it("should return onSuccess and onError handlers", () => {
      const handlers = createMutationHandlers({
        successMessage: "Success!",
      });

      expect(typeof handlers.onSuccess).toBe("function");
      expect(typeof handlers.onError).toBe("function");
    });

    it("should call showMutationSuccess on onSuccess", async () => {
      const { toast } = await import("sonner");
      const handlers = createMutationHandlers({
        successMessage: "User saved",
        successDescription: "All changes saved",
      });

      handlers.onSuccess();

      expect(toast.success).toHaveBeenCalledWith("User saved", {
        description: "All changes saved",
        duration: 3000,
      });
    });

    it("should call showMutationError on onError", async () => {
      const { toast } = await import("sonner");
      const handlers = createMutationHandlers({
        successMessage: "Success",
      });

      handlers.onError(new Error("Failed"), undefined);

      expect(toast.error).toHaveBeenCalled();
    });

    it("should pass retry function to error handler", async () => {
      const { toast } = await import("sonner");
      const onRetry = vi.fn();
      const handlers = createMutationHandlers<number>({
        successMessage: "Done",
        onRetry,
      });

      handlers.onError(new Error("Network error"), 42);

      expect(toast.error).toHaveBeenCalledWith(
        "Network error",
        expect.objectContaining({
          action: expect.objectContaining({
            label: "Retry",
          }),
        })
      );
    });
  });
});
