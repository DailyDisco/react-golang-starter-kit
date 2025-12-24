import { createFileRoute, Link } from "@tanstack/react-router";
import { XCircle } from "lucide-react";

import { Button } from "../../components/ui/button";

export const Route = createFileRoute("/(dashboard)/billing/cancel")({
  component: BillingCancelPage,
});

function BillingCancelPage() {
  return (
    <div className="mx-auto max-w-md px-4 py-16 text-center">
      <div className="mb-6 flex justify-center">
        <div className="rounded-full bg-gray-100 p-3">
          <XCircle className="h-12 w-12 text-gray-500" />
        </div>
      </div>

      <h1 className="mb-4 text-2xl font-bold">Checkout Canceled</h1>

      <p className="text-muted-foreground mb-8">
        Your checkout was canceled. No charges were made. If you'd like to try again or have any questions, please don't
        hesitate to reach out.
      </p>

      <div className="space-y-3">
        <Button
          asChild
          className="w-full"
        >
          <Link to="/billing">Try Again</Link>
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
