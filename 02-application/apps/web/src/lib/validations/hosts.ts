import { z } from "zod";

export enum HostType {
  SSH = "ssh",
  RDP = "rdp",
  VNC = "vnc",
}

const hostTypeSchema = z
  .union([z.nativeEnum(HostType), z.enum(["ssh", "rdp", "vnc"])])
  .transform((v) => (typeof v === "string" ? v : (HostType[v] as unknown as string)));

export const createHostSchema = z.object({
  name: z.string().min(1).max(120),
  host: z.string().min(1).max(255),
  address: z.string().min(1).max(255).optional(),
  port: z.coerce.number().int().min(1).max(65535).default(22),
  type: hostTypeSchema.default("ssh"),
  hostType: hostTypeSchema.default("ssh"),
  username: z.string().optional(),
  labels: z.array(z.string()).optional(),
  tags: z.array(z.string()).optional(),
  description: z.string().optional(),
  color: z.string().optional(),
  favorite: z.boolean().optional(),
  groupId: z.string().optional(),

  sshOptions: z
    .object({
      identityFile: z.string().optional(),
      jumpHost: z.string().optional(),
      keepAlive: z.boolean().optional(),
      compress: z.boolean().optional(),
      forwardingAgent: z.boolean().optional(),
      x11Forwarding: z.boolean().optional(),
    })
    .optional(),

  rdpOptions: z
    .object({
      domain: z.string().optional(),
      width: z.number().optional(),
      height: z.number().optional(),
      colorDepth: z.enum(["16", "24", "32"]).default("32").optional(),
      redirectDrives: z.boolean().optional(),
      redirectPrinters: z.boolean().optional(),
      redirectClipboard: z.boolean().optional(),
      clipboardRedirect: z.boolean().optional(),
      driveRedirect: z.boolean().optional(),
    })
    .optional(),

  vncOptions: z
    .object({
      quality: z.enum(["low", "medium", "high", "auto"]).default("auto").optional(),
      compression: z.enum(["low", "medium", "high", "auto"]).default("auto").optional(),
      colorDepth: z.enum(["8", "16", "24", "32"]).default("32").optional(),
      viewOnly: z.boolean().optional(),
      shared: z.boolean().optional(),
    })
    .optional(),
});

export const updateHostSchema = createHostSchema.partial();

export type CreateHostInput = z.infer<typeof createHostSchema>;
export type UpdateHostInput = z.infer<typeof updateHostSchema>;

export const hostFilterSchema = z.object({
  type: z.nativeEnum(HostType).optional(),
  status: z.enum(["online", "offline", "unknown"]).optional(),
  favorite: z.boolean().optional(),
  groupId: z.string().optional(),
  tags: z.array(z.string()).optional(),
  search: z.string().optional(),
  sortBy: z.enum(["name", "createdAt", "lastConnected"]).optional(),
  sortOrder: z.enum(["asc", "desc"]).optional(),
  page: z.coerce.number().int().min(1).optional(),
  limit: z.coerce.number().int().min(1).max(1000).optional(),
});

export type HostFilterInput = z.infer<typeof hostFilterSchema>;

export interface HostListResponse {
  hosts?: HostResponse[];
  data?: HostResponse[];
  total: number;
  page: number;
  limit: number;
}

export interface SSHOptions {
  identityFile?: string;
  jumpHost?: string;
  keepAlive?: boolean;
  compress?: boolean;
  forwardingAgent?: boolean;
  x11Forwarding?: boolean;
}

export interface RDPOptions {
  domain?: string;
  width?: number;
  height?: number;
  colorDepth?: "16" | "24" | "32";
  redirectDrives?: boolean;
  redirectPrinters?: boolean;
  redirectClipboard?: boolean;
  clipboardRedirect?: boolean;
  driveRedirect?: boolean;
}

export interface VNCOptions {
  quality?: "low" | "medium" | "high" | "auto";
  compression?: "low" | "medium" | "high" | "auto";
  colorDepth?: "8" | "16" | "24" | "32";
  viewOnly?: boolean;
  shared?: boolean;
}

export interface HostResponse {
  id: string;
  name: string;
  address: string;
  host: string;
  hostType: "ssh" | "rdp" | "vnc";
  port: number;
  type: HostType;
  username?: string;
  labels?: string[];
  tags?: string[];
  color?: string;
  favorite?: boolean;
  status?: "online" | "offline" | "unknown";
  description?: string;
  groupName?: string;
  groupId?: string;
  lastConnected?: string;
  sshOptions?: SSHOptions;
  rdpOptions?: RDPOptions;
  vncOptions?: VNCOptions;
  createdAt: string;
  updatedAt: string;
}

export const hostSchema = createHostSchema;
export type HostInput = CreateHostInput;
