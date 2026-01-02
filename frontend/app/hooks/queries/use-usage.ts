import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { CACHE_TIMES } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { UsageService } from "../../services/usage/usageService";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to fetch current billing period usage summary
 */
export function useCurrentUsage() {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.usage.current(),
    queryFn: () => UsageService.getCurrentUsage(),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.USAGE,
    refetchInterval: 60000, // Refetch every minute to keep usage current
  });
}

/**
 * Hook to fetch usage history for past billing periods
 */
export function useUsageHistory(months: number = 6) {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.usage.history(months),
    queryFn: () => UsageService.getUsageHistory(months),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.USAGE * 5, // History is less volatile
  });
}

/**
 * Hook to fetch unacknowledged usage alerts
 */
export function useUsageAlerts() {
  const { isAuthenticated } = useAuthStore();

  return useQuery({
    queryKey: queryKeys.usage.alerts(),
    queryFn: () => UsageService.getAlerts(),
    enabled: isAuthenticated,
    staleTime: CACHE_TIMES.USAGE,
  });
}

/**
 * Hook to acknowledge a usage alert
 */
export function useAcknowledgeAlert() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (alertId: number) => UsageService.acknowledgeAlert(alertId),
    onSuccess: () => {
      // Invalidate alerts query to refresh the list
      queryClient.invalidateQueries({ queryKey: queryKeys.usage.alerts() });
    },
  });
}
