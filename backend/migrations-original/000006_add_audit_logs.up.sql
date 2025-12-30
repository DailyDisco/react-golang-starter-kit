-- Audit logs table for tracking user actions
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL, -- Actor (null for system actions)
    target_type VARCHAR(50) NOT NULL, -- "user", "subscription", "file", "settings"
    target_id INTEGER, -- ID of the affected resource
    action VARCHAR(50) NOT NULL, -- "create", "update", "delete", "login", "logout", "impersonate"
    changes JSONB, -- Before/after diff for updates
    ip_address VARCHAR(45), -- IPv4 or IPv6
    user_agent TEXT,
    metadata JSONB, -- Additional context
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common query patterns
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_target_type ON audit_logs(target_type);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);
