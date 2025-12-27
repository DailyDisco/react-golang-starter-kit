-- Login history for security auditing and user transparency
CREATE TABLE IF NOT EXISTS login_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100), -- 'invalid_password', 'account_locked', '2fa_failed', 'account_inactive'
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}', -- {browser, browser_version, os, os_version, device_type}
    location JSONB DEFAULT '{}', -- {country, country_code, city, region}
    auth_method VARCHAR(20) DEFAULT 'password' CHECK (auth_method IN ('password', 'oauth_google', 'oauth_github', 'refresh_token', '2fa')),
    session_id INTEGER REFERENCES user_sessions(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_login_history_user_id ON login_history(user_id);
CREATE INDEX idx_login_history_created ON login_history(created_at DESC);
CREATE INDEX idx_login_history_user_time ON login_history(user_id, created_at DESC);
CREATE INDEX idx_login_history_ip ON login_history(ip_address);
CREATE INDEX idx_login_history_failed ON login_history(user_id, success) WHERE success = false;

-- Partition by month for large-scale deployments (optional, can be enabled later)
-- This keeps the table manageable as login history grows
