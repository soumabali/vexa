"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Download, QrCode, Copy } from "lucide-react"
import type { Tunnel, TunnelConfig } from "@/types/tunnel"
import { getTunnelConfig } from "@/lib/api/tunnels"

interface TunnelConfigModalProps {
  tunnel: Tunnel
}

export function TunnelConfigModal({ tunnel }: TunnelConfigModalProps) {
  const [config, setConfig] = useState<TunnelConfig | null>(null)
  const [open, setOpen] = useState(false)

  const loadConfig = async () => {
    const cfg = await getTunnelConfig(tunnel.id)
    setConfig(cfg)
  }

  const handleOpen = (open: boolean) => {
    setOpen(open)
    if (open) loadConfig()
  }

  const confText = config
    ? `[Interface]
PrivateKey = ${config.interface.private_key}
Address = ${config.interface.address}
DNS = ${config.interface.dns.join(", ")}

[Peer]
PublicKey = ${config.peer.public_key}
${config.peer.preshared_key ? `PresharedKey = ${config.peer.preshared_key}\n` : ""}AllowedIPs = ${config.peer.allowed_ips.join(", ")}
Endpoint = ${config.peer.endpoint}
PersistentKeepalive = ${config.peer.persistent_keepalive}`
    : ""

  const handleCopy = () => {
    navigator.clipboard.writeText(confText)
  }

  const handleDownload = () => {
    const blob = new Blob([confText], { type: "text/plain" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${tunnel.interface_name}.conf`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <Download className="mr-2 h-4 w-4" />
          Config
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>WireGuard Configuration</DialogTitle>
          <DialogDescription>
            Download or copy the configuration for {tunnel.interface_name}.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {config ? (
            <>
              <pre className="bg-muted p-3 rounded text-xs overflow-auto max-h-64">
                {confText}
              </pre>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={handleCopy}>
                  <Copy className="mr-2 h-4 w-4" />
                  Copy
                </Button>
                <Button variant="outline" size="sm" onClick={handleDownload}>
                  <Download className="mr-2 h-4 w-4" />
                  Download .conf
                </Button>
              </div>
            </>
          ) : (
            <p className="text-muted-foreground">Loading configuration…</p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
