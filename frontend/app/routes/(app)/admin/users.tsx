import { useMemo, useState } from "react";

import { ConfirmDialog } from "@/components/admin";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { requireAdmin } from "@/lib/guards";
import { AdminService } from "@/services/admin";
import { apiClient } from "@/services/api/client";
import type { User } from "@/services/types";
import { useAuthStore } from "@/stores/auth-store";
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
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

interface UsersResponse {
  users: User[];
  count: number;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export const Route = createFileRoute("/(app)/admin/users")({
  beforeLoad: () => requireAdmin(),
  component: AdminUsersPage,
});

type SortField = "name" | "email" | "role" | "created_at";
type SortDirection = "asc" | "desc";

function SortHeader({
  field,
  children,
  onSort,
}: {
  field: SortField;
  children: React.ReactNode;
  onSort: (field: SortField) => void;
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      className="-ml-3 h-8 font-medium"
      onClick={() => onSort(field)}
    >
      {children}
      <ArrowUpDown className="ml-2 h-4 w-4" />
    </Button>
  );
}

function AdminUsersPage() {
  const { t } = useTranslation("admin");
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

  const filteredAndSortedUsers = useMemo(() => {
    if (!data?.users) return [];

    let filtered = data.users;

    if (search.trim()) {
      const searchLower = search.toLowerCase();
      filtered = filtered.filter(
        (user) =>
          user.name.toLowerCase().includes(searchLower) ||
          user.email.toLowerCase().includes(searchLower) ||
          user.role?.toLowerCase().includes(searchLower)
      );
    }

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
      toast.success(t("users.toast.impersonating", { name: response.user.name }));
      window.location.href = "/";
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("users.toast.impersonateError"));
    },
  });

  const deactivateMutation = useMutation({
    mutationFn: (userId: number) => AdminService.deactivateUser(userId),
    onSuccess: () => {
      toast.success(t("users.toast.deactivated"));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("users.toast.deactivateError"));
    },
  });

  const reactivateMutation = useMutation({
    mutationFn: (userId: number) => AdminService.reactivateUser(userId),
    onSuccess: () => {
      toast.success(t("users.toast.reactivated"));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("users.toast.reactivateError"));
    },
  });

  const bulkDeactivateMutation = useMutation({
    mutationFn: async (userIds: number[]) => {
      await Promise.all(userIds.map((id) => AdminService.deactivateUser(id)));
    },
    onSuccess: () => {
      toast.success(t("users.toast.bulkDeactivated", { count: selectedUsers.size }));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setSelectedUsers(new Set());
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("users.toast.bulkDeactivateError"));
    },
  });

  const bulkReactivateMutation = useMutation({
    mutationFn: async (userIds: number[]) => {
      await Promise.all(userIds.map((id) => AdminService.reactivateUser(id)));
    },
    onSuccess: () => {
      toast.success(t("users.toast.bulkReactivated", { count: selectedUsers.size }));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      setSelectedUsers(new Set());
      setConfirmDialog((prev) => ({ ...prev, open: false }));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("users.toast.bulkReactivateError"));
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

    const csvContent = [headers.join(","), ...rows.map((row) => row.map((cell) => `"${cell}"`).join(","))].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `users-export-${new Date().toISOString().split("T")[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success(t("users.toast.exported"));
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">{t("users.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("users.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={exportToCSV}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            <span className="hidden sm:inline">{t("users.export")}</span>
          </Button>
          <Button
            variant="outline"
            onClick={() => refetch()}
            className="gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            <span className="hidden sm:inline">{t("users.refresh")}</span>
          </Button>
        </div>
      </div>

      {/* Search and Bulk Actions */}
      <Card>
        <CardContent className="py-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="relative flex-1 sm:max-w-xs">
              <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
              <Input
                placeholder={t("users.searchPlaceholder")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>
            {selectedUsers.size > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground text-sm">
                  {t("users.selected", { count: selectedUsers.size })}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setConfirmDialog({
                      open: true,
                      title: t("users.dialog.deactivateTitle"),
                      description: t("users.dialog.bulkDeactivateDescription", { count: selectedUsers.size }),
                      action: () => bulkDeactivateMutation.mutate([...selectedUsers]),
                      variant: "destructive",
                    })
                  }
                  className="text-destructive hover:text-destructive"
                >
                  <Ban className="mr-2 h-4 w-4" />
                  {t("users.deactivate")}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setConfirmDialog({
                      open: true,
                      title: t("users.dialog.reactivateTitle"),
                      description: t("users.dialog.bulkReactivateDescription", { count: selectedUsers.size }),
                      action: () => bulkReactivateMutation.mutate([...selectedUsers]),
                      variant: "default",
                    })
                  }
                  className="text-success hover:text-success"
                >
                  <CheckCircle className="mr-2 h-4 w-4" />
                  {t("users.reactivate")}
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
              <RefreshCw className="text-muted-foreground h-6 w-6 animate-spin" />
              <span className="text-muted-foreground ml-2">{t("users.loading")}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Error State */}
      {error && (
        <Card className="border-destructive/30 bg-destructive/5">
          <CardHeader>
            <CardTitle className="text-destructive">{t("users.error.title")}</CardTitle>
            <CardDescription className="text-destructive/80">
              {error instanceof Error ? error.message : t("users.error.default")}
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
                        filteredAndSortedUsers.length > 0 && selectedUsers.size === filteredAndSortedUsers.length
                      }
                      onCheckedChange={handleSelectAll}
                    />
                  </TableHead>
                  <TableHead>
                    <SortHeader
                      field="name"
                      onSort={handleSort}
                    >
                      {t("users.table.user")}
                    </SortHeader>
                  </TableHead>
                  <TableHead>
                    <SortHeader
                      field="role"
                      onSort={handleSort}
                    >
                      {t("users.table.role")}
                    </SortHeader>
                  </TableHead>
                  <TableHead>{t("users.table.status")}</TableHead>
                  <TableHead>
                    <SortHeader
                      field="created_at"
                      onSort={handleSort}
                    >
                      {t("users.table.created")}
                    </SortHeader>
                  </TableHead>
                  <TableHead className="text-right">{t("users.table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredAndSortedUsers.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className="h-24 text-center"
                    >
                      <div className="text-muted-foreground flex flex-col items-center justify-center">
                        <UsersIcon className="mb-2 h-8 w-8" />
                        <p>{t("users.empty.title")}</p>
                        {search && <p className="text-sm">{t("users.empty.searchHint")}</p>}
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredAndSortedUsers.map((user) => (
                    <TableRow
                      key={user.id}
                      data-state={selectedUsers.has(user.id) && "selected"}
                    >
                      <TableCell>
                        <Checkbox
                          checked={selectedUsers.has(user.id)}
                          onCheckedChange={() => handleSelectUser(user.id)}
                        />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar className="h-9 w-9">
                            <AvatarFallback>{user.name.charAt(0).toUpperCase()}</AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">{user.name}</p>
                            <p className="text-muted-foreground text-sm">{user.email}</p>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={getRoleBadgeVariant(user.role)}>{user.role || "user"}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {user.is_active ? (
                            <Badge variant="success">
                              <CheckCircle className="mr-1 h-3 w-3" />
                              {t("users.status.active")}
                            </Badge>
                          ) : (
                            <Badge variant="destructive">
                              <Ban className="mr-1 h-3 w-3" />
                              {t("users.status.inactive")}
                            </Badge>
                          )}
                          {user.email_verified && <Badge variant="info">{t("users.status.verified")}</Badge>}
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
                              <TooltipContent>{t("users.tooltips.impersonate")}</TooltipContent>
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
                                        title: t("users.dialog.deactivateTitle"),
                                        description: t("users.dialog.deactivateDescription", { name: user.name }),
                                        action: () => deactivateMutation.mutate(user.id),
                                        variant: "destructive",
                                      })
                                    }
                                    className="text-destructive hover:text-destructive"
                                  >
                                    <Ban className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>{t("users.tooltips.deactivate")}</TooltipContent>
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
                                        title: t("users.dialog.reactivateTitle"),
                                        description: t("users.dialog.reactivateDescription", { name: user.name }),
                                        action: () => reactivateMutation.mutate(user.id),
                                        variant: "default",
                                      })
                                    }
                                    className="text-success hover:text-success"
                                  >
                                    <CheckCircle className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>{t("users.tooltips.reactivate")}</TooltipContent>
                              </Tooltip>
                            ))}
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => toast.info(t("users.toast.detailsComingSoon"))}
                              >
                                <UserCog className="h-4 w-4" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>{t("users.tooltips.viewDetails")}</TooltipContent>
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
              <p className="text-muted-foreground text-sm">
                {t("users.pagination.showing", {
                  from: (data.page - 1) * data.limit + 1,
                  to: Math.min(data.page * data.limit, data.total),
                  total: data.total,
                })}
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  <ChevronLeft className="mr-1 h-4 w-4" />
                  {t("users.pagination.previous")}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  {t("users.pagination.next")}
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
