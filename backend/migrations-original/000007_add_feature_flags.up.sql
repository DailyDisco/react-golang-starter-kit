-- Feature flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE, -- "dark_mode", "new_billing_ui", etc.
    name VARCHAR(255) NOT NULL, -- Human readable name
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT false, -- Global on/off
    rollout_percentage INTEGER NOT NULL DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100), -- 0-100 percentage
    allowed_roles TEXT[], -- Roles that have access even if not in rollout (e.g., ["admin", "super_admin"])
    metadata JSONB, -- Additional configuration
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User feature flag overrides (for specific user targeting)
CREATE TABLE IF NOT EXISTS user_feature_flags (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_flag_id INTEGER NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL, -- Override value for this user
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_flag_id)
);

-- Indexes
CREATE INDEX idx_feature_flags_key ON feature_flags(key);
CREATE INDEX idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX idx_user_feature_flags_user_id ON user_feature_flags(user_id);
CREATE INDEX idx_user_feature_flags_flag_id ON user_feature_flags(feature_flag_id);

-- Add impersonation tracking to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonated_by INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonation_started_at TIMESTAMP;
