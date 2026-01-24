import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useClearFeatureFlags, useFeatureFlag, useFeatureFlags, useInvalidateFeatureFlags } from "./use-feature-flags";

// Mock dependencies
const mockInvalidateQueries = vi.fn();
const mockRemoveQueries = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: mockInvalidateQueries,
    removeQueries: mockRemoveQueries,
  })),
}));

vi.mock("../../lib/query-keys", () => ({
  queryKeys: {
    featureFlags: {
      all: ["featureFlags"],
      user: () => ["featureFlags", "user"],
    },
  },
}));

vi.mock("../../services/admin/adminService", () => ({
  FeatureFlagService: {
    getFlags: vi.fn(),
  },
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

describe("useFeatureFlags", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useFeatureFlags).toBeDefined();
    expect(typeof useFeatureFlags).toBe("function");
  });

  it("returns empty flags object when loading", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue(null);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as UseQueryResult<Record<string, boolean>, Error>);

    const result = useFeatureFlags();
    expect(result.flags).toEqual({});
    expect(result.isLoading).toBe(true);
  });

  it("returns flags when loaded", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    // Mock the new response format with FeatureFlagDetail objects
    const mockFlagDetails = {
      newFeature: { enabled: true, gated_by_plan: false },
      betaFeature: { enabled: false, gated_by_plan: false },
    };
    const expectedFlags = { newFeature: true, betaFeature: false };

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: mockFlagDetails,
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<typeof mockFlagDetails, Error>);

    const result = useFeatureFlags();
    expect(result.flags).toEqual(expectedFlags);
    expect(result.flagDetails).toEqual(mockFlagDetails);
    expect(result.isLoading).toBe(false);
  });

  it("provides refetch function that invalidates queries", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: {},
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<Record<string, boolean>, Error>);

    const result = useFeatureFlags();
    result.refetch();

    expect(mockInvalidateQueries).toHaveBeenCalledWith({
      queryKey: ["featureFlags", "user"],
    });
  });

  it("is disabled when user is not authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue(null);
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as UseQueryResult<Record<string, boolean>, Error>);

    useFeatureFlags();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });
});

describe("useFeatureFlag", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useFeatureFlag).toBeDefined();
    expect(typeof useFeatureFlag).toBe("function");
  });

  it("returns default value when loading", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as UseQueryResult<Record<string, boolean>, Error>);

    const result = useFeatureFlag("newFeature", true);
    expect(result.enabled).toBe(true);
    expect(result.isLoading).toBe(true);
  });

  it("returns flag value when loaded", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    // Mock the new response format with FeatureFlagDetail objects
    const mockFlagDetails = {
      newFeature: { enabled: true, gated_by_plan: false },
      betaFeature: { enabled: false, gated_by_plan: false },
    };

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: mockFlagDetails,
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<typeof mockFlagDetails, Error>);

    const result = useFeatureFlag("newFeature");
    expect(result.enabled).toBe(true);
    expect(result.isLoading).toBe(false);
    expect(result.gatedByPlan).toBe(false);
  });

  it("returns default value (false) for non-existent flag", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: {},
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<Record<string, unknown>, Error>);

    const result = useFeatureFlag("nonExistent");
    expect(result.enabled).toBe(false);
    expect(result.gatedByPlan).toBe(false);
    expect(result.requiredPlan).toBeUndefined();
  });

  it("returns gating info when feature is gated by plan", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    // Mock a feature that's gated by plan
    const mockFlagDetails = {
      premiumFeature: { enabled: false, gated_by_plan: true, required_plan: "pro" },
    };

    vi.mocked(useAuthStore).mockReturnValue({ id: 1 });
    vi.mocked(useQuery).mockReturnValue({
      data: mockFlagDetails,
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<typeof mockFlagDetails, Error>);

    const result = useFeatureFlag("premiumFeature");
    expect(result.enabled).toBe(false);
    expect(result.gatedByPlan).toBe(true);
    expect(result.requiredPlan).toBe("pro");
  });
});

describe("useInvalidateFeatureFlags", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useInvalidateFeatureFlags).toBeDefined();
    expect(typeof useInvalidateFeatureFlags).toBe("function");
  });

  it("returns a function that invalidates feature flags queries", () => {
    const invalidate = useInvalidateFeatureFlags();

    expect(typeof invalidate).toBe("function");

    invalidate();

    expect(mockInvalidateQueries).toHaveBeenCalledWith({
      queryKey: ["featureFlags"],
    });
  });
});

describe("useClearFeatureFlags", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useClearFeatureFlags).toBeDefined();
    expect(typeof useClearFeatureFlags).toBe("function");
  });

  it("returns a function that removes feature flags queries", () => {
    const clear = useClearFeatureFlags();

    expect(typeof clear).toBe("function");

    clear();

    expect(mockRemoveQueries).toHaveBeenCalledWith({
      queryKey: ["featureFlags"],
    });
  });
});
