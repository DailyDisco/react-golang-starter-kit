import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Globe, Lock, Mail, Shield, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { AdminPageHeader, ConfirmDialog } from "../../components/admin";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Switch } from "../../components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/ui/tabs";
import { Textarea } from "../../components/ui/textarea";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "../../components/ui/tooltip";
import { requireAdmin } from "../../lib/guards";
import {
  AdminSettingsService,
  type CreateIPBlockRequest,
  type EmailSettings,
  type IPBlock,
  type SecuritySettings,
  type SiteSettings,
} from "../../services/admin";

export const Route = createFileRoute("/admin/settings")({
  beforeLoad: () => requireAdmin(),
  component: AdminSettingsPage,
});

function AdminSettingsPage() {
  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Settings"
        description="Configure application-wide settings and preferences"
        breadcrumbs={[{ label: "Settings" }]}
      />

      <Tabs defaultValue="email" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2 lg:w-auto lg:grid-cols-4">
          <TabsTrigger value="email" className="gap-2">
            <Mail className="h-4 w-4" />
            <span className="hidden sm:inline">Email/SMTP</span>
            <span className="sm:hidden">Email</span>
          </TabsTrigger>
          <TabsTrigger value="security" className="gap-2">
            <Lock className="h-4 w-4" />
            <span>Security</span>
          </TabsTrigger>
          <TabsTrigger value="site" className="gap-2">
            <Globe className="h-4 w-4" />
            <span>Site</span>
          </TabsTrigger>
          <TabsTrigger value="ip-blocklist" className="gap-2">
            <Shield className="h-4 w-4" />
            <span className="hidden sm:inline">IP Blocklist</span>
            <span className="sm:hidden">IPs</span>
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
  );
}

