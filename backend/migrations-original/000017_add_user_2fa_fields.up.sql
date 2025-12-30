-- Add 2FA and account deletion tracking fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_password_change_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45);
ALTER TABLE users ADD COLUMN IF NOT EXISTS deletion_requested_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deletion_reason TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_data_export_at TIMESTAMP;

-- Index for finding locked accounts
CREATE INDEX idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;

-- Index for finding accounts scheduled for deletion
CREATE INDEX idx_users_deletion ON users(deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;
