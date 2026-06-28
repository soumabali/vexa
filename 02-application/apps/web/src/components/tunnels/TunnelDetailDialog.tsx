"use client"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { LiveStatsPanel } from "./LiveStatsPanel"
import { TunnelStatusBadge } from "./TunnelStatusBadge"
import { getTunnel } from "@/lib/api"
import type { Tunnel } from "@/types/tunnel"

interface TunnelDetailDialogProps {
  tunnel: Tunnel | null
  onOpenChange: (open: boolean) => void
}

export function TunnelDetailDialog({
  tunnel,
  onOpenChange,
}: TunnelDetailDialogProps) {
  const [tab, setTab] = useState("stats")
  const open = !!tunnel

  const { data: fresh } = useQuery({
    queryKey: ["tunnel", tunnel?.id],
    queryFn: () => getTunnel(tunnel!.id),
    enabled: !!tunnel,
  })

  const current = fresh ?? tunnel

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg" data-testid="tunnel-detail-dialog">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {current?.interface_name ?? "Tunnel"}
            {current && (
              <TunnelStatusBadge status={current.is_enabled ? "up" : "disabled"} />
            )}
          </DialogTitle>
          <DialogDescription>
            Client IP {current?.client_ip ?? "—"} · Listen port{" "}
            {current?.listen_port ?? "—"}
          </DialogDescription>
        </DialogHeader>

        {current && (
          <Tabs value={tab} onValueChange={setTab}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="stats">Live Stats</TabsTrigger>
              <TabsTrigger value="overview">Overview</TabsTrigger>
            </TabsList>

            <TabsContent value="stats" className="space-y-3">
              <LiveStatsPanel tunnelId={current.id} />
            </TabsContent>

            <TabsContent value="overview" className="space-y-3">
              <OverviewCard current={current} />
            </TabsContent>
          </Tabs>
        )}
      </DialogContent>
    </Dialog>
  )
}

function OverviewCard({ current }: { current: Tunnel }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium">Details</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        <Row label="Interface" value={current.interface_name} />
        <Row label="Host ID" value={current.host_id} mono />
        <Row label="Client IP" value={current.client_ip} mono />
        <Row label="Listen Port" value={String(current.listen_port)} mono />
        <Row
          label="Allowed IPs"
          value={current.allowed_ips.join(", ") || "—"}
          mono
        />
        <Row
          label="Created"
          value={new Date(current.createdAt).toLocaleString()}
        />
        <Row
          label="Last Rotated"
          value={
            current.last_rotated_at
              ? new Date(current.last_rotated_at).toLocaleString()
              : "Never"
          }
        />
      </CardContent>
    </Card>
  )
}

function Row({
  label,
  value,
  mono,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="flex items-start justify-between gap-3 border-b border-border/40 pb-1 last:border-0">
      <span className="text-muted-foreground">{label}</span>
      <span className={mono ? "font-mono text-xs" : ""}>{value}</span>
    </div>
  )
}