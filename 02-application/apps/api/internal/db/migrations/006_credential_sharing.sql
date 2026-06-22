-- 006_credential_sharing.sql
-- Credential sharing with encrypted key transfer, role-based permissions, and time-limited access.

-- Credential shares: owner grants recipient encrypted access to a credential.
CREATE TABLE credential_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'write')),
    encrypted_key TEXT,                      -- Credential's encrypted_data re-encrypted for recipient (NULL = use owner key)
    share_salt BYTEA,                        -- Salt for deriving share-specific key
    expires_at TIMESTAMPTZ,                 -- NULL = never expires
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_share_per_recipient UNIQUE (credential_id, recipient_id)
);

CREATE INDEX idx_credential_shares_credential ON credential_shares(credential_id);
CREATE INDEX idx_credential_shares_recipient ON credential_shares(recipient_id);
CREATE INDEX idx_credential_shares_owner ON credential_shares(owner_id);
CREATE INDEX idx_credential_shares_active ON credential_shares(credential_id, recipient_id)
    WHERE (expires_at IS NULL OR expires_at > '1970-01-01'::timestamptz);

-- Audit for sharing operations (extends audit_log)
CREATE TABLE credential_share_audit (
    id BIGSERIAL PRIMARY KEY,
    share_id UUID NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('created', 'accessed', 'revoked', 'expired')),
    performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    performed_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    details JSONB
);

CREATE INDEX idx_credential_share_audit_share ON credential_share_audit(share_id);
CREATE INDEX idx_credential_share_audit_time ON credential_share_audit(performed_at);

-- Master key versions: per-user key rotation tracking
CREATE TABLE master_key_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'argon2id-aes-256-gcm',
    salt BYTEA NOT NULL,
    key_hash TEXT NOT NULL,                 -- Hash of derived key for integrity check
    encrypted_key BYTEA,                     -- Key material (encrypted with password)
    params JSONB NOT NULL DEFAULT '{"memory":65536,"time":3,"threads":4,"saltLen":32,"keyLen":32}',
    rotated_at TIMESTAMPTZ DEFAULT NOW(),
    rotated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_key_version_per_user UNIQUE (user_id, version)
);

CREATE INDEX idx_master_key_versions_user ON master_key_versions(user_id, version DESC);

-- Update credentials table: add tags and metadata columns (retroactive)
ALTER TABLE credentials ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE credentials ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
ALTER TABLE credentials ADD COLUMN IF NOT EXISTS description TEXT;

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('006_credential_sharing', 'Add credential_shares, master_key_versions, and credential metadata')
ON CONFLICT (version) DO NOTHING;