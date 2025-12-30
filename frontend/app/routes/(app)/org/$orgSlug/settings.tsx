import { useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { OrganizationService } from "@/services/organizations/organizationService";
import { useCurrentOrg, useIsOrgOwner, useOrgStore } from "@/stores/org-store";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { AlertTriangle, Building2, Loader2, Trash2 } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/org/$orgSlug/settings")({
  component: OrgSettingsPage,
});

function OrgSettingsPage() {
  const { orgSlug } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const currentOrg = useCurrentOrg();
  const isOwner = useIsOrgOwner();
  const { updateOrganization, removeOrganization } = useOrgStore();

  const [name, setName] = useState(currentOrg?.name || "");
  const [deleteConfirmation, setDeleteConfirmation] = useState("");

  // Update organization mutation
  const updateMutation = useMutation({
    mutationFn: () => OrganizationService.updateOrganization(orgSlug, { name }),
    onSuccess: (updatedOrg) => {
      updateOrganization(orgSlug, { name: updatedOrg.name });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      toast.success("Organization updated successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Delete organization mutation
  const deleteMutation = useMutation({
    mutationFn: () => OrganizationService.deleteOrganization(orgSlug),
    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      toast.success("Organization deleted successfully");
      navigate({ to: "/" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Leave organization mutation
  const leaveMutation = useMutation({
    mutationFn: () => OrganizationService.leaveOrganization(orgSlug),
    onSuccess: () => {
      removeOrganization(orgSlug);
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      toast.success("You have left the organization");
      navigate({ to: "/" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (!currentOrg) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold">Organization Settings</h2>
        <p className="text-muted-foreground text-sm">Manage settings for {currentOrg.name}</p>
      </div>

      {/* General Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            General
          </CardTitle>
          <CardDescription>Basic organization information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Organization Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Organization"
              disabled={!isOwner && currentOrg.role !== "admin"}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="slug">Organization Slug</Label>
            <Input
              id="slug"
              value={currentOrg.slug}
              disabled
            />
            <p className="text-muted-foreground text-xs">The slug cannot be changed after creation.</p>
          </div>
          <div className="grid gap-2">
            <Label>Plan</Label>
            <Input
              value={currentOrg.plan}
              disabled
              className="capitalize"
            />
          </div>
          {(isOwner || currentOrg.role === "admin") && (
            <Button
              onClick={() => updateMutation.mutate()}
              disabled={updateMutation.isPending || name === currentOrg.name}
            >
              {updateMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Save Changes
            </Button>
          )}
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="text-destructive flex items-center gap-2">
            <AlertTriangle className="h-5 w-5" />
            Danger Zone
          </CardTitle>
          <CardDescription>Irreversible actions that affect the organization</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Leave Organization (non-owners) */}
          {!isOwner && (
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <h4 className="font-medium">Leave Organization</h4>
                <p className="text-muted-foreground text-sm">Remove yourself from this organization</p>
              </div>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="outline">Leave</Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Leave Organization?</AlertDialogTitle>
                    <AlertDialogDescription>
                      You will lose access to {currentOrg.name}. You will need to be re-invited to rejoin.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={() => leaveMutation.mutate()}
                      disabled={leaveMutation.isPending}
                    >
                      {leaveMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                      Leave Organization
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          )}

          {/* Delete Organization (owners only) */}
          {isOwner && (
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <h4 className="font-medium">Delete Organization</h4>
                <p className="text-muted-foreground text-sm">Permanently delete this organization and all its data</p>
              </div>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive">
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete Organization?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This action cannot be undone. All organization data, members, and invitations will be permanently
                      deleted.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <div className="py-4">
                    <Label htmlFor="confirm">
                      Type <span className="font-mono font-bold">{currentOrg.slug}</span> to confirm
                    </Label>
                    <Input
                      id="confirm"
                      value={deleteConfirmation}
                      onChange={(e) => setDeleteConfirmation(e.target.value)}
                      placeholder={currentOrg.slug}
                      className="mt-2"
                    />
                  </div>
                  <AlertDialogFooter>
                    <AlertDialogCancel onClick={() => setDeleteConfirmation("")}>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={() => deleteMutation.mutate()}
                      disabled={deleteConfirmation !== currentOrg.slug || deleteMutation.isPending}
                      className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                    >
                      {deleteMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                      Delete Organization
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
