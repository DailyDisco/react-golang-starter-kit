-- System settings table for admin-configurable options
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value JSONB NOT NULL DEFAULT '{}',
    category VARCHAR(50) NOT NULL,
    description TEXT,
    is_sensitive BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for system_settings
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
('rate_limit_api_per_minute', '100', 'ratelimit', 'API requests per minute', false);
