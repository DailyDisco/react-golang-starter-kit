-- Remove refresh token fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token;
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token_expires;

-- Drop token blacklist table
DROP TABLE IF EXISTS token_blacklist;
