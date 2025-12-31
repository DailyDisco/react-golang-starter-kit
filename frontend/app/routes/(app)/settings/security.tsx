import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { AuthService } from "@/services/auth/authService";
import { SettingsService, type TwoFactorSetupResponse, type UserSession } from "@/services/settings/settingsService";
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
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/security")({
  component: SecuritySettingsPage,
});

function SecuritySettingsPage() {
  const { t } = useTranslation("settings");
  const { data: sessions, isLoading: sessionsLoading } = useQuery({
    queryKey: ["user-sessions"],
    queryFn: () => SettingsService.getSessions(),
  });

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold">{t("security.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("security.subtitle")}</p>
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
    </SettingsLayout>
  );
}

function PasswordChangeCard() {
  const { t } = useTranslation("settings");
  const [showPasswords, setShowPasswords] = useState(false);
  const [formData, setFormData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  const changePasswordMutation = useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) => SettingsService.changePassword(data),
    onSuccess: () => {
      toast.success(t("security.password.toast.changed"));
      setFormData({ currentPassword: "", newPassword: "", confirmPassword: "" });
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (formData.newPassword !== formData.confirmPassword) {
      toast.error(t("security.password.toast.mismatch"));
      return;
    }

    if (formData.newPassword.length < 8) {
      toast.error(t("security.password.minLength"));
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
          {t("security.password.title")}
        </CardTitle>
        <CardDescription>{t("security.password.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="space-y-4"
        >
          <div className="space-y-2">
            <Label htmlFor="currentPassword">{t("security.password.current")}</Label>
            <div className="relative">
              <Input
                id="currentPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.currentPassword}
                onChange={(e) => setFormData({ ...formData, currentPassword: e.target.value })}
                placeholder={t("security.password.currentPlaceholder")}
              />
              <button
                type="button"
                className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
                onClick={() => setShowPasswords(!showPasswords)}
              >
                {showPasswords ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="newPassword">{t("security.password.new")}</Label>
              <Input
                id="newPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.newPassword}
                onChange={(e) => setFormData({ ...formData, newPassword: e.target.value })}
                placeholder={t("security.password.newPlaceholder")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{t("security.password.confirm")}</Label>
              <Input
                id="confirmPassword"
                type={showPasswords ? "text" : "password"}
                value={formData.confirmPassword}
                onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                placeholder={t("security.password.confirmPlaceholder")}
              />
            </div>
          </div>

          {formData.newPassword && formData.newPassword !== formData.confirmPassword && (
            <p className="text-destructive text-sm">{t("security.password.mismatch")}</p>
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
                {t("security.password.changing")}
              </>
            ) : (
              t("security.password.change")
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function TwoFactorCard() {
  const { t } = useTranslation("settings");
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
      toast.success(t("security.twoFactor.toast.enabled"));
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
      toast.success(t("security.twoFactor.toast.disabled"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const copyBackupCodes = () => {
    navigator.clipboard.writeText(backupCodes.join("\n"));
    toast.success(t("security.twoFactor.toast.codesCopied"));
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          {t("security.twoFactor.title")}
        </CardTitle>
        <CardDescription>{t("security.twoFactor.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!is2FAEnabled && !setupData && (
          <div className="bg-warning/10 border-warning/30 flex items-center justify-between rounded-lg border p-4">
            <div className="flex items-center gap-3">
              <AlertTriangle className="text-warning h-5 w-5" />
              <div>
                <p className="font-medium">{t("security.twoFactor.notEnabled")}</p>
                <p className="text-muted-foreground text-sm">{t("security.twoFactor.notEnabledHint")}</p>
              </div>
            </div>
            <Button
              onClick={() => setup2FAMutation.mutate()}
              disabled={setup2FAMutation.isPending}
            >
              {setup2FAMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                t("security.twoFactor.enable")
              )}
            </Button>
          </div>
        )}

        {setupData && (
          <div className="space-y-4 rounded-lg border p-4">
            <h4 className="font-medium">{t("security.twoFactor.setupTitle")}</h4>
            <ol className="text-muted-foreground list-inside list-decimal space-y-2 text-sm">
              <li>{t("security.twoFactor.setupSteps.step1")}</li>
              <li>{t("security.twoFactor.setupSteps.step2")}</li>
              <li>{t("security.twoFactor.setupSteps.step3")}</li>
            </ol>

            <div className="flex flex-col items-center gap-4 py-4">
              <div className="rounded-lg border bg-white p-4">
                <img
                  src={`data:image/png;base64,${setupData.qr_code}`}
                  alt="QR Code for 2FA"
                  className="h-48 w-48"
                />
              </div>

              <div className="text-center">
                <p className="text-muted-foreground text-xs">{t("security.twoFactor.manualEntry")}</p>
                <code className="bg-muted mt-1 block rounded px-3 py-1 font-mono text-sm">{setupData.secret}</code>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="verifyCode">{t("security.twoFactor.verificationCode")}</Label>
              <div className="flex gap-2">
                <Input
                  id="verifyCode"
                  value={verifyCode}
                  onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  placeholder={t("security.twoFactor.verificationPlaceholder")}
                  maxLength={6}
                  className="font-mono"
                />
                <Button
                  onClick={() => verify2FAMutation.mutate(verifyCode)}
                  disabled={verifyCode.length !== 6 || verify2FAMutation.isPending}
                >
                  {verify2FAMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    t("security.twoFactor.verify")
                  )}
                </Button>
              </div>
            </div>

            <Button
              variant="ghost"
              onClick={() => setSetupData(null)}
              className="w-full"
            >
              {t("security.twoFactor.cancelSetup")}
            </Button>
          </div>
        )}

        {is2FAEnabled && !showDisable && (
          <div className="bg-success/10 border-success/30 flex items-center justify-between rounded-lg border p-4">
            <div className="flex items-center gap-3">
              <Check className="text-success h-5 w-5" />
              <div>
                <p className="font-medium">{t("security.twoFactor.enabled")}</p>
                <p className="text-muted-foreground text-sm">{t("security.twoFactor.enabledHint")}</p>
              </div>
            </div>
            <Button
              variant="outline"
              onClick={() => setShowDisable(true)}
            >
              {t("security.twoFactor.disable")}
            </Button>
          </div>
        )}

        {showDisable && (
          <div className="bg-destructive/10 border-destructive/30 space-y-4 rounded-lg border p-4">
            <div className="text-destructive flex items-center gap-2">
              <AlertTriangle className="h-5 w-5" />
              <p className="font-medium">{t("security.twoFactor.disableConfirm")}</p>
            </div>
            <p className="text-destructive/80 text-sm">{t("security.twoFactor.disableWarning")}</p>
            <div className="flex gap-2">
              <Input
                value={disableCode}
                onChange={(e) => setDisableCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                placeholder={t("security.twoFactor.verificationPlaceholder")}
                maxLength={6}
                className="font-mono"
              />
              <Button
                variant="destructive"
                onClick={() => disable2FAMutation.mutate(disableCode)}
                disabled={disableCode.length !== 6 || disable2FAMutation.isPending}
              >
                {disable2FAMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  t("security.twoFactor.disable")
                )}
              </Button>
              <Button
                variant="ghost"
                onClick={() => setShowDisable(false)}
              >
                {t("security.twoFactor.cancel")}
              </Button>
            </div>
          </div>
        )}

        {/* Backup Codes Modal */}
        {showBackupCodes && backupCodes.length > 0 && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="bg-background w-full max-w-md rounded-lg p-6">
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-lg font-semibold">{t("security.twoFactor.backupCodes.title")}</h3>
                <button onClick={() => setShowBackupCodes(false)}>
                  <X className="text-muted-foreground h-5 w-5" />
                </button>
              </div>
              <p className="text-muted-foreground mb-4 text-sm">{t("security.twoFactor.backupCodes.hint")}</p>
              <div className="bg-muted grid grid-cols-2 gap-2 rounded-lg p-4 font-mono text-sm">
                {backupCodes.map((code, i) => (
                  <div
                    key={i}
                    className="bg-background rounded px-2 py-1 text-center"
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
                  {t("security.twoFactor.backupCodes.copy")}
                </Button>
                <Button
                  onClick={() => setShowBackupCodes(false)}
                  className="flex-1"
                >
                  {t("security.twoFactor.backupCodes.done")}
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
  const { t } = useTranslation("settings");
  const queryClient = useQueryClient();

  const revokeSessionMutation = useMutation({
    mutationFn: (sessionId: number) => SettingsService.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-sessions"] });
      toast.success(t("security.sessions.toast.revoked"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const revokeAllMutation = useMutation({
    mutationFn: () => SettingsService.revokeAllSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-sessions"] });
      toast.success(t("security.sessions.toast.allRevoked"));
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
              {t("security.sessions.title")}
            </CardTitle>
            <CardDescription>{t("security.sessions.subtitle")}</CardDescription>
          </div>
          {sessions.length > 1 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => revokeAllMutation.mutate()}
              disabled={revokeAllMutation.isPending}
            >
              <LogOut className="mr-2 h-4 w-4" />
              {t("security.sessions.signOutAll")}
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
                className="bg-muted h-20 animate-pulse rounded-lg"
              />
            ))}
          </div>
        ) : sessions.length === 0 ? (
          <p className="text-muted-foreground text-center">{t("security.sessions.noSessions")}</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                className={`flex items-center justify-between rounded-lg border p-4 ${
                  session.is_current ? "border-primary/30 bg-primary/5" : ""
                }`}
              >
                <div className="flex items-center gap-4">
                  <div className="text-muted-foreground">{getDeviceIcon(session.device_info)}</div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{session.device_info || t("security.sessions.unknownDevice")}</span>
                      {session.is_current && (
                        <Badge
                          variant="secondary"
                          className="text-xs"
                        >
                          {t("security.sessions.current")}
                        </Badge>
                      )}
                    </div>
                    <div className="text-muted-foreground text-sm">
                      {session.location || t("security.sessions.unknownLocation")} • {session.ip_address}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      {t("security.sessions.lastActive")} {new Date(session.last_active_at).toLocaleString()}
                    </div>
                  </div>
                </div>
                {!session.is_current && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => revokeSessionMutation.mutate(session.id)}
                    disabled={revokeSessionMutation.isPending}
                    className="text-destructive hover:bg-destructive/10 hover:text-destructive"
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
