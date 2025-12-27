import { createFileRoute, Link } from "@tanstack/react-router";
import { Check, Loader2, X } from "lucide-react";

import { Button } from "../../components/ui/button";
import { useCreateCheckout } from "../../hooks/mutations/use-billing-mutations";
import { useBillingPlans } from "../../hooks/queries/use-billing";
import { BillingService } from "../../services/billing/billingService";
import type { BillingPlan } from "../../services/types";
import { useAuthStore } from "../../stores/auth-store";

export const Route = createFileRoute("/(public)/pricing")({
  component: PricingPage,
});

// Static fallback plans when Stripe is not configured
const staticPlans = [
  {
    id: "free",
    name: "Free",
    description: "Perfect for getting started",
    amount: 0,
    currency: "usd",
    interval: "month",
    features: [
      "Up to 3 projects",
      "Basic analytics",
      "Community support",
      "1GB storage",
      "API access (100 req/day)",
    ],
    limitations: [
      "No priority support",
      "No custom branding",
      "No team members",
    ],
  },
  {
    id: "pro",
    name: "Pro",
    description: "For professionals and growing teams",
    amount: 1900, // $19.00
    currency: "usd",
    interval: "month",
    features: [
      "Unlimited projects",
      "Advanced analytics",
      "Priority email support",
      "25GB storage",
      "API access (10,000 req/day)",
      "Custom branding",
      "Up to 5 team members",
      "Two-factor authentication",
    ],
    limitations: [
      "No dedicated support",
    ],
    popular: true,
  },
  {
    id: "enterprise",
    name: "Enterprise",
    description: "For large organizations with advanced needs",
    amount: 9900, // $99.00
    currency: "usd",
    interval: "month",
    features: [
      "Everything in Pro",
      "Unlimited storage",
      "Unlimited API requests",
      "Unlimited team members",
      "Dedicated account manager",
      "24/7 phone support",
      "Custom integrations",
      "SLA guarantee",
      "SSO/SAML",
      "Audit logs",
    ],
    limitations: [],
  },
];

