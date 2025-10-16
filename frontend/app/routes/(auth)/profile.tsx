import { createFileRoute } from "@tanstack/react-router";

import { ProtectedRoute } from "../../components/auth/ProtectedRoute";
import { UserProfile } from "../../components/auth/UserProfile";

export const Route = createFileRoute("/(auth)/profile")({
  component: ProfilePage,
});

function ProfilePage() {
  return (
    <ProtectedRoute>
      <div className="container mx-auto px-4 py-8">
        <UserProfile />
      </div>
    </ProtectedRoute>
  );
}
