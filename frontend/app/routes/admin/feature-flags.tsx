import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Flag, Pencil, Plus, RefreshCw, Save, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import { AdminPageHeader, ConfirmDialog } from "../../components/admin";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Progress } from "../../components/ui/progress";
import { Slider } from "../../components/ui/slider";
import { Switch } from "../../components/ui/switch";
import { Tooltip, TooltipContent, TooltipTrigger } from "../../components/ui/tooltip";
import { requireAdmin } from "../../lib/guards";
import {
  AdminService,
  type CreateFeatureFlagRequest,
  type FeatureFlag,
  type FeatureFlagsResponse,
} from "../../services/admin";

export const Route = createFileRoute("/admin/feature-flags")({
  beforeLoad: () => requireAdmin(),
  component: FeatureFlagsPage,
});

function FeatureFlagsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery<FeatureFlagsResponse>({
    queryKey: ["admin", "feature-flags"],
    queryFn: () => AdminService.getFeatureFlags(),
  });

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Feature Flags"
        description="Control feature rollouts and A/B testing"
        breadcrumbs={[{ label: "Feature Flags" }]}
        actions={
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => refetch()}
              className="gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              <span className="hidden sm:inline">Refresh</span>
            </Button>
            <Button
              onClick={() => setShowCreate(true)}
              className="gap-2"
            >
              <Plus className="h-4 w-4" />
              <span className="hidden sm:inline">Create Flag</span>
            </Button>
          </div>
        }
      />

      {/* Create Form */}
      {showCreate && (
        <CreateFeatureFlagForm
          onClose={() => setShowCreate(false)}
          onSuccess={() => {
            setShowCreate(false);
            queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
          }}
        />
      )}

      {/* Loading State */}
      {isLoading && (
        <Card>
          <CardContent className="py-8">
            <div className="flex items-center justify-center">
              <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
              <span className="ml-2 text-gray-500">Loading feature flags...</span>
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
              {error instanceof Error ? error.message : "Failed to load feature flags"}
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {/* Feature Flags List */}
      {data && (
        <div className="space-y-4">
          {data.flags.length === 0 ? (
            <Card>
              <CardContent className="py-12">
                <div className="flex flex-col items-center justify-center text-center">
                  <Flag className="mb-4 h-12 w-12 text-gray-300" />
                  <p className="font-medium text-gray-900 dark:text-gray-100">No feature flags created yet</p>
                  <p className="mt-1 text-sm text-gray-500">Click "Create Flag" to add your first feature flag.</p>
                </div>
              </CardContent>
            </Card>
          ) : (
            data.flags.map((flag) => (
              <FeatureFlagCard
                key={flag.id}
                flag={flag}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function CreateFeatureFlagForm({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) {
  const [formData, setFormData] = useState<CreateFeatureFlagRequest>({
    key: "",
    name: "",
    description: "",
    enabled: false,
    rollout_percentage: 0,
    allowed_roles: [],
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateFeatureFlagRequest) => AdminService.createFeatureFlag(data),
    onSuccess: () => {
      toast.success("Feature flag created successfully");
      onSuccess();
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to create feature flag");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Create Feature Flag</CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-4"
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                placeholder="feature_key"
                value={formData.key}
                onChange={(e) =>
                  setFormData({ ...formData, key: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, "_") })
                }
                required
              />
              <p className="text-muted-foreground text-xs">Lowercase letters, numbers, and underscores only</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="Feature Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              placeholder="What does this feature flag control?"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-4">
              <Label>Rollout Percentage: {formData.rollout_percentage}%</Label>
              <Slider
                value={[formData.rollout_percentage || 0]}
                onValueChange={([value]) => setFormData({ ...formData, rollout_percentage: value })}
                max={100}
                step={1}
              />
            </div>
            <div className="flex items-center justify-between rounded-lg border p-4 dark:border-gray-700">
              <div>
                <Label
                  htmlFor="enabled"
                  className="cursor-pointer"
                >
                  Enabled
                </Label>
                <p className="text-muted-foreground text-xs">Turn this flag on immediately</p>
              </div>
              <Switch
                id="enabled"
                checked={formData.enabled}
                onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={createMutation.isPending}
              className="gap-2"
            >
              {createMutation.isPending ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              Create
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function FeatureFlagCard({ flag }: { flag: FeatureFlag }) {
  const [isEditing, setIsEditing] = useState(false);
  const [editData, setEditData] = useState({
    name: flag.name,
    description: flag.description || "",
    enabled: flag.enabled,
    rollout_percentage: flag.rollout_percentage,
  });
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: () => AdminService.updateFeatureFlag(flag.key, editData),
    onSuccess: () => {
      toast.success("Feature flag updated");
      setIsEditing(false);
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to update");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => AdminService.deleteFeatureFlag(flag.key),
    onSuccess: () => {
      toast.success("Feature flag deleted");
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to delete");
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) => AdminService.updateFeatureFlag(flag.key, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
      toast.success(flag.enabled ? "Feature flag disabled" : "Feature flag enabled");
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to toggle");
    },
  });

  return (
    <>
      <Card className="overflow-hidden">
        <CardContent className="p-0">
          <div className="flex flex-col gap-4 p-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0 flex-1 space-y-3">
              {/* Header with key and toggle */}
              <div className="flex flex-wrap items-center gap-3">
                <code className="rounded bg-gray-100 px-2 py-1 font-mono text-sm dark:bg-gray-800">{flag.key}</code>
                <Switch
                  checked={flag.enabled}
                  onCheckedChange={(checked) => toggleMutation.mutate(checked)}
                  disabled={toggleMutation.isPending}
                />
                <Badge variant={flag.enabled ? "default" : "secondary"}>{flag.enabled ? "Enabled" : "Disabled"}</Badge>
              </div>

              {isEditing ? (
                <div className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                      <Label>Name</Label>
                      <Input
                        value={editData.name}
                        onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                        placeholder="Name"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Description</Label>
                      <Input
                        value={editData.description}
                        onChange={(e) => setEditData({ ...editData, description: e.target.value })}
                        placeholder="Description"
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>Rollout Percentage: {editData.rollout_percentage}%</Label>
                    <Slider
                      value={[editData.rollout_percentage]}
                      onValueChange={([value]) => setEditData({ ...editData, rollout_percentage: value })}
                      max={100}
                      step={1}
                    />
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={() => updateMutation.mutate()}
                      disabled={updateMutation.isPending}
                    >
                      {updateMutation.isPending ? "Saving..." : "Save"}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        setIsEditing(false);
                        setEditData({
                          name: flag.name,
                          description: flag.description || "",
                          enabled: flag.enabled,
                          rollout_percentage: flag.rollout_percentage,
                        });
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              ) : (
                <>
                  <div>
                    <h3 className="font-medium">{flag.name}</h3>
                    {flag.description && <p className="text-muted-foreground mt-1 text-sm">{flag.description}</p>}
                  </div>
                  <div className="flex flex-wrap items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground text-sm">Rollout:</span>
                      <div className="flex items-center gap-2">
                        <Progress
                          value={flag.rollout_percentage}
                          className="h-2 w-20"
                        />
                        <span className="text-sm font-medium">{flag.rollout_percentage}%</span>
                      </div>
                    </div>
                    {flag.allowed_roles && flag.allowed_roles.length > 0 && (
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground text-sm">Roles:</span>
                        <div className="flex gap-1">
                          {flag.allowed_roles.map((role) => (
                            <Badge
                              key={role}
                              variant="outline"
                              className="text-xs"
                            >
                              {role}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </>
              )}
            </div>

            {!isEditing && (
              <div className="flex gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setIsEditing(true)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Edit flag</TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => setDeleteConfirm(true)}
                      disabled={deleteMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Delete flag</TooltipContent>
                </Tooltip>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteConfirm}
        onOpenChange={setDeleteConfirm}
        title="Delete Feature Flag"
        description={`Are you sure you want to delete the feature flag "${flag.key}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        loading={deleteMutation.isPending}
      />
    </>
  );
}
