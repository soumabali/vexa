-- Rollback: Credential Sharing
-- Reverses migration 008

DROP TABLE IF EXISTS share_audit_log;
DROP TABLE IF EXISTS credential_shares;
DROP TABLE IF EXISTS user_share_keys;

DROP FUNCTION IF EXISTS update_share_timestamp();
DROP FUNCTION IF EXISTS expire_shares();
