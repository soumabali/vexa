import {
  CreateHostInput,
  UpdateHostInput,
  HostFilterInput,
  HostResponse,
  HostListResponse,
  SSHOptions,
  RDPOptions,
  VNCOptions,
} from "@/lib/validations/hosts";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";

// Helper to get auth token
function getAuthToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("access_token") || localStorage.getItem("token") || null;
  }
  return null;
}

// Map frontend host schema to backend API model
function toBackendHost(data: CreateHostInput | UpdateHostInput) {
  const payload: Record<string, unknown> = {};
  if (data.name !== undefined) payload.name = data.name;
  if (data.host !== undefined) payload.address = data.host;
  if (data.address !== undefined) payload.address = data.address;
  if (data.port !== undefined) payload.port = data.port;
  if (data.hostType !== undefined) payload.protocol = data.hostType;
  if (data.type !== undefined) payload.protocol = data.type;
  if (data.tags !== undefined) payload.tags = data.tags;
  if (data.description !== undefined) payload.description = data.description;
  if (data.groupId !== undefined) payload.group_path = data.groupId;
  if (data.favorite !== undefined) payload.is_active = data.favorite !== undefined ? true : undefined;
  return payload;
}

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
    credentials: "include",
    ...options,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({
      message: "An error occurred",
    }));
    throw new Error(error.message || `HTTP error! status: ${response.status}`);
  }

  const contentLength = response.headers.get("content-length");
  if (response.status === 204 || contentLength === "0") {
    return {} as T;
  }

  return response.json();
}

export interface HostStats {
  total: number;
  ssh: number;
  rdp: number;
  vnc: number;
  favorites: number;
  online: number;
  offline: number;
}

export const hostsApi = {
  // List hosts with filtering, sorting, pagination
  list: async (filters?: HostFilterInput): Promise<HostListResponse> => {
    const resp = await apiRequest<HostListResponse>("/api/v1/hosts", { method: "GET" });
    const hosts = (resp.hosts || resp.data || []).map((h) => ({
      ...h,
      host: h.host ?? h.address ?? "",
      hostType: h.hostType ?? (h as { protocol?: string }).protocol ?? "ssh",
      status: h.status ?? "unknown",
    }));
    return { ...resp, hosts };
  },

  // Get single host
  get: (id: string): Promise<HostResponse> =>
    apiRequest<HostResponse>(`/api/v1/hosts/${id}`),

  // Create host
  create: (data: CreateHostInput | Record<string, unknown>): Promise<HostResponse> => {
    const payload = toBackendHost(data as CreateHostInput);
    return apiRequest<HostResponse>("/api/v1/hosts", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  // Update host
  update: (id: string, data: UpdateHostInput | Record<string, unknown>): Promise<HostResponse> => {
    const payload = toBackendHost(data as UpdateHostInput);
    return apiRequest<HostResponse>(`/api/v1/hosts/${id}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
  },

  // Delete host
  delete: (id: string): Promise<{ message: string }> =>
    apiRequest<{ message: string }>(`/api/v1/hosts/${id}`, {
      method: "DELETE",
    }),

  // Bulk delete hosts
  bulkDelete: (ids: string[]): Promise<{ message: string; deleted: number }> =>
    apiRequest<{ message: string; deleted: number }>("/api/v1/hosts/bulk-delete", {
      method: "POST",
      body: JSON.stringify({ ids }),
    }),

  // Get host stats
  stats: (): Promise<HostStats> =>
    apiRequest<HostStats>("/api/v1/hosts/stats"),

  // Toggle favorite
  toggleFavorite: (id: string): Promise<HostResponse> =>
    apiRequest<HostResponse>(`/api/v1/hosts/${id}/favorite`, {
      method: "PATCH",
    }),

  // Connect to host (returns session info)
  connect: (id: string): Promise<{ sessionId: string; token: string }> =>
    apiRequest<{ sessionId: string; token: string }>(`/api/v1/hosts/${id}/connect`, {
      method: "POST",
    }),

  // Get host connection options
  getConnectionOptions: (
    id: string
  ): Promise<{
    hostType: string;
    sshOptions?: SSHOptions;
    rdpOptions?: RDPOptions;
    vncOptions?: VNCOptions;
  }> => apiRequest(`/api/v1/hosts/${id}/connect/options`),

  // Test host connectivity
  testConnection: (id: string): Promise<{ online: boolean; latency?: number }> =>
    apiRequest<{ online: boolean; latency?: number }>(`/api/v1/hosts/${id}/test`),

  // Import hosts
  import: (
    hosts: CreateHostInput[],
    overwrite = false
  ): Promise<{ imported: number; skipped: number }> =>
    apiRequest<{ imported: number; skipped: number }>("/api/v1/hosts/import", {
      method: "POST",
      body: JSON.stringify({ hosts, overwrite }),
    }),

  // Export hosts
  export: (
    filter?: HostFilterInput
  ): Promise<HostResponse[]> =>
    apiRequest<HostResponse[]>("/api/v1/hosts/export", {
      method: "POST",
      body: JSON.stringify(filter || {}),
    }),

  // Get host groups
  getGroups: (): Promise<
    Array<{ id: string; name: string; color?: string; count: number }>
  > => apiRequest("/api/v1/hosts/groups"),

  // Create host group
  createGroup: (data: {
    name: string;
    color?: string;
  }): Promise<{ id: string; name: string; color?: string }> =>
    apiRequest("/api/v1/hosts/groups", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Update host group
  updateGroup: (
    id: string,
    data: { name?: string; color?: string }
  ): Promise<{ id: string; name: string; color?: string }> =>
    apiRequest(`/api/v1/hosts/groups/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  // Delete host group
  deleteGroup: (id: string): Promise<{ message: string }> =>
    apiRequest(`/api/v1/hosts/groups/${id}`, {
      method: "DELETE",
    }),
};
