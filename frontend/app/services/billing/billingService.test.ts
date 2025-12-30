import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, authenticatedFetch, parseErrorResponse } from "../api/client";
import type { Subscription } from "../types";
import { BillingService } from "./billingService";

// Mock the API client module
vi.mock("../api/client", () => ({
  API_BASE_URL: "http://localhost:8080",
  ApiError: class ApiError extends Error {
    code: string;
    statusCode: number;
    constructor(message: string, code: string, statusCode: number) {
      super(message);
      this.name = "ApiError";
      this.code = code;
      this.statusCode = statusCode;
    }
  },
  authenticatedFetch: vi.fn(),
  parseErrorResponse: vi.fn(),
}));

// Mock global fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock window.location
const mockLocation = { href: "" };
Object.defineProperty(window, "location", {
  value: mockLocation,
  writable: true,
});

describe("BillingService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocation.href = "";
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("getConfig", () => {
    it("should return billing configuration", async () => {
      const mockConfig = { publishable_key: "pk_test_123" };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockConfig,
      } as Response);

      const result = await BillingService.getConfig();

      expect(result).toEqual(mockConfig);
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/billing/config");
    });

    it("should throw error on failure", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Config not found", "NOT_FOUND", 500));

      await expect(BillingService.getConfig()).rejects.toThrow("Config not found");
    });
  });

  describe("getPlans", () => {
    it("should return available plans", async () => {
      const mockPlans = [
        { id: "plan_1", name: "Basic", amount: 999, currency: "usd", interval: "month" },
        { id: "plan_2", name: "Pro", amount: 1999, currency: "usd", interval: "month" },
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockPlans,
      } as Response);

      const result = await BillingService.getPlans();

      expect(result).toEqual(mockPlans);
      expect(mockFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/billing/plans");
    });

    it("should return empty array when billing is not configured (503)", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 503,
      } as Response);

      const result = await BillingService.getPlans();

      expect(result).toEqual([]);
    });

    it("should throw error on other failures", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(BillingService.getPlans()).rejects.toThrow("Server error");
    });
  });

  describe("createCheckoutSession", () => {
    it("should create checkout session", async () => {
      const mockSession = {
        session_id: "cs_test_123",
        url: "https://checkout.stripe.com/pay/cs_test_123",
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSession,
      } as Response);

      const result = await BillingService.createCheckoutSession({ price_id: "price_123" });

      expect(result).toEqual(mockSession);
      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/billing/checkout",
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ price_id: "price_123" }),
        })
      );
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid price ID", "INVALID_PRICE", 400));

      await expect(BillingService.createCheckoutSession({ price_id: "invalid" })).rejects.toThrow("Invalid price ID");
    });
  });

  describe("createPortalSession", () => {
    it("should create portal session", async () => {
      const mockSession = {
        url: "https://billing.stripe.com/session/ses_123",
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSession,
      } as Response);

      const result = await BillingService.createPortalSession();

      expect(result).toEqual(mockSession);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/billing/portal", {
        method: "POST",
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("No subscription found", "NOT_FOUND", 404));

      await expect(BillingService.createPortalSession()).rejects.toThrow("No subscription found");
    });
  });

  describe("getSubscription", () => {
    it("should return current subscription", async () => {
      const mockSubscription = {
        id: 1,
        user_id: 1,
        status: "active",
        stripe_price_id: "price_123",
        current_period_start: "2024-01-01",
        current_period_end: "2024-02-01",
        cancel_at_period_end: false,
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSubscription,
      } as Response);

      const result = await BillingService.getSubscription();

      expect(result).toEqual(mockSubscription);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/billing/subscription");
    });

    it("should return null when no subscription exists (404)", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);

      const result = await BillingService.getSubscription();

      expect(result).toBeNull();
    });

    it("should throw error on other failures", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(BillingService.getSubscription()).rejects.toThrow("Server error");
    });
  });

  describe("redirectToCheckout", () => {
    it("should redirect to checkout URL", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          session_id: "cs_test_123",
          url: "https://checkout.stripe.com/pay/cs_test_123",
        }),
      } as Response);

      await BillingService.redirectToCheckout("price_123");

      expect(mockLocation.href).toBe("https://checkout.stripe.com/pay/cs_test_123");
    });

    it("should throw error if no URL returned", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ session_id: "cs_test_123" }),
      } as Response);

      await expect(BillingService.redirectToCheckout("price_123")).rejects.toThrow("No checkout URL returned");
    });
  });

  describe("redirectToPortal", () => {
    it("should redirect to portal URL", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          url: "https://billing.stripe.com/session/ses_123",
        }),
      } as Response);

      await BillingService.redirectToPortal();

      expect(mockLocation.href).toBe("https://billing.stripe.com/session/ses_123");
    });

    it("should throw error if no URL returned", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      } as Response);

      await expect(BillingService.redirectToPortal()).rejects.toThrow("No portal URL returned");
    });
  });

  describe("formatPrice", () => {
    it("should format price from cents to USD", () => {
      expect(BillingService.formatPrice(999, "usd")).toBe("$9.99");
      expect(BillingService.formatPrice(1999, "usd")).toBe("$19.99");
      expect(BillingService.formatPrice(10000, "usd")).toBe("$100.00");
    });

    it("should format price from cents to EUR", () => {
      // Note: Intl.NumberFormat may format EUR differently depending on locale
      const result = BillingService.formatPrice(999, "eur");
      expect(result).toContain("9.99");
    });

    it("should handle zero amount", () => {
      expect(BillingService.formatPrice(0, "usd")).toBe("$0.00");
    });
  });

  describe("isSubscriptionActive", () => {
    it("should return true for active subscription", () => {
      const subscription: Subscription = {
        id: 1,
        user_id: 1,
        status: "active",
        stripe_price_id: "price_123",
        current_period_start: "2024-01-01",
        current_period_end: "2024-02-01",
        cancel_at_period_end: false,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      };

      expect(BillingService.isSubscriptionActive(subscription)).toBe(true);
    });

    it("should return true for trialing subscription", () => {
      const subscription: Subscription = {
        id: 1,
        user_id: 1,
        status: "trialing",
        stripe_price_id: "price_123",
        current_period_start: "2024-01-01",
        current_period_end: "2024-02-01",
        cancel_at_period_end: false,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      };

      expect(BillingService.isSubscriptionActive(subscription)).toBe(true);
    });

    it("should return false for canceled subscription", () => {
      const subscription: Subscription = {
        id: 1,
        user_id: 1,
        status: "canceled",
        stripe_price_id: "price_123",
        current_period_start: "2024-01-01",
        current_period_end: "2024-02-01",
        cancel_at_period_end: false,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      };

      expect(BillingService.isSubscriptionActive(subscription)).toBe(false);
    });

    it("should return false for past_due subscription", () => {
      const subscription: Subscription = {
        id: 1,
        user_id: 1,
        status: "past_due",
        stripe_price_id: "price_123",
        current_period_start: "2024-01-01",
        current_period_end: "2024-02-01",
        cancel_at_period_end: false,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      };

      expect(BillingService.isSubscriptionActive(subscription)).toBe(false);
    });

    it("should return false for null subscription", () => {
      expect(BillingService.isSubscriptionActive(null)).toBe(false);
    });
  });
});
