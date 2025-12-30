-- Organization invitations table for pending invites
CREATE TABLE IF NOT EXISTS organization_invitations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Relationship
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Invite details
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',

    -- Token for accepting invitation
    token VARCHAR(64) NOT NULL,

    -- Who invited
    invited_by_user_id INTEGER NOT NULL REFERENCES users(id),

    -- Expiration and acceptance
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,

    -- Only one pending invitation per email per org
    UNIQUE(organization_id, email)
);

-- Unique index on token for lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_invitations_token
ON organization_invitations(token);

-- Index for email lookup (for when user registers)
CREATE INDEX IF NOT EXISTS idx_org_invitations_email
ON organization_invitations(email);

-- Index for expiration cleanup
CREATE INDEX IF NOT EXISTS idx_org_invitations_expires
ON organization_invitations(expires_at);

-- Index for organization's pending invitations
CREATE INDEX IF NOT EXISTS idx_org_invitations_org_id
ON organization_invitations(organization_id);