function EmailSettingsTab() {
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
      toast.success("Email settings updated successfully");
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
        <CardTitle>Email / SMTP Configuration</CardTitle>
        <CardDescription>Configure the SMTP server for sending transactional emails</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="flex items-center justify-between rounded-lg border p-4 dark:border-gray-700">
            <div className="space-y-0.5">
              <Label htmlFor="smtp_enabled" className="text-base">Enable Email Sending</Label>
              <p className="text-sm text-muted-foreground">
                When enabled, the system will send transactional emails
              </p>
            </div>
            <Switch
              id="smtp_enabled"
              checked={formData.smtp_enabled ?? false}
              onCheckedChange={(checked) => setFormData({ ...formData, smtp_enabled: checked })}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="smtp_host">SMTP Host</Label>
              <Input
                id="smtp_host"
                value={formData.smtp_host ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_host: e.target.value })}
                placeholder="smtp.example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_port">SMTP Port</Label>
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
              <Label htmlFor="smtp_username">SMTP Username</Label>
              <Input
                id="smtp_username"
                value={formData.smtp_username ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_username: e.target.value })}
                placeholder="your-username"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_password">SMTP Password</Label>
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
              <Label htmlFor="smtp_from_email">From Email</Label>
              <Input
                id="smtp_from_email"
                type="email"
                value={formData.smtp_from_email ?? ""}
                onChange={(e) => setFormData({ ...formData, smtp_from_email: e.target.value })}
                placeholder="noreply@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_from_name">From Name</Label>
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
              {testMutation.isPending ? "Sending..." : "Send Test Email"}
            </Button>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function SecuritySettingsTab() {
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
      toast.success("Security settings updated successfully");
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
        <CardTitle>Security Settings</CardTitle>
        <CardDescription>Configure password policies, session settings, and authentication requirements</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-8">
          {/* Password Requirements */}
          <div className="space-y-4">
            <h4 className="font-medium">Password Requirements</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="password_min_length">Minimum Length</Label>
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
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <Label htmlFor="require_uppercase" className="cursor-pointer">Require uppercase</Label>
                <Switch
                  id="require_uppercase"
                  checked={formData.password_require_uppercase ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_uppercase: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <Label htmlFor="require_lowercase" className="cursor-pointer">Require lowercase</Label>
                <Switch
                  id="require_lowercase"
                  checked={formData.password_require_lowercase ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_lowercase: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <Label htmlFor="require_number" className="cursor-pointer">Require number</Label>
                <Switch
                  id="require_number"
                  checked={formData.password_require_number ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, password_require_number: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <Label htmlFor="require_special" className="cursor-pointer">Require special character</Label>
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
            <h4 className="font-medium">Session Settings</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="session_timeout_minutes">Session Timeout (minutes)</Label>
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
            <h4 className="font-medium">Login Protection</h4>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="max_login_attempts">Max Login Attempts</Label>
                <Input
                  id="max_login_attempts"
                  type="number"
                  min={1}
                  value={formData.max_login_attempts ?? 5}
                  onChange={(e) => setFormData({ ...formData, max_login_attempts: parseInt(e.target.value) })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lockout_duration_minutes">Lockout Duration (minutes)</Label>
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
            <h4 className="font-medium">Additional Security</h4>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <div>
                  <Label htmlFor="require_2fa" className="cursor-pointer">Require 2FA for admins</Label>
                  <p className="text-xs text-muted-foreground">Admin accounts must use two-factor auth</p>
                </div>
                <Switch
                  id="require_2fa"
                  checked={formData.require_2fa_for_admin ?? false}
                  onCheckedChange={(checked) => setFormData({ ...formData, require_2fa_for_admin: checked })}
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-3 dark:border-gray-700">
                <div>
                  <Label htmlFor="allow_registration" className="cursor-pointer">Allow registration</Label>
                  <p className="text-xs text-muted-foreground">New users can create accounts</p>
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
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function SiteSettingsTab() {
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
      toast.success("Site settings updated successfully");
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
        <CardTitle>Site Settings</CardTitle>
        <CardDescription>Configure site branding and general settings</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="site_name">Site Name</Label>
              <Input
                id="site_name"
                value={formData.site_name ?? ""}
                onChange={(e) => setFormData({ ...formData, site_name: e.target.value })}
                placeholder="My Application"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="site_logo_url">Logo URL</Label>
              <Input
                id="site_logo_url"
                value={formData.site_logo_url ?? ""}
                onChange={(e) => setFormData({ ...formData, site_logo_url: e.target.value })}
                placeholder="https://example.com/logo.png"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="site_description">Site Description</Label>
            <Textarea
              id="site_description"
              value={formData.site_description ?? ""}
              onChange={(e) => setFormData({ ...formData, site_description: e.target.value })}
              placeholder="A brief description of your site"
              rows={3}
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="contact_email">Contact Email</Label>
              <Input
                id="contact_email"
                type="email"
                value={formData.contact_email ?? ""}
                onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                placeholder="contact@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="support_url">Support URL</Label>
              <Input
                id="support_url"
                value={formData.support_url ?? ""}
                onChange={(e) => setFormData({ ...formData, support_url: e.target.value })}
                placeholder="https://support.example.com"
              />
            </div>
          </div>

          {/* Maintenance Mode */}
          <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4 dark:border-yellow-900 dark:bg-yellow-950">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <AlertTriangle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                <div>
                  <Label htmlFor="maintenance_mode" className="cursor-pointer font-medium text-yellow-800 dark:text-yellow-200">
                    Maintenance Mode
                  </Label>
                  <p className="text-sm text-yellow-700 dark:text-yellow-300">
                    Temporarily disable access to the site
                  </p>
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
                <Label htmlFor="maintenance_message" className="text-yellow-800 dark:text-yellow-200">
                  Maintenance Message
                </Label>
                <Textarea
                  id="maintenance_message"
                  value={formData.maintenance_message ?? ""}
                  onChange={(e) => setFormData({ ...formData, maintenance_message: e.target.value })}
                  placeholder="We're currently undergoing maintenance. Please check back soon."
                  rows={2}
                  className="border-yellow-300 bg-white dark:border-yellow-800 dark:bg-yellow-900"
                />
              </div>
            )}
          </div>

          <div className="flex justify-end">
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function IPBlocklistTab() {
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

  const blockMutation = useMutation({
    mutationFn: (data: CreateIPBlockRequest) => AdminSettingsService.blockIP(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "ip-blocklist"] });
      setShowAddForm(false);
      setNewBlock({ ip_address: "", reason: "", block_type: "manual" });
      toast.success("IP address added to blocklist");
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
      toast.success("IP address removed from blocklist");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const [newBlock, setNewBlock] = useState<CreateIPBlockRequest>({
    ip_address: "",
    reason: "",
    block_type: "manual",
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
              <CardTitle>IP Blocklist</CardTitle>
              <CardDescription>Block IP addresses from accessing the application</CardDescription>
            </div>
            <Button onClick={() => setShowAddForm(true)}>Block IP</Button>
          </div>
        </CardHeader>
        <CardContent>
          {showAddForm && (
            <div className="mb-6 rounded-lg border p-4 dark:border-gray-700">
              <h4 className="mb-4 font-medium">Add IP to Blocklist</h4>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="new_ip">IP Address</Label>
                  <Input
                    id="new_ip"
                    value={newBlock.ip_address}
                    onChange={(e) => setNewBlock({ ...newBlock, ip_address: e.target.value })}
                    placeholder="192.168.1.1"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="new_reason">Reason</Label>
                  <Input
                    id="new_reason"
                    value={newBlock.reason}
                    onChange={(e) => setNewBlock({ ...newBlock, reason: e.target.value })}
                    placeholder="Spam, abuse, etc."
                  />
                </div>
              </div>
              <div className="mt-4 flex justify-end gap-2">
                <Button variant="outline" onClick={() => setShowAddForm(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={() => blockMutation.mutate(newBlock)}
                  disabled={blockMutation.isPending || !newBlock.ip_address}
                >
                  {blockMutation.isPending ? "Blocking..." : "Block IP"}
                </Button>
              </div>
            </div>
          )}

          {blocks?.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-500">
              <Shield className="mb-3 h-12 w-12 text-gray-300" />
              <p className="font-medium">No blocked IPs</p>
              <p className="text-sm">Click "Block IP" to add an IP address to the blocklist</p>
            </div>
          ) : (
            <div className="space-y-3">
              {blocks?.map((block) => (
                <div
                  key={block.id}
                  className="flex items-center justify-between rounded-lg border p-4 transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800"
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono font-medium">{block.ip_address}</span>
                      {block.ip_range && (
                        <Badge variant="outline" className="text-xs">
                          Range: {block.ip_range}
                        </Badge>
                      )}
                      <Badge variant={block.is_active ? "default" : "secondary"}>
                        {block.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">{block.reason}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Added: {new Date(block.created_at).toLocaleString()} | Hits: {block.hit_count}
                    </p>
                  </div>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-600 hover:text-red-700"
                        onClick={() => setConfirmDialog({ open: true, id: block.id, ip: block.ip_address })}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Remove from blocklist</TooltipContent>
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
        title="Remove IP from Blocklist"
        description={`Are you sure you want to unblock ${confirmDialog.ip}? This IP will be able to access the application again.`}
        onConfirm={() => unblockMutation.mutate(confirmDialog.id)}
        variant="destructive"
        confirmLabel="Unblock"
        loading={unblockMutation.isPending}
      />
    </div>
  );
}

function SettingsSkeleton() {
  return (
    <Card>
      <CardHeader>
        <div className="h-6 w-48 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
        <div className="h-4 w-64 animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-10 animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
