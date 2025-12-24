-- River job queue tables
-- These tables are required for the River job queue system
-- Reference: https://github.com/riverqueue/river

CREATE TABLE IF NOT EXISTS river_migration (
    id SERIAL PRIMARY KEY,
    version INTEGER NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS river_job (
    id BIGSERIAL PRIMARY KEY,
    state VARCHAR(255) NOT NULL DEFAULT 'available',
    attempt INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 25,
    attempted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    finalized_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    priority INTEGER NOT NULL DEFAULT 1,
    args JSONB NOT NULL DEFAULT '{}',
    attempted_by TEXT[],
    errors JSONB[],
    kind TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    queue TEXT NOT NULL DEFAULT 'default',
    tags VARCHAR(255)[] NOT NULL DEFAULT '{}',
    unique_key BYTEA,
    unique_states BIT(8)
);

CREATE INDEX IF NOT EXISTS river_job_state_idx ON river_job (state);
CREATE INDEX IF NOT EXISTS river_job_kind_idx ON river_job (kind);
CREATE INDEX IF NOT EXISTS river_job_queue_idx ON river_job (queue);
CREATE INDEX IF NOT EXISTS river_job_scheduled_at_idx ON river_job (scheduled_at);
CREATE INDEX IF NOT EXISTS river_job_prioritized_fetching_idx ON river_job (state, queue, priority DESC, scheduled_at, id) WHERE state = 'available';
CREATE UNIQUE INDEX IF NOT EXISTS river_job_unique_key_idx ON river_job (unique_key) WHERE unique_key IS NOT NULL AND unique_states IS NOT NULL;

CREATE TABLE IF NOT EXISTS river_leader (
    elected_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    leader_id TEXT NOT NULL,
    name TEXT PRIMARY KEY NOT NULL
);

CREATE INDEX IF NOT EXISTS river_leader_name_idx ON river_leader (name);

CREATE TABLE IF NOT EXISTS river_queue (
    name TEXT PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB NOT NULL DEFAULT '{}',
    paused_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert migration version marker
INSERT INTO river_migration (version) VALUES (6) ON CONFLICT (version) DO NOTHING;
