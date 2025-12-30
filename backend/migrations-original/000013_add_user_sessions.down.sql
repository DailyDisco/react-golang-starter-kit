DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP INDEX IF EXISTS idx_user_sessions_last_active;
DROP INDEX IF EXISTS idx_user_sessions_expires;
DROP INDEX IF EXISTS idx_user_sessions_token_hash;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP TABLE IF EXISTS user_sessions;
