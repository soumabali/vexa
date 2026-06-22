-- ROLLBACK_006.sql
-- Rollback 006_credential_sharing.sql

BEGIN;

-- Drop indexes first
DROP INDEX IF EXISTS idx_master_key_versions_user;
DROP INDEX IF EXISTS idx_credential_shares_active;
DROP INDEX IF EXISTS idx_credential_shares_owner;
DROP INDEX IF EXISTS idx_credential_shares_recipient;
DROP INDEX IF EXISTS idx_credential_shares_credential;
DROP INDEX IF EXISTS idx_credential_share_audit_time;
DROP INDEX IF EXISTS idx_credential_share_audit_share;

-- Drop tables
DROP TABLE IF EXISTS master_key_versions;
DROP TABLE IF EXISTS credential_share_audit;
DROP TABLE IF EXISTS credential_shares;

-- Remove added columns from credentials (PostgreSQL allows this)
ALTER TABLE credentials DROP COLUMN IF EXISTS tags;
ALTER TABLE credentials DROP COLUMN IF EXISTS metadata;
ALTER TABLE credentials DROP COLUMN IF EXISTS description;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = '006_credential_sharing';

COMMIT;