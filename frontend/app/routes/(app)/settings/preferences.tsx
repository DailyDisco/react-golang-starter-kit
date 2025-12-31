import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { cn } from "@/lib/utils";
import {
  DATE_FORMATS,
  LANGUAGES,
  SettingsService,
  TIMEZONES,
  type UpdatePreferencesRequest,
  type UserPreferences,
} from "@/services/settings/settingsService";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Calendar, Clock, Globe, Loader2, Moon, Palette, Save, Sun } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/preferences")({
  component: PreferencesSettingsPage,
});

function PreferencesSettingsPage() {
  const { t } = useTranslation("settings");
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
      toast.success(t("preferences.toast.saved"));
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
      <SettingsLayout>
        <div className="space-y-6">
          <div className="bg-muted h-8 w-48 animate-pulse rounded" />
          <div className="bg-muted h-64 animate-pulse rounded-lg" />
          <div className="bg-muted h-48 animate-pulse rounded-lg" />
        </div>
      </SettingsLayout>
    );
  }

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold">{t("preferences.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("preferences.subtitle")}</p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="space-y-6"
        >
          {/* Theme Settings */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Palette className="h-5 w-5" />
                {t("preferences.appearance.title")}
              </CardTitle>
              <CardDescription>{t("preferences.appearance.subtitle")}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 sm:grid-cols-3">
                <ThemeOption
                  value="light"
                  label={t("preferences.appearance.light")}
                  icon={<Sun className="h-5 w-5" />}
                  selected={formData.theme === "light"}
                  onClick={() => setFormData({ ...formData, theme: "light" })}
                />
                <ThemeOption
                  value="dark"
                  label={t("preferences.appearance.dark")}
                  icon={<Moon className="h-5 w-5" />}
                  selected={formData.theme === "dark"}
                  onClick={() => setFormData({ ...formData, theme: "dark" })}
                />
                <ThemeOption
                  value="system"
                  label={t("preferences.appearance.system")}
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
                {t("preferences.regional.title")}
              </CardTitle>
              <CardDescription>{t("preferences.regional.subtitle")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-6 md:grid-cols-2">
                {/* Timezone */}
                <div className="space-y-2">
                  <Label htmlFor="timezone">{t("preferences.regional.timezone")}</Label>
                  <Select
                    value={formData.timezone}
                    onValueChange={(value) => setFormData({ ...formData, timezone: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("preferences.regional.timezonePlaceholder")} />
                    </SelectTrigger>
                    <SelectContent>
                      {TIMEZONES.map((tz) => (
                        <SelectItem
                          key={tz}
                          value={tz}
                        >
                          {tz.replace(/_/g, " ")}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-muted-foreground text-xs">
                    {t("preferences.regional.currentTime")}{" "}
                    {new Date().toLocaleTimeString("en-US", { timeZone: formData.timezone })}
                  </p>
                </div>

                {/* Language */}
                <div className="space-y-2">
                  <Label htmlFor="language">{t("preferences.regional.language")}</Label>
                  <Select
                    value={formData.language}
                    onValueChange={(value) => setFormData({ ...formData, language: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("preferences.regional.languagePlaceholder")} />
                    </SelectTrigger>
                    <SelectContent>
                      {LANGUAGES.map((lang) => (
                        <SelectItem
                          key={lang.code}
                          value={lang.code}
                        >
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
                {t("preferences.dateTime.title")}
              </CardTitle>
              <CardDescription>{t("preferences.dateTime.subtitle")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-6 md:grid-cols-2">
                {/* Date Format */}
                <div className="space-y-2">
                  <Label htmlFor="dateFormat">{t("preferences.dateTime.dateFormat")}</Label>
                  <Select
                    value={formData.date_format}
                    onValueChange={(value) => setFormData({ ...formData, date_format: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("preferences.dateTime.dateFormatPlaceholder")} />
                    </SelectTrigger>
                    <SelectContent>
                      {DATE_FORMATS.map((format) => (
                        <SelectItem
                          key={format.value}
                          value={format.value}
                        >
                          {format.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Time Format */}
                <div className="space-y-2">
                  <Label>{t("preferences.dateTime.timeFormat")}</Label>
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      variant={formData.time_format === "12h" ? "default" : "outline"}
                      className="flex-1"
                      onClick={() => setFormData({ ...formData, time_format: "12h" })}
                    >
                      <Clock className="mr-2 h-4 w-4" />
                      {t("preferences.dateTime.12hour")}
                    </Button>
                    <Button
                      type="button"
                      variant={formData.time_format === "24h" ? "default" : "outline"}
                      className="flex-1"
                      onClick={() => setFormData({ ...formData, time_format: "24h" })}
                    >
                      <Clock className="mr-2 h-4 w-4" />
                      {t("preferences.dateTime.24hour")}
                    </Button>
                  </div>
                  <p className="text-muted-foreground text-xs">
                    {t("preferences.dateTime.example")}{" "}
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
                  {t("preferences.saving")}
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  {t("preferences.save")}
                </>
              )}
            </Button>
          </div>
        </form>
      </div>
    </SettingsLayout>
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
        selected ? "border-primary bg-primary/5" : "border-border hover:border-muted-foreground/30 hover:bg-muted"
      )}
    >
      <div
        className={cn("rounded-full p-3", selected ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground")}
      >
        {icon}
      </div>
      <span className={cn("font-medium", selected ? "text-primary" : "text-foreground")}>{label}</span>
    </button>
  );
}
