// store/tunnel-store.ts — Zustand store for real-time tunnel state
import { create } from "zustand"
import type { TunnelStats } from "@/types/tunnel"

interface TunnelState {
  tunnelStats: Record<string, TunnelStats>
  connectedTunnelIds: Set<string>
  setStats: (tunnelId: string, stats: TunnelStats) => void
  removeStats: (tunnelId: string) => void
  addConnected: (tunnelId: string) => void
  removeConnected: (tunnelId: string) => void
  isConnected: (tunnelId: string) => boolean
}

export const useTunnelStore = create<TunnelState>((set, get) => ({
  tunnelStats: {},
  connectedTunnelIds: new Set(),

  setStats: (tunnelId: string, stats: TunnelStats) =>
    set((state) => ({
      tunnelStats: { ...state.tunnelStats, [tunnelId]: stats },
    })),

  removeStats: (tunnelId: string) =>
    set((state) => {
      const next = { ...state.tunnelStats }
      delete next[tunnelId]
      return { tunnelStats: next }
    }),

  addConnected: (tunnelId: string) =>
    set((state) => {
      const next = new Set(state.connectedTunnelIds)
      next.add(tunnelId)
      return { connectedTunnelIds: next }
    }),

  removeConnected: (tunnelId: string) =>
    set((state) => {
      const next = new Set(state.connectedTunnelIds)
      next.delete(tunnelId)
      return { connectedTunnelIds: next }
    }),

  isConnected: (tunnelId: string) => get().connectedTunnelIds.has(tunnelId),
}))
