-- ROLLBACK_005.sql
-- Rollback: final indexes

DROP INDEX IF EXISTS idx_sessions_user_active_time;
DROP INDEX IF EXISTS idx_sessions_recording;
DROP INDEX IF EXISTS idx_credentials_owner;
DROP INDEX IF EXISTS idx_credentials_expiry;
DROP INDEX IF EXISTS idx_devices_trusted;
DROP INDEX IF EXISTS idx_hosts_active;
DROP INDEX IF EXISTS idx_hosts_protocol_port;
DROP INDEX IF EXISTS idx_users_email_active;
DROP INDEX IF EXISTS idx_conn_history_range;
DROP INDEX IF EXISTS idx_vault_keys_active_user;

DELETE FROM schema_migrations WHERE version = '005_final_indexes';

-- ROLLBACK_004.sql
-- Rollback: host groups hierarchy

BEGIN;

DROP INDEX IF EXISTS idx_host_groups_ltree_path;
DROP INDEX IF EXISTS idx_host_groups_path_pattern;
DROP FUNCTION IF EXISTS get_group_descendants(TEXT);
DROP FUNCTION IF EXISTS path_to_ltree(TEXT);
DROP VIEW IF EXISTS v_host_groups_hierarchy;
ALTER TABLE host_groups DROP COLUMN IF EXISTS ltree_path;

DELETE FROM schema_migrations WHERE version = '004_host_groups_hierarchy';

COMMIT;
