import { createFileRoute } from "@tanstack/react-router";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { requireAdmin } from "../../lib/guards";

export const Route = createFileRoute("/admin/settings")({
  beforeLoad: () => requireAdmin(),
  component: AdminSettingsPage,
});

function AdminSettingsPage() {
  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Settings</h2>

      <Card>
        <CardHeader>
          <CardTitle>Application Settings</CardTitle>
          <CardDescription>Configure application-wide settings and preferences.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-gray-500">Settings management coming soon. This section will allow you to configure:</p>
          <ul className="mt-4 space-y-2 text-sm text-gray-600">
            <li>- Email configuration</li>
            <li>- OAuth provider settings</li>
            <li>- Rate limiting rules</li>
            <li>- Cache settings</li>
            <li>- Backup and restore</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
