import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

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
    staleTime: 60 * 60 * 1000, // 1 hour - config rarely changes
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
    staleTime: 5 * 60 * 1000, // 5 minutes
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
    staleTime: 60 * 1000, // 1 minute - subscription status may change
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
