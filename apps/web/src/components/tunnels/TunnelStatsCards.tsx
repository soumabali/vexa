"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { TunnelStats } from "@/types/tunnel"

interface TunnelStatsCardsProps {
  stats?: TunnelStats
}

export function TunnelStatsCards({ stats }: TunnelStatsCardsProps) {
  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i]
  }

  return (
    <div className="grid grid-cols-3 gap-4">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-xs font-medium text-muted-foreground">
            Received
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold">
            {stats ? formatBytes(stats.bytes_received) : "—"}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-xs font-medium text-muted-foreground">
            Sent
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold">
            {stats ? formatBytes(stats.bytes_sent) : "—"}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-xs font-medium text-muted-foreground">
            Last Handshake
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold">
            {stats?.last_handshake
              ? new Date(stats.last_handshake).toLocaleTimeString()
              : "—"}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
