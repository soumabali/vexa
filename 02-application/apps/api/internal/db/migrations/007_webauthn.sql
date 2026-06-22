-- 007_webauthn.sql
-- WebAuthn / FIDO2 credential storage for passwordless authentication.
-- Supports platform authenticator (Touch ID, Windows Hello), roaming authenticator (YubiKey),
-- resident keys (discoverable credentials), backup eligibility, and attestation metadata.

CREATE TABLE webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id BYTEA NOT NULL,              -- Raw credential ID (unique per authenticator)
    public_key BYTEA NOT NULL,                 -- COSE public key bytes
    attestation_type TEXT,                     -- Attestation type (none, indirect, direct, enterprise)
    transport TEXT[] DEFAULT '{}',             -- Supported transports: usb, nfc, ble, hybrid, internal
    authenticator_aaguid BYTEA,                -- AAGUID of the authenticator
    authenticator_sign_count INT NOT NULL DEFAULT 0,
    authenticator_clone_warning BOOLEAN NOT NULL DEFAULT FALSE,
    authenticator_attachment TEXT,               -- platform | cross-platform
    is_resident_key BOOLEAN NOT NULL DEFAULT FALSE,
    is_backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    is_backed_up BOOLEAN NOT NULL DEFAULT FALSE,
    user_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name TEXT NOT NULL DEFAULT 'Unnamed Credential',
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_credential_id UNIQUE (credential_id)
);

CREATE INDEX idx_webauthn_credentials_user ON webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON webauthn_credentials USING HASH (credential_id);

-- Track pending WebAuthn registration challenges (ephemeral; production may use Redis)
CREATE TABLE webauthn_registration_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge BYTEA NOT NULL,
    session_data JSONB NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '5 minutes'),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_registration_session_user UNIQUE (user_id)
);

CREATE INDEX idx_webauthn_registration_sessions_expires ON webauthn_registration_sessions(expires_at);

-- Track pending WebAuthn login/assertion challenges for discoverable credential flows
CREATE TABLE webauthn_login_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge BYTEA NOT NULL,
    session_data JSONB NOT NULL DEFAULT '{}',
    allowed_credential_ids BYTEA[] DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '5 minutes'),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webauthn_login_sessions_expires ON webauthn_login_sessions(expires_at);

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('007_webauthn', 'Add WebAuthn/FIDO2 credential tables and session stores')
ON CONFLICT (version) DO NOTHING;
