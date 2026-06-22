import { apiRequest } from "./client";
import type {
  Tunnel,
  TunnelCreateInput,
  TunnelUpdateInput,
  TunnelStats,
  TunnelConfig,
} from "@/types/tunnel"

export async function listTunnels(): Promise<Tunnel[]> {
  const resp = await apiRequest<{ tunnels: Tunnel[] }>("/api/v1/tunnels", { method: "GET" });
  return resp.tunnels ?? [];
}

export async function getTunnel(id: string): Promise<Tunnel> {
  return apiRequest<Tunnel>(`/api/v1/tunnels/${id}`, { method: "GET" })
}

export async function createTunnel(input: TunnelCreateInput): Promise<Tunnel> {
  return apiRequest<Tunnel>("/api/v1/tunnels", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

export async function updateTunnel(id: string, input: TunnelUpdateInput): Promise<Tunnel> {
  return apiRequest<Tunnel>(`/api/v1/tunnels/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  })
}

export async function deleteTunnel(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/tunnels/${id}`, { method: "DELETE" })
}

export async function rotateTunnelKeys(id: string): Promise<Tunnel> {
  return apiRequest<Tunnel>(`/api/v1/tunnels/${id}/rotate`, { method: "POST" })
}

export async function getTunnelConfig(id: string): Promise<TunnelConfig> {
  return apiRequest<TunnelConfig>(`/api/v1/tunnels/${id}/config`, { method: "GET" })
}

export async function enableTunnel(id: string): Promise<Tunnel> {
  return apiRequest<Tunnel>(`/api/v1/tunnels/${id}/enable`, { method: "POST" })
}

export async function disableTunnel(id: string): Promise<Tunnel> {
  return apiRequest<Tunnel>(`/api/v1/tunnels/${id}/disable`, { method: "POST" })
}

export async function getTunnelStats(id: string): Promise<TunnelStats> {
  return apiRequest<TunnelStats>(`/api/v1/tunnels/${id}/stats`, { method: "GET" })
}
