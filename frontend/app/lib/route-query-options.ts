/**
 * Query options for use in TanStack Router loaders
 *
 * These options can be used with queryClient.ensureQueryData() in route loaders
 * to prefetch data before the route component renders.
 *
 * Usage in a route:
 * ```typescript
 * export const Route = createFileRoute('/settings/preferences')({
 *   loader: async ({ context }) => {
 *     await context.queryClient.ensureQueryData(preferencesQueryOptions());
 *   },
 *   component: PreferencesPage,
 * });
 * ```
 */

import { queryOptions } from "@tanstack/react-query";

import { AuthService, type User } from "../services";
import { AdminService, type AdminStats } from "../services/admin/adminService";
import { BillingService } from "../services/billing/billingService";
import { FileService } from "../services/files/fileService";
import { OrganizationService, type OrganizationMember } from "../services/organizations/organizationService";
import { SettingsService } from "../services/settings/settingsService";
import { CACHE_TIMES, GC_TIMES, SWR_CONFIG } from "./cache-config";
import { queryKeys } from "./query-keys";

/**
 * Query options for fetching the current authenticated user
 */
export const currentUserQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.auth.user,
    queryFn: () => AuthService.getCurrentUser(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false, // Don't retry auth failures
  });

/**
 * Query options for admin dashboard stats
 */
export const adminStatsQueryOptions = () =>
  queryOptions<AdminStats>({
    queryKey: ["admin", "stats"],
    queryFn: () => AdminService.getStats(),
    staleTime: 30 * 1000, // 30 seconds - stats change frequently
  });

/**
 * Query options for admin feature flags
 */
export const adminFeatureFlagsQueryOptions = () =>
  queryOptions({
    queryKey: ["admin", "feature-flags"],
    queryFn: () => AdminService.getFeatureFlags(),
    staleTime: 60 * 1000, // 1 minute
  });

/**
 * Query options for admin audit logs
 */
export const adminAuditLogsQueryOptions = (filter?: {
  page?: number;
  limit?: number;
  action?: string;
  target_type?: string;
}) =>
  queryOptions({
    queryKey: ["admin", "audit-logs", filter],
    queryFn: () => AdminService.getAuditLogs(filter),
    staleTime: 30 * 1000, // 30 seconds
  });

/**
 * Query options for organization members
 */
export const orgMembersQueryOptions = (orgSlug: string) =>
  queryOptions<OrganizationMember[]>({
    queryKey: queryKeys.organizations.members(orgSlug),
    queryFn: () => OrganizationService.listMembers(orgSlug),
    staleTime: 60 * 1000, // 1 minute
  });

/**
 * Query options for organization invitations
 */
export const orgInvitationsQueryOptions = (orgSlug: string) =>
  queryOptions({
    queryKey: queryKeys.organizations.invitations(orgSlug),
    queryFn: () => OrganizationService.listInvitations(orgSlug),
    staleTime: 60 * 1000, // 1 minute
  });

// ============================================================================
// Settings Query Options
// ============================================================================

/**
 * Query options for user preferences
 */
export const preferencesQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.settings.preferences(),
    queryFn: () => SettingsService.getPreferences(),
    staleTime: CACHE_TIMES.PREFERENCES,
  });

/**
 * Query options for active sessions
 */
export const sessionsQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.settings.sessions(),
    queryFn: () => SettingsService.getSessions(),
    staleTime: CACHE_TIMES.SESSIONS,
  });

/**
 * Query options for API keys
 */
export const apiKeysQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.settings.apiKeys(),
    queryFn: () => SettingsService.getAPIKeys(),
    staleTime: CACHE_TIMES.API_KEYS,
  });

/**
 * Query options for login history
 */
export const loginHistoryQueryOptions = (limit: number = 50) =>
  queryOptions({
    queryKey: queryKeys.settings.loginHistory(),
    queryFn: () => SettingsService.getLoginHistory(limit),
    staleTime: CACHE_TIMES.LOGIN_HISTORY,
  });

/**
 * Query options for connected OAuth accounts
 */
export const connectedAccountsQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.settings.connectedAccounts(),
    queryFn: () => SettingsService.getConnectedAccounts(),
    staleTime: CACHE_TIMES.PREFERENCES,
  });

// ============================================================================
// Files Query Options
// ============================================================================

/**
 * Query options for files list
 */
export const filesQueryOptions = (limit?: number, offset?: number) =>
  queryOptions({
    queryKey: queryKeys.files.list(limit, offset),
    queryFn: () => FileService.fetchFiles(limit, offset),
    staleTime: CACHE_TIMES.FILES,
  });

/**
 * Query options for storage status
 */
export const storageStatusQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.files.storageStatus(),
    queryFn: () => FileService.getStorageStatus(),
    staleTime: CACHE_TIMES.STORAGE_STATUS,
  });

// ============================================================================
// Billing Query Options (with SWR pattern)
// ============================================================================

/**
 * Query options for billing configuration.
 * Uses SWR pattern since this almost never changes.
 */
export const billingConfigQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.billing.config(),
    queryFn: () => BillingService.getConfig(),
    ...SWR_CONFIG.STABLE,
    gcTime: GC_TIMES.BILLING,
  });

/**
 * Query options for subscription plans.
 * Uses SWR pattern since plans rarely change.
 */
export const billingPlansQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.billing.plans(),
    queryFn: () => BillingService.getPlans(),
    ...SWR_CONFIG.STABLE,
    gcTime: GC_TIMES.BILLING,
  });

/**
 * Query options for user's subscription
 */
export const subscriptionQueryOptions = () =>
  queryOptions({
    queryKey: queryKeys.billing.subscription(),
    queryFn: () => BillingService.getSubscription(),
    staleTime: CACHE_TIMES.SUBSCRIPTION,
  });
