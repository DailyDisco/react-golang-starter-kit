-- Performance optimization indexes
-- These indexes improve query performance for common access patterns

-- Composite index for user login queries (email + is_active with soft delete check)
-- Speeds up: SELECT * FROM users WHERE email = ? AND deleted_at IS NULL
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_active
ON users(email, is_active) WHERE deleted_at IS NULL;

-- Composite index for token blacklist expiry cleanup and validation
-- Speeds up: SELECT * FROM token_blacklist WHERE token_hash = ? AND expires_at > NOW()
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_blacklist_expires_hash
ON token_blacklist(expires_at, token_hash);

-- Composite index for audit log queries by user and date range
-- Speeds up: SELECT * FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_created
ON audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;

-- Index for file queries by user with soft delete check
-- Speeds up: SELECT * FROM files WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_user_created
ON files(user_id, created_at DESC) WHERE deleted_at IS NULL;

-- Index for user sessions lookup by user_id
-- Speeds up: SELECT * FROM user_sessions WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_user_created
ON user_sessions(user_id, created_at DESC);

-- Index for feature flags lookup by key (if not already indexed)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_feature_flags_key
ON feature_flags(key) WHERE deleted_at IS NULL;

-- Index for login history by user
-- Speeds up: SELECT * FROM login_history WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_login_history_user_created
ON login_history(user_id, created_at DESC);

-- Partial index for active announcements
-- Speeds up: SELECT * FROM announcement_banners WHERE is_active = true AND (end_date IS NULL OR end_date > NOW())
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_announcements_active
ON announcement_banners(start_date, end_date) WHERE is_active = true;
