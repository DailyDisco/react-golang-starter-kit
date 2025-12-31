import { useState } from "react";

import { ConfirmDialog } from "@/components/admin";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { requireAdmin } from "@/lib/guards";
import { AdminSettingsService, type Announcement, type CreateAnnouncementRequest } from "@/services/admin";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertCircle, Bell, CheckCircle, Info, Plus, Trash2, XCircle } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/announcements")({
  beforeLoad: () => requireAdmin(),
  component: AnnouncementsPage,
});

function AnnouncementsPage() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; id: number; title: string }>({
    open: false,
    id: 0,
    title: "",
  });

  const { data: announcements, isLoading } = useQuery<Announcement[]>({
    queryKey: ["admin", "announcements"],
    queryFn: () => AdminSettingsService.getAnnouncements(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateAnnouncementRequest) => AdminSettingsService.createAnnouncement(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "announcements"] });
      setShowCreateForm(false);
      toast.success(t("announcements.toast.created"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => AdminSettingsService.deleteAnnouncement(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "announcements"] });
      setDeleteConfirm({ open: false, id: 0, title: "" });
      toast.success(t("announcements.toast.deleted"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: number; is_active: boolean }) =>
      AdminSettingsService.updateAnnouncement(id, { is_active }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "announcements"] });
    },
  });

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "info":
        return <Info className="text-info h-4 w-4" />;
      case "warning":
        return <AlertCircle className="text-warning h-4 w-4" />;
      case "success":
        return <CheckCircle className="text-success h-4 w-4" />;
      case "error":
        return <XCircle className="text-destructive h-4 w-4" />;
      default:
        return <Bell className="h-4 w-4" />;
    }
  };

  const getTypeBadgeVariant = (type: string): "info" | "warning" | "success" | "destructive" | "secondary" => {
    switch (type) {
      case "info":
        return "info";
      case "warning":
        return "warning";
      case "success":
        return "success";
      case "error":
        return "destructive";
      default:
        return "secondary";
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold">{t("announcements.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("announcements.subtitle")}</p>
        </div>
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="bg-muted h-6 w-48 animate-pulse rounded" />
              </CardHeader>
              <CardContent>
                <div className="bg-muted/50 h-16 animate-pulse rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">{t("announcements.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("announcements.subtitle")}</p>
        </div>
        <Button
          onClick={() => setShowCreateForm(true)}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">{t("announcements.createAnnouncement")}</span>
        </Button>
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <Card>
          <CardHeader>
            <CardTitle>{t("announcements.create.title")}</CardTitle>
            <CardDescription>{t("announcements.create.subtitle")}</CardDescription>
          </CardHeader>
          <CardContent>
            <CreateAnnouncementForm
              onSubmit={(data) => createMutation.mutate(data)}
              onCancel={() => setShowCreateForm(false)}
              isLoading={createMutation.isPending}
            />
          </CardContent>
        </Card>
      )}

      {/* Announcements List */}
      <div className="space-y-4">
        {announcements?.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Bell className="text-muted-foreground/50 h-12 w-12" />
              <p className="mt-4 font-medium">{t("announcements.empty.title")}</p>
              <p className="text-muted-foreground text-sm">{t("announcements.empty.subtitle")}</p>
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setShowCreateForm(true)}
              >
                {t("announcements.empty.createFirst")}
              </Button>
            </CardContent>
          </Card>
        ) : (
          announcements?.map((announcement) => (
            <Card
              key={announcement.id}
              className={announcement.is_active ? "" : "opacity-60"}
            >
              <CardHeader>
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex flex-wrap items-center gap-2">
                    {getTypeIcon(announcement.type)}
                    <CardTitle className="text-lg">{announcement.title}</CardTitle>
                    <Badge variant={getTypeBadgeVariant(announcement.type)}>{announcement.type}</Badge>
                    {!announcement.is_active && <Badge variant="secondary">{t("announcements.card.inactive")}</Badge>}
                    {announcement.is_dismissible && (
                      <Badge
                        variant="outline"
                        className="text-xs"
                      >
                        {t("announcements.card.dismissible")}
                      </Badge>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-2">
                      <Label
                        htmlFor={`toggle-${announcement.id}`}
                        className="text-muted-foreground text-sm"
                      >
                        {announcement.is_active ? t("announcements.card.active") : t("announcements.card.inactive")}
                      </Label>
                      <Switch
                        id={`toggle-${announcement.id}`}
                        checked={announcement.is_active}
                        onCheckedChange={(checked) =>
                          toggleMutation.mutate({ id: announcement.id, is_active: checked })
                        }
                      />
                    </div>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={() =>
                            setDeleteConfirm({ open: true, id: announcement.id, title: announcement.title })
                          }
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>{t("announcements.tooltips.delete")}</TooltipContent>
                    </Tooltip>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">{announcement.message}</p>
                {announcement.link_url && (
                  <p className="text-primary mt-2 text-sm">
                    {t("announcements.card.link")}: {announcement.link_text || announcement.link_url}
                  </p>
                )}
                {announcement.target_roles && announcement.target_roles.length > 0 && (
                  <div className="mt-2 flex items-center gap-2">
                    <span className="text-muted-foreground text-sm">{t("announcements.card.targetRoles")}:</span>
                    <div className="flex gap-1">
                      {announcement.target_roles.map((role) => (
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
              </CardContent>
            </Card>
          ))
        )}
      </div>

      <ConfirmDialog
        open={deleteConfirm.open}
        onOpenChange={(open) => setDeleteConfirm((prev) => ({ ...prev, open }))}
        title={t("announcements.dialog.deleteTitle")}
        description={t("announcements.dialog.deleteDescription", { title: deleteConfirm.title })}
        confirmLabel={t("announcements.dialog.delete")}
        variant="destructive"
        onConfirm={() => deleteMutation.mutate(deleteConfirm.id)}
        loading={deleteMutation.isPending}
      />
    </div>
  );
}

function CreateAnnouncementForm({
  onSubmit,
  onCancel,
  isLoading,
}: {
  onSubmit: (data: CreateAnnouncementRequest) => void;
  onCancel: () => void;
  isLoading: boolean;
}) {
  const { t } = useTranslation("admin");
  const [formData, setFormData] = useState<CreateAnnouncementRequest>({
    title: "",
    message: "",
    type: "info",
    is_dismissible: true,
    is_active: true,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4"
    >
      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="title">{t("announcements.create.form.title")}</Label>
          <Input
            id="title"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            placeholder={t("announcements.create.form.titlePlaceholder")}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="type">{t("announcements.create.form.type")}</Label>
          <Select
            value={formData.type}
            onValueChange={(value) =>
              setFormData({
                ...formData,
                type: value as "info" | "warning" | "success" | "error",
              })
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="info">{t("announcements.create.form.types.info")}</SelectItem>
              <SelectItem value="warning">{t("announcements.create.form.types.warning")}</SelectItem>
              <SelectItem value="success">{t("announcements.create.form.types.success")}</SelectItem>
              <SelectItem value="error">{t("announcements.create.form.types.error")}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="message">{t("announcements.create.form.message")}</Label>
        <Textarea
          id="message"
          value={formData.message}
          onChange={(e) => setFormData({ ...formData, message: e.target.value })}
          placeholder={t("announcements.create.form.messagePlaceholder")}
          rows={3}
          required
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="link_url">{t("announcements.create.form.linkUrl")}</Label>
          <Input
            id="link_url"
            type="url"
            value={formData.link_url || ""}
            onChange={(e) => setFormData({ ...formData, link_url: e.target.value })}
            placeholder={t("announcements.create.form.linkUrlPlaceholder")}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="link_text">{t("announcements.create.form.linkText")}</Label>
          <Input
            id="link_text"
            value={formData.link_text || ""}
            onChange={(e) => setFormData({ ...formData, link_text: e.target.value })}
            placeholder={t("announcements.create.form.linkTextPlaceholder")}
          />
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-6">
        <div className="flex items-center gap-2">
          <Switch
            id="is_dismissible"
            checked={formData.is_dismissible}
            onCheckedChange={(checked) => setFormData({ ...formData, is_dismissible: checked })}
          />
          <Label
            htmlFor="is_dismissible"
            className="cursor-pointer"
          >
            {t("announcements.create.form.isDismissible")}
          </Label>
        </div>
        <div className="flex items-center gap-2">
          <Switch
            id="is_active"
            checked={formData.is_active}
            onCheckedChange={(checked) => setFormData({ ...formData, is_active: checked })}
          />
          <Label
            htmlFor="is_active"
            className="cursor-pointer"
          >
            {t("announcements.create.form.isActive")}
          </Label>
        </div>
      </div>

      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
        >
          {t("announcements.create.form.cancel")}
        </Button>
        <Button
          type="submit"
          disabled={isLoading}
        >
          {isLoading ? t("announcements.create.form.creating") : t("announcements.createAnnouncement")}
        </Button>
      </div>
    </form>
  );
}
