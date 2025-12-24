-- Remove impersonation columns
ALTER TABLE users DROP COLUMN IF EXISTS impersonation_started_at;
ALTER TABLE users DROP COLUMN IF EXISTS impersonated_by;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_feature_flags_flag_id;
DROP INDEX IF EXISTS idx_user_feature_flags_user_id;
DROP INDEX IF EXISTS idx_feature_flags_enabled;
DROP INDEX IF EXISTS idx_feature_flags_key;

-- Drop tables
DROP TABLE IF EXISTS user_feature_flags;
DROP TABLE IF EXISTS feature_flags;
