import { useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  AlertTriangle,
  Check,
  Copy,
  Eye,
  EyeOff,
  Key,
  Loader2,
  Lock,
  LogOut,
  Monitor,
  Shield,
  Smartphone,
  X,
} from "lucide-react";
import { toast } from "sonner";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { requireAuth } from "../../lib/guards";
import { AuthService } from "../../services/auth/authService";
import {
  SettingsService,
  type TwoFactorSetupResponse,
  type UserSession,
} from "../../services/settings/settingsService";

export const Route = createFileRoute("/settings/security")({
  beforeLoad: () => requireAuth(),
  component: SecuritySettingsPage,
});

function SecuritySettingsPage() {
  const queryClient = useQueryClient();

  const { data: user } = useQuery({
    queryKey: ["currentUser"],
    queryFn: () => AuthService.getCurrentUser(),
    staleTime: 60 * 1000,
  });

  const { data: sessions, isLoading: sessionsLoading } = useQuery({
    queryKey: ["user-sessions"],
    queryFn: () => SettingsService.getSessions(),
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Security</h2>
        <p className="text-sm text-gray-500">Manage your password, two-factor authentication, and active sessions</p>
      </div>

      {/* Password Change */}
      <PasswordChangeCard />

      {/* Two-Factor Authentication */}
      <TwoFactorCard />

      {/* Active Sessions */}
      <SessionsCard
        sessions={sessions || []}
        isLoading={sessionsLoading}
      />
    </div>
  );
}

function PasswordChangeCard() {
  const [showPasswords, setShowPasswords] = useState(false);
  const [formData, setFormData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  const changePasswordMutation = useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) => SettingsService.changePassword(data),
    onSuccess: () => {
      toast.success("Your password has been changed successfully.");
      setFormData({ currentPassword: "", newPassword: "", confirmPassword: "" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (formData.newPassword !== formData.confirmPassword) {
      toast.error("New passwords do not match.");
      return;
    }

    if (formData.newPassword.length < 8) {
      toast.error("Password must be at least 8 characters long.");
      return;
    }

    changePasswordMutation.mutate({
      current_password: formData.currentPassword,
      new_password: formData.newPassword,
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="h-5 w-5" />
          Change Password
        </CardTitle>
        <CardDescription>Update your password to keep your account secure</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-4"
        >
          <div className="space-y-2">
            <Label htmlFor="currentPassword">Current Password</Label>
            <div className="relative">
              <Input
                id="currentPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.currentPassword}
                onChange={(e) => setFormData({ ...formData, currentPassword: e.target.value })}
                placeholder="Enter your current password"
              />
              <button
                type="button"
                className="absolute top-1/2 right-3 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                onClick={() => setShowPasswords(!showPasswords)}
              >
                {showPasswords ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="newPassword">New Password</Label>
              <Input
                id="newPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.newPassword}
                onChange={(e) => setFormData({ ...formData, newPassword: e.target.value })}
                placeholder="Enter new password"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">Confirm New Password</Label>
              <Input
                id="confirmPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.confirmPassword}
                onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                placeholder="Confirm new password"
              />
            </div>
          </div>

          {formData.newPassword && formData.newPassword !== formData.confirmPassword && (
            <p className="text-sm text-red-600">Passwords do not match</p>
          )}

          <Button
            type="submit"
            disabled={
              !formData.currentPassword ||
              !formData.newPassword ||
              formData.newPassword !== formData.confirmPassword ||
              changePasswordMutation.isPending
            }
          >
            {changePasswordMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Changing...
              </>
            ) : (
              "Change Password"
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function TwoFactorCard() {
  const queryClient = useQueryClient();
  const [setupData, setSetupData] = useState<TwoFactorSetupResponse | null>(null);
  const [verifyCode, setVerifyCode] = useState("");
  const [disableCode, setDisableCode] = useState("");
  const [showDisable, setShowDisable] = useState(false);
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [showBackupCodes, setShowBackupCodes] = useState(false);

  const { data: user } = useQuery({
    queryKey: ["currentUser"],
    queryFn: () => AuthService.getCurrentUser(),
    staleTime: 60 * 1000,
  });

  // Check if 2FA is enabled (you'll need to add this to your User type)
  const is2FAEnabled = false; // TODO: Get from user.two_factor_enabled

  const setup2FAMutation = useMutation({
    mutationFn: () => SettingsService.setup2FA(),
    onSuccess: (data) => {
      setSetupData(data);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const verify2FAMutation = useMutation({
    mutationFn: (code: string) => SettingsService.verify2FA(code),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      setBackupCodes(data.backup_codes);
      setShowBackupCodes(true);
      setSetupData(null);
      setVerifyCode("");
      toast.success("Two-factor authentication has been enabled.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const disable2FAMutation = useMutation({
    mutationFn: (code: string) => SettingsService.disable2FA(code),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      setShowDisable(false);
      setDisableCode("");
      toast.success("Two-factor authentication has been disabled.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const copyBackupCodes = () => {
    navigator.clipboard.writeText(backupCodes.join("\n"));
    toast.success("Backup codes copied to clipboard.");
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Two-Factor Authentication
        </CardTitle>
        <CardDescription>Add an extra layer of security to your account</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!is2FAEnabled && !setupData && (
          <div className="flex items-center justify-between rounded-lg border border-amber-200 bg-amber-50 p-4">
            <div className="flex items-center gap-3">
              <AlertTriangle className="h-5 w-5 text-amber-600" />
              <div>
                <p className="font-medium text-amber-900">2FA Not Enabled</p>
                <p className="text-sm text-amber-700">Protect your account with two-factor authentication</p>
              </div>
            </div>
            <Button
              onClick={() => setup2FAMutation.mutate()}
              disabled={setup2FAMutation.isPending}
            >
              {setup2FAMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Enable 2FA"}
            </Button>
          </div>
        )}

        {setupData && (
          <div className="space-y-4 rounded-lg border p-4">
            <h4 className="font-medium">Set up your authenticator app</h4>
            <ol className="list-inside list-decimal space-y-2 text-sm text-gray-600">
              <li>Install an authenticator app (Google Authenticator, Authy, etc.)</li>
              <li>Scan the QR code below or enter the secret manually</li>
              <li>Enter the 6-digit code from your app to verify</li>
            </ol>

            <div className="flex flex-col items-center gap-4 py-4">
              {/* QR Code */}
              <div className="rounded-lg border bg-white p-4">
                <img
                  src={`data:image/png;base64,${setupData.qr_code}`}
                  alt="QR Code for 2FA"
                  className="h-48 w-48"
                />
              </div>

              {/* Manual Entry Secret */}
              <div className="text-center">
                <p className="text-xs text-gray-500">Or enter this code manually:</p>
                <code className="mt-1 block rounded bg-gray-100 px-3 py-1 font-mono text-sm">{setupData.secret}</code>
              </div>
            </div>

            {/* Verification Input */}
            <div className="space-y-2">
              <Label htmlFor="verifyCode">Verification Code</Label>
              <div className="flex gap-2">
                <Input
                  id="verifyCode"
                  value={verifyCode}
                  onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  placeholder="Enter 6-digit code"
                  maxLength={6}
                  className="font-mono"
                />
                <Button
                  onClick={() => verify2FAMutation.mutate(verifyCode)}
                  disabled={verifyCode.length !== 6 || verify2FAMutation.isPending}
                >
                  {verify2FAMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Verify"}
                </Button>
              </div>
            </div>

            <Button
              variant="ghost"
              onClick={() => setSetupData(null)}
              className="w-full"
            >
              Cancel Setup
            </Button>
          </div>
        )}

        {is2FAEnabled && !showDisable && (
          <div className="flex items-center justify-between rounded-lg border border-green-200 bg-green-50 p-4">
            <div className="flex items-center gap-3">
              <Check className="h-5 w-5 text-green-600" />
              <div>
                <p className="font-medium text-green-900">2FA Enabled</p>
                <p className="text-sm text-green-700">Your account is protected with two-factor authentication</p>
              </div>
            </div>
            <Button
              variant="outline"
              onClick={() => setShowDisable(true)}
            >
              Disable 2FA
            </Button>
          </div>
        )}

        {showDisable && (
          <div className="space-y-4 rounded-lg border border-red-200 bg-red-50 p-4">
            <div className="flex items-center gap-2 text-red-700">
              <AlertTriangle className="h-5 w-5" />
              <p className="font-medium">Disable Two-Factor Authentication?</p>
            </div>
            <p className="text-sm text-red-600">
              This will make your account less secure. Enter your authentication code to confirm.
            </p>
            <div className="flex gap-2">
              <Input
                value={disableCode}
                onChange={(e) => setDisableCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                placeholder="Enter 6-digit code"
                maxLength={6}
                className="font-mono"
              />
              <Button
                variant="destructive"
                onClick={() => disable2FAMutation.mutate(disableCode)}
                disabled={disableCode.length !== 6 || disable2FAMutation.isPending}
              >
                {disable2FAMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Disable"}
              </Button>
              <Button
                variant="ghost"
                onClick={() => setShowDisable(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        )}

        {/* Backup Codes Modal */}
        {showBackupCodes && backupCodes.length > 0 && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="w-full max-w-md rounded-lg bg-white p-6">
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-lg font-semibold">Backup Codes</h3>
                <button onClick={() => setShowBackupCodes(false)}>
                  <X className="h-5 w-5 text-gray-400" />
                </button>
              </div>
              <p className="mb-4 text-sm text-gray-600">
                Save these backup codes in a secure location. Each code can only be used once.
              </p>
              <div className="grid grid-cols-2 gap-2 rounded-lg bg-gray-50 p-4 font-mono text-sm">
                {backupCodes.map((code, i) => (
                  <div
                    key={i}
                    className="rounded bg-white px-2 py-1 text-center"
                  >
                    {code}
                  </div>
                ))}
              </div>
              <div className="mt-4 flex gap-2">
                <Button
                  onClick={copyBackupCodes}
                  variant="outline"
                  className="flex-1"
                >
                  <Copy className="mr-2 h-4 w-4" />
                  Copy Codes
                </Button>
                <Button
                  onClick={() => setShowBackupCodes(false)}
                  className="flex-1"
                >
                  Done
                </Button>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function SessionsCard({ sessions, isLoading }: { sessions: UserSession[]; isLoading: boolean }) {
  const queryClient = useQueryClient();

  const revokeSessionMutation = useMutation({
    mutationFn: (sessionId: number) => SettingsService.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-sessions"] });
      toast.success("Session has been revoked.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const revokeAllMutation = useMutation({
    mutationFn: () => SettingsService.revokeAllSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-sessions"] });
      toast.success("All other sessions have been revoked.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const getDeviceIcon = (deviceInfo: string) => {
    const lower = deviceInfo.toLowerCase();
    if (lower.includes("mobile") || lower.includes("iphone") || lower.includes("android")) {
      return <Smartphone className="h-5 w-5" />;
    }
    return <Monitor className="h-5 w-5" />;
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              Active Sessions
            </CardTitle>
            <CardDescription>Manage your active sessions and sign out from other devices</CardDescription>
          </div>
          {sessions.length > 1 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => revokeAllMutation.mutate()}
              disabled={revokeAllMutation.isPending}
            >
              <LogOut className="mr-2 h-4 w-4" />
              Sign Out All Others
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-20 animate-pulse rounded-lg bg-gray-100"
              />
            ))}
          </div>
        ) : sessions.length === 0 ? (
          <p className="text-center text-gray-500">No active sessions found</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                className={`flex items-center justify-between rounded-lg border p-4 ${
                  session.is_current ? "border-blue-200 bg-blue-50" : ""
                }`}
              >
                <div className="flex items-center gap-4">
                  <div className="text-gray-400">{getDeviceIcon(session.device_info)}</div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{session.device_info || "Unknown Device"}</span>
                      {session.is_current && (
                        <Badge
                          variant="secondary"
                          className="text-xs"
                        >
                          Current
                        </Badge>
                      )}
                    </div>
                    <div className="text-sm text-gray-500">
                      {session.location || "Unknown Location"} • {session.ip_address}
                    </div>
                    <div className="text-xs text-gray-400">
                      Last active: {new Date(session.last_active_at).toLocaleString()}
                    </div>
                  </div>
                </div>
                {!session.is_current && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => revokeSessionMutation.mutate(session.id)}
                    disabled={revokeSessionMutation.isPending}
                    className="text-red-600 hover:bg-red-50 hover:text-red-700"
                  >
                    <LogOut className="h-4 w-4" />
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
