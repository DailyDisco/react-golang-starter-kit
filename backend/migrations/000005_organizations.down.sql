-- Rollback organizations schema
DROP TABLE IF EXISTS organization_invitations CASCADE;
DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
