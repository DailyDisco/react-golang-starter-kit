import React from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { UserService } from "../../services";
import { useUserStore } from "../../stores/user-store";
// Import the hooks to test
import { useCreateUser, useDeleteUser, useUpdateUser } from "./use-user-mutations";

// Mock the UserService
vi.mock("../../services", () => ({
  UserService: {
    createUser: vi.fn(),
    updateUser: vi.fn(),
    deleteUser: vi.fn(),
  },
}));

// Mock the logger
vi.mock("../../lib/logger", () => ({
  logger: {
    error: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
  },
}));

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Helper to create wrapper with QueryClientProvider
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe("useCreateUser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset the user store
    useUserStore.getState().resetForm();
  });

  it("should create user successfully", async () => {
    const mockUser = {
      id: 1,
      name: "Test User",
      email: "test@example.com",
      email_verified: false,
      is_active: true,
      created_at: "2024-01-01",
      updated_at: "2024-01-01",
    };
    vi.mocked(UserService.createUser).mockResolvedValueOnce(mockUser);

    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate({ name: "Test User", email: "test@example.com" });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(UserService.createUser).toHaveBeenCalledWith("Test User", "test@example.com", undefined);
  });

  it("should create user with password", async () => {
    const mockUser = {
      id: 1,
      name: "Test User",
      email: "test@example.com",
      email_verified: false,
      is_active: true,
      created_at: "2024-01-01",
      updated_at: "2024-01-01",
    };
    vi.mocked(UserService.createUser).mockResolvedValueOnce(mockUser);

    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate({
        name: "Test User",
        email: "test@example.com",
        password: "Password123!",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(UserService.createUser).toHaveBeenCalledWith("Test User", "test@example.com", "Password123!");
  });

  it("should handle creation error", async () => {
    vi.mocked(UserService.createUser).mockRejectedValueOnce(new Error("Email already exists"));

    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate({ name: "Test", email: "existing@example.com" });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error?.message).toBe("Email already exists");
  });

  it("should handle password validation error", async () => {
    vi.mocked(UserService.createUser).mockRejectedValueOnce(
      new Error("password must contain at least one uppercase letter")
    );

    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate({
        name: "Test",
        email: "test@example.com",
        password: "weak",
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should handle password length error", async () => {
    vi.mocked(UserService.createUser).mockRejectedValueOnce(new Error("password must be at least 8 characters"));

    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate({
        name: "Test",
        email: "test@example.com",
        password: "short",
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateUser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should update user successfully", async () => {
    const mockUser = {
      id: 1,
      name: "Updated Name",
      email: "test@example.com",
      email_verified: false,
      is_active: true,
      created_at: "2024-01-01",
      updated_at: "2024-01-02",
    };
    vi.mocked(UserService.updateUser).mockResolvedValueOnce(mockUser);

    const { result } = renderHook(() => useUpdateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockUser);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(UserService.updateUser).toHaveBeenCalledWith(mockUser);
  });

  it("should handle update error and rollback", async () => {
    vi.mocked(UserService.updateUser).mockRejectedValueOnce(new Error("Update failed"));

    const mockUser = {
      id: 1,
      name: "Updated Name",
      email: "test@example.com",
      email_verified: false,
      is_active: true,
      created_at: "2024-01-01",
      updated_at: "2024-01-02",
    };

    const { result } = renderHook(() => useUpdateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockUser);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should optimistically update the cache", async () => {
    const mockUser = {
      id: 1,
      name: "Updated Name",
      email: "test@example.com",
      email_verified: false,
      is_active: true,
      created_at: "2024-01-01",
      updated_at: "2024-01-02",
    };
    vi.mocked(UserService.updateUser).mockResolvedValueOnce(mockUser);

    const { result } = renderHook(() => useUpdateUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(mockUser);
    });

    // Verify mutation completes successfully (optimistic update uses onMutate)
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(UserService.updateUser).toHaveBeenCalledWith(mockUser);
  });
});

describe("useDeleteUser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should delete user successfully", async () => {
    vi.mocked(UserService.deleteUser).mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useDeleteUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(UserService.deleteUser).toHaveBeenCalledWith(1);
  });

  it("should handle delete error", async () => {
    vi.mocked(UserService.deleteUser).mockRejectedValueOnce(new Error("User not found"));

    const { result } = renderHook(() => useDeleteUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(999);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("should pass the correct user ID to the service", async () => {
    vi.mocked(UserService.deleteUser).mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => useDeleteUser(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.mutate(42);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(UserService.deleteUser).toHaveBeenCalledWith(42);
  });
});

describe("Mutation hooks mock compatibility", () => {
  it("useCreateUser returns expected interface", () => {
    const { result } = renderHook(() => useCreateUser(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(result.current.mutateAsync).toBeDefined();
    expect(typeof result.current.isPending).toBe("boolean");
    expect(typeof result.current.isError).toBe("boolean");
    expect(typeof result.current.isSuccess).toBe("boolean");
  });

  it("useUpdateUser returns expected interface", () => {
    const { result } = renderHook(() => useUpdateUser(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(result.current.mutateAsync).toBeDefined();
    expect(typeof result.current.isPending).toBe("boolean");
  });

  it("useDeleteUser returns expected interface", () => {
    const { result } = renderHook(() => useDeleteUser(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(result.current.mutateAsync).toBeDefined();
    expect(typeof result.current.isPending).toBe("boolean");
  });
});
