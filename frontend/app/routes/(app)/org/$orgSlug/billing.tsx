import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { queryKeys } from "@/lib/query-keys";
import { OrganizationService, type OrgBillingInfo } from "@/services/organizations/organizationService";
import { useCurrentOrg, useIsOrgOwner } from "@/stores/org-store";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { format } from "date-fns";
import { CreditCard, ExternalLink, Loader2, Settings, Sparkles, Users, Zap } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/(app)/org/$orgSlug/billing")({
  component: OrgBillingPage,
});

function OrgBillingPage() {
  const { orgSlug } = Route.useParams();
  const queryClient = useQueryClient();
  const currentOrg = useCurrentOrg();
  const isOwner = useIsOrgOwner();

  // Fetch billing info
  const {
    data: billing,
    isLoading,
    error,
  } = useQuery({
    queryKey: queryKeys.organizations.billing(orgSlug),
    queryFn: () => OrganizationService.getBilling(orgSlug),
  });

  // Create checkout session mutation
  const checkoutMutation = useMutation({
    mutationFn: (priceId: string) => OrganizationService.createCheckoutSession(orgSlug, priceId),
    onSuccess: (data) => {
      // Redirect to Stripe checkout
      window.location.href = data.url;
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Create portal session mutation
  const portalMutation = useMutation({
    mutationFn: () => OrganizationService.createPortalSession(orgSlug),
    onSuccess: (data) => {
      // Redirect to Stripe portal
      window.location.href = data.url;
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return <BillingPageSkeleton />;
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-destructive">Failed to load billing information</p>
      </div>
    );
  }

  if (!billing || !currentOrg) {
    return null;
  }

  const seatUsagePercent = billing.seat_limit > 0 ? (billing.seat_count / billing.seat_limit) * 100 : 0;
  const isNearLimit = seatUsagePercent >= 80;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold">Organization Billing</h2>
        <p className="text-muted-foreground text-sm">Manage billing and subscription for {currentOrg.name}</p>
      </div>

      {/* Current Plan Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Sparkles className="h-5 w-5" />
                Current Plan
              </CardTitle>
              <CardDescription>Your organization's subscription details</CardDescription>
            </div>
            <Badge
              variant={billing.plan === "free" ? "secondary" : "default"}
              className="px-3 py-1 text-base capitalize"
            >
              {billing.plan}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Subscription Status */}
          {billing.subscription ? (
            <div className="space-y-3 rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Status</span>
                <SubscriptionStatusBadge status={billing.subscription.status} />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Current Period</span>
                <span className="text-muted-foreground text-sm">
                  {format(new Date(billing.subscription.current_period_start), "MMM d, yyyy")} -{" "}
                  {format(new Date(billing.subscription.current_period_end), "MMM d, yyyy")}
                </span>
              </div>
              {billing.subscription.cancel_at_period_end && (
                <div className="text-sm text-amber-600 dark:text-amber-400">
                  Subscription will cancel at the end of the current period
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-lg border border-dashed p-4 text-center">
              <p className="text-muted-foreground text-sm">No active subscription</p>
            </div>
          )}

          {/* Seat Usage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Users className="text-muted-foreground h-4 w-4" />
                <span className="text-sm font-medium">Team Seats</span>
              </div>
              <span className={`text-sm ${isNearLimit ? "text-amber-600" : "text-muted-foreground"}`}>
                {billing.seat_count} / {billing.seat_limit === -1 ? "Unlimited" : billing.seat_limit}
              </span>
            </div>
            {billing.seat_limit > 0 && (
              <Progress
                value={seatUsagePercent}
                className={`h-2 ${isNearLimit ? "[&>div]:bg-amber-500" : ""}`}
              />
            )}
            {isNearLimit && billing.seat_limit > 0 && (
              <p className="text-xs text-amber-600">
                You're approaching your seat limit. Consider upgrading for more seats.
              </p>
            )}
          </div>

          {/* Actions */}
          {isOwner && (
            <div className="flex flex-wrap gap-3 pt-2">
              {billing.has_subscription ? (
                <Button
                  variant="outline"
                  onClick={() => portalMutation.mutate()}
                  disabled={portalMutation.isPending}
                >
                  {portalMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Settings className="mr-2 h-4 w-4" />
                  )}
                  Manage Subscription
                </Button>
              ) : (
                <Button
                  onClick={() => checkoutMutation.mutate("price_pro_monthly")}
                  disabled={checkoutMutation.isPending}
                >
                  {checkoutMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Zap className="mr-2 h-4 w-4" />
                  )}
                  Upgrade to Pro
                </Button>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Plan Comparison */}
      {billing.plan === "free" && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5" />
              Upgrade Your Plan
            </CardTitle>
            <CardDescription>Get more features and seats for your team</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              {/* Pro Plan */}
              <div className="space-y-4 rounded-lg border p-4">
                <div>
                  <h3 className="text-lg font-semibold">Pro</h3>
                  <p className="text-2xl font-bold">
                    $25<span className="text-muted-foreground text-sm font-normal">/month</span>
                  </p>
                </div>
                <ul className="space-y-2 text-sm">
                  <li className="flex items-center gap-2">
                    <Users className="text-primary h-4 w-4" />
                    Up to 25 team members
                  </li>
                  <li className="flex items-center gap-2">
                    <CreditCard className="text-primary h-4 w-4" />
                    Priority support
                  </li>
                  <li className="flex items-center gap-2">
                    <Sparkles className="text-primary h-4 w-4" />
                    Advanced features
                  </li>
                </ul>
                {isOwner && (
                  <Button
                    className="w-full"
                    onClick={() => checkoutMutation.mutate("price_pro_monthly")}
                    disabled={checkoutMutation.isPending}
                  >
                    {checkoutMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                    Upgrade to Pro
                  </Button>
                )}
              </div>

              {/* Enterprise Plan */}
              <div className="space-y-4 rounded-lg border p-4">
                <div>
                  <h3 className="text-lg font-semibold">Enterprise</h3>
                  <p className="text-2xl font-bold">Custom</p>
                </div>
                <ul className="space-y-2 text-sm">
                  <li className="flex items-center gap-2">
                    <Users className="text-primary h-4 w-4" />
                    Unlimited team members
                  </li>
                  <li className="flex items-center gap-2">
                    <CreditCard className="text-primary h-4 w-4" />
                    Dedicated support
                  </li>
                  <li className="flex items-center gap-2">
                    <Sparkles className="text-primary h-4 w-4" />
                    Custom integrations
                  </li>
                </ul>
                <Button
                  variant="outline"
                  className="w-full"
                  asChild
                >
                  <a href="mailto:sales@example.com">
                    Contact Sales
                    <ExternalLink className="ml-2 h-4 w-4" />
                  </a>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Billing Portal Link for subscribers */}
      {billing.has_subscription && isOwner && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CreditCard className="h-5 w-5" />
              Payment & Invoices
            </CardTitle>
            <CardDescription>Manage payment methods and view invoices</CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              variant="outline"
              onClick={() => portalMutation.mutate()}
              disabled={portalMutation.isPending}
            >
              {portalMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <ExternalLink className="mr-2 h-4 w-4" />
              )}
              Open Billing Portal
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Non-owner notice */}
      {!isOwner && (
        <Card className="border-amber-200 dark:border-amber-800">
          <CardContent className="pt-6">
            <p className="text-muted-foreground text-sm">
              Only organization owners can manage billing and subscriptions. Contact your organization owner for billing
              changes.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function SubscriptionStatusBadge({ status }: { status: string }) {
  const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
    active: "default",
    trialing: "secondary",
    past_due: "destructive",
    canceled: "outline",
    incomplete: "secondary",
  };

  return (
    <Badge
      variant={variants[status] || "secondary"}
      className="capitalize"
    >
      {status.replace("_", " ")}
    </Badge>
  );
}

function BillingPageSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <Skeleton className="h-8 w-48" />
        <Skeleton className="mt-2 h-4 w-64" />
      </div>
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-10 w-40" />
        </CardContent>
      </Card>
    </div>
  );
}
