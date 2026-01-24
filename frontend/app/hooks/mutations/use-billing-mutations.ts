import { useMutation, useQueryClient } from "@tanstack/react-query";

import { logger } from "../../lib/logger";
import { showMutationError, showMutationSuccess } from "../../lib/mutation-toast";
import { queryKeys } from "../../lib/query-keys";
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
      showMutationError({ error, context: "Failed to start checkout" });
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
      showMutationError({ error, context: "Failed to open billing portal" });
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
      await queryClient.invalidateQueries({ queryKey: queryKeys.billing.subscription() });
      await queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onSuccess: () => {
      showMutationSuccess({ message: "Subscription updated" });
    },
  });
}
