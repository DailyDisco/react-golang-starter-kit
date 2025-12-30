-- Drop subscriptions table and its indexes
DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE IF EXISTS subscriptions;

-- Remove Stripe customer ID from users table
DROP INDEX IF EXISTS idx_users_stripe_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
