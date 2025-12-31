import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { AdminLayout } from "@/layouts/AdminLayout";
import { AdminSettingsService, type EmailTemplate, type UpdateEmailTemplateRequest } from "@/services/admin";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  AlertCircle,
  CheckCircle,
  ChevronRight,
  Clock,
  Code,
  Copy,
  Eye,
  FileCode,
  FileText,
  FolderOpen,
  Loader2,
  Mail,
  MailCheck,
  Monitor,
  PanelLeftClose,
  PanelRightClose,
  RefreshCw,
  Rocket,
  Save,
  Search,
  Send,
  Shield,
  Smartphone,
  Sparkles,
  Tablet,
  X,
  Zap,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/admin/email-templates")({
  component: EmailTemplatesPage,
});

const SUBJECT_MAX_LENGTH = 78;
const SUBJECT_WARN_LENGTH = 60;

type DeviceSize = "desktop" | "tablet" | "mobile";
type EditorTab = "html" | "text" | "preview";
type LayoutMode = "sidebar" | "split";

const DEVICE_WIDTHS: Record<DeviceSize, string> = {
  desktop: "100%",
  tablet: "768px",
  mobile: "375px",
};

// Template categories based on common email types
const TEMPLATE_CATEGORIES: Record<string, { icon: typeof Mail; color: string; label: string }> = {
  welcome: { icon: Sparkles, color: "text-emerald-500", label: "Welcome" },
  password: { icon: Shield, color: "text-amber-500", label: "Password" },
  verification: { icon: MailCheck, color: "text-blue-500", label: "Verification" },
  notification: { icon: Zap, color: "text-purple-500", label: "Notification" },
  security: { icon: Shield, color: "text-red-500", label: "Security" },
  default: { icon: Mail, color: "text-slate-500", label: "General" },
};

function getCategoryFromKey(key: string): keyof typeof TEMPLATE_CATEGORIES {
  const lowerKey = key.toLowerCase();
  if (lowerKey.includes("welcome")) return "welcome";
  if (lowerKey.includes("password") || lowerKey.includes("reset")) return "password";
  if (lowerKey.includes("verif")) return "verification";
  if (lowerKey.includes("security") || lowerKey.includes("2fa") || lowerKey.includes("login")) return "security";
  if (lowerKey.includes("notif")) return "notification";
  return "default";
}

