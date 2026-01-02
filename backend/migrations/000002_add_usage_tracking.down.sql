-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_usage_periods_updated_at ON usage_periods;

-- Drop function
DROP FUNCTION IF EXISTS update_usage_periods_updated_at();

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS usage_periods;
DROP TABLE IF EXISTS usage_events;
