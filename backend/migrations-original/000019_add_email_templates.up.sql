-- Email templates for customizable transactional emails
CREATE TABLE IF NOT EXISTS email_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) NOT NULL UNIQUE, -- e.g., 'welcome', 'password_reset', 'verification'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT, -- Plain text fallback
    available_variables JSONB DEFAULT '[]', -- [{name: "user_name", description: "User's full name"}]
    is_active BOOLEAN DEFAULT true,
    is_system BOOLEAN DEFAULT false, -- System templates cannot be deleted
    last_sent_at TIMESTAMP,
    send_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_templates_key ON email_templates(key);
CREATE INDEX idx_email_templates_active ON email_templates(is_active) WHERE is_active = true;

-- Insert default email templates
INSERT INTO email_templates (key, name, description, subject, body_html, body_text, available_variables, is_system) VALUES
(
    'welcome',
    'Welcome Email',
    'Sent to new users after registration',
    'Welcome to {{site_name}}!',
    '<h1>Welcome, {{user_name}}!</h1><p>Thank you for joining {{site_name}}. We''re excited to have you on board.</p><p>Get started by <a href="{{login_url}}">logging into your account</a>.</p><p>Best regards,<br>The {{site_name}} Team</p>',
    'Welcome, {{user_name}}!\n\nThank you for joining {{site_name}}. We''re excited to have you on board.\n\nGet started by logging into your account: {{login_url}}\n\nBest regards,\nThe {{site_name}} Team',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "site_name", "description": "Site name"}, {"name": "login_url", "description": "Login page URL"}]',
    true
),
(
    'email_verification',
    'Email Verification',
    'Sent to verify user email address',
    'Verify your email address',
    '<h1>Verify Your Email</h1><p>Hi {{user_name}},</p><p>Please click the link below to verify your email address:</p><p><a href="{{verification_url}}">Verify Email</a></p><p>This link will expire in {{expiry_hours}} hours.</p><p>If you didn''t create an account, you can safely ignore this email.</p>',
    'Hi {{user_name}},\n\nPlease click the link below to verify your email address:\n\n{{verification_url}}\n\nThis link will expire in {{expiry_hours}} hours.\n\nIf you didn''t create an account, you can safely ignore this email.',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "verification_url", "description": "Email verification URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]',
    true
),
(
    'password_reset',
    'Password Reset',
    'Sent when user requests password reset',
    'Reset your password',
    '<h1>Password Reset Request</h1><p>Hi {{user_name}},</p><p>We received a request to reset your password. Click the link below to set a new password:</p><p><a href="{{reset_url}}">Reset Password</a></p><p>This link will expire in {{expiry_hours}} hours.</p><p>If you didn''t request this, you can safely ignore this email. Your password will remain unchanged.</p>',
    'Hi {{user_name}},\n\nWe received a request to reset your password. Click the link below to set a new password:\n\n{{reset_url}}\n\nThis link will expire in {{expiry_hours}} hours.\n\nIf you didn''t request this, you can safely ignore this email.',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "reset_url", "description": "Password reset URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]',
    true
),
(
    'password_changed',
    'Password Changed Notification',
    'Sent when password is successfully changed',
    'Your password has been changed',
    '<h1>Password Changed</h1><p>Hi {{user_name}},</p><p>Your password was successfully changed on {{change_date}}.</p><p>If you did not make this change, please <a href="{{reset_url}}">reset your password immediately</a> and contact support.</p><p>Details:</p><ul><li>IP Address: {{ip_address}}</li><li>Device: {{device_info}}</li></ul>',
    'Hi {{user_name}},\n\nYour password was successfully changed on {{change_date}}.\n\nIf you did not make this change, please reset your password immediately: {{reset_url}}\n\nDetails:\n- IP Address: {{ip_address}}\n- Device: {{device_info}}',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "change_date", "description": "Date/time of change"}, {"name": "reset_url", "description": "Password reset URL"}, {"name": "ip_address", "description": "IP address"}, {"name": "device_info", "description": "Device information"}]',
    true
),
(
    'login_alert',
    'New Login Alert',
    'Sent when login from new device/location detected',
    'New login to your account',
    '<h1>New Login Detected</h1><p>Hi {{user_name}},</p><p>We noticed a new login to your account:</p><ul><li>Time: {{login_time}}</li><li>Location: {{location}}</li><li>Device: {{device_info}}</li><li>IP Address: {{ip_address}}</li></ul><p>If this was you, no action is needed.</p><p>If you don''t recognize this activity, please <a href="{{security_url}}">secure your account</a> immediately.</p>',
    'Hi {{user_name}},\n\nWe noticed a new login to your account:\n\n- Time: {{login_time}}\n- Location: {{location}}\n- Device: {{device_info}}\n- IP Address: {{ip_address}}\n\nIf this was you, no action is needed.\n\nIf you don''t recognize this activity, please secure your account: {{security_url}}',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "login_time", "description": "Login timestamp"}, {"name": "location", "description": "Login location"}, {"name": "device_info", "description": "Device information"}, {"name": "ip_address", "description": "IP address"}, {"name": "security_url", "description": "Security settings URL"}]',
    true
),
(
    '2fa_enabled',
    '2FA Enabled Confirmation',
    'Sent when two-factor authentication is enabled',
    'Two-factor authentication enabled',
    '<h1>2FA Enabled</h1><p>Hi {{user_name}},</p><p>Two-factor authentication has been successfully enabled on your account.</p><p>Your account is now more secure. You will need to enter a verification code from your authenticator app each time you log in.</p><p>Make sure to store your backup codes in a safe place.</p>',
    'Hi {{user_name}},\n\nTwo-factor authentication has been successfully enabled on your account.\n\nYour account is now more secure. You will need to enter a verification code from your authenticator app each time you log in.\n\nMake sure to store your backup codes in a safe place.',
    '[{"name": "user_name", "description": "User''s full name"}]',
    true
),
(
    'account_deletion_requested',
    'Account Deletion Requested',
    'Sent when user requests account deletion',
    'Account deletion request received',
    '<h1>Account Deletion Requested</h1><p>Hi {{user_name}},</p><p>We received your request to delete your account. Your account and all associated data will be permanently deleted on {{deletion_date}}.</p><p>If you change your mind, you can cancel this request by logging in before the deletion date.</p><p><a href="{{cancel_url}}">Cancel Deletion Request</a></p>',
    'Hi {{user_name}},\n\nWe received your request to delete your account. Your account and all associated data will be permanently deleted on {{deletion_date}}.\n\nIf you change your mind, you can cancel this request by logging in before the deletion date.\n\nCancel here: {{cancel_url}}',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "deletion_date", "description": "Scheduled deletion date"}, {"name": "cancel_url", "description": "URL to cancel deletion"}]',
    true
),
(
    'data_export_ready',
    'Data Export Ready',
    'Sent when user data export is ready for download',
    'Your data export is ready',
    '<h1>Data Export Ready</h1><p>Hi {{user_name}},</p><p>Your data export is ready for download. Click the link below to download your data:</p><p><a href="{{download_url}}">Download Data</a></p><p>This link will expire in {{expiry_hours}} hours.</p>',
    'Hi {{user_name}},\n\nYour data export is ready for download:\n\n{{download_url}}\n\nThis link will expire in {{expiry_hours}} hours.',
    '[{"name": "user_name", "description": "User''s full name"}, {"name": "download_url", "description": "Data download URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]',
    true
);
