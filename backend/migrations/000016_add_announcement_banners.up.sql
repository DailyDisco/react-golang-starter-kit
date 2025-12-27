-- Announcement banners for site-wide notifications
CREATE TABLE IF NOT EXISTS announcement_banners (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'info' CHECK (type IN ('info', 'warning', 'error', 'success', 'maintenance')),
    link_url VARCHAR(500),
    link_text VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    is_dismissible BOOLEAN DEFAULT true,
    show_on_pages JSONB DEFAULT '["*"]', -- ["*"] for all, or specific paths like ["/dashboard", "/settings"]
    target_roles TEXT[] DEFAULT NULL, -- NULL means all users, or specific roles
    priority INTEGER DEFAULT 0, -- Higher priority shows first
    starts_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ends_at TIMESTAMP,
    view_count INTEGER DEFAULT 0,
    dismiss_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_announcement_banners_active ON announcement_banners(is_active) WHERE is_active = true;
CREATE INDEX idx_announcement_banners_dates ON announcement_banners(starts_at, ends_at);
CREATE INDEX idx_announcement_banners_priority ON announcement_banners(priority DESC);

-- Track dismissed announcements per user
CREATE TABLE IF NOT EXISTS user_dismissed_announcements (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id INTEGER NOT NULL REFERENCES announcement_banners(id) ON DELETE CASCADE,
    dismissed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, announcement_id)
);
