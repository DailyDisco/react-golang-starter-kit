import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Check, Chrome, Clock, Globe, MapPin, Monitor, RefreshCw, Smartphone, X } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { requireAuth } from "../../lib/guards";
import { SettingsService, type LoginHistoryEntry } from "../../services/settings/settingsService";

export const Route = createFileRoute("/settings/login-history")({
  beforeLoad: () => requireAuth(),
  component: LoginHistoryPage,
});

function LoginHistoryPage() {
  const {
    data: loginHistory,
    isLoading,
    refetch,
    isRefetching,
  } = useQuery({
    queryKey: ["login-history"],
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
    // For simplicity, just return Chrome icon for all browsers
    // In production, you might want to add more browser icons
    return <Chrome className="h-4 w-4" />;
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins} minutes ago`;
    if (diffHours < 24) return `${diffHours} hours ago`;
    if (diffDays < 7) return `${diffDays} days ago`;

    return date.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
      hour: "numeric",
      minute: "2-digit",
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Login History</h2>
          <p className="text-sm text-gray-500">Review your recent login activity and detect suspicious access</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          disabled={isRefetching}
        >
          <RefreshCw className={`mr-2 h-4 w-4 ${isRefetching ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-green-100 p-3">
                <Check className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-bold">{loginHistory?.filter((l) => l.success).length || 0}</p>
                <p className="text-sm text-gray-500">Successful Logins</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-red-100 p-3">
                <X className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <p className="text-2xl font-bold">{loginHistory?.filter((l) => !l.success).length || 0}</p>
                <p className="text-sm text-gray-500">Failed Attempts</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="rounded-full bg-blue-100 p-3">
                <Globe className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-bold">{new Set(loginHistory?.map((l) => l.location)).size || 0}</p>
                <p className="text-sm text-gray-500">Unique Locations</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Login History List */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Your login history for the past 30 days</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <div
                  key={i}
                  className="h-20 animate-pulse rounded-lg bg-gray-100"
                />
              ))}
            </div>
          ) : !loginHistory || loginHistory.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Clock className="mb-4 h-12 w-12 text-gray-300" />
              <p className="text-gray-500">No login history available</p>
            </div>
          ) : (
            <div className="space-y-4">
              {loginHistory.map((entry) => (
                <div
                  key={entry.id}
                  className={`flex items-start gap-4 rounded-lg border p-4 ${
                    !entry.success ? "border-red-200 bg-red-50" : ""
                  }`}
                >
                  {/* Status Icon */}
                  <div
                    className={`rounded-full p-2 ${
                      entry.success ? "bg-green-100 text-green-600" : "bg-red-100 text-red-600"
                    }`}
                  >
                    {entry.success ? <Check className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
                  </div>

                  {/* Main Content */}
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-medium">{entry.success ? "Successful login" : "Failed login attempt"}</span>
                      {!entry.success && entry.failure_reason && (
                        <Badge
                          variant="destructive"
                          className="text-xs"
                        >
                          {entry.failure_reason}
                        </Badge>
                      )}
                    </div>

                    <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500">
                      {/* Device */}
                      <div className="flex items-center gap-1">
                        {getDeviceIcon(entry.device_type)}
                        <span>{entry.device_type || "Unknown"}</span>
                      </div>

                      {/* Browser */}
                      <div className="flex items-center gap-1">
                        {getBrowserIcon(entry.browser)}
                        <span>{entry.browser || "Unknown"}</span>
                      </div>

                      {/* OS */}
                      <div className="flex items-center gap-1">
                        <Monitor className="h-4 w-4" />
                        <span>{entry.os || "Unknown"}</span>
                      </div>
                    </div>

                    <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500">
                      {/* Location */}
                      <div className="flex items-center gap-1">
                        <MapPin className="h-4 w-4" />
                        <span>{entry.location || "Unknown location"}</span>
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
                    <div className="text-sm font-medium text-gray-900">{formatDate(entry.created_at)}</div>
                    <div className="text-xs text-gray-400">{new Date(entry.created_at).toLocaleTimeString()}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
