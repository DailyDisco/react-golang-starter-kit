import { API_BASE_URL, authenticatedFetch, parseErrorResponse } from "../api/client";
import type {
  BillingConfig,
  BillingPlan,
  CheckoutSessionResponse,
  CreateCheckoutRequest,
  PortalSessionResponse,
  Subscription,
} from "../types";

export class BillingService {
  /**
   * Get public billing configuration (Stripe publishable key)
   */
  static async getConfig(): Promise<BillingConfig> {
    const response = await fetch(`${API_BASE_URL}/api/v1/billing/config`);
    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to get billing config");
    }

    return response.json();
  }

  /**
   * Get available subscription plans
   */
  static async getPlans(): Promise<BillingPlan[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/billing/plans`);
    if (!response.ok) {
      // If billing is not configured, return empty array
      if (response.status === 503) {
        return [];
      }
      throw await parseErrorResponse(response, "Failed to get plans");
    }

    return response.json();
  }

  /**
   * Create a checkout session for subscription purchase
   */
  static async createCheckoutSession(request: CreateCheckoutRequest): Promise<CheckoutSessionResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/billing/checkout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to create checkout session");
    }

    return response.json();
  }

  /**
   * Create a billing portal session for subscription management
   */
  static async createPortalSession(): Promise<PortalSessionResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/billing/portal`, {
      method: "POST",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to create portal session");
    }

    return response.json();
  }

  /**
   * Get current user's subscription
   */
  static async getSubscription(): Promise<Subscription | null> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/billing/subscription`);

    if (!response.ok) {
      // No subscription found is not an error
      if (response.status === 404) {
        return null;
      }
      throw await parseErrorResponse(response, "Failed to get subscription");
    }

    return response.json();
  }

  /**
   * Redirect to Stripe Checkout
   */
  static async redirectToCheckout(priceId: string): Promise<void> {
    const session = await this.createCheckoutSession({ price_id: priceId });
    if (session.url) {
      window.location.href = session.url;
    } else {
      throw new Error("No checkout URL returned");
    }
  }

  /**
   * Redirect to Stripe Customer Portal
   */
  static async redirectToPortal(): Promise<void> {
    const session = await this.createPortalSession();
    if (session.url) {
      window.location.href = session.url;
    } else {
      throw new Error("No portal URL returned");
    }
  }

  /**
   * Format price from cents to display string
   */
  static formatPrice(amount: number, currency: string): string {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency.toUpperCase(),
    }).format(amount / 100);
  }

  /**
   * Check if subscription is active
   */
  static isSubscriptionActive(subscription: Subscription | null): boolean {
    if (!subscription) return false;
    return subscription.status === "active" || subscription.status === "trialing";
  }
}
