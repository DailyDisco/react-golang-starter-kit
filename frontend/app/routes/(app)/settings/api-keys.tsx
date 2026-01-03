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
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { requireAuth } from "@/lib/guards";
import { queryKeys } from "@/lib/query-keys";
import { apiKeysQueryOptions } from "@/lib/route-query-options";
import { SettingsService, type CreateAPIKeyRequest, type UserAPIKey } from "@/services/settings/settingsService";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Brain, Check, Eye, EyeOff, Key, Loader2, Plus, Sparkles, TestTube, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/api-keys")({
  // Ensure user is authenticated before loading data
  beforeLoad: async (ctx) => {
    await requireAuth(ctx);
  },
  // Prefetch API keys data before component renders for faster navigation
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(apiKeysQueryOptions());
  },
  component: APIKeysSettingsPage,
});

const API_KEY_PROVIDERS = [
  {
    value: "gemini" as const,
    label: "Google Gemini",
    icon: Sparkles,
    description: "Google's Gemini AI models for chat, vision, and embeddings",
  },
  {
    value: "openai" as const,
    label: "OpenAI",
    icon: Brain,
    description: "GPT models and DALL-E for chat and image generation",
  },
  {
    value: "anthropic" as const,
    label: "Anthropic",
    icon: Brain,
    description: "Claude models for advanced reasoning and analysis",
  },
];

