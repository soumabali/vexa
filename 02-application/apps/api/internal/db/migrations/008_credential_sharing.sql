-- Migration 008: Credential Sharing
-- Adds support for E2E encrypted credential sharing between users
-- Run order: 008 (after 006)
-- Rollback: ROLLBACK_008.sql

-- ============================================================
-- Tables for Credential Sharing
-- ============================================================

-- User share key pairs (derived from vault key)
CREATE TABLE IF NOT EXISTS user_share_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_key      TEXT NOT NULL, -- base64-encoded ECDH P-256 public key
    private_key     TEXT NOT NULL, -- base64-encoded ECDH P-256 private key (encrypted with vault key)
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    
    CONSTRAINT unique_user_share_keys_user UNIQUE (user_id)
);

-- Ensure columns required by the E2E sharing model exist.
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS encrypted_data TEXT;
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS expiry_time TIMESTAMPTZ;
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'accepted'
    CHECK (status IN ('pending', 'accepted', 'rejected', 'revoked', 'expired'));
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ;
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

-- Credential share invitations/shares (006 created a base credential_shares table)
-- Add sender_id to align with the E2E sharing model.
ALTER TABLE credential_shares ADD COLUMN IF NOT EXISTS sender_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE credential_shares SET sender_id = owner_id WHERE sender_id IS NULL AND owner_id IS NOT NULL;
ALTER TABLE credential_shares ALTER COLUMN sender_id SET NOT NULL;
ALTER TABLE credential_shares ADD CONSTRAINT no_self_share CHECK (sender_id != recipient_id);

-- Ensure permission allows E2E sharing roles if migrating from 006
ALTER TABLE credential_shares DROP CONSTRAINT IF EXISTS credential_shares_permission_check;
ALTER TABLE credential_shares ADD CONSTRAINT credential_shares_permission_check
    CHECK (permission IN ('read_only', 'read_write', 'admin'));

-- Backfill status for existing rows
UPDATE credential_shares SET status = 'accepted' WHERE status IS NULL;
ALTER TABLE credential_shares ALTER COLUMN status SET NOT NULL;

-- Audit trail for sharing events
CREATE TABLE IF NOT EXISTS share_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    share_id        UUID NOT NULL REFERENCES credential_shares(id) ON DELETE CASCADE,
    action          VARCHAR(20) NOT NULL
        CHECK (action IN ('created', 'accepted', 'rejected', 'revoked', 'accessed', 'modified', 'expired')),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    permission      VARCHAR(20),
    ip_address      INET,
    user_agent      TEXT,
    timestamp       TIMESTAMPTZ DEFAULT now(),
    success         BOOLEAN NOT NULL DEFAULT true,
    error_message   TEXT,
    
    -- Integrity chain (tamper-evident)
    prev_hash       TEXT,
    current_hash    TEXT NOT NULL
);

-- ============================================================
-- Indexes
-- ============================================================

-- User share keys
CREATE INDEX IF NOT EXISTS idx_user_share_keys_user 
    ON user_share_keys(user_id);

-- Credential shares - common query patterns
CREATE INDEX IF NOT EXISTS idx_credential_shares_sender 
    ON credential_shares(sender_id, status);
CREATE INDEX IF NOT EXISTS idx_credential_shares_recipient 
    ON credential_shares(recipient_id, status);
CREATE INDEX IF NOT EXISTS idx_credential_shares_credential 
    ON credential_shares(credential_id);
CREATE INDEX IF NOT EXISTS idx_credential_shares_status 
    ON credential_shares(status);
CREATE INDEX IF NOT EXISTS idx_credential_shares_expiry 
    ON credential_shares(expiry_time) 
    WHERE status IN ('pending', 'accepted');

-- Unique constraint: one active share per credential-recipient pair
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_share 
    ON credential_shares(credential_id, recipient_id) 
    WHERE status = 'accepted' AND revoked_at IS NULL;

-- Share audit log
CREATE INDEX IF NOT EXISTS idx_share_audit_share 
    ON share_audit_log(share_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_share_audit_user 
    ON share_audit_log(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_share_audit_timestamp 
    ON share_audit_log(timestamp DESC);

-- ============================================================
-- Row-Level Security Policies
-- ============================================================

-- User share keys: only owner can see their keys
ALTER TABLE user_share_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_share_keys_owner ON user_share_keys
    FOR ALL
    USING (user_id = current_setting('app.current_user_id')::UUID);

-- Credential shares: sender and recipient can view
ALTER TABLE credential_shares ENABLE ROW LEVEL SECURITY;

CREATE POLICY credential_shares_view ON credential_shares
    FOR SELECT
    USING (
        sender_id = current_setting('app.current_user_id')::UUID
        OR recipient_id = current_setting('app.current_user_id')::UUID
    );

-- Only sender can insert/update shares
CREATE POLICY credential_shares_modify ON credential_shares
    FOR INSERT
    WITH CHECK (sender_id = current_setting('app.current_user_id')::UUID);

CREATE POLICY credential_shares_update ON credential_shares
    FOR UPDATE
    USING (sender_id = current_setting('app.current_user_id')::UUID)
    WITH CHECK (sender_id = current_setting('app.current_user_id')::UUID);

-- Share audit log: admins can view all, users can view their own
ALTER TABLE share_audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY share_audit_owner ON share_audit_log
    FOR SELECT
    USING (
        user_id = current_setting('app.current_user_id')::UUID
        OR EXISTS (
            SELECT 1 FROM users 
            WHERE id = current_setting('app.current_user_id')::UUID 
            AND role = 'admin'
        )
    );

-- ============================================================
-- Functions
-- ============================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_share_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update timestamp
CREATE TRIGGER trigger_update_credential_shares
    BEFORE UPDATE ON credential_shares
    FOR EACH ROW
    EXECUTE FUNCTION update_share_timestamp();

-- Function to check and auto-expire shares
CREATE OR REPLACE FUNCTION expire_shares()
RETURNS void AS $$
BEGIN
    UPDATE credential_shares
    SET status = 'expired',
        updated_at = now()
    WHERE status IN ('pending', 'accepted')
      AND expiry_time IS NOT NULL
      AND expiry_time < now();
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Cron job for periodic expiry check (optional)
-- ============================================================
-- In production, run this every minute via cron/pg_cron:
-- SELECT cron.schedule('expire-shares', '* * * * *', 'SELECT expire_shares()');

-- ============================================================
-- Comment documenting rollback
-- ============================================================
COMMENT ON TABLE credential_shares IS 
'E2E encrypted credential sharing. Rollback with ROLLBACK_008.sql';
