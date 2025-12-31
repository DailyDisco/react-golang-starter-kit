-- Migration: Add idempotency keys table for safe POST request retries
-- This prevents duplicate actions when clients retry failed requests

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id SERIAL PRIMARY KEY,

    -- The idempotency key provided by the client
    key VARCHAR(255) NOT NULL,

    -- User who made the request (NULL for unauthenticated requests)
    user_id INT REFERENCES users(id) ON DELETE CASCADE,

    -- Request fingerprint (method + path + body hash)
    request_hash VARCHAR(64) NOT NULL,

    -- Cached response
    response_status INT NOT NULL,
    response_body JSONB,
    response_headers JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Unique constraint: one key per user (or per anonymous session)
    CONSTRAINT unique_idempotency_key UNIQUE (key, user_id)
);

-- Index for fast lookups by key and user
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_lookup
    ON idempotency_keys(key, user_id);

-- Index for cleanup job (expired keys)
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires
    ON idempotency_keys(expires_at);

-- Comment for documentation
COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys and cached responses for safe request retries';
COMMENT ON COLUMN idempotency_keys.key IS 'Client-provided idempotency key (UUID recommended)';
COMMENT ON COLUMN idempotency_keys.expires_at IS 'Keys expire after 24 hours by default';
