"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { MaterialIcon } from "@/components/ui/material-icon";
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
      <Card className="bg-surface-container-low border border-outline-variant">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-label-md text-on-surface">
            <MaterialIcon name="bolt" size="sm" className="text-primary" />
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
      <Card className="bg-surface-container-low border border-outline-variant">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-label-md text-on-surface">
            <MaterialIcon name="bolt" size="sm" className="text-primary" />
            Quick Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-body-md text-on-surface-variant">
            <MaterialIcon name="error" size="sm" className="text-error" />
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
      icon: "schedule",
      label: "Last Connected",
      value: formatRelative(data.last_connected_at),
    },
    {
      icon: "sensors",
      label: "Total Sessions",
      value: data.total_sessions,
    },
    {
      icon: "shield",
      label: "Tunnels",
      value: `${data.active_tunnels}/${data.tunnel_count}`,
      sublabel: "active / total",
    },
    {
      icon: "bolt",
      label: "Type",
      value: host.hostType?.toUpperCase() ?? "—",
    },
  ];

  return (
    <Card className="bg-surface-container-low border border-outline-variant">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-label-md text-on-surface">
          <MaterialIcon name="bolt" size="sm" className="text-primary" />
          Quick Stats
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {items.map((item) => {
            return (
              <div
                key={item.label}
                className="flex flex-col gap-1 p-3 rounded-md bg-surface-container border border-outline-variant"
              >
                <div className="flex items-center gap-2 text-label-md text-on-surface-variant">
                  <MaterialIcon name={item.icon} size="sm" />
                  <span>{item.label}</span>
                </div>
                <div
                  className="text-body-md font-semibold text-on-surface truncate"
                  title={String(item.value)}
                >
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