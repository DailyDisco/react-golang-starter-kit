import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { logger } from "../../lib/logger";
import { BillingService } from "../../services/billing/billingService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to create a checkout session and redirect to Stripe
 */
export function useCreateCheckout() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async (priceId: string) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required");
      }
      return BillingService.redirectToCheckout(priceId);
    },
    onError: (error: Error) => {
      logger.error("Checkout creation error", error);
      toast.error("Failed to start checkout", {
        description: error.message || "Please try again later",
      });
    },
  });
}

/**
 * Hook to create a billing portal session and redirect
 */
export function useCreatePortalSession() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async () => {
      if (!isAuthenticated) {
        throw new Error("Authentication required");
      }
      return BillingService.redirectToPortal();
    },
    onError: (error: Error) => {
      logger.error("Portal session creation error", error);
      toast.error("Failed to open billing portal", {
        description: error.message || "Please try again later",
      });
    },
  });
}

/**
 * Hook to refresh subscription data after checkout
 */
export function useRefreshSubscription() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await queryClient.invalidateQueries({ queryKey: ["billing", "subscription"] });
      await queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
    },
    onSuccess: () => {
      toast.success("Subscription updated!");
    },
  });
}
