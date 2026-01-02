-- =============================================================================
-- Organization Billing Integration
-- Allows subscriptions to be associated with organizations (not just users)
-- =============================================================================

-- Add organization_id column to subscriptions
-- NULL = user-level subscription, non-NULL = org-level subscription
ALTER TABLE subscriptions
ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE;

-- Drop the existing unique constraint on user_id to allow org subscriptions
-- where user_id may be the billing admin but not the sole owner
ALTER TABLE subscriptions
DROP CONSTRAINT IF EXISTS subscriptions_user_id_key;

-- Create index for org subscriptions
CREATE INDEX idx_subscriptions_organization_id ON subscriptions(organization_id)
WHERE organization_id IS NOT NULL;

-- Create partial unique index for user subscriptions (user can only have one personal subscription)
CREATE UNIQUE INDEX idx_subscriptions_user_unique ON subscriptions(user_id)
WHERE organization_id IS NULL AND deleted_at IS NULL;

-- Create partial unique index for org subscriptions (org can only have one subscription)
CREATE UNIQUE INDEX idx_subscriptions_org_unique ON subscriptions(organization_id)
WHERE organization_id IS NOT NULL AND deleted_at IS NULL;

-- Add stripe_subscription_id to organizations for quick lookup
ALTER TABLE organizations
ADD COLUMN stripe_subscription_id VARCHAR(255);

-- Add plan_features JSONB to organizations for plan-specific settings (seat limits, etc.)
ALTER TABLE organizations
ADD COLUMN plan_features JSONB DEFAULT '{}';

COMMENT ON COLUMN subscriptions.organization_id IS 'NULL for user subscriptions, set for organization subscriptions';
COMMENT ON COLUMN organizations.stripe_subscription_id IS 'Stripe subscription ID for org billing';
COMMENT ON COLUMN organizations.plan_features IS 'Plan-specific features like seat_limit, storage_limit, etc.';
