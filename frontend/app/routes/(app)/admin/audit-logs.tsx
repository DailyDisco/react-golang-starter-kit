import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { AdminLayout } from "@/layouts/AdminLayout";
import { AdminService, type AuditLogFilter, type AuditLogsResponse } from "@/services/admin";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight, Download, FileText, Filter, RefreshCw, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/audit-logs")({
  component: AuditLogsPage,
});

function AuditLogsPage() {
  const { t } = useTranslation("admin");
  const [filter, setFilter] = useState<AuditLogFilter>({
    page: 1,
    limit: 20,
  });
  const [showFilters, setShowFilters] = useState(false);

  const { data, isLoading, error, refetch } = useQuery<AuditLogsResponse>({
    queryKey: ["admin", "audit-logs", filter],
    queryFn: () => AdminService.getAuditLogs(filter),
  });

  const handleFilterChange = (key: keyof AuditLogFilter, value: string | number | undefined) => {
    setFilter((prev) => ({
      ...prev,
      [key]: value,
      page: key !== "page" ? 1 : (value as number),
    }));
  };

  const clearFilters = () => {
    setFilter({ page: 1, limit: 20 });
  };

  const hasActiveFilters = filter.target_type || filter.action || filter.start_date || filter.end_date;

  const exportToCSV = () => {
    if (!data?.logs) return;

    const headers = ["ID", "Time", "User", "Email", "Action", "Target Type", "Target ID", "IP Address"];
    const rows = data.logs.map((log) => [
      log.id,
      new Date(log.created_at).toISOString(),
      log.user_name || t("auditLogs.system"),
      log.user_email || "-",
      log.action,
      log.target_type,
      log.target_id || "-",
      log.ip_address || "-",
    ]);

    const csvContent = [headers.join(","), ...rows.map((row) => row.map((cell) => `"${cell}"`).join(","))].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split("T")[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success(t("auditLogs.toast.exported"));
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getActionBadgeVariant = (action: string): "default" | "secondary" | "destructive" | "success" | "info" => {
    switch (action) {
      case "create":
        return "success";
      case "delete":
        return "destructive";
      case "update":
        return "info";
      default:
        return "secondary";
    }
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">{t("auditLogs.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("auditLogs.subtitle")}</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="gap-2"
            >
              <Filter className="h-4 w-4" />
              <span className="hidden sm:inline">{t("auditLogs.filters")}</span>
              {hasActiveFilters && (
                <Badge
                  variant="secondary"
                  className="ml-1"
                >
                  {t("auditLogs.filtersActive")}
                </Badge>
              )}
            </Button>
            <Button
              variant="outline"
              onClick={exportToCSV}
              className="gap-2"
            >
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">{t("auditLogs.export")}</span>
            </Button>
            <Button
              variant="outline"
              onClick={() => refetch()}
              className="gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              <span className="hidden sm:inline">{t("auditLogs.refresh")}</span>
            </Button>
          </div>
        </div>

        {/* Filters */}
        {showFilters && (
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{t("auditLogs.filters")}</CardTitle>
                {hasActiveFilters && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={clearFilters}
                    className="gap-2"
                  >
                    <X className="h-4 w-4" />
                    {t("auditLogs.clearAll")}
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
                <div className="space-y-2">
                  <Label>{t("auditLogs.filter.targetType")}</Label>
                  <Select
                    value={filter.target_type || "all"}
                    onValueChange={(value) => handleFilterChange("target_type", value === "all" ? undefined : value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("auditLogs.filter.types.all")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">{t("auditLogs.filter.types.all")}</SelectItem>
                      <SelectItem value="user">{t("auditLogs.filter.types.user")}</SelectItem>
                      <SelectItem value="subscription">{t("auditLogs.filter.types.subscription")}</SelectItem>
                      <SelectItem value="file">{t("auditLogs.filter.types.file")}</SelectItem>
                      <SelectItem value="feature_flag">{t("auditLogs.filter.types.featureFlag")}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t("auditLogs.filter.action")}</Label>
                  <Select
                    value={filter.action || "all"}
                    onValueChange={(value) => handleFilterChange("action", value === "all" ? undefined : value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("auditLogs.filter.actions.all")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">{t("auditLogs.filter.actions.all")}</SelectItem>
                      <SelectItem value="create">{t("auditLogs.filter.actions.create")}</SelectItem>
                      <SelectItem value="update">{t("auditLogs.filter.actions.update")}</SelectItem>
                      <SelectItem value="delete">{t("auditLogs.filter.actions.delete")}</SelectItem>
                      <SelectItem value="login">{t("auditLogs.filter.actions.login")}</SelectItem>
                      <SelectItem value="logout">{t("auditLogs.filter.actions.logout")}</SelectItem>
                      <SelectItem value="impersonate">{t("auditLogs.filter.actions.impersonate")}</SelectItem>
                      <SelectItem value="role_change">{t("auditLogs.filter.actions.roleChange")}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="start_date">{t("auditLogs.filter.startDate")}</Label>
                  <Input
                    id="start_date"
                    type="datetime-local"
                    value={filter.start_date || ""}
                    onChange={(e) => handleFilterChange("start_date", e.target.value || undefined)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="end_date">{t("auditLogs.filter.endDate")}</Label>
                  <Input
                    id="end_date"
                    type="datetime-local"
                    value={filter.end_date || ""}
                    onChange={(e) => handleFilterChange("end_date", e.target.value || undefined)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Loading State */}
        {isLoading && (
          <Card>
            <CardContent className="py-8">
              <div className="flex items-center justify-center">
                <RefreshCw className="text-muted-foreground h-6 w-6 animate-spin" />
                <span className="text-muted-foreground ml-2">{t("auditLogs.loading")}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Error State */}
        {error && (
          <Card className="border-destructive/30 bg-destructive/5">
            <CardHeader>
              <CardTitle className="text-destructive">{t("auditLogs.error.title")}</CardTitle>
              <CardDescription className="text-destructive/80">
                {error instanceof Error ? error.message : t("auditLogs.error.default")}
              </CardDescription>
            </CardHeader>
          </Card>
        )}

        {/* Audit Logs Table */}
        {data && (
          <>
            <Card>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("auditLogs.table.time")}</TableHead>
                    <TableHead>{t("auditLogs.table.user")}</TableHead>
                    <TableHead>{t("auditLogs.table.action")}</TableHead>
                    <TableHead>{t("auditLogs.table.target")}</TableHead>
                    <TableHead>{t("auditLogs.table.ipAddress")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.logs.length === 0 ? (
                    <TableRow>
                      <TableCell
                        colSpan={5}
                        className="h-24 text-center"
                      >
                        <div className="text-muted-foreground flex flex-col items-center justify-center">
                          <FileText className="mb-2 h-8 w-8" />
                          <p>{t("auditLogs.empty.title")}</p>
                          {hasActiveFilters && <p className="text-sm">{t("auditLogs.empty.filterHint")}</p>}
                        </div>
                      </TableCell>
                    </TableRow>
                  ) : (
                    data.logs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell className="text-muted-foreground">{formatDate(log.created_at)}</TableCell>
                        <TableCell>
                          {log.user_name ? (
                            <div>
                              <p className="font-medium">{log.user_name}</p>
                              <p className="text-muted-foreground text-sm">{log.user_email}</p>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">{t("auditLogs.system")}</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant={getActionBadgeVariant(log.action)}>{log.action}</Badge>
                        </TableCell>
                        <TableCell>
                          <span className="font-medium">{log.target_type}</span>
                          {log.target_id && <span className="text-muted-foreground ml-1">#{log.target_id}</span>}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{log.ip_address || "-"}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </Card>

            {/* Pagination */}
            {data.total_pages > 1 && (
              <div className="flex items-center justify-between">
                <p className="text-muted-foreground text-sm">
                  {t("auditLogs.pagination.showing", {
                    from: (data.page - 1) * data.limit + 1,
                    to: Math.min(data.page * data.limit, data.total),
                    total: data.total,
                  })}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={data.page <= 1}
                    onClick={() => handleFilterChange("page", data.page - 1)}
                  >
                    <ChevronLeft className="mr-1 h-4 w-4" />
                    {t("auditLogs.pagination.previous")}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={data.page >= data.total_pages}
                    onClick={() => handleFilterChange("page", data.page + 1)}
                  >
                    {t("auditLogs.pagination.next")}
                    <ChevronRight className="ml-1 h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </AdminLayout>
  );
}
