DROP FUNCTION IF EXISTS is_ip_blocked(VARCHAR);
DROP INDEX IF EXISTS idx_ip_blocklist_expires;
DROP INDEX IF EXISTS idx_ip_blocklist_active;
DROP INDEX IF EXISTS idx_ip_blocklist_ip;
DROP TABLE IF EXISTS ip_blocklist;
