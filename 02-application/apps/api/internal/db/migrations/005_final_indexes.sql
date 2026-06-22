-- 005_final_indexes_and_optimizations.sql
-- Final indexes and performance optimizations

BEGIN;

-- Sessions additional indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_active_time ON sessions(user_id, status, started_at DESC) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_sessions_recording ON sessions(recording_path) WHERE recording_path IS NOT NULL;

-- Credentials: index for owner lookup + expiry tracking
CREATE INDEX IF NOT EXISTS idx_credentials_owner ON credentials(owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credentials_expiry ON credentials(expires_at) WHERE expires_at IS NOT NULL;

-- Devices: index for trust status
CREATE INDEX IF NOT EXISTS idx_devices_trusted ON devices(trusted) WHERE trusted = FALSE;

-- Hosts: additional optimized indexes
CREATE INDEX IF NOT EXISTS idx_hosts_active ON hosts(is_active, owner_id) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_hosts_protocol_port ON hosts(protocol, port);

-- Users: index for login tracking
CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, is_active) WHERE is_active = TRUE;

-- Connection history: partition-aware index (for time-range queries)
CREATE INDEX IF NOT EXISTS idx_conn_history_range ON connection_history(started_at DESC, user_id);

-- Vault keys: quick lookup for active key by user
CREATE INDEX IF NOT EXISTS idx_vault_keys_active_user ON vault_keys(user_id) WHERE is_active = TRUE;

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('005_final_indexes', 'Add optimized indexes for all queries in Sprint 0-3')
ON CONFLICT (version) DO NOTHING;

COMMIT;
