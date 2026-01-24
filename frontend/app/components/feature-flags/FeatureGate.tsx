import type React from "react";

import { useFeatureFlag } from "../../hooks/queries/use-feature-flags";
import { useUsageWarning, type UsageType } from "../../hooks/useUsageWarning";

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
  /**
   * Render function for upgrade prompt when gated by plan.
   * If provided and feature is gated by plan, this will be rendered instead of fallback.
   * @param requiredPlan - The plan required to unlock this feature ("pro" or "enterprise")
   */
  upgradePrompt?: (requiredPlan: string) => React.ReactNode;
  /**
   * Optional usage type to check limits for.
   * When specified, the feature will be blocked if usage is exceeded.
   */
  usageType?: UsageType;
  /**
   * Render function for usage exceeded state.
   * If provided and usage limit is exceeded, this will be rendered.
   * @param percentage - Current usage percentage
   * @param message - Descriptive warning message
   */
  usageExceededPrompt?: (percentage: number, message: string) => React.ReactNode;
  /**
   * Whether to block the feature when usage is exceeded (default: true).
   * Set to false to show the feature but with a warning via usageWarningPrompt.
   */
  blockOnUsageExceeded?: boolean;
  /**
   * Render function for usage warning (approaching limit but not exceeded).
   * This is rendered alongside children, not instead of them.
   * @param percentage - Current usage percentage
   * @param message - Descriptive warning message
   */
  usageWarningPrompt?: (percentage: number, message: string) => React.ReactNode;
}

/**
 * Component for conditional rendering based on feature flags and usage limits
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
 *
 * @example
 * // With upgrade prompt for plan-gated features
 * <FeatureGate
 *   flag="advanced_analytics"
 *   upgradePrompt={(plan) => <UpgradeBanner plan={plan} />}
 * >
 *   <AdvancedAnalytics />
 * </FeatureGate>
 *
 * @example
 * // With usage-based gating - block when usage exceeded
 * <FeatureGate
 *   flag="file_uploads"
 *   usageType="file_uploads"
 *   usageExceededPrompt={(pct, msg) => <UsageLimitBanner message={msg} />}
 *   usageWarningPrompt={(pct, msg) => <UsageWarning percentage={pct} />}
 * >
 *   <FileUploader />
 * </FeatureGate>
 */
export function FeatureGate({
  flag,
  children,
  fallback = null,
  loading = null,
  defaultValue = false,
  upgradePrompt,
  usageType,
  usageExceededPrompt,
  blockOnUsageExceeded = true,
  usageWarningPrompt,
}: FeatureGateProps) {
  const { enabled, isLoading, gatedByPlan, requiredPlan } = useFeatureFlag(flag, defaultValue);

  // Only call useUsageWarning when usageType is provided
  // We need to call hooks unconditionally, so use a dummy value when not needed
  const { warning, isExceeded, shouldWarn, isLoading: isUsageLoading } = useUsageWarning(usageType ?? "api_calls");

  // Only consider usage data if usageType was explicitly provided
  const checkUsage = usageType !== undefined;
  const usageIsExceeded = checkUsage && isExceeded;
  const usageShouldWarn = checkUsage && shouldWarn && !isExceeded;

  if (isLoading && loading !== null) {
    return <>{loading}</>;
  }

  // If feature flag is not enabled, show appropriate fallback
  if (!enabled) {
    // Show upgrade prompt if feature is gated by plan and prompt is provided
    if (gatedByPlan && upgradePrompt && requiredPlan) {
      return <>{upgradePrompt(requiredPlan)}</>;
    }
    return <>{fallback}</>;
  }

  // Feature is enabled, now check usage limits
  if (usageIsExceeded && blockOnUsageExceeded) {
    // Usage exceeded and blocking is enabled
    if (usageExceededPrompt) {
      return <>{usageExceededPrompt(warning.percentage, warning.message)}</>;
    }
    // No prompt provided, show fallback
    return <>{fallback}</>;
  }

  // Feature is enabled and either usage is OK or not blocking on exceeded
  // Show warning alongside content if approaching limit
  if (usageShouldWarn && usageWarningPrompt) {
    return (
      <>
        {usageWarningPrompt(warning.percentage, warning.message)}
        {children}
      </>
    );
  }

  // Show exceeded warning (non-blocking mode) alongside content
  if (usageIsExceeded && !blockOnUsageExceeded && usageExceededPrompt) {
    return (
      <>
        {usageExceededPrompt(warning.percentage, warning.message)}
        {children}
      </>
    );
  }

  return <>{children}</>;
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
}: Omit<FeatureGateProps, "usageType" | "usageExceededPrompt" | "blockOnUsageExceeded" | "usageWarningPrompt">) {
  const { enabled, isLoading } = useFeatureFlag(flag, defaultValue);

  if (isLoading && loading !== null) {
    return <>{loading}</>;
  }

  if (!enabled) {
    return <>{children}</>;
  }

  return <>{fallback}</>;
}
