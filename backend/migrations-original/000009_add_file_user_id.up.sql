-- Add user_id to files table for ownership tracking
-- This enables per-user file isolation and IDOR protection

-- Add user_id column (nullable initially for existing data)
ALTER TABLE files ADD COLUMN IF NOT EXISTS user_id INTEGER;

-- Create index for fast user-based file lookups
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);

-- Add foreign key constraint to users table
-- Using SET NULL on delete to preserve file records if user is deleted
ALTER TABLE files ADD CONSTRAINT fk_files_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
