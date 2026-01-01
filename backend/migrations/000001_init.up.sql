-- =============================================================================
-- INIT MIGRATION - Complete Database Schema
-- Consolidated from migrations 000001-000013
-- =============================================================================

-- =============================================================================
-- SECTION 1: USERS & CORE
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

    -- Security tracking / Account lockout
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_failed_login TIMESTAMPTZ,
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
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_stripe_customer_id ON users(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX idx_users_refresh_token ON users(refresh_token) WHERE refresh_token IS NOT NULL;
CREATE UNIQUE INDEX idx_users_password_reset_token ON users(password_reset_token) WHERE password_reset_token IS NOT NULL AND password_reset_token != '';
CREATE INDEX idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
CREATE INDEX idx_users_deletion ON users(deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;
CREATE INDEX idx_users_oauth_provider_id ON users(oauth_provider_id) WHERE oauth_provider_id IS NOT NULL AND oauth_provider_id != '';
CREATE INDEX idx_users_active ON users(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_failed_attempts ON users(failed_login_attempts) WHERE failed_login_attempts > 0;

COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN users.locked_until IS 'Account locked until this time (NULL if not locked)';
COMMENT ON COLUMN users.last_failed_login IS 'Timestamp of last failed login attempt';

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
CREATE INDEX idx_data_exports_cleanup ON data_exports(status, expires_at) WHERE status IN ('pending', 'failed');

-- =============================================================================
-- SECTION 2: AUTHENTICATION
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

CREATE INDEX idx_token_blacklist_hash_expires ON token_blacklist(token_hash, expires_at);
CREATE INDEX idx_token_blacklist_user_id ON token_blacklist(user_id);

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
CREATE INDEX idx_login_history_failed_ip ON login_history(ip_address, created_at DESC) WHERE success = false;

-- =============================================================================
-- SECTION 3: PAYMENTS
-- =============================================================================

-- Subscriptions table for Stripe billing
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_price_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMPTZ
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at);
CREATE INDEX idx_subscriptions_status_user ON subscriptions(user_id, status);

-- =============================================================================
-- SECTION 4: ORGANIZATIONS
-- =============================================================================

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url VARCHAR(500),
    plan VARCHAR(50) DEFAULT 'free',
    stripe_customer_id VARCHAR(255),
    settings JSONB DEFAULT '{}',
    created_by_user_id INTEGER NOT NULL REFERENCES users(id)
);

CREATE UNIQUE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_organizations_stripe_customer ON organizations(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id != '';
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX idx_organizations_plan ON organizations(plan);
CREATE INDEX idx_organizations_created_by ON organizations(created_by_user_id);

-- Organization members table
CREATE TABLE IF NOT EXISTS organization_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    invited_by_user_id INTEGER REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    accepted_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX idx_org_members_org_id ON organization_members(organization_id);
CREATE INDEX idx_org_members_status ON organization_members(status);
CREATE INDEX idx_org_members_role ON organization_members(role);
CREATE INDEX idx_org_members_user_status ON organization_members(user_id, status);
CREATE INDEX idx_organization_members_invited_by ON organization_members(invited_by_user_id);
CREATE INDEX idx_org_members_org_role ON organization_members(organization_id, role);

-- Organization invitations table
CREATE TABLE IF NOT EXISTS organization_invitations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    token VARCHAR(64) NOT NULL,
    invited_by_user_id INTEGER NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    UNIQUE(organization_id, email)
);

CREATE UNIQUE INDEX idx_org_invitations_token ON organization_invitations(token);
CREATE INDEX idx_org_invitations_email ON organization_invitations(email);
CREATE INDEX idx_org_invitations_expires ON organization_invitations(expires_at);
CREATE INDEX idx_org_invitations_org_id ON organization_invitations(organization_id);
CREATE INDEX idx_organization_invitations_invited_by ON organization_invitations(invited_by_user_id);

-- =============================================================================
-- SECTION 5: ADMIN
-- =============================================================================

-- Audit logs table for tracking user actions
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id INTEGER,
    action VARCHAR(50) NOT NULL,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_target_type ON audit_logs(target_type);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action, created_at DESC);

-- Feature flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT false,
    rollout_percentage INTEGER NOT NULL DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    allowed_roles TEXT[],
    metadata JSONB,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_feature_flags_enabled ON feature_flags(enabled) WHERE enabled = true;
