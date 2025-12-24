# ADR 001: Database Migrations with golang-migrate

## Status

Accepted

## Context

The application needs a reliable way to manage database schema changes across development, staging, and production environments. GORM's AutoMigrate was initially used for convenience during development, but it has limitations:

- No support for rollbacks
- Limited control over schema changes
- Risky in production environments
- No version tracking
- Difficult to coordinate schema changes across team members

## Decision

We will use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations with the following approach:

1. **SQL-based migrations**: Write migrations in raw SQL for maximum control
2. **Sequential versioning**: Use numbered migration files (000001_name.up.sql, 000001_name.down.sql)
3. **Makefile commands**: Provide easy-to-use make targets for common operations
4. **CI integration**: Validate migrations in CI pipeline before merging
5. **Optional auto-run**: Support `RUN_MIGRATIONS=true` for development convenience

### Migration Commands

```bash
make migrate-up        # Run all pending migrations
make migrate-down      # Rollback last migration
make migrate-create name=add_users  # Create new migration
make migrate-version   # Show current version
make migrate-validate  # Test up/down/up cycle
```

## Consequences

### Positive

- Full control over schema changes with raw SQL
- Rollback capability for safe deployments
- Version tracking in the schema_migrations table
- CI validates migrations before merge
- Team can review schema changes in PRs

### Negative

- More manual work than AutoMigrate for simple changes
- Need to maintain both up and down migrations
- Developers must remember to create migrations for model changes

## Alternatives Considered

### Option 1: Continue with GORM AutoMigrate

- Pros: Simple, automatic, no extra files
- Cons: No rollbacks, risky in production, no version control

### Option 2: goose

- Pros: Popular, good CLI, supports Go migrations
- Cons: Less active maintenance, golang-migrate has broader adoption

### Option 3: Atlas

- Pros: Modern, declarative, great tooling
- Cons: Steeper learning curve, overkill for this project size

## References

- [golang-migrate GitHub](https://github.com/golang-migrate/migrate)
- [backend/migrations/](../../backend/migrations/)
- [backend/Makefile](../../backend/Makefile)
