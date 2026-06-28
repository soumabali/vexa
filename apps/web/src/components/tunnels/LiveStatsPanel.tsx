"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { Activity, ArrowDownToLine, ArrowUpFromLine, Clock, RefreshCw } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Skeleton } from "@/components/ui/skeleton"
import { getTunnelStats } from "@/lib/api"
import type { TunnelStats } from "@/types/tunnel"

interface LiveStatsPanelProps {
  tunnelId: string
  refreshIntervalMs?: number
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B"
  const k = 1024
  const units = ["B", "KB", "MB", "GB", "TB"]
  const i = Math.min(units.length - 1, Math.floor(Math.log(bytes) / Math.log(k)))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${units[i]}`
}

function formatRelativeTime(iso: string | null, now: number): string {
  if (!iso) return "Never"
  const ts = new Date(iso).getTime()
  if (!Number.isFinite(ts)) return "Never"
  const diffSec = Math.max(0, Math.floor((now - ts) / 1000))
  if (diffSec < 5) return "just now"
  if (diffSec < 60) return `${diffSec}s ago`
  const min = Math.floor(diffSec / 60)
  if (min < 60) return `${min} min ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr} h ago`
  const day = Math.floor(hr / 24)
  return `${day} d ago`
}

export function LiveStatsPanel({
  tunnelId,
  refreshIntervalMs = 10000,
}: LiveStatsPanelProps) {
  const [stats, setStats] = useState<TunnelStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [now, setNow] = useState<number>(() => Date.now()) // tick to refresh relative time

  const controllerRef = useRef<AbortController | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const mountedRef = useRef(true)

  const fetchStats = useCallback(async () => {
    controllerRef.current?.abort()
    const controller = new AbortController()
    controllerRef.current = controller
    try {
      setError(null)
      const data = await getTunnelStats(tunnelId)
      if (!mountedRef.current || controller.signal.aborted) return
      setStats(data)
      setLoading(false)
    } catch (err) {
      if (!mountedRef.current || controller.signal.aborted) return
      setError(err instanceof Error ? err.message : "Failed to load stats")
      setLoading(false)
    }
  }, [tunnelId])

  // Re-fetch on id change + interval tick + autoRefresh toggle
  useEffect(() => {
    mountedRef.current = true
    Promise.resolve().then(fetchStats)
    return () => {
      mountedRef.current = false
      controllerRef.current?.abort()
    }
  }, [fetchStats])

  // Polling loop
  useEffect(() => {
    if (!autoRefresh) return

    const schedule = () => {
      timerRef.current = setTimeout(() => {
        if (document.hidden) {
          schedule()
          return
        }
        fetchStats().finally(schedule)
      }, refreshIntervalMs)
    }

    schedule()
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [autoRefresh, refreshIntervalMs, fetchStats])

  // Pause/resume on visibility change
  useEffect(() => {
    const onVisibility = () => {
      if (document.hidden) {
        if (timerRef.current) {
          clearTimeout(timerRef.current)
          timerRef.current = null
        }
      } else if (autoRefresh) {
        fetchStats()
      }
    }
    document.addEventListener("visibilitychange", onVisibility)
    return () => document.removeEventListener("visibilitychange", onVisibility)
  }, [autoRefresh, fetchStats])

  // Relative time tick (every 30s) — only when visible
  useEffect(() => {
    const id = setInterval(() => {
      if (!document.hidden) setNow(Date.now())
    }, 30000)
    return () => clearInterval(id)
  }, [])

  const handshakeOk =
    stats?.last_handshake != null &&
    now - new Date(stats.last_handshake).getTime() < 3 * 60 * 1000

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium flex items-center gap-2">
          <Activity className="h-4 w-4" />
          Live Stats
          <span
            className={`ml-2 inline-block h-2.5 w-2.5 rounded-full ${
              handshakeOk ? "bg-green-500" : "bg-red-500"
            }`}
            aria-label={handshakeOk ? "connected" : "disconnected"}
            data-testid="stats-status-dot"
          />
        </CardTitle>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer">
            <Switch
              checked={autoRefresh}
              onCheckedChange={setAutoRefresh}
              aria-label="Auto-refresh"
              data-testid="stats-autorefresh"
            />
            Auto-refresh
          </label>
          <Button
            variant="ghost"
            size="icon"
            onClick={fetchStats}
            disabled={loading}
            aria-label="Refresh stats"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {error && !loading && (
          <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 mb-3 flex items-center justify-between gap-3">
            <p className="text-sm text-destructive">{error}</p>
            <Button variant="outline" size="sm" onClick={fetchStats}>
              Retry
            </Button>
          </div>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3" data-testid="stats-grid">
          <StatTile
            icon={<ArrowDownToLine className="h-4 w-4 text-muted-foreground" />}
            label="Bytes Received"
            loading={loading && !stats}
            value={stats ? formatBytes(stats.bytes_received) : "—"}
          />
          <StatTile
            icon={<ArrowUpFromLine className="h-4 w-4 text-muted-foreground" />}
            label="Bytes Sent"
            loading={loading && !stats}
            value={stats ? formatBytes(stats.bytes_sent) : "—"}
          />
          <StatTile
            icon={<Clock className="h-4 w-4 text-muted-foreground" />}
            label="Last Handshake"
            loading={loading && !stats}
            value={stats ? formatRelativeTime(stats.last_handshake, now) : "—"}
          />
        </div>
      </CardContent>
    </Card>
  )
}

function StatTile({
  icon,
  label,
  value,
  loading,
}: {
  icon: React.ReactNode
  label: string
  value: string
  loading?: boolean
}) {
  return (
    <div className="rounded-md border bg-muted/30 p-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        {icon}
        {label}
      </div>
      {loading ? (
        <Skeleton className="h-7 w-24 mt-2" />
      ) : (
        <p className="text-xl font-semibold mt-2">{value}</p>
      )}
    </div>
  )
}