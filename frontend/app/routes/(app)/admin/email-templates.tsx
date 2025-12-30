import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { AdminSettingsService, type EmailTemplate, type UpdateEmailTemplateRequest } from "@/services/admin";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Eye, Loader2, Mail, Save, Variable } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/email-templates")({
  component: EmailTemplatesPage,
});

function EmailTemplatesPage() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();
  const [selectedTemplate, setSelectedTemplate] = useState<EmailTemplate | null>(null);
  const [previewHtml, setPreviewHtml] = useState<string>("");
  const [showPreview, setShowPreview] = useState(false);

  const { data: templates, isLoading } = useQuery<EmailTemplate[]>({
    queryKey: ["admin", "email-templates"],
    queryFn: () => AdminSettingsService.getEmailTemplates(),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateEmailTemplateRequest }) =>
      AdminSettingsService.updateEmailTemplate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "email-templates"] });
      toast.success(t("emailTemplates.toast.updated"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const previewMutation = useMutation({
    mutationFn: ({ id, variables }: { id: number; variables: Record<string, string> }) =>
      AdminSettingsService.previewEmailTemplate(id, variables),
    onSuccess: (data) => {
      setPreviewHtml(data.body_html);
      setShowPreview(true);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold">{t("emailTemplates.title")}</h2>
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="space-y-4">
            {[...Array(4)].map((_, i) => (
              <div
                key={i}
                className="bg-muted h-20 animate-pulse rounded-lg"
              />
            ))}
          </div>
          <div className="lg:col-span-2">
            <div className="bg-muted h-96 animate-pulse rounded-lg" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold">{t("emailTemplates.title")}</h2>
        <p className="text-muted-foreground text-sm">{t("emailTemplates.subtitle")}</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Template List */}
        <div className="space-y-2">
          {templates?.map((template) => (
            <button
              key={template.id}
              onClick={() => setSelectedTemplate(template)}
              className={`w-full rounded-lg border p-4 text-left transition-colors ${
                selectedTemplate?.id === template.id
                  ? "border-primary bg-primary/5"
                  : "hover:border-muted-foreground/30 hover:bg-muted"
              }`}
            >
              <div className="flex items-center gap-2">
                <Mail className="text-muted-foreground h-4 w-4" />
                <span className="font-medium">{template.name}</span>
              </div>
              <p className="text-muted-foreground mt-1 text-xs">{template.key}</p>
              <div className="mt-2 flex gap-2">
                {template.is_system && (
                  <Badge
                    variant="secondary"
                    className="text-xs"
                  >
                    {t("emailTemplates.badges.system")}
                  </Badge>
                )}
                <Badge
                  variant={template.is_active ? "success" : "secondary"}
                  className="text-xs"
                >
                  {template.is_active ? t("emailTemplates.badges.active") : t("emailTemplates.badges.inactive")}
                </Badge>
              </div>
            </button>
          ))}
        </div>

        {/* Editor */}
        <div className="lg:col-span-2">
          {selectedTemplate ? (
            <TemplateEditor
              template={selectedTemplate}
              onSave={(data) => updateMutation.mutate({ id: selectedTemplate.id, data })}
              onPreview={(variables) => previewMutation.mutate({ id: selectedTemplate.id, variables })}
              isSaving={updateMutation.isPending}
              isPreviewing={previewMutation.isPending}
            />
          ) : (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Mail className="text-muted-foreground/30 h-12 w-12" />
                <p className="text-muted-foreground mt-4">{t("emailTemplates.selectTemplate")}</p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* Preview Modal */}
      {showPreview && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-background max-h-[80vh] w-full max-w-3xl overflow-auto rounded-lg p-6">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-semibold">{t("emailTemplates.preview.title")}</h3>
              <Button
                variant="ghost"
                onClick={() => setShowPreview(false)}
              >
                {t("emailTemplates.preview.close")}
              </Button>
            </div>
            <div
              className="bg-background rounded border p-4"
              dangerouslySetInnerHTML={{ __html: previewHtml }}
            />
          </div>
        </div>
      )}
    </div>
  );
}

function TemplateEditor({
  template,
  onSave,
  onPreview,
  isSaving,
  isPreviewing,
}: {
  template: EmailTemplate;
  onSave: (data: UpdateEmailTemplateRequest) => void;
  onPreview: (variables: Record<string, string>) => void;
  isSaving: boolean;
  isPreviewing: boolean;
}) {
  const { t } = useTranslation("admin");
  const [formData, setFormData] = useState({
    name: template.name,
    description: template.description || "",
    subject: template.subject,
    body_html: template.body_html,
    body_text: template.body_text || "",
    is_active: template.is_active,
  });

  const [previewVars, setPreviewVars] = useState<Record<string, string>>(() => {
    const vars: Record<string, string> = {};
    if (template.available_variables)
      for (const v of template.available_variables) {
        vars[v.name] = `[${v.name}]`;
      }
    return vars;
  });

  // Update form when template changes
  if (formData.name !== template.name && formData.subject === template.subject) {
    setFormData({
      name: template.name,
      description: template.description || "",
      subject: template.subject,
      body_html: template.body_html,
      body_text: template.body_text || "",
      is_active: template.is_active,
    });
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{template.name}</CardTitle>
            <CardDescription>{template.description}</CardDescription>
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onPreview(previewVars)}
              disabled={isPreviewing}
            >
              {isPreviewing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Eye className="mr-2 h-4 w-4" />}
              {isPreviewing ? t("emailTemplates.preview.loading") : t("emailTemplates.preview.button")}
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isSaving}
            >
              {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
              {isSaving ? t("emailTemplates.save.saving") : t("emailTemplates.save.button")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <form className="space-y-6">
          {/* Available Variables */}
          {template.available_variables && template.available_variables.length > 0 && (
            <div className="bg-muted rounded-lg p-4">
              <div className="mb-2 flex items-center gap-2">
                <Variable className="h-4 w-4" />
                <span className="text-sm font-medium">{t("emailTemplates.variables.title")}</span>
              </div>
              <div className="flex flex-wrap gap-2">
                {template.available_variables.map((variable) => (
                  <Badge
                    key={variable.name}
                    variant="outline"
                    className="cursor-pointer"
                    title={variable.description}
                  >
                    {"{{"}
                    {variable.name}
                    {"}}"}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">{t("emailTemplates.form.templateName")}</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>
            <div className="flex items-center gap-2 pt-8">
              <input
                type="checkbox"
                id="is_active"
                checked={formData.is_active}
                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="is_active">{t("emailTemplates.form.active")}</Label>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="subject">{t("emailTemplates.form.subjectLine")}</Label>
            <Input
              id="subject"
              value={formData.subject}
              onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="body_html">{t("emailTemplates.form.htmlBody")}</Label>
            <Textarea
              id="body_html"
              value={formData.body_html}
              onChange={(e) => setFormData({ ...formData, body_html: e.target.value })}
              rows={12}
              className="font-mono text-sm"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="body_text">{t("emailTemplates.form.plainTextBody")}</Label>
            <Textarea
              id="body_text"
              value={formData.body_text}
              onChange={(e) => setFormData({ ...formData, body_text: e.target.value })}
              rows={6}
              className="font-mono text-sm"
            />
          </div>

          {/* Preview Variables */}
          <div className="rounded-lg border p-4">
            <h4 className="mb-3 text-sm font-medium">{t("emailTemplates.form.previewVariables")}</h4>
            <div className="grid gap-3 md:grid-cols-2">
              {template.available_variables?.map((variable) => (
                <div
                  key={variable.name}
                  className="space-y-1"
                >
                  <Label
                    htmlFor={`var-${variable.name}`}
                    className="text-xs"
                  >
                    {variable.name}
                  </Label>
                  <Input
                    id={`var-${variable.name}`}
                    value={previewVars[variable.name] || ""}
                    onChange={(e) => setPreviewVars({ ...previewVars, [variable.name]: e.target.value })}
                    placeholder={variable.description}
                    className="text-sm"
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Stats */}
          <div className="bg-muted text-muted-foreground rounded-lg p-4 text-sm">
            <div className="flex gap-6">
              <div>
                <span className="font-medium">{t("emailTemplates.stats.sendCount")}:</span> {template.send_count}
              </div>
              {template.last_sent_at && (
                <div>
                  <span className="font-medium">{t("emailTemplates.stats.lastSent")}:</span>{" "}
                  {new Date(template.last_sent_at).toLocaleString()}
                </div>
              )}
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
