-- ROLLBACK_001.sql
-- Rollback 001_initial.sql: drops all base schema tables
-- WARNING: This destroys ALL data. Use with extreme caution.

BEGIN;

-- Drop in reverse dependency order
DROP TRIGGER IF EXISTS update_credentials_updated_at ON credentials;
DROP TRIGGER IF EXISTS update_hosts_updated_at ON hosts;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop all tables (CASCADE will handle FK dependencies)
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS connection_history CASCADE;
DROP TABLE IF EXISTS hosts CASCADE;
DROP TABLE IF EXISTS host_groups CASCADE;
DROP TABLE IF EXISTS credentials CASCADE;
DROP TABLE IF EXISTS devices CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;

-- Note: audit_log partitions are dropped via CASCADE
-- Note: pgcrypto/ltree/citext extensions should NOT be dropped
--       as they may be shared with other databases on the same instance

DELETE FROM schema_migrations WHERE version = '001_initial';

COMMIT;
