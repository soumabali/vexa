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
import {
  Key,
  Copy,
  Trash2,
  Plus,
  Eye,
  EyeOff,
  Clock,
  Shield,
} from "lucide-react";

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
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<string | null>(null);
  const [showKeyValue, setShowKeyValue] = useState<Record<string, boolean>>({});

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
  };

  const handleRevokeKey = (id: string) => {
    setApiKeys(
      apiKeys.map((k) => (k.id === id ? { ...k, isRevoked: true } : k))
    );
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
  };

  const toggleShowKey = (id: string) => {
    setShowKeyValue({ ...showKeyValue, [id]: !showKeyValue[id] });
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">API Keys</h1>
        <p className="text-muted-foreground mt-2">
          Generate and manage API keys for programmatic access
        </p>
      </div>

      <Card className="mb-6">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              Your API Keys
            </CardTitle>
            <CardDescription>
              {apiKeys.filter((k) => !k.isRevoked).length} active keys
            </CardDescription>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Generate New Key
          </Button>
        </CardHeader>
        <CardContent className="space-y-4">
          {apiKeys.map((apiKey) => (
            <div
              key={apiKey.id}
              className={`p-4 border rounded-lg ${
                apiKey.isRevoked ? "opacity-50" : ""
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Key className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">{apiKey.name}</span>
                  {apiKey.isRevoked && (
                    <Badge variant="destructive">Revoked</Badge>
                  )}
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleShowKey(apiKey.id)}
                  >
                    {showKeyValue[apiKey.id] ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleCopyKey(apiKey.key)}
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                  {!apiKey.isRevoked && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRevokeKey(apiKey.id)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2 mb-2">
                <code className="text-sm bg-muted px-2 py-1 rounded">
                  {showKeyValue[apiKey.id]
                    ? apiKey.key
                    : `${apiKey.prefix}_...${apiKey.key.slice(-4)}`}
                </code>
              </div>

              <div className="flex flex-wrap gap-1 mb-2">
                {apiKey.scopes.map((scope) => (
                  <Badge key={scope} variant="outline" className="text-xs">
                    {scope}
                  </Badge>
                ))}
              </div>

              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span>Created: {apiKey.createdAt.toLocaleDateString()}</span>
                {apiKey.lastUsed && (
                  <span>Last used: {apiKey.lastUsed.toLocaleDateString()}</span>
                )}
                {apiKey.expiresAt && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    Expires: {apiKey.expiresAt.toLocaleDateString()}
                  </span>
                )}
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Create Key Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Generate New API Key</DialogTitle>
            <DialogDescription>
              Create a new API key with specific permissions
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="keyName">Key Name</Label>
              <Input
                id="keyName"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="e.g., Production API Key"
              />
            </div>

            <div className="space-y-2">
              <Label>Scopes</Label>
              <div className="space-y-2">
                {availableScopes.map((scope) => (
                  <div
                    key={scope.value}
                    className="flex items-center justify-between p-3 border rounded-lg cursor-pointer hover:bg-muted"
                    onClick={() => {
                      if (newKeyScopes.includes(scope.value)) {
                        setNewKeyScopes(newKeyScopes.filter((s) => s !== scope.value));
                      } else {
                        setNewKeyScopes([...newKeyScopes, scope.value]);
                      }
                    }}
                  >
                    <div>
                      <p className="font-medium">{scope.label}</p>
                      <p className="text-sm text-muted-foreground">
                        {scope.description}
                      </p>
                    </div>
                    {newKeyScopes.includes(scope.value) && (
                      <Shield className="h-5 w-5 text-primary" />
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowCreateDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateKey}
              disabled={!newKeyName || newKeyScopes.length === 0}
            >
              Generate Key
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Show New Key Dialog */}
      <Dialog open={showKeyDialog} onOpenChange={setShowKeyDialog}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle className="text-destructive">API Key Created</DialogTitle>
            <DialogDescription>
              Copy this key now. You won&apos;t be able to see it again!
            </DialogDescription>
          </DialogHeader>

          {newlyCreatedKey && (
            <div className="space-y-4 py-4">
              <code className="block p-4 bg-muted rounded-lg text-sm break-all">
                {newlyCreatedKey}
              </code>

              <Button
                className="w-full"
                onClick={() => handleCopyKey(newlyCreatedKey)}
              >
                <Copy className="h-4 w-4 mr-2" />
                Copy to Clipboard
              </Button>
            </div>
          )}

          <div className="flex justify-end">
            <Button
              variant="outline"
              onClick={() => {
                setShowKeyDialog(false);
                setNewlyCreatedKey(null);
              }}
            >
              I&apos;ve Saved the Key
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
