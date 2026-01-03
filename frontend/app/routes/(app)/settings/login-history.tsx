import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { requireAuth } from "@/lib/guards";
import { queryKeys } from "@/lib/query-keys";
import { loginHistoryQueryOptions } from "@/lib/route-query-options";
import { SettingsService, type LoginHistoryEntry } from "@/services/settings/settingsService";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Check, Chrome, Clock, Globe, MapPin, Monitor, RefreshCw, Smartphone, X } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/settings/login-history")({
  // Ensure user is authenticated before loading data
  beforeLoad: async (ctx) => {
    await requireAuth(ctx);
  },
  // Prefetch login history data before component renders for faster navigation
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(loginHistoryQueryOptions());
  },
  component: LoginHistoryPage,
});

function LoginHistoryPage() {
  const { t } = useTranslation("settings");
  const {
    data: loginHistory,
    isLoading,
    refetch,
    isRefetching,
  } = useQuery({
    queryKey: queryKeys.settings.loginHistory(),
    queryFn: () => SettingsService.getLoginHistory(50),
  });

  const getDeviceIcon = (deviceType: string) => {
    const lower = deviceType.toLowerCase();
    if (lower.includes("mobile") || lower.includes("phone") || lower.includes("tablet")) {
      return <Smartphone className="h-5 w-5" />;
    }
    return <Monitor className="h-5 w-5" />;
  };

  const getBrowserIcon = (browser: string) => {
    return <Chrome className="h-4 w-4" />;
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return t("loginHistory.time.justNow");
    if (diffMins < 60) return t("loginHistory.time.minutesAgo", { count: diffMins });
    if (diffHours < 24) return t("loginHistory.time.hoursAgo", { count: diffHours });
    if (diffDays < 7) return t("loginHistory.time.daysAgo", { count: diffDays });

    return date.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
      hour: "numeric",
      minute: "2-digit",
    });
  };

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">{t("loginHistory.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("loginHistory.subtitle")}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isRefetching}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${isRefetching ? "animate-spin" : ""}`} />
            {t("loginHistory.refresh")}
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 sm:grid-cols-3">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="bg-success/10 rounded-full p-3">
                  <Check className="text-success h-5 w-5" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{loginHistory?.filter((l) => l.success).length || 0}</p>
                  <p className="text-muted-foreground text-sm">{t("loginHistory.stats.successful")}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="bg-destructive/10 rounded-full p-3">
                  <X className="text-destructive h-5 w-5" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{loginHistory?.filter((l) => !l.success).length || 0}</p>
                  <p className="text-muted-foreground text-sm">{t("loginHistory.stats.failed")}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="bg-info/10 rounded-full p-3">
                  <Globe className="text-info h-5 w-5" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{new Set(loginHistory?.map((l) => l.location)).size || 0}</p>
                  <p className="text-muted-foreground text-sm">{t("loginHistory.stats.locations")}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Login History List */}
        <Card>
          <CardHeader>
            <CardTitle>{t("loginHistory.recentActivity")}</CardTitle>
            <CardDescription>{t("loginHistory.recentActivityHint")}</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="bg-muted h-20 animate-pulse rounded-lg"
                  />
                ))}
              </div>
            ) : !loginHistory || loginHistory.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <Clock className="text-muted-foreground/50 mb-4 h-12 w-12" />
                <p className="text-muted-foreground">{t("loginHistory.noHistory")}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {loginHistory.map((entry) => (
                  <div
                    key={entry.id}
                    className={`flex items-start gap-4 rounded-lg border p-4 ${
                      !entry.success ? "border-destructive/30 bg-destructive/5" : ""
                    }`}
                  >
                    {/* Status Icon */}
                    <div
                      className={`rounded-full p-2 ${
                        entry.success ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive"
                      }`}
                    >
                      {entry.success ? <Check className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
                    </div>

                    {/* Main Content */}
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium">
                          {entry.success ? t("loginHistory.successful") : t("loginHistory.failedAttempt")}
                        </span>
                        {!entry.success && entry.failure_reason && (
                          <Badge
                            variant="destructive"
                            className="text-xs"
                          >
                            {entry.failure_reason}
                          </Badge>
                        )}
                      </div>

                      <div className="text-muted-foreground mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">
                        {/* Device */}
                        <div className="flex items-center gap-1">
                          {getDeviceIcon(entry.device_type)}
                          <span>{entry.device_type || t("loginHistory.unknown")}</span>
                        </div>

                        {/* Browser */}
                        <div className="flex items-center gap-1">
                          {getBrowserIcon(entry.browser)}
                          <span>{entry.browser || t("loginHistory.unknown")}</span>
                        </div>

                        {/* OS */}
                        <div className="flex items-center gap-1">
                          <Monitor className="h-4 w-4" />
                          <span>{entry.os || t("loginHistory.unknown")}</span>
                        </div>
                      </div>

                      <div className="text-muted-foreground mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">
                        {/* Location */}
                        <div className="flex items-center gap-1">
                          <MapPin className="h-4 w-4" />
                          <span>{entry.location || t("loginHistory.unknownLocation")}</span>
                        </div>

                        {/* IP Address */}
                        <div className="flex items-center gap-1">
                          <Globe className="h-4 w-4" />
                          <span>{entry.ip_address}</span>
                        </div>
                      </div>
                    </div>

                    {/* Time */}
                    <div className="shrink-0 text-right">
                      <div className="text-sm font-medium">{formatDate(entry.created_at)}</div>
                      <div className="text-muted-foreground text-xs">
                        {new Date(entry.created_at).toLocaleTimeString()}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </SettingsLayout>
  );
}
