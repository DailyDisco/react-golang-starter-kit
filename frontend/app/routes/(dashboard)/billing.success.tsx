import { useEffect } from "react";

import { createFileRoute, Link } from "@tanstack/react-router";
import { CheckCircle } from "lucide-react";

import { Button } from "../../components/ui/button";
import { useRefreshSubscription } from "../../hooks/mutations/use-billing-mutations";

export const Route = createFileRoute("/(dashboard)/billing/success")({
  component: BillingSuccessPage,
});

function BillingSuccessPage() {
  const refreshMutation = useRefreshSubscription();

  // Refresh subscription data when the page loads
  useEffect(() => {
    refreshMutation.mutate();
  }, []);

  return (
    <div className="mx-auto max-w-md px-4 py-16 text-center">
      <div className="mb-6 flex justify-center">
        <div className="rounded-full bg-green-100 p-3">
          <CheckCircle className="h-12 w-12 text-green-600" />
        </div>
      </div>

      <h1 className="mb-4 text-2xl font-bold">Subscription Successful!</h1>

      <p className="text-muted-foreground mb-8">
        Thank you for subscribing! Your premium features are now active. You can manage your subscription at any time
        from your billing settings.
      </p>

      <div className="space-y-3">
        <Button
          asChild
          className="w-full"
        >
          <Link to="/billing">View Billing Settings</Link>
        </Button>
        <Button
          asChild
          variant="outline"
          className="w-full"
        >
          <Link to="/">Go to Dashboard</Link>
        </Button>
      </div>
    </div>
  );
}
