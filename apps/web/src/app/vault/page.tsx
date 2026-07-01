"use client";

import { useState, useEffect, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { MaterialIcon } from "@/components/ui/material-icon";
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

const typeIcon: Record<string, string> = {
  password: "password",
  ssh_key: "key",
  api_token: "code",
};

const typeTint: Record<string, string> = {
  password: "bg-primary-container text-on-primary-container",
  ssh_key: "bg-tertiary-container text-on-tertiary-container",
  api_token: "bg-secondary-container text-on-secondary-container",
};

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

  const handleCopy = (cred: VaultCredential) => {
    const value = revealId === cred.id && decrypted
      ? Object.values(decrypted).map((v) => (v ? String(v) : "—")).join("\n")
      : cred.name;
    navigator.clipboard?.writeText(value).then(() => toast.success("Copied to clipboard"));
  };

  return (
    <DashboardLayout>
      <div className="container mx-auto py-8 px-4 space-y-8">
        {/* Header */}
        <div className="flex items-end justify-between gap-4">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <MaterialIcon name="encrypted" className="text-primary text-[40px]" />
              <h1 className="text-headline-lg text-on-surface">Secure Vault</h1>
            </div>
            <p className="text-body-lg text-on-surface-variant">
              Manage and organize your encrypted credentials and private SSH keys.
            </p>
          </div>
          <div className="flex gap-3">
            {unlocked ? (
              <>
                <Button
                  variant="outline"
                  className="border-outline-variant hover:bg-surface-container-highest text-label-md"
                  onClick={() => lockMutation.mutate()}
                >
                  <MaterialIcon name="lock" size="sm" className="mr-2" /> Lock Vault
                </Button>
                <Button
                  className="bg-primary text-on-primary hover:opacity-90 active:scale-[0.98] text-label-md"
                  onClick={() => setShowAddDialog(true)}
                >
                  <MaterialIcon name="add" size="sm" className="mr-2" /> Add Credential
                </Button>
              </>
            ) : null}
          </div>
        </div>

        {/* Locked State */}
        {!unlocked ? (
          <div className="flex items-center justify-center min-h-[60vh]">
            <div className="glass-card p-xl rounded-xl w-full max-w-md flex flex-col items-center text-center shadow-2xl">
              <div className="w-20 h-20 bg-surface-container-high rounded-full flex items-center justify-center mb-lg">
                <MaterialIcon name="shield" className="text-primary text-[40px]" />
              </div>
              <h2 className="text-headline-sm text-on-surface mb-2">Vault Locked</h2>
              <p className="text-body-md text-on-surface-variant mb-xl">
                Your credentials are protected with AES-256 encryption. Enter your master password
                to decrypt and access the vault.
              </p>
              <form
                className="w-full space-y-4"
                onSubmit={(e) => {
                  e.preventDefault();
                  if (masterPassword) unlockMutation.mutate(masterPassword);
                }}
              >
                <Input
                  type="password"
                  placeholder="Master Password"
                  className="bg-surface-container-low border border-outline-variant text-on-surface placeholder:text-on-surface-variant focus:border-primary"
                  value={masterPassword}
                  onChange={(e) => setMasterPassword(e.target.value)}
                />
                <Button
                  type="submit"
                  className="w-full py-3 bg-primary text-on-primary hover:bg-primary-container font-bold"
                  disabled={!masterPassword || unlockMutation.isPending}
                >
                  <MaterialIcon name="lock_open" size="sm" className="mr-2" /> Unlock Vault
                </Button>
              </form>
              <div className="mt-xl flex items-center gap-2 text-label-md text-on-surface-variant">
                <MaterialIcon name="verified_user" size="sm" /> Hardware Security Module Active
              </div>
            </div>
          </div>
        ) : (
          <>
            {/* Toolbar */}
            <div className="relative">
              <MaterialIcon
                name="search"
                size="sm"
                className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
              />
              <Input
                placeholder="Search credentials..."
                className="pl-10 bg-surface-container-low border border-outline-variant text-on-surface placeholder:text-on-surface-variant"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>

            {/* Credential Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {credsLoading &&
                Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="glass-card rounded-xl p-5 flex flex-col gap-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-lg bg-surface-container-low animate-pulse" />
                        <div className="space-y-2">
                          <div className="h-4 w-32 rounded-md bg-surface-container-low animate-pulse" />
                          <div className="h-3 w-20 rounded-md bg-surface-container-low animate-pulse" />
                        </div>
                      </div>
                    </div>
                    <div className="h-3 w-full rounded-md bg-surface-container-low animate-pulse" />
                    <div className="h-3 w-2/3 rounded-md bg-surface-container-low animate-pulse" />
                    <div className="flex gap-2 mt-auto">
                      <div className="h-8 w-24 rounded-md bg-surface-container-low animate-pulse" />
                      <div className="h-8 w-24 rounded-md bg-surface-container-low animate-pulse" />
                    </div>
                  </div>
                ))}
              {!credsLoading &&
                filtered.map((cred) => (
                <div key={cred.id} className="glass-card rounded-xl p-5 flex flex-col gap-3 group">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div
                        className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                          typeTint[cred.type] ?? "bg-surface-container-high text-on-surface-variant"
                        }`}
                      >
                        <MaterialIcon name={typeIcon[cred.type] ?? "lock"} className="text-[24px]" />
                      </div>
                      <div>
                        <h3 className="text-body-lg font-bold text-on-surface truncate max-w-[180px]">
                          {cred.name}
                        </h3>
                        <p className="text-label-md text-on-surface-variant capitalize">{cred.type.replace("_", " ")}</p>
                      </div>
                    </div>
                    <Badge variant="outline" className="border-outline-variant text-on-surface-variant capitalize">
                      {cred.type.replace("_", " ")}
                    </Badge>
                  </div>

                  {cred.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {cred.tags.map((tag) => (
                        <Badge key={tag} variant="secondary" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  )}

                  {revealId === cred.id && decrypted && (
                    <div className="bg-surface-container-high rounded p-2 text-xs font-mono space-y-1">
                      {Object.entries(decrypted).map(([k, v]) => (
                        <div key={k}>
                          <span className="font-semibold">{k}: </span>
                          {v ? String(v).slice(0, 80) : "—"}
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Actions — reveal on hover */}
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity mt-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => handleCopy(cred)}
                    >
                      <MaterialIcon name="content_copy" size="sm" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => handleReveal(cred.id)}
                    >
                      <MaterialIcon name={revealId === cred.id ? "visibility_off" : "visibility"} size="sm" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={() => {
                        const data = prompt("Enter new secret value to rotate credential data:");
                        if (data) rotateMutation.mutate({ id: cred.id, data });
                      }}
                    >
                      <MaterialIcon name="autorenew" size="sm" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0 hover:bg-error/10"
                      onClick={() => {
                        if (confirm("Delete this credential?")) deleteMutation.mutate(cred.id);
                      }}
                    >
                      <MaterialIcon name="delete" size="sm" className="text-error" />
                    </Button>
                  </div>

                  {/* Footer */}
                  <div className="border-t border-outline-variant pt-3 mt-auto flex items-center justify-between text-label-md text-on-surface-variant">
                    <span>Last modified {new Date(cred.updated_at).toLocaleDateString()}</span>
                    {cred.expires_at && (
                      <span className="text-error">Expires {new Date(cred.expires_at).toLocaleDateString()}</span>
                    )}
                  </div>
                </div>
              ))}

              {filtered.length === 0 && !credsLoading && (
                <p className="text-on-surface-variant col-span-full">No credentials found.</p>
              )}
            </div>
          </>
        )}

        <AddCredentialDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
          onSubmit={(values) => createMutation.mutate(values)}
        />
      </div>
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
      Promise.resolve().then(() => {
        setName("");
        setType("password");
        setData("");
        setTags("");
        setExpiresAt("");
      });
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md bg-surface-container border-outline-variant">
        <DialogHeader>
          <DialogTitle className="text-on-surface">Add Credential</DialogTitle>
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
            <Label htmlFor="name" className="text-on-surface">Name</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="type" className="text-on-surface">Type</Label>
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
            <Label htmlFor="data" className="text-on-surface">Secret data</Label>
            <Input id="data" type="password" value={data} onChange={(e) => setData(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="tags" className="text-on-surface">Tags (comma separated)</Label>
            <Input id="tags" value={tags} onChange={(e) => setTags(e.target.value)} placeholder="prod, aws" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="expires" className="text-on-surface">Expires at (optional)</Label>
            <Input id="expires" type="datetime-local" value={expiresAt} onChange={(e) => setExpiresAt(e.target.value)} />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" className="border-outline-variant" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" className="bg-primary text-on-primary" disabled={!name || !data}>
              Save
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}