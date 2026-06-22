"use client";

import { useState, useEffect, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { Lock, Unlock, Plus, Search, Trash2, Eye, EyeOff, RotateCcw } from "lucide-react";
import { toast } from "sonner";

interface VaultCredential {
  id: string;
  name: string;
  type: string;
  tags: string[];
  version: number;
  last_rotated_at?: string | null;
  expires_at?: string | null;
  created_at: string;
  updated_at: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function apiFetch(path: string, options: RequestInit = {}) {
  const token = localStorage.getItem("access_token");
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
    credentials: "include",
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: "Request failed" }));
    throw new Error(err.message || `HTTP ${res.status}`);
  }
  return res.status === 204 ? null : res.json();
}

function useVaultStatus() {
  return useQuery({
    queryKey: ["vault-status"],
    queryFn: () => apiFetch("/api/v1/vault/status") as Promise<{ unlocked: boolean; key_version: number }>,
    staleTime: 30000,
  });
}

function useVaultCredentials(enabled: boolean) {
  return useQuery({
    queryKey: ["vault-credentials"],
    queryFn: () => apiFetch("/api/v1/vault/credentials") as Promise<{ credentials: VaultCredential[]; total: number }>,
    enabled,
    staleTime: 30000,
  });
}

export default function VaultPage() {
  const queryClient = useQueryClient();
  const [masterPassword, setMasterPassword] = useState("");
  const [search, setSearch] = useState("");
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [revealId, setRevealId] = useState<string | null>(null);
  const [decrypted, setDecrypted] = useState<Record<string, unknown> | null>(null);

  const { data: status, refetch: refetchStatus } = useVaultStatus();
  const unlocked = status?.unlocked ?? false;

  const { data: credsData, isLoading: credsLoading } = useVaultCredentials(unlocked);
  const credentials = credsData?.credentials ?? [];

  const unlockMutation = useMutation({
    mutationFn: (password: string) => apiFetch("/api/v1/vault/unlock", { method: "POST", body: JSON.stringify({ master_password: password }) }),
    onSuccess: () => {
      toast.success("Vault unlocked");
      setMasterPassword("");
      refetchStatus();
      queryClient.invalidateQueries({ queryKey: ["vault-credentials"] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const lockMutation = useMutation({
    mutationFn: () => apiFetch("/api/v1/vault/lock", { method: "POST" }),
    onSuccess: () => {
      toast.success("Vault locked");
      refetchStatus();
      queryClient.invalidateQueries({ queryKey: ["vault-credentials"] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const createMutation = useMutation({
    mutationFn: (payload: { name: string; type: string; data: string; tags?: string[]; expires_at?: string }) =>
      apiFetch("/api/v1/vault/credentials", { method: "POST", body: JSON.stringify(payload) }),
    onSuccess: () => {
      toast.success("Credential created");
      setShowAddDialog(false);
      queryClient.invalidateQueries({ queryKey: ["vault-credentials"] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiFetch(`/api/v1/vault/credentials/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      toast.success("Credential deleted");
      queryClient.invalidateQueries({ queryKey: ["vault-credentials"] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const rotateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: string }) =>
      apiFetch(`/api/v1/vault/credentials/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ data, name: credentials.find((c) => c.id === id)?.name }),
      }),
    onSuccess: () => {
      toast.success("Credential rotated");
      queryClient.invalidateQueries({ queryKey: ["vault-credentials"] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const filtered = useMemo(() => {
    const q = search.toLowerCase();
    return credentials.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.type.toLowerCase().includes(q) ||
        c.tags.some((t) => t.toLowerCase().includes(q))
    );
  }, [credentials, search]);

  const handleReveal = async (id: string) => {
    if (revealId === id) {
      setRevealId(null);
      setDecrypted(null);
      return;
    }
    try {
      const data = (await apiFetch(`/api/v1/vault/credentials/${id}/decrypt`)) as Record<string, unknown>;
      setRevealId(id);
      setDecrypted(data);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to decrypt");
    }
  };

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Lock className="h-6 w-6" />
              Vault
            </h1>
            <p className="text-muted-foreground">Manage encrypted credentials</p>
          </div>
          <div className="flex items-center gap-2">
            {unlocked ? (
              <Button variant="outline" size="sm" onClick={() => lockMutation.mutate()}>
                <Lock className="h-4 w-4 mr-1" /> Lock Vault
              </Button>
            ) : (
              <div className="flex items-center gap-2">
                <Input
                  type="password"
                  placeholder="Master password"
                  className="w-48"
                  value={masterPassword}
                  onChange={(e) => setMasterPassword(e.target.value)}
                />
                <Button size="sm" onClick={() => unlockMutation.mutate(masterPassword)} disabled={!masterPassword}>
                  <Unlock className="h-4 w-4 mr-1" /> Unlock
                </Button>
              </div>
            )}
            <Button size="sm" disabled={!unlocked} onClick={() => setShowAddDialog(true)}>
              <Plus className="h-4 w-4 mr-1" /> Add
            </Button>
          </div>
        </div>

        {!unlocked && (
          <Card className="bg-yellow-50 dark:bg-yellow-950/20">
            <CardContent className="py-6">
              <p className="text-sm text-muted-foreground">Vault is locked. Enter master password to view and manage credentials.</p>
            </CardContent>
          </Card>
        )}

        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search credentials..."
            className="pl-10"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            disabled={!unlocked}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {credsLoading && unlocked && (
            <p className="text-muted-foreground col-span-full">Loading credentials...</p>
          )}
          {filtered.map((cred) => (
            <Card key={cred.id}>
              <CardHeader className="pb-2">
                <CardTitle className="text-base flex items-center justify-between">
                  <span className="truncate">{cred.name}</span>
                  <Badge variant="outline">{cred.type}</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="flex flex-wrap gap-1">
                  {cred.tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                  ))}
                </div>
                <p className="text-xs text-muted-foreground">Updated {new Date(cred.updated_at).toLocaleString()}</p>
                {cred.expires_at && (
                  <p className="text-xs text-destructive">Expires {new Date(cred.expires_at).toLocaleString()}</p>
                )}
                {revealId === cred.id && decrypted && (
                  <div className="bg-muted rounded p-2 text-xs font-mono space-y-1">
                    {Object.entries(decrypted).map(([k, v]) => (
                      <div key={k}>
                        <span className="font-semibold">{k}: </span>
                        {v ? String(v).slice(0, 80) : "—"}
                      </div>
                    ))}
                  </div>
                )}
                <div className="flex items-center gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => handleReveal(cred.id)}>
                    {revealId === cred.id ? <EyeOff className="h-4 w-4 mr-1" /> : <Eye className="h-4 w-4 mr-1" />}
                    {revealId === cred.id ? "Hide" : "Reveal"}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      const data = prompt("Enter new secret value to rotate credential data:");
                      if (data) rotateMutation.mutate({ id: cred.id, data });
                    }}
                  >
                    <RotateCcw className="h-4 w-4 mr-1" /> Rotate
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => {
                      if (confirm("Delete this credential?")) deleteMutation.mutate(cred.id);
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
          {filtered.length === 0 && unlocked && !credsLoading && (
            <p className="text-muted-foreground col-span-full">No credentials found.</p>
          )}
        </div>
      </div>

      <AddCredentialDialog
        open={showAddDialog}
        onOpenChange={setShowAddDialog}
        onSubmit={(values) => createMutation.mutate(values)}
      />
    </DashboardLayout>
  );
}

function AddCredentialDialog({
  open,
  onOpenChange,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  onSubmit: (values: { name: string; type: string; data: string; tags?: string[]; expires_at?: string }) => void;
}) {
  const [name, setName] = useState("");
  const [type, setType] = useState("password");
  const [data, setData] = useState("");
  const [tags, setTags] = useState("");
  const [expiresAt, setExpiresAt] = useState("");

  useEffect(() => {
    if (!open) {
      setName("");
      setType("password");
      setData("");
      setTags("");
      setExpiresAt("");
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add Credential</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit({
              name,
              type,
              data,
              tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
              expires_at: expiresAt || undefined,
            });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="type">Type</Label>
            <Select value={type} onValueChange={setType}>
              <SelectTrigger id="type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="password">Password</SelectItem>
                <SelectItem value="ssh_key">SSH Key</SelectItem>
                <SelectItem value="api_token">API Token</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="data">Secret data</Label>
            <Input id="data" type="password" value={data} onChange={(e) => setData(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="tags">Tags (comma separated)</Label>
            <Input id="tags" value={tags} onChange={(e) => setTags(e.target.value)} placeholder="prod, aws"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="expires">Expires at (optional)</Label>
            <Input id="expires" type="datetime-local" value={expiresAt} onChange={(e) => setExpiresAt(e.target.value)} />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={!name || !data}>Save</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
