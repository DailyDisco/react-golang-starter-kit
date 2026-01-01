import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { CACHE_TIMES } from "../../lib/cache-config";
import { logger } from "../../lib/logger";
import { BillingService } from "../../services/billing/billingService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to fetch billing configuration (publishable key)
 */
export function useBillingConfig() {
  return useQuery({
    queryKey: ["billing", "config"],
    queryFn: () => BillingService.getConfig(),
    staleTime: CACHE_TIMES.BILLING_CONFIG,
    retry: 1,
  });
}

/**
 * Hook to fetch available subscription plans
 */
export function useBillingPlans() {
  return useQuery({
    queryKey: ["billing", "plans"],
    queryFn: () => BillingService.getPlans(),
    staleTime: CACHE_TIMES.BILLING_PLANS,
    retry: 1,
  });
}

/**
 * Hook to fetch current user's subscription
 */
export function useSubscription() {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: ["billing", "subscription"],
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
