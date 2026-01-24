import { useQuery, useQueryClient } from "@tanstack/react-query";

import { GC_TIMES, SWR_CONFIG } from "../../lib/cache-config";
import { queryKeys } from "../../lib/query-keys";
import { FeatureFlagService } from "../../services/admin/adminService";
import type {
  FeatureFlagDetail,
  UseFeatureFlagResult,
  UseFeatureFlagsResult,
  UserFeatureFlagsResponse,
} from "../../services/feature-flags/types";
import { useAuthStore } from "../../stores/auth-store";

/** Empty placeholder for when data is not yet loaded */
const EMPTY_FLAG_DETAILS: UserFeatureFlagsResponse = {};

/**
 * Hook to fetch all feature flags for the current user.
 * Uses stale-while-revalidate pattern since flags rarely change.
 * Shows cached data immediately while refreshing in background.
 * Real-time updates are handled via WebSocket cache_invalidate messages.
 *
 * Returns both simple boolean flags (backward compatible) and detailed
 * flag information with plan gating metadata.
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
    placeholderData: EMPTY_FLAG_DETAILS,
  });

  const flagDetails: Record<string, FeatureFlagDetail> = query.data ?? EMPTY_FLAG_DETAILS;

  // Transform to simple boolean map for backward compatibility
  const flags: Record<string, boolean> = {};
  for (const [key, detail] of Object.entries(flagDetails)) {
    flags[key] = detail.enabled;
  }

  return {
    flags,
    flagDetails,
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
 *
 * Returns enabled status along with plan gating information for UI prompts.
 */
export function useFeatureFlag(flagKey: string, defaultValue: boolean = false): UseFeatureFlagResult {
  const { flags, flagDetails, isLoading } = useFeatureFlags();

  const detail = flagDetails[flagKey];

  return {
    enabled: isLoading ? defaultValue : (flags[flagKey] ?? defaultValue),
    isLoading,
    gatedByPlan: detail?.gated_by_plan ?? false,
    requiredPlan: detail?.required_plan,
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
