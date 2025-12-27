-- Two-factor authentication table
-- Security: TOTP secret MUST be encrypted before storage using AES-256-GCM
-- Backup codes MUST be hashed with bcrypt before storage
CREATE TABLE IF NOT EXISTS user_two_factor (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    -- TOTP secret encrypted with AES-256-GCM (base64 encoded)
    -- Format: nonce:ciphertext (nonce is 12 bytes, prepended to ciphertext)
    encrypted_secret TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT false,
    -- Backup codes stored as bcrypt hashes in JSON array
    -- Each code is 8 alphanumeric characters, hashed individually
    backup_codes_hash JSONB DEFAULT '[]',
    backup_codes_remaining INTEGER DEFAULT 10,
    -- Tracking fields
    verified_at TIMESTAMP,
    last_used_at TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_two_factor_user_id ON user_two_factor(user_id);
CREATE INDEX idx_user_two_factor_enabled ON user_two_factor(is_enabled) WHERE is_enabled = true;
