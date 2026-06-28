"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  Pencil,
  Trash2,
  Copy as CopyIcon,
  Network,
  Server,
  Loader2,
} from "lucide-react";
import { hostsApi } from "@/lib/api/hosts";
import type { HostDetailStats } from "@/lib/api/hosts";
import { createTunnel } from "@/lib/api/tunnels";
import { HostResponse } from "@/lib/validations/hosts";
import { HostForm } from "@/components/hosts/HostForm";
import { ConnectionButton } from "@/components/hosts/ConnectionButton";
import { HostStatsCard } from "@/components/hosts/HostStatsCard";
import { HostMetadataPanel } from "@/components/hosts/HostMetadataPanel";
import { CopyableField } from "@/components/hosts/CopyableField";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { useAsyncData } from "@/hooks/useAsyncData";
import { toast } from "sonner";

export default function HostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [editingHost, setEditingHost] = useState<HostResponse | null>(null);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showTunnelDialog, setShowTunnelDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [tunnelName, setTunnelName] = useState("");
  const [isCreatingTunnel, setIsCreatingTunnel] = useState(false);

  // Host detail fetch (existing pattern preserved)
  const {
    data: host,
    loading,
    error,
    reload: refetchHost,
  } = useAsyncData<HostResponse>(async () => {
    const resp = await hostsApi.get(id);
    return resp;
  });

  // Per-host stats fetch (new)
  const {
    data: stats,
    loading: statsLoading,
    error: statsError,
  } = useAsyncData<HostDetailStats>(async () => {
    return await hostsApi.getStats(id);
  });

  // Sync edit form when host loads
  useEffect(() => {
    if (host) setEditingHost(host);
  }, [host]);

  const handleCopyHost = async () => {
    if (!host) return;
    try {
      await navigator.clipboard.writeText(`${host.address}:${host.port}`);
      toast.success("Address copied");
    } catch {
      toast.error("Copy failed");
    }
  };

  const handleDelete = async () => {
    if (!host) return;
    setIsDeleting(true);
    try {
      await hostsApi.delete(host.id);
      toast.success("Host deleted");
      router.push("/hosts");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Delete failed");
      setIsDeleting(false);
    }
  };

  const handleCreateTunnel = async () => {
    if (!host) return;
    if (!tunnelName.trim()) {
      toast.error("Tunnel name required");
      return;
    }
    setIsCreatingTunnel(true);
    try {
      await createTunnel({
        hostId: host.id,
        name: tunnelName.trim(),
      });
      toast.success("Tunnel created");
      setShowTunnelDialog(false);
      setTunnelName("");
      // Stats changed; refresh.
      window.location.reload();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tunnel creation failed");
    } finally {
      setIsCreatingTunnel(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6 flex items-center justify-center min-h-[400px]">
        <LoadingSpinner className="h-8 w-8" />
      </div>
    );
  }

  if (error || !host) {
    return (
      <div className="container mx-auto p-6 space-y-4">
        <Button variant="ghost" onClick={() => router.push("/hosts")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to hosts
        </Button>
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {error?.message ?? "Host not found"}
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div className="flex items-start gap-3 min-w-0">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => router.push("/hosts")}
            aria-label="Back to hosts"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <Server className="h-5 w-5 text-muted-foreground shrink-0" />
              <h1 className="text-2xl font-bold truncate">{host.name}</h1>
              <Badge variant={host.favorite ? "default" : "secondary"}>
                {host.favorite ? "active" : "inactive"}
              </Badge>
              {host.hostType && (
                <Badge variant="outline" className="uppercase">
                  {host.hostType}
                </Badge>
              )}
            </div>
            <div className="mt-2">
              <CopyableField
                value={`${host.address}:${host.port}`}
                ariaLabel="Copy host:port"
              />
            </div>
            {host.description && (
              <p className="mt-2 text-sm text-muted-foreground">
                {host.description}
              </p>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 flex-wrap">
          <ConnectionButton host={host} />
          <Button variant="outline" size="sm" onClick={handleCopyHost}>
            <CopyIcon className="h-4 w-4 mr-1" />
            Copy
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowTunnelDialog(true)}
          >
            <Network className="h-4 w-4 mr-1" />
            Tunnel
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowEditDialog(true)}
          >
            <Pencil className="h-4 w-4 mr-1" />
            Edit
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setShowDeleteConfirm(true)}
          >
            <Trash2 className="h-4 w-4 mr-1" />
            Delete
          </Button>
        </div>
      </div>

      {/* Quick stats */}
      <HostStatsCard
        host={host}
        stats={stats}
        loading={statsLoading}
        error={statsError}
      />

      {/* Metadata */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <HostMetadataPanel host={host} />

        {/* Recent activity placeholder — keeps card grid balanced.
            Future: surface last few connection_history rows. */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            {stats?.last_connected_at ? (
              <p>
                Last connected{" "}
                {new Date(stats.last_connected_at).toLocaleString()} ·{" "}
                {stats.total_sessions} total session
                {stats.total_sessions === 1 ? "" : "s"}
              </p>
            ) : (
              <p>No prior connections recorded.</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Edit dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Host</DialogTitle>
            <DialogDescription>
              Update connection details for {host.name}.
            </DialogDescription>
          </DialogHeader>
          {editingHost && (
            <HostForm
              host={editingHost}
              onSuccess={() => {
                setShowEditDialog(false);
                refetchHost();
              }}
              onCancel={() => setShowEditDialog(false)}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete host?</DialogTitle>
            <DialogDescription>
              This will permanently remove <strong>{host.name}</strong> and all
              associated credentials. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-3 mt-4">
            <Button
              variant="outline"
              onClick={() => setShowDeleteConfirm(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Create tunnel dialog */}
      <Dialog open={showTunnelDialog} onOpenChange={setShowTunnelDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create WireGuard Tunnel</DialogTitle>
            <DialogDescription>
              Provision a WireGuard tunnel to <strong>{host.name}</strong>.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 mt-2">
            <div className="space-y-2">
              <Label htmlFor="tunnel-name">Tunnel name</Label>
              <Input
                id="tunnel-name"
                value={tunnelName}
                onChange={(e) => setTunnelName(e.target.value)}
                placeholder="e.g. production-vpn"
                disabled={isCreatingTunnel}
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-3">
              <Button
                variant="outline"
                onClick={() => {
                  setShowTunnelDialog(false);
                  setTunnelName("");
                }}
                disabled={isCreatingTunnel}
              >
                Cancel
              </Button>
              <Button
                onClick={handleCreateTunnel}
                disabled={isCreatingTunnel || !tunnelName.trim()}
              >
                {isCreatingTunnel && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Create
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
