-- 012_wireguard_tunnels.sql
-- WireGuard tunnel configurations per user/host

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE tunnel_status AS ENUM (
    'creating',
    'connected',
    'disconnected',
    'error',
    'disabled'
);

CREATE TABLE wireguard_tunnels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    interface_name VARCHAR(16) NOT NULL,
    server_private_key TEXT NOT NULL,
    server_public_key TEXT NOT NULL,
    client_public_key TEXT NOT NULL,
    preshared_key TEXT,
    server_ip INET NOT NULL,
    client_ip INET NOT NULL,
    listen_port INT NOT NULL CHECK (listen_port BETWEEN 1 AND 65535),
    allowed_ips CIDR[] DEFAULT '{0.0.0.0/0}'::CIDR[],
    dns_servers INET[] DEFAULT '{}'::INET[],
    mtu INT DEFAULT 1420,
    persistent_keepalive INT DEFAULT 25,
    status tunnel_status DEFAULT 'creating',
    is_enabled BOOLEAN DEFAULT false,
    last_handshake_at TIMESTAMPTZ,
    bytes_sent BIGINT DEFAULT 0,
    bytes_received BIGINT DEFAULT 0,
    last_rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT unique_user_tunnel_per_host UNIQUE (user_id, host_id),
    CONSTRAINT unique_interface_name UNIQUE (interface_name)
);

CREATE INDEX idx_wg_tunnels_user ON wireguard_tunnels(user_id);
CREATE INDEX idx_wg_tunnels_host ON wireguard_tunnels(host_id);
CREATE INDEX idx_wg_tunnels_enabled ON wireguard_tunnels(is_enabled);
CREATE INDEX idx_wg_tunnels_status ON wireguard_tunnels(status);

-- Add wireguard fields to hosts table
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS wireguard_enabled BOOLEAN DEFAULT false;
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS wireguard_port INT DEFAULT 51820;

-- WireGuard updated-at trigger
CREATE TRIGGER update_wireguard_tunnels_updated_at BEFORE UPDATE ON wireguard_tunnels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
