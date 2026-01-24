import { useCallback } from "react";

import { useQueryClient, type QueryClient } from "@tanstack/react-query";

import { handleCacheInvalidation, type CacheInvalidatePayload } from "../../lib/cache-invalidation";
import { queryKeys } from "../../lib/query-keys";
import type {
  FeatureFlagUpdatePayload,
  MemberUpdatePayload,
  OrgUpdatePayload,
  SubscriptionUpdatePayload,
  UsageAlertPayload,
  UserUpdatePayload,
} from "./types";

interface UseWebSocketInvalidationReturn {
  /** Invalidate queries based on user update */
  invalidateUserUpdate: (payload: UserUpdatePayload) => void;
  /** Invalidate queries based on cache invalidate message */
  invalidateCacheMessage: (payload: CacheInvalidatePayload) => void;
  /** Invalidate queries based on usage alert */
  invalidateUsageQueries: () => void;
  /** Invalidate queries based on subscription update */
  invalidateSubscriptionQueries: () => void;
  /** Invalidate queries based on org update */
  invalidateOrgQueries: (payload: OrgUpdatePayload) => void;
  /** Invalidate queries based on member update */
  invalidateMemberQueries: (payload: MemberUpdatePayload) => void;
  /** Invalidate queries based on feature flag update */
  invalidateFeatureFlagQueries: () => void;
  /** Invalidate notification center queries */
  invalidateNotificationQueries: () => void;
}

/**
 * Hook for managing TanStack Query cache invalidation based on WebSocket messages.
 * Provides granular control over which queries to invalidate for each message type.
 */
export function useWebSocketInvalidation(): UseWebSocketInvalidationReturn {
  const queryClient = useQueryClient();

  const invalidateUserUpdate = useCallback(
    (payload: UserUpdatePayload) => {
      if (payload.field === "profile" || payload.field === "role") {
        queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
      } else if (payload.field === "preferences") {
        queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences() });
      } else if (payload.field === "sessions") {
        queryClient.invalidateQueries({ queryKey: queryKeys.settings.sessions() });
      }
    },
    [queryClient]
  );

  const invalidateCacheMessage = useCallback(
    (payload: CacheInvalidatePayload) => {
      handleCacheInvalidation(queryClient, payload);
    },
    [queryClient]
  );

  const invalidateUsageQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.usage.all });
  }, [queryClient]);

  const invalidateSubscriptionQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.billing.subscription() });
    queryClient.invalidateQueries({ queryKey: queryKeys.billing.all });
    queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    queryClient.invalidateQueries({ queryKey: queryKeys.usage.all });
  }, [queryClient]);

  const invalidateOrgQueries = useCallback(
    (payload: OrgUpdatePayload) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      if (payload.orgSlug) {
        queryClient.invalidateQueries({ queryKey: queryKeys.organizations.detail(payload.orgSlug) });

        if (payload.event === "billing_changed") {
          queryClient.invalidateQueries({ queryKey: queryKeys.organizations.billing(payload.orgSlug) });
        }
      }
    },
    [queryClient]
  );

  const invalidateMemberQueries = useCallback(
    (payload: MemberUpdatePayload) => {
      if (payload.orgSlug) {
        queryClient.invalidateQueries({ queryKey: queryKeys.organizations.members(payload.orgSlug) });

        if (payload.event === "invitation_sent" || payload.event === "invitation_revoked") {
          queryClient.invalidateQueries({ queryKey: queryKeys.organizations.invitations(payload.orgSlug) });
        }
      }

      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
    },
    [queryClient]
  );

  const invalidateFeatureFlagQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
  }, [queryClient]);

  const invalidateNotificationQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
  }, [queryClient]);

  return {
    invalidateUserUpdate,
    invalidateCacheMessage,
    invalidateUsageQueries,
    invalidateSubscriptionQueries,
    invalidateOrgQueries,
    invalidateMemberQueries,
    invalidateFeatureFlagQueries,
    invalidateNotificationQueries,
  };
}

/**
 * Standalone function for cache invalidation (useful for testing or non-hook contexts)
 */
export function invalidateQueriesForMessage(queryClient: QueryClient, messageType: string, payload: unknown): void {
  switch (messageType) {
    case "user_update": {
      const p = payload as UserUpdatePayload;
      if (p.field === "profile" || p.field === "role") {
        queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
      } else if (p.field === "preferences") {
        queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences() });
      } else if (p.field === "sessions") {
        queryClient.invalidateQueries({ queryKey: queryKeys.settings.sessions() });
      }
      break;
    }
    case "usage_alert":
      queryClient.invalidateQueries({ queryKey: queryKeys.usage.all });
      break;
    case "subscription_update":
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.subscription() });
      queryClient.invalidateQueries({ queryKey: queryKeys.billing.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
      queryClient.invalidateQueries({ queryKey: queryKeys.usage.all });
      break;
    case "org_update": {
      const p = payload as OrgUpdatePayload;
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      if (p.orgSlug) {
        queryClient.invalidateQueries({ queryKey: queryKeys.organizations.detail(p.orgSlug) });
        if (p.event === "billing_changed") {
          queryClient.invalidateQueries({ queryKey: queryKeys.organizations.billing(p.orgSlug) });
        }
      }
      break;
    }
    case "member_update": {
      const p = payload as MemberUpdatePayload;
      if (p.orgSlug) {
        queryClient.invalidateQueries({ queryKey: queryKeys.organizations.members(p.orgSlug) });
        if (p.event === "invitation_sent" || p.event === "invitation_revoked") {
          queryClient.invalidateQueries({ queryKey: queryKeys.organizations.invitations(p.orgSlug) });
        }
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.all });
      break;
    }
    case "feature_flag_update":
      queryClient.invalidateQueries({ queryKey: queryKeys.featureFlags.all });
      break;
  }
}
