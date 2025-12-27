-- Rollback performance optimization indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_users_email_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_token_blacklist_expires_hash;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_files_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_sessions_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_feature_flags_key;
DROP INDEX CONCURRENTLY IF EXISTS idx_login_history_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_announcements_active;
