import type { UseQueryResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CACHE_TIMES } from "../../lib/cache-config";
import { useBillingConfig, useBillingPlans, useHasActiveSubscription, useSubscription } from "./use-billing";

// Mock dependencies
vi.mock("@tanstack/react-query", () => ({
  useQuery: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("../../lib/logger", () => ({
  logger: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("../../services/billing/billingService", () => ({
  BillingService: {
    getConfig: vi.fn(),
    getPlans: vi.fn(),
    getSubscription: vi.fn(),
    isSubscriptionActive: vi.fn(),
  },
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

describe("useBillingConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useBillingConfig).toBeDefined();
    expect(typeof useBillingConfig).toBe("function");
  });

  it("can be mocked to return config data", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockConfig = { publishableKey: "pk_test_123" };
    vi.mocked(useQuery).mockReturnValue({
      data: mockConfig,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<typeof mockConfig, Error>);

    const result = useBillingConfig();
    expect(result.data).toEqual(mockConfig);
    expect(result.isSuccess).toBe(true);
  });

  it("has correct stale time configuration", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<unknown, Error>);

    useBillingConfig();

    // Uses SWR_CONFIG.STABLE which has staleTime: Infinity
    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        staleTime: Infinity,
        retry: 1,
      })
    );
  });
});

describe("useBillingPlans", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useBillingPlans).toBeDefined();
    expect(typeof useBillingPlans).toBe("function");
  });

  it("can be mocked to return plans data", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    const mockPlans = [
      { id: "basic", name: "Basic", price: 10 },
      { id: "pro", name: "Pro", price: 20 },
    ];

    vi.mocked(useQuery).mockReturnValue({
      data: mockPlans,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<typeof mockPlans, Error>);

    const result = useBillingPlans();
    expect(result.data).toEqual(mockPlans);
  });

  it("has correct stale time configuration", async () => {
    const { useQuery } = await import("@tanstack/react-query");

    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<unknown, Error>);

    useBillingPlans();

    // Uses SWR_CONFIG.STABLE which has staleTime: Infinity
    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        staleTime: Infinity,
        retry: 1,
      })
    );
  });
});

describe("useSubscription", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useSubscription).toBeDefined();
    expect(typeof useSubscription).toBe("function");
  });

  it("is enabled when user is authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as UseQueryResult<unknown, Error>);

    useSubscription();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: true,
      })
    );
  });

  it("is disabled when user is not authenticated", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: false });
    vi.mocked(useQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as UseQueryResult<unknown, Error>);

    useSubscription();

    expect(useQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        enabled: false,
      })
    );
  });
});

describe("useHasActiveSubscription", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useHasActiveSubscription).toBeDefined();
    expect(typeof useHasActiveSubscription).toBe("function");
  });

  it("returns hasActiveSubscription based on BillingService check", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");
    const { BillingService } = await import("../../services/billing/billingService");

    const mockSubscription = { status: "active", plan: "pro" };

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: mockSubscription,
      isLoading: false,
      isError: false,
      isSuccess: true,
    } as unknown as UseQueryResult<typeof mockSubscription, Error>);
    vi.mocked(BillingService.isSubscriptionActive).mockReturnValue(true);

    const result = useHasActiveSubscription();
    expect(result.hasActiveSubscription).toBe(true);
    expect(result.isLoading).toBe(false);
    expect(result.subscription).toEqual(mockSubscription);
  });

  it("returns false when subscription is not active", async () => {
    const { useQuery } = await import("@tanstack/react-query");
    const { useAuthStore } = await import("../../stores/auth-store");
    const { BillingService } = await import("../../services/billing/billingService");

    vi.mocked(useAuthStore).mockReturnValue({ isAuthenticated: true });
    vi.mocked(useQuery).mockReturnValue({
      data: null,
      isLoading: false,
      isSuccess: true,
    } as unknown as UseQueryResult<null, Error>);
    vi.mocked(BillingService.isSubscriptionActive).mockReturnValue(false);

    const result = useHasActiveSubscription();
    expect(result.hasActiveSubscription).toBe(false);
  });
});
