/**
 * Prefetch utilities for TanStack Query
 *
 * These utilities allow prefetching data before navigation to improve UX.
 * Use them on hover, focus, or when you anticipate user navigation.
 */
import { useCallback } from "react";

import { useQueryClient } from "@tanstack/react-query";

import { AuthService, FileService } from "../services";
import { OrganizationService } from "../services/organizations/organizationService";
import { SettingsService } from "../services/settings/settingsService";
import { CACHE_TIMES } from "./cache-config";
import { queryKeys } from "./query-keys";

/**
 * Prefetch user profile data
 */
export function usePrefetchUser() {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.auth.user,
      queryFn: () => AuthService.getCurrentUser(),
      staleTime: CACHE_TIMES.USER_DATA,
    });
  }, [queryClient]);
}

/**
 * Prefetch user preferences
 */
export function usePrefetchPreferences() {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.settings.preferences(),
      queryFn: () => SettingsService.getPreferences(),
      staleTime: CACHE_TIMES.PREFERENCES,
    });
  }, [queryClient]);
}

/**
 * Prefetch user sessions
 */
export function usePrefetchSessions() {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.settings.sessions(),
      queryFn: () => SettingsService.getSessions(),
      staleTime: CACHE_TIMES.SESSIONS,
    });
  }, [queryClient]);
}

/**
 * Prefetch API keys
 */
export function usePrefetchAPIKeys() {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.settings.apiKeys(),
      queryFn: () => SettingsService.getAPIKeys(),
      staleTime: CACHE_TIMES.API_KEYS,
    });
  }, [queryClient]);
}

/**
 * Prefetch login history
 */
export function usePrefetchLoginHistory() {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.settings.loginHistory(),
      queryFn: () => SettingsService.getLoginHistory(50),
      staleTime: CACHE_TIMES.LOGIN_HISTORY,
    });
  }, [queryClient]);
}

/**
 * Prefetch files list
 */
export function usePrefetchFiles(limit?: number, offset?: number) {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.files.list(limit, offset),
      queryFn: () => FileService.fetchFiles(limit, offset),
      staleTime: CACHE_TIMES.FILES,
    });
  }, [queryClient, limit, offset]);
}

/**
 * Prefetch organization members
 */
export function usePrefetchOrgMembers(orgSlug: string) {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.organizations.members(orgSlug),
      queryFn: () => OrganizationService.listMembers(orgSlug),
      staleTime: CACHE_TIMES.ORGANIZATIONS,
    });
  }, [queryClient, orgSlug]);
}

/**
 * Prefetch organization invitations
 */
export function usePrefetchOrgInvitations(orgSlug: string) {
  const queryClient = useQueryClient();

  return useCallback(() => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.organizations.invitations(orgSlug),
      queryFn: () => OrganizationService.listInvitations(orgSlug),
      staleTime: CACHE_TIMES.ORGANIZATIONS,
    });
  }, [queryClient, orgSlug]);
}

/**
 * Prefetch all settings data for the settings page
 * Useful when user hovers over settings navigation
 */
export function usePrefetchSettingsData() {
  const prefetchPreferences = usePrefetchPreferences();
  const prefetchSessions = usePrefetchSessions();
  const prefetchAPIKeys = usePrefetchAPIKeys();
  const prefetchLoginHistory = usePrefetchLoginHistory();

  return useCallback(() => {
    prefetchPreferences();
    prefetchSessions();
    prefetchAPIKeys();
    prefetchLoginHistory();
  }, [prefetchPreferences, prefetchSessions, prefetchAPIKeys, prefetchLoginHistory]);
}

/**
 * Prefetch organization data for team page
 */
export function usePrefetchOrgData(orgSlug: string) {
  const prefetchMembers = usePrefetchOrgMembers(orgSlug);
  const prefetchInvitations = usePrefetchOrgInvitations(orgSlug);

  return useCallback(() => {
    prefetchMembers();
    prefetchInvitations();
  }, [prefetchMembers, prefetchInvitations]);
}
