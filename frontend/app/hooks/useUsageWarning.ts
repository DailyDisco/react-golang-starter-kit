import { useMemo } from "react";

import { useCurrentUsage } from "./queries/use-usage";

export type UsageType = "api_calls" | "storage_bytes" | "compute_ms" | "file_uploads";
export type WarningLevel = "none" | "approaching" | "warning" | "critical" | "exceeded";

interface UsageWarning {
  level: WarningLevel;
  percentage: number;
  current: number;
  limit: number;
  message: string;
}

const WARNING_THRESHOLDS = {
  approaching: 70, // 70-79%
  warning: 80, // 80-89%
  critical: 90, // 90-99%
  exceeded: 100, // 100%+
} as const;

const USAGE_TYPE_LABELS: Record<UsageType, string> = {
  api_calls: "API calls",
  storage_bytes: "storage",
  compute_ms: "compute time",
  file_uploads: "file uploads",
};

function getWarningLevel(percentage: number): WarningLevel {
  if (percentage >= WARNING_THRESHOLDS.exceeded) return "exceeded";
  if (percentage >= WARNING_THRESHOLDS.critical) return "critical";
  if (percentage >= WARNING_THRESHOLDS.warning) return "warning";
  if (percentage >= WARNING_THRESHOLDS.approaching) return "approaching";
  return "none";
}

function getWarningMessage(type: UsageType, level: WarningLevel, percentage: number): string {
  const label = USAGE_TYPE_LABELS[type];
  const rounded = Math.round(percentage);

  switch (level) {
    case "exceeded":
      return `You've exceeded your ${label} limit. Upgrade your plan to continue.`;
    case "critical":
      return `You've used ${rounded}% of your ${label} limit.`;
    case "warning":
      return `You're approaching your ${label} limit (${rounded}% used).`;
    case "approaching":
      return `${rounded}% of your ${label} limit used.`;
    default:
      return "";
  }
}

/**
 * Hook to check usage warnings for a specific usage type.
 * Use this in feature components to show inline warnings when approaching limits.
 *
 * @example
 * const { warning, isApproachingLimit } = useUsageWarning("file_uploads");
 *
 * {isApproachingLimit && <UsageWarningBanner warning={warning} />}
 */
export function useUsageWarning(type: UsageType): {
  warning: UsageWarning;
  isApproachingLimit: boolean;
  shouldWarn: boolean;
  isExceeded: boolean;
  isLoading: boolean;
} {
  const { data: usage, isLoading } = useCurrentUsage();

  const warning = useMemo<UsageWarning>(() => {
    if (!usage) {
      return {
        level: "none",
        percentage: 0,
        current: 0,
        limit: 0,
        message: "",
      };
    }

    const percentage = usage.percentages[type];
    const level = getWarningLevel(percentage);
    const message = getWarningMessage(type, level, percentage);

    return {
      level,
      percentage,
      current: usage.totals[type],
      limit: usage.limits[type],
      message,
    };
  }, [usage, type]);

  return {
    warning,
    isApproachingLimit: warning.level !== "none",
    shouldWarn: warning.level === "warning" || warning.level === "critical" || warning.level === "exceeded",
    isExceeded: warning.level === "exceeded",
    isLoading,
  };
}

/**
 * Hook to check all usage warnings at once.
 * Useful for dashboard or settings pages.
 */
export function useAllUsageWarnings(): {
  warnings: Record<UsageType, UsageWarning>;
  hasAnyWarnings: boolean;
  hasExceeded: boolean;
  mostCritical: UsageWarning | null;
  isLoading: boolean;
} {
  const { data: usage, isLoading } = useCurrentUsage();

  const warnings = useMemo<Record<UsageType, UsageWarning>>(() => {
    const types: UsageType[] = ["api_calls", "storage_bytes", "compute_ms", "file_uploads"];
    const result = {} as Record<UsageType, UsageWarning>;

    for (const type of types) {
      if (!usage) {
        result[type] = {
          level: "none",
          percentage: 0,
          current: 0,
          limit: 0,
          message: "",
        };
      } else {
        const percentage = usage.percentages[type];
        const level = getWarningLevel(percentage);
        result[type] = {
          level,
          percentage,
          current: usage.totals[type],
          limit: usage.limits[type],
          message: getWarningMessage(type, level, percentage),
        };
      }
    }

    return result;
  }, [usage]);

  const { hasAnyWarnings, hasExceeded, mostCritical } = useMemo(() => {
    const all = Object.values(warnings);
    const withWarnings = all.filter((w) => w.level !== "none");
    const exceeded = all.some((w) => w.level === "exceeded");

    // Sort by percentage descending to find most critical
    const sorted = [...withWarnings].sort((a, b) => b.percentage - a.percentage);

    return {
      hasAnyWarnings: withWarnings.length > 0,
      hasExceeded: exceeded,
      mostCritical: sorted[0] ?? null,
    };
  }, [warnings]);

  return {
    warnings,
    hasAnyWarnings,
    hasExceeded,
    mostCritical,
    isLoading,
  };
}
