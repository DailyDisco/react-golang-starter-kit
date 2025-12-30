-- Initial schema based on GORM models
-- Models: User, File

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at VARCHAR(255),
    updated_at VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verification_expires VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user'
);

-- User indexes (matching GORM annotations)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    created_at VARCHAR(255),
    updated_at VARCHAR(255),
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(255),
    file_size BIGINT,
    location VARCHAR(255),
    content BYTEA,
    storage_type VARCHAR(255) DEFAULT 'database'
);
