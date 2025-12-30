-- =============================================================================
-- AUTH - Token blacklist, OAuth, Sessions, 2FA, Login History, IP Blocklist
-- Consolidated from: 000004, 000005, 000013, 000014, 000015, 000018
-- =============================================================================

-- Token blacklist for revoked JWT tokens
CREATE TABLE IF NOT EXISTS token_blacklist (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    reason VARCHAR(50) DEFAULT 'logout'
);

CREATE INDEX idx_token_blacklist_token_hash ON token_blacklist(token_hash);
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);
CREATE INDEX idx_token_blacklist_expires_hash ON token_blacklist(expires_at, token_hash);

-- OAuth providers table for social login
CREATE TABLE IF NOT EXISTS oauth_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id),
    UNIQUE(user_id, provider)
);

CREATE INDEX idx_oauth_providers_user_id ON oauth_providers(user_id);
CREATE INDEX idx_oauth_providers_provider_user ON oauth_providers(provider, provider_user_id);

COMMENT ON TABLE oauth_providers IS 'Stores OAuth provider connections for social login';

-- User sessions table for active session management
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token_hash VARCHAR(64) NOT NULL UNIQUE,
    device_info JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    location JSONB DEFAULT '{}',
    is_current BOOLEAN DEFAULT false,
    last_active_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(session_token_hash);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_last_active ON user_sessions(last_active_at);
CREATE INDEX idx_user_sessions_user_created ON user_sessions(user_id, created_at DESC);

-- Cleanup function for expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Two-factor authentication table
CREATE TABLE IF NOT EXISTS user_two_factor (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    encrypted_secret TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT false,
    backup_codes_hash JSONB DEFAULT '[]',
    backup_codes_remaining INTEGER DEFAULT 10,
    verified_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_two_factor_user_id ON user_two_factor(user_id);
CREATE INDEX idx_user_two_factor_enabled ON user_two_factor(is_enabled) WHERE is_enabled = true;

-- IP blocklist for security
CREATE TABLE IF NOT EXISTS ip_blocklist (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    ip_range VARCHAR(50),
    reason VARCHAR(500),
    block_type VARCHAR(20) DEFAULT 'manual' CHECK (block_type IN ('manual', 'auto_rate_limit', 'auto_brute_force', 'auto_suspicious')),
    blocked_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    hit_count INTEGER DEFAULT 0,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ip_blocklist_ip ON ip_blocklist(ip_address);
CREATE INDEX idx_ip_blocklist_active ON ip_blocklist(is_active) WHERE is_active = true;
CREATE INDEX idx_ip_blocklist_expires ON ip_blocklist(expires_at) WHERE expires_at IS NOT NULL;

-- Function to check if an IP is blocked
CREATE OR REPLACE FUNCTION is_ip_blocked(check_ip VARCHAR(45))
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM ip_blocklist
        WHERE is_active = true
        AND (expires_at IS NULL OR expires_at > NOW())
        AND (
            ip_address = check_ip
            OR (ip_range IS NOT NULL AND check_ip::inet <<= ip_range::inet)
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Login history for security auditing
CREATE TABLE IF NOT EXISTS login_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',
    location JSONB DEFAULT '{}',
    auth_method VARCHAR(20) DEFAULT 'password' CHECK (auth_method IN ('password', 'oauth_google', 'oauth_github', 'refresh_token', '2fa')),
    session_id INTEGER REFERENCES user_sessions(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_login_history_user_id ON login_history(user_id);
CREATE INDEX idx_login_history_created ON login_history(created_at DESC);
CREATE INDEX idx_login_history_user_time ON login_history(user_id, created_at DESC);
CREATE INDEX idx_login_history_ip ON login_history(ip_address);
CREATE INDEX idx_login_history_failed ON login_history(user_id, success) WHERE success = false;
