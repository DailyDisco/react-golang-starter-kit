-- Rollback: Remove idempotency keys table

DROP INDEX IF EXISTS idx_idempotency_keys_expires;
DROP INDEX IF EXISTS idx_idempotency_keys_lookup;
DROP TABLE IF EXISTS idempotency_keys;
