import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { queryKeys } from "@/lib/query-keys";
import { orgMembersQueryOptions } from "@/lib/route-query-options";
import {
  OrganizationService,
  type OrganizationInvitation,
  type OrganizationMember,
} from "@/services/organizations/organizationService";
import { useAuthStore } from "@/stores/auth-store";
import { useCurrentOrg, useIsOrgAdmin } from "@/stores/org-store";
import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  AlertTriangle,
  Building2,
  Crown,
  Loader2,
  MoreHorizontal,
  RefreshCw,
  Shield,
  Trash2,
  User,
  UserPlus,
  Users,
} from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/org/$orgSlug/team")({
  loader: ({ context, params }) => context.queryClient.ensureQueryData(orgMembersQueryOptions(params.orgSlug)),
  pendingMs: 200,
  pendingComponent: TeamPagePending,
  errorComponent: TeamPageError,
  notFoundComponent: OrgNotFound,
  component: TeamPage,
});

function TeamPage() {
  const { orgSlug } = Route.useParams();
  const currentOrg = useCurrentOrg();
  const isAdmin = useIsOrgAdmin();
  const currentUser = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<"admin" | "member">("member");

  // Members are guaranteed by the loader - use suspense query
  const { data: members } = useSuspenseQuery(orgMembersQueryOptions(orgSlug));

  // Fetch invitations (still using regular query since it depends on isAdmin)
  const { data: invitations = [] } = useQuery({
    queryKey: queryKeys.organizations.invitations(orgSlug),
    queryFn: () => OrganizationService.listInvitations(orgSlug),
    enabled: isAdmin,
  });

  // Invite member mutation
  const inviteMutation = useMutation({
    mutationFn: () => OrganizationService.inviteMember(orgSlug, { email: inviteEmail, role: inviteRole }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.invitations(orgSlug) });
      toast.success("Invitation sent successfully");
      setInviteDialogOpen(false);
      setInviteEmail("");
      setInviteRole("member");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Update role mutation
  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: "owner" | "admin" | "member" }) =>
      OrganizationService.updateMemberRole(orgSlug, userId, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.members(orgSlug) });
      toast.success("Role updated successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Remove member mutation
  const removeMemberMutation = useMutation({
    mutationFn: (userId: number) => OrganizationService.removeMember(orgSlug, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.members(orgSlug) });
      toast.success("Member removed successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Cancel invitation mutation
  const cancelInvitationMutation = useMutation({
    mutationFn: (invitationId: number) => OrganizationService.cancelInvitation(orgSlug, invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations.invitations(orgSlug) });
      toast.success("Invitation cancelled");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const getRoleIcon = (role: string) => {
    switch (role) {
      case "owner":
        return <Crown className="h-4 w-4 text-yellow-500" />;
      case "admin":
        return <Shield className="h-4 w-4 text-blue-500" />;
      default:
        return <User className="h-4 w-4 text-gray-500" />;
    }
  };

  const getRoleBadge = (role: string) => {
    switch (role) {
      case "owner":
        return <Badge variant="default">Owner</Badge>;
      case "admin":
        return <Badge variant="secondary">Admin</Badge>;
      default:
        return <Badge variant="outline">Member</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Team</h2>
          <p className="text-muted-foreground text-sm">
            Manage members and invitations for {currentOrg?.name || orgSlug}
          </p>
        </div>
        {isAdmin && (
          <Dialog
            open={inviteDialogOpen}
            onOpenChange={setInviteDialogOpen}
          >
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="mr-2 h-4 w-4" />
                Invite Member
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Invite Team Member</DialogTitle>
                <DialogDescription>
                  Send an invitation to join {currentOrg?.name || "the organization"}.
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="email">Email address</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="colleague@example.com"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="role">Role</Label>
                  <Select
                    value={inviteRole}
                    onValueChange={(v) => setInviteRole(v as "admin" | "member")}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="member">Member</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-muted-foreground text-xs">Admins can manage members and organization settings.</p>
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setInviteDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => inviteMutation.mutate()}
                  disabled={!inviteEmail || inviteMutation.isPending}
                >
                  {inviteMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Send Invitation
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* Members Table */}
      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
          <CardDescription>{members.length} team members</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Member</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Joined</TableHead>
                {isAdmin && <TableHead className="w-[50px]"></TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.map((member) => (
                <TableRow key={member.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="bg-muted flex h-10 w-10 items-center justify-center rounded-full">
                        {getRoleIcon(member.role)}
                      </div>
                      <div>
                        <p className="font-medium">{member.name || "Unnamed"}</p>
                        <p className="text-muted-foreground text-sm">{member.email}</p>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{getRoleBadge(member.role)}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {member.joined_at ? new Date(member.joined_at).toLocaleDateString() : "—"}
                  </TableCell>
                  {isAdmin && (
                    <TableCell>
                      {member.user_id !== currentUser?.id && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                            >
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            {currentOrg?.role === "owner" && member.role !== "owner" && (
                              <>
                                <DropdownMenuItem
                                  onClick={() =>
                                    updateRoleMutation.mutate({
                                      userId: member.user_id,
                                      role: member.role === "admin" ? "member" : "admin",
                                    })
                                  }
                                >
                                  {member.role === "admin" ? "Demote to Member" : "Promote to Admin"}
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                              </>
                            )}
                            {member.role !== "owner" && (
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => removeMemberMutation.mutate(member.user_id)}
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                Remove
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Pending Invitations */}
      {isAdmin && invitations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Pending Invitations</CardTitle>
            <CardDescription>{invitations.length} pending invitations</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Invited By</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {invitations.map((invitation) => (
                  <TableRow key={invitation.id}>
                    <TableCell className="font-medium">{invitation.email}</TableCell>
                    <TableCell>{getRoleBadge(invitation.role)}</TableCell>
                    <TableCell className="text-muted-foreground">{invitation.invited_by}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(invitation.expires_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => cancelInvitationMutation.mutate(invitation.id)}
                        disabled={cancelInvitationMutation.isPending}
                      >
                        <Trash2 className="text-destructive h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function TeamPagePending() {
  return (
    <div className="space-y-6">
      {/* Header skeleton */}
      <div className="flex items-center justify-between">
        <div>
          <Skeleton className="h-8 w-24" />
          <Skeleton className="mt-2 h-4 w-64" />
        </div>
        <Skeleton className="h-10 w-32" />
      </div>

      {/* Members table skeleton */}
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-24" />
          <Skeleton className="h-4 w-32" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div
                key={i}
                className="flex items-center justify-between py-2"
              >
                <div className="flex items-center gap-3">
                  <Skeleton className="h-10 w-10 rounded-full" />
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                </div>
                <Skeleton className="h-6 w-16" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function TeamPageError({ error, reset }: { error: Error; reset?: () => void }) {
  const { orgSlug } = Route.useParams();

  return (
    <div className="space-y-6">
      <Card className="border-destructive/30 bg-destructive/5">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="bg-destructive/10 rounded-full p-2">
              <AlertTriangle className="text-destructive h-5 w-5" />
            </div>
            <div>
              <CardTitle className="text-destructive">Failed to load team</CardTitle>
              <CardDescription className="text-destructive/80">
                {error.message || "An unexpected error occurred while loading the team."}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="flex gap-3">
          {reset && (
            <Button
              variant="outline"
              onClick={reset}
              className="gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              Try Again
            </Button>
          )}
          <Button
            variant="ghost"
            asChild
          >
            <Link
              to="/org/$orgSlug/settings"
              params={{ orgSlug }}
            >
              Go to Settings
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

function OrgNotFound() {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center">
      <div className="text-center">
        <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
          <Building2 className="text-muted-foreground h-8 w-8" />
        </div>
        <h2 className="mb-2 text-xl font-semibold">Organization not found</h2>
        <p className="text-muted-foreground mb-6 max-w-sm">
          The organization you're looking for doesn't exist or you don't have permission to view it.
        </p>
        <div className="flex justify-center gap-3">
          <Button
            variant="outline"
            asChild
          >
            <Link to="/dashboard">Go to Dashboard</Link>
          </Button>
          <Button asChild>
            <Link to="/settings">
              <Users className="mr-2 h-4 w-4" />
              View Your Organizations
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
