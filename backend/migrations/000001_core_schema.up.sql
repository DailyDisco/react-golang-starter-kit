-- =============================================================================
-- CORE SCHEMA - Users, Files, Profile Fields, Preferences
-- Consolidated from: 000001, 000008, 000009, 000012, 000017, 000020, 000025
-- =============================================================================

-- Users table - Core user data with authentication and profile fields
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    -- Basic info
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,

    -- Email verification
    email_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verification_expires TIMESTAMPTZ,

    -- Account status
    is_active BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',

    -- Profile fields
    bio TEXT,
    location VARCHAR(255),
    avatar_url VARCHAR(500),
    social_links JSONB DEFAULT '{}',

    -- OAuth support (quick access fields)
    oauth_provider VARCHAR(50),
    oauth_provider_id VARCHAR(255),

    -- Stripe integration
    stripe_customer_id VARCHAR(255) UNIQUE,

    -- Refresh token
    refresh_token VARCHAR(64),
    refresh_token_expires TIMESTAMPTZ,

    -- Password reset
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMPTZ,

    -- 2FA fields
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    last_password_change_at TIMESTAMPTZ,

    -- Security tracking
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(45),

    -- Account deletion tracking
    deletion_requested_at TIMESTAMPTZ,
    deletion_scheduled_at TIMESTAMPTZ,
    deletion_reason TEXT,
    last_data_export_at TIMESTAMPTZ,

    -- Admin impersonation
    impersonated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    impersonation_started_at TIMESTAMPTZ
);

-- User indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_verification_token ON users(verification_token) WHERE verification_token IS NOT NULL;
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_stripe_customer_id ON users(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX idx_users_refresh_token ON users(refresh_token) WHERE refresh_token IS NOT NULL;
CREATE UNIQUE INDEX idx_users_password_reset_token ON users(password_reset_token) WHERE password_reset_token IS NOT NULL AND password_reset_token != '';
CREATE INDEX idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
CREATE INDEX idx_users_deletion ON users(deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;
CREATE INDEX idx_users_email_active ON users(email, is_active) WHERE deleted_at IS NULL;

-- Files table - File storage metadata with user ownership
CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(255),
    file_size BIGINT,
    location VARCHAR(255),
    content BYTEA,
    storage_type VARCHAR(50) DEFAULT 'database',
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_files_deleted_at ON files(deleted_at);
CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_user_created ON files(user_id, created_at DESC) WHERE deleted_at IS NULL;

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'system' CHECK (theme IN ('light', 'dark', 'system')),
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',
    time_format VARCHAR(10) DEFAULT '12h' CHECK (time_format IN ('12h', '24h')),
    email_notifications JSONB DEFAULT '{"marketing": false, "security": true, "updates": true, "weekly_digest": false}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Data exports table
CREATE TABLE IF NOT EXISTS data_exports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    download_url VARCHAR(500),
    file_path VARCHAR(500),
    file_size BIGINT,
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_data_exports_user_id ON data_exports(user_id);
CREATE INDEX idx_data_exports_status ON data_exports(status);
