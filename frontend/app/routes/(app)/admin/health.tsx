import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { AdminLayout } from "@/layouts/AdminLayout";
import { requireAdmin } from "@/lib/guards";
import { AdminSettingsService, type HealthComponent, type SystemHealth } from "@/services/admin";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Activity, Database, HardDrive, RefreshCw, Server, Wifi } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/admin/health")({
  beforeLoad: () => requireAdmin(),
  component: SystemHealthPage,
});

function SystemHealthPage() {
  const { t } = useTranslation("admin");
  const {
    data: health,
    isLoading,
    error,
    refetch,
    isFetching,
  } = useQuery<SystemHealth>({
    queryKey: ["admin", "health"],
    queryFn: () => AdminSettingsService.getSystemHealth(),
    refetchInterval: 30000,
  });

  const getStatusBadgeVariant = (status: string): "success" | "warning" | "destructive" | "secondary" => {
    switch (status) {
      case "healthy":
        return "success";
      case "degraded":
        return "warning";
      case "unhealthy":
        return "destructive";
      default:
        return "secondary";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy":
        return "border-success/30 bg-success/5";
      case "degraded":
        return "border-warning/30 bg-warning/5";
      case "unhealthy":
        return "border-destructive/30 bg-destructive/5";
      default:
        return "";
    }
  };

  const getComponentIcon = (name: string) => {
    switch (name.toLowerCase()) {
      case "database":
        return <Database className="h-5 w-5" />;
      case "cache":
        return <Server className="h-5 w-5" />;
      case "storage":
        return <HardDrive className="h-5 w-5" />;
      case "api":
        return <Wifi className="h-5 w-5" />;
      default:
        return <Activity className="h-5 w-5" />;
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "healthy":
        return t("health.status.healthy");
      case "degraded":
        return t("health.status.degraded");
      case "unhealthy":
        return t("health.status.unhealthy");
      default:
        return t("health.status.unknown");
    }
  };

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="space-y-6">
          <div>
            <h2 className="text-2xl font-bold">{t("health.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("health.subtitle")}</p>
          </div>
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {[...Array(4)].map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <div className="bg-muted h-6 w-32 animate-pulse rounded" />
                </CardHeader>
                <CardContent>
                  <div className="bg-muted/50 h-20 animate-pulse rounded" />
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="space-y-6">
          <div>
            <h2 className="text-2xl font-bold">{t("health.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("health.subtitle")}</p>
          </div>
          <Card className="border-destructive/30 bg-destructive/5">
            <CardHeader>
              <CardTitle className="text-destructive">{t("health.error.title")}</CardTitle>
              <CardDescription className="text-destructive/80">
                {error instanceof Error ? error.message : t("health.error.default")}
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">{t("health.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("health.subtitle")}</p>
          </div>
          <div className="flex items-center gap-4">
            <Badge
              variant={getStatusBadgeVariant(health?.status || "unknown")}
              className="text-sm"
            >
              {getStatusText(health?.status || "unknown").toUpperCase()}
            </Badge>
            <Button
              variant="outline"
              onClick={() => refetch()}
              disabled={isFetching}
              className="gap-2"
            >
              <RefreshCw className={`h-4 w-4 ${isFetching ? "animate-spin" : ""}`} />
              {t("health.refresh")}
            </Button>
          </div>
        </div>

        {/* Last Updated */}
        <p className="text-muted-foreground text-sm">
          {t("health.lastUpdated", {
            time: health?.timestamp ? new Date(health.timestamp).toLocaleString() : t("health.notAvailable"),
          })}{" "}
          (auto-refreshes every 30s)
        </p>

        {/* Overall Status Card */}
        <Card className={getStatusColor(health?.status || "unknown")}>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Activity className="h-6 w-6" />
              <CardTitle>{t("health.overallStatus")}</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold capitalize">{getStatusText(health?.status || "unknown")}</p>
          </CardContent>
        </Card>

        {/* Components Grid */}
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {health?.components?.map((component: HealthComponent) => (
            <Card
              key={component.name}
              className={getStatusColor(component.status)}
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {getComponentIcon(component.name)}
                    <CardTitle className="capitalize">{component.name}</CardTitle>
                  </div>
                  <Badge variant={getStatusBadgeVariant(component.status)}>{getStatusText(component.status)}</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {component.message && <p className="text-sm">{component.message}</p>}
                {component.latency && (
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">{t("health.latency")}</span>
                    <span className="font-medium">{component.latency}</span>
                  </div>
                )}
                {component.details && Object.keys(component.details).length > 0 && (
                  <div className="bg-background/50 mt-2 rounded p-3">
                    <p className="text-muted-foreground mb-2 text-xs font-medium uppercase">{t("health.details")}</p>
                    <dl className="grid grid-cols-2 gap-2 text-sm">
                      {Object.entries(component.details).map(([key, value]) => (
                        <div
                          key={key}
                          className="flex flex-col"
                        >
                          <dt className="text-muted-foreground capitalize">{key.replace(/_/g, " ")}</dt>
                          <dd className="font-medium">{String(value)}</dd>
                        </div>
                      ))}
                    </dl>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Metrics Section */}
        {health?.metrics && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">{t("health.metrics.title")}</h3>

            {/* Database Metrics */}
            {health.metrics.database && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Database className="h-5 w-5" />
                    <CardTitle>{t("health.metrics.database.title")}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                    <MetricItem
                      label={t("health.metrics.database.activeConnections")}
                      value={health.metrics.database.connections_active}
                    />
                    <MetricItem
                      label={t("health.metrics.database.idleConnections")}
                      value={health.metrics.database.connections_idle}
                    />
                    <MetricItem
                      label={t("health.metrics.database.maxConnections")}
                      value={health.metrics.database.connections_max}
                    />
                    <MetricItem
                      label={t("health.metrics.database.avgQueryTime")}
                      value={health.metrics.database.avg_query_time}
                    />
                  </div>
                  {health.metrics.database.connections_max > 0 && (
                    <div className="mt-4 space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">
                          {t("health.metrics.database.connectionPoolUsage")}
                        </span>
                        <span className="font-medium">
                          {Math.round(
                            ((health.metrics.database.connections_active + health.metrics.database.connections_idle) /
                              health.metrics.database.connections_max) *
                              100
                          )}
                          %
                        </span>
                      </div>
                      <Progress
                        value={
                          ((health.metrics.database.connections_active + health.metrics.database.connections_idle) /
                            health.metrics.database.connections_max) *
                          100
                        }
                      />
                    </div>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Cache Metrics */}
            {health.metrics.cache && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Server className="h-5 w-5" />
                    <CardTitle>{t("health.metrics.cache.title")}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                    <MetricItem
                      label={t("health.metrics.cache.memoryUsed")}
                      value={health.metrics.cache.memory_used}
                    />
                    <MetricItem
                      label={t("health.metrics.cache.memoryMax")}
                      value={health.metrics.cache.memory_max}
                    />
                    <MetricItem
                      label={t("health.metrics.cache.hitRate")}
                      value={`${health.metrics.cache.hit_rate.toFixed(1)}%`}
                    />
                    <MetricItem
                      label={t("health.metrics.cache.keys")}
                      value={health.metrics.cache.keys}
                    />
                  </div>
                  <div className="mt-4 space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">{t("health.metrics.cache.cacheHitRate")}</span>
                      <span className="font-medium">{health.metrics.cache.hit_rate.toFixed(1)}%</span>
                    </div>
                    <Progress value={health.metrics.cache.hit_rate} />
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Storage Metrics */}
            {health.metrics.storage && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <HardDrive className="h-5 w-5" />
                    <CardTitle>{t("health.metrics.storage.title")}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                    <MetricItem
                      label={t("health.metrics.storage.used")}
                      value={health.metrics.storage.used}
                    />
                    <MetricItem
                      label={t("health.metrics.storage.available")}
                      value={health.metrics.storage.available}
                    />
                    <MetricItem
                      label={t("health.metrics.storage.total")}
                      value={health.metrics.storage.total}
                    />
                    <MetricItem
                      label={t("health.metrics.storage.files")}
                      value={health.metrics.storage.file_count}
                    />
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        )}
      </div>
    </AdminLayout>
  );
}

function MetricItem({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <p className="text-muted-foreground text-sm">{label}</p>
      <p className="text-xl font-bold">{value}</p>
    </div>
  );
}
