import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Ban, CheckCircle, RefreshCw, Shield, UserCog, Users as UsersIcon } from "lucide-react";
import { toast } from "sonner";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { requireAdmin } from "../../lib/guards";
import { AdminService } from "../../services/admin";
import { apiClient } from "../../services/api/client";
import type { User } from "../../services/types";
import { useAuthStore } from "../../stores/auth-store";

interface UsersResponse {
  users: User[];
  count: number;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export const Route = createFileRoute("/admin/users")({
  beforeLoad: () => requireAdmin(),
  component: AdminUsersPage,
});

function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const { user: currentUser, login } = useAuthStore();
  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery<UsersResponse>({
    queryKey: ["users", page, search],
    queryFn: async () => {
      const result = await apiClient.get<{ data: UsersResponse }>(`/users?page=${page}&limit=20`);
      // Handle both direct response and wrapped response formats
      return (result as unknown as { data: UsersResponse }).data || (result as unknown as UsersResponse);
    },
  });

  const impersonateMutation = useMutation({
    mutationFn: (userId: number) => AdminService.impersonateUser({ user_id: userId }),
    onSuccess: (response) => {
      login(response.user);
      localStorage.setItem("impersonating", "true");
      localStorage.setItem("original_user_id", response.original_user_id.toString());
      toast.success(`Now impersonating ${response.user.name}`);
      window.location.href = "/";
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to impersonate");
    },
  });

  const deactivateMutation = useMutation({
    mutationFn: (userId: number) => AdminService.deactivateUser(userId),
    onSuccess: () => {
      toast.success("User deactivated");
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to deactivate");
    },
  });

  const reactivateMutation = useMutation({
    mutationFn: (userId: number) => AdminService.reactivateUser(userId),
    onSuccess: () => {
      toast.success("User reactivated");
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to reactivate");
    },
  });

  const getRoleBadgeColor = (role?: string) => {
    switch (role) {
      case "super_admin":
        return "bg-purple-100 text-purple-800";
      case "admin":
        return "bg-blue-100 text-blue-800";
      case "premium":
        return "bg-yellow-100 text-yellow-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">User Management</h2>
        <Button
          variant="outline"
          onClick={() => refetch()}
        >
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      {/* Search */}
      <Card>
        <CardContent className="py-4">
          <div className="flex gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search users..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Loading State */}
      {isLoading && (
        <Card>
          <CardContent className="py-8">
            <div className="flex items-center justify-center">
              <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
              <span className="ml-2 text-gray-500">Loading users...</span>
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
              {error instanceof Error ? error.message : "Failed to load users"}
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {/* Users List */}
      {data && (
        <>
          <Card>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="border-b bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">User</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Role</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Status</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-500">Created</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-gray-500">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {data.users.map((user) => (
                      <tr
                        key={user.id}
                        className="hover:bg-gray-50"
                      >
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200">
                              <span className="text-sm font-medium text-gray-600">
                                {user.name.charAt(0).toUpperCase()}
                              </span>
                            </div>
                            <div>
                              <p className="font-medium text-gray-900">{user.name}</p>
                              <p className="text-sm text-gray-500">{user.email}</p>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${getRoleBadgeColor(user.role)}`}
                          >
                            {user.role || "user"}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            {user.is_active ? (
                              <span className="flex items-center text-sm text-green-600">
                                <CheckCircle className="mr-1 h-4 w-4" />
                                Active
                              </span>
                            ) : (
                              <span className="flex items-center text-sm text-red-600">
                                <Ban className="mr-1 h-4 w-4" />
                                Inactive
                              </span>
                            )}
                            {user.email_verified && (
                              <span className="rounded bg-blue-50 px-2 py-0.5 text-xs text-blue-600">Verified</span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-500">
                          {new Date(user.created_at).toLocaleDateString()}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex justify-end gap-2">
                            {/* Can't impersonate self or other super_admins */}
                            {user.id !== currentUser?.id && user.role !== "super_admin" && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => impersonateMutation.mutate(user.id)}
                                disabled={impersonateMutation.isPending}
                                title="Impersonate user"
                              >
                                <Shield className="h-4 w-4" />
                              </Button>
                            )}
                            {/* Toggle active status */}
                            {user.id !== currentUser?.id &&
                              (user.is_active ? (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => deactivateMutation.mutate(user.id)}
                                  disabled={deactivateMutation.isPending}
                                  title="Deactivate user"
                                  className="text-red-600 hover:text-red-700"
                                >
                                  <Ban className="h-4 w-4" />
                                </Button>
                              ) : (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => reactivateMutation.mutate(user.id)}
                                  disabled={reactivateMutation.isPending}
                                  title="Reactivate user"
                                  className="text-green-600 hover:text-green-700"
                                >
                                  <CheckCircle className="h-4 w-4" />
                                </Button>
                              ))}
                            <Link to={`/admin/users/${user.id}`}>
                              <Button
                                variant="outline"
                                size="sm"
                                title="Manage user"
                              >
                                <UserCog className="h-4 w-4" />
                              </Button>
                            </Link>
                          </div>
                        </td>
                      </tr>
                    ))}
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
                {data.total} users
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  disabled={page >= data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
