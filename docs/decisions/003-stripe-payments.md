# ADR 003: Stripe Payments Integration

## Status

Accepted

## Context

The application needs to support subscription-based billing with the following requirements:

- Monthly/annual subscription plans
- Self-service subscription management
- Secure payment processing
- Webhook-based event handling
- Role-based access for premium features

## Decision

We will use Stripe with the following components:

### Stripe Checkout

Server-side checkout session creation that redirects to Stripe's hosted checkout page. This provides:
- PCI compliance out of the box
- Mobile-optimized payment forms
- Support for multiple payment methods
- Built-in fraud protection

### Stripe Customer Portal

Self-service portal for customers to:
- Update payment methods
- View invoices
- Cancel/modify subscriptions
- Download receipts

### Webhook Processing

Server-side webhook handler for real-time event processing:
- `checkout.session.completed` - Create subscription record
- `customer.subscription.updated` - Sync subscription status
- `customer.subscription.deleted` - Handle cancellation
- `invoice.payment_failed` - Handle failed payments

### Architecture

```
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│   Frontend   │────>│  Backend API  │────>│    Stripe    │
│   /billing   │     │   /checkout   │     │   Checkout   │
└──────────────┘     └───────────────┘     └──────────────┘
                                                  │
                                                  ▼
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│   Database   │<────│    Webhook    │<────│    Stripe    │
│ subscriptions│     │    Handler    │     │   Webhooks   │
└──────────────┘     └───────────────┘     └──────────────┘
```

### Role Sync Strategy

| Subscription Status | User Role |
|---------------------|-----------|
| active, trialing | premium |
| past_due | premium (grace period) |
| canceled, unpaid | user |

## Consequences

### Positive

- PCI compliance without handling card data
- Stripe handles tax, invoicing, receipts
- Customer Portal reduces support burden
- Webhooks ensure data consistency
- Proven, reliable payment infrastructure

### Negative

- Stripe fees (2.9% + $0.30 per transaction)
- Dependency on external service
- Need to handle webhook reliability
- Must keep Stripe data in sync with local DB

## Configuration

```bash
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_SUCCESS_URL=http://localhost:5173/billing/success
STRIPE_CANCEL_URL=http://localhost:5173/billing/cancel
```

### Testing with Stripe CLI

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Forward webhooks locally
stripe listen --forward-to localhost:8080/api/webhooks/stripe

# Trigger test events
stripe trigger checkout.session.completed
stripe trigger customer.subscription.updated
```

## Alternatives Considered

### Option 1: Paddle

- Pros: Merchant of record, handles tax/VAT
- Cons: Higher fees, less customization

### Option 2: LemonSqueezy

- Pros: Simple API, good for digital products
- Cons: Smaller ecosystem, less mature

### Option 3: Self-hosted (BTCPay, etc.)

- Pros: No fees, full control
- Cons: PCI compliance burden, more complexity

## References

- [Stripe Checkout Docs](https://stripe.com/docs/payments/checkout)
- [Stripe Customer Portal](https://stripe.com/docs/billing/subscriptions/customer-portal)
- [Stripe Webhooks](https://stripe.com/docs/webhooks)
- [backend/internal/stripe/](../../backend/internal/stripe/)
