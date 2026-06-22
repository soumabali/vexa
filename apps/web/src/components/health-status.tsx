'use client';

import React from 'react';
import { Activity, AlertCircle, Wifi, WifiOff } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

export type HealthStatus = 'healthy' | 'unhealthy' | 'unknown' | 'checking';

interface HealthStatusProps {
  status: HealthStatus;
  latency?: number; // ms
  lastChecked?: string;
  showLatency?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const statusConfig = {
  healthy: {
    icon: Wifi,
    color: 'text-green-500',
    bgColor: 'bg-green-500/10',
    badge: 'default' as const,
    label: 'Online',
  },
  unhealthy: {
    icon: WifiOff,
    color: 'text-red-500',
    bgColor: 'bg-red-500/10',
    badge: 'destructive' as const,
    label: 'Offline',
  },
  unknown: {
    icon: AlertCircle,
    color: 'text-amber-500',
    bgColor: 'bg-amber-500/10',
    badge: 'secondary' as const,
    label: 'Unknown',
  },
  checking: {
    icon: Activity,
    color: 'text-blue-500',
    bgColor: 'bg-blue-500/10',
    badge: 'outline' as const,
    label: 'Checking...',
  },
};

const sizeConfig = {
  sm: { icon: 14, badge: 'text-[10px] px-1.5 py-0' },
  md: { icon: 16, badge: 'text-xs px-2 py-0.5' },
  lg: { icon: 20, badge: 'text-sm px-2.5 py-0.5' },
};

export default function HealthStatus({
  status,
  latency,
  lastChecked,
  showLatency = true,
  size = 'md',
}: HealthStatusProps) {
  const config = statusConfig[status];
  const Icon = config.icon;
  const sizes = sizeConfig[size];

  const formatLatency = (ms: number) => {
    if (ms < 1) return `<1ms`;
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  const getLatencyColor = (ms: number) => {
    if (ms < 50) return 'text-green-500';
    if (ms < 200) return 'text-amber-500';
    return 'text-red-500';
  };

  const tooltipContent = (
    <div className="space-y-1">
      <p className="font-medium">{config.label}</p>
      {latency !== undefined && latency > 0 && (
        <p>Latency: {formatLatency(latency)}</p>
      )}
      {lastChecked && (
        <p className="text-xs text-muted-foreground">
          Last checked: {new Date(lastChecked).toLocaleTimeString()}
        </p>
      )}
    </div>
  );

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="inline-flex items-center gap-1.5">
            <div className={`p-1 rounded-full ${config.bgColor}`}>
              <Icon
                className={`${config.color}`}
                size={sizes.icon}
              />
            </div>
            <Badge variant={config.badge} className={sizes.badge}>
              {config.label}
            </Badge>
            {showLatency && latency !== undefined && latency > 0 && status === 'healthy' && (
              <span className={`text-xs ${getLatencyColor(latency)}`}>
                {formatLatency(latency)}
              </span>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent>{tooltipContent}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

// Host health card component
interface HostHealthCardProps {
  hostName: string;
  address: string;
  status: HealthStatus;
  latency?: number;
  lastChecked?: string;
  uptime?: number; // percentage
  consecutiveFails?: number;
}

export function HostHealthCard({
  hostName,
  address,
  status,
  latency,
  lastChecked,
  uptime,
  consecutiveFails,
}: HostHealthCardProps) {
  return (
    <div className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors">
      <div className="min-w-0">
        <p className="font-medium truncate">{hostName}</p>
        <p className="text-xs text-muted-foreground truncate">{address}</p>
      </div>
      <div className="flex items-center gap-3">
        {uptime !== undefined && (
          <span className={`text-xs ${uptime >= 99 ? 'text-green-500' : uptime >= 95 ? 'text-amber-500' : 'text-red-500'}`}>
            {uptime.toFixed(1)}%
          </span>
        )}
        {(consecutiveFails || 0) > 0 && (
          <span className="text-xs text-red-500">
            {(consecutiveFails || 0)} fail{(consecutiveFails || 0) > 1 ? 's' : ''}
          </span>
        )}
        <HealthStatus
          status={status}
          latency={latency}
          lastChecked={lastChecked}
          size="sm"
        />
      </div>
    </div>
  );
}

// Health history mini chart
interface HealthHistoryProps {
  history: { status: HealthStatus; checkedAt: string }[];
  maxItems?: number;
}

export function HealthHistory({ history, maxItems = 24 }: HealthHistoryProps) {
  const displayHistory = history.slice(-maxItems);
  
  const getStatusColor = (status: HealthStatus) => {
    switch (status) {
      case 'healthy': return 'bg-green-500';
      case 'unhealthy': return 'bg-red-500';
      case 'checking': return 'bg-blue-500';
      default: return 'bg-gray-300';
    }
  };

  return (
    <div className="flex items-end gap-0.5 h-8">
      {displayHistory.map((h, i) => (
        <TooltipProvider key={i}>
          <Tooltip>
            <TooltipTrigger asChild>
              <div
                className={`w-1.5 rounded-sm ${getStatusColor(h.status)} hover:opacity-80 transition-opacity`}
                style={{ height: h.status === 'healthy' ? '100%' : '40%' }}
              />
            </TooltipTrigger>
            <TooltipContent side="top">
              <p className="text-xs">
                {new Date(h.checkedAt).toLocaleTimeString()} — {statusConfig[h.status].label}
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      ))}
      {displayHistory.length === 0 && (
        <p className="text-xs text-muted-foreground">No history available</p>
      )}
    </div>
  );
}
