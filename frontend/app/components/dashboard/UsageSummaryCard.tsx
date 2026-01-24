import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentUsage } from "@/hooks/queries/use-usage";
import { useAllUsageWarnings, type UsageType } from "@/hooks/useUsageWarning";
import { cn } from "@/lib/utils";
import { UsageService } from "@/services/usage/usageService";
import { Link } from "@tanstack/react-router";
import { Activity, AlertTriangle, ArrowRight, TrendingUp } from "lucide-react";
import { useTranslation } from "react-i18next";

interface UsageItemProps {
  type: UsageType;
  label: string;
  current: number;
  limit: number;
  percentage: number;
  formatValue: (value: number) => string;
}

function UsageItem({ label, current, limit, percentage, formatValue }: UsageItemProps) {
  const getProgressColor = (pct: number) => {
    if (pct >= 90) return "bg-red-500";
    if (pct >= 70) return "bg-amber-500";
    return "bg-primary";
  };

  const clampedPercentage = Math.min(percentage, 100);

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">
          {formatValue(current)} / {formatValue(limit)}
        </span>
      </div>
      <div className="bg-primary/20 relative h-2 w-full overflow-hidden rounded-full">
        <div
          className={cn("h-full transition-all", getProgressColor(percentage))}
          style={{ width: `${clampedPercentage}%` }}
        />
      </div>
    </div>
  );
}

function UsageSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-5 w-16" />
        </div>
        <Skeleton className="h-4 w-48" />
      </CardHeader>
      <CardContent className="space-y-4">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="space-y-1.5"
          >
            <div className="flex justify-between">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 w-24" />
            </div>
            <Skeleton className="h-2 w-full" />
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

export function UsageSummaryCard() {
  const { t } = useTranslation("dashboard");
  const { data: usage, isLoading } = useCurrentUsage();
  const { hasAnyWarnings, hasExceeded, mostCritical } = useAllUsageWarnings();

  if (isLoading) {
    return <UsageSkeleton />;
  }

  if (!usage) {
    return null;
  }

  const usageItems: Array<{
    type: UsageType;
    label: string;
    formatValue: (value: number) => string;
  }> = [
    {
      type: "api_calls",
      label: t("usage.apiCalls", "API Calls"),
      formatValue: UsageService.formatCount,
    },
    {
      type: "storage_bytes",
      label: t("usage.storage", "Storage"),
      formatValue: UsageService.formatBytes,
    },
    {
      type: "file_uploads",
      label: t("usage.fileUploads", "File Uploads"),
      formatValue: UsageService.formatCount,
    },
    {
      type: "compute_ms",
      label: t("usage.computeTime", "Compute Time"),
      formatValue: UsageService.formatComputeTime,
    },
  ];

  return (
    <Card className={hasExceeded ? "border-red-500/50" : hasAnyWarnings ? "border-amber-500/50" : undefined}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <TrendingUp className="text-primary h-5 w-5" />
            {t("usage.title", "Usage Summary")}
          </CardTitle>
          {hasExceeded ? (
            <Badge
              variant="destructive"
              className="gap-1"
            >
              <AlertTriangle
                className="h-3 w-3"
                aria-hidden="true"
              />
              {t("usage.exceeded", "Limit Exceeded")}
            </Badge>
          ) : hasAnyWarnings ? (
            <Badge
              variant="outline"
              className="gap-1 border-amber-500/50 text-amber-600"
            >
              <Activity
                className="h-3 w-3"
                aria-hidden="true"
              />
              {t("usage.approaching", "Approaching Limit")}
            </Badge>
          ) : null}
        </div>
        <CardDescription>
          {usage.period_start && usage.period_end
            ? t("usage.period", "Current billing period: {{start}} - {{end}}", {
                start: new Date(usage.period_start).toLocaleDateString(),
                end: new Date(usage.period_end).toLocaleDateString(),
              })
            : t("usage.currentPeriod", "Current billing period")}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {usageItems.map((item) => (
          <UsageItem
            key={item.type}
            type={item.type}
            label={item.label}
            current={usage.totals[item.type]}
            limit={usage.limits[item.type]}
            percentage={usage.percentages[item.type]}
            formatValue={item.formatValue}
          />
        ))}

        {mostCritical && mostCritical.level !== "none" && (
          <div className="mt-4 rounded-lg bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-400">
            {mostCritical.message}
          </div>
        )}

        <Link
          to="/billing"
          className="group text-muted-foreground hover:text-primary mt-2 flex items-center justify-center gap-2 text-sm transition-colors"
        >
          {t("usage.viewDetails", "View billing details")}
          <ArrowRight
            className="h-4 w-4 transition-transform group-hover:translate-x-1"
            aria-hidden="true"
          />
        </Link>
      </CardContent>
    </Card>
  );
}