function APIKeysSettingsPage() {
  const { t } = useTranslation("settings");
  const queryClient = useQueryClient();
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState<UserAPIKey | null>(null);

  const { data: apiKeys, isLoading } = useQuery({
    queryKey: queryKeys.settings.apiKeys(),
    queryFn: () => SettingsService.getAPIKeys(),
  });

  const deleteKeyMutation = useMutation({
    mutationFn: (id: number) => SettingsService.deleteAPIKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
      toast.success(t("apiKeys.toast.deleted"));
      setShowDeleteDialog(null);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const testKeyMutation = useMutation({
    mutationFn: (id: number) => SettingsService.testAPIKey(id),
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t("apiKeys.toast.testSuccess"));
      } else {
        toast.error(data.message || t("apiKeys.toast.testFailed"));
      }
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const toggleKeyMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      SettingsService.updateAPIKey(id, { is_active: isActive }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
      toast.success(t("apiKeys.toast.updated"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const getProviderInfo = (provider: string) => {
    return API_KEY_PROVIDERS.find((p) => p.value === provider) || API_KEY_PROVIDERS[0];
  };

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">{t("apiKeys.title")}</h2>
            <p className="text-muted-foreground text-sm">{t("apiKeys.subtitle")}</p>
          </div>
          <Button onClick={() => setShowAddDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            {t("apiKeys.addKey")}
          </Button>
        </div>

        {/* Info Card */}
        <Card className="border-blue-500/30 bg-blue-500/10">
          <CardContent className="flex items-start gap-4 p-4">
            <div className="rounded-lg bg-blue-500/20 p-2">
              <Key className="h-5 w-5 text-blue-500" />
            </div>
            <div>
              <p className="font-medium text-blue-700 dark:text-blue-300">{t("apiKeys.info.title")}</p>
              <p className="text-muted-foreground text-sm">{t("apiKeys.info.description")}</p>
            </div>
          </CardContent>
        </Card>

        {/* API Keys List */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              {t("apiKeys.yourKeys")}
            </CardTitle>
            <CardDescription>{t("apiKeys.yourKeysDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-3">
                {[1, 2].map((i) => (
                  <div
                    key={i}
                    className="bg-muted h-20 animate-pulse rounded-lg"
                  />
                ))}
              </div>
            ) : !apiKeys || apiKeys.length === 0 ? (
              <div className="py-8 text-center">
                <Key className="text-muted-foreground mx-auto h-12 w-12" />
                <p className="text-muted-foreground mt-4">{t("apiKeys.noKeys")}</p>
                <Button
                  className="mt-4"
                  variant="outline"
                  onClick={() => setShowAddDialog(true)}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  {t("apiKeys.addFirstKey")}
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                {apiKeys.map((key) => {
                  const providerInfo = getProviderInfo(key.provider);
                  const Icon = providerInfo.icon;
                  return (
                    <div
                      key={key.id}
                      className={`flex items-center justify-between rounded-lg border p-4 ${
                        key.is_active ? "" : "opacity-60"
                      }`}
                    >
                      <div className="flex items-center gap-4">
                        <div
                          className={`rounded-lg p-2 ${
                            key.provider === "gemini"
                              ? "bg-purple-500/20"
                              : key.provider === "openai"
                                ? "bg-green-500/20"
                                : "bg-orange-500/20"
                          }`}
                        >
                          <Icon
                            className={`h-5 w-5 ${
                              key.provider === "gemini"
                                ? "text-purple-500"
                                : key.provider === "openai"
                                  ? "text-green-500"
                                  : "text-orange-500"
                            }`}
                          />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{key.name}</span>
                            <Badge variant={key.is_active ? "default" : "secondary"}>{providerInfo.label}</Badge>
                          </div>
                          <div className="text-muted-foreground text-sm">
                            <code className="bg-muted rounded px-1">{key.key_preview}</code>
                            {key.last_used_at && (
                              <span className="ml-2">
                                {t("apiKeys.lastUsed")}: {new Date(key.last_used_at).toLocaleDateString()}
                              </span>
                            )}
                            {key.usage_count > 0 && (
                              <span className="ml-2">
                                ({key.usage_count} {t("apiKeys.uses")})
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={key.is_active}
                          onCheckedChange={(checked) => toggleKeyMutation.mutate({ id: key.id, isActive: checked })}
                        />
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => testKeyMutation.mutate(key.id)}
                          disabled={testKeyMutation.isPending}
                        >
                          {testKeyMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <TestTube className="h-4 w-4" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setShowDeleteDialog(key)}
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Security Note */}
        <Card className="border-amber-500/30 bg-amber-500/10">
          <CardContent className="flex items-start gap-4 p-4">
            <div className="rounded-lg bg-amber-500/20 p-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
            </div>
            <div>
              <p className="font-medium text-amber-700 dark:text-amber-300">{t("apiKeys.security.title")}</p>
              <p className="text-muted-foreground text-sm">{t("apiKeys.security.description")}</p>
            </div>
          </CardContent>
        </Card>

        {/* Add API Key Dialog */}
        <AddAPIKeyDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
        />

        {/* Delete Confirmation Dialog */}
        <Dialog
          open={!!showDeleteDialog}
          onOpenChange={() => setShowDeleteDialog(null)}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("apiKeys.deleteDialog.title")}</DialogTitle>
              <DialogDescription>
                {t("apiKeys.deleteDialog.description", { name: showDeleteDialog?.name })}
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setShowDeleteDialog(null)}
              >
                {t("apiKeys.deleteDialog.cancel")}
              </Button>
              <Button
                variant="destructive"
                onClick={() => showDeleteDialog && deleteKeyMutation.mutate(showDeleteDialog.id)}
                disabled={deleteKeyMutation.isPending}
              >
                {deleteKeyMutation.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Trash2 className="mr-2 h-4 w-4" />
                )}
                {t("apiKeys.deleteDialog.confirm")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </SettingsLayout>
  );
}

function AddAPIKeyDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t } = useTranslation("settings");
  const queryClient = useQueryClient();
  const [showKey, setShowKey] = useState(false);
  const [formData, setFormData] = useState<CreateAPIKeyRequest>({
    provider: "gemini",
    name: "",
    api_key: "",
  });

  const createKeyMutation = useMutation({
    mutationFn: (req: CreateAPIKeyRequest) => SettingsService.createAPIKey(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.apiKeys() });
      toast.success(t("apiKeys.toast.created"));
      onOpenChange(false);
      setFormData({ provider: "gemini", name: "", api_key: "" });
      setShowKey(false);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim() || !formData.api_key.trim()) {
      toast.error(t("apiKeys.validation.required"));
      return;
    }
    createKeyMutation.mutate(formData);
  };

  const selectedProvider = API_KEY_PROVIDERS.find((p) => p.value === formData.provider);

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{t("apiKeys.addDialog.title")}</DialogTitle>
          <DialogDescription>{t("apiKeys.addDialog.description")}</DialogDescription>
        </DialogHeader>
        <form
          onSubmit={handleSubmit}
          className="space-y-4"
        >
          <div className="space-y-2">
            <Label htmlFor="provider">{t("apiKeys.addDialog.provider")}</Label>
            <Select
              value={formData.provider}
              onValueChange={(value: "gemini" | "openai" | "anthropic") =>
                setFormData({ ...formData, provider: value })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {API_KEY_PROVIDERS.map((provider) => {
                  const Icon = provider.icon;
                  return (
                    <SelectItem
                      key={provider.value}
                      value={provider.value}
                    >
                      <div className="flex items-center gap-2">
                        <Icon className="h-4 w-4" />
                        {provider.label}
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
            {selectedProvider && <p className="text-muted-foreground text-xs">{selectedProvider.description}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">{t("apiKeys.addDialog.name")}</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder={t("apiKeys.addDialog.namePlaceholder")}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="api_key">{t("apiKeys.addDialog.apiKey")}</Label>
            <div className="relative">
              <Input
                id="api_key"
                type={showKey ? "text" : "password"}
                value={formData.api_key}
                onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                placeholder={t("apiKeys.addDialog.apiKeyPlaceholder")}
                className="pr-10"
              />
              <button
                type="button"
                className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
                onClick={() => setShowKey(!showKey)}
              >
                {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t("apiKeys.addDialog.cancel")}
            </Button>
            <Button
              type="submit"
              disabled={createKeyMutation.isPending}
            >
              {createKeyMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Check className="mr-2 h-4 w-4" />
              )}
              {t("apiKeys.addDialog.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
