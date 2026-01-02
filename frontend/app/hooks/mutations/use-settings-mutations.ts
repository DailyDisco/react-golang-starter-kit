import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

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
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name?: string; email?: string; bio?: string; location?: string; social_links?: string }) =>
      SettingsService.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useUploadAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (file: File) => SettingsService.uploadAvatar(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useDeleteAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => SettingsService.deleteAvatar(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

// ============================================================================
// Password & Security Mutations
// ============================================================================

export function useChangePassword() {
  return useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) => SettingsService.changePassword(data),
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useSetup2FA() {
  return useMutation({
    mutationFn: () => SettingsService.setup2FA(),
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useVerify2FA() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (code: string) => SettingsService.verify2FA(code),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useDisable2FA() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (code: string) => SettingsService.disable2FA(code),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

// ============================================================================
// Session Mutations
// ============================================================================

export function useRevokeSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (sessionId: number) => SettingsService.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.sessions() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useRevokeAllSessions() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => SettingsService.revokeAllSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.sessions() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

// ============================================================================
// Preferences Mutations
// ============================================================================

export function useUpdatePreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdatePreferencesRequest) => SettingsService.updatePreferences(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useUpdateNotifications() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: EmailNotificationSettings) => SettingsService.updatePreferences({ email_notifications: data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

// ============================================================================
// API Key Mutations
// ============================================================================

export function useCreateAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: CreateAPIKeyRequest) => SettingsService.createAPIKey(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useDeleteAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => SettingsService.deleteAPIKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useUpdateAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: { is_active?: boolean } }) => SettingsService.updateAPIKey(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}

export function useTestAPIKey() {
  return useMutation({
    mutationFn: (id: number) => SettingsService.testAPIKey(id),
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });
}
