-- Remove security email templates
DELETE FROM email_templates WHERE key IN ('email_change_verify', 'two_factor_code', 'account_locked');
