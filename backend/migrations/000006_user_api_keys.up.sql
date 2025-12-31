-- =============================================================================
-- USER API KEYS - Store user-provided API keys for external services
-- =============================================================================

-- User API keys table
CREATE TABLE IF NOT EXISTS user_api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- e.g., 'gemini', 'openai', 'anthropic'
    name VARCHAR(100) NOT NULL,     -- User-friendly name
    key_hash VARCHAR(64) NOT NULL,  -- SHA-256 hash for verification
    key_encrypted TEXT NOT NULL,    -- AES-256 encrypted key
    key_preview VARCHAR(20),        -- Last 4 chars for display (e.g., "...xyzA")
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMPTZ,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider)       -- One key per provider per user
);

CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);
CREATE INDEX idx_user_api_keys_provider ON user_api_keys(provider);
CREATE INDEX idx_user_api_keys_user_provider ON user_api_keys(user_id, provider) WHERE is_active = true;
