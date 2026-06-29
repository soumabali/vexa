"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import {
  Shield,
  Power,
  RefreshCw,
  Trash2,
  Download,
  QrCode,
  Copy,
} from "lucide-react"
import type { Tunnel, TunnelStats } from "@/types/tunnel"
import { disableTunnel, enableTunnel, rotateTunnelKeys } from "@/lib/api/tunnels"
import { useTunnelStore } from "@/store/tunnel-store"
import { TunnelConfigModal } from "./TunnelConfigModal"
import { TunnelDeleteDialog } from "./TunnelDeleteDialog"
import { TunnelStatsCards } from "./TunnelStatsCards"

interface TunnelTabProps {
  hostId: string
  tunnel?: Tunnel
}

export function TunnelTab({ hostId, tunnel }: TunnelTabProps) {
  const [rotating, setRotating] = useState(false)
  const [enabling, setEnabling] = useState(false)
  const stats = useTunnelStore((s) =>
    tunnel ? s.tunnelStats[tunnel.id] : undefined
  )

  if (!tunnel) {
    return <TunnelEmptyState hostId={hostId} />
  }

  const handleEnable = async () => {
    setEnabling(true)
    try {
      if (tunnel.is_enabled) {
        await disableTunnel(tunnel.id)
      } else {
        await enableTunnel(tunnel.id)
      }
    } finally {
      setEnabling(false)
    }
  }

  const handleRotate = async () => {
    setRotating(true)
    try {
      await rotateTunnelKeys(tunnel.id)
    } finally {
      setRotating(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Status Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="h-5 w-5 text-info" />
          <div>
            <p className="font-medium">WireGuard Tunnel</p>
            <p className="text-sm text-muted-foreground">
              {tunnel.interface_name} · {tunnel.client_ip}
            </p>
          </div>
        </div>
        <TunnelStatusBadge enabled={tunnel.is_enabled} />
      </div>

      <Separator />

      {/* Actions */}
      <div className="flex flex-wrap gap-2">
        <Button
          variant={tunnel.is_enabled ? "destructive" : "default"}
          size="sm"
          onClick={handleEnable}
          disabled={enabling}
        >
          <Power className="mr-2 h-4 w-4" />
          {tunnel.is_enabled ? "Disable" : "Enable"}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRotate}
          disabled={rotating}
        >
          <RefreshCw className={`mr-2 h-4 w-4 ${rotating ? "animate-spin" : ""}`} />
          Rotate Keys
        </Button>
        <TunnelConfigModal tunnel={tunnel} />
        <TunnelDeleteDialog tunnel={tunnel} />
      </div>

      {/* Stats */}
      {tunnel.is_enabled && <TunnelStatsCards stats={stats} />}

      {/* Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Server Public Key</span>
            <span className="font-mono">{tunnel.server_public_key.slice(0, 32)}…</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Listen Port</span>
            <span>{tunnel.listen_port}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Allowed IPs</span>
            <span>{tunnel.allowed_ips.join(", ")}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Last Rotated</span>
            <span>{tunnel.last_rotated_at ? new Date(tunnel.last_rotated_at).toLocaleDateString() : "Never"}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function TunnelEmptyState({ hostId }: { hostId: string }) {
  const [open, setOpen] = useState(false)
  const [allowedIps, setAllowedIps] = useState("0.0.0.0/0, ::/0")
  const [port, setPort] = useState("51820")
  const [creating, setCreating] = useState(false)

  const handleCreate = async () => {
    setCreating(true)
    try {
      const { createTunnel } = await import("@/lib/api/tunnels")
      await createTunnel({
        host_id: hostId,
        allowed_ips: allowedIps.split(",").map((s) => s.trim()),
        listen_port: parseInt(port, 10),
      })
      setOpen(false)
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="flex flex-col items-center justify-center py-12 space-y-4">
      <Shield className="h-12 w-12 text-muted-foreground" />
      <p className="text-lg font-medium">No WireGuard tunnel configured</p>
      <p className="text-sm text-muted-foreground text-center max-w-sm">
        Create a tunnel to connect to this host&apos;s network securely over an encrypted WireGuard tunnel.
      </p>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button>Create Tunnel</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create WireGuard Tunnel</DialogTitle>
            <DialogDescription>
              Configure tunnel parameters for this host.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="allowed-ips">Allowed IPs (CIDR)</Label>
              <Input
                id="allowed-ips"
                value={allowedIps}
                onChange={(e) => setAllowedIps(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="port">Listen Port</Label>
              <Input
                id="port"
                type="number"
                value={port}
                onChange={(e) => setPort(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={creating}>
              {creating ? "Creating…" : "Create Tunnel"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function TunnelStatusBadge({ enabled }: { enabled: boolean }) {
  return enabled ? (
    <Badge variant="default" className="bg-success text-success-foreground">
      ● Active
    </Badge>
  ) : (
    <Badge variant="secondary">● Inactive</Badge>
  )
}
