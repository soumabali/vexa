"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
  Activity,
  Radio,
  Shield,
  Clock,
  AlertCircle,
} from "lucide-react";
import { HostResponse } from "@/lib/validations/hosts";

export interface HostDetailStats {
  total_sessions: number;
  last_connected_at: string | null;
  tunnel_count: number;
  active_tunnels: number;
}

interface HostStatsCardProps {
  host: HostResponse;
  stats: HostDetailStats | null;
  loading: boolean;
  error: Error | null;
}

function formatRelative(iso: string | null): string {
  if (!iso) return "Never";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "Unknown";
  const diffMs = Date.now() - then;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return "Just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  return new Date(iso).toLocaleDateString();
}

export function HostStatsCard({ host, stats, loading, error }: HostStatsCardProps) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Activity className="h-4 w-4" aria-hidden="true" />
            Quick Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {[0, 1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Activity className="h-4 w-4" aria-hidden="true" />
            Quick Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <AlertCircle className="h-4 w-4 text-destructive" aria-hidden="true" />
            <span>Stats unavailable: {error.message}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  const data = stats ?? {
    total_sessions: 0,
    last_connected_at: null,
    tunnel_count: 0,
    active_tunnels: 0,
  };

  const items = [
    {
      icon: Clock,
      label: "Last Connected",
      value: formatRelative(data.last_connected_at),
    },
    {
      icon: Radio,
      label: "Total Sessions",
      value: data.total_sessions,
    },
    {
      icon: Shield,
      label: "Tunnels",
      value: `${data.active_tunnels}/${data.tunnel_count}`,
      sublabel: "active / total",
    },
    {
      icon: Activity,
      label: "Type",
      value: host.hostType?.toUpperCase() ?? "—",
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Activity className="h-4 w-4" aria-hidden="true" />
          Quick Stats
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {items.map((item) => {
            const Icon = item.icon;
            return (
              <div
                key={item.label}
                className="flex flex-col gap-1 p-3 rounded-md bg-muted/30 border border-border"
              >
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Icon className="h-3.5 w-3.5" aria-hidden="true" />
                  <span>{item.label}</span>
                </div>
                <div className="text-xl font-semibold truncate" title={String(item.value)}>
                  {item.value}
                </div>
                {item.sublabel && (
                  <Badge variant="outline" className="text-[10px] w-fit">
                    {item.sublabel}
                  </Badge>
                )}
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
