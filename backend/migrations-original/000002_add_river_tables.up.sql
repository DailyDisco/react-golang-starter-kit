-- River job queue schema
-- NOTE: River manages its own migrations via rivermigrate package.
-- This file is intentionally empty to avoid conflicts with River's internal migrator.
-- The River schema (tables, indexes, functions) is created automatically when
-- the job system initializes in internal/jobs/client.go
--
-- Reference: https://github.com/riverqueue/river

-- No-op: River handles its own schema
SELECT 1;
