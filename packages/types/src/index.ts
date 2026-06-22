export interface User {
  id: string;
  email: string;
  name: string;
  role: "admin" | "user";
  createdAt: string;
  updatedAt: string;
}

export interface Host {
  id: string;
  name: string;
  hostname: string;
  port: number;
  protocol: "ssh" | "telnet" | "rdp" | "vnc";
  username?: string;
  tags?: string[];
  createdAt: string;
  updatedAt: string;
}

export interface Credential {
  id: string;
  hostId: string;
  type: "password" | "key" | "otp";
  label: string;
  createdAt: string;
}

export interface Session {
  id: string;
  hostId: string;
  userId: string;
  status: "active" | "closed" | "error";
  startedAt: string;
  endedAt?: string;
}

export interface AuditEvent {
  id: string;
  userId: string;
  action: string;
  resource: string;
  timestamp: string;
  details?: Record<string, unknown>;
}

export interface TerminalSize {
  rows: number;
  cols: number;
}

export interface GatewayConfig {
  protocol: "ssh" | "telnet" | "rdp" | "vnc";
  host: string;
  port: number;
  credentials: {
    username: string;
    password?: string;
    privateKey?: string;
  };
}