CREATE INDEX idx_feature_flags_key_active ON feature_flags(key) WHERE deleted_at IS NULL;

-- User feature flag overrides
CREATE TABLE IF NOT EXISTS user_feature_flags (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_flag_id INTEGER NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_flag_id)
);

CREATE INDEX idx_user_feature_flags_user_id ON user_feature_flags(user_id);
CREATE INDEX idx_user_feature_flags_flag_id ON user_feature_flags(feature_flag_id);

-- System settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value JSONB NOT NULL DEFAULT '{}',
    category VARCHAR(50) NOT NULL,
    description TEXT,
    is_sensitive BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_settings_key ON system_settings(key);
CREATE INDEX idx_system_settings_category ON system_settings(category);

-- Insert default settings
INSERT INTO system_settings (key, value, category, description, is_sensitive) VALUES
-- Email settings
('smtp_host', '"localhost"', 'email', 'SMTP server hostname', false),
('smtp_port', '587', 'email', 'SMTP server port', false),
('smtp_user', '""', 'email', 'SMTP username', false),
('smtp_password', '""', 'email', 'SMTP password', true),
('smtp_from_email', '"noreply@example.com"', 'email', 'From email address', false),
('smtp_from_name', '"My App"', 'email', 'From name', false),
('smtp_enabled', 'false', 'email', 'Enable email sending', false),
-- Site settings
('site_name', '"My Application"', 'site', 'Site name displayed in UI', false),
('site_logo_url', '""', 'site', 'URL to site logo', false),
('maintenance_mode', 'false', 'site', 'Enable maintenance mode', false),
('maintenance_message', '"We are performing scheduled maintenance. Please check back soon."', 'site', 'Maintenance mode message', false),
-- Security settings
('password_min_length', '8', 'security', 'Minimum password length', false),
('password_require_uppercase', 'true', 'security', 'Require uppercase letter in password', false),
('password_require_lowercase', 'true', 'security', 'Require lowercase letter in password', false),
('password_require_number', 'true', 'security', 'Require number in password', false),
('password_require_special', 'false', 'security', 'Require special character in password', false),
('session_timeout_minutes', '10080', 'security', 'Session timeout in minutes (default 7 days)', false),
('max_login_attempts', '5', 'security', 'Maximum failed login attempts before lockout', false),
('lockout_duration_minutes', '15', 'security', 'Account lockout duration in minutes', false),
('require_2fa_for_admins', 'false', 'security', 'Require 2FA for admin users', false),
-- Rate limiting settings
('rate_limit_login_per_minute', '10', 'ratelimit', 'Login attempts per minute', false),
('rate_limit_register_per_minute', '5', 'ratelimit', 'Registration attempts per minute', false),
('rate_limit_api_per_minute', '100', 'ratelimit', 'API requests per minute', false)
ON CONFLICT (key) DO NOTHING;

-- Announcement banners table (includes enhancements from 000013)
CREATE TABLE IF NOT EXISTS announcement_banners (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'info' CHECK (type IN ('info', 'warning', 'error', 'success', 'maintenance')),
    link_url VARCHAR(500),
    link_text VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    is_dismissible BOOLEAN DEFAULT true,
    show_on_pages JSONB DEFAULT '["*"]',
    target_roles TEXT[] DEFAULT NULL,
    priority INTEGER DEFAULT 0,
    starts_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ends_at TIMESTAMPTZ,
    view_count INTEGER DEFAULT 0,
    dismiss_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    -- Enhancement fields
    display_type VARCHAR(20) DEFAULT 'banner' CHECK (display_type IN ('banner', 'modal')),
    category VARCHAR(20) DEFAULT 'update' CHECK (category IN ('update', 'feature', 'bugfix')),
    email_sent BOOLEAN DEFAULT false,
    email_sent_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT announcement_valid_dates CHECK (ends_at IS NULL OR starts_at <= ends_at)
);

