import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Calendar, Clock, Globe, Loader2, Moon, Palette, Save, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Label } from "../../components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../components/ui/select";
import { requireAuth } from "../../lib/guards";
import {
  DATE_FORMATS,
  LANGUAGES,
  SettingsService,
  TIMEZONES,
  type UpdatePreferencesRequest,
  type UserPreferences,
} from "../../services/settings/settingsService";
import { cn } from "../../lib/utils";

export const Route = createFileRoute("/settings/preferences")({
  beforeLoad: () => requireAuth(),
  component: PreferencesSettingsPage,
});

function PreferencesSettingsPage() {
  const queryClient = useQueryClient();

  const { data: preferences, isLoading } = useQuery({
    queryKey: ["user-preferences"],
    queryFn: () => SettingsService.getPreferences(),
  });

  const [formData, setFormData] = useState<Partial<UserPreferences>>({
    theme: "system",
    timezone: "UTC",
    language: "en",
    date_format: "MM/DD/YYYY",
    time_format: "12h",
  });

  // Update form when preferences load
  useEffect(() => {
    if (preferences) {
      setFormData({
        theme: preferences.theme,
        timezone: preferences.timezone,
        language: preferences.language,
        date_format: preferences.date_format,
        time_format: preferences.time_format,
      });
    }
  }, [preferences]);

  const updateMutation = useMutation({
    mutationFn: (data: UpdatePreferencesRequest) => SettingsService.updatePreferences(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-preferences"] });
      toast.success("Your preferences have been saved.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate({
      theme: formData.theme,
      timezone: formData.timezone,
      language: formData.language,
      date_format: formData.date_format,
      time_format: formData.time_format,
    });
  };

  const hasChanges =
    preferences &&
    (formData.theme !== preferences.theme ||
      formData.timezone !== preferences.timezone ||
      formData.language !== preferences.language ||
      formData.date_format !== preferences.date_format ||
      formData.time_format !== preferences.time_format);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse rounded bg-gray-200" />
        <div className="h-64 animate-pulse rounded-lg bg-gray-200" />
        <div className="h-48 animate-pulse rounded-lg bg-gray-200" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Preferences</h2>
        <p className="text-sm text-gray-500">
          Customize your experience with theme, language, and display settings
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Theme Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Palette className="h-5 w-5" />
              Appearance
            </CardTitle>
            <CardDescription>
              Choose your preferred color theme
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-3">
              <ThemeOption
                value="light"
                label="Light"
                icon={<Sun className="h-5 w-5" />}
                selected={formData.theme === "light"}
                onClick={() => setFormData({ ...formData, theme: "light" })}
              />
              <ThemeOption
                value="dark"
                label="Dark"
                icon={<Moon className="h-5 w-5" />}
                selected={formData.theme === "dark"}
                onClick={() => setFormData({ ...formData, theme: "dark" })}
              />
              <ThemeOption
                value="system"
                label="System"
                icon={
                  <div className="relative">
                    <Sun className="h-4 w-4" />
                    <Moon className="absolute -right-1 -bottom-1 h-3 w-3" />
                  </div>
                }
                selected={formData.theme === "system"}
                onClick={() => setFormData({ ...formData, theme: "system" })}
              />
            </div>
          </CardContent>
        </Card>

        {/* Regional Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Regional Settings
            </CardTitle>
            <CardDescription>
              Set your timezone and language preferences
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
              {/* Timezone */}
              <div className="space-y-2">
                <Label htmlFor="timezone">Timezone</Label>
                <Select
                  value={formData.timezone}
                  onValueChange={(value) => setFormData({ ...formData, timezone: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select timezone" />
                  </SelectTrigger>
                  <SelectContent>
                    {TIMEZONES.map((tz) => (
                      <SelectItem key={tz} value={tz}>
                        {tz.replace(/_/g, " ")}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-gray-500">
                  Current time: {new Date().toLocaleTimeString("en-US", { timeZone: formData.timezone })}
                </p>
              </div>

              {/* Language */}
              <div className="space-y-2">
                <Label htmlFor="language">Language</Label>
                <Select
                  value={formData.language}
                  onValueChange={(value) => setFormData({ ...formData, language: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select language" />
                  </SelectTrigger>
                  <SelectContent>
                    {LANGUAGES.map((lang) => (
                      <SelectItem key={lang.code} value={lang.code}>
                        {lang.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Date & Time Format */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="h-5 w-5" />
              Date & Time Format
            </CardTitle>
            <CardDescription>
              Choose how dates and times are displayed
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
              {/* Date Format */}
              <div className="space-y-2">
                <Label htmlFor="dateFormat">Date Format</Label>
                <Select
                  value={formData.date_format}
                  onValueChange={(value) => setFormData({ ...formData, date_format: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select date format" />
                  </SelectTrigger>
                  <SelectContent>
                    {DATE_FORMATS.map((format) => (
                      <SelectItem key={format.value} value={format.value}>
                        {format.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Time Format */}
              <div className="space-y-2">
                <Label>Time Format</Label>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    variant={formData.time_format === "12h" ? "default" : "outline"}
                    className="flex-1"
                    onClick={() => setFormData({ ...formData, time_format: "12h" })}
                  >
                    <Clock className="mr-2 h-4 w-4" />
                    12-hour
                  </Button>
                  <Button
                    type="button"
                    variant={formData.time_format === "24h" ? "default" : "outline"}
                    className="flex-1"
                    onClick={() => setFormData({ ...formData, time_format: "24h" })}
                  >
                    <Clock className="mr-2 h-4 w-4" />
                    24-hour
                  </Button>
                </div>
                <p className="text-xs text-gray-500">
                  Example:{" "}
                  {formData.time_format === "12h"
                    ? new Date().toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit", hour12: true })
                    : new Date().toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", hour12: false })}
                </p>
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

function ThemeOption({
  value,
  label,
  icon,
  selected,
  onClick,
}: {
  value: string;
  label: string;
  icon: React.ReactNode;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex flex-col items-center gap-3 rounded-lg border-2 p-6 transition-all",
        selected
          ? "border-blue-500 bg-blue-50"
          : "border-gray-200 hover:border-gray-300 hover:bg-gray-50"
      )}
    >
      <div
        className={cn(
          "rounded-full p-3",
          selected ? "bg-blue-100 text-blue-600" : "bg-gray-100 text-gray-600"
        )}
      >
        {icon}
      </div>
      <span className={cn("font-medium", selected ? "text-blue-700" : "text-gray-700")}>
        {label}
      </span>
    </button>
  );
}
