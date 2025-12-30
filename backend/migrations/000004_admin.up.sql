-- =============================================================================
-- ADMIN - Feature flags, Audit logs, Settings, Announcements, Email templates
-- Consolidated from: 000006, 000007, 000011, 000016, 000019
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

CREATE INDEX idx_feature_flags_key ON feature_flags(key);
CREATE INDEX idx_feature_flags_enabled ON feature_flags(enabled);
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

-- Announcement banners table
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
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_announcement_banners_active ON announcement_banners(is_active) WHERE is_active = true;
CREATE INDEX idx_announcement_banners_dates ON announcement_banners(starts_at, ends_at);
CREATE INDEX idx_announcement_banners_priority ON announcement_banners(priority DESC);

-- Track dismissed announcements per user
CREATE TABLE IF NOT EXISTS user_dismissed_announcements (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id INTEGER NOT NULL REFERENCES announcement_banners(id) ON DELETE CASCADE,
    dismissed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, announcement_id)
);

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

-- Insert default email templates
INSERT INTO email_templates (key, name, description, subject, body_html, body_text, available_variables, is_system) VALUES
('welcome', 'Welcome Email', 'Sent to new users after registration', 'Welcome to {{site_name}}!',
 '<h1>Welcome, {{user_name}}!</h1><p>Thank you for joining {{site_name}}. We''re excited to have you on board.</p><p>Get started by <a href="{{login_url}}">logging into your account</a>.</p><p>Best regards,<br>The {{site_name}} Team</p>',
 'Welcome, {{user_name}}!\n\nThank you for joining {{site_name}}.\n\nGet started: {{login_url}}',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "site_name", "description": "Site name"}, {"name": "login_url", "description": "Login page URL"}]', true),
('email_verification', 'Email Verification', 'Sent to verify user email address', 'Verify your email address',
 '<h1>Verify Your Email</h1><p>Hi {{user_name}},</p><p>Please click the link below to verify your email address:</p><p><a href="{{verification_url}}">Verify Email</a></p><p>This link will expire in {{expiry_hours}} hours.</p>',
 'Hi {{user_name}},\n\nVerify your email: {{verification_url}}\n\nExpires in {{expiry_hours}} hours.',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "verification_url", "description": "Email verification URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]', true),
('password_reset', 'Password Reset', 'Sent when user requests password reset', 'Reset your password',
 '<h1>Password Reset Request</h1><p>Hi {{user_name}},</p><p>Click the link below to set a new password:</p><p><a href="{{reset_url}}">Reset Password</a></p><p>This link will expire in {{expiry_hours}} hours.</p>',
 'Hi {{user_name}},\n\nReset your password: {{reset_url}}\n\nExpires in {{expiry_hours}} hours.',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "reset_url", "description": "Password reset URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]', true),
('password_changed', 'Password Changed Notification', 'Sent when password is successfully changed', 'Your password has been changed',
 '<h1>Password Changed</h1><p>Hi {{user_name}},</p><p>Your password was changed on {{change_date}}.</p><p>If you did not make this change, <a href="{{reset_url}}">reset your password immediately</a>.</p>',
 'Hi {{user_name}},\n\nYour password was changed on {{change_date}}.\n\nNot you? Reset: {{reset_url}}',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "change_date", "description": "Date/time of change"}, {"name": "reset_url", "description": "Password reset URL"}]', true),
('login_alert', 'New Login Alert', 'Sent when login from new device/location detected', 'New login to your account',
 '<h1>New Login Detected</h1><p>Hi {{user_name}},</p><p>New login:</p><ul><li>Time: {{login_time}}</li><li>Location: {{location}}</li><li>Device: {{device_info}}</li></ul><p>Not you? <a href="{{security_url}}">Secure your account</a>.</p>',
 'Hi {{user_name}},\n\nNew login at {{login_time}} from {{location}}.\n\nNot you? {{security_url}}',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "login_time", "description": "Login timestamp"}, {"name": "location", "description": "Login location"}, {"name": "device_info", "description": "Device information"}, {"name": "security_url", "description": "Security settings URL"}]', true),
('2fa_enabled', '2FA Enabled Confirmation', 'Sent when two-factor authentication is enabled', 'Two-factor authentication enabled',
 '<h1>2FA Enabled</h1><p>Hi {{user_name}},</p><p>Two-factor authentication has been successfully enabled on your account.</p><p>Store your backup codes safely.</p>',
 'Hi {{user_name}},\n\n2FA is now enabled on your account. Store your backup codes safely.',
 '[{"name": "user_name", "description": "User''s full name"}]', true)
ON CONFLICT (key) DO NOTHING;
