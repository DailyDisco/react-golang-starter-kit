// Feature flag types for the frontend

export interface UseFeatureFlagResult {
  enabled: boolean;
  isLoading: boolean;
}

export interface UseFeatureFlagsResult {
  flags: Record<string, boolean>;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
}

/**
 * Common feature flag keys - type-safe constants
 * Use these to avoid typos and enable autocomplete
 */
export const FeatureFlagKeys = {
  // UI Features
  DARK_MODE: "dark_mode",
  NEW_DASHBOARD: "new_dashboard",

  // Beta/Experimental
  BETA_FEATURES: "beta_features",

  // Premium/Billing
  PREMIUM_FEATURES: "premium_features",

  // File Features
  FILE_PREVIEW: "file_preview",

  // Auth Features
  OAUTH_LOGIN: "oauth_login",

  // Admin Features
  ADMIN_IMPERSONATION: "admin_impersonation",
  ADVANCED_ANALYTICS: "advanced_analytics",
} as const;

export type FeatureFlagKey = (typeof FeatureFlagKeys)[keyof typeof FeatureFlagKeys];
