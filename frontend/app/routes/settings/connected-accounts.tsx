import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Check, Github, Link2, Loader2, Unlink } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { requireAuth } from "../../lib/guards";
import { SettingsService, type ConnectedAccount } from "../../services/settings/settingsService";
import { API_BASE_URL } from "../../services/api/client";

export const Route = createFileRoute("/settings/connected-accounts")({
  beforeLoad: () => requireAuth(),
  component: ConnectedAccountsPage,
});

// Provider configurations
const providers = [
  {
    id: "google",
    name: "Google",
    icon: GoogleIcon,
    description: "Sign in with your Google account",
    color: "bg-white border-gray-300 hover:bg-gray-50",
  },
  {
    id: "github",
    name: "GitHub",
    icon: Github,
    description: "Sign in with your GitHub account",
    color: "bg-gray-900 text-white hover:bg-gray-800",
  },
];

function ConnectedAccountsPage() {
  const queryClient = useQueryClient();

  const { data: connectedAccounts, isLoading } = useQuery({
    queryKey: ["connected-accounts"],
    queryFn: () => SettingsService.getConnectedAccounts(),
  });

  const disconnectMutation = useMutation({
    mutationFn: (provider: string) => SettingsService.disconnectAccount(provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["connected-accounts"] });
      toast.success("Account has been disconnected.");
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const handleConnect = (providerId: string) => {
    // Redirect to OAuth flow
    window.location.href = `${API_BASE_URL}/api/auth/${providerId}?link=true`;
  };

  const isConnected = (providerId: string) => {
    return connectedAccounts?.some((acc) => acc.provider === providerId);
  };

  const getConnectedAccount = (providerId: string) => {
    return connectedAccounts?.find((acc) => acc.provider === providerId);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Connected Accounts</h2>
        <p className="text-sm text-gray-500">
          Link third-party accounts for easier sign-in
        </p>
      </div>

      {/* Providers List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Link2 className="h-5 w-5" />
            OAuth Providers
          </CardTitle>
          <CardDescription>
            Connect your accounts to enable single sign-on
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2].map((i) => (
                <div key={i} className="h-20 animate-pulse rounded-lg bg-gray-100" />
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
                      connected ? "border-green-200 bg-green-50" : ""
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <div
                        className={`flex h-12 w-12 items-center justify-center rounded-lg ${
                          provider.id === "github"
                            ? "bg-gray-900 text-white"
                            : "border bg-white"
                        }`}
                      >
                        <Icon className="h-6 w-6" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-medium">{provider.name}</p>
                          {connected && (
                            <Badge variant="secondary" className="bg-green-100 text-green-700">
                              <Check className="mr-1 h-3 w-3" />
                              Connected
                            </Badge>
                          )}
                        </div>
                        <p className="text-sm text-gray-500">
                          {connected && account
                            ? account.email
                            : provider.description}
                        </p>
                        {connected && account && (
                          <p className="text-xs text-gray-400">
                            Connected on{" "}
                            {new Date(account.connected_at).toLocaleDateString()}
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
                        className="text-red-600 hover:bg-red-50 hover:text-red-700"
                      >
                        {disconnectMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <>
                            <Unlink className="mr-2 h-4 w-4" />
                            Disconnect
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
                        Connect
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
            <div className="rounded-full bg-blue-100 p-2">
              <Link2 className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <h4 className="font-medium">Why connect accounts?</h4>
              <ul className="mt-2 list-inside list-disc space-y-1 text-sm text-gray-600">
                <li>Sign in faster without entering your password</li>
                <li>Securely link your identity across platforms</li>
                <li>Enable features that integrate with third-party services</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Security Notice */}
      <Card className="border-amber-200 bg-amber-50">
        <CardContent className="py-4">
          <div className="flex items-start gap-3">
            <div className="mt-0.5 text-amber-600">
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
              <p className="font-medium text-amber-900">Security Note</p>
              <p className="mt-1 text-sm text-amber-700">
                Disconnecting an account won't delete any data from that service.
                You can always reconnect later. We only store your email and unique
                identifier from connected accounts.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// Google icon component
function GoogleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
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