CREATE INDEX idx_announcement_banners_active ON announcement_banners(is_active) WHERE is_active = true;
CREATE INDEX idx_announcement_banners_dates ON announcement_banners(starts_at, ends_at);
CREATE INDEX idx_announcement_banners_priority ON announcement_banners(priority DESC);
CREATE INDEX idx_announcement_banners_published ON announcement_banners(published_at DESC) WHERE is_active = true;
CREATE INDEX idx_announcement_banners_category ON announcement_banners(category);
CREATE INDEX idx_announcement_banners_display_type ON announcement_banners(display_type);
CREATE INDEX idx_announcement_banners_active_dates ON announcement_banners(is_active, starts_at, ends_at) WHERE is_active = true;

-- Track dismissed announcements per user
CREATE TABLE IF NOT EXISTS user_dismissed_announcements (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id INTEGER NOT NULL REFERENCES announcement_banners(id) ON DELETE CASCADE,
    dismissed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, announcement_id)
);

-- Track modal reads (separate from dismissals for modals that must be acknowledged)
CREATE TABLE IF NOT EXISTS user_announcement_reads (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id INTEGER NOT NULL REFERENCES announcement_banners(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, announcement_id)
);

CREATE INDEX idx_user_announcement_reads_user ON user_announcement_reads(user_id);

-- Email templates table
CREATE TABLE IF NOT EXISTS email_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    available_variables JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    is_system BOOLEAN DEFAULT false,
    last_sent_at TIMESTAMPTZ,
    send_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_templates_key ON email_templates(key);
CREATE INDEX idx_email_templates_active ON email_templates(is_active) WHERE is_active = true;

-- Insert email templates with final styled versions
INSERT INTO email_templates (key, name, description, subject, body_html, body_text, available_variables, is_system) VALUES
('welcome', 'Welcome Email', 'Sent to new users after registration', 'Welcome to {{site_name}}!',
'<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding-bottom: 24px;">
            <div style="width: 64px; height: 64px; background-color: #dbeafe; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">&#127881;</span>
            </div>
        </td>
    </tr>
</table>

<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3; text-align: center;">
    Welcome to {{site_name}}!
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Thank you for joining {{site_name}}! We''re excited to have you on board and can''t wait for you to explore everything we have to offer.
</p>

<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Get started by logging into your account and setting up your profile.
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{login_url}}" style="background-color: #2563eb; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Get Started
            </a>
        </td>
    </tr>
</table>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    If you have any questions, our support team is here to help.
</p>

<p style="margin: 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Best regards,<br>
    <strong>The {{site_name}} Team</strong>
</p>',
'Hi {{user_name}},

Welcome to {{site_name}}! We''re excited to have you on board.

Get started by logging in: {{login_url}}

If you have any questions, our support team is here to help.

Best regards,
The {{site_name}} Team',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "site_name", "description": "Site name"}, {"name": "login_url", "description": "Login page URL"}]', true),

('email_verification', 'Email Verification', 'Sent to verify user email address', 'Verify your email address',
'<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding-bottom: 24px;">
            <div style="width: 64px; height: 64px; background-color: #dbeafe; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">&#9993;</span>
            </div>
        </td>
    </tr>
</table>

<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3; text-align: center;">
    Verify Your Email
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Thanks for signing up! To complete your registration and unlock all features, please verify your email address by clicking the button below:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{verification_url}}" style="background-color: #2563eb; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Verify Email Address
            </a>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 0 0 24px 0;">
    <tr>
        <td style="background-color: #f3f4f6; border-radius: 8px; padding: 16px;">
            <p style="margin: 0 0 8px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">
                Or copy this link:
            </p>
            <p style="margin: 0; font-size: 14px; color: #374151; word-break: break-all; font-family: monospace;">
                {{verification_url}}
            </p>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #dbeafe; border-left: 4px solid #2563eb; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #1e40af;">
                <strong>This link will expire in {{expiry_hours}} hours.</strong> If you need a new verification link, you can request one from your account settings.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0; font-size: 14px; color: #6b7280;">
    If you didn''t create an account, you can safely ignore this email.
</p>',
'Hi {{user_name}},

Thanks for signing up! Please verify your email address:

{{verification_url}}

This link will expire in {{expiry_hours}} hours.

If you didn''t create an account, you can safely ignore this email.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "verification_url", "description": "Email verification URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]', true),

