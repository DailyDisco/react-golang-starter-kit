import { useMemo, useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  ArrowUpDown,
  Ban,
  CheckCircle,
  ChevronLeft,
  ChevronRight,
  Download,
  RefreshCw,
  Search,
  Shield,
  UserCog,
  Users as UsersIcon,
} from "lucide-react";
import { toast } from "sonner";

import { AdminPageHeader, ConfirmDialog } from "../../components/admin";
import { Avatar, AvatarFallback } from "../../components/ui/avatar";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Checkbox } from "../../components/ui/checkbox";
import { Input } from "../../components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "../../components/ui/tooltip";
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

type SortField = "name" | "email" | "role" | "created_at";
type SortDirection = "asc" | "desc";

function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [selectedUsers, setSelectedUsers] = useState<Set<number>>(new Set());
  const [sortField, setSortField] = useState<SortField>("created_at");
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    title: string;
    description: string;
    action: () => void;
    variant: "default" | "destructive";
  }>({ open: false, title: "", description: "", action: () => {}, variant: "default" });

  const { user: currentUser, login } = useAuthStore();
  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery<UsersResponse>({
    queryKey: ["users", page, search],
    queryFn: async () => {
      const result = await apiClient.get<{ data: UsersResponse }>(`/users?page=${page}&limit=20`);
      return (result as unknown as { data: UsersResponse }).data || (result as unknown as UsersResponse);
    },
  });

  // Filter and sort users
  const filteredAndSortedUsers = useMemo(() => {
    if (!data?.users) return [];

    let filtered = data.users;

    // Filter by search
    if (search.trim()) {
      const searchLower = search.toLowerCase();
      filtered = filtered.filter(
        (user) =>
          user.name.toLowerCase().includes(searchLower) ||
          user.email.toLowerCase().includes(searchLower) ||
          (user.role && user.role.toLowerCase().includes(searchLower))
      );
    }

    // Sort
    filtered.sort((a, b) => {
      let comparison = 0;
      switch (sortField) {
        case "name":
          comparison = a.name.localeCompare(b.name);
          break;
        case "email":
          comparison = a.email.localeCompare(b.email);
          break;
        case "role":
          comparison = (a.role || "user").localeCompare(b.role || "user");
          break;
        case "created_at":
          comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
      }
      return sortDirection === "asc" ? comparison : -comparison;
    });

    return filtered;
  }, [data?.users, search, sortField, sortDirection]);

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
      setConfirmDialog((prev) => ({ ...prev, open: false }));
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
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to reactivate");
    },
  });

  // Bulk actions
  const bulkDeactivateMutation = useMutation({
    mutationFn: async (userIds: number[]) => {
      await Promise.all(userIds.map((id) => AdminService.deactivateUser(id)));
    },
    onSuccess: () => {
      toast.success(`${selectedUsers.size} users deactivated`);
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setSelectedUsers(new Set());
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to deactivate users");
    },
  });

  const bulkReactivateMutation = useMutation({
    mutationFn: async (userIds: number[]) => {
      await Promise.all(userIds.map((id) => AdminService.reactivateUser(id)));
    },
    onSuccess: () => {
      toast.success(`${selectedUsers.size} users reactivated`);
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setSelectedUsers(new Set());
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to reactivate users");
    },
  });

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortDirection("asc");
    }
  };

  const handleSelectAll = () => {
    if (selectedUsers.size === filteredAndSortedUsers.length) {
      setSelectedUsers(new Set());
    } else {
      setSelectedUsers(new Set(filteredAndSortedUsers.map((u) => u.id)));
    }
  };

  const handleSelectUser = (userId: number) => {
    const newSelected = new Set(selectedUsers);
    if (newSelected.has(userId)) {
      newSelected.delete(userId);
    } else {
      newSelected.add(userId);
    }
    setSelectedUsers(newSelected);
  };

  const exportToCSV = () => {
    if (!data?.users) return;

    const headers = ["ID", "Name", "Email", "Role", "Status", "Verified", "Created At"];
    const rows = data.users.map((user) => [
      user.id,
      user.name,
      user.email,
      user.role || "user",
      user.is_active ? "Active" : "Inactive",
      user.email_verified ? "Yes" : "No",
      new Date(user.created_at).toISOString(),
    ]);

    const csvContent = [
      headers.join(","),
      ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `users-export-${new Date().toISOString().split("T")[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success("Users exported to CSV");
  };

  const getRoleBadgeVariant = (role?: string) => {
    switch (role) {
      case "super_admin":
        return "default";
      case "admin":
        return "secondary";
      case "premium":
        return "outline";
      default:
        return "outline";
    }
  };

  const SortHeader = ({ field, children }: { field: SortField; children: React.ReactNode }) => (
    <Button
      variant="ghost"
      size="sm"
      className="-ml-3 h-8 font-medium"
      onClick={() => handleSort(field)}
    >
      {children}
      <ArrowUpDown className="ml-2 h-4 w-4" />
    </Button>
  );

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="User Management"
        description="Manage users, roles, and permissions"
        breadcrumbs={[{ label: "Users" }]}
        actions={
          <div className="flex gap-2">
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

      {/* Search and Bulk Actions */}
      <Card>
        <CardContent className="py-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="relative flex-1 sm:max-w-xs">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <Input
                placeholder="Search users..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>
            {selectedUsers.size > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-500">{selectedUsers.size} selected</span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setConfirmDialog({
                      open: true,
                      title: "Deactivate Users",
                      description: `Are you sure you want to deactivate ${selectedUsers.size} users? They will no longer be able to access the platform.`,
                      action: () => bulkDeactivateMutation.mutate(Array.from(selectedUsers)),
                      variant: "destructive",
                    })
                  }
                  className="text-red-600 hover:text-red-700"
                >
                  <Ban className="mr-2 h-4 w-4" />
                  Deactivate
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setConfirmDialog({
                      open: true,
                      title: "Reactivate Users",
                      description: `Are you sure you want to reactivate ${selectedUsers.size} users?`,
                      action: () => bulkReactivateMutation.mutate(Array.from(selectedUsers)),
                      variant: "default",
                    })
                  }
                  className="text-green-600 hover:text-green-700"
                >
                  <CheckCircle className="mr-2 h-4 w-4" />
                  Reactivate
                </Button>
              </div>
            )}
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
        <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950">
          <CardHeader>
            <CardTitle className="text-red-600">Error</CardTitle>
            <CardDescription className="text-red-500">
              {error instanceof Error ? error.message : "Failed to load users"}
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {/* Users Table */}
      {data && (
        <>
          <Card>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <Checkbox
                      checked={
                        filteredAndSortedUsers.length > 0 &&
                        selectedUsers.size === filteredAndSortedUsers.length
                      }
                      onCheckedChange={handleSelectAll}
                    />
                  </TableHead>
                  <TableHead>
                    <SortHeader field="name">User</SortHeader>
                  </TableHead>
                  <TableHead>
                    <SortHeader field="role">Role</SortHeader>
                  </TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>
                    <SortHeader field="created_at">Created</SortHeader>
                  </TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredAndSortedUsers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="h-24 text-center">
                      <div className="flex flex-col items-center justify-center text-gray-500">
                        <UsersIcon className="mb-2 h-8 w-8" />
                        <p>No users found</p>
                        {search && <p className="text-sm">Try adjusting your search</p>}
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredAndSortedUsers.map((user) => (
                    <TableRow key={user.id} data-state={selectedUsers.has(user.id) && "selected"}>
                      <TableCell>
                        <Checkbox
                          checked={selectedUsers.has(user.id)}
                          onCheckedChange={() => handleSelectUser(user.id)}
                        />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar className="h-9 w-9">
                            <AvatarFallback>
                              {user.name.charAt(0).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">{user.name}</p>
                            <p className="text-sm text-muted-foreground">{user.email}</p>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={getRoleBadgeVariant(user.role)}>
                          {user.role || "user"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {user.is_active ? (
                            <Badge variant="outline" className="border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950 dark:text-green-400">
                              <CheckCircle className="mr-1 h-3 w-3" />
                              Active
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
                              <Ban className="mr-1 h-3 w-3" />
                              Inactive
                            </Badge>
                          )}
                          {user.email_verified && (
                            <Badge variant="outline" className="border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-400">
                              Verified
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(user.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          {user.id !== currentUser?.id && user.role !== "super_admin" && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => impersonateMutation.mutate(user.id)}
                                  disabled={impersonateMutation.isPending}
                                >
                                  <Shield className="h-4 w-4" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>Impersonate user</TooltipContent>
                            </Tooltip>
                          )}
                          {user.id !== currentUser?.id &&
                            (user.is_active ? (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() =>
                                      setConfirmDialog({
                                        open: true,
                                        title: "Deactivate User",
                                        description: `Are you sure you want to deactivate ${user.name}? They will no longer be able to access the platform.`,
                                        action: () => deactivateMutation.mutate(user.id),
                                        variant: "destructive",
                                      })
                                    }
                                    className="text-red-600 hover:text-red-700"
                                  >
                                    <Ban className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>Deactivate user</TooltipContent>
                              </Tooltip>
                            ) : (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() =>
                                      setConfirmDialog({
                                        open: true,
                                        title: "Reactivate User",
                                        description: `Are you sure you want to reactivate ${user.name}?`,
                                        action: () => reactivateMutation.mutate(user.id),
                                        variant: "default",
                                      })
                                    }
                                    className="text-green-600 hover:text-green-700"
                                  >
                                    <CheckCircle className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>Reactivate user</TooltipContent>
                              </Tooltip>
                            ))}
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => toast.info(`User details page for ${user.name} coming soon`)}
                              >
                                <UserCog className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>View user details</TooltipContent>
                          </Tooltip>
                        </div>
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
                {data.total} users
              </p>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  <ChevronLeft className="mr-1 h-4 w-4" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                  <ChevronRight className="ml-1 h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Confirm Dialog */}
      <ConfirmDialog
        open={confirmDialog.open}
        onOpenChange={(open) => setConfirmDialog((prev) => ({ ...prev, open }))}
        title={confirmDialog.title}
        description={confirmDialog.description}
        onConfirm={confirmDialog.action}
        variant={confirmDialog.variant}
        loading={
          deactivateMutation.isPending ||
          reactivateMutation.isPending ||
          bulkDeactivateMutation.isPending ||
          bulkReactivateMutation.isPending
        }
      />
    </div>
  );
}
