import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { SettingsService, type EmailNotificationSettings } from "@/services/settings/settingsService";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Bell, Loader2, Mail, Megaphone, Save, Shield, Sparkles } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/notifications")({
  component: NotificationsSettingsPage,
});

function NotificationsSettingsPage() {
  const { t } = useTranslation("settings");
  const queryClient = useQueryClient();

  const { data: preferences, isLoading } = useQuery({
    queryKey: ["user-preferences"],
    queryFn: () => SettingsService.getPreferences(),
  });

  const [notifications, setNotifications] = useState<EmailNotificationSettings>({
    marketing: false,
    security: true,
    updates: true,
    weekly_digest: false,
  });

  useEffect(() => {
    if (preferences?.email_notifications) {
      setNotifications(preferences.email_notifications);
    }
  }, [preferences]);

  const updateMutation = useMutation({
    mutationFn: (data: EmailNotificationSettings) => SettingsService.updatePreferences({ email_notifications: data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-preferences"] });
      toast.success(t("notifications.toast.saved"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleToggle = (key: keyof EmailNotificationSettings) => {
    setNotifications((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(notifications);
  };

  const hasChanges =
    preferences?.email_notifications &&
    (notifications.marketing !== preferences.email_notifications.marketing ||
      notifications.security !== preferences.email_notifications.security ||
      notifications.updates !== preferences.email_notifications.updates ||
      notifications.weekly_digest !== preferences.email_notifications.weekly_digest);

  if (isLoading) {
    return (
      <SettingsLayout>
        <div className="space-y-6">
          <div className="bg-muted h-8 w-48 animate-pulse rounded" />
          <div className="bg-muted h-64 animate-pulse rounded-lg" />
        </div>
      </SettingsLayout>
    );
  }

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold">{t("notifications.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("notifications.subtitle")}</p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="space-y-6"
        >
          {/* Email Notifications */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="h-5 w-5" />
                {t("notifications.email.title")}
              </CardTitle>
              <CardDescription>{t("notifications.email.subtitle")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Security Alerts */}
              <NotificationToggle
                icon={<Shield className="text-destructive h-5 w-5" />}
                title={t("notifications.security.title")}
                description={t("notifications.security.description")}
                checked={notifications.security}
                onCheckedChange={() => handleToggle("security")}
                recommendedLabel={t("notifications.recommended")}
                recommended
              />

              {/* Product Updates */}
              <NotificationToggle
                icon={<Sparkles className="text-primary h-5 w-5" />}
                title={t("notifications.updates.title")}
                description={t("notifications.updates.description")}
                checked={notifications.updates}
                onCheckedChange={() => handleToggle("updates")}
                recommendedLabel={t("notifications.recommended")}
              />

              {/* Marketing */}
              <NotificationToggle
                icon={<Megaphone className="text-info h-5 w-5" />}
                title={t("notifications.marketing.title")}
                description={t("notifications.marketing.description")}
                checked={notifications.marketing}
                onCheckedChange={() => handleToggle("marketing")}
                recommendedLabel={t("notifications.recommended")}
              />

              {/* Weekly Digest */}
              <NotificationToggle
                icon={<Bell className="text-warning h-5 w-5" />}
                title={t("notifications.weeklyDigest.title")}
                description={t("notifications.weeklyDigest.description")}
                checked={notifications.weekly_digest}
                onCheckedChange={() => handleToggle("weekly_digest")}
                recommendedLabel={t("notifications.recommended")}
              />
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardContent className="py-4">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-sm font-medium">{t("notifications.quickActions.title")}</p>
                  <p className="text-muted-foreground text-xs">{t("notifications.quickActions.subtitle")}</p>
                </div>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setNotifications({
                        marketing: false,
                        security: true,
                        updates: false,
                        weekly_digest: false,
                      })
                    }
                  >
                    {t("notifications.quickActions.disableOptional")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setNotifications({
                        marketing: true,
                        security: true,
                        updates: true,
                        weekly_digest: true,
                      })
                    }
                  >
                    {t("notifications.quickActions.enableAll")}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Save Button */}
          <div className="flex items-center justify-end gap-4 border-t pt-6">
            <Button
              type="submit"
              disabled={!hasChanges || updateMutation.isPending}
            >
              {updateMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {t("notifications.saving")}
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  {t("notifications.save")}
                </>
              )}
            </Button>
          </div>
        </form>
      </div>
    </SettingsLayout>
  );
}

function NotificationToggle({
  icon,
  title,
  description,
  checked,
  onCheckedChange,
  recommended,
  recommendedLabel,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  checked: boolean;
  onCheckedChange: () => void;
  recommended?: boolean;
  recommendedLabel?: string;
}) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg border p-4">
      <div className="flex items-start gap-4">
        <div className="mt-0.5">{icon}</div>
        <div>
          <div className="flex items-center gap-2">
            <Label className="font-medium">{title}</Label>
            {recommended && (
              <span className="bg-success/10 text-success rounded-full px-2 py-0.5 text-xs font-medium">
                {recommendedLabel}
              </span>
            )}
          </div>
          <p className="text-muted-foreground text-sm">{description}</p>
        </div>
      </div>
      <Switch
        checked={checked}
        onCheckedChange={onCheckedChange}
        className="shrink-0"
      />
    </div>
  );
}
