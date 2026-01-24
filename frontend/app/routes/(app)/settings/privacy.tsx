import { useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useRequestDataExport } from "@/hooks/mutations";
import { useDataExportStatus } from "@/hooks/queries";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { queryKeys } from "@/lib/query-keys";
import { SettingsService, type DataExportStatus } from "@/services/settings/settingsService";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  AlertTriangle,
  Download,
  Eye,
  EyeOff,
  FileArchive,
  Loader2,
  RefreshCw,
  ShieldAlert,
  Trash2,
  X,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

/**
 * Format bytes to human-readable string
 */
function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = parseFloat((bytes / Math.pow(k, i)).toFixed(1));
  return value + " " + sizes[i];
}

export const Route = createFileRoute("/(app)/settings/privacy")({
  component: PrivacySettingsPage,
});

function PrivacySettingsPage() {
  const { t } = useTranslation("settings");

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold">{t("privacy.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("privacy.subtitle")}</p>
        </div>

        {/* Data Export */}
        <DataExportCard />

        {/* Account Deletion */}
        <AccountDeletionCard />
      </div>
    </SettingsLayout>
  );
}

function DataExportCard() {
  const { t } = useTranslation("settings");

  // Use the new dedicated hooks
  const { data: exportStatus, isLoading } = useDataExportStatus();
  const requestExportMutation = useRequestDataExport();

  const getStatusBadge = (status: DataExportStatus["status"]) => {
    switch (status) {
      case "pending":
        return <Badge variant="secondary">{t("privacy.export.status.pending")}</Badge>;
      case "processing":
        return <Badge variant="info">{t("privacy.export.status.processing")}</Badge>;
      case "completed":
        return <Badge variant="success">{t("privacy.export.status.ready")}</Badge>;
      case "failed":
        return <Badge variant="destructive">{t("privacy.export.status.failed")}</Badge>;
      case "expired":
        return <Badge variant="outline">{t("privacy.export.status.expired", { defaultValue: "Expired" })}</Badge>;
      default:
        return null;
    }
  };

  // Check if user can request a new export (no active export or previous is done)
  const canRequestNew =
    !exportStatus ||
    exportStatus.status === "completed" ||
    exportStatus.status === "failed" ||
    exportStatus.status === "expired";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileArchive className="h-5 w-5" />
          {t("privacy.export.title")}
        </CardTitle>
        <CardDescription>{t("privacy.export.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="bg-muted rounded-lg p-4">
          <h4 className="font-medium">{t("privacy.export.included")}</h4>
          <ul className="text-muted-foreground mt-2 list-inside list-disc space-y-1 text-sm">
            <li>{t("privacy.export.items.profile")}</li>
            <li>{t("privacy.export.items.loginHistory")}</li>
            <li>{t("privacy.export.items.files")}</li>
            <li>{t("privacy.export.items.notifications")}</li>
            <li>{t("privacy.export.items.other")}</li>
          </ul>
        </div>

        {isLoading ? (
          <div className="text-muted-foreground flex items-center gap-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>{t("privacy.export.checking")}</span>
          </div>
        ) : exportStatus ? (
          <div className="space-y-3">
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div className="flex items-center gap-4">
                <Download className="text-muted-foreground h-8 w-8" />
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{t("privacy.export.dataExport")}</p>
                    {getStatusBadge(exportStatus.status)}
                  </div>
                  <p className="text-muted-foreground text-sm">
                    {t("privacy.export.requestedOn")} {new Date(exportStatus.created_at).toLocaleDateString()}
                  </p>
                  {/* File size display */}
                  {exportStatus.file_size && exportStatus.status === "completed" && (
                    <p className="text-muted-foreground text-xs">
                      {t("privacy.export.fileSize", { defaultValue: "Size" })}: {formatBytes(exportStatus.file_size)}
                    </p>
                  )}
                  {/* Expiration date */}
                  {exportStatus.expires_at && exportStatus.status === "completed" && (
                    <p className="text-warning text-xs">
                      {t("privacy.export.expiresOn")} {new Date(exportStatus.expires_at).toLocaleDateString()}
                    </p>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                {/* Download button */}
                {exportStatus.status === "completed" && exportStatus.download_url && (
                  <Button asChild>
                    <a
                      href={exportStatus.download_url}
                      download
                    >
                      <Download className="mr-2 h-4 w-4" />
                      {t("privacy.export.download")}
                    </a>
                  </Button>
                )}
                {/* Loading spinner while processing */}
                {(exportStatus.status === "pending" || exportStatus.status === "processing") && (
                  <Loader2 className="text-primary h-5 w-5 animate-spin" />
                )}
              </div>
            </div>

            {/* Error message display for failed exports */}
            {exportStatus.status === "failed" && (
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  {exportStatus.error_message ||
                    t("privacy.export.defaultError", { defaultValue: "Export failed. Please try again." })}
                </AlertDescription>
              </Alert>
            )}

            {/* Request new export button (shown when previous export is done/failed/expired) */}
            {canRequestNew && (
              <Button
                variant="outline"
                onClick={() => requestExportMutation.mutate()}
                disabled={requestExportMutation.isPending}
              >
                {requestExportMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {t("privacy.export.requesting")}
                  </>
                ) : (
                  <>
                    <RefreshCw className="mr-2 h-4 w-4" />
                    {t("privacy.export.requestNew", { defaultValue: "Request New Export" })}
                  </>
                )}
              </Button>
            )}
          </div>
        ) : (
          <Button
            onClick={() => requestExportMutation.mutate()}
            disabled={requestExportMutation.isPending}
          >
            {requestExportMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {t("privacy.export.requesting")}
              </>
            ) : (
              <>
                <Download className="mr-2 h-4 w-4" />
                {t("privacy.export.request")}
              </>
            )}
          </Button>
        )}

        <p className="text-muted-foreground text-xs">{t("privacy.export.hint")}</p>
      </CardContent>
    </Card>
  );
}

function AccountDeletionCard() {
  const { t } = useTranslation("settings");
  const { t: tCommon } = useTranslation("common");
  const queryClient = useQueryClient();
  const [showConfirm, setShowConfirm] = useState(false);
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [reason, setReason] = useState("");
  const [confirmText, setConfirmText] = useState("");

  const requestDeletionMutation = useMutation({
    mutationFn: ({ password, reason }: { password: string; reason?: string }) =>
      SettingsService.requestAccountDeletion(password, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });
      toast.success(t("privacy.delete.toast.scheduled"));
      setShowConfirm(false);
      setPassword("");
      setReason("");
      setConfirmText("");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (confirmText !== "DELETE") {
      toast.error(t("privacy.delete.toast.typeDelete"));
      return;
    }
    requestDeletionMutation.mutate({ password, reason: reason || undefined });
  };

  return (
    <Card className="border-destructive/30">
      <CardHeader>
        <CardTitle className="text-destructive flex items-center gap-2">
          <ShieldAlert className="h-5 w-5" />
          {t("privacy.delete.title")}
        </CardTitle>
        <CardDescription>{t("privacy.delete.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!showConfirm ? (
          <>
            <div className="bg-destructive/10 border-destructive/30 rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="text-destructive mt-0.5 h-5 w-5" />
                <div>
                  <p className="font-medium">{t("privacy.delete.warning")}</p>
                  <p className="text-muted-foreground mt-1 text-sm">{t("privacy.delete.warningText")}</p>
                  <ul className="text-muted-foreground mt-2 list-inside list-disc text-sm">
                    <li>{t("privacy.delete.items.profile")}</li>
                    <li>{t("privacy.delete.items.files")}</li>
                    <li>{t("privacy.delete.items.loginHistory")}</li>
                    <li>{t("privacy.delete.items.billing")}</li>
                  </ul>
                </div>
              </div>
            </div>

            <Button
              variant="destructive"
              onClick={() => setShowConfirm(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {t("privacy.delete.button")}
            </Button>
          </>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="space-y-4"
          >
            {/* Password */}
            <div className="space-y-2">
              <Label htmlFor="delete-password">{t("privacy.delete.confirmPassword")}</Label>
              <div className="relative">
                <Input
                  id="delete-password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={t("privacy.delete.passwordPlaceholder")}
                  required
                />
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {/* Reason (optional) */}
            <div className="space-y-2">
              <Label htmlFor="delete-reason">{t("privacy.delete.reason")}</Label>
              <Textarea
                id="delete-reason"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder={t("privacy.delete.reasonPlaceholder")}
                rows={3}
              />
            </div>

            {/* Confirmation Text */}
            <div className="space-y-2">
              <Label htmlFor="confirm-text">
                {
                  t("privacy.delete.confirmType", {
                    defaultValue: "Type <1>DELETE</1> to confirm",
                  }).split("<1>")[0]
                }
                <span className="text-destructive font-mono font-bold">DELETE</span>
                {
                  t("privacy.delete.confirmType", {
                    defaultValue: "Type <1>DELETE</1> to confirm",
                  }).split("</1>")[1]
                }
              </Label>
              <Input
                id="confirm-text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={t("privacy.delete.confirmPlaceholder")}
                className="font-mono"
              />
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setShowConfirm(false);
                  setPassword("");
                  setReason("");
                  setConfirmText("");
                }}
              >
                <X className="mr-2 h-4 w-4" />
                {tCommon("buttons.cancel")}
              </Button>
              <Button
                type="submit"
                variant="destructive"
                disabled={!password || confirmText !== "DELETE" || requestDeletionMutation.isPending}
              >
                {requestDeletionMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {t("privacy.delete.processing")}
                  </>
                ) : (
                  <>
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t("privacy.delete.confirmButton")}
                  </>
                )}
              </Button>
            </div>

            <p className="text-muted-foreground text-xs">{t("privacy.delete.gracePeriod")}</p>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
