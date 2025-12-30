-- Remove password reset token fields from users table

DROP INDEX IF EXISTS idx_users_password_reset_token;

ALTER TABLE users DROP COLUMN IF EXISTS password_reset_token;
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_expires;
