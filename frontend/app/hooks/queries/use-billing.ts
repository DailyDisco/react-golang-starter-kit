import { useQuery } from "@tanstack/react-query";

import { CACHE_TIMES, GC_TIMES, SWR_CONFIG } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { BillingService } from "../../services/billing/billingService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to fetch billing configuration (publishable key).
 * Uses stale-while-revalidate pattern since billing config almost never changes.
 * Shows cached data immediately while refreshing in background.
 */
export function useBillingConfig() {
  return useQuery({
    queryKey: queryKeys.billing.config(),
    queryFn: () => BillingService.getConfig(),
    ...SWR_CONFIG.STABLE,
    gcTime: GC_TIMES.BILLING,
    retry: 1,
  });
}

/**
 * Hook to fetch available subscription plans.
 * Uses stale-while-revalidate pattern since plans rarely change.
 * Shows cached data immediately while refreshing in background.
 */
export function useBillingPlans() {
  return useQuery({
    queryKey: queryKeys.billing.plans(),
    queryFn: () => BillingService.getPlans(),
    ...SWR_CONFIG.STABLE,
    gcTime: GC_TIMES.BILLING,
    retry: 1,
  });
}

/**
 * Hook to fetch current user's subscription
 */
export function useSubscription() {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.billing.subscription(),
    queryFn: () => BillingService.getSubscription(),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.SUBSCRIPTION,
    retry: 1,
  });
}

/**
 * Hook to check if user has an active subscription
 */
export function useHasActiveSubscription() {
  const { data: subscription, isLoading } = useSubscription();

  return {
    hasActiveSubscription: BillingService.isSubscriptionActive(subscription ?? null),
    isLoading,
    subscription,
  };
}
