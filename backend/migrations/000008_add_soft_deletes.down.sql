-- Remove soft delete column from users table
DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

-- Remove soft delete column from files table
DROP INDEX IF EXISTS idx_files_deleted_at;
ALTER TABLE files DROP COLUMN IF EXISTS deleted_at;

-- Remove soft delete column from subscriptions table
DROP INDEX IF EXISTS idx_subscriptions_deleted_at;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS deleted_at;
