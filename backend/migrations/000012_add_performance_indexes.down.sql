-- Rollback: Remove performance indexes

DROP INDEX IF EXISTS idx_files_user_created;
DROP INDEX IF EXISTS idx_subscriptions_status_user;
DROP INDEX IF EXISTS idx_feature_flags_enabled;
DROP INDEX IF EXISTS idx_audit_logs_user_action;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_org_members_org_role;
DROP INDEX IF EXISTS idx_users_oauth_provider_id;
