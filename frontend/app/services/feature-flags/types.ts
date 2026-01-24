// Feature flag types for the frontend

/** Detailed flag information from backend including plan gating */
export interface FeatureFlagDetail {
  enabled: boolean;
  gated_by_plan: boolean;
  required_plan?: "pro" | "enterprise";
}

/** User-facing API response type - map of flag key to detail */
export type UserFeatureFlagsResponse = Record<string, FeatureFlagDetail>;

export interface UseFeatureFlagResult {
  enabled: boolean;
  isLoading: boolean;
  /** True if feature is disabled due to plan restrictions */
  gatedByPlan: boolean;
  /** Plan required to unlock (if gated) */
  requiredPlan?: string;
}

export interface UseFeatureFlagsResult {
  /** Simple boolean flags (backward compatible) */
  flags: Record<string, boolean>;
  /** Detailed flag information with gating metadata */
  flagDetails: Record<string, FeatureFlagDetail>;
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
