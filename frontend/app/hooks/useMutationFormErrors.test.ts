import { renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useMutationFormErrors } from "./useMutationFormErrors";

// Mock ApiError
class MockApiError extends Error {
  code: string;
  statusCode: number;

  constructor(message: string, code: string, statusCode: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.statusCode = statusCode;
  }
}

// Create mock form
const createMockForm = () => {
  const watchCallbacks: Array<(values: unknown, info: { name?: string }) => void> = [];

  return {
    setError: vi.fn(),
    clearErrors: vi.fn(),
    formState: {
      errors: {} as Record<string, { type?: string; message?: string }>,
    },
    watch: vi.fn((callback?: (values: unknown, info: { name?: string }) => void) => {
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
    // Helper to trigger watch
    _triggerWatch: (name?: string) => {
      for (const cb of watchCallbacks) {
        cb({}, { name });
      }
    },
  };
};

// Create mock mutation
const createMockMutation = (error: Error | null = null) => ({
  error,
  isPending: false,
  isError: error !== null,
  isSuccess: false,
  data: undefined,
  mutate: vi.fn(),
  mutateAsync: vi.fn(),
  reset: vi.fn(),
  status: error ? "error" : "idle",
});

describe("useMutationFormErrors", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does nothing when no error", () => {
    const mockForm = createMockForm();
    const mockMutation = createMockMutation(null);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).not.toHaveBeenCalled();
  });

  it("sets root error for non-field validation errors", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Something went wrong", "ERROR", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("root", {
      type: "server",
      message: "Something went wrong",
    });
  });

  it("parses field:message format and sets field error", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("email: already exists", "VALIDATION_ERROR", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("email", {
      type: "server",
      message: "already exists",
    });
  });

  it("uses field mapping to translate field names", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("user_email: invalid format", "VALIDATION_ERROR", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() =>
      useMutationFormErrors(mockForm as any, mockMutation as any, {
        fieldMapping: {
          user_email: "email",
        },
      })
    );

    expect(mockForm.setError).toHaveBeenCalledWith("email", {
      type: "server",
      message: "invalid format",
    });
  });

  it("detects email exists error pattern", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Email already exists", "CONFLICT", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("email", {
      type: "server",
      message: "This email is already registered",
    });
  });

  it("detects password uppercase error pattern", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Password must contain at least one uppercase letter", "VALIDATION_ERROR", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("password", {
      type: "server",
      message: "Password must contain at least one uppercase letter",
    });
  });

  it("detects password length error pattern", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Password must be at least 8 characters", "VALIDATION_ERROR", 400);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("password", {
      type: "server",
      message: "Password must be at least 8 characters",
    });
  });

  it("clears server error when field changes", () => {
    const mockForm = createMockForm();
    // Set up existing server error
    mockForm.formState.errors = {
      email: { type: "server", message: "Already exists" },
    };

    const mockMutation = createMockMutation(null);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    // Simulate user typing in email field
    mockForm._triggerWatch("email");

    expect(mockForm.clearErrors).toHaveBeenCalledWith("email");
  });

  it("clears root server error when any field changes", () => {
    const mockForm = createMockForm();
    mockForm.formState.errors = {
      root: { type: "server", message: "Server error" },
    };

    const mockMutation = createMockMutation(null);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    mockForm._triggerWatch("anyField");

    expect(mockForm.clearErrors).toHaveBeenCalledWith("root");
  });

  it("does not clear non-server errors", () => {
    const mockForm = createMockForm();
    mockForm.formState.errors = {
      email: { type: "required", message: "Email is required" },
    };

    const mockMutation = createMockMutation(null);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    mockForm._triggerWatch("email");

    // Should not clear because it's not a server error
    expect(mockForm.clearErrors).not.toHaveBeenCalled();
  });

  it("respects clearOnChange: false option", () => {
    const mockForm = createMockForm();
    mockForm.formState.errors = {
      email: { type: "server", message: "Already exists" },
    };

    const mockMutation = createMockMutation(null);

    renderHook(() =>
      useMutationFormErrors(mockForm as any, mockMutation as any, {
        clearOnChange: false,
      })
    );

    mockForm._triggerWatch("email");

    expect(mockForm.clearErrors).not.toHaveBeenCalled();
  });

  it("sets root error for server errors (500+)", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Internal server error", "INTERNAL_ERROR", 500);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("root", {
      type: "server",
      message: "Something went wrong on our end",
    });
  });

  it("sets root error for auth errors", () => {
    const mockForm = createMockForm();
    const error = new MockApiError("Unauthorized", "UNAUTHORIZED", 401);
    const mockMutation = createMockMutation(error);

    renderHook(() => useMutationFormErrors(mockForm as any, mockMutation as any));

    expect(mockForm.setError).toHaveBeenCalledWith("root", {
      type: "server",
      message: "Please log in to continue",
    });
  });
});
