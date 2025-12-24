import { beforeEach, describe, expect, it, vi } from "vitest";

// Import the mocked hooks
import { useLogin, useRegister } from "./use-auth-mutations";

// The hooks are globally mocked in setup.ts
// This test verifies the mock integration works correctly

describe("useLogin mock", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns a mock function", () => {
    expect(useLogin).toBeDefined();
    expect(typeof useLogin).toBe("function");
  });

  it("can be mocked to return specific mutation state", () => {
    const mockMutate = vi.fn();
    vi.mocked(useLogin).mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      reset: vi.fn(),
      variables: undefined,
      status: "idle",
      failureCount: 0,
      failureReason: null,
      isIdle: true,
      isPaused: false,
      context: undefined,
      submittedAt: 0,
    });

    const result = useLogin();
    expect(result.mutate).toBe(mockMutate);
    expect(result.isPending).toBe(false);
  });

  it("can simulate pending state", () => {
    vi.mocked(useLogin).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: true,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      reset: vi.fn(),
      variables: undefined,
      status: "pending",
      failureCount: 0,
      failureReason: null,
      isIdle: false,
      isPaused: false,
      context: undefined,
      submittedAt: Date.now(),
    });

    const result = useLogin();
    expect(result.isPending).toBe(true);
  });

  it("can simulate error state", () => {
    const mockError = new Error("Login failed");
    vi.mocked(useLogin).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isError: true,
      isSuccess: false,
      error: mockError,
      data: undefined,
      reset: vi.fn(),
      variables: undefined,
      status: "error",
      failureCount: 1,
      failureReason: mockError,
      isIdle: false,
      isPaused: false,
      context: undefined,
      submittedAt: 0,
    });

    const result = useLogin();
    expect(result.isError).toBe(true);
    expect(result.error).toBe(mockError);
  });
});

describe("useRegister mock", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns a mock function", () => {
    expect(useRegister).toBeDefined();
    expect(typeof useRegister).toBe("function");
  });

  it("can be mocked to return specific mutation state", () => {
    const mockMutate = vi.fn();
    vi.mocked(useRegister).mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isError: false,
      isSuccess: true,
      error: null,
      data: { user: { id: 1, name: "Test" }, token: "token" },
      reset: vi.fn(),
      variables: undefined,
      status: "success",
      failureCount: 0,
      failureReason: null,
      isIdle: false,
      isPaused: false,
      context: undefined,
      submittedAt: 0,
    });

    const result = useRegister();
    expect(result.mutate).toBe(mockMutate);
    expect(result.isSuccess).toBe(true);
  });
});
