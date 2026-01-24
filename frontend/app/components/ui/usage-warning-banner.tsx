import { Link } from "@tanstack/react-router";
import { AlertTriangle, TrendingUp, XCircle } from "lucide-react";
import { useTranslation } from "react-i18next";

import type { WarningLevel } from "@/hooks/useUsageWarning";
import { cn } from "@/lib/utils";

import { Alert, AlertDescription, AlertTitle } from "./alert";
import { Button } from "./button";
import { Progress } from "./progress";

interface UsageWarningBannerProps {
  /** Warning level determines styling */
  level: WarningLevel;
  /** Percentage used (0-100+) */
  percentage: number;
  /** Warning message to display */
  message: string;
  /** Whether to show progress bar */
  showProgress?: boolean;
  /** Whether to show upgrade CTA */
  showUpgrade?: boolean;
  /** Optional callback when dismissed */
  onDismiss?: () => void;
  /** Additional CSS classes */
  className?: string;
}

const LEVEL_STYLES: Record<WarningLevel, { variant: "default" | "destructive"; icon: typeof AlertTriangle }> = {
  none: { variant: "default", icon: TrendingUp },
  approaching: { variant: "default", icon: TrendingUp },
  warning: { variant: "default", icon: AlertTriangle },
  critical: { variant: "destructive", icon: AlertTriangle },
  exceeded: { variant: "destructive", icon: XCircle },
};

/**
 * Inline usage warning banner for feature components.
 * Shows contextual warnings when users are approaching limits.
 *
 * @example
 * const { warning, shouldWarn } = useUsageWarning("file_uploads");
 *
 * {shouldWarn && (
 *   <UsageWarningBanner
 *     level={warning.level}
 *     percentage={warning.percentage}
 *     message={warning.message}
 *     showUpgrade
 *   />
 * )}
 */
export function UsageWarningBanner({
  level,
  percentage,
  message,
  showProgress = false,
  showUpgrade = true,
  onDismiss,
  className,
}: UsageWarningBannerProps) {
  const { t } = useTranslation("billing");

  if (level === "none") return null;

  const { variant, icon: Icon } = LEVEL_STYLES[level];
  const isExceeded = level === "exceeded";

  return (
    <Alert
      variant={variant}
      className={cn("mb-4", className)}
    >
      <Icon className="h-4 w-4" />
      <AlertTitle className="flex items-center justify-between">
        <span>
          {isExceeded
            ? t("usage.exceeded", "Usage Limit Exceeded")
            : t("usage.approaching", "Approaching Usage Limit")}
        </span>
        {onDismiss && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={onDismiss}
            aria-label="Dismiss"
          >
            ×
          </Button>
        )}
      </AlertTitle>
      <AlertDescription className="space-y-2">
        <p>{message}</p>

        {showProgress && (
          <Progress
            value={Math.min(percentage, 100)}
            className={cn(
              "h-2",
              level === "critical" || level === "exceeded" ? "bg-destructive/20" : "bg-muted"
            )}
          />
        )}

        {showUpgrade && (
          <div className="pt-1">
            <Button
              variant={isExceeded ? "default" : "outline"}
              size="sm"
              asChild
            >
              <Link to="/billing">
                {isExceeded
                  ? t("usage.upgradeNow", "Upgrade Now")
                  : t("usage.viewPlans", "View Plans")}
              </Link>
            </Button>
          </div>
        )}
      </AlertDescription>
    </Alert>
  );
}

/**
 * Compact inline usage indicator for tight spaces.
 * Shows just the percentage with color coding.
 */
export function UsageIndicator({
  level,
  percentage,
  label,
  className,
}: {
  level: WarningLevel;
  percentage: number;
  label?: string;
  className?: string;
}) {
  if (level === "none") return null;

  const colorClass =
    level === "exceeded" || level === "critical"
      ? "text-destructive"
      : level === "warning"
        ? "text-warning"
        : "text-muted-foreground";

  return (
    <span className={cn("inline-flex items-center gap-1 text-xs", colorClass, className)}>
      <AlertTriangle className="h-3 w-3" />
      {label && <span>{label}:</span>}
      <span>{Math.round(percentage)}%</span>
    </span>
  );
}
