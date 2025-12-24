import { createFileRoute } from "@tanstack/react-router";
import { Mail, Shield, User } from "lucide-react";

import { Button } from "../../components/ui/button";
import { SettingsSkeleton } from "../../components/ui/skeletons";
import { AuthService } from "../../services/auth/authService";

export const Route = createFileRoute("/(dashboard)/settings")({
  component: SettingsPage,
  pendingComponent: SettingsSkeleton,
  // Fetch current user settings from API
  loader: async ({ context }) => {
    const user = await context.queryClient.ensureQueryData({
      queryKey: ["currentUser"],
      queryFn: () => AuthService.getCurrentUser(),
      staleTime: 60 * 1000, // 1 minute
    });

    return { user };
  },
});

function SettingsPage() {
  const { user } = Route.useLoaderData();

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold">Settings</h1>

      <div className="space-y-6">
        {/* Profile Settings */}
        <div className="bg-card rounded-lg border p-6">
          <div className="mb-4 flex items-center gap-2">
            <User className="h-5 w-5" />
            <h2 className="text-lg font-semibold">Profile</h2>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Name</span>
              <span className="font-medium">{user.name}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Email</span>
              <span className="font-medium">{user.email}</span>
            </div>
          </div>
          <div className="mt-4">
            <Button variant="outline">Edit Profile</Button>
          </div>
        </div>

        {/* Account Status */}
        <div className="bg-card rounded-lg border p-6">
          <div className="mb-4 flex items-center gap-2">
            <Shield className="h-5 w-5" />
            <h2 className="text-lg font-semibold">Account Status</h2>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Role</span>
              <span className="bg-primary/10 text-primary rounded-full px-3 py-1 text-sm font-medium capitalize">
                {user.role || "user"}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Account Active</span>
              <span
                className={`rounded-full px-3 py-1 text-sm font-medium ${user.is_active ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}
              >
                {user.is_active ? "Active" : "Inactive"}
              </span>
            </div>
          </div>
        </div>

        {/* Email Verification */}
        <div className="bg-card rounded-lg border p-6">
          <div className="mb-4 flex items-center gap-2">
            <Mail className="h-5 w-5" />
            <h2 className="text-lg font-semibold">Email Verification</h2>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Email Verified</span>
            <span
              className={`rounded-full px-3 py-1 text-sm font-medium ${user.email_verified ? "bg-green-100 text-green-700" : "bg-yellow-100 text-yellow-700"}`}
            >
              {user.email_verified ? "Verified" : "Pending"}
            </span>
          </div>
          {!user.email_verified && (
            <div className="mt-4">
              <Button variant="outline">Resend Verification Email</Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
