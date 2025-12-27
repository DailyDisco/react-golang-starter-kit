import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertCircle, Bell, CheckCircle, Info, Plus, Trash2, XCircle } from "lucide-react";
import { toast } from "sonner";

import { AdminPageHeader, ConfirmDialog } from "../../components/admin";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/ui/select";
import { Switch } from "../../components/ui/switch";
import { Textarea } from "../../components/ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "../../components/ui/tooltip";
import { requireAdmin } from "../../lib/guards";
import { AdminSettingsService, type Announcement, type CreateAnnouncementRequest } from "../../services/admin";

export const Route = createFileRoute("/admin/announcements")({
  beforeLoad: () => requireAdmin(),
  component: AnnouncementsPage,
});

function AnnouncementsPage() {
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
      toast.success("Announcement created successfully");
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
      toast.success("Announcement deleted");
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
        return <Info className="h-4 w-4 text-blue-500" />;
      case "warning":
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "error":
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Bell className="h-4 w-4" />;
    }
  };

  const getTypeBadgeVariant = (type: string): "default" | "secondary" | "outline" | "destructive" => {
    switch (type) {
      case "info":
        return "default";
      case "warning":
        return "secondary";
      case "success":
        return "outline";
      case "error":
        return "destructive";
      default:
        return "outline";
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <AdminPageHeader
          title="Announcements"
          description="Manage site-wide announcement banners"
          breadcrumbs={[{ label: "Announcements" }]}
        />
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="h-6 w-48 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
              </CardHeader>
              <CardContent>
                <div className="h-16 animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Announcements"
        description="Manage site-wide announcement banners"
        breadcrumbs={[{ label: "Announcements" }]}
        actions={
          <Button
            onClick={() => setShowCreateForm(true)}
            className="gap-2"
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Create Announcement</span>
          </Button>
        }
      />

      {/* Create Form */}
      {showCreateForm && (
        <Card>
          <CardHeader>
            <CardTitle>Create New Announcement</CardTitle>
            <CardDescription>This announcement will be shown to users across the site</CardDescription>
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
              <Bell className="h-12 w-12 text-gray-300" />
              <p className="mt-4 font-medium text-gray-900 dark:text-gray-100">No announcements yet</p>
              <p className="text-sm text-gray-500">Create your first announcement to display site-wide banners</p>
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setShowCreateForm(true)}
              >
                Create your first announcement
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
                    {!announcement.is_active && <Badge variant="secondary">Inactive</Badge>}
                    {announcement.is_dismissible && (
                      <Badge
                        variant="outline"
                        className="text-xs"
                      >
                        Dismissible
                      </Badge>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-2">
                      <Label
                        htmlFor={`toggle-${announcement.id}`}
                        className="text-muted-foreground text-sm"
                      >
                        {announcement.is_active ? "Active" : "Inactive"}
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
                          className="text-red-600 hover:text-red-700"
                          onClick={() =>
                            setDeleteConfirm({ open: true, id: announcement.id, title: announcement.title })
                          }
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>Delete announcement</TooltipContent>
                    </Tooltip>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">{announcement.message}</p>
                {announcement.link_url && (
                  <p className="mt-2 text-sm text-blue-600">Link: {announcement.link_text || announcement.link_url}</p>
                )}
                {announcement.target_roles && announcement.target_roles.length > 0 && (
                  <div className="mt-2 flex items-center gap-2">
                    <span className="text-muted-foreground text-sm">Target roles:</span>
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
        title="Delete Announcement"
        description={`Are you sure you want to delete "${deleteConfirm.title}"? This action cannot be undone.`}
        confirmLabel="Delete"
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
          <Label htmlFor="title">Title</Label>
          <Input
            id="title"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            placeholder="Announcement title"
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="type">Type</Label>
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
              <SelectItem value="info">Info</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="success">Success</SelectItem>
              <SelectItem value="error">Error</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="message">Message</Label>
        <Textarea
          id="message"
          value={formData.message}
          onChange={(e) => setFormData({ ...formData, message: e.target.value })}
          placeholder="Announcement message"
          rows={3}
          required
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="link_url">Link URL (optional)</Label>
          <Input
            id="link_url"
            type="url"
            value={formData.link_url || ""}
            onChange={(e) => setFormData({ ...formData, link_url: e.target.value })}
            placeholder="https://example.com"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="link_text">Link Text (optional)</Label>
          <Input
            id="link_text"
            value={formData.link_text || ""}
            onChange={(e) => setFormData({ ...formData, link_text: e.target.value })}
            placeholder="Learn more"
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
            Allow users to dismiss
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
            Active immediately
          </Label>
        </div>
      </div>

      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={isLoading}
        >
          {isLoading ? "Creating..." : "Create Announcement"}
        </Button>
      </div>
    </form>
  );
}
