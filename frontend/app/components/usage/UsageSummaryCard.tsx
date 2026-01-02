import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentUsage } from "@/hooks/queries/use-usage";
import { cn } from "@/lib/utils";
import { UsageService } from "@/services/usage/usageService";
import { Link } from "@tanstack/react-router";
import { Activity, AlertTriangle, BarChart3, FileUp, HardDrive, Zap } from "lucide-react";
import { useTranslation } from "react-i18next";

interface UsageSummaryCardProps {
  /** Show only the top N metrics by usage percentage */
  maxMetrics?: number;
  /** Show upgrade prompt when usage exceeds this threshold */
  warningThreshold?: number;
  /** Use compact layout */
  compact?: boolean;
  /** Show link to detailed usage page */
  showDetailsLink?: boolean;
  /** Custom class name */
  className?: string;
}

interface UsageMetric {
  key: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  current: number;
  limit: number;
  percentage: number;
  format: (v: number) => string;
  color: string;
  bgColor: string;
}

/**
 * Compact usage summary card for embedding in other pages (like billing)
 * Shows top usage metrics with warning indicators
 */
export function UsageSummaryCard({
  maxMetrics = 2,
  warningThreshold = 80,
  compact = false,
  showDetailsLink = true,
  className,
}: UsageSummaryCardProps) {
  const { t } = useTranslation("settings");
  const { data: usage, isLoading, error } = useCurrentUsage();

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className={compact ? "pb-2" : undefined}>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <div className={cn("space-y-3", compact && "space-y-2")}>
            <Skeleton className="h-12" />
            <Skeleton className="h-12" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error || !usage) {
    return null; // Silently fail - this is supplementary info
  }

  const allMetrics: UsageMetric[] = [
    {
      key: "api_calls",
      label: t("usage.metrics.apiCalls", "API Calls"),
      icon: Activity,
      current: usage.totals.api_calls,
      limit: usage.limits.api_calls,
      percentage: usage.percentages.api_calls,
      format: (v: number) => UsageService.formatCount(v),
      color: "text-blue-500",
      bgColor: "bg-blue-500/10",
    },
    {
      key: "storage_bytes",
      label: t("usage.metrics.storage", "Storage"),
      icon: HardDrive,
      current: usage.totals.storage_bytes,
      limit: usage.limits.storage_bytes,
      percentage: usage.percentages.storage_bytes,
      format: (v: number) => UsageService.formatBytes(v),
      color: "text-purple-500",
      bgColor: "bg-purple-500/10",
    },
    {
      key: "compute_ms",
      label: t("usage.metrics.computeTime", "Compute Time"),
      icon: Zap,
      current: usage.totals.compute_ms,
      limit: usage.limits.compute_ms,
      percentage: usage.percentages.compute_ms,
      format: (v: number) => UsageService.formatComputeTime(v),
      color: "text-amber-500",
      bgColor: "bg-amber-500/10",
    },
    {
      key: "file_uploads",
      label: t("usage.metrics.fileUploads", "File Uploads"),
      icon: FileUp,
      current: usage.totals.file_uploads,
      limit: usage.limits.file_uploads,
      percentage: usage.percentages.file_uploads,
      format: (v: number) => v.toLocaleString(),
      color: "text-emerald-500",
      bgColor: "bg-emerald-500/10",
    },
  ];

  // Sort by percentage descending and take top N
  const topMetrics = [...allMetrics].sort((a, b) => b.percentage - a.percentage).slice(0, maxMetrics);

  // Check if any metric exceeds warning threshold
  const hasWarning = topMetrics.some((m) => m.percentage >= warningThreshold);
  const hasExceeded = topMetrics.some((m) => m.percentage >= 100);

  const getProgressColor = (percentage: number) => {
    if (percentage >= 100) return "bg-destructive";
    if (percentage >= 90) return "bg-destructive";
    if (percentage >= warningThreshold) return "bg-warning";
    return "bg-primary";
  };

  return (
    <Card
      className={cn(
        hasExceeded && "border-destructive/50",
        hasWarning && !hasExceeded && "border-warning/50",
        className
      )}
    >
      <CardHeader className={compact ? "pb-2" : undefined}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <BarChart3 className="text-muted-foreground h-4 w-4" />
            <CardTitle className={cn("text-base", compact && "text-sm")}>
              {t("usage.summary.title", "Current Usage")}
            </CardTitle>
          </div>
          {hasExceeded && (
            <Badge
              variant="destructive"
              className="text-xs"
            >
              <AlertTriangle className="mr-1 h-3 w-3" />
              {t("usage.exceeded", "Exceeded")}
            </Badge>
          )}
          {hasWarning && !hasExceeded && (
            <Badge
              variant="warning"
              className="text-xs"
            >
              <AlertTriangle className="mr-1 h-3 w-3" />
              {t("usage.nearLimit", "Near Limit")}
            </Badge>
          )}
        </div>
        {!compact && (
          <CardDescription>
            {usage.limits_exceeded
              ? t("usage.summary.exceeded", "You've exceeded your plan limits")
              : t("usage.summary.description", "Your resource usage this billing period")}
          </CardDescription>
        )}
      </CardHeader>
      <CardContent className={compact ? "pt-0" : undefined}>
        <div className={cn("space-y-3", compact && "space-y-2")}>
          {topMetrics.map((metric) => {
            const Icon = metric.icon;
            return (
              <div
                key={metric.key}
                className="space-y-1.5"
              >
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <div className={cn("rounded p-1", metric.bgColor)}>
                      <Icon className={cn("h-3 w-3", metric.color)} />
                    </div>
                    <span className={cn("font-medium", compact && "text-xs")}>{metric.label}</span>
                  </div>
                  <span className={cn("text-muted-foreground", compact && "text-xs")}>
                    {metric.format(metric.current)} / {metric.format(metric.limit)}
                  </span>
                </div>
                <Progress
                  value={Math.min(metric.percentage, 100)}
                  className={cn("h-1.5", getProgressColor(metric.percentage))}
                />
              </div>
            );
          })}
        </div>

        {showDetailsLink && (
          <div className={cn("mt-4 flex items-center justify-between", compact && "mt-3")}>
            <Link
              to="/settings/usage"
              className="text-primary text-sm hover:underline"
            >
              {t("usage.summary.viewDetails", "View all usage details")}
            </Link>
            {(hasWarning || hasExceeded) && (
              <Button
                asChild
                size="sm"
                variant={hasExceeded ? "destructive" : "outline"}
              >
                <Link to="/billing">{t("usage.summary.upgrade", "Upgrade Plan")}</Link>
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
