import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight, Filter, RefreshCw } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
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
      page: key !== "page" ? 1 : (value as number), // Reset to page 1 when filter changes
    }));
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case "create":
        return "bg-green-100 text-green-800";
      case "update":
        return "bg-blue-100 text-blue-800";
      case "delete":
        return "bg-red-100 text-red-800";
      case "login":
        return "bg-purple-100 text-purple-800";
      case "logout":
        return "bg-gray-100 text-gray-800";
      case "impersonate":
        return "bg-yellow-100 text-yellow-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Audit Logs</h2>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="mr-2 h-4 w-4" />
            Filters
          </Button>
          <Button
            variant="outline"
            onClick={() => refetch()}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Filters</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
              <div>
                <Label htmlFor="target_type">Target Type</Label>
                <select
                  id="target_type"
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  value={filter.target_type || ""}
                  onChange={(e) => handleFilterChange("target_type", e.target.value || undefined)}
                >
                  <option value="">All</option>
                  <option value="user">User</option>
                  <option value="subscription">Subscription</option>
                  <option value="file">File</option>
                  <option value="feature_flag">Feature Flag</option>
                </select>
              </div>
              <div>
                <Label htmlFor="action">Action</Label>
                <select
                  id="action"
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  value={filter.action || ""}
                  onChange={(e) => handleFilterChange("action", e.target.value || undefined)}
                >
                  <option value="">All</option>
                  <option value="create">Create</option>
                  <option value="update">Update</option>
                  <option value="delete">Delete</option>
                  <option value="login">Login</option>
                  <option value="logout">Logout</option>
                  <option value="impersonate">Impersonate</option>
                  <option value="role_change">Role Change</option>
                </select>
              </div>
              <div>
                <Label htmlFor="start_date">Start Date</Label>
                <Input
                  id="start_date"
                  type="datetime-local"
                  value={filter.start_date || ""}
                  onChange={(e) => handleFilterChange("start_date", e.target.value || undefined)}
                />
              </div>
              <div>
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
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-red-600">Error</CardTitle>
            <CardDescription className="text-red-500">
              {error instanceof Error ? error.message : "Failed to load audit logs"}
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {/* Audit Logs Table */}
      {data && (
        <>
          <Card>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="border-b bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Time</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">User</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Action</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Target</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">IP Address</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {data.logs.map((log) => (
                      <tr
                        key={log.id}
                        className="hover:bg-gray-50"
                      >
                        <td className="px-4 py-3 text-sm text-gray-500">{formatDate(log.created_at)}</td>
                        <td className="px-4 py-3">
                          {log.user_name ? (
                            <div>
                              <p className="text-sm font-medium text-gray-900">{log.user_name}</p>
                              <p className="text-xs text-gray-500">{log.user_email}</p>
                            </div>
                          ) : (
                            <span className="text-sm text-gray-400">System</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${getActionColor(log.action)}`}
                          >
                            {log.action}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-sm">
                          <span className="text-gray-900">{log.target_type}</span>
                          {log.target_id && <span className="ml-1 text-gray-500">#{log.target_id}</span>}
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-500">{log.ip_address || "-"}</td>
                      </tr>
                    ))}
                    {data.logs.length === 0 && (
                      <tr>
                        <td
                          colSpan={5}
                          className="px-4 py-8 text-center text-gray-500"
                        >
                          No audit logs found
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Pagination */}
          {data.total_pages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-500">
                Showing {(data.page - 1) * data.limit + 1} to {Math.min(data.page * data.limit, data.total)} of{" "}
                {data.total} results
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  disabled={data.page <= 1}
                  onClick={() => handleFilterChange("page", data.page - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  disabled={data.page >= data.total_pages}
                  onClick={() => handleFilterChange("page", data.page + 1)}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
