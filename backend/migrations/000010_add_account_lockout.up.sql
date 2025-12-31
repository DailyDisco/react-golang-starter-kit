-- Migration: Add account lockout fields for brute-force protection
-- These fields track failed login attempts and temporary account locks

-- Add lockout fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS last_failed_login TIMESTAMP WITH TIME ZONE;

-- Index for locked accounts (for admin queries and cleanup jobs)
CREATE INDEX IF NOT EXISTS idx_users_locked_until
    ON users(locked_until)
    WHERE locked_until IS NOT NULL;

-- Partial index for accounts with failed attempts (for monitoring)
CREATE INDEX IF NOT EXISTS idx_users_failed_attempts
    ON users(failed_login_attempts)
    WHERE failed_login_attempts > 0;

-- Comment for documentation
COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN users.locked_until IS 'Account locked until this time (NULL if not locked)';
COMMENT ON COLUMN users.last_failed_login IS 'Timestamp of last failed login attempt';
