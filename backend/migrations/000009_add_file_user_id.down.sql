-- Remove user_id from files table

-- Drop foreign key constraint
ALTER TABLE files DROP CONSTRAINT IF EXISTS fk_files_user_id;

-- Drop index
DROP INDEX IF EXISTS idx_files_user_id;

-- Drop column
ALTER TABLE files DROP COLUMN IF EXISTS user_id;
