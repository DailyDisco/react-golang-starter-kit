import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useHealthCheck } from "./use-health";

// Mock dependencies
vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
}));

vi.mock("../../lib/query-keys", () => ({
  queryKeys: {
    health: {
      status: ["health", "status"],
    },
  },
}));

vi.mock("../../services", () => ({
  API_BASE_URL: "http://localhost:8080",
}));

describe("useHealthCheck", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useHealthCheck).toBeDefined();
    expect(typeof useHealthCheck).toBe("function");
  });

  it("can be mocked to return healthy status", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockHealthData = {
      status: "healthy",
      timestamp: "2024-01-01T00:00:00Z",
      version: "1.0.0",
    };

    vi.mocked(useQuery).mockReturnValue({
      data: mockHealthData,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<typeof mockHealthData, Error>);

    const result = useHealthCheck();
    expect(result.data).toEqual(mockHealthData);
    expect(result.isSuccess).toBe(true);
  });

  it("can be mocked to return unhealthy status", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockHealthData = {
      status: "unhealthy",
      error: "Database connection failed",
    };

    vi.mocked(useQuery).mockReturnValue({
      data: mockHealthData,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<typeof mockHealthData, Error>);

    const result = useHealthCheck();
    expect(result.data?.status).toBe("unhealthy");
  });

  it("can be mocked to return error state", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockError = new Error("Health check failed");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: mockError,
    } as unknown as UseQueryResult<unknown, Error>);

    const result = useHealthCheck();
    expect(result.isError).toBe(true);
    expect(result.error).toBe(mockError);
  });

  it("has correct refetch interval", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<unknown, Error>);

    useHealthCheck();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        refetchInterval: 30000, // 30 seconds
      })
    );
  });

  it("uses correct query key", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<unknown, Error>);

    useHealthCheck();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["health", "status"],
      })
    );
  });
});
