-- Add min_plan column to feature_flags table for plan-based feature gating
ALTER TABLE feature_flags
ADD COLUMN min_plan VARCHAR(20) DEFAULT '' NOT NULL;

-- Create index for filtering by plan
CREATE INDEX idx_feature_flags_min_plan ON feature_flags(min_plan);

-- Add comment for documentation
COMMENT ON COLUMN feature_flags.min_plan IS 'Minimum subscription plan required: empty string = all plans, pro, enterprise';
