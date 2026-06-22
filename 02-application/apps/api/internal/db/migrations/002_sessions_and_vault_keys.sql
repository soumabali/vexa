-- 002_sessions_and_vault_keys.sql
-- Active sessions tracking + vault encryption key management

-- Sessions: active terminal/proxy sessions
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    host_id UUID REFERENCES hosts(id) ON DELETE CASCADE,
    protocol TEXT NOT NULL CHECK (protocol IN ('ssh', 'rdp', 'vnc')),
    session_type TEXT DEFAULT 'terminal' CHECK (session_type IN ('terminal', 'sftp', 'tunnel', 'proxy')),
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    bytes_in BIGINT DEFAULT 0,
    bytes_out BIGINT DEFAULT 0,
    recording_path TEXT,
    terminated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    termination_reason TEXT CHECK (termination_reason IN ('normal', 'killed', 'timeout', 'error')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed', 'killed', 'timeout'))
);

CREATE INDEX idx_sessions_user ON sessions(user_id, started_at);
CREATE INDEX idx_sessions_host ON sessions(host_id, started_at);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_active ON sessions(user_id, status) WHERE status = 'active';
CREATE INDEX idx_sessions_device ON sessions(device_id);

-- Update session triggers helper functions (must be defined before the triggers)
CREATE OR REPLACE FUNCTION sync_session_to_connection_history()
RETURNS TRIGGER AS $$
BEGIN
    -- Sync active session start to connection_history
    INSERT INTO connection_history (id, user_id, host_id, protocol, started_at, status)
    VALUES (NEW.id, NEW.user_id, NEW.host_id, NEW.protocol, NEW.started_at, 'active')
    ON CONFLICT (id) DO NOTHING;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION sync_session_end_to_connection_history()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.status != OLD.status AND NEW.status IN ('closed', 'killed', 'timeout') THEN
        UPDATE connection_history
        SET ended_at = NEW.ended_at,
            bytes_in = NEW.bytes_in,
            bytes_out = NEW.bytes_out,
            status = CASE NEW.status
                WHEN 'killed' THEN 'killed'
                ELSE 'closed'
            END,
            terminated_by = NEW.terminated_by
        WHERE id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Update session triggers
CREATE TRIGGER update_connection_history_sessions_sync
    AFTER INSERT ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION sync_session_to_connection_history();

CREATE TRIGGER trigger_sync_session_end
    AFTER UPDATE OF status, ended_at, bytes_in, bytes_out, terminated_by ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION sync_session_end_to_connection_history();

-- Vault keys: per-user vault encryption key metadata
CREATE TABLE vault_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INT NOT NULL DEFAULT 1,
    algorithm TEXT NOT NULL DEFAULT 'argon2id-aes-256-gcm' CHECK (algorithm IN ('argon2id-aes-256-gcm')),
    key_hash TEXT NOT NULL,                    -- SHA-256 hash of derived key for verification
    salt BYTEA NOT NULL,                       -- Argon2id salt
    wrapping_salt BYTEA NOT NULL,              -- Secondary salt for key wrap
    wrapped_key BYTEA NOT NULL,                -- Master key wrapped with device-derived key
    is_active BOOLEAN DEFAULT TRUE,
    rotated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_active_vault_key_per_user UNIQUE (user_id) -- only one active key
);

CREATE INDEX idx_vault_keys_user ON vault_keys(user_id, is_active);
CREATE INDEX idx_vault_keys_active ON vault_keys(is_active) WHERE is_active = TRUE;

-- Record initial migration
INSERT INTO schema_migrations (version, description)
VALUES ('002_sessions_and_vault_keys', 'Add sessions and vault_keys tables')
ON CONFLICT (version) DO NOTHING;
