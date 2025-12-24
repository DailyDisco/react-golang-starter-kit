import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Flag, Plus, RefreshCw, Save, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
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
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Feature Flags</h2>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => refetch()}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Flag
          </Button>
        </div>
      </div>

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
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-red-600">Error</CardTitle>
            <CardDescription className="text-red-500">
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
              <CardContent className="py-8">
                <div className="text-center">
                  <Flag className="mx-auto mb-4 h-12 w-12 text-gray-300" />
                  <p className="text-gray-500">No feature flags created yet.</p>
                  <p className="mt-1 text-sm text-gray-400">Click "Create Flag" to add your first feature flag.</p>
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
          <div className="grid grid-cols-2 gap-4">
            <div>
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
              <p className="mt-1 text-xs text-gray-500">Lowercase letters, numbers, and underscores only</p>
            </div>
            <div>
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
          <div>
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              placeholder="What does this feature flag control?"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="rollout">Rollout Percentage</Label>
              <Input
                id="rollout"
                type="number"
                min={0}
                max={100}
                value={formData.rollout_percentage}
                onChange={(e) => setFormData({ ...formData, rollout_percentage: parseInt(e.target.value) || 0 })}
              />
            </div>
            <div className="flex items-center gap-2 pt-6">
              <input
                type="checkbox"
                id="enabled"
                checked={formData.enabled}
                onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="enabled">Enabled</Label>
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
            >
              {createMutation.isPending ? (
                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Save className="mr-2 h-4 w-4" />
              )}
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
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : "Failed to toggle");
    },
  });

  return (
    <Card>
      <CardContent className="py-4">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3">
              <code className="rounded bg-gray-100 px-2 py-1 font-mono text-sm">{flag.key}</code>
              <button
                onClick={() => toggleMutation.mutate(!flag.enabled)}
                disabled={toggleMutation.isPending}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  flag.enabled ? "bg-blue-600" : "bg-gray-200"
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    flag.enabled ? "translate-x-6" : "translate-x-1"
                  }`}
                />
              </button>
              <span className={`text-sm ${flag.enabled ? "text-green-600" : "text-gray-500"}`}>
                {flag.enabled ? "Enabled" : "Disabled"}
              </span>
            </div>
            {isEditing ? (
              <div className="mt-4 space-y-3">
                <Input
                  value={editData.name}
                  onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                  placeholder="Name"
                />
                <Input
                  value={editData.description}
                  onChange={(e) => setEditData({ ...editData, description: e.target.value })}
                  placeholder="Description"
                />
                <div className="flex items-center gap-2">
                  <Label>Rollout:</Label>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    className="w-24"
                    value={editData.rollout_percentage}
                    onChange={(e) => setEditData({ ...editData, rollout_percentage: parseInt(e.target.value) || 0 })}
                  />
                  <span className="text-sm text-gray-500">%</span>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    onClick={() => updateMutation.mutate()}
                    disabled={updateMutation.isPending}
                  >
                    Save
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setIsEditing(false)}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <h3 className="mt-2 font-medium text-gray-900">{flag.name}</h3>
                {flag.description && <p className="mt-1 text-sm text-gray-500">{flag.description}</p>}
                <div className="mt-2 flex items-center gap-4 text-sm text-gray-500">
                  <span>Rollout: {flag.rollout_percentage}%</span>
                  {flag.allowed_roles && flag.allowed_roles.length > 0 && (
                    <span>Roles: {flag.allowed_roles.join(", ")}</span>
                  )}
                </div>
              </>
            )}
          </div>
          <div className="flex gap-2">
            {!isEditing && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setIsEditing(true)}
                >
                  Edit
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="text-red-600 hover:text-red-700"
                  onClick={() => {
                    if (confirm(`Delete feature flag "${flag.key}"?`)) {
                      deleteMutation.mutate();
                    }
                  }}
                  disabled={deleteMutation.isPending}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
