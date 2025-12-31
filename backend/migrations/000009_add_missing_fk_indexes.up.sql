-- Migration: Add missing foreign key indexes
-- These indexes improve query performance for foreign key lookups

-- Index for organization_invitations.invited_by_user_id
-- Speeds up queries like "who sent this invitation" and join operations
CREATE INDEX IF NOT EXISTS idx_organization_invitations_invited_by
    ON organization_invitations(invited_by_user_id);

-- Index for organization_members.invited_by_user_id
-- Speeds up queries like "who invited this member" and join operations
CREATE INDEX IF NOT EXISTS idx_organization_members_invited_by
    ON organization_members(invited_by_user_id);

-- Index for token_blacklist.user_id
-- Speeds up user-specific token revocation lookups
CREATE INDEX IF NOT EXISTS idx_token_blacklist_user_id
    ON token_blacklist(user_id);
