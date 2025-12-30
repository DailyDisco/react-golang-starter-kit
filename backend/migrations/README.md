# Consolidated Migrations

This directory contains a consolidated version of all database migrations, reduced from 25 files to 5 logical groups for easier understanding and faster setup.

## Migration Groups

| File | Description | Tables |
|------|-------------|--------|
| `000001_core_schema` | Core user and file management | `users`, `files`, `user_preferences`, `data_exports` |
| `000002_auth` | Authentication and security | `token_blacklist`, `oauth_providers`, `user_sessions`, `user_two_factor`, `ip_blocklist`, `login_history` |
| `000003_payments` | Stripe billing | `subscriptions` |
| `000004_admin` | Admin features | `audit_logs`, `feature_flags`, `user_feature_flags`, `system_settings`, `announcement_banners`, `user_dismissed_announcements`, `email_templates` |
| `000005_organizations` | Multi-tenancy | `organizations`, `organization_members`, `organization_invitations` |

## Usage

To use the consolidated migrations instead of the original 25 files:

```bash
# Backup existing data if needed
pg_dump -U your_user your_database > backup.sql

# Option 1: Fresh database (recommended for new projects)
# Simply replace the migrations folder:
mv backend/migrations backend/migrations-original
mv backend/migrations-consolidated backend/migrations

# Option 2: Reset existing database
make db-reset
make dev
```

## Notes

- **River Job Queue**: River manages its own migrations via the `rivermigrate` package. No migration file is needed.
- **GORM Auto-migrate**: GORM's auto-migrate is disabled in favor of explicit SQL migrations for better control.
- **Performance Indexes**: All performance indexes are included in their respective migration files.

## Switching Back

If you need the original granular migrations:

```bash
mv backend/migrations backend/migrations-consolidated
mv backend/migrations-original backend/migrations
```
