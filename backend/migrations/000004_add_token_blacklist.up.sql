-- Token blacklist for revoked JWT tokens
CREATE TABLE IF NOT EXISTS token_blacklist (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason VARCHAR(50) NOT NULL DEFAULT 'logout'
);

-- Index for fast token lookup
CREATE INDEX idx_token_blacklist_token_hash ON token_blacklist(token_hash);

-- Index for cleanup job (remove expired tokens)
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- Add refresh token fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS refresh_token VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS refresh_token_expires TIMESTAMP;

-- Index for refresh token lookup
CREATE INDEX idx_users_refresh_token ON users(refresh_token) WHERE refresh_token IS NOT NULL;
