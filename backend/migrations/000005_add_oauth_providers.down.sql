-- Remove OAuth-related columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;

-- Drop indexes
DROP INDEX IF EXISTS idx_oauth_providers_user_id;
DROP INDEX IF EXISTS idx_oauth_providers_provider_user;

-- Drop OAuth providers table
DROP TABLE IF EXISTS oauth_providers;
