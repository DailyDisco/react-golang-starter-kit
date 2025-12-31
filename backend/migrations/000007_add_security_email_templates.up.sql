-- =============================================================================
-- Add security-related email templates
-- =============================================================================

INSERT INTO email_templates (key, name, description, subject, body_html, body_text, available_variables, is_system) VALUES
-- Email change verification
('email_change_verify', 'Email Change Verification', 'Sent to verify a new email address when user changes their email',
 'Confirm your new email address',
 '<h1>Confirm Your New Email</h1>
<p>Hi {{user_name}},</p>
<p>You requested to change your account email address to this email. Click the button below to confirm this change:</p>
<p style="text-align: center; margin: 24px 0;">
  <a href="{{verification_url}}" style="background-color: #2563eb; color: #ffffff; padding: 14px 32px; border-radius: 6px; text-decoration: none; font-weight: 600;">Confirm Email Change</a>
</p>
<p style="font-size: 14px; color: #6b7280;">Or copy this link: {{verification_url}}</p>
<p style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 12px; margin: 16px 0; font-size: 14px;">
  <strong>This link will expire in {{expiry_hours}} hours.</strong> After confirmation, your old email address will no longer be associated with your account.
</p>
<p style="font-size: 14px; color: #6b7280;">If you didn''t request this change, please secure your account immediately by changing your password.</p>',
 'Hi {{user_name}},

You requested to change your email address. Confirm here: {{verification_url}}

This link expires in {{expiry_hours}} hours.

Didn''t request this? Secure your account immediately.',
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "verification_url", "description": "Email verification URL"}, {"name": "expiry_hours", "description": "Hours until link expires"}]',
 true),

-- Two-factor authentication code
('two_factor_code', 'Two-Factor Authentication Code', 'Sent when user needs to enter a 2FA code during login',
 'Your verification code: {{code}}',
 '<div style="text-align: center;">
  <div style="width: 64px; height: 64px; background-color: #dbeafe; border-radius: 50%; margin: 0 auto 24px; display: flex; align-items: center; justify-content: center;">
    <span style="font-size: 32px;">🔐</span>
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
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "code", "description": "6-digit verification code"}, {"name": "expiry_minutes", "description": "Minutes until code expires"}]',
 true),

-- Account locked notification
('account_locked', 'Account Locked Notification', 'Sent when account is locked due to failed login attempts',
 'Your account has been temporarily locked',
 '<div style="text-align: center;">
  <div style="width: 64px; height: 64px; background-color: #fef2f2; border-radius: 50%; margin: 0 auto 24px; display: flex; align-items: center; justify-content: center;">
    <span style="font-size: 32px;">🔒</span>
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
 '[{"name": "user_name", "description": "User''s full name"}, {"name": "lock_duration", "description": "How long account is locked"}, {"name": "failed_attempts", "description": "Number of failed attempts"}, {"name": "reset_url", "description": "Password reset URL"}]',
 true)

ON CONFLICT (key) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  subject = EXCLUDED.subject,
  body_html = EXCLUDED.body_html,
  body_text = EXCLUDED.body_text,
  available_variables = EXCLUDED.available_variables,
  updated_at = CURRENT_TIMESTAMP;
