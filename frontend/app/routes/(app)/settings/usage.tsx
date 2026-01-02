import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { useAcknowledgeAlert, useCurrentUsage, useUsageAlerts, useUsageHistory } from "@/hooks/queries/use-usage";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { cn } from "@/lib/utils";
import { UsageService } from "@/services/usage/usageService";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  Activity,
  AlertTriangle,
  BarChart3,
  Bell,
  Check,
  Clock,
  Database,
  FileUp,
  HardDrive,
  TrendingUp,
  Zap,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/usage")({
  component: UsageSettingsPage,
});

function UsageSettingsPage() {
  const { t } = useTranslation("settings");
  const { data: usage, isLoading: usageLoading, error: usageError } = useCurrentUsage();
  const { data: history, isLoading: historyLoading } = useUsageHistory(6);
  const { data: alertsData, isLoading: alertsLoading } = useUsageAlerts();
  const acknowledgeAlert = useAcknowledgeAlert();

  const handleAcknowledgeAlert = (alertId: number) => {
    acknowledgeAlert.mutate(alertId, {
      onSuccess: () => {
        toast.success(t("usage.alerts.acknowledged", "Alert acknowledged"));
      },
      onError: () => {
        toast.error(t("usage.alerts.acknowledgeFailed", "Failed to acknowledge alert"));
      },
    });
  };

  if (usageLoading) {
    return (
      <SettingsLayout>
        <div className="space-y-6">
          <Skeleton className="h-48 rounded-xl" />
          <div className="grid gap-4 md:grid-cols-2">
            <Skeleton className="h-32" />
            <Skeleton className="h-32" />
            <Skeleton className="h-32" />
            <Skeleton className="h-32" />
          </div>
        </div>
      </SettingsLayout>
    );
  }

  if (usageError || !usage) {
    return (
      <SettingsLayout>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <AlertTriangle className="text-muted-foreground mb-4 h-12 w-12" />
            <p className="text-muted-foreground">{t("usage.error", "Failed to load usage data")}</p>
          </CardContent>
        </Card>
      </SettingsLayout>
    );
  }

  const usageMetrics = [
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

  const getProgressColor = (percentage: number) => {
    if (percentage >= 90) return "bg-destructive";
    if (percentage >= 75) return "bg-warning";
    return "bg-primary";
  };

  const formatPeriodDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
    });
  };

  return (
    <SettingsLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="from-primary via-primary/80 to-primary/60 relative overflow-hidden rounded-xl bg-gradient-to-r p-8 shadow-lg">
          <div className="absolute inset-0 bg-black/10" />
          <div className="relative flex items-center gap-4">
            <div className="rounded-full bg-white/20 p-4">
              <BarChart3 className="h-8 w-8 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-white">{t("usage.title", "Usage")}</h1>
              <p className="text-primary-foreground/80 mt-1">
                {t("usage.subtitle", "Monitor your resource usage and limits")}
              </p>
            </div>
          </div>
        </div>

        {/* Current Period Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Clock className="text-muted-foreground h-5 w-5" />
                <CardTitle>{t("usage.currentPeriod.title", "Current Billing Period")}</CardTitle>
              </div>
              <Badge variant={usage.limits_exceeded ? "destructive" : "secondary"}>
                {formatPeriodDate(usage.period_start)} - {formatPeriodDate(usage.period_end)}
              </Badge>
            </div>
            <CardDescription>
              {usage.limits_exceeded
                ? t("usage.currentPeriod.exceeded", "You have exceeded your usage limits")
                : t("usage.currentPeriod.description", "Track your resource consumption for this billing cycle")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              {usageMetrics.map((metric) => {
                const Icon = metric.icon;
                return (
                  <div
                    key={metric.key}
                    className="space-y-3 rounded-lg border p-4"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className={cn("rounded-full p-2", metric.bgColor)}>
                          <Icon className={cn("h-4 w-4", metric.color)} />
                        </div>
                        <span className="font-medium">{metric.label}</span>
                      </div>
                      <span className="text-muted-foreground text-sm">{metric.percentage}%</span>
                    </div>
                    <Progress
                      value={Math.min(metric.percentage, 100)}
                      className={cn("h-2", getProgressColor(metric.percentage))}
                    />
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">
                        {metric.format(metric.current)} {t("usage.of", "of")} {metric.format(metric.limit)}
                      </span>
                      {metric.percentage >= 90 && (
                        <Badge
                          variant="destructive"
                          className="text-xs"
                        >
                          {metric.percentage >= 100
                            ? t("usage.exceeded", "Exceeded")
                            : t("usage.nearLimit", "Near Limit")}
                        </Badge>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Alerts Section */}
        {alertsData?.alerts && alertsData.alerts.length > 0 && (
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Bell className="text-warning h-5 w-5" />
                <CardTitle>{t("usage.alerts.title", "Usage Alerts")}</CardTitle>
              </div>
              <CardDescription>
                {t("usage.alerts.description", "Active alerts that require your attention")}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {alertsData.alerts.map((alert) => {
                  const alertInfo = UsageService.getAlertTypeInfo(alert.alert_type);
                  return (
                    <div
                      key={alert.id}
                      className="border-warning/30 bg-warning/5 flex items-center justify-between rounded-lg border p-4"
                    >
                      <div className="flex items-center gap-3">
                        <AlertTriangle className="text-warning h-5 w-5" />
                        <div>
                          <p className="font-medium">
                            {UsageService.getUsageTypeLabel(alert.usage_type)} - {alertInfo.label}
                          </p>
                          <p className="text-muted-foreground text-sm">
                            {alert.percentage_used}% {t("usage.alerts.used", "used")} (
                            {alert.current_usage.toLocaleString()} / {alert.usage_limit.toLocaleString()})
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleAcknowledgeAlert(alert.id)}
                        disabled={acknowledgeAlert.isPending}
                      >
                        <Check className="mr-1 h-4 w-4" />
                        {t("usage.alerts.acknowledge", "Acknowledge")}
                      </Button>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Usage History */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <TrendingUp className="text-muted-foreground h-5 w-5" />
              <CardTitle>{t("usage.history.title", "Usage History")}</CardTitle>
            </div>
            <CardDescription>
              {t("usage.history.description", "Your resource usage over the past 6 months")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {historyLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <Skeleton
                    key={i}
                    className="h-20"
                  />
                ))}
              </div>
            ) : history?.history && history.history.length > 0 ? (
              <div className="space-y-4">
                {history.history.map((period, index) => (
                  <div
                    key={index}
                    className="hover:bg-muted/50 rounded-lg border p-4 transition-colors"
                  >
                    <div className="mb-3 flex items-center justify-between">
                      <span className="font-medium">
                        {formatPeriodDate(period.period_start)} - {formatPeriodDate(period.period_end)}
                      </span>
                      {period.limits_exceeded && <Badge variant="destructive">{t("usage.exceeded", "Exceeded")}</Badge>}
                    </div>
                    <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                      <div className="text-center">
                        <div className="text-muted-foreground flex items-center justify-center gap-1 text-xs">
                          <Activity className="h-3 w-3" />
                          {t("usage.metrics.apiCalls", "API Calls")}
                        </div>
                        <p className="font-medium">{UsageService.formatCount(period.totals.api_calls)}</p>
                      </div>
                      <div className="text-center">
                        <div className="text-muted-foreground flex items-center justify-center gap-1 text-xs">
                          <Database className="h-3 w-3" />
                          {t("usage.metrics.storage", "Storage")}
                        </div>
                        <p className="font-medium">{UsageService.formatBytes(period.totals.storage_bytes)}</p>
                      </div>
                      <div className="text-center">
                        <div className="text-muted-foreground flex items-center justify-center gap-1 text-xs">
                          <Zap className="h-3 w-3" />
                          {t("usage.metrics.computeTime", "Compute")}
                        </div>
                        <p className="font-medium">{UsageService.formatComputeTime(period.totals.compute_ms)}</p>
                      </div>
                      <div className="text-center">
                        <div className="text-muted-foreground flex items-center justify-center gap-1 text-xs">
                          <FileUp className="h-3 w-3" />
                          {t("usage.metrics.fileUploads", "Uploads")}
                        </div>
                        <p className="font-medium">{period.totals.file_uploads.toLocaleString()}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-muted-foreground flex flex-col items-center justify-center py-8">
                <BarChart3 className="mb-2 h-8 w-8 opacity-50" />
                <p>{t("usage.history.noData", "No usage history available yet")}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Upgrade CTA Card */}
        <Card className={cn(usage.limits_exceeded && "border-destructive/50 bg-destructive/5")}>
          <CardContent className="pt-6">
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-4">
                <div className={cn("rounded-full p-3", usage.limits_exceeded ? "bg-destructive/10" : "bg-info/10")}>
                  <Zap className={cn("h-5 w-5", usage.limits_exceeded ? "text-destructive" : "text-info")} />
                </div>
                <div>
                  <h4 className="font-medium">
                    {usage.limits_exceeded
                      ? t("usage.info.exceededTitle", "You've exceeded your plan limits")
                      : t("usage.info.title", "Need more resources?")}
                  </h4>
                  <p className="text-muted-foreground mt-1 text-sm">
                    {usage.limits_exceeded
                      ? t("usage.info.exceededDescription", "Upgrade now to restore full access and get higher limits.")
                      : t(
                          "usage.info.description",
                          "Upgrade your plan to get higher limits and unlock additional features."
                        )}
                  </p>
                </div>
              </div>
              <Button
                asChild
                variant={usage.limits_exceeded ? "destructive" : "default"}
              >
                <Link to="/billing">{t("usage.info.upgrade", "Upgrade Plan")}</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </SettingsLayout>
  );
}
