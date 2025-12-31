import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { User } from "../../services";
import { createMockUser } from "../../test/test-utils";
import { useUser, useUsers } from "./use-users";

// Mock dependencies
vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
}));

vi.mock("../../lib/query-keys", () => ({
  queryKeys: {
    users: {
      all: ["users"],
      lists: () => ["users", "list"],
      list: (filters: Record<string, unknown>) => ["users", "list", filters],
      details: () => ["users", "detail"],
      detail: (id: number) => ["users", "detail", id],
    },
  },
}));

vi.mock("../../services", () => ({
  UserService: {
    fetchUsers: vi.fn(),
    getUserById: vi.fn(),
  },
}));

vi.mock("../../stores/user-store", () => ({
  useUserStore: vi.fn(),
}));

describe("useUsers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useUsers).toBeDefined();
    expect(typeof useUsers).toBe("function");
  });

  it("can be mocked to return users list", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    const mockUsers = [
      createMockUser({ id: 1, name: "User 1" }),
      createMockUser({ id: 2, name: "User 2" }),
      createMockUser({ id: 3, name: "User 3" }),
    ];

    vi.mocked(useUserStore).mockReturnValue({ search: "", role: "", isActive: true });
    vi.mocked(useQuery).mockReturnValue({
      data: mockUsers,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<User[], Error>);

    const result = useUsers();
    expect(result.data).toEqual(mockUsers);
    expect(result.isSuccess).toBe(true);
  });

  it("can be mocked to return loading state", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    vi.mocked(useUserStore).mockReturnValue({ search: "", role: "", isActive: true });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isPending: true,
      isError: false,
      isSuccess: false,
    } as unknown as UseQueryResult<User[], Error>);

    const result = useUsers();
    expect(result.isLoading).toBe(true);
    expect(result.data).toBeUndefined();
  });

  it("can be mocked to return error state", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    const mockError = new Error("Failed to fetch users");

    vi.mocked(useUserStore).mockReturnValue({ search: "", role: "", isActive: true });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: mockError,
    } as unknown as UseQueryResult<User[], Error>);

    const result = useUsers();
    expect(result.isError).toBe(true);
    expect(result.error).toBe(mockError);
  });

  it("includes filters from user store in query key", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    const filters = { search: "test", role: "admin", isActive: true };
    vi.mocked(useUserStore).mockReturnValue(filters);
    vi.mocked(useQuery).mockReturnValue({
      data: [],
      isLoading: false,
    } as unknown as UseQueryResult<User[], Error>);

    useUsers();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["users", "list", filters],
      })
    );
  });
});

describe("useUser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useUser).toBeDefined();
    expect(typeof useUser).toBe("function");
  });

  it("can be mocked to return single user", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    const mockUser = createMockUser({ id: 5, name: "Selected User" });

    vi.mocked(useUserStore).mockReturnValue(5);
    vi.mocked(useQuery).mockReturnValue({
      data: mockUser,
      isLoading: false,
      isSuccess: true,
    } as unknown as UseQueryResult<User, Error>);

    const result = useUser();
    expect(result.data).toEqual(mockUser);
    expect(result.isSuccess).toBe(true);
  });

  it("is enabled when selectedUserId exists", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    vi.mocked(useUserStore).mockReturnValue(10);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<User, Error>);

    useUser();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: true,
      })
    );
  });

  it("is disabled when selectedUserId is null", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    vi.mocked(useUserStore).mockReturnValue(null);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as UseQueryResult<User, Error>);

    useUser();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });

  it("uses correct query key with user id", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useUserStore } = await import("../../stores/user-store");

    vi.mocked(useUserStore).mockReturnValue(42);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<User, Error>);

    useUser();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["users", "detail", 42],
      })
    );
  });
});
