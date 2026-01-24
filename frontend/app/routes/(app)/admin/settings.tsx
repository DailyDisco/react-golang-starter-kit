import { useState } from "react";

import { AdminPageHeader } from "@/components/admin";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { AdminLayout } from "@/layouts/AdminLayout";
import {
  AdminSettingsService,
  type CreateIPBlockRequest,
  type EmailSettings,
  type IPBlock,
  type SecuritySettings,
  type SiteSettings,
} from "@/services/admin";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Globe, Lock, Mail, Shield, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/settings")({
  component: AdminSettingsPage,
});

function AdminSettingsPage() {
  const { t } = useTranslation("admin");

  return (
    <AdminLayout>
      <div className="space-y-6">
        <AdminPageHeader
          title={t("settings.title")}
          description={t("settings.subtitle")}
          breadcrumbs={[{ label: t("settings.title") }]}
        />

        <Tabs
          defaultValue="email"
          className="space-y-6"
        >
          <TabsList className="grid w-full grid-cols-2 lg:w-auto lg:grid-cols-4">
            <TabsTrigger
              value="email"
              className="gap-2"
            >
              <Mail className="h-4 w-4" />
              <span className="hidden sm:inline">{t("settings.tabs.email")}</span>
              <span className="sm:hidden">{t("settings.tabs.emailShort")}</span>
            </TabsTrigger>
            <TabsTrigger
              value="security"
              className="gap-2"
            >
              <Lock className="h-4 w-4" />
              <span>{t("settings.tabs.security")}</span>
            </TabsTrigger>
            <TabsTrigger
              value="site"
              className="gap-2"
            >
              <Globe className="h-4 w-4" />
              <span>{t("settings.tabs.site")}</span>
            </TabsTrigger>
            <TabsTrigger
              value="ip-blocklist"
              className="gap-2"
            >
              <Shield className="h-4 w-4" />
              <span className="hidden sm:inline">{t("settings.tabs.ipBlocklist")}</span>
              <span className="sm:hidden">{t("settings.tabs.ipBlocklistShort")}</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="email">
            <EmailSettingsTab />
          </TabsContent>
          <TabsContent value="security">
            <SecuritySettingsTab />
          </TabsContent>
          <TabsContent value="site">
            <SiteSettingsTab />
          </TabsContent>
          <TabsContent value="ip-blocklist">
            <IPBlocklistTab />
          </TabsContent>
        </Tabs>
      </div>
    </AdminLayout>
  );
}