('password_reset', 'Password Reset', 'Sent when user requests password reset', 'Reset your password',
'<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3;">
    Reset Your Password
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    We received a request to reset your password. Click the button below to create a new password:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{reset_url}}" style="background-color: #2563eb; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Reset Password
            </a>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 0 0 24px 0;">
    <tr>
        <td style="background-color: #f3f4f6; border-radius: 8px; padding: 16px;">
            <p style="margin: 0 0 8px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">
                Or copy this link:
            </p>
            <p style="margin: 0; font-size: 14px; color: #374151; word-break: break-all; font-family: monospace;">
                {{reset_url}}
            </p>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #92400e;">
                <strong>This link will expire in {{expiry_hours}} hours.</strong> For security reasons, please complete your password reset promptly.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0; font-size: 14px; color: #6b7280;">
    If you didn''t request a password reset, you can safely ignore this email. Your password will remain unchanged.
</p>',
'Hi {{user_name}},

We received a request to reset your password. Click the link below:

{{reset_url}}

This link will expire in {{expiry_hours}} hours.

If you didn''t request this, you can ignore this email.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "reset_url", "description": "Password reset URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]', true),

('password_changed', 'Password Changed Notification', 'Sent when password is successfully changed', 'Your password has been changed',
'<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding-bottom: 24px;">
            <div style="width: 64px; height: 64px; background-color: #dcfce7; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">&#10003;</span>
            </div>
        </td>
    </tr>
</table>

<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3; text-align: center;">
    Password Changed Successfully
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Your account password was successfully changed on {{change_date}}.
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #dcfce7; border-left: 4px solid #22c55e; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #166534;">
                Your password has been updated. You can now use your new password to sign in.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151; font-weight: 600;">
    Didn''t make this change?
</p>

<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    If you didn''t change your password, your account may have been compromised. Please take action immediately:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{reset_url}}" style="background-color: #dc2626; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Reset Password Now
            </a>
        </td>
    </tr>
</table>

<p style="margin: 0; font-size: 14px; color: #6b7280;">
    We also recommend enabling two-factor authentication for added security.
</p>',
'Hi {{user_name}},

Your password was successfully changed on {{change_date}}.

If you didn''t make this change, reset your password immediately: {{reset_url}}

We recommend enabling two-factor authentication for added security.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "change_date", "description": "Date/time of change"}, {"name": "reset_url", "description": "Password reset URL"}]', true),

('login_alert', 'New Login Alert', 'Sent when login from new device/location detected', 'New login to your account',
'<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding-bottom: 24px;">
            <div style="width: 64px; height: 64px; background-color: #fef3c7; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">&#128187;</span>
            </div>
        </td>
    </tr>
</table>

<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3; text-align: center;">
    New Sign-in Detected
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    We detected a new sign-in to your account from an unrecognized device or location:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 0 0 24px 0;">
    <tr>
        <td style="background-color: #f3f4f6; border-radius: 8px; padding: 16px;">
            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
                <tr>
                    <td style="padding: 4px 0; font-size: 14px; color: #6b7280; width: 100px;">Time:</td>
                    <td style="padding: 4px 0; font-size: 14px; color: #374151; font-weight: 500;">{{login_time}}</td>
                </tr>
                <tr>
                    <td style="padding: 4px 0; font-size: 14px; color: #6b7280;">Location:</td>
                    <td style="padding: 4px 0; font-size: 14px; color: #374151; font-weight: 500;">{{location}}</td>
                </tr>
                <tr>
                    <td style="padding: 4px 0; font-size: 14px; color: #6b7280;">Device:</td>
                    <td style="padding: 4px 0; font-size: 14px; color: #374151; font-weight: 500;">{{device_info}}</td>
                </tr>
                <tr>
                    <td style="padding: 4px 0; font-size: 14px; color: #6b7280;">IP Address:</td>
                    <td style="padding: 4px 0; font-size: 14px; color: #374151; font-weight: 500;">{{ip_address}}</td>
                </tr>
            </table>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #dbeafe; border-left: 4px solid #2563eb; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #1e40af;">
                If this was you, no action is needed. You can safely ignore this email.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151; font-weight: 600;">
    Wasn''t you?
</p>

<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    If you don''t recognize this activity, your account may be compromised. Please secure your account immediately:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{security_url}}" style="background-color: #dc2626; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Secure My Account
            </a>
        </td>
    </tr>
</table>',
'Hi {{user_name}},

We detected a new sign-in to your account:

