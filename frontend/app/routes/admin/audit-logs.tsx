import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight, Download, FileText, Filter, RefreshCw, X } from "lucide-react";
import { toast } from "sonner";

import { AdminPageHeader } from "../../components/admin";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../components/ui/table";
import { requireAdmin } from "../../lib/guards";
import { AdminService, type AuditLogFilter, type AuditLogsResponse } from "../../services/admin";

export const Route = createFileRoute("/admin/audit-logs")({
  beforeLoad: () => requireAdmin(),
  component: AuditLogsPage,
});

function AuditLogsPage() {
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
      log.user_name || "System",
      log.user_email || "-",
      log.action,
      log.target_type,
      log.target_id || "-",
      log.ip_address || "-",
    ]);

    const csvContent = [
      headers.join(","),
      ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split("T")[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success("Audit logs exported to CSV");
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getActionBadgeVariant = (action: string): "default" | "secondary" | "outline" | "destructive" => {
    switch (action) {
      case "create":
        return "default";
      case "delete":
        return "destructive";
      case "update":
        return "secondary";
      default:
        return "outline";
    }
  };

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Audit Logs"
        description="Track all system activity and user actions"
        breadcrumbs={[{ label: "Audit Logs" }]}
        actions={
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="gap-2"
            >
              <Filter className="h-4 w-4" />
              <span className="hidden sm:inline">Filters</span>
              {hasActiveFilters && (
                <Badge variant="secondary" className="ml-1">
                  Active
                </Badge>
              )}
            </Button>
            <Button variant="outline" onClick={exportToCSV} className="gap-2">
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">Export</span>
            </Button>
            <Button variant="outline" onClick={() => refetch()} className="gap-2">
              <RefreshCw className="h-4 w-4" />
              <span className="hidden sm:inline">Refresh</span>
            </Button>
          </div>
        }
      />

      {/* Filters */}
      {showFilters && (
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">Filters</CardTitle>
              {hasActiveFilters && (
                <Button variant="ghost" size="sm" onClick={clearFilters} className="gap-2">
                  <X className="h-4 w-4" />
                  Clear all
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
              <div className="space-y-2">
                <Label>Target Type</Label>
                <Select
                  value={filter.target_type || "all"}
                  onValueChange={(value) => handleFilterChange("target_type", value === "all" ? undefined : value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All types" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All</SelectItem>
                    <SelectItem value="user">User</SelectItem>
                    <SelectItem value="subscription">Subscription</SelectItem>
                    <SelectItem value="file">File</SelectItem>
                    <SelectItem value="feature_flag">Feature Flag</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Action</Label>
                <Select
                  value={filter.action || "all"}
                  onValueChange={(value) => handleFilterChange("action", value === "all" ? undefined : value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All actions" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All</SelectItem>
                    <SelectItem value="create">Create</SelectItem>
                    <SelectItem value="update">Update</SelectItem>
                    <SelectItem value="delete">Delete</SelectItem>
                    <SelectItem value="login">Login</SelectItem>
                    <SelectItem value="logout">Logout</SelectItem>
                    <SelectItem value="impersonate">Impersonate</SelectItem>
                    <SelectItem value="role_change">Role Change</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="start_date">Start Date</Label>
                <Input
                  id="start_date"
                  type="datetime-local"
                  value={filter.start_date || ""}
                  onChange={(e) => handleFilterChange("start_date", e.target.value || undefined)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="end_date">End Date</Label>
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
              <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
              <span className="ml-2 text-gray-500">Loading audit logs...</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Error State */}
      {error && (
        <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950">
          <CardHeader>
            <CardTitle className="text-red-600 dark:text-red-400">Error</CardTitle>
            <CardDescription className="text-red-500 dark:text-red-400">
              {error instanceof Error ? error.message : "Failed to load audit logs"}
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
                  <TableHead>Time</TableHead>
                  <TableHead>User</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>IP Address</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.logs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="h-24 text-center">
                      <div className="flex flex-col items-center justify-center text-gray-500">
                        <FileText className="mb-2 h-8 w-8" />
                        <p>No audit logs found</p>
                        {hasActiveFilters && <p className="text-sm">Try adjusting your filters</p>}
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  data.logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-muted-foreground">
                        {formatDate(log.created_at)}
                      </TableCell>
                      <TableCell>
                        {log.user_name ? (
                          <div>
                            <p className="font-medium">{log.user_name}</p>
                            <p className="text-sm text-muted-foreground">{log.user_email}</p>
                          </div>
                        ) : (
                          <span className="text-muted-foreground">System</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={getActionBadgeVariant(log.action)}>
                          {log.action}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="font-medium">{log.target_type}</span>
                        {log.target_id && (
                          <span className="ml-1 text-muted-foreground">#{log.target_id}</span>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {log.ip_address || "-"}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </Card>

          {/* Pagination */}
          {data.total_pages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Showing {(data.page - 1) * data.limit + 1} to {Math.min(data.page * data.limit, data.total)} of{" "}
                {data.total} results
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={data.page <= 1}
                  onClick={() => handleFilterChange("page", data.page - 1)}
                >
                  <ChevronLeft className="mr-1 h-4 w-4" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={data.page >= data.total_pages}
                  onClick={() => handleFilterChange("page", data.page + 1)}
                >
                  Next
                  <ChevronRight className="ml-1 h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
