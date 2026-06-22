-- Migration 009: API Keys Table
-- Adds API key support for machine-to-machine authentication

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    prefix VARCHAR(20) NOT NULL, -- e.g., "sk_live_abc1"
    hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 of full key
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_prefix_format CHECK (prefix ~ '^sk_(live|test)_[a-f0-9]{4}'),
    CONSTRAINT chk_name_length CHECK (LENGTH(name) BETWEEN 1 AND 100)
);

-- Indexes for fast lookups
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(hash) WHERE is_active = true;
CREATE INDEX idx_api_keys_active ON api_keys(user_id, is_active);

-- Unique active key names per user
CREATE UNIQUE INDEX idx_api_keys_user_name ON api_keys(user_id, name) WHERE is_active = true;

-- Auto-disable expired keys function
CREATE OR REPLACE FUNCTION disable_expired_api_keys()
RETURNS void AS $$
BEGIN
    UPDATE api_keys
    SET is_active = false
    WHERE is_active = true
      AND expires_at IS NOT NULL
      AND expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Run cleanup every 5 minutes via pg_cron (optional)
-- SELECT cron.schedule('api-key-cleanup', '*/5 * * * *', 'SELECT disable_expired_api_keys()');

COMMENT ON TABLE api_keys IS 'API keys for machine-to-machine authentication. Stores only SHA-256 hashes, never plaintext.';
