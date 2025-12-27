-- IP blocklist for security - blocks malicious IPs from accessing the system
CREATE TABLE IF NOT EXISTS ip_blocklist (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL, -- Supports IPv4 and IPv6
    ip_range VARCHAR(50), -- CIDR notation for ranges (e.g., 192.168.1.0/24)
    reason VARCHAR(500),
    block_type VARCHAR(20) DEFAULT 'manual' CHECK (block_type IN ('manual', 'auto_rate_limit', 'auto_brute_force', 'auto_suspicious')),
    blocked_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    hit_count INTEGER DEFAULT 0, -- Number of blocked requests
    expires_at TIMESTAMP, -- NULL for permanent blocks
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ip_blocklist_ip ON ip_blocklist(ip_address);
CREATE INDEX idx_ip_blocklist_active ON ip_blocklist(is_active) WHERE is_active = true;
CREATE INDEX idx_ip_blocklist_expires ON ip_blocklist(expires_at) WHERE expires_at IS NOT NULL;

-- Function to check if an IP is blocked
CREATE OR REPLACE FUNCTION is_ip_blocked(check_ip VARCHAR(45))
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM ip_blocklist
        WHERE is_active = true
        AND (expires_at IS NULL OR expires_at > NOW())
        AND (
            ip_address = check_ip
            OR (ip_range IS NOT NULL AND check_ip::inet <<= ip_range::inet)
        )
    );
END;
$$ LANGUAGE plpgsql;
