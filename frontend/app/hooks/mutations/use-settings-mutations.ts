import { useCreateMutation, useCreateOptimisticMutation } from "../../lib/create-mutation";
import { queryKeys } from "../../lib/query-keys";
import {
  SettingsService,
  type CreateAPIKeyRequest,
  type EmailNotificationSettings,
  type UpdatePreferencesRequest,
} from "../../services/settings/settingsService";

// ============================================================================
// Profile Mutations
// ============================================================================

export function useUpdateProfile() {
  return useCreateMutation({
    mutationFn: (data: { name?: string; email?: string; bio?: string; location?: string; social_links?: string }) =>
      SettingsService.updateProfile(data),
    successMessage: "Profile updated",
    invalidateKeys: [queryKeys.auth.user],
  });
}

export function useUploadAvatar() {
  return useCreateMutation({
    mutationFn: (file: File) => SettingsService.uploadAvatar(file),
    successMessage: "Avatar uploaded",
    invalidateKeys: [queryKeys.auth.user],
  });
}

export function useDeleteAvatar() {
  return useCreateMutation({
    mutationFn: () => SettingsService.deleteAvatar(),
    successMessage: "Avatar removed",
    invalidateKeys: [queryKeys.auth.user],
  });
}

// ============================================================================
// Password & Security Mutations
// ============================================================================

export function useChangePassword() {
  return useCreateMutation({
    mutationFn: (data: { current_password: string; new_password: string }) => SettingsService.changePassword(data),
    successMessage: "Password changed",
  });
}

export function useSetup2FA() {
  return useCreateMutation({
    mutationFn: () => SettingsService.setup2FA(),
    showSuccessToast: false,
  });
}

export function useVerify2FA() {
  return useCreateMutation({
    mutationFn: (code: string) => SettingsService.verify2FA(code),
    successMessage: "Two-factor authentication enabled",
    invalidateKeys: [queryKeys.auth.user],
  });
}

export function useDisable2FA() {
  return useCreateMutation({
    mutationFn: (code: string) => SettingsService.disable2FA(code),
    successMessage: "Two-factor authentication disabled",
    invalidateKeys: [queryKeys.auth.user],
  });
}

// ============================================================================
// Session Mutations
// ============================================================================

export function useRevokeSession() {
  return useCreateOptimisticMutation({
    mutationFn: (sessionId: number) => SettingsService.revokeSession(sessionId),
    queryKey: queryKeys.settings.sessions(),
    optimisticUpdate: (old, sessionId) => {
      if (!old || !Array.isArray(old)) return old;
      return old.filter((session: { id: number }) => session.id !== sessionId);
    },
    successMessage: "Session revoked",
    invalidateKeys: [queryKeys.settings.sessions()],
  });
}

export function useRevokeAllSessions() {
  return useCreateMutation({
    mutationFn: () => SettingsService.revokeAllSessions(),
    successMessage: "All other sessions revoked",
    invalidateKeys: [queryKeys.settings.sessions()],
  });
}

// ============================================================================
// Preferences Mutations
// ============================================================================

export function useUpdatePreferences() {
  return useCreateOptimisticMutation({
    mutationFn: (data: UpdatePreferencesRequest) => SettingsService.updatePreferences(data),
    queryKey: queryKeys.settings.preferences(),
    optimisticUpdate: (old, newData) => {
      if (!old || typeof old !== "object") return old;
      return { ...old, ...newData };
    },
    successMessage: "Preferences saved",
    invalidateKeys: [queryKeys.settings.preferences()],
  });
}

export function useUpdateNotifications() {
  return useCreateMutation({
    mutationFn: (data: EmailNotificationSettings) => SettingsService.updatePreferences({ email_notifications: data }),
    successMessage: "Notification settings saved",
    invalidateKeys: [queryKeys.settings.preferences()],
  });
}

// ============================================================================
// API Key Mutations
// ============================================================================

export function useCreateAPIKey() {
  return useCreateMutation({
    mutationFn: (req: CreateAPIKeyRequest) => SettingsService.createAPIKey(req),
    successMessage: "API key created",
    invalidateKeys: [queryKeys.settings.apiKeys()],
  });
}

export function useDeleteAPIKey() {
  return useCreateOptimisticMutation({
    mutationFn: (id: number) => SettingsService.deleteAPIKey(id),
    queryKey: queryKeys.settings.apiKeys(),
    optimisticUpdate: (old, keyId) => {
      if (!old || !Array.isArray(old)) return old;
      return old.filter((key: { id: number }) => key.id !== keyId);
    },
    successMessage: "API key deleted",
    invalidateKeys: [queryKeys.settings.apiKeys()],
  });
}

export function useUpdateAPIKey() {
  return useCreateOptimisticMutation({
    mutationFn: ({ id, data }: { id: number; data: { is_active?: boolean } }) => SettingsService.updateAPIKey(id, data),
    queryKey: queryKeys.settings.apiKeys(),
    optimisticUpdate: (old, { id, data }) => {
      if (!old || !Array.isArray(old)) return old;
      return old.map((key: { id: number; is_active?: boolean }) => (key.id === id ? { ...key, ...data } : key));
    },
    successMessage: "API key updated",
    invalidateKeys: [queryKeys.settings.apiKeys()],
  });
}

export function useTestAPIKey() {
  return useCreateMutation({
    mutationFn: (id: number) => SettingsService.testAPIKey(id),
    successMessage: "API key is valid",
  });
}

// ============================================================================
// Data Export Mutations
// ============================================================================

export function useRequestDataExport() {
  return useCreateMutation({
    mutationFn: () => SettingsService.requestDataExport(),
    successMessage: "Data export requested",
    successDescription: "You'll receive an email when your export is ready.",
    invalidateKeys: [queryKeys.settings.dataExportStatus()],
  });
}
