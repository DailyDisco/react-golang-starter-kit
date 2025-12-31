-- Migration: Add performance indexes identified during optimization audit
-- These indexes improve query performance for common access patterns

-- Index for OAuth provider lookups (login via OAuth)
CREATE INDEX IF NOT EXISTS idx_users_oauth_provider_id
    ON users(oauth_provider_id)
    WHERE oauth_provider_id IS NOT NULL AND oauth_provider_id != '';

-- Composite index for organization member queries (common access pattern)
CREATE INDEX IF NOT EXISTS idx_org_members_org_role
    ON organization_members(organization_id, role);

-- Partial index for active (non-deleted) users - speeds up most queries
CREATE INDEX IF NOT EXISTS idx_users_active
    ON users(id)
    WHERE deleted_at IS NULL;

-- Composite index for audit log queries (common filters)
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action
    ON audit_logs(user_id, action, created_at DESC);

-- Index for feature flag lookups by enabled status
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled
    ON feature_flags(enabled)
    WHERE enabled = true;

-- Index for subscription status queries
CREATE INDEX IF NOT EXISTS idx_subscriptions_status_user
    ON subscriptions(user_id, status);

-- Index for file ownership lookups
CREATE INDEX IF NOT EXISTS idx_files_user_created
    ON files(user_id, created_at DESC);