function EmailSettingsTab() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();

  const { data: settings, isLoading } = useQuery<EmailSettings>({
    queryKey: ["admin", "settings", "email"],
    queryFn: () => AdminSettingsService.getEmailSettings(),
  });

  const [formData, setFormData] = useState<Partial<EmailSettings>>({});

  const updateMutation = useMutation({
    mutationFn: (data: Partial<EmailSettings>) => AdminSettingsService.updateEmailSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "settings", "email"] });
      toast.success(t("settings.email.toast.updated"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const testMutation = useMutation({
    mutationFn: () => AdminSettingsService.testEmailSettings(),
    onSuccess: (data) => {
      if (data.success) {
        toast.success(data.message);
      } else {
        toast.error(data.message);
      }
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (settings && Object.keys(formData).length === 0) {
    setFormData(settings);
  }

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("settings.email.title")}</CardTitle>
        <CardDescription>{t("settings.email.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-6"
        >
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label
                htmlFor="smtp_enabled"
                className="text-base"
              >
                {t("settings.email.enableSending")}
              </Label>
              <p className="text-muted-foreground text-sm">{t("settings.email.enableSendingHint")}</p>
            </div>
            <Switch
              id="smtp_enabled"
              checked={formData.smtp_enabled ?? false}
              onCheckedChange={(checked) => setFormData({ ...formData, smtp_enabled: checked })}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="smtp_host">{t("settings.email.smtpHost")}</Label>
              <Input
                id="smtp_host"
                value={formData.smtp_host ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_host: e.target.value })}
                placeholder="smtp.example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_port">{t("settings.email.smtpPort")}</Label>
              <Input
                id="smtp_port"
                type="number"
                value={formData.smtp_port ?? 587}
                onChange={(e) => setFormData({ ...formData, smtp_port: parseInt(e.target.value) })}
                placeholder="587"
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="smtp_username">{t("settings.email.smtpUsername")}</Label>
              <Input
                id="smtp_username"
                value={formData.smtp_username ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_username: e.target.value })}
                placeholder="your-username"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_password">{t("settings.email.smtpPassword")}</Label>
              <Input
                id="smtp_password"
                type="password"
                value={formData.smtp_password ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_password: e.target.value })}
                placeholder="••••••••"
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="smtp_from_email">{t("settings.email.fromEmail")}</Label>
              <Input
                id="smtp_from_email"
                type="email"
                value={formData.smtp_from_email ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_from_email: e.target.value })}
                placeholder="noreply@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_from_name">{t("settings.email.fromName")}</Label>
              <Input
                id="smtp_from_name"
                value={formData.smtp_from_name ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_from_name: e.target.value })}
                placeholder="My Application"
              />
            </div>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row sm:justify-between">
            <Button
              type="button"
              variant="outline"
              onClick={() => testMutation.mutate()}
              disabled={testMutation.isPending}
            >
              {testMutation.isPending ? t("settings.email.sendingTest") : t("settings.email.sendTest")}
            </Button>
            <Button
              type="submit"
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? t("settings.email.saving") : t("settings.email.saveChanges")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function SecuritySettingsTab() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();

  const { data: settings, isLoading } = useQuery<SecuritySettings>({
    queryKey: ["admin", "settings", "security"],
    queryFn: () => AdminSettingsService.getSecuritySettings(),
  });

  const [formData, setFormData] = useState<Partial<SecuritySettings>>({});

  const updateMutation = useMutation({
    mutationFn: (data: Partial<SecuritySettings>) => AdminSettingsService.updateSecuritySettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "settings", "security"] });
      toast.success(t("settings.security.toast.updated"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (settings && Object.keys(formData).length === 0) {
    setFormData(settings);
  }

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("settings.security.title")}</CardTitle>
        <CardDescription>{t("settings.security.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-8"
        >
          {/* Password Requirements */}
          <div className="space-y-4">
            <h4 className="font-medium">{t("settings.security.password.title")}</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="password_min_length">{t("settings.security.password.minLength")}</Label>
                <Input
                  id="password_min_length"
                  type="number"
                  min={6}
                  max={32}
                  value={formData.password_min_length ?? 8}
                  onChange={(e) => setFormData({ ...formData, password_min_length: parseInt(e.target.value) })}
                />
              </div>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center justify-between rounded-lg border p-3">
                <Label
                  htmlFor="require_uppercase"
                  className="cursor-pointer"
                >
                  {t("settings.security.password.requireUppercase")}
                </Label>
                <Switch
                  id="require_uppercase"
                  checked={formData.password_require_uppercase ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_uppercase: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3">
                <Label
                  htmlFor="require_lowercase"
                  className="cursor-pointer"
                >
                  {t("settings.security.password.requireLowercase")}
                </Label>
                <Switch
                  id="require_lowercase"
                  checked={formData.password_require_lowercase ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_lowercase: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3">
                <Label
                  htmlFor="require_number"
                  className="cursor-pointer"
                >
                  {t("settings.security.password.requireNumber")}
                </Label>
                <Switch
                  id="require_number"
                  checked={formData.password_require_number ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_number: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3">
                <Label
                  htmlFor="require_special"
                  className="cursor-pointer"
                >
                  {t("settings.security.password.requireSpecial")}
                </Label>
                <Switch
                  id="require_special"
                  checked={formData.password_require_special ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_special: checked })}
                />
              </div>
            </div>
          </div>

          {/* Session Settings */}
          <div className="space-y-4">
            <h4 className="font-medium">{t("settings.security.session.title")}</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="session_timeout_minutes">{t("settings.security.session.timeoutMinutes")}</Label>
                <Input
                  id="session_timeout_minutes"
                  type="number"
                  min={5}
                  value={formData.session_timeout_minutes ?? 30}
                  onChange={(e) => setFormData({ ...formData, session_timeout_minutes: parseInt(e.target.value) })}
                />
              </div>
            </div>
          </div>

          {/* Login Protection */}
          <div className="space-y-4">
            <h4 className="font-medium">{t("settings.security.login.title")}</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="max_login_attempts">{t("settings.security.login.maxAttempts")}</Label>
                <Input
                  id="max_login_attempts"
                  type="number"
                  min={1}
                  value={formData.max_login_attempts ?? 5}
                  onChange={(e) => setFormData({ ...formData, max_login_attempts: parseInt(e.target.value) })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lockout_duration_minutes">{t("settings.security.login.lockoutDuration")}</Label>
                <Input
                  id="lockout_duration_minutes"
                  type="number"
                  min={1}
                  value={formData.lockout_duration_minutes ?? 15}
                  onChange={(e) => setFormData({ ...formData, lockout_duration_minutes: parseInt(e.target.value) })}
                />
              </div>
            </div>
          </div>

          {/* Other Settings */}
          <div className="space-y-4">
            <h4 className="font-medium">{t("settings.security.additional.title")}</h4>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center justify-between rounded-lg border p-3">
                <div>
                  <Label
                    htmlFor="require_2fa"
                    className="cursor-pointer"
                  >
                    {t("settings.security.additional.require2fa")}
                  </Label>
                  <p className="text-muted-foreground text-xs">{t("settings.security.additional.require2faHint")}</p>
                </div>
                <Switch
                  id="require_2fa"
                  checked={formData.require_2fa_for_admin ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, require_2fa_for_admin: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3">
                <div>
                  <Label
                    htmlFor="allow_registration"
                    className="cursor-pointer"
                  >
                    {t("settings.security.additional.allowRegistration")}
                  </Label>
                  <p className="text-muted-foreground text-xs">
                    {t("settings.security.additional.allowRegistrationHint")}
                  </p>
                </div>
                <Switch
                  id="allow_registration"
                  checked={formData.allow_registration ?? true}
                  onCheckedChange={(checked) => setFormData({ ...formData, allow_registration: checked })}
                />
              </div>
            </div>
          </div>

          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? t("settings.security.saving") : t("settings.security.saveChanges")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function SiteSettingsTab() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();

  const { data: settings, isLoading } = useQuery<SiteSettings>({
    queryKey: ["admin", "settings", "site"],
    queryFn: () => AdminSettingsService.getSiteSettings(),
  });

  const [formData, setFormData] = useState<Partial<SiteSettings>>({});

  const updateMutation = useMutation({
    mutationFn: (data: Partial<SiteSettings>) => AdminSettingsService.updateSiteSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "settings", "site"] });
      toast.success(t("settings.site.toast.updated"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (settings && Object.keys(formData).length === 0) {
    setFormData(settings);
  }

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("settings.site.title")}</CardTitle>
        <CardDescription>{t("settings.site.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-6"
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="site_name">{t("settings.site.siteName")}</Label>
              <Input
                id="site_name"
                value={formData.site_name ?? ""}
                onChange={(e) => setFormData({ ...formData, site_name: e.target.value })}
                placeholder="My Application"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="site_logo_url">{t("settings.site.logoUrl")}</Label>
              <Input
                id="site_logo_url"
                value={formData.site_logo_url ?? ""}
                onChange={(e) => setFormData({ ...formData, site_logo_url: e.target.value })}
                placeholder="https://example.com/logo.png"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="site_description">{t("settings.site.siteDescription")}</Label>
            <Textarea
              id="site_description"
              value={formData.site_description ?? ""}
              onChange={(e) => setFormData({ ...formData, site_description: e.target.value })}
              placeholder={t("settings.site.siteDescriptionPlaceholder")}
              rows={3}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="contact_email">{t("settings.site.contactEmail")}</Label>
              <Input
                id="contact_email"
                type="email"
                value={formData.contact_email ?? ""}
                onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                placeholder="contact@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="support_url">{t("settings.site.supportUrl")}</Label>
              <Input
                id="support_url"
                value={formData.support_url ?? ""}
                onChange={(e) => setFormData({ ...formData, support_url: e.target.value })}
                placeholder="https://support.example.com"
              />
            </div>
          </div>

          {/* Maintenance Mode */}
          <div className="border-warning/30 bg-warning/5 rounded-lg border p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <AlertTriangle className="text-warning h-5 w-5" />
                <div>
                  <Label
                    htmlFor="maintenance_mode"
                    className="cursor-pointer font-medium"
                  >
                    {t("settings.site.maintenance.title")}
                  </Label>
                  <p className="text-muted-foreground text-sm">{t("settings.site.maintenance.subtitle")}</p>
                </div>
              </div>
              <Switch
                id="maintenance_mode"
                checked={formData.maintenance_mode ?? false}
                onCheckedChange={(checked) => setFormData({ ...formData, maintenance_mode: checked })}
              />
            </div>
            {formData.maintenance_mode && (
              <div className="mt-4 space-y-2">
                <Label htmlFor="maintenance_message">{t("settings.site.maintenance.message")}</Label>
                <Textarea
                  id="maintenance_message"
                  value={formData.maintenance_message ?? ""}
                  onChange={(e) => setFormData({ ...formData, maintenance_message: e.target.value })}
                  placeholder={t("settings.site.maintenance.messagePlaceholder")}
                  rows={2}
                />
              </div>
            )}
          </div>

          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending ? t("settings.site.saving") : t("settings.site.saveChanges")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function IPBlocklistTab() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    id: number;
    ip: string;
  }>({ open: false, id: 0, ip: "" });

  const { data: blocks, isLoading } = useQuery<IPBlock[]>({
    queryKey: ["admin", "ip-blocklist"],
    queryFn: () => AdminSettingsService.getIPBlocklist(),
  });

  const [newBlock, setNewBlock] = useState<CreateIPBlockRequest>({
    ip_address: "",
    reason: "",
    block_type: "manual",
  });

  const blockMutation = useMutation({
    mutationFn: (data: CreateIPBlockRequest) => AdminSettingsService.blockIP(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "ip-blocklist"] });
      setShowAddForm(false);
      setNewBlock({ ip_address: "", reason: "", block_type: "manual" });
      toast.success(t("settings.ipBlocklist.toast.blocked"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const unblockMutation = useMutation({
    mutationFn: (id: number) => AdminSettingsService.unblockIP(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "ip-blocklist"] });
      setConfirmDialog({ open: false, id: 0, ip: "" });
      toast.success(t("settings.ipBlocklist.toast.unblocked"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t("settings.ipBlocklist.title")}</CardTitle>
              <CardDescription>{t("settings.ipBlocklist.subtitle")}</CardDescription>
            </div>
            <Button onClick={() => setShowAddForm(true)}>{t("settings.ipBlocklist.blockIp")}</Button>
          </div>
        </CardHeader>
        <CardContent>
          {showAddForm && (
            <div className="mb-6 rounded-lg border p-4">
              <h4 className="mb-4 font-medium">{t("settings.ipBlocklist.addTitle")}</h4>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="new_ip">{t("settings.ipBlocklist.ipAddress")}</Label>
                  <Input
                    id="new_ip"
                    value={newBlock.ip_address}
                    onChange={(e) => setNewBlock({ ...newBlock, ip_address: e.target.value })}
                    placeholder="192.168.1.1"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="new_reason">{t("settings.ipBlocklist.reason")}</Label>
                  <Input
                    id="new_reason"
                    value={newBlock.reason}
                    onChange={(e) => setNewBlock({ ...newBlock, reason: e.target.value })}
                    placeholder={t("settings.ipBlocklist.reasonPlaceholder")}
                  />
                </div>
              </div>
              <div className="mt-4 flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowAddForm(false)}
                >
                  {t("settings.ipBlocklist.cancel")}
                </Button>
                <Button
                  onClick={() => blockMutation.mutate(newBlock)}
                  disabled={blockMutation.isPending || !newBlock.ip_address}
                >
                  {blockMutation.isPending ? t("settings.ipBlocklist.blocking") : t("settings.ipBlocklist.blockIp")}
                </Button>
              </div>
            </div>
          )}

          {blocks?.length === 0 ? (
            <div className="text-muted-foreground flex flex-col items-center justify-center py-12">
              <Shield className="text-muted-foreground/30 mb-3 h-12 w-12" />
              <p className="font-medium">{t("settings.ipBlocklist.empty.title")}</p>
              <p className="text-sm">{t("settings.ipBlocklist.empty.hint")}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {blocks?.map((block) => (
                <div
                  key={block.id}
                  className="hover:bg-muted flex items-center justify-between rounded-lg border p-4 transition-colors"
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono font-medium">{block.ip_address}</span>
                      {block.ip_range && (
                        <Badge
                          variant="outline"
                          className="text-xs"
                        >
                          {t("settings.ipBlocklist.item.range")}: {block.ip_range}
                        </Badge>
                      )}
                      <Badge variant={block.is_active ? "success" : "secondary"}>
                        {block.is_active
                          ? t("settings.ipBlocklist.item.active")
                          : t("settings.ipBlocklist.item.inactive")}
                      </Badge>
                    </div>
                    <p className="text-muted-foreground mt-1 text-sm">{block.reason}</p>
                    <p className="text-muted-foreground mt-1 text-xs">
                      {t("settings.ipBlocklist.item.added", {
                        date: new Date(block.created_at).toLocaleString(),
                        count: block.hit_count,
                      })}
                    </p>
                  </div>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive"
                        onClick={() => setConfirmDialog({ open: true, id: block.id, ip: block.ip_address })}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>{t("settings.ipBlocklist.tooltips.remove")}</TooltipContent>
                  </Tooltip>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={confirmDialog.open}
        onOpenChange={(open) => setConfirmDialog((prev) => ({ ...prev, open }))}
        title={t("settings.ipBlocklist.dialog.removeTitle")}
        description={t("settings.ipBlocklist.dialog.removeDescription", { ip: confirmDialog.ip })}
        onConfirm={() => unblockMutation.mutate(confirmDialog.id)}
        variant="destructive"
        confirmLabel={t("settings.ipBlocklist.dialog.unblock")}
        loading={unblockMutation.isPending}
      />
    </div>
  );
}

function SettingsSkeleton() {
  return (
    <Card>
      <CardHeader>
        <div className="bg-muted h-6 w-48 animate-pulse rounded" />
        <div className="bg-muted h-4 w-64 animate-pulse rounded" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {[...Array(4)].map((_, i) => (
            <div
              key={i}
              className="bg-muted h-10 animate-pulse rounded"
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
