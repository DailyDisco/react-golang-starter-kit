-- Remove min_plan column from feature_flags table
DROP INDEX IF EXISTS idx_feature_flags_min_plan;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS min_plan;
