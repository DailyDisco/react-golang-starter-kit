-- Test Database Initialization Script
-- This runs automatically when the test container starts

-- Create test schema
CREATE SCHEMA IF NOT EXISTS test;

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA test TO testuser;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA test TO testuser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA test TO testuser;

-- Create a function for test isolation via savepoints
CREATE OR REPLACE FUNCTION test_reset_sequence(seq_name TEXT)
RETURNS VOID AS $$
BEGIN
  EXECUTE format('ALTER SEQUENCE %I RESTART WITH 1', seq_name);
END;
$$ LANGUAGE plpgsql;

-- Create cleanup function for test data
CREATE OR REPLACE FUNCTION truncate_all_tables()
RETURNS VOID AS $$
DECLARE
  table_name TEXT;
BEGIN
  FOR table_name IN
    SELECT tablename FROM pg_tables
    WHERE schemaname = 'public'
    AND tablename NOT IN ('schema_migrations', 'goose_db_version')
  LOOP
    EXECUTE format('TRUNCATE TABLE %I CASCADE', table_name);
  END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Disable statement timeout for test setup
SET statement_timeout = 0;

-- Log that init completed
DO $$
BEGIN
  RAISE NOTICE 'Test database initialized successfully';
END $$;
