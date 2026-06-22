"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HostResponse } from "@/lib/validations/hosts";
import {
  ServerIcon,
  MonitorIcon,
  StarIcon,
  StarOffIcon,
  MoreHorizontalIcon,
  WifiIcon,
  WifiOffIcon,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const hostTypeIcons = {
  ssh: ServerIcon,
  rdp: MonitorIcon,
  vnc: MonitorIcon,
};

const hostTypeLabels = {
  ssh: "SSH",
  rdp: "RDP",
  vnc: "VNC",
};

const statusColors = {
  online: "bg-emerald-500",
  offline: "bg-red-500",
  unknown: "bg-gray-400",
};

interface HostCardProps {
  host: HostResponse;
  onConnect?: (host: HostResponse) => void;
  onEdit?: (host: HostResponse) => void;
  onDelete?: (host: HostResponse) => void;
  onToggleFavorite?: (host: HostResponse) => void;
  onDuplicate?: (host: HostResponse) => void;
  compact?: boolean;
  className?: string;
}

export function HostCard({
  host,
  onConnect,
  onEdit,
  onDelete,
  onToggleFavorite,
  onDuplicate,
  compact = false,
  className,
}: HostCardProps) {
  const Icon = hostTypeIcons[host.hostType];
  const statusColor = statusColors[host.status ?? "unknown"];

  return (
    <div
      className={cn(
        "group relative flex flex-col gap-3 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:shadow-md",
        compact && "p-3",
        className
      )}
      style={host.color ? { borderLeftColor: host.color, borderLeftWidth: 3 } : undefined}
    >
      {/* Status indicator */}
      <div className="absolute right-3 top-3 flex items-center gap-2">
        <div className={cn("h-2 w-2 rounded-full", statusColor)} title={`Status: ${host.status}`} />
        {host.favorite && <StarIcon className="h-3.5 w-3.5 fill-yellow-400 text-yellow-400" />}
      </div>

      {/* Header */}
      <div className="flex items-start gap-3 pr-8">
        <div
          className={cn(
            "flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-white",
            host.hostType === "ssh" && "bg-blue-600",
            host.hostType === "rdp" && "bg-purple-600",
            host.hostType === "vnc" && "bg-teal-600"
          )}
        >
          <Icon className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="truncate font-medium text-sm">{host.name}</h3>
          <p className="truncate text-xs text-muted-foreground">
            {host.host}:{host.port}
          </p>
        </div>
      </div>

      {/* Description */}
      {host.description && !compact && (
        <p className="line-clamp-2 text-xs text-muted-foreground">
          {host.description}
        </p>
      )}

      {/* Tags */}
      {host.tags && host.tags.length > 0 && !compact && (
        <div className="flex flex-wrap gap-1">
          {host.tags.slice(0, 3).map((tag) => (
            <Badge key={tag} variant="outline" className="text-xs px-1.5 py-0">
              {tag}
            </Badge>
          ))}
          {host.tags.length > 3 && (
            <Badge variant="outline" className="text-xs px-1.5 py-0">
              +{host.tags.length - 3}
            </Badge>
          )}
        </div>
      )}

      {/* Group badge */}
      {host.groupName && !compact && (
        <Badge variant="secondary" className="w-fit text-xs">
          {host.groupName}
        </Badge>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between gap-2 mt-auto">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Badge variant="outline" className="text-xs">
            {hostTypeLabels[host.hostType]}
          </Badge>
          {host.username && (
            <span className="truncate max-w-[100px]">{host.username}</span>
          )}
        </div>

        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="default"
            className="h-7 px-2 text-xs"
            onClick={() => onConnect?.(host)}
          >
            Connect
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="icon" variant="ghost" className="h-7 w-7">
                <MoreHorizontalIcon className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit?.(host)}>
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDuplicate?.(host)}>
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onToggleFavorite?.(host)}>
                {host.favorite ? "Remove from favorites" : "Add to favorites"}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => onDelete?.(host)}
                className="text-destructive"
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Last connected */}
      {host.lastConnected && !compact && (
        <p className="text-xs text-muted-foreground/60">
          Last connected: {new Date(host.lastConnected).toLocaleDateString()}
        </p>
      )}
    </div>
  );
}