Time: {{login_time}}
Location: {{location}}
Device: {{device_info}}
IP: {{ip_address}}

If this was you, no action needed.

If this wasn''t you, secure your account: {{security_url}}',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "login_time", "description": "Login timestamp"}, {"name": "location", "description": "Login location"}, {"name": "device_info", "description": "Device information"}, {"name": "security_url", "description": "Security settings URL"}, {"name": "ip_address", "description": "IP address"}]', true),

('2fa_enabled', '2FA Enabled Confirmation', 'Sent when two-factor authentication is enabled', 'Two-factor authentication enabled',
'<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding-bottom: 24px;">
            <div style="width: 64px; height: 64px; background-color: #dcfce7; border-radius: 50%; display: inline-flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">&#128274;</span>
            </div>
        </td>
    </tr>
</table>

<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3; text-align: center;">
    Two-Factor Authentication Enabled
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Great news! Two-factor authentication has been successfully enabled on your account. Your account is now more secure.
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #dcfce7; border-left: 4px solid #22c55e; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #166534;">
                <strong>Security upgrade complete!</strong> You''ll now need to enter a verification code from your authenticator app when signing in.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151; font-weight: 600;">
    Important: Save Your Backup Codes
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Make sure you''ve saved your backup codes in a secure location. You''ll need them if you ever lose access to your authenticator app.
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #92400e;">
                <strong>If you didn''t enable 2FA,</strong> someone may have access to your account. Change your password immediately and contact support.
            </p>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{security_url}}" style="background-color: #2563eb; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Review Security Settings
            </a>
        </td>
    </tr>
</table>',
'Hi {{user_name}},

Two-factor authentication has been successfully enabled on your account.

IMPORTANT: Make sure you''ve saved your backup codes in a secure location.

If you didn''t enable 2FA, change your password immediately: {{security_url}}',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "security_url", "description": "Security settings URL"}]', true),

('email_change_verify', 'Email Change Verification', 'Sent to verify a new email address when user changes their email',
'Confirm your new email address',
'<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3;">
    Confirm Your New Email
</h1>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    Hi {{user_name}},
</p>

<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #374151;">
    You requested to change your account email address to this email. Click the button below to confirm this change:
</p>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
    <tr>
        <td align="center" style="padding: 8px 0 24px 0;">
            <a href="{{verification_url}}" style="background-color: #2563eb; border-radius: 6px; color: #ffffff; display: inline-block; font-size: 16px; font-weight: 600; padding: 14px 32px; text-decoration: none;">
                Confirm Email Change
            </a>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 0 0 24px 0;">
    <tr>
        <td style="background-color: #f3f4f6; border-radius: 8px; padding: 16px;">
            <p style="margin: 0 0 8px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">
                Or copy this link:
            </p>
            <p style="margin: 0; font-size: 14px; color: #374151; word-break: break-all; font-family: monospace;">
                {{verification_url}}
            </p>
        </td>
    </tr>
</table>

<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
    <tr>
        <td style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; border-radius: 0 8px 8px 0;">
            <p style="margin: 0; font-size: 14px; color: #92400e;">
                <strong>This link will expire in {{expiry_hours}} hours.</strong> After confirmation, your old email address will no longer be associated with your account.
            </p>
        </td>
    </tr>
</table>

<p style="margin: 0; font-size: 14px; color: #6b7280;">
    If you didn''t request this change, please secure your account immediately by changing your password and contacting support.
</p>',
'Hi {{user_name}},

You requested to change your email address. Confirm here: {{verification_url}}

This link expires in {{expiry_hours}} hours.

If you didn''t request this, secure your account immediately.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "verification_url", "description": "Email verification URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]', true),

('two_factor_code', 'Two-Factor Authentication Code', 'Sent when user needs to enter a 2FA code during login',
'Your verification code: {{code}}',
'<div style="text-align: center;">
  <div style="width: 64px; height: 64px; background-color: #dbeafe; border-radius: 50%; margin: 0 auto 24px; display: flex; align-items: center; justify-content: center;">
    <span style="font-size: 32px;">&#128274;</span>
  </div>
  <h1>Your Verification Code</h1>
</div>
<p>Hi {{user_name}},</p>
<p>Use the following code to complete your sign-in:</p>
<div style="text-align: center; margin: 24px 0;">
  <div style="background-color: #f5f5f5; padding: 24px 32px; border-radius: 8px; display: inline-block;">
    <span style="font-size: 36px; font-weight: 700; letter-spacing: 8px; font-family: monospace;">{{code}}</span>
  </div>
</div>
<p style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 12px; margin: 16px 0; font-size: 14px;">
  <strong>This code expires in {{expiry_minutes}} minutes.</strong> Never share this code with anyone.
</p>
<p style="font-size: 14px; color: #6b7280;">If you didn''t try to sign in, someone may be trying to access your account. Change your password immediately.</p>',
'Hi {{user_name}},

Your verification code is: {{code}}

This code expires in {{expiry_minutes}} minutes. Never share this code.

Didn''t request this? Change your password immediately.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "code", "description": "6-digit verification code"}, {"name": "expiry_minutes", "description": "Minutes until code expires"}]', true),

