import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { queryKeys } from "@/lib/query-keys";
import { API_BASE_URL } from "@/services/api/client";
import { SettingsService, type ConnectedAccount } from "@/services/settings/settingsService";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Check, Github, Link2, Loader2, Unlink } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/settings/connected-accounts")({
  component: ConnectedAccountsPage,
});

function ConnectedAccountsPage() {
  const { t } = useTranslation("settings");
  const queryClient = useQueryClient();

  // Provider configurations with translation keys
  const providers = [
    {
      id: "google",
      nameKey: "connectedAccounts.providers.google.name",
      descriptionKey: "connectedAccounts.providers.google.description",
      icon: GoogleIcon,
      color: "bg-white border-gray-300 hover:bg-gray-50",
    },
    {
      id: "github",
      nameKey: "connectedAccounts.providers.github.name",
      descriptionKey: "connectedAccounts.providers.github.description",
      icon: Github,
      color: "bg-gray-900 text-white hover:bg-gray-800",
    },
  ];

  const { data: connectedAccounts, isLoading } = useQuery({
    queryKey: queryKeys.settings.connectedAccounts(),
    queryFn: () => SettingsService.getConnectedAccounts(),
  });

  const disconnectMutation = useMutation({
    mutationFn: (provider: string) => SettingsService.disconnectAccount(provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.connectedAccounts() });
      toast.success(t("connectedAccounts.toast.disconnected"));
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleConnect = (providerId: string) => {
    window.location.assign(`${API_BASE_URL}/api/v1/auth/${providerId}?link=true`);
  };

  const isConnected = (providerId: string) => {
    return connectedAccounts?.some((acc) => acc.provider === providerId);
  };

  const getConnectedAccount = (providerId: string) => {
    return connectedAccounts?.find((acc) => acc.provider === providerId);
  };

  return (
    <SettingsLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold">{t("connectedAccounts.title")}</h2>
          <p className="text-muted-foreground text-sm">{t("connectedAccounts.subtitle")}</p>
        </div>

        {/* Providers List */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5" />
              {t("connectedAccounts.providers.title")}
            </CardTitle>
            <CardDescription>{t("connectedAccounts.providers.subtitle")}</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-4">
                {[1, 2].map((i) => (
                  <div
                    key={i}
                    className="bg-muted h-20 animate-pulse rounded-lg"
                  />
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                {providers.map((provider) => {
                  const connected = isConnected(provider.id);
                  const account = getConnectedAccount(provider.id);
                  const Icon = provider.icon;

                  return (
                    <div
                      key={provider.id}
                      className={`flex items-center justify-between rounded-lg border p-4 ${
                        connected ? "border-success/30 bg-success/5" : ""
                      }`}
                    >
                      <div className="flex items-center gap-4">
                        <div
                          className={`flex h-12 w-12 items-center justify-center rounded-lg ${
                            provider.id === "github" ? "bg-foreground text-background" : "bg-background border"
                          }`}
                        >
                          <Icon className="h-6 w-6" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-medium">{t(provider.nameKey as never)}</p>
                            {connected && <Badge variant="success">{t("connectedAccounts.connected")}</Badge>}
                          </div>
                          <p className="text-muted-foreground text-sm">
                            {connected && account ? account.email : t(provider.descriptionKey as never)}
                          </p>
                          {connected && account && (
                            <p className="text-muted-foreground text-xs">
                              {t("connectedAccounts.connectedOn")} {new Date(account.connected_at).toLocaleDateString()}
                            </p>
                          )}
                        </div>
                      </div>

                      {connected ? (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => disconnectMutation.mutate(provider.id)}
                          disabled={disconnectMutation.isPending}
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        >
                          {disconnectMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <>
                              <Unlink className="mr-2 h-4 w-4" />
                              {t("connectedAccounts.disconnect")}
                            </>
                          )}
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleConnect(provider.id)}
                        >
                          <Link2 className="mr-2 h-4 w-4" />
                          {t("connectedAccounts.connect")}
                        </Button>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Info Card */}
        <Card>
          <CardContent className="py-4">
            <div className="flex items-start gap-4">
              <div className="bg-info/10 rounded-full p-2">
                <Link2 className="text-info h-5 w-5" />
              </div>
              <div>
                <h4 className="font-medium">{t("connectedAccounts.whyConnect.title")}</h4>
                <ul className="text-muted-foreground mt-2 list-inside list-disc space-y-1 text-sm">
                  <li>{t("connectedAccounts.whyConnect.reasons.faster")}</li>
                  <li>{t("connectedAccounts.whyConnect.reasons.secure")}</li>
                  <li>{t("connectedAccounts.whyConnect.reasons.features")}</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Security Notice */}
        <Card className="border-warning/30 bg-warning/5">
          <CardContent className="py-4">
            <div className="flex items-start gap-3">
              <div className="text-warning mt-0.5">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  className="h-5 w-5"
                >
                  <path
                    fillRule="evenodd"
                    d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25zM12.75 9a.75.75 0 00-1.5 0v2.25H9a.75.75 0 000 1.5h2.25V15a.75.75 0 001.5 0v-2.25H15a.75.75 0 000-1.5h-2.25V9z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div>
                <p className="font-medium">{t("connectedAccounts.securityNote.title")}</p>
                <p className="text-muted-foreground mt-1 text-sm">{t("connectedAccounts.securityNote.text")}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </SettingsLayout>
  );
}

// Google icon component
function GoogleIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
    >
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      />
    </svg>
  );
}
