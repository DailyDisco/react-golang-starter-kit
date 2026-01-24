import { useEffect, useRef } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { UsageSummaryCard } from "@/components/usage/UsageSummaryCard";
import { useCreateCheckout, useCreatePortalSession } from "@/hooks/mutations/use-billing-mutations";
import { useBillingPlans, useSubscription } from "@/hooks/queries/use-billing";
import { queryKeys } from "@/lib/query-keys";
import { BillingService } from "@/services/billing/billingService";
import type { BillingPlan, Subscription } from "@/services/types";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { CreditCard, ExternalLink, Loader2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";

const billingSearchSchema = z.object({
  session_id: z.string().optional(),
  success: z.coerce.boolean().optional(),
  canceled: z.coerce.boolean().optional(),
});

export const Route = createFileRoute("/(app)/billing")({
  validateSearch: billingSearchSchema,
  component: BillingPage,
});

function BillingPage() {
  const { t } = useTranslation("billing");
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const search = Route.useSearch();
  const hasHandledCheckoutReturn = useRef(false);

  const { data: subscription, isLoading: isLoadingSubscription } = useSubscription();
  const { data: plans, isLoading: isLoadingPlans } = useBillingPlans();
  const portalMutation = useCreatePortalSession();
  const checkoutMutation = useCreateCheckout();

  // Handle return from Stripe checkout
  useEffect(() => {
    // Only process once per page load to avoid duplicate toasts
    if (hasHandledCheckoutReturn.current) return;

    const handleCheckoutReturn = async () => {
      if (search.session_id || search.success) {
        hasHandledCheckoutReturn.current = true;

        // Refresh subscription data
        await queryClient.invalidateQueries({ queryKey: queryKeys.billing.subscription() });
        await queryClient.invalidateQueries({ queryKey: queryKeys.auth.user });

        toast.success(t("checkout.success", "Subscription updated successfully!"));

        // Clean up URL params
        navigate({ to: "/billing", search: {}, replace: true });
      } else if (search.canceled) {
        hasHandledCheckoutReturn.current = true;
        toast.info(t("checkout.canceled", "Checkout was canceled"));
        navigate({ to: "/billing", search: {}, replace: true });
      }
    };

    handleCheckoutReturn();
  }, [search, queryClient, navigate, t]);

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
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
      </div>

      {/* Current Subscription */}
      <SubscriptionCard
        subscription={subscription}
        onManage={handleManageSubscription}
        isManaging={portalMutation.isPending}
      />

      {/* Usage Summary - shows current usage with upgrade prompts */}
      <UsageSummaryCard
        maxMetrics={2}
        warningThreshold={80}
        showDetailsLink={true}
      />

      {/* Available Plans */}
      {!BillingService.isSubscriptionActive(subscription ?? null) && plans && plans.length > 0 && (
        <div className="bg-card rounded-lg border p-6">
          <h2 className="mb-4 text-lg font-semibold">{t("availablePlans.title")}</h2>
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
          <p className="text-muted-foreground">{t("availablePlans.noPlans")}</p>
        </div>
      ) : null}
    </div>
  );
}

interface SubscriptionCardProps {
  subscription: Subscription | null | undefined;
  onManage: () => void;
  isManaging: boolean;
}

function SubscriptionCard({ subscription, onManage, isManaging }: SubscriptionCardProps) {
  const { t } = useTranslation("billing");

  return (
    <div className="bg-card rounded-lg border p-6">
      <div className="mb-4 flex items-center gap-2">
        <CreditCard className="h-5 w-5" />
        <h2 className="text-lg font-semibold">{t("currentSubscription.title")}</h2>
      </div>

      {subscription ? (
        <div className="space-y-4">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("currentSubscription.status")}</span>
              <StatusBadge status={subscription.status} />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("currentSubscription.currentPeriod")}</span>
              <span className="text-sm">
                {new Date(subscription.current_period_start).toLocaleDateString()} -{" "}
                {new Date(subscription.current_period_end).toLocaleDateString()}
              </span>
            </div>
            {subscription.cancel_at_period_end && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("currentSubscription.cancelsAt")}</span>
                <span className="text-warning text-sm">
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
            {t("currentSubscription.manageSubscription")}
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-muted-foreground">{t("currentSubscription.noSubscription")}</p>
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
  const { t } = useTranslation("billing");

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
              <span className="text-success mr-2">✓</span>
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
        {t("availablePlans.subscribe")}
      </Button>
    </div>
  );
}

interface StatusBadgeProps {
  status: string;
}

function StatusBadge({ status }: StatusBadgeProps) {
  const { t } = useTranslation("billing");

  const statusVariants: Record<string, "success" | "info" | "warning" | "destructive" | "secondary"> = {
    active: "success",
    trialing: "info",
    past_due: "warning",
    canceled: "secondary",
    unpaid: "destructive",
  };

  return (
    <Badge variant={statusVariants[status] || "secondary"}>
      {t(`status.${status}`, { defaultValue: status.replace("_", " ") })}
    </Badge>
  );
}
