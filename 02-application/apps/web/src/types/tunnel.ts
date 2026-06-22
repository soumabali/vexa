export interface Tunnel {
  id: string;
  name: string;
  hostId: string;
  host_id: string;
  interface_name: string;
  client_ip: string;
  listen_port: number;
  is_enabled: boolean;
  enabled: boolean;
  publicKey: string;
  server_public_key: string;
  allowed_ips: string[];
  last_rotated_at: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface TunnelCreateInput {
  name?: string;
  hostId?: string;
  host_id?: string;
  allowed_ips?: string[];
  listen_port?: number;
  dns_servers?: string[];
  use_psk?: boolean;
  mtu?: number;
}

export interface TunnelUpdateInput {
  name?: string;
  enabled?: boolean;
}

export interface TunnelStats {
  bytes_received: number;
  bytes_sent: number;
  last_handshake: string | null;
}

export interface TunnelConfig {
  interface: {
    private_key: string;
    address: string;
    dns: string[];
  };
  peer: {
    public_key: string;
    preshared_key?: string;
    allowed_ips: string[];
    endpoint: string;
    persistent_keepalive: number;
  };
}
