import { createFileRoute, Link } from "@tanstack/react-router";
import { Check, Loader2 } from "lucide-react";

import { Button } from "../../components/ui/button";
import { useCreateCheckout } from "../../hooks/mutations/use-billing-mutations";
import { useBillingPlans } from "../../hooks/queries/use-billing";
import { BillingService } from "../../services/billing/billingService";
import type { BillingPlan } from "../../services/types";
import { useAuthStore } from "../../stores/auth-store";

export const Route = createFileRoute("/(public)/pricing")({
  component: PricingPage,
});

function PricingPage() {
  const { data: plans, isLoading, error } = useBillingPlans();
  const checkoutMutation = useCreateCheckout();
  const { isAuthenticated } = useAuthStore();

  const navigate = Route.useNavigate();

  const handleSubscribe = (priceId: string) => {
    if (!isAuthenticated) {
      // Redirect to login with return URL
      void navigate({ to: "/login", search: { redirect: "/pricing" } });
      return;
    }
    checkoutMutation.mutate(priceId);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (error || !plans || plans.length === 0) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-16 text-center">
        <h1 className="mb-4 text-3xl font-bold">Pricing</h1>
        <p className="text-muted-foreground">Pricing information is currently unavailable. Please check back later.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-16">
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="mb-4 text-4xl font-bold">Simple, Transparent Pricing</h1>
        <p className="text-muted-foreground mx-auto max-w-2xl text-lg">
          Choose the plan that's right for you. All plans include a 14-day free trial.
        </p>
      </div>

      {/* Pricing Cards */}
      <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
        {/* Free Plan */}
        <div className="bg-card rounded-2xl border p-8">
          <h3 className="mb-2 text-xl font-semibold">Free</h3>
          <p className="text-muted-foreground mb-4 text-sm">Perfect for getting started</p>
          <div className="mb-6">
            <span className="text-4xl font-bold">$0</span>
            <span className="text-muted-foreground">/month</span>
          </div>
          <ul className="mb-8 space-y-3">
            <FeatureItem>Basic features</FeatureItem>
            <FeatureItem>Community support</FeatureItem>
            <FeatureItem>Up to 3 projects</FeatureItem>
          </ul>
          <Button
            variant="outline"
            className="w-full"
            asChild
          >
            <Link to="/register">Get Started</Link>
          </Button>
        </div>

        {/* Dynamic Plans from Stripe */}
        {plans.map((plan) => (
          <PricingCard
            key={plan.id}
            plan={plan}
            onSubscribe={() => handleSubscribe(plan.price_id)}
            isLoading={checkoutMutation.isPending}
            isAuthenticated={isAuthenticated}
          />
        ))}
      </div>

      {/* FAQ or additional info */}
      <div className="mt-16 text-center">
        <p className="text-muted-foreground">
          Have questions?{" "}
          <a
            href="mailto:support@example.com"
            className="text-primary hover:underline"
          >
            Contact our sales team
          </a>
        </p>
      </div>
    </div>
  );
}

interface PricingCardProps {
  plan: BillingPlan;
  onSubscribe: () => void;
  isLoading: boolean;
  isAuthenticated: boolean;
}

function PricingCard({ plan, onSubscribe, isLoading, isAuthenticated }: PricingCardProps) {
  const isPopular = plan.name.toLowerCase().includes("premium") || plan.name.toLowerCase().includes("pro");

  return (
    <div
      className={`relative rounded-2xl border p-8 ${isPopular ? "border-primary bg-primary/5 shadow-lg" : "bg-card"}`}
    >
      {isPopular && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <span className="bg-primary text-primary-foreground rounded-full px-3 py-1 text-xs font-medium">
            Most Popular
          </span>
        </div>
      )}

      <h3 className="mb-2 text-xl font-semibold">{plan.name}</h3>
      {plan.description && <p className="text-muted-foreground mb-4 text-sm">{plan.description}</p>}

      <div className="mb-6">
        <span className="text-4xl font-bold">{BillingService.formatPrice(plan.amount, plan.currency)}</span>
        <span className="text-muted-foreground">/{plan.interval}</span>
      </div>

      <ul className="mb-8 space-y-3">
        {plan.features && plan.features.length > 0 ? (
          plan.features.map((feature, index) => <FeatureItem key={index}>{feature}</FeatureItem>)
        ) : (
          <>
            <FeatureItem>All Free features</FeatureItem>
            <FeatureItem>Priority support</FeatureItem>
            <FeatureItem>Unlimited projects</FeatureItem>
            <FeatureItem>Advanced analytics</FeatureItem>
          </>
        )}
      </ul>

      <Button
        onClick={onSubscribe}
        disabled={isLoading}
        className="w-full"
      >
        {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
        {isAuthenticated ? "Subscribe Now" : "Sign Up to Subscribe"}
      </Button>
    </div>
  );
}

function FeatureItem({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-center gap-2">
      <Check className="h-4 w-4 text-green-500" />
      <span className="text-sm">{children}</span>
    </li>
  );
}
