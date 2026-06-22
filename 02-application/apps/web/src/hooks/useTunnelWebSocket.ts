// hooks/useTunnelWebSocket.ts — WebSocket hook for tunnel stats
import { useEffect, useRef } from "react"
import { useTunnelStore } from "@/store/tunnel-store"
import type { TunnelStats } from "@/types/tunnel"

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "wss://localhost:8080/ws"

export function useTunnelWebSocket(tunnelIds: string[]) {
  const wsRef = useRef<WebSocket | null>(null)
  const setStats = useTunnelStore((s: any) => s.setStats)
  const addConnected = useTunnelStore((s: any) => s.addConnected)
  const removeConnected = useTunnelStore((s: any) => s.removeConnected)

  useEffect(() => {
    if (tunnelIds.length === 0) return

    const ws = new WebSocket(WS_URL)
    wsRef.current = ws

    ws.onopen = () => {
      tunnelIds.forEach((id) => {
        ws.send(JSON.stringify({ type: "tunnel:subscribe", tunnel_id: id }))
        addConnected(id)
      })
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === "tunnel:stats" && msg.tunnel_id) {
          setStats(msg.tunnel_id, msg.stats as TunnelStats)
        }
      } catch {
        // ignore invalid messages
      }
    }

    ws.onclose = () => {
      tunnelIds.forEach((id) => removeConnected(id))
    }

    return () => {
      ws.close()
      tunnelIds.forEach((id) => removeConnected(id))
    }
  }, [tunnelIds.join(",")])

  return wsRef
}
