import { useQuery, useQueryClient } from "@tanstack/react-query";

import { GC_TIMES, SWR_CONFIG } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { FeatureFlagService } from "../../services/admin/adminService";
import type { UseFeatureFlagResult, UseFeatureFlagsResult } from "../../services/feature-flags/types";
import { useAuthStore } from "../../stores/auth-store";

/**
 * Hook to fetch all feature flags for the current user.
 * Uses stale-while-revalidate pattern since flags rarely change.
 * Shows cached data immediately while refreshing in background.
 * Real-time updates are handled via WebSocket cache_invalidate messages.
 */
export function useFeatureFlags(): UseFeatureFlagsResult {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = !!user;
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: queryKeys.featureFlags.user(),
    queryFn: () => FeatureFlagService.getFlags(),
    enabled: isAuthenticated,
    ...SWR_CONFIG.STABLE,
    gcTime: GC_TIMES.FEATURE_FLAGS,
    placeholderData: {},
  });

  return {
    flags: query.data ?? {},
    isLoading: query.isLoading,
    isError: query.isError,
    refetch: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.user() });
    },
  };
}

/**
 * Hook to check if a specific feature flag is enabled
 * @param flagKey - The feature flag key to check
 * @param defaultValue - Default value while loading (defaults to false)
 */
export function useFeatureFlag(flagKey: string, defaultValue: boolean = false): UseFeatureFlagResult {
  const { flags, isLoading } = useFeatureFlags();

  return {
    enabled: isLoading ? defaultValue : (flags[flagKey] ?? defaultValue),
    isLoading,
  };
}

/**
 * Hook to invalidate feature flags cache
 * Useful for forcing a refresh after admin changes
 */
export function useInvalidateFeatureFlags(): () => void {
  const queryClient = useQueryClient();

  return () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
  };
}

/**
 * Hook to clear feature flags from cache (on logout)
 */
export function useClearFeatureFlags(): () => void {
  const queryClient = useQueryClient();

  return () => {
    queryClient.removeQueries({ queryKey: queryKeys.featureFlags.all });
  };
}
