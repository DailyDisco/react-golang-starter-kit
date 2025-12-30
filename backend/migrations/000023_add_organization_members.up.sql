-- Organization members table for user-organization relationships
CREATE TABLE IF NOT EXISTS organization_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Relationship
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Role within organization: owner, admin, member
    role VARCHAR(50) NOT NULL DEFAULT 'member',

    -- Invitation tracking
    invited_by_user_id INTEGER REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE,
    accepted_at TIMESTAMP WITH TIME ZONE,

    -- Status: pending, active, suspended
    status VARCHAR(50) DEFAULT 'active',

    -- Ensure unique membership per org
    UNIQUE(organization_id, user_id)
);

-- Index for finding user's organizations
CREATE INDEX IF NOT EXISTS idx_org_members_user_id
ON organization_members(user_id);

-- Index for finding organization's members
CREATE INDEX IF NOT EXISTS idx_org_members_org_id
ON organization_members(organization_id);

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_org_members_status
ON organization_members(status);

-- Index for filtering by role
CREATE INDEX IF NOT EXISTS idx_org_members_role
ON organization_members(role);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_org_members_user_status
ON organization_members(user_id, status);
