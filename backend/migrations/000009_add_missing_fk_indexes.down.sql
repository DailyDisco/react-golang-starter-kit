-- Rollback: Remove foreign key indexes added in 000009

DROP INDEX IF EXISTS idx_organization_invitations_invited_by;
DROP INDEX IF EXISTS idx_organization_members_invited_by;
DROP INDEX IF EXISTS idx_token_blacklist_user_id;