function EmailTemplatesPage() {
  const { t } = useTranslation("admin");
  const queryClient = useQueryClient();
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [layoutMode, setLayoutMode] = useState<LayoutMode>("sidebar");
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const {
    data: templates,
    isLoading,
    refetch,
    isRefetching,
  } = useQuery<EmailTemplate[]>({
    queryKey: ["admin", "email-templates"],
    queryFn: () => AdminSettingsService.getEmailTemplates(),
  });

  const selectedTemplate = templates?.find((t) => t.id === selectedTemplateId) ?? null;

  // Group templates by category
  const groupedTemplates = useMemo(() => {
    if (!templates) return {};

    const filtered = searchQuery.trim()
      ? templates.filter(
          (t) =>
            t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            t.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
            t.description?.toLowerCase().includes(searchQuery.toLowerCase())
        )
      : templates;

    return filtered.reduce(
      (acc, template) => {
        const category = getCategoryFromKey(template.key);
        if (!acc[category]) acc[category] = [];
        acc[category].push(template);
        return acc;
      },
      {} as Record<string, EmailTemplate[]>
    );
  }, [templates, searchQuery]);

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

  // Auto-select first template
  useEffect(() => {
    if (templates && templates.length > 0 && !selectedTemplateId) {
      setSelectedTemplateId(templates[0].id);
    }
  }, [templates, selectedTemplateId]);

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="bg-muted h-7 w-40 animate-pulse rounded" />
            <div className="bg-muted h-8 w-24 animate-pulse rounded" />
          </div>
          <div className="grid gap-4 lg:grid-cols-12">
            <div className="space-y-2 lg:col-span-3">
              {[...Array(5)].map((_, i) => (
                <div
                  key={i}
                  className="bg-muted h-14 animate-pulse rounded-lg"
                />
              ))}
            </div>
            <div className="lg:col-span-9">
              <div className="bg-muted h-[500px] animate-pulse rounded-lg" />
            </div>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <TooltipProvider>
        <div className="space-y-4">
          {/* Header - More compact */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h2 className="text-xl font-semibold">{t("emailTemplates.title")}</h2>
              <Badge
                variant="outline"
                className="gap-1 text-xs"
              >
                <Mail className="h-3 w-3" />
                {templates?.length || 0}
              </Badge>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="flex items-center rounded-md border p-0.5">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={layoutMode === "sidebar" ? "secondary" : "ghost"}
                      size="sm"
                      className="h-7 w-7 p-0"
                      onClick={() => setLayoutMode("sidebar")}
                    >
                      <PanelLeftClose className="h-3.5 w-3.5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Sidebar layout</TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={layoutMode === "split" ? "secondary" : "ghost"}
                      size="sm"
                      className="h-7 w-7 p-0"
                      onClick={() => setLayoutMode("split")}
                    >
                      <PanelRightClose className="h-3.5 w-3.5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Split view</TooltipContent>
                </Tooltip>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => refetch()}
                disabled={isRefetching}
                className="h-8 gap-1.5 px-2.5"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${isRefetching ? "animate-spin" : ""}`} />
              </Button>
            </div>
          </div>

          <div className={`grid gap-4 ${layoutMode === "split" ? "lg:grid-cols-2" : "lg:grid-cols-12"}`}>
            {/* Template List - More compact */}
            <div className={`${layoutMode === "split" ? "" : sidebarCollapsed ? "lg:col-span-1" : "lg:col-span-3"}`}>
              <Card className="sticky top-4 overflow-hidden">
                {sidebarCollapsed && layoutMode === "sidebar" ? (
                  <div className="flex flex-col items-center py-3">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSidebarCollapsed(false)}
                          className="mb-2"
                        >
                          <ChevronRight className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent side="right">Expand</TooltipContent>
                    </Tooltip>
                    {templates?.slice(0, 10).map((template) => {
                      const category = getCategoryFromKey(template.key);
                      const CategoryIcon = TEMPLATE_CATEGORIES[category].icon;
                      return (
                        <Tooltip key={template.id}>
                          <TooltipTrigger asChild>
                            <button
                              onClick={() => setSelectedTemplateId(template.id)}
                              className={`mb-1 flex h-8 w-8 items-center justify-center rounded-md transition-all ${
                                selectedTemplateId === template.id
                                  ? "bg-primary text-primary-foreground"
                                  : "hover:bg-muted"
                              }`}
                            >
                              <CategoryIcon
                                className={`h-3.5 w-3.5 ${selectedTemplateId !== template.id ? TEMPLATE_CATEGORIES[category].color : ""}`}
                              />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="right">{template.name}</TooltipContent>
                        </Tooltip>
                      );
                    })}
                  </div>
                ) : (
                  <>
                    <div className="flex items-center justify-between border-b p-2.5">
                      <span className="text-sm font-medium">Templates</span>
                      {layoutMode === "sidebar" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0"
                          onClick={() => setSidebarCollapsed(true)}
                        >
                          <PanelLeftClose className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </div>
                    <div className="border-b p-2">
                      <div className="relative">
                        <Search className="text-muted-foreground absolute top-1/2 left-2.5 h-3.5 w-3.5 -translate-y-1/2" />
                        <Input
                          placeholder="Search..."
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                          className="h-8 pr-8 pl-8 text-sm"
                        />
                        {searchQuery && (
                          <button
                            onClick={() => setSearchQuery("")}
                            className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 -translate-y-1/2"
                          >
                            <X className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </div>
                    <ScrollArea className="h-[calc(100vh-280px)]">
                      <div className="p-2">
                        {Object.keys(groupedTemplates).length === 0 ? (
                          <div className="flex flex-col items-center justify-center py-8 text-center">
                            <Mail className="text-muted-foreground/40 mb-2 h-8 w-8" />
                            <p className="text-muted-foreground text-sm">{t("emailTemplates.noResults")}</p>
                          </div>
                        ) : (
                          Object.entries(groupedTemplates).map(([category, categoryTemplates]) => {
                            const categoryConfig = TEMPLATE_CATEGORIES[category];
                            const CategoryIcon = categoryConfig.icon;
                            return (
                              <div
                                key={category}
                                className="mb-3"
                              >
                                <div className="mb-1.5 flex items-center gap-1.5 px-1">
                                  <CategoryIcon className={`h-3 w-3 ${categoryConfig.color}`} />
                                  <span className="text-muted-foreground text-[10px] font-medium tracking-wide uppercase">
                                    {categoryConfig.label}
                                  </span>
                                </div>
                                <div className="space-y-0.5">
                                  {categoryTemplates.map((template) => (
                                    <TemplateListItem
                                      key={template.id}
                                      template={template}
                                      isSelected={selectedTemplateId === template.id}
                                      onSelect={() => setSelectedTemplateId(template.id)}
                                    />
                                  ))}
                                </div>
                              </div>
                            );
                          })
                        )}
                      </div>
                    </ScrollArea>
                  </>
                )}
              </Card>
            </div>

            {/* Editor */}
            <div className={layoutMode === "split" ? "" : sidebarCollapsed ? "lg:col-span-11" : "lg:col-span-9"}>
              {selectedTemplate ? (
                <TemplateEditor
                  key={selectedTemplate.id}
                  template={selectedTemplate}
                  onSave={(data) => updateMutation.mutate({ id: selectedTemplate.id, data })}
                  isSaving={updateMutation.isPending}
                  layoutMode={layoutMode}
                />
              ) : (
                <EmptyState />
              )}
            </div>
          </div>
        </div>
      </TooltipProvider>
    </AdminLayout>
  );
}

interface TemplateListItemProps {
  template: EmailTemplate;
  isSelected: boolean;
  onSelect: () => void;
}

function TemplateListItem({ template, isSelected, onSelect }: TemplateListItemProps) {
  return (
    <button
      onClick={onSelect}
      className={`group w-full rounded-lg p-2 text-left transition-all ${
        isSelected ? "bg-primary text-primary-foreground" : "hover:bg-muted/80"
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="min-w-0 flex-1">
          <span className="block truncate text-sm font-medium">{template.name}</span>
          <span
            className={`block truncate font-mono text-[10px] ${
              isSelected ? "text-primary-foreground/70" : "text-muted-foreground"
            }`}
          >
            {template.key}
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          {template.send_count > 0 && (
            <span className={`text-[10px] ${isSelected ? "text-primary-foreground/60" : "text-muted-foreground"}`}>
              {template.send_count > 999 ? `${(template.send_count / 1000).toFixed(1)}k` : template.send_count}
            </span>
          )}
          <div className={`h-2 w-2 rounded-full ${template.is_active ? "bg-emerald-500" : "bg-muted-foreground/30"}`} />
        </div>
      </div>
    </button>
  );
}

