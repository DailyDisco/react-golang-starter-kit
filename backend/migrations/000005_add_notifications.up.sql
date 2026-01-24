-- Create notifications table for in-app notification center
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'info',
    title VARCHAR(255) NOT NULL,
    message TEXT,
    link VARCHAR(500),
    read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Index for fetching user notifications ordered by time
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Index for filtering unread notifications
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, read) WHERE read = FALSE;

-- Add constraint for valid notification types
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('info', 'success', 'warning', 'error', 'announcement', 'billing', 'security'));