('account_locked', 'Account Locked Notification', 'Sent when account is locked due to failed login attempts',
'Your account has been temporarily locked',
'<div style="text-align: center;">
  <div style="width: 64px; height: 64px; background-color: #fef2f2; border-radius: 50%; margin: 0 auto 24px; display: flex; align-items: center; justify-content: center;">
    <span style="font-size: 32px;">&#128274;</span>
  </div>
  <h1 style="color: #dc2626;">Account Temporarily Locked</h1>
</div>
<p>Hi {{user_name}},</p>
<p>Your account has been temporarily locked due to multiple failed login attempts. This is a security measure to protect your account.</p>
<p style="background-color: #fef2f2; border-left: 4px solid #dc2626; padding: 12px; margin: 16px 0; font-size: 14px; color: #991b1b;">
  <strong>Your account is locked for {{lock_duration}}.</strong> After this period, you can try logging in again.
</p>
<p><strong>What happened?</strong></p>
<p>We detected {{failed_attempts}} failed login attempts. This could mean you forgot your password, or someone is trying to access your account.</p>
<p><strong>What should you do?</strong></p>
<p>You can reset your password to regain access immediately:</p>
<p style="text-align: center; margin: 24px 0;">
  <a href="{{reset_url}}" style="background-color: #2563eb; color: #ffffff; padding: 14px 32px; border-radius: 6px; text-decoration: none; font-weight: 600;">Reset Password</a>
</p>
<p style="font-size: 14px; color: #6b7280;">If you weren''t trying to log in, we recommend resetting your password and enabling two-factor authentication.</p>',
'Hi {{user_name}},

Your account has been locked for {{lock_duration}} due to {{failed_attempts}} failed login attempts.

Reset your password: {{reset_url}}

If this wasn''t you, secure your account immediately.',
'[{"name": "user_name", "description": "User''s full name"}, {"name": "lock_duration", "description": "How long account is locked"}, {"name": "failed_attempts", "description": "Number of failed attempts"}, {"name": "reset_url", "description": "Password reset URL"}]', true)

ON CONFLICT (key) DO NOTHING;

-- =============================================================================
-- SECTION 6: USER API KEYS
-- =============================================================================

-- User API keys table
CREATE TABLE IF NOT EXISTS user_api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(64) NOT NULL,
    key_encrypted TEXT NOT NULL,
    key_preview VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMPTZ,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider)
);

CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);
CREATE INDEX idx_user_api_keys_provider ON user_api_keys(provider);
CREATE INDEX idx_user_api_keys_user_provider ON user_api_keys(user_id, provider) WHERE is_active = true;

-- =============================================================================
-- SECTION 7: IDEMPOTENCY
-- =============================================================================

-- Idempotency keys table for safe POST request retries
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    request_hash VARCHAR(64) NOT NULL,
    response_status INT NOT NULL,
    response_body JSONB,
    response_headers JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT unique_idempotency_key UNIQUE (key, user_id)
);

CREATE INDEX idx_idempotency_keys_lookup ON idempotency_keys(key, user_id);
CREATE INDEX idx_idempotency_keys_expires ON idempotency_keys(expires_at);

COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys and cached responses for safe request retries';
COMMENT ON COLUMN idempotency_keys.key IS 'Client-provided idempotency key (UUID recommended)';
COMMENT ON COLUMN idempotency_keys.expires_at IS 'Keys expire after 24 hours by default';