function EmptyState() {
  const { t } = useTranslation("admin");

  return (
    <Card className="flex h-[400px] flex-col items-center justify-center">
      <div className="text-center">
        <div className="from-primary/10 to-primary/5 mx-auto flex h-14 w-14 items-center justify-center rounded-xl bg-linear-to-br">
          <Mail className="text-primary h-7 w-7" />
        </div>
        <h3 className="mt-4 font-semibold">{t("emailTemplates.selectTemplate")}</h3>
        <p className="text-muted-foreground mt-1 text-sm">{t("emailTemplates.selectTemplateHint")}</p>
      </div>
    </Card>
  );
}

interface TemplateEditorProps {
  template: EmailTemplate;
  onSave: (data: UpdateEmailTemplateRequest) => void;
  isSaving: boolean;
  layoutMode: LayoutMode;
}

function TemplateEditor({ template, onSave, isSaving, layoutMode }: TemplateEditorProps) {
  const { t } = useTranslation("admin");
  const htmlTextareaRef = useRef<HTMLTextAreaElement>(null);
  const textTextareaRef = useRef<HTMLTextAreaElement>(null);

  // Form state
  const [formData, setFormData] = useState({
    subject: template.subject,
    body_html: template.body_html,
    body_text: template.body_text || "",
    is_active: template.is_active,
  });

  // UI state
  const [activeTab, setActiveTab] = useState<EditorTab>("html");

  // Preview state
  const [previewVars, setPreviewVars] = useState<Record<string, string>>(() => {
    const vars: Record<string, string> = {};
    for (const v of template.available_variables ?? []) {
      vars[v.name] = getSampleValue(v.name);
    }
    return vars;
  });
  const [showPreview, setShowPreview] = useState(false);
  const [previewTab, setPreviewTab] = useState<"html" | "text">("html");
  const [deviceSize, setDeviceSize] = useState<DeviceSize>("desktop");
  const [previewData, setPreviewData] = useState<{
    subject: string;
    body_html: string;
    body_text: string;
  } | null>(null);

  // Test email state
  const [showTestEmail, setShowTestEmail] = useState(false);
  const [testEmailAddress, setTestEmailAddress] = useState("");

  // Track if form has changes
  const hasChanges =
    formData.subject !== template.subject ||
    formData.body_html !== template.body_html ||
    formData.body_text !== (template.body_text || "") ||
    formData.is_active !== template.is_active;

  // Preview mutation
  const previewMutation = useMutation({
    mutationFn: ({ id, variables }: { id: number; variables: Record<string, string> }) =>
      AdminSettingsService.previewEmailTemplate(id, variables),
    onSuccess: (data) => {
      setPreviewData(data);
      if (layoutMode === "sidebar") {
        setShowPreview(true);
      }
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Auto-preview in split mode
  useEffect(() => {
    if (layoutMode === "split" && !previewData) {
      previewMutation.mutate({ id: template.id, variables: previewVars });
    }
  }, [layoutMode, template.id]);

  // Subject line character count helpers
  const subjectLength = formData.subject.length;
  const subjectStatus =
    subjectLength > SUBJECT_MAX_LENGTH ? "error" : subjectLength > SUBJECT_WARN_LENGTH ? "warning" : "ok";

  // Insert variable at cursor position
  const insertVariable = useCallback(
    (variableName: string, target: "html" | "text") => {
      const textarea = target === "html" ? htmlTextareaRef.current : textTextareaRef.current;
      if (!textarea) return;

      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const field = target === "html" ? "body_html" : "body_text";
      const currentValue = formData[field];
      const insertion = `{{${variableName}}}`;

      const newValue = currentValue.substring(0, start) + insertion + currentValue.substring(end);

      setFormData((prev) => ({ ...prev, [field]: newValue }));

      setTimeout(() => {
        textarea.focus();
        textarea.setSelectionRange(start + insertion.length, start + insertion.length);
      }, 0);
    },
    [formData]
  );

  // Copy variable to clipboard
  const copyVariable = useCallback(
    (variableName: string) => {
      navigator.clipboard.writeText(`{{${variableName}}}`);
      toast.success(t("emailTemplates.variables.copied"));
    },
    [t]
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const handlePreview = () => {
    previewMutation.mutate({ id: template.id, variables: previewVars });
  };

  const handleSendTestEmail = () => {
    toast.success(t("emailTemplates.testEmail.sent", { email: testEmailAddress }));
    setShowTestEmail(false);
    setTestEmailAddress("");
  };

  const category = getCategoryFromKey(template.key);
  const CategoryIcon = TEMPLATE_CATEGORIES[category].icon;

  return (
    <TooltipProvider>
      <div className="space-y-3">
        {/* Header Bar - Compact */}
        <Card>
          <CardContent className="p-3">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div
                  className={`flex h-9 w-9 items-center justify-center rounded-lg ${
                    template.is_active
                      ? "bg-linear-to-br from-emerald-100 to-emerald-50 dark:from-emerald-900/30 dark:to-emerald-900/10"
                      : "bg-muted"
                  }`}
                >
                  <CategoryIcon
                    className={`h-4 w-4 ${template.is_active ? TEMPLATE_CATEGORIES[category].color : "text-muted-foreground"}`}
                  />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold">{template.name}</h3>
                    {hasChanges && (
                      <Badge
                        variant="outline"
                        className="h-5 border-amber-200 bg-amber-50 px-1.5 text-[10px] text-amber-700"
                      >
                        Unsaved
                      </Badge>
                    )}
                  </div>
                  <p className="text-muted-foreground text-xs">{template.key}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-2 rounded-md border px-2 py-1">
                  <Switch
                    id="is_active"
                    checked={formData.is_active}
                    onCheckedChange={(checked) => setFormData((prev) => ({ ...prev, is_active: checked }))}
                    className="h-4 w-7"
                  />
                  <Label
                    htmlFor="is_active"
                    className="cursor-pointer text-xs"
                  >
                    {formData.is_active ? "Active" : "Inactive"}
                  </Label>
                </div>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => setShowTestEmail(true)}
                    >
                      <Send className="h-3.5 w-3.5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Send test</TooltipContent>
                </Tooltip>
                {layoutMode === "sidebar" && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-8 w-8 p-0"
                        onClick={handlePreview}
                        disabled={previewMutation.isPending}
                      >
                        {previewMutation.isPending ? (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <Eye className="h-3.5 w-3.5" />
                        )}
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Preview</TooltipContent>
                  </Tooltip>
                )}
                <Button
                  size="sm"
                  onClick={handleSubmit}
                  disabled={isSaving || !hasChanges}
                  className="h-8 gap-1.5 px-3"
                >
                  {isSaving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Save className="h-3.5 w-3.5" />}
                  Save
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Main Editor Area */}
        <div className={layoutMode === "split" ? "grid gap-3 lg:grid-cols-2" : ""}>
          <div className="space-y-3">
            {/* Subject + Variables Combined */}
            <Card>
              <CardContent className="p-3">
                <div className="space-y-3">
                  {/* Subject */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <Label
                        htmlFor="subject"
                        className="text-xs font-medium"
                      >
                        Subject Line
                      </Label>
                      <span
                        className={`text-[10px] font-medium ${
                          subjectStatus === "error"
                            ? "text-red-600"
                            : subjectStatus === "warning"
                              ? "text-amber-600"
                              : "text-muted-foreground"
                        }`}
                      >
                        {subjectLength}/{SUBJECT_MAX_LENGTH}
                      </span>
                    </div>
                    <Input
                      id="subject"
                      value={formData.subject}
                      onChange={(e) => setFormData((prev) => ({ ...prev, subject: e.target.value }))}
                      className={`h-9 text-sm ${subjectStatus === "error" ? "border-red-500" : ""}`}
                      placeholder="Enter email subject..."
                    />
                    {subjectStatus !== "ok" && (
                      <p
                        className={`flex items-center gap-1 text-[10px] ${subjectStatus === "error" ? "text-red-600" : "text-amber-600"}`}
                      >
                        <AlertCircle className="h-3 w-3" />
                        {subjectStatus === "error" ? "Too long" : "May be cut off on mobile"}
                      </p>
                    )}
                  </div>

                  {/* Variables inline */}
                  {template.available_variables && template.available_variables.length > 0 && (
                    <div className="space-y-1.5">
                      <div className="flex items-center justify-between">
                        <Label className="text-xs font-medium">Variables</Label>
                        <span className="text-muted-foreground text-[10px]">
                          {template.available_variables.length} available
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {template.available_variables.map((variable) => (
                          <Tooltip key={variable.name}>
                            <TooltipTrigger asChild>
                              <button
                                type="button"
                                onClick={() => insertVariable(variable.name, activeTab === "text" ? "text" : "html")}
                                className="group bg-muted/50 hover:bg-primary hover:text-primary-foreground inline-flex items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[10px] transition-all"
                              >
                                {variable.name}
                                <Copy
                                  className="h-2.5 w-2.5 opacity-0 group-hover:opacity-100"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    copyVariable(variable.name);
                                  }}
                                />
                              </button>
                            </TooltipTrigger>
                            <TooltipContent
                              side="bottom"
                              className="text-xs"
                            >
                              {variable.description}
                            </TooltipContent>
                          </Tooltip>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Editor Tabs */}
            <Card className="overflow-hidden">
              <Tabs
                value={activeTab}
                onValueChange={(v) => setActiveTab(v as EditorTab)}
              >
                <div className="flex items-center justify-between border-b px-3 py-1.5">
                  <TabsList className="h-8 bg-transparent p-0">
                    <TabsTrigger
                      value="html"
                      className="data-[state=active]:bg-muted h-7 gap-1.5 rounded px-2 text-xs data-[state=active]:shadow-none"
                    >
                      <FileCode className="h-3 w-3" />
                      HTML
                    </TabsTrigger>
                    <TabsTrigger
                      value="text"
                      className="data-[state=active]:bg-muted h-7 gap-1.5 rounded px-2 text-xs data-[state=active]:shadow-none"
                    >
                      <FileText className="h-3 w-3" />
                      Text
                    </TabsTrigger>
                    {layoutMode === "sidebar" && (
                      <TabsTrigger
                        value="preview"
                        className="data-[state=active]:bg-muted h-7 gap-1.5 rounded px-2 text-xs data-[state=active]:shadow-none"
                      >
                        <Eye className="h-3 w-3" />
                        Variables
                      </TabsTrigger>
                    )}
                  </TabsList>
                  <span className="text-muted-foreground text-[10px]">
                    {activeTab === "html"
                      ? formData.body_html.length.toLocaleString()
                      : formData.body_text.length.toLocaleString()}{" "}
                    chars
                  </span>
                </div>

                <TabsContent
                  value="html"
                  className="m-0"
                >
                  <Textarea
                    ref={htmlTextareaRef}
                    value={formData.body_html}
                    onChange={(e) => setFormData((prev) => ({ ...prev, body_html: e.target.value }))}
                    rows={layoutMode === "split" ? 18 : 14}
                    className="resize-none rounded-none border-0 font-mono text-xs leading-relaxed focus-visible:ring-0 focus-visible:ring-offset-0"
                    spellCheck={false}
                  />
                </TabsContent>

                <TabsContent
                  value="text"
                  className="m-0"
                >
                  <Textarea
                    ref={textTextareaRef}
                    value={formData.body_text}
                    onChange={(e) => setFormData((prev) => ({ ...prev, body_text: e.target.value }))}
                    rows={layoutMode === "split" ? 18 : 14}
                    className="resize-none rounded-none border-0 font-mono text-xs focus-visible:ring-0 focus-visible:ring-offset-0"
                    spellCheck={false}
                    placeholder="Plain text version..."
                  />
                </TabsContent>

                <TabsContent
                  value="preview"
                  className="m-0 p-3"
                >
                  <div className="grid gap-2 sm:grid-cols-2">
                    {template.available_variables?.map((variable) => (
                      <div
                        key={variable.name}
                        className="space-y-1"
                      >
                        <Label
                          htmlFor={`var-${variable.name}`}
                          className="text-[10px] font-medium"
                        >
                          {variable.name}
                        </Label>
                        <Input
                          id={`var-${variable.name}`}
                          value={previewVars[variable.name] || ""}
                          onChange={(e) => setPreviewVars((prev) => ({ ...prev, [variable.name]: e.target.value }))}
                          placeholder={variable.description}
                          className="h-8 text-xs"
                        />
                      </div>
                    ))}
                  </div>
                </TabsContent>
              </Tabs>
            </Card>

            {/* Stats - Inline compact */}
            <div className="text-muted-foreground flex items-center gap-4 px-1 text-xs">
              <span className="flex items-center gap-1">
                <Rocket className="h-3 w-3" />
                {template.send_count.toLocaleString()} sent
              </span>
              {template.last_sent_at && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Last: {new Date(template.last_sent_at).toLocaleDateString()}
                </span>
              )}
            </div>
          </div>

          {/* Split View Preview */}
          {layoutMode === "split" && (
            <Card className="sticky top-4 h-fit overflow-hidden">
              <div className="flex items-center justify-between border-b px-3 py-2">
                <span className="text-sm font-medium">Preview</span>
                <div className="flex items-center gap-1.5">
                  <Tabs
                    value={previewTab}
                    onValueChange={(v) => setPreviewTab(v as "html" | "text")}
                  >
                    <TabsList className="h-7">
                      <TabsTrigger
                        value="html"
                        className="h-6 px-2 text-[10px]"
                      >
                        HTML
                      </TabsTrigger>
                      <TabsTrigger
                        value="text"
                        className="h-6 px-2 text-[10px]"
                      >
                        Text
                      </TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <div className="flex items-center rounded border p-0.5">
                    {(["desktop", "tablet", "mobile"] as DeviceSize[]).map((size) => {
                      const Icon = size === "desktop" ? Monitor : size === "tablet" ? Tablet : Smartphone;
                      return (
                        <Button
                          key={size}
                          variant={deviceSize === size ? "secondary" : "ghost"}
                          size="sm"
                          className="h-6 w-6 p-0"
                          onClick={() => setDeviceSize(size)}
                        >
                          <Icon className="h-3 w-3" />
                        </Button>
                      );
                    })}
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={handlePreview}
                    disabled={previewMutation.isPending}
                  >
                    {previewMutation.isPending ? (
                      <Loader2 className="h-3 w-3 animate-spin" />
                    ) : (
                      <RefreshCw className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              {previewData && (
                <div className="bg-muted/30 border-b px-3 py-1.5">
                  <p className="truncate text-xs font-medium">{previewData.subject}</p>
                </div>
              )}
              <div className="flex justify-center overflow-auto bg-gray-100 p-3 dark:bg-gray-900/50">
                {previewTab === "html" ? (
                  <div
                    className="overflow-hidden rounded-md border bg-white shadow-sm transition-all"
                    style={{ width: DEVICE_WIDTHS[deviceSize], maxWidth: "100%" }}
                  >
                    {previewData ? (
                      <iframe
                        srcDoc={previewData.body_html}
                        className="h-[400px] w-full"
                        title="Preview"
                        sandbox="allow-same-origin"
                      />
                    ) : (
                      <div className="flex h-[400px] items-center justify-center">
                        <Loader2 className="text-muted-foreground h-6 w-6 animate-spin" />
                      </div>
                    )}
                  </div>
                ) : (
                  <ScrollArea className="h-[400px] w-full">
                    <pre className="rounded-md bg-white p-3 font-mono text-xs whitespace-pre-wrap dark:bg-gray-800">
                      {previewData?.body_text || "No plain text version"}
                    </pre>
                  </ScrollArea>
                )}
              </div>
            </Card>
          )}
        </div>
      </div>

      {/* Preview Modal */}
      <Dialog
        open={showPreview}
        onOpenChange={setShowPreview}
      >
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base">
              <Eye className="h-4 w-4" />
              Preview
            </DialogTitle>
            <DialogDescription className="text-xs">{previewData?.subject}</DialogDescription>
          </DialogHeader>

          <div className="flex items-center justify-between border-b pb-3">
            <Tabs
              value={previewTab}
              onValueChange={(v) => setPreviewTab(v as "html" | "text")}
            >
              <TabsList className="h-8">
                <TabsTrigger
                  value="html"
                  className="gap-1.5 text-xs"
                >
                  <Code className="h-3 w-3" />
                  HTML
                </TabsTrigger>
                <TabsTrigger
                  value="text"
                  className="gap-1.5 text-xs"
                >
                  <FileText className="h-3 w-3" />
                  Text
                </TabsTrigger>
              </TabsList>
            </Tabs>

            {previewTab === "html" && (
              <div className="flex items-center gap-0.5 rounded-md border p-0.5">
                {(["desktop", "tablet", "mobile"] as DeviceSize[]).map((size) => {
                  const Icon = size === "desktop" ? Monitor : size === "tablet" ? Tablet : Smartphone;
                  return (
                    <Button
                      key={size}
                      variant={deviceSize === size ? "secondary" : "ghost"}
                      size="sm"
                      className="h-7 w-7 p-0"
                      onClick={() => setDeviceSize(size)}
                    >
                      <Icon className="h-3.5 w-3.5" />
                    </Button>
                  );
                })}
              </div>
            )}
          </div>

          {previewTab === "html" ? (
            <div className="flex justify-center overflow-auto rounded-lg bg-gray-100 p-3 dark:bg-gray-900">
              <div
                className="overflow-hidden rounded-lg border bg-white shadow-md transition-all"
                style={{ width: DEVICE_WIDTHS[deviceSize], maxWidth: "100%" }}
              >
                {previewData && (
                  <iframe
                    srcDoc={previewData.body_html}
                    className="h-[55vh] w-full"
                    title="Preview"
                    sandbox="allow-same-origin"
                  />
                )}
              </div>
            </div>
          ) : (
            <ScrollArea className="h-[55vh]">
              <pre className="rounded-lg bg-gray-100 p-3 font-mono text-xs whitespace-pre-wrap dark:bg-gray-900">
                {previewData?.body_text || "No plain text version"}
              </pre>
            </ScrollArea>
          )}
        </DialogContent>
      </Dialog>

      {/* Test Email Modal */}
      <Dialog
        open={showTestEmail}
        onOpenChange={setShowTestEmail}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base">
              <Send className="h-4 w-4" />
              Send Test Email
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div className="space-y-1.5">
              <Label
                htmlFor="test-email"
                className="text-xs"
              >
                Email Address
              </Label>
              <Input
                id="test-email"
                type="email"
                placeholder="you@example.com"
                value={testEmailAddress}
                onChange={(e) => setTestEmailAddress(e.target.value)}
                className="h-9"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowTestEmail(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleSendTestEmail}
                disabled={!testEmailAddress?.includes("@")}
                className="gap-1.5"
              >
                <Send className="h-3.5 w-3.5" />
                Send
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}

function getSampleValue(variableName: string): string {
  const samples: Record<string, string> = {
    name: "John Doe",
    Name: "John Doe",
    user_name: "John Doe",
    email: "john@example.com",
    Email: "john@example.com",
    app_name: "MyApp",
    AppName: "MyApp",
    reset_url: "https://example.com/reset?token=abc123",
    verification_url: "https://example.com/verify?token=xyz789",
    support_email: "support@example.com",
    date: new Date().toLocaleDateString(),
    code: "123456",
    otp: "123456",
  };

  return samples[variableName] || `[${variableName}]`;
}
