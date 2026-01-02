-- =============================================================================
-- Rollback Organization Billing Integration
-- =============================================================================

-- Remove comments first
COMMENT ON COLUMN subscriptions.organization_id IS NULL;
COMMENT ON COLUMN organizations.stripe_subscription_id IS NULL;
COMMENT ON COLUMN organizations.plan_features IS NULL;

-- Drop the new indexes
DROP INDEX IF EXISTS idx_subscriptions_org_unique;
DROP INDEX IF EXISTS idx_subscriptions_user_unique;
DROP INDEX IF EXISTS idx_subscriptions_organization_id;

-- Restore unique constraint on user_id
ALTER TABLE subscriptions
ADD CONSTRAINT subscriptions_user_id_key UNIQUE (user_id);

-- Remove columns from organizations
ALTER TABLE organizations
DROP COLUMN IF EXISTS plan_features;

ALTER TABLE organizations
DROP COLUMN IF EXISTS stripe_subscription_id;

-- Remove organization_id from subscriptions
ALTER TABLE subscriptions
DROP COLUMN IF EXISTS organization_id;
