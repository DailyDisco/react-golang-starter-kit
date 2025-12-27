import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, Download, Eye, EyeOff, FileArchive, Loader2, ShieldAlert, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Textarea } from "../../components/ui/textarea";
import { requireAuth } from "../../lib/guards";
import { SettingsService, type DataExportStatus } from "../../services/settings/settingsService";

export const Route = createFileRoute("/settings/privacy")({
  beforeLoad: () => requireAuth(),
  component: PrivacySettingsPage,
});

function PrivacySettingsPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Privacy</h2>
        <p className="text-sm text-gray-500">Manage your data and account privacy settings</p>
      </div>

      {/* Data Export */}
      <DataExportCard />

      {/* Account Deletion */}
      <AccountDeletionCard />
    </div>
  );
}

function DataExportCard() {
  const queryClient = useQueryClient();

  const { data: exportStatus, isLoading } = useQuery({
    queryKey: ["data-export-status"],
    queryFn: () => SettingsService.getDataExportStatus(),
    refetchInterval: (query) => {
      // Poll every 5 seconds if export is pending/processing
      const data = query.state.data;
      if (data && (data.status === "pending" || data.status === "processing")) {
        return 5000;
      }
      return false;
    },
  });

  const requestExportMutation = useMutation({
    mutationFn: () => SettingsService.requestDataExport(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["data-export-status"] });
      toast.success("Data export has been requested. We'll notify you when it's ready.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const getStatusBadge = (status: DataExportStatus["status"]) => {
    switch (status) {
      case "pending":
        return <Badge variant="secondary">Pending</Badge>;
      case "processing":
        return (
          <Badge
            variant="secondary"
            className="bg-blue-100 text-blue-700"
          >
            Processing
          </Badge>
        );
      case "completed":
        return (
          <Badge
            variant="default"
            className="bg-green-500"
          >
            Ready
          </Badge>
        );
      case "failed":
        return <Badge variant="destructive">Failed</Badge>;
      default:
        return null;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileArchive className="h-5 w-5" />
          Export Your Data
        </CardTitle>
        <CardDescription>Download a copy of all your personal data stored in our system</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="rounded-lg bg-gray-50 p-4">
          <h4 className="font-medium">What's included in your export:</h4>
          <ul className="mt-2 list-inside list-disc space-y-1 text-sm text-gray-600">
            <li>Profile information and settings</li>
            <li>Login history and activity logs</li>
            <li>Files and documents you've uploaded</li>
            <li>Notification preferences</li>
            <li>Any other data associated with your account</li>
          </ul>
        </div>

        {isLoading ? (
          <div className="flex items-center gap-2 text-gray-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Checking export status...</span>
          </div>
        ) : exportStatus ? (
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="flex items-center gap-4">
              <Download className="h-8 w-8 text-gray-400" />
              <div>
                <div className="flex items-center gap-2">
                  <p className="font-medium">Data Export</p>
                  {getStatusBadge(exportStatus.status)}
                </div>
                <p className="text-sm text-gray-500">
                  Requested on {new Date(exportStatus.created_at).toLocaleDateString()}
                </p>
                {exportStatus.expires_at && exportStatus.status === "completed" && (
                  <p className="text-xs text-amber-600">
                    Expires on {new Date(exportStatus.expires_at).toLocaleDateString()}
                  </p>
                )}
              </div>
            </div>
            {exportStatus.status === "completed" && exportStatus.download_url && (
              <Button asChild>
                <a
                  href={exportStatus.download_url}
                  download
                >
                  <Download className="mr-2 h-4 w-4" />
                  Download
                </a>
              </Button>
            )}
            {(exportStatus.status === "pending" || exportStatus.status === "processing") && (
              <Loader2 className="h-5 w-5 animate-spin text-blue-500" />
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
                Requesting...
              </>
            ) : (
              <>
                <Download className="mr-2 h-4 w-4" />
                Request Data Export
              </>
            )}
          </Button>
        )}

        <p className="text-xs text-gray-500">
          Your data will be prepared as a downloadable archive. This process may take a few minutes. You'll receive an
          email notification when your data is ready.
        </p>
      </CardContent>
    </Card>
  );
}

function AccountDeletionCard() {
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
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      toast.success("Account deletion has been scheduled. You can cancel this within the grace period.");
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
      toast.error("Please type DELETE to confirm.");
      return;
    }
    requestDeletionMutation.mutate({ password, reason: reason || undefined });
  };

  return (
    <Card className="border-red-200">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-red-700">
          <ShieldAlert className="h-5 w-5" />
          Delete Account
        </CardTitle>
        <CardDescription>Permanently delete your account and all associated data</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!showConfirm ? (
          <>
            <div className="rounded-lg border border-red-200 bg-red-50 p-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="mt-0.5 h-5 w-5 text-red-600" />
                <div>
                  <p className="font-medium text-red-900">Warning: This action is permanent</p>
                  <p className="mt-1 text-sm text-red-700">
                    Deleting your account will permanently remove all your data, including:
                  </p>
                  <ul className="mt-2 list-inside list-disc text-sm text-red-700">
                    <li>Your profile and settings</li>
                    <li>All files and documents</li>
                    <li>Login history and activity logs</li>
                    <li>Any subscriptions or billing information</li>
                  </ul>
                </div>
              </div>
            </div>

            <Button
              variant="destructive"
              onClick={() => setShowConfirm(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete My Account
            </Button>
          </>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="space-y-4"
          >
            {/* Password */}
            <div className="space-y-2">
              <Label htmlFor="delete-password">Enter your password to confirm</Label>
              <div className="relative">
                <Input
                  id="delete-password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  required
                />
                <button
                  type="button"
                  className="absolute top-1/2 right-3 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {/* Reason (optional) */}
            <div className="space-y-2">
              <Label htmlFor="delete-reason">Why are you leaving? (optional)</Label>
              <Textarea
                id="delete-reason"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Help us improve by sharing your feedback..."
                rows={3}
              />
            </div>

            {/* Confirmation Text */}
            <div className="space-y-2">
              <Label htmlFor="confirm-text">
                Type <span className="font-mono font-bold text-red-600">DELETE</span> to confirm
              </Label>
              <Input
                id="confirm-text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="Type DELETE"
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
                Cancel
              </Button>
              <Button
                type="submit"
                variant="destructive"
                disabled={!password || confirmText !== "DELETE" || requestDeletionMutation.isPending}
              >
                {requestDeletionMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Permanently Delete Account
                  </>
                )}
              </Button>
            </div>

            <p className="text-xs text-gray-500">
              After requesting deletion, you'll have a 14-day grace period to cancel. After that, your account and all
              data will be permanently deleted.
            </p>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