function PricingPage() {
  const { data: plans, isLoading, error } = useBillingPlans();
  const checkoutMutation = useCreateCheckout();
  const { isAuthenticated } = useAuthStore();

  const navigate = Route.useNavigate();

  // Use Stripe plans if available, otherwise use static fallback
  const hasStripePlans = plans && plans.length > 0;
  const displayPlans = hasStripePlans ? plans : null;

  const handleSubscribe = (priceId: string) => {
    if (!isAuthenticated) {
      void navigate({ to: "/login", search: { redirect: "/pricing" } });
      return;
    }
    if (hasStripePlans) {
      checkoutMutation.mutate(priceId);
    } else {
      // For static plans, redirect to contact or register
      void navigate({ to: "/register" });
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-16">
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="mb-4 text-4xl font-bold">Simple, Transparent Pricing</h1>
        <p className="text-muted-foreground mx-auto max-w-2xl text-lg">
          Choose the plan that's right for you. All paid plans include a 14-day free trial.
        </p>
      </div>

      {/* Pricing Cards */}
      <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
        {hasStripePlans ? (
          <>
            {/* Free Plan (always shown) */}
            <StaticPlanCard
              plan={staticPlans[0]}
              onAction={() => navigate({ to: "/register" })}
              actionLabel="Get Started"
              variant="outline"
            />
            {/* Dynamic Plans from Stripe */}
            {displayPlans?.map((plan) => (
              <PricingCard
                key={plan.id}
                plan={plan}
                onSubscribe={() => handleSubscribe(plan.price_id)}
                isLoading={checkoutMutation.isPending}
                isAuthenticated={isAuthenticated}
              />
            ))}
          </>
        ) : (
          /* Static fallback plans */
          staticPlans.map((plan, index) => (
            <StaticPlanCard
              key={plan.id}
              plan={plan}
              onAction={() => {
                if (plan.id === "free") {
                  void navigate({ to: "/register" });
                } else if (plan.id === "enterprise") {
                  window.location.href = "mailto:sales@example.com?subject=Enterprise%20Plan%20Inquiry";
                } else {
                  void navigate({ to: "/register" });
                }
              }}
              actionLabel={
                plan.id === "free" ? "Get Started" :
                plan.id === "enterprise" ? "Contact Sales" :
                "Start Free Trial"
              }
              variant={plan.id === "free" ? "outline" : "default"}
            />
          ))
        )}
      </div>

      {/* Feature Comparison Table */}
      {!hasStripePlans && (
        <div className="mt-16">
          <h2 className="mb-8 text-center text-2xl font-bold">Compare Plans</h2>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="border-b">
                  <th className="p-4 text-left font-medium">Feature</th>
                  <th className="p-4 text-center font-medium">Free</th>
                  <th className="bg-primary/5 p-4 text-center font-medium">Pro</th>
                  <th className="p-4 text-center font-medium">Enterprise</th>
                </tr>
              </thead>
              <tbody>
                <ComparisonRow feature="Projects" free="3" pro="Unlimited" enterprise="Unlimited" />
                <ComparisonRow feature="Storage" free="1GB" pro="25GB" enterprise="Unlimited" />
                <ComparisonRow feature="API Requests" free="100/day" pro="10,000/day" enterprise="Unlimited" />
                <ComparisonRow feature="Team Members" free={false} pro="5" enterprise="Unlimited" />
                <ComparisonRow feature="Custom Branding" free={false} pro={true} enterprise={true} />
                <ComparisonRow feature="Priority Support" free={false} pro={true} enterprise={true} />
                <ComparisonRow feature="Two-Factor Auth" free={true} pro={true} enterprise={true} />
                <ComparisonRow feature="SSO/SAML" free={false} pro={false} enterprise={true} />
                <ComparisonRow feature="Audit Logs" free={false} pro={false} enterprise={true} />
                <ComparisonRow feature="SLA Guarantee" free={false} pro={false} enterprise={true} />
                <ComparisonRow feature="Dedicated Support" free={false} pro={false} enterprise={true} />
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* FAQ Section */}
      <div className="mt-16">
        <h2 className="mb-8 text-center text-2xl font-bold">Frequently Asked Questions</h2>
        <div className="mx-auto grid max-w-4xl gap-6 md:grid-cols-2">
          <FaqItem
            question="Can I change plans later?"
            answer="Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and we'll prorate your billing."
          />
          <FaqItem
            question="What payment methods do you accept?"
            answer="We accept all major credit cards (Visa, Mastercard, American Express) and can arrange invoicing for Enterprise plans."
          />
          <FaqItem
            question="Is there a free trial?"
            answer="Yes! All paid plans come with a 14-day free trial. No credit card required to start."
          />
          <FaqItem
            question="What happens when I exceed my limits?"
            answer="We'll notify you when you're approaching your limits. You can upgrade anytime, or we'll temporarily pause non-critical features."
          />
        </div>
      </div>

      {/* Contact CTA */}
      <div className="mt-16 text-center">
        <p className="text-muted-foreground">
          Have more questions?{" "}
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

function LimitationItem({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-center gap-2">
      <X className="text-muted-foreground h-4 w-4" />
      <span className="text-muted-foreground text-sm">{children}</span>
    </li>
  );
}

interface StaticPlan {
  id: string;
  name: string;
  description: string;
  amount: number;
  currency: string;
  interval: string;
  features: string[];
  limitations?: string[];
  popular?: boolean;
}

interface StaticPlanCardProps {
  plan: StaticPlan;
  onAction: () => void;
  actionLabel: string;
  variant?: "default" | "outline";
}

function StaticPlanCard({ plan, onAction, actionLabel, variant = "default" }: StaticPlanCardProps) {
  return (
    <div
      className={`relative rounded-2xl border p-8 ${
        plan.popular ? "border-primary bg-primary/5 shadow-lg" : "bg-card"
      }`}
    >
      {plan.popular && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <span className="bg-primary text-primary-foreground rounded-full px-3 py-1 text-xs font-medium">
            Most Popular
          </span>
        </div>
      )}

      <h3 className="mb-2 text-xl font-semibold">{plan.name}</h3>
      <p className="text-muted-foreground mb-4 text-sm">{plan.description}</p>

      <div className="mb-6">
        <span className="text-4xl font-bold">
          {plan.amount === 0 ? "$0" : BillingService.formatPrice(plan.amount, plan.currency)}
        </span>
        <span className="text-muted-foreground">/{plan.interval}</span>
      </div>

      <ul className="mb-4 space-y-2">
        {plan.features.map((feature, index) => (
          <FeatureItem key={index}>{feature}</FeatureItem>
        ))}
      </ul>

      {plan.limitations && plan.limitations.length > 0 && (
        <ul className="mb-6 space-y-2">
          {plan.limitations.map((limitation, index) => (
            <LimitationItem key={index}>{limitation}</LimitationItem>
          ))}
        </ul>
      )}

      <Button onClick={onAction} variant={variant} className="w-full">
        {actionLabel}
      </Button>
    </div>
  );
}

interface ComparisonRowProps {
  feature: string;
  free: boolean | string;
  pro: boolean | string;
  enterprise: boolean | string;
}

function ComparisonRow({ feature, free, pro, enterprise }: ComparisonRowProps) {
  const renderValue = (value: boolean | string) => {
    if (typeof value === "boolean") {
      return value ? (
        <Check className="mx-auto h-5 w-5 text-green-500" />
      ) : (
        <X className="text-muted-foreground mx-auto h-5 w-5" />
      );
    }
    return <span className="text-sm">{value}</span>;
  };

  return (
    <tr className="border-b">
      <td className="p-4 text-left text-sm">{feature}</td>
      <td className="p-4 text-center">{renderValue(free)}</td>
      <td className="bg-primary/5 p-4 text-center">{renderValue(pro)}</td>
      <td className="p-4 text-center">{renderValue(enterprise)}</td>
    </tr>
  );
}

interface FaqItemProps {
  question: string;
  answer: string;
}

function FaqItem({ question, answer }: FaqItemProps) {
  return (
    <div className="bg-card rounded-lg border p-6">
      <h3 className="mb-2 font-semibold">{question}</h3>
      <p className="text-muted-foreground text-sm">{answer}</p>
    </div>
  );
}
