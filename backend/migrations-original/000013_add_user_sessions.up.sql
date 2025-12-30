-- User sessions table for active session management
-- Security: Store only SHA-256 hash of session token, never the actual token
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of refresh token
    device_info JSONB DEFAULT '{}', -- {browser, browser_version, os, os_version, device_type, device_name}
    ip_address VARCHAR(45), -- Supports both IPv4 and IPv6
    user_agent TEXT,
    location JSONB DEFAULT '{}', -- {country, country_code, city, region, latitude, longitude}
    is_current BOOLEAN DEFAULT false,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(session_token_hash);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_last_active ON user_sessions(last_active_at);

-- Cleanup function for expired sessions (can be called by cron job)
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
