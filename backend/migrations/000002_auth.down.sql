-- Rollback auth schema
DROP FUNCTION IF EXISTS is_ip_blocked(VARCHAR);
DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP TABLE IF EXISTS login_history CASCADE;
DROP TABLE IF EXISTS ip_blocklist CASCADE;
DROP TABLE IF EXISTS user_two_factor CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS oauth_providers CASCADE;
DROP TABLE IF EXISTS token_blacklist CASCADE;
