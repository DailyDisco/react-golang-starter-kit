// Mutation hooks - TanStack Query wrappers for data mutations

// Auth
export { useLogin, useRegister } from "./use-auth-mutations";

// Users
export { useCreateUser, useDeleteUser, useUpdateUser } from "./use-user-mutations";

// Files
export { useFileDelete, useFileUpload } from "./use-file-mutations";

// Billing
export { useCreateCheckout, useCreatePortalSession, useRefreshSubscription } from "./use-billing-mutations";

// AI
export { useAIChat, useAIChatAdvanced, useAIAnalyzeImage, useAIEmbeddings } from "./use-ai-mutations";

// Settings
export {
  useChangePassword,
  useCreateAPIKey,
  useDeleteAPIKey,
  useDeleteAvatar,
  useDisable2FA,
  useRequestDataExport,
  useRevokeAllSessions,
  useRevokeSession,
  useSetup2FA,
  useTestAPIKey,
  useUpdateAPIKey,
  useUpdateNotifications,
  useUpdatePreferences,
  useUpdateProfile,
  useUploadAvatar,
  useVerify2FA,
} from "./use-settings-mutations";

// Organizations
export {
  useCancelInvitation,
  useDeleteOrganization,
  useInviteMember,
  useLeaveOrganization,
  useRemoveMember,
  useUpdateMemberRole,
  useUpdateOrganization,
} from "./use-org-mutations";
