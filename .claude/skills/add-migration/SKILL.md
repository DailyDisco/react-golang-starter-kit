---
name: add-migration
description: Generate database migrations using project conventions. Use when adding tables, columns, indexes, or schema changes.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(ls:*), Bash(make:*), AskUserQuestion
context-files:
  - backend/migrations/000001_init.up.sql
  - backend/internal/models/models.go
  - backend/internal/database/migrate.go
---

# Database Migration Generator

You are my migration generator for this React + Go starter kit. Generate migrations following the exact project conventions.

## Project Context

This project uses:
- **golang-migrate** for migration execution
- **Sequential numbering**: 000001, 000002, 000003, etc.
- **Paired files**: `{number}_{name}.up.sql` and `{number}_{name}.down.sql`
- **GORM models** in `backend/internal/models/`
- **PostgreSQL** as the database

### Existing Tables (from 000001_init.up.sql)

**Section 1 - Users & Core:**
- `users` - Core user table with auth, OAuth, 2FA, security fields
- `files` - User file storage (S3/DB)
- `user_preferences` - JSON preferences per user
- `data_exports` - GDPR data export requests

**Section 2 - Authentication:**
- `token_blacklist` - Revoked JWT tokens
- `oauth_providers` - Linked OAuth accounts
- `user_sessions` - Active sessions
- `user_two_factor` - 2FA backup codes
- `ip_blocklist` - Blocked IPs
- `login_history` - Login audit trail

**Section 3 - Payments:**
- `subscriptions` - Stripe subscriptions (user or org level)

**Section 4 - Organizations:**
- `organizations` - Multi-tenant orgs
- `organization_members` - Org membership
- `organization_invitations` - Pending invites

**Section 5 - Admin:**
- `audit_logs` - Admin audit trail
- `feature_flags` - Feature toggles
- `user_feature_flags` - Per-user overrides
- `system_settings` - App configuration
- `announcement_banners` - UI announcements
- `email_templates` - Email content

**Section 6 - API Keys:**
- `user_api_keys` - External API keys (OpenAI, etc.)

**Section 7 - Idempotency:**
- `idempotency_keys` - Request deduplication

**From 000002 - Usage Tracking:**
- `usage_events` - Granular usage log
- `usage_periods` - Aggregated usage
- `usage_alerts` - Threshold notifications

## Objective

Generate safe, reversible database migrations that:
- Follow project naming conventions
- Include proper indexes
- Have complete rollback scripts
- Update GORM models when needed

## Hard Rules

1. ALWAYS generate BOTH .up.sql AND .down.sql files
2. ALWAYS use sequential numbering (check existing migrations first)
3. NEVER drop columns/tables without explicit user confirmation
4. ALWAYS use `IF NOT EXISTS` / `IF EXISTS` for safety
5. ALWAYS add indexes on foreign keys
6. Use `TIMESTAMPTZ` not `TIMESTAMP` for timestamps
7. Use `BIGSERIAL` for IDs, `TEXT` for strings, `JSONB` for JSON

## Guided Workflow

### Phase 1: Gather Requirements

Ask the user:

1. **Migration type**:
   - [ ] Add new table
   - [ ] Add column(s) to existing table
   - [ ] Add index
   - [ ] Modify column
   - [ ] Add constraint
   - [ ] Other

2. **Details** based on type:
   - For new table: table name, columns with types
   - For new column: table name, column name, type, nullable, default
   - For index: table, columns, unique?

### Phase 2: Determine Migration Number

```bash
# Check existing migrations
ls backend/migrations/*.up.sql | tail -1
```

Next number = highest + 1 (e.g., 000004)

### Phase 3: Generate Migration Files

**File: `backend/migrations/000XXX_{name}.up.sql`**

```sql
-- Migration: {description}
-- Created: {timestamp}

BEGIN;

-- Add your schema changes here
CREATE TABLE IF NOT EXISTS {table_name} (
    id BIGSERIAL PRIMARY KEY,
    -- columns
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_{table}_{column} ON {table}({column});

COMMIT;
```

**File: `backend/migrations/000XXX_{name}.down.sql`**

```sql
-- Rollback: {description}

BEGIN;

-- Reverse the changes in exact opposite order
DROP TABLE IF EXISTS {table_name};

COMMIT;
```

### Phase 4: Update GORM Model (if new table)

Add to `backend/internal/models/models.go`:

```go
// {Entity} represents a {description}
type {Entity} struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

    // Add fields matching migration
}

// TableName overrides the table name
func ({Entity}) TableName() string {
    return "{table_name}"
}
```

### Phase 5: Provide Makefile Commands

```bash
# Apply migration
make migrate-up

# Verify
make migrate-version

# If issues, rollback
make migrate-down

# Test the cycle
make migrate-validate
```

## Common Patterns

### Adding a Column

```sql
-- up.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- down.sql
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
```

### Adding a NOT NULL Column (safe way)

```sql
-- up.sql
ALTER TABLE users ADD COLUMN status TEXT;
UPDATE users SET status = 'active' WHERE status IS NULL;
ALTER TABLE users ALTER COLUMN status SET NOT NULL;
ALTER TABLE users ALTER COLUMN status SET DEFAULT 'active';

-- down.sql
ALTER TABLE users ALTER COLUMN status DROP NOT NULL;
ALTER TABLE users ALTER COLUMN status DROP DEFAULT;
ALTER TABLE users DROP COLUMN IF EXISTS status;
```

### Adding Foreign Key with Index

```sql
-- up.sql
ALTER TABLE orders ADD COLUMN user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- down.sql
DROP INDEX IF EXISTS idx_orders_user_id;
ALTER TABLE orders DROP COLUMN IF EXISTS user_id;
```

### Adding a New Table

```sql
-- up.sql
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'info',
    read BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id) WHERE read = FALSE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);

-- down.sql
DROP TABLE IF EXISTS notifications;
```

## Output Format

```
## Migration Generated

**Files created:**
- `backend/migrations/000004_add_notifications.up.sql`
- `backend/migrations/000004_add_notifications.down.sql`

**GORM model updated:**
- `backend/internal/models/models.go` (if applicable)

**Next steps:**
1. Review the generated SQL
2. Run: `make migrate-up`
3. Verify: `make migrate-version`
4. Test rollback: `make migrate-down && make migrate-up`

**Rollback command:**
```bash
make migrate-down
```
```

## Constraints

- Migration names: snake_case, descriptive (e.g., `add_user_avatar`, `create_notifications_table`)
- Max 1 major change per migration (don't mix creating tables with altering others)
- Always wrap in `BEGIN;`/`COMMIT;` for transactional safety
- Include comments explaining the purpose
