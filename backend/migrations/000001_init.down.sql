-- =============================================================================
-- ROLLBACK INIT MIGRATION - Drop all tables in reverse dependency order
-- =============================================================================

-- Drop functions first
DROP FUNCTION IF EXISTS is_ip_blocked(VARCHAR);
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Section 7: Idempotency
DROP TABLE IF EXISTS idempotency_keys CASCADE;

-- Section 6: User API Keys
DROP TABLE IF EXISTS user_api_keys CASCADE;

-- Section 5: Admin (reverse order)
DROP TABLE IF EXISTS email_templates CASCADE;
DROP TABLE IF EXISTS user_announcement_reads CASCADE;
DROP TABLE IF EXISTS user_dismissed_announcements CASCADE;
DROP TABLE IF EXISTS announcement_banners CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;
DROP TABLE IF EXISTS user_feature_flags CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;

-- Section 4: Organizations (reverse order)
DROP TABLE IF EXISTS organization_invitations CASCADE;
DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- Section 3: Payments
DROP TABLE IF EXISTS subscriptions CASCADE;

-- Section 2: Authentication (reverse order)
DROP TABLE IF EXISTS login_history CASCADE;
DROP TABLE IF EXISTS ip_blocklist CASCADE;
DROP TABLE IF EXISTS user_two_factor CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS oauth_providers CASCADE;
DROP TABLE IF EXISTS token_blacklist CASCADE;

-- Section 1: Core (reverse order)
DROP TABLE IF EXISTS data_exports CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS users CASCADE;
