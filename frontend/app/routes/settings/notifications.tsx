import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Bell, Loader2, Mail, Megaphone, Save, Shield, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Label } from "../../components/ui/label";
import { Switch } from "../../components/ui/switch";
import { requireAuth } from "../../lib/guards";
import {
  SettingsService,
  type EmailNotificationSettings,
} from "../../services/settings/settingsService";

export const Route = createFileRoute("/settings/notifications")({
  beforeLoad: () => requireAuth(),
  component: NotificationsSettingsPage,
});

function NotificationsSettingsPage() {
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

  // Update form when preferences load
  useEffect(() => {
    if (preferences?.email_notifications) {
      setNotifications(preferences.email_notifications);
    }
  }, [preferences]);

  const updateMutation = useMutation({
    mutationFn: (data: EmailNotificationSettings) =>
      SettingsService.updatePreferences({ email_notifications: data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-preferences"] });
      toast.success("Notification preferences have been saved.");
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
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse rounded bg-gray-200" />
        <div className="h-64 animate-pulse rounded-lg bg-gray-200" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Notifications</h2>
        <p className="text-sm text-gray-500">
          Manage your email notification preferences
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Email Notifications */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Email Notifications
            </CardTitle>
            <CardDescription>
              Choose which emails you'd like to receive
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Security Alerts */}
            <NotificationToggle
              icon={<Shield className="h-5 w-5 text-red-500" />}
              title="Security Alerts"
              description="Get notified about security events like new logins, password changes, and suspicious activity"
              checked={notifications.security}
              onCheckedChange={() => handleToggle("security")}
              recommended
            />

            {/* Product Updates */}
            <NotificationToggle
              icon={<Sparkles className="h-5 w-5 text-purple-500" />}
              title="Product Updates"
              description="Stay informed about new features, improvements, and important product announcements"
              checked={notifications.updates}
              onCheckedChange={() => handleToggle("updates")}
            />

            {/* Marketing */}
            <NotificationToggle
              icon={<Megaphone className="h-5 w-5 text-blue-500" />}
              title="Marketing & Promotions"
              description="Receive special offers, tips, and promotional content"
              checked={notifications.marketing}
              onCheckedChange={() => handleToggle("marketing")}
            />

            {/* Weekly Digest */}
            <NotificationToggle
              icon={<Bell className="h-5 w-5 text-amber-500" />}
              title="Weekly Digest"
              description="Get a weekly summary of your account activity and important updates"
              checked={notifications.weekly_digest}
              onCheckedChange={() => handleToggle("weekly_digest")}
            />
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card>
          <CardContent className="py-4">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div>
                <p className="text-sm font-medium">Quick Actions</p>
                <p className="text-xs text-gray-500">
                  Quickly enable or disable all optional notifications
                </p>
              </div>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setNotifications({
                      marketing: false,
                      security: true, // Keep security enabled
                      updates: false,
                      weekly_digest: false,
                    })
                  }
                >
                  Disable Optional
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
                  Enable All
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
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Preferences
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}

function NotificationToggle({
  icon,
  title,
  description,
  checked,
  onCheckedChange,
  recommended,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  checked: boolean;
  onCheckedChange: () => void;
  recommended?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg border p-4">
      <div className="flex items-start gap-4">
        <div className="mt-0.5">{icon}</div>
        <div>
          <div className="flex items-center gap-2">
            <Label className="font-medium">{title}</Label>
            {recommended && (
              <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                Recommended
              </span>
            )}
          </div>
          <p className="text-sm text-gray-500">{description}</p>
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
