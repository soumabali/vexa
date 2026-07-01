"use client"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Tunnel } from "@/types/tunnel"
import { tunnelsApi, enableTunnelApi, disableTunnelApi, rotateTunnelKeys } from "@/lib/api"
import { TunnelStatusBadge } from "@/components/tunnels/TunnelStatusBadge"
import { TunnelConfigModal } from "@/components/tunnels/TunnelConfigModal"
import { TunnelDeleteDialog } from "@/components/tunnels/TunnelDeleteDialog"
import { TunnelDetailDialog } from "@/components/tunnels/TunnelDetailDialog"
import { DashboardLayout } from "@/components/layouts/DashboardLayout"
import { MaterialIcon } from "@/components/ui/material-icon";

export default function TunnelsPage() {
  const [filter, setFilter] = useState("")
  const [selected, setSelected] = useState<Tunnel | null>(null)
  const { data: tunnels, refetch } = useQuery({
    queryKey: ["tunnels"],
    queryFn: () => tunnelsApi(),
  })

  const filtered = tunnels?.filter(
    (t) =>
      t.interface_name.toLowerCase().includes(filter.toLowerCase()) ||
      t.client_ip.includes(filter) ||
      t.host_id.includes(filter)
  )

  const toggleTunnel = async (tunnel: Tunnel) => {
    if (tunnel.is_enabled) {
      await disableTunnelApi(tunnel.id)
    } else {
      await enableTunnelApi(tunnel.id)
    }
    refetch()
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <MaterialIcon name="shield" className="h-6 w-6" />
              Tunnels
            </h1>
            <p className="text-on-surface-variant">
              Manage your WireGuard tunnels
            </p>
          </div>
          <Input
            placeholder="Filter by name, IP, or host…"
            className="w-64"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>

        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Interface</TableHead>
                <TableHead>Host</TableHead>
                <TableHead>Client IP</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Port</TableHead>
                <TableHead>Last Rotated</TableHead>
                <TableHead className="w-12"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-12">
                    <div className="flex flex-col items-center justify-center gap-2">
                      <MaterialIcon name="vpn_lock" size="xl" className="text-on-surface-variant mb-2" />
                      <p className="text-headline-sm text-on-surface mb-1">No tunnels configured</p>
                      <p className="text-body-md text-on-surface-variant">
                        Create a WireGuard tunnel from a host detail page
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
              {filtered?.map((tunnel) => (
                <TableRow
                  key={tunnel.id}
                  className="cursor-pointer"
                  onClick={() => setSelected(tunnel)}
                  data-testid="tunnel-row"
                >
                  <TableCell className="font-mono">{tunnel.interface_name}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{tunnel.host_id.slice(0, 8)}…</Badge>
                  </TableCell>
                  <TableCell>{tunnel.client_ip}</TableCell>
                  <TableCell>
                    <TunnelStatusBadge
                      status={
                        tunnel.is_enabled
                          ? "up"
                          : "disabled"
                      }
                    />
                  </TableCell>
                  <TableCell>{tunnel.listen_port}</TableCell>
                  <TableCell>
                    {tunnel.last_rotated_at
                      ? new Date(tunnel.last_rotated_at).toLocaleDateString()
                      : "Never"}
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MaterialIcon name="more_horiz" className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => toggleTunnel(tunnel)}>
                          <MaterialIcon name="power" className="mr-2 h-4 w-4" />
                          {tunnel.is_enabled ? "Disable" : "Enable"}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={async () => { await rotateTunnelKeys(tunnel.id); refetch(); }}>
                          <MaterialIcon name="refresh" className="mr-2 h-4 w-4" />
                          Rotate Keys
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-error">
                          <MaterialIcon name="delete" className="mr-2 h-4 w-4" />
                          <TunnelDeleteDialog tunnel={tunnel} onDeleted={refetch} />
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>

      <TunnelDetailDialog
        tunnel={selected}
        onOpenChange={(open) => !open && setSelected(null)}
      />
    </DashboardLayout>
  )
}
