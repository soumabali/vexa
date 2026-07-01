"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MaterialIcon } from "@/components/ui/material-icon";
import { toast } from "sonner";

interface APIKey {
  id: string;
  name: string;
  key: string;
  prefix: string;
  scopes: string[];
  createdAt: Date;
  lastUsed?: Date;
  expiresAt?: Date;
  isRevoked: boolean;
}

export default function APIKeysPage() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([
    {
      id: "1",
      name: "Production API Key",
      key: "ssh_prod_abc123def456",
      prefix: "ssh_prod",
      scopes: ["hosts:read", "hosts:write", "sessions:read"],
      createdAt: new Date("2026-01-01"),
      lastUsed: new Date("2026-05-28"),
      expiresAt: new Date("2026-12-31"),
      isRevoked: false,
    },
    {
      id: "2",
      name: "CI/CD Key",
      key: "ssh_ci_xyz789uvw456",
      prefix: "ssh_ci",
      scopes: ["hosts:read", "vault:read"],
      createdAt: new Date("2026-03-01"),
      lastUsed: new Date("2026-05-27"),
      isRevoked: false,
    },
    {
      id: "3",
      name: "Old Key",
      key: "ssh_old_def789ghi012",
      prefix: "ssh_old",
      scopes: ["hosts:read"],
      createdAt: new Date("2026-01-15"),
      isRevoked: true,
    },
  ]);

  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showKeyDialog, setShowKeyDialog] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [newKeyScopes, setNewKeyScopes] = useState<string[]>(["hosts:read"]);
  const [newKeyExpiry, setNewKeyExpiry] = useState("30d");
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null);

  const availableScopes = [
    { value: "hosts:read", label: "Read Hosts", description: "View host configurations" },
    { value: "hosts:write", label: "Write Hosts", description: "Create and modify hosts" },
    { value: "sessions:read", label: "Read Sessions", description: "View session history" },
    { value: "sessions:write", label: "Write Sessions", description: "Create new sessions" },
    { value: "vault:read", label: "Read Vault", description: "View credentials" },
    { value: "vault:write", label: "Write Vault", description: "Manage credentials" },
    { value: "users:read", label: "Read Users", description: "View user list" },
    { value: "admin", label: "Admin", description: "Full access" },
  ];

  const handleCreateKey = () => {
    const key = `ssh_${Math.random().toString(36).substring(7)}_${Math.random().toString(36).substring(7)}`;
    const newKey: APIKey = {
      id: Math.random().toString(36).substring(7),
      name: newKeyName,
      key,
      prefix: key.split("_").slice(0, 2).join("_"),
      scopes: newKeyScopes,
      createdAt: new Date(),
      isRevoked: false,
    };
    setApiKeys([...apiKeys, newKey]);
    setNewlyCreatedKey(key);
    setShowCreateDialog(false);
    setShowKeyDialog(true);
    setNewKeyName("");
    setNewKeyScopes(["hosts:read"]);
    setNewKeyExpiry("30d");
  };

  const handleRevokeKey = (id: string) => {
    setApiKeys(
      apiKeys.map((k) => (k.id === id ? { ...k, isRevoked: true } : k))
    );
    toast.success("API key revoked");
  };

  const handleRotateKey = (id: string) => {
    const rotated = `ssh_${Math.random().toString(36).substring(7)}_${Math.random().toString(36).substring(7)}`;
    setApiKeys(
      apiKeys.map((k) =>
        k.id === id
          ? { ...k, key: rotated, prefix: rotated.split("_").slice(0, 2).join("_"), createdAt: new Date() }
          : k
      )
    );
    toast.success("API key rotated");
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    toast.success("Copied to clipboard");
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-5xl">
      <div className="mb-8">
        <h1 className="text-headline-lg text-on-surface mb-2">API Keys</h1>
        <p className="text-body-lg text-on-surface-variant">
          Manage programmatic access third-party integrations.
        </p>
      </div>

      <div className="flex justify-end mb-6">
        <Button onClick={() => setShowCreateDialog(true)} className="bg-primary text-on-primary hover:bg-primary/90">
          <MaterialIcon name="add" size="sm" className="mr-2" />
          Create New Key
        </Button>
      </div>

      {apiKeys.length === 0 ? (
        <Card className="bg-surface-container border border-outline-variant rounded-xl">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 flex items-center justify-center rounded-full bg-surface-container-high text-on-surface-variant mb-4">
              <MaterialIcon name="key" className="text-[32px]" />
            </div>
            <h3 className="text-on-surface font-semibold mb-1">No API keys</h3>
            <p className="text-on-surface-variant text-sm mb-4">Create your first key to enable programmatic access.</p>
            <Button onClick={() => setShowCreateDialog(true)} className="bg-primary text-on-primary hover:bg-primary/90">
              <MaterialIcon name="add" size="sm" className="mr-2" />
              Create your first key
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Card className="bg-surface-container rounded-xl border border-outline-variant overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-surface-container-high border-b border-outline-variant">
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Name</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Key</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Created</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Last Used</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Scopes</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Status</th>
                  <th className="text-label-md text-on-surface-variant uppercase tracking-wide text-left px-4 py-3 font-semibold">Actions</th>
                </tr>
              </thead>
              <tbody>
                {apiKeys.map((apiKey) => (
                  <tr
                    key={apiKey.id}
                    className={`border-b border-outline-variant hover:bg-surface-variant/30 ${apiKey.isRevoked ? "opacity-50" : ""}`}
                  >
                    <td className="px-4 py-3 text-on-surface font-medium">{apiKey.name}</td>
                    <td className="px-4 py-3">
                      <code className="font-mono text-sm bg-surface-container-lowest border border-outline-variant text-on-surface-variant px-2 py-1 rounded">
                        {apiKey.isRevoked
                          ? "••••••••••••"
                          : `${apiKey.prefix}_••••••${apiKey.key.slice(-4)}`}
                      </code>
                    </td>
                    <td className="px-4 py-3 text-on-surface-variant text-sm">{apiKey.createdAt.toLocaleDateString()}</td>
                    <td className="px-4 py-3 text-on-surface-variant text-sm">
                      {apiKey.lastUsed ? apiKey.lastUsed.toLocaleDateString() : "Never"}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {apiKey.scopes.map((scope) => (
                          <Badge key={scope} variant="outline" className="text-xs border-outline-variant text-on-surface-variant">
                            {scope}
                          </Badge>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {apiKey.isRevoked ? (
                        <Badge className="bg-error/10 text-error border border-error/30">Revoked</Badge>
                      ) : (
                        <Badge className="bg-success/10 text-success border border-success/30">Active</Badge>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleCopyKey(apiKey.key)}
                          className="text-on-surface-variant hover:text-on-surface"
                          aria-label="Copy key"
                        >
                          <MaterialIcon name="content_copy" size="sm" />
                        </Button>
                        {!apiKey.isRevoked && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRotateKey(apiKey.id)}
                            className="text-on-surface-variant hover:text-on-surface"
                            aria-label="Rotate key"
                          >
                            <MaterialIcon name="refresh" size="sm" />
                          </Button>
                        )}
                        {!apiKey.isRevoked && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRevokeKey(apiKey.id)}
                            className="text-error hover:text-error"
                            aria-label="Revoke key"
                          >
                            <MaterialIcon name="delete_forever" size="sm" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Create Key Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-lg bg-surface-container-high border-outline-variant">
          <DialogHeader>
            <DialogTitle className="text-on-surface">Create New API Key</DialogTitle>
            <DialogDescription className="text-on-surface-variant">
              Create a new API key with specific permissions
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="keyName" className="text-on-surface">Key Name</Label>
              <Input
                id="keyName"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="e.g., Production API Key"
                className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
              />
            </div>

            <div className="space-y-2">
              <Label className="text-on-surface">Scopes</Label>
              <div className="space-y-2">
                {availableScopes.map((scope) => (
                  <div
                    key={scope.value}
                    className={`flex items-center justify-between p-3 border rounded-lg cursor-pointer transition-colors ${
                      newKeyScopes.includes(scope.value)
                        ? "border-primary bg-primary/10"
                        : "border-outline-variant hover:bg-surface-container-lowest"
                    }`}
                    onClick={() => {
                      if (newKeyScopes.includes(scope.value)) {
                        setNewKeyScopes(newKeyScopes.filter((s) => s !== scope.value));
                      } else {
                        setNewKeyScopes([...newKeyScopes, scope.value]);
                      }
                    }}
                  >
                    <div>
                      <p className="font-medium text-on-surface">{scope.label}</p>
                      <p className="text-sm text-on-surface-variant">{scope.description}</p>
                    </div>
                    {newKeyScopes.includes(scope.value) && (
                      <MaterialIcon name="check" size="sm" className="text-primary" />
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-on-surface">Expiry</Label>
              <Select value={newKeyExpiry} onValueChange={setNewKeyExpiry}>
                <SelectTrigger className="bg-surface-container-lowest border-outline-variant text-on-surface">
                  <SelectValue placeholder="Select expiry" />
                </SelectTrigger>
                <SelectContent className="bg-surface-container-high border-outline-variant">
                  <SelectItem value="7d">7 days</SelectItem>
                  <SelectItem value="30d">30 days</SelectItem>
                  <SelectItem value="90d">90 days</SelectItem>
                  <SelectItem value="365d">1 year</SelectItem>
                  <SelectItem value="never">Never</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowCreateDialog(false)}
              className="border-outline-variant text-on-surface"
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateKey}
              disabled={!newKeyName || newKeyScopes.length === 0}
              className="bg-primary text-on-primary hover:bg-primary/90"
            >
              <MaterialIcon name="add" size="sm" className="mr-2" />
              Create Key
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Show New Key Dialog */}
      <Dialog open={showKeyDialog} onOpenChange={setShowKeyDialog}>
        <DialogContent className="max-w-lg bg-surface-container-high border-outline-variant">
          <DialogHeader>
            <DialogTitle className="text-on-surface flex items-center gap-2">
              <MaterialIcon name="key" size="sm" className="text-primary" />
              API Key Created
            </DialogTitle>
            <DialogDescription className="text-on-surface-variant">
              Copy this key now. You won&apos;t be able to see it again!
            </DialogDescription>
          </DialogHeader>

          {newlyCreatedKey && (
            <div className="space-y-4 py-4">
              <div className="flex items-center gap-2">
                <code className="flex-1 block p-4 bg-surface-container-lowest border border-outline-variant rounded-lg text-sm break-all font-mono text-on-surface">
                  {newlyCreatedKey}
                </code>
                <Button
                  onClick={() => handleCopyKey(newlyCreatedKey)}
                  className="bg-primary text-on-primary hover:bg-primary/90"
                  aria-label="Copy key"
                >
                  <MaterialIcon name="content_copy" size="sm" />
                </Button>
              </div>
            </div>
          )}

          <div className="flex justify-end">
            <Button
              variant="outline"
              onClick={() => {
                setShowKeyDialog(false);
                setNewlyCreatedKey(null);
              }}
              className="border-outline-variant text-on-surface"
            >
              I&apos;ve Saved the Key
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}