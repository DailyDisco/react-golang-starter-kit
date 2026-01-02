import { useState } from "react";

import { ConfirmDialog } from "@/components/admin";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { AdminLayout } from "@/layouts/AdminLayout";
import { requireAdmin } from "@/lib/guards";
import {
  AdminService,
  type CreateFeatureFlagRequest,
  type FeatureFlag,
  type FeatureFlagsResponse,
} from "@/services/admin";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Flag, Pencil, Plus, RefreshCw, Save, Trash2, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/feature-flags")({
  beforeLoad: async (ctx) => requireAdmin(ctx),
  component: FeatureFlagsPage,
});

function FeatureFlagsPage() {
  const { t } = useTranslation("admin");
  const [showCreate, setShowCreate] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery<FeatureFlagsResponse>({
    queryKey: ["admin", "feature-flags"],
    queryFn: () => AdminService.getFeatureFlags(),
  });

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">{t("featureFlags.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("featureFlags.subtitle")}</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => refetch()}
              className="gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              <span className="hidden sm:inline">{t("featureFlags.refresh")}</span>
            </Button>
            <Button
              onClick={() => setShowCreate(true)}
              className="gap-2"
            >
              <Plus className="h-4 w-4" />
              <span className="hidden sm:inline">{t("featureFlags.createFlag")}</span>
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
                <RefreshCw className="text-muted-foreground h-6 w-6 animate-spin" />
                <span className="text-muted-foreground ml-2">{t("featureFlags.loading")}</span>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Error State */}
        {error && (
          <Card className="border-destructive/30 bg-destructive/5">
            <CardHeader>
              <CardTitle className="text-destructive">{t("featureFlags.error.title")}</CardTitle>
              <CardDescription className="text-destructive/80">
                {error instanceof Error ? error.message : t("featureFlags.error.default")}
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
                    <Flag className="text-muted-foreground/50 mb-4 h-12 w-12" />
                    <p className="font-medium">{t("featureFlags.empty.title")}</p>
                    <p className="text-muted-foreground mt-1 text-sm">{t("featureFlags.empty.hint")}</p>
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
    </AdminLayout>
  );
}

function CreateFeatureFlagForm({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) {
  const { t } = useTranslation("admin");
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
      toast.success(t("featureFlags.toast.created"));
      onSuccess();
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("featureFlags.toast.createError"));
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
          <CardTitle>{t("featureFlags.create.title")}</CardTitle>
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
              <Label htmlFor="key">{t("featureFlags.create.key")}</Label>
              <Input
                id="key"
                placeholder={t("featureFlags.create.keyPlaceholder")}
                value={formData.key}
                onChange={(e) =>
                  setFormData({ ...formData, key: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, "_") })
                }
                required
              />
              <p className="text-muted-foreground text-xs">{t("featureFlags.create.keyHint")}</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="name">{t("featureFlags.create.name")}</Label>
              <Input
                id="name"
                placeholder={t("featureFlags.create.namePlaceholder")}
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">{t("featureFlags.create.description")}</Label>
            <Input
              id="description"
              placeholder={t("featureFlags.create.descriptionPlaceholder")}
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-4">
              <Label>{t("featureFlags.create.rolloutPercentage", { percentage: formData.rollout_percentage })}</Label>
              <Slider
                value={[formData.rollout_percentage || 0]}
                onValueChange={([value]) => setFormData({ ...formData, rollout_percentage: value })}
                max={100}
                step={1}
              />
            </div>
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <Label
                  htmlFor="enabled"
                  className="cursor-pointer"
                >
                  {t("featureFlags.create.enabled")}
                </Label>
                <p className="text-muted-foreground text-xs">{t("featureFlags.create.enabledHint")}</p>
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
              {t("featureFlags.create.cancel")}
            </Button>
            <Button
              type="submit"
              disabled={createMutation.isPending}
              className="gap-2"
            >
              {createMutation.isPending ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              {t("featureFlags.create.create")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function FeatureFlagCard({ flag }: { flag: FeatureFlag }) {
  const { t } = useTranslation("admin");
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
      toast.success(t("featureFlags.toast.updated"));
      setIsEditing(false);
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("featureFlags.toast.updateError"));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => AdminService.deleteFeatureFlag(flag.key),
    onSuccess: () => {
      toast.success(t("featureFlags.toast.deleted"));
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("featureFlags.toast.deleteError"));
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) => AdminService.updateFeatureFlag(flag.key, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "feature-flags"] });
      toast.success(flag.enabled ? t("featureFlags.toast.disabled") : t("featureFlags.toast.enabled"));
    },
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : t("featureFlags.toast.toggleError"));
    },
  });

  return (
    <>
      <Card className="overflow-hidden">
        <CardContent className="p-0">
          <div className="flex flex-col gap-4 p-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0 flex-1 space-y-3">
              <div className="flex flex-wrap items-center gap-3">
                <code className="bg-muted rounded px-2 py-1 font-mono text-sm">{flag.key}</code>
                <Switch
                  checked={flag.enabled}
                  onCheckedChange={(checked) => toggleMutation.mutate(checked)}
                  disabled={toggleMutation.isPending}
                />
                <Badge variant={flag.enabled ? "success" : "secondary"}>
                  {flag.enabled ? t("featureFlags.card.enabled") : t("featureFlags.card.disabled")}
                </Badge>
              </div>

              {isEditing ? (
                <div className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                      <Label>{t("featureFlags.create.name")}</Label>
                      <Input
                        value={editData.name}
                        onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                        placeholder={t("featureFlags.create.namePlaceholder")}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>{t("featureFlags.create.description")}</Label>
                      <Input
                        value={editData.description}
                        onChange={(e) => setEditData({ ...editData, description: e.target.value })}
                        placeholder={t("featureFlags.create.descriptionPlaceholder")}
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>{t("featureFlags.create.rolloutPercentage", { value: editData.rollout_percentage })}</Label>
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
                      {updateMutation.isPending ? t("featureFlags.card.saving") : t("featureFlags.card.save")}
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
                      {t("featureFlags.card.cancel")}
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
                      <span className="text-muted-foreground text-sm">{t("featureFlags.card.rollout")}:</span>
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
                        <span className="text-muted-foreground text-sm">{t("featureFlags.card.roles")}:</span>
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
                  <TooltipContent>{t("featureFlags.tooltips.edit")}</TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      onClick={() => setDeleteConfirm(true)}
                      disabled={deleteMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>{t("featureFlags.tooltips.delete")}</TooltipContent>
                </Tooltip>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteConfirm}
        onOpenChange={setDeleteConfirm}
        title={t("featureFlags.dialog.deleteTitle")}
        description={t("featureFlags.dialog.deleteDescription", { key: flag.key })}
        confirmLabel={t("featureFlags.dialog.delete")}
        variant="destructive"
        onConfirm={() => deleteMutation.mutate()}
        loading={deleteMutation.isPending}
      />
    </>
  );
}
