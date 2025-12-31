-- =============================================================================
-- Update all email templates with consistent, professional styling
-- =============================================================================

-- Welcome Email
UPDATE email_templates SET
  body_html = '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
  body_text = 'Hi {{user_name}},

Welcome to {{site_name}}! We''re excited to have you on board.

Get started by logging in: {{login_url}}

If you have any questions, our support team is here to help.

Best regards,
The {{site_name}} Team',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'welcome';

-- Email Verification
UPDATE email_templates SET
  body_html = '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
  body_text = 'Hi {{user_name}},

Thanks for signing up! Please verify your email address:

{{verification_url}}

This link will expire in {{expiry_hours}} hours.

If you didn''t create an account, you can safely ignore this email.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'email_verification';

-- Password Reset
UPDATE email_templates SET
  body_html = '<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3;">
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
  body_text = 'Hi {{user_name}},

We received a request to reset your password. Click the link below:

{{reset_url}}

This link will expire in {{expiry_hours}} hours.

If you didn''t request this, you can ignore this email.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'password_reset';

-- Password Changed
UPDATE email_templates SET
  body_html = '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
  body_text = 'Hi {{user_name}},

Your password was successfully changed on {{change_date}}.

If you didn''t make this change, reset your password immediately: {{reset_url}}

We recommend enabling two-factor authentication for added security.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'password_changed';

-- Login Alert (New Device)
UPDATE email_templates SET
  body_html = '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
  body_text = 'Hi {{user_name}},

We detected a new sign-in to your account:

Time: {{login_time}}
Location: {{location}}
Device: {{device_info}}
IP: {{ip_address}}

If this was you, no action needed.

If this wasn''t you, secure your account: {{security_url}}',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'login_alert';

-- 2FA Enabled
UPDATE email_templates SET
  body_html = '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0">
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
  body_text = 'Hi {{user_name}},

Two-factor authentication has been successfully enabled on your account.

IMPORTANT: Make sure you''ve saved your backup codes in a secure location.

If you didn''t enable 2FA, change your password immediately: {{security_url}}',
  updated_at = CURRENT_TIMESTAMP
WHERE key = '2fa_enabled';

-- Email Change Verify
UPDATE email_templates SET
  body_html = '<h1 style="margin: 0 0 24px 0; font-size: 28px; font-weight: 700; color: #2563eb; line-height: 1.3;">
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
  body_text = 'Hi {{user_name}},

You requested to change your email address. Confirm here: {{verification_url}}

This link expires in {{expiry_hours}} hours.

If you didn''t request this, secure your account immediately.',
  updated_at = CURRENT_TIMESTAMP
WHERE key = 'email_change_verify';
