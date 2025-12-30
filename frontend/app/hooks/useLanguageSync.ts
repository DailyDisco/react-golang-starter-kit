import { useEffect, useRef } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { supportedLanguages, type SupportedLanguage } from "../i18n";
import { queryKeys } from "../lib/query-keys";
import { SettingsService } from "../services/settings/settingsService";
import { useAuthStore } from "../stores/auth-store";
import { useLanguageStore } from "../stores/language-store";

/**
 * Hook to sync language preference between frontend and backend.
 * - On login: syncs language from backend to frontend
 * - On language change: updates backend (debounced)
 */
export function useLanguageSync() {
  const queryClient = useQueryClient();
  const { language, syncFromBackend, isInitialized } = useLanguageStore();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const previousLanguageRef = useRef(language);
  const hasInitialSyncedRef = useRef(false);

  // Fetch user preferences when authenticated
  const { data: preferences } = useQuery({
    queryKey: queryKeys.settings.preferences,
    queryFn: () => SettingsService.getPreferences(),
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Mutation to update language preference
  const { mutate: updateLanguage, isPending: isUpdating } = useMutation({
    mutationFn: (newLanguage: string) => SettingsService.updatePreferences({ language: newLanguage }),
    onSuccess: () => {
      // Invalidate preferences cache
      void queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences });
    },
  });

  // Sync language from backend when user logs in (backend takes precedence)
  useEffect(() => {
    if (isAuthenticated && preferences?.language && isInitialized && !hasInitialSyncedRef.current) {
      const backendLang = preferences.language;
      if (supportedLanguages.includes(backendLang as SupportedLanguage)) {
        syncFromBackend(backendLang);
        previousLanguageRef.current = backendLang as SupportedLanguage;
        hasInitialSyncedRef.current = true;
      }
    }
  }, [isAuthenticated, preferences?.language, isInitialized, syncFromBackend]);

  // Reset initial sync flag when user logs out
  useEffect(() => {
    if (!isAuthenticated) {
      hasInitialSyncedRef.current = false;
    }
  }, [isAuthenticated]);

  // Update backend when language changes (only after initial sync)
  useEffect(() => {
    if (isAuthenticated && isInitialized && hasInitialSyncedRef.current && language !== previousLanguageRef.current) {
      previousLanguageRef.current = language;
      updateLanguage(language);
    }
  }, [language, isAuthenticated, isInitialized, updateLanguage]);

  return {
    language,
    isLoading: isUpdating,
  };
}
