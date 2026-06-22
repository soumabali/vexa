"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { hostsApi } from "@/lib/api/hosts";
import { createTunnel } from "@/lib/api/tunnels";
import { HostResponse } from "@/lib/validations/hosts";
import { HostForm } from "@/components/hosts/HostForm";
import { ConnectionButton } from "@/components/hosts/ConnectionButton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  ArrowLeftIcon,
  ServerIcon,
  MonitorIcon,
  StarIcon,
  ClockIcon,
  PencilIcon,
  TrashIcon,
  CopyIcon,
  Network,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";

const hostTypeIcons = {
  ssh: ServerIcon,
  rdp: MonitorIcon,
  vnc: MonitorIcon,
};

const hostTypeLabels = {
  ssh: "SSH",
  rdp: "RDP",
  vnc: "VNC",
};

const statusColors = {
  online: "bg-emerald-500",
  offline: "bg-red-500",
  unknown: "bg-gray-400",
};

export default function HostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const hostId = params.id as string;

  const [host, setHost] = useState<HostResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showTunnelDialog, setShowTunnelDialog] = useState(false);
  const [tunnelAllowedIPs, setTunnelAllowedIPs] = useState("0.0.0.0/0, ::/0");
  const [tunnelPort, setTunnelPort] = useState("51820");
  const [tunnelDNS, setTunnelDNS] = useState("1.1.1.1, 8.8.8.8");
  const [isCreatingTunnel, setIsCreatingTunnel] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    const loadHost = async () => {
      setIsLoading(true);
      setError("");
      try {
        const data = await hostsApi.get(hostId);
        setHost(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load host");
      } finally {
        setIsLoading(false);
      }
    };
    if (hostId) loadHost();
  }, [hostId]);

  const handleEditSuccess = (updated: HostResponse) => {
    setHost(updated);
    setShowEditDialog(false);
  };

  const handleDelete = async () => {
    if (!host) return;
    setIsDeleting(true);
    try {
      await hostsApi.delete(host.id);
      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete host");
      setIsDeleting(false);
    }
  };

  const handleCopyHost = () => {
    if (host) {
      navigator.clipboard.writeText(`${host.host}:${host.port}`);
    }
  };

  const handleCreateTunnel = async () => {
    if (!host) return;
    setIsCreatingTunnel(true);
    setError("");
    try {
      await createTunnel({
        host_id: host.id,
        allowed_ips: tunnelAllowedIPs.split(",").map((s) => s.trim()).filter(Boolean),
        listen_port: parseInt(tunnelPort, 10) || 51820,
        dns_servers: tunnelDNS.split(",").map((s) => s.trim()).filter(Boolean),
      });
      setShowTunnelDialog(false);
      router.push("/tunnels");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create tunnel");
      setIsCreatingTunnel(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner />
        <span className="ml-2 text-muted-foreground">Loading host...</span>
      </div>
    );
  }

  if (error && !host) {
    return (
      <div className="container py-6">
        <Button variant="ghost" onClick={() => router.push("/hosts")} className="mb-4">
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Hosts
        </Button>
        <ErrorDisplay message={error} />
      </div>
    );
  }

  if (!host) {
    return (
      <div className="container py-6">
        <Button variant="ghost" onClick={() => router.push("/hosts")} className="mb-4">
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Hosts
        </Button>
        <p className="text-muted-foreground">Host not found</p>
      </div>
    );
  }

  const Icon = hostTypeIcons[host.hostType];

  return (
    <div className="container py-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <Button variant="ghost" onClick={() => router.push("/hosts")}>
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back
          </Button>
          <div
            className={`flex h-12 w-12 items-center justify-center rounded-lg text-white ${
              host.hostType === "ssh" ? "bg-blue-600" :
              host.hostType === "rdp" ? "bg-purple-600" : "bg-teal-600"
            }`}
            style={host.color ? { backgroundColor: host.color } : undefined}
          >
            <Icon className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-semibold">{host.name}</h1>
              {host.favorite && <StarIcon className="h-5 w-5 fill-yellow-400 text-yellow-400" />}
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <code className="rounded bg-muted px-1.5 py-0.5 font-mono">
                {host.host}:{host.port}
              </code>
              <span>•</span>
              <Badge variant="outline">{hostTypeLabels[host.hostType]}</Badge>
              <span>•</span>
              <div className={`h-2 w-2 rounded-full ${statusColors[host.status ?? "unknown"]}`} />
              <span className="capitalize">{host.status}</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <ConnectionButton host={host} />
          <Button variant="outline" size="icon" onClick={handleCopyHost}>
            <CopyIcon className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setShowTunnelDialog(true)}>
            <Network className="h-4 w-4 mr-1" />
            Tunnel
          </Button>
          <Button variant="outline" size="icon" onClick={() => setShowEditDialog(true)}>
            <PencilIcon className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="icon" onClick={() => setShowDeleteConfirm(true)}>
            <TrashIcon className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </div>

      {/* Info cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Connection Details */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Connection</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Host</span>
              <code className="font-mono">{host.host}</code>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Port</span>
              <span>{host.port}</span>
            </div>
            {host.username && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Username</span>
                <span>{host.username}</span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Group & Tags */}
        {(host.groupName || (host.tags?.length ?? 0) > 0) && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Organization</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {host.groupName && (
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Group</span>
                  <Badge variant="secondary">{host.groupName}</Badge>
                </div>
              )}
              {host.tags && host.tags.length > 0 && (
                <div className="space-y-1.5">
                  <span className="text-sm text-muted-foreground">Tags</span>
                  <div className="flex flex-wrap gap-1.5">
                    {host.tags.map((tag) => (
                      <Badge key={tag} variant="outline">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Status */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <div className={`h-3 w-3 rounded-full ${statusColors[host.status ?? "unknown"]}`} />
              <span className="text-sm capitalize">{host.status}</span>
            </div>
            {host.lastConnected && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <ClockIcon className="h-4 w-4" />
                <span>Last connected {new Date(host.lastConnected).toLocaleDateString()}</span>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Description */}
      {host.description && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Description</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{host.description}</p>
          </CardContent>
        </Card>
      )}

      {/* Connection options */}
      {(host.sshOptions || host.rdpOptions || host.vncOptions) && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Connection Options</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="rounded bg-muted p-4 text-xs overflow-auto">
              {JSON.stringify(
                host.sshOptions || host.rdpOptions || host.vncOptions,
                null,
                2
              )}
            </pre>
          </CardContent>
        </Card>
      )}

      {/* Metadata */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Metadata</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Created</span>
            <span>{new Date(host.createdAt).toLocaleString()}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Updated</span>
            <span>{new Date(host.updatedAt).toLocaleString()}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">ID</span>
            <code className="text-xs font-mono">{host.id}</code>
          </div>
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Host</DialogTitle>
            <DialogDescription>
              Update connection details for {host.name}
            </DialogDescription>
          </DialogHeader>
          <HostForm
            host={host}
            onSuccess={handleEditSuccess}
            onCancel={() => setShowEditDialog(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Delete Confirm Dialog */}
      <Dialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Host</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{host.name}&quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-3 mt-4">
            <Button variant="outline" onClick={() => setShowDeleteConfirm(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
              {isDeleting ? <LoadingSpinner className="h-4 w-4 mr-2" /> : null}
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Create Tunnel Dialog */}
      <Dialog open={showTunnelDialog} onOpenChange={setShowTunnelDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Create WireGuard Tunnel</DialogTitle>
            <DialogDescription>
              Create a tunnel for {host.name} ({host.host}).
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="allowed-ips">Allowed IPs</Label>
              <Input
                id="allowed-ips"
                value={tunnelAllowedIPs}
                onChange={(e) => setTunnelAllowedIPs(e.target.value)}
                placeholder="0.0.0.0/0, ::/0"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tunnel-port">Listen Port</Label>
              <Input
                id="tunnel-port"
                type="number"
                value={tunnelPort}
                onChange={(e) => setTunnelPort(e.target.value)}
                placeholder="51820"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tunnel-dns">DNS Servers</Label>
              <Input
                id="tunnel-dns"
                value={tunnelDNS}
                onChange={(e) => setTunnelDNS(e.target.value)}
                placeholder="1.1.1.1, 8.8.8.8"
              />
            </div>
            {error && (
              <p className="text-sm text-destructive">{error}</p>
            )}
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={() => setShowTunnelDialog(false)} disabled={isCreatingTunnel}>
              Cancel
            </Button>
            <Button onClick={handleCreateTunnel} disabled={isCreatingTunnel}>
              {isCreatingTunnel && <LoadingSpinner className="h-4 w-4 mr-2" />}
              Create Tunnel
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}