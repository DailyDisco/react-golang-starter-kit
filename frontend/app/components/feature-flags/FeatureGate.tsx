import type React from "react";

import { useFeatureFlag } from "../../hooks/queries/use-feature-flags";

interface FeatureGateProps {
  /** The feature flag key to check */
  flag: string;
  /** Content to render when flag is enabled */
  children: React.ReactNode;
  /** Optional fallback content when flag is disabled */
  fallback?: React.ReactNode;
  /** Optional loading component while flags are being fetched */
  loading?: React.ReactNode;
  /** Default value while loading (defaults to false - hide feature) */
  defaultValue?: boolean;
}

/**
 * Component for conditional rendering based on feature flags
 *
 * @example
 * // Basic usage - hide content when flag is disabled
 * <FeatureGate flag="new_dashboard">
 *   <NewDashboard />
 * </FeatureGate>
 *
 * @example
 * // With fallback for disabled state
 * <FeatureGate flag="premium_features" fallback={<UpgradePrompt />}>
 *   <PremiumFeature />
 * </FeatureGate>
 *
 * @example
 * // Show feature while loading (optimistic)
 * <FeatureGate flag="beta_features" defaultValue={true}>
 *   <BetaFeature />
 * </FeatureGate>
 */
export function FeatureGate({
  flag,
  children,
  fallback = null,
  loading = null,
  defaultValue = false,
}: FeatureGateProps) {
  const { enabled, isLoading } = useFeatureFlag(flag, defaultValue);

  if (isLoading && loading !== null) {
    return <>{loading}</>;
  }

  if (enabled) {
    return <>{children}</>;
  }

  return <>{fallback}</>;
}

/**
 * Component for rendering content when flag is DISABLED
 * Inverse of FeatureGate
 */
export function FeatureGateDisabled({
  flag,
  children,
  fallback = null,
  loading = null,
  defaultValue = false,
}: FeatureGateProps) {
  const { enabled, isLoading } = useFeatureFlag(flag, defaultValue);

  if (isLoading && loading !== null) {
    return <>{loading}</>;
  }

  if (!enabled) {
    return <>{children}</>;
  }

  return <>{fallback}</>;
}
