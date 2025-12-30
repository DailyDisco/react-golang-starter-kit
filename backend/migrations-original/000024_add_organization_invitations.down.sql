-- Drop organization invitations table
DROP INDEX IF EXISTS idx_org_invitations_org_id;
DROP INDEX IF EXISTS idx_org_invitations_expires;
DROP INDEX IF EXISTS idx_org_invitations_email;
DROP INDEX IF EXISTS idx_org_invitations_token;
DROP TABLE IF EXISTS organization_invitations;
