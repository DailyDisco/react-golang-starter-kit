-- OAuth providers table for social login
CREATE TABLE IF NOT EXISTS oauth_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,          -- 'google', 'github'
    provider_user_id VARCHAR(255) NOT NULL, -- User ID from the OAuth provider
    email VARCHAR(255),                      -- Email from OAuth provider (may differ from user's email)
    access_token TEXT,                       -- OAuth access token (encrypted)
    refresh_token TEXT,                      -- OAuth refresh token (encrypted)
    token_expires_at TIMESTAMP,              -- When the OAuth token expires
    raw_data JSONB,                          -- Raw profile data from provider
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id),      -- Each provider user can only be linked once
    UNIQUE(user_id, provider)                -- Each user can only have one account per provider
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_oauth_providers_user_id ON oauth_providers(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_providers_provider_user ON oauth_providers(provider, provider_user_id);

-- Add OAuth-related fields to users table for quick access
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- Comment on table
COMMENT ON TABLE oauth_providers IS 'Stores OAuth provider connections for social login';
