// Main hooks barrel export
// Provides convenient access to all hooks from a single import

// Query hooks (data fetching)
export {
  // Auth
  useCurrentUser,
  // Billing
  useBillingConfig,
  useBillingPlans,
  useHasActiveSubscription,
  useSubscription,
  // Feature Flags
  useClearFeatureFlags,
  useFeatureFlag,
  useFeatureFlags,
  useInvalidateFeatureFlags,
  // Files
  useFileDownload,
  useFiles,
  useFileUrl,
  useStorageStatus,
  // Health
  useHealthCheck,
  // Users
  useUser,
  useUsers,
} from "./queries";

// Mutation hooks (data mutations)
export {
  // Auth
  useLogin,
  useRegister,
  // Billing
  useCreateCheckout,
  useCreatePortalSession,
  useRefreshSubscription,
  // Files
  useFileDelete,
  useFileUpload,
  // Users
  useCreateUser,
  useDeleteUser,
  useUpdateUser,
} from "./mutations";

// Auth hook (combines store + mutations)
export { useAuth } from "./auth/useAuth";

// Utility hooks
export { useIsMobile } from "./use-mobile";
export { useLanguageSync } from "./useLanguageSync";
export { useOnlineStatus } from "./useOnlineStatus";
export { useWebSocket } from "./useWebSocket";
