-- =============================================================================
-- ORGANIZATIONS - Multi-tenancy support
-- Consolidated from: 000022, 000023, 000024
-- =============================================================================

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url VARCHAR(500),
    plan VARCHAR(50) DEFAULT 'free',
    stripe_customer_id VARCHAR(255),
    settings JSONB DEFAULT '{}',
    created_by_user_id INTEGER NOT NULL REFERENCES users(id)
);

CREATE UNIQUE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_organizations_stripe_customer ON organizations(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id != '';
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX idx_organizations_plan ON organizations(plan);
CREATE INDEX idx_organizations_created_by ON organizations(created_by_user_id);

-- Organization members table
CREATE TABLE IF NOT EXISTS organization_members (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    invited_by_user_id INTEGER REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    accepted_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX idx_org_members_org_id ON organization_members(organization_id);
CREATE INDEX idx_org_members_status ON organization_members(status);
CREATE INDEX idx_org_members_role ON organization_members(role);
CREATE INDEX idx_org_members_user_status ON organization_members(user_id, status);

-- Organization invitations table
CREATE TABLE IF NOT EXISTS organization_invitations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    token VARCHAR(64) NOT NULL,
    invited_by_user_id INTEGER NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    UNIQUE(organization_id, email)
);

CREATE UNIQUE INDEX idx_org_invitations_token ON organization_invitations(token);
CREATE INDEX idx_org_invitations_email ON organization_invitations(email);
CREATE INDEX idx_org_invitations_expires ON organization_invitations(expires_at);
CREATE INDEX idx_org_invitations_org_id ON organization_invitations(organization_id);
