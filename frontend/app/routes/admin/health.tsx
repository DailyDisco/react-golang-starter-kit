import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Activity, Database, HardDrive, RefreshCw, Server, Wifi } from "lucide-react";

import { AdminPageHeader } from "../../components/admin";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Progress } from "../../components/ui/progress";
import { requireAdmin } from "../../lib/guards";
import { AdminSettingsService, type HealthComponent, type SystemHealth } from "../../services/admin";

export const Route = createFileRoute("/admin/health")({
  beforeLoad: () => requireAdmin(),
  component: SystemHealthPage,
});

function SystemHealthPage() {
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

  if (isLoading) {
    return (
      <div className="space-y-6">
        <AdminPageHeader
          title="System Health"
          description="Monitor system status and performance metrics"
          breadcrumbs={[{ label: "System Health" }]}
        />
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="h-6 w-32 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
              </CardHeader>
              <CardContent>
                <div className="h-20 animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <AdminPageHeader
          title="System Health"
          description="Monitor system status and performance metrics"
          breadcrumbs={[{ label: "System Health" }]}
        />
        <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950">
          <CardHeader>
            <CardTitle className="text-red-600 dark:text-red-400">Error Loading Health Status</CardTitle>
            <CardDescription className="text-red-500 dark:text-red-400">
              {error instanceof Error ? error.message : "Failed to load system health"}
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const getStatusBadgeVariant = (status: string): "default" | "secondary" | "outline" | "destructive" => {
    switch (status) {
      case "healthy":
        return "default";
      case "degraded":
        return "secondary";
      case "unhealthy":
        return "destructive";
      default:
        return "outline";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy":
        return "border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950";
      case "degraded":
        return "border-yellow-200 bg-yellow-50 dark:border-yellow-900 dark:bg-yellow-950";
      case "unhealthy":
        return "border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950";
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

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="System Health"
        description="Monitor system status and performance metrics"
        breadcrumbs={[{ label: "System Health" }]}
        actions={
          <div className="flex items-center gap-4">
            <Badge
              variant={getStatusBadgeVariant(health?.status || "unknown")}
              className="text-sm"
            >
              {health?.status?.toUpperCase() || "Unknown"}
            </Badge>
            <Button
              variant="outline"
              onClick={() => refetch()}
              disabled={isFetching}
              className="gap-2"
            >
              <RefreshCw className={`h-4 w-4 ${isFetching ? "animate-spin" : ""}`} />
              Refresh
            </Button>
          </div>
        }
      />

      {/* Last Updated */}
      <p className="text-muted-foreground text-sm">
        Last updated: {health?.timestamp ? new Date(health.timestamp).toLocaleString() : "N/A"} (auto-refreshes every
        30s)
      </p>

      {/* Overall Status Card */}
      <Card className={getStatusColor(health?.status || "unknown")}>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Activity className="h-6 w-6" />
            <CardTitle>Overall System Status</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold capitalize">{health?.status || "Unknown"}</p>
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
                <Badge variant={getStatusBadgeVariant(component.status)}>{component.status}</Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              {component.message && <p className="text-sm">{component.message}</p>}
              {component.latency && (
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Latency</span>
                  <span className="font-medium">{component.latency}</span>
                </div>
              )}
              {component.details && Object.keys(component.details).length > 0 && (
                <div className="mt-2 rounded bg-white/50 p-3 dark:bg-black/20">
                  <p className="text-muted-foreground mb-2 text-xs font-medium uppercase">Details</p>
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
          <h3 className="text-lg font-semibold">System Metrics</h3>

          {/* Database Metrics */}
          {health.metrics.database && (
            <Card>
              <CardHeader>
                <div className="flex items-center gap-2">
                  <Database className="h-5 w-5" />
                  <CardTitle>Database Metrics</CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                  <MetricItem
                    label="Active Connections"
                    value={health.metrics.database.connections_active}
                  />
                  <MetricItem
                    label="Idle Connections"
                    value={health.metrics.database.connections_idle}
                  />
                  <MetricItem
                    label="Max Connections"
                    value={health.metrics.database.connections_max}
                  />
                  <MetricItem
                    label="Avg Query Time"
                    value={health.metrics.database.avg_query_time}
                  />
                </div>
                {health.metrics.database.connections_max > 0 && (
                  <div className="mt-4 space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Connection Pool Usage</span>
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
                  <CardTitle>Cache Metrics</CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                  <MetricItem
                    label="Memory Used"
                    value={health.metrics.cache.memory_used}
                  />
                  <MetricItem
                    label="Memory Max"
                    value={health.metrics.cache.memory_max}
                  />
                  <MetricItem
                    label="Hit Rate"
                    value={`${health.metrics.cache.hit_rate.toFixed(1)}%`}
                  />
                  <MetricItem
                    label="Keys"
                    value={health.metrics.cache.keys}
                  />
                </div>
                <div className="mt-4 space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Cache Hit Rate</span>
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
                  <CardTitle>Storage Metrics</CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
                  <MetricItem
                    label="Used"
                    value={health.metrics.storage.used}
                  />
                  <MetricItem
                    label="Available"
                    value={health.metrics.storage.available}
                  />
                  <MetricItem
                    label="Total"
                    value={health.metrics.storage.total}
                  />
                  <MetricItem
                    label="Files"
                    value={health.metrics.storage.file_count}
                  />
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
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
