-- Organizations table for multi-tenancy support
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Basic info
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    logo_url VARCHAR(500),

    -- Billing/plan info
    plan VARCHAR(50) DEFAULT 'free',
    stripe_customer_id VARCHAR(255),

    -- Settings stored as JSONB
    settings JSONB DEFAULT '{}',

    -- Owner tracking
    created_by_user_id INTEGER NOT NULL REFERENCES users(id)
);

-- Unique constraint on slug for active (non-deleted) organizations
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_slug
ON organizations(slug) WHERE deleted_at IS NULL;

-- Index for Stripe customer lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_stripe_customer
ON organizations(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id != '';

-- Index for soft deletes
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at
ON organizations(deleted_at);

-- Index for plan filtering
CREATE INDEX IF NOT EXISTS idx_organizations_plan
ON organizations(plan);

-- Index for creator lookup
CREATE INDEX IF NOT EXISTS idx_organizations_created_by
ON organizations(created_by_user_id);
