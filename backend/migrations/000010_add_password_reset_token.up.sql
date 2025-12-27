-- Add separate password reset token fields to users table
-- This separates password reset tokens from email verification tokens for better security

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_token VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_expires VARCHAR(255);

-- Add unique index for password reset token lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_password_reset_token ON users(password_reset_token) WHERE password_reset_token IS NOT NULL AND password_reset_token != '';

-- Comment on new columns
COMMENT ON COLUMN users.password_reset_token IS 'Token for password reset requests (separate from email verification)';
COMMENT ON COLUMN users.password_reset_expires IS 'Expiration time for password reset token';
