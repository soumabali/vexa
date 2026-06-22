-- ROLLBACK_002.sql
-- Rollback: sessions and vault_keys tables

BEGIN;

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_sync_session_end ON sessions;
DROP TRIGGER IF EXISTS trigger_sync_session_end ON connection_history;
DROP TRIGGER IF EXISTS update_connection_history_sessions_sync ON sessions;
DROP FUNCTION IF EXISTS sync_session_end_to_connection_history();
DROP FUNCTION IF EXISTS sync_session_to_connection_history();

-- Drop sessions table
DROP TABLE IF EXISTS sessions;

-- Drop vault_keys table
DROP TABLE IF EXISTS vault_keys;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = '002_sessions_and_vault_keys';

COMMIT;
