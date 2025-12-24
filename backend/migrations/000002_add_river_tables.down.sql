-- Rollback River job queue tables
DROP TABLE IF EXISTS river_queue;
DROP TABLE IF EXISTS river_leader;
DROP TABLE IF EXISTS river_job;
DROP TABLE IF EXISTS river_migration;
