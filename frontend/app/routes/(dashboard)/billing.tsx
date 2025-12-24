import { createFileRoute } from "@tanstack/react-router";
import { CreditCard, ExternalLink, Loader2 } from "lucide-react";

import { Button } from "../../components/ui/button";
import { useCreateCheckout, useCreatePortalSession } from "../../hooks/mutations/use-billing-mutations";
import { useBillingPlans, useSubscription } from "../../hooks/queries/use-billing";
import { BillingService } from "../../services/billing/billingService";
import type { BillingPlan, Subscription } from "../../services/types";

export const Route = createFileRoute("/(dashboard)/billing")({
  component: BillingPage,
});

function BillingPage() {
  const { data: subscription, isLoading: isLoadingSubscription } = useSubscription();
  const { data: plans, isLoading: isLoadingPlans } = useBillingPlans();
  const portalMutation = useCreatePortalSession();
  const checkoutMutation = useCreateCheckout();

  const handleManageSubscription = () => {
    portalMutation.mutate();
  };

  const handleSubscribe = (priceId: string) => {
    checkoutMutation.mutate(priceId);
  };

  if (isLoadingSubscription || isLoadingPlans) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold">Billing & Subscription</h1>

      <div className="space-y-6">
        {/* Current Subscription */}
        <SubscriptionCard
          subscription={subscription}
          onManage={handleManageSubscription}
          isManaging={portalMutation.isPending}
        />

        {/* Available Plans */}
        {!BillingService.isSubscriptionActive(subscription ?? null) && plans && plans.length > 0 && (
          <div className="bg-card rounded-lg border p-6">
            <h2 className="mb-4 text-lg font-semibold">Available Plans</h2>
            <div className="grid gap-4 md:grid-cols-2">
              {plans.map((plan) => (
                <PlanCard
                  key={plan.id}
                  plan={plan}
                  onSubscribe={() => handleSubscribe(plan.price_id)}
                  isLoading={checkoutMutation.isPending}
                />
              ))}
            </div>
          </div>
        )}

        {/* No Plans Available */}
        {!plans || plans.length === 0 ? (
          <div className="bg-card rounded-lg border p-6 text-center">
            <p className="text-muted-foreground">
              No subscription plans are currently available. Please check back later.
            </p>
          </div>
        ) : null}
      </div>
    </div>
  );
}

interface SubscriptionCardProps {
  subscription: Subscription | null | undefined;
  onManage: () => void;
  isManaging: boolean;
}

function SubscriptionCard({ subscription, onManage, isManaging }: SubscriptionCardProps) {
  const isActive = BillingService.isSubscriptionActive(subscription ?? null);

  return (
    <div className="bg-card rounded-lg border p-6">
      <div className="mb-4 flex items-center gap-2">
        <CreditCard className="h-5 w-5" />
        <h2 className="text-lg font-semibold">Current Subscription</h2>
      </div>

      {subscription ? (
        <div className="space-y-4">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Status</span>
              <StatusBadge status={subscription.status} />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Current Period</span>
              <span className="text-sm">
                {new Date(subscription.current_period_start).toLocaleDateString()} -{" "}
                {new Date(subscription.current_period_end).toLocaleDateString()}
              </span>
            </div>
            {subscription.cancel_at_period_end && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Cancels At</span>
                <span className="text-sm text-yellow-600">
                  {new Date(subscription.current_period_end).toLocaleDateString()}
                </span>
              </div>
            )}
          </div>

          <Button
            onClick={onManage}
            disabled={isManaging}
            className="w-full sm:w-auto"
          >
            {isManaging ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ExternalLink className="mr-2 h-4 w-4" />}
            Manage Subscription
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-muted-foreground">
            You don't have an active subscription. Choose a plan below to get started.
          </p>
        </div>
      )}
    </div>
  );
}

interface PlanCardProps {
  plan: BillingPlan;
  onSubscribe: () => void;
  isLoading: boolean;
}

function PlanCard({ plan, onSubscribe, isLoading }: PlanCardProps) {
  return (
    <div className="rounded-lg border p-4 transition-shadow hover:shadow-md">
      <h3 className="text-lg font-semibold">{plan.name}</h3>
      {plan.description && <p className="text-muted-foreground mt-1 text-sm">{plan.description}</p>}
      <div className="mt-4">
        <span className="text-2xl font-bold">{BillingService.formatPrice(plan.amount, plan.currency)}</span>
        <span className="text-muted-foreground">/{plan.interval}</span>
      </div>
      {plan.features && plan.features.length > 0 && (
        <ul className="mt-4 space-y-2">
          {plan.features.map((feature, index) => (
            <li
              key={index}
              className="flex items-center text-sm"
            >
              <span className="mr-2 text-green-500">✓</span>
              {feature}
            </li>
          ))}
        </ul>
      )}
      <Button
        onClick={onSubscribe}
        disabled={isLoading}
        className="mt-4 w-full"
      >
        {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
        Subscribe
      </Button>
    </div>
  );
}

interface StatusBadgeProps {
  status: string;
}

function StatusBadge({ status }: StatusBadgeProps) {
  const statusStyles: Record<string, string> = {
    active: "bg-green-100 text-green-700",
    trialing: "bg-blue-100 text-blue-700",
    past_due: "bg-yellow-100 text-yellow-700",
    canceled: "bg-gray-100 text-gray-700",
    unpaid: "bg-red-100 text-red-700",
  };

  return (
    <span
      className={`rounded-full px-3 py-1 text-sm font-medium capitalize ${statusStyles[status] || "bg-gray-100 text-gray-700"}`}
    >
      {status.replace("_", " ")}
    </span>
  );
}
