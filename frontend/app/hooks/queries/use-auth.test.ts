import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { User } from "../../services";
import { createMockUser } from "../../test/test-utils";
import { useCurrentUser } from "./use-auth";

// Mock the dependencies
vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
}));

vi.mock("../../lib/query-keys", () => ({
  queryKeys: {
    auth: {
      user: ["auth", "user"],
    },
  },
}));

vi.mock("../../services", () => ({
  AuthService: {
    getCurrentUser: vi.fn(),
  },
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

describe("useCurrentUser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useCurrentUser).toBeDefined();
    expect(typeof useCurrentUser).toBe("function");
  });

  it("can be mocked to return loading state", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue(true);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
      isPending: true,
      isSuccess: false,
      status: "pending",
    } as unknown as UseQueryResult<User, Error>);

    const result = useCurrentUser();
    expect(result.isLoading).toBe(true);
    expect(result.data).toBeUndefined();
  });

  it("can be mocked to return user data", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    const mockUser = createMockUser({ id: 1, name: "Test User" });

    vi.mocked(useAuthStore).mockReturnValue(true);
    vi.mocked(useQuery).mockReturnValue({
      data: mockUser,
      isLoading: false,
      isError: false,
      error: null,
      isPending: false,
      isSuccess: true,
      status: "success",
    } as unknown as UseQueryResult<User, Error>);

    const result = useCurrentUser();
    expect(result.isLoading).toBe(false);
    expect(result.isSuccess).toBe(true);
    expect(result.data).toEqual(mockUser);
  });

  it("can be mocked to return error state", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    const mockError = new Error("Failed to fetch user");

    vi.mocked(useAuthStore).mockReturnValue(true);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: mockError,
      isPending: false,
      isSuccess: false,
      status: "error",
    } as unknown as UseQueryResult<User, Error>);

    const result = useCurrentUser();
    expect(result.isError).toBe(true);
    expect(result.error).toBe(mockError);
  });

  it("query is disabled when not authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue(false);

    useCurrentUser();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });

  it("query is enabled when authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue(true);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<User, Error>);

    useCurrentUser();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: true,
      })
    );
  });
});
