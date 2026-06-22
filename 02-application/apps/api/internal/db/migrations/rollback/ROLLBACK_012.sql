-- ROLLBACK_012.sql
-- Remove WireGuard tunnel support

DROP TRIGGER IF EXISTS update_wireguard_tunnels_updated_at ON wireguard_tunnels;
DROP TABLE IF EXISTS wireguard_tunnels;
DROP TYPE IF EXISTS tunnel_status;

ALTER TABLE hosts DROP COLUMN IF EXISTS wireguard_enabled;
ALTER TABLE hosts DROP COLUMN IF EXISTS wireguard_port;
