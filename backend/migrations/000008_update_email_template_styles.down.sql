-- =============================================================================
-- Revert email templates to basic styling (rollback)
-- Note: This restores simplified versions - exact originals may differ
-- =============================================================================

UPDATE email_templates SET
  body_html = '<h1>Welcome, {{user_name}}!</h1><p>Thank you for joining {{site_name}}. We''re excited to have you on board.</p><p>Get started by <a href="{{login_url}}">logging into your account</a>.</p><p>Best regards,<br>The {{site_name}} Team</p>',
  body_text = 'Hi {{user_name}}, Welcome to {{site_name}}! Get started: {{login_url}}',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'welcome';

UPDATE email_templates SET
  body_html = '<h1>Verify Your Email</h1><p>Hi {{user_name}},</p><p>Please click the link below to verify your email address:</p><p><a href="{{verification_url}}">Verify Email</a></p><p>This link will expire in {{expiry_hours}} hours.</p>',
  body_text = 'Hi {{user_name}}, Verify your email: {{verification_url}}. Expires in {{expiry_hours}} hours.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'email_verification';

UPDATE email_templates SET
  body_html = '<h1>Password Reset Request</h1><p>Hi {{user_name}},</p><p>Click the link below to set a new password:</p><p><a href="{{reset_url}}">Reset Password</a></p><p>This link will expire in {{expiry_hours}} hours.</p>',
  body_text = 'Hi {{user_name}}, Reset your password: {{reset_url}}. Expires in {{expiry_hours}} hours.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'password_reset';

UPDATE email_templates SET
  body_html = '<h1>Password Changed</h1><p>Hi {{user_name}},</p><p>Your password was changed on {{change_date}}.</p><p>If you did not make this change, <a href="{{reset_url}}">reset your password immediately</a>.</p>',
  body_text = 'Hi {{user_name}}, Your password was changed on {{change_date}}. If not you, reset it: {{reset_url}}',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'password_changed';

UPDATE email_templates SET
  body_html = '<h1>New Login Detected</h1><p>Hi {{user_name}},</p><p>New login:</p><ul><li>Time: {{login_time}}</li><li>Location: {{location}}</li><li>Device: {{device_info}}</li></ul><p>Not you? <a href="{{security_url}}">Secure your account</a>.</p>',
  body_text = 'Hi {{user_name}}, New login at {{login_time}} from {{location}} ({{device_info}}). Not you? {{security_url}}',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'login_alert';

UPDATE email_templates SET
  body_html = '<h1>2FA Enabled</h1><p>Hi {{user_name}},</p><p>Two-factor authentication has been successfully enabled on your account.</p><p>Store your backup codes safely.</p>',
  body_text = 'Hi {{user_name}}, 2FA has been enabled. Store your backup codes safely.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = '2fa_enabled';

UPDATE email_templates SET
  body_html = '<h1>Confirm Your New Email</h1><p>Hi {{user_name}},</p><p>You requested to change your account email address. Click below to confirm:</p><p><a href="{{verification_url}}">Confirm Email Change</a></p><p>Expires in {{expiry_hours}} hours.</p>',
  body_text = 'Hi {{user_name}}, Confirm email change: {{verification_url}}. Expires in {{expiry_hours}} hours.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'email_change_verify';
