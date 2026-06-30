"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { HostResponse } from "@/lib/validations/hosts";
import { MaterialIcon } from "@/components/ui/material-icon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const hostTypeIcons = {
  ssh: "server",
  rdp: "monitor",
  vnc: "monitor",
};

const hostTypeLabels = {
  ssh: "SSH",
  rdp: "RDP",
  vnc: "VNC",
};

const statusBadgeStyles = {
  online: "bg-green-500/10 text-green-400",
  offline: "bg-red-500/10 text-red-400",
  unknown: "bg-yellow-500/10 text-yellow-400",
};

const statusLabels = {
  online: "Online",
  offline: "Offline",
  unknown: "Busy",
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
  const iconName = hostTypeIcons[host.hostType];
  const statusKey = (host.status ?? "unknown") as keyof typeof statusBadgeStyles;
  const isOffline = host.status === "offline";

  return (
    <div
      className={cn(
        "glass-card relative rounded-xl p-5 flex flex-col gap-4 transition-all hover:shadow-lg",
        isOffline && "opacity-75 grayscale-[0.5]",
        compact && "p-3",
        className
      )}
      style={host.color ? { borderLeftColor: host.color, borderLeftWidth: 3 } : undefined}
    >
      {/* Header: icon + title + status */}
      <div className="flex items-start gap-3">
        <div className="w-12 h-12 shrink-0 rounded-lg bg-tertiary-container/10 flex items-center justify-center text-on-surface">
          <MaterialIcon name={iconName} size="lg" className="text-on-surface" />
        </div>
        <div className="min-w-0 flex-1 flex flex-col gap-1">
          <div className="flex items-center justify-between gap-2">
            <h3 className="truncate text-headline-sm text-on-surface">{host.name}</h3>
            <span
              className={cn(
                "shrink-0 rounded-full px-2 py-0.5 text-xs",
                statusBadgeStyles[statusKey]
              )}
            >
              {statusLabels[statusKey]}
            </span>
          </div>
          <p className="truncate font-mono-code text-mono-code text-on-surface-variant/70">
            {host.host}:{host.port}
          </p>
        </div>
      </div>

      {/* Description */}
      {host.description && !compact && (
        <p className="line-clamp-2 text-xs text-on-surface-variant">
          {host.description}
        </p>
      )}

      {/* Tags */}
      {host.tags && host.tags.length > 0 && !compact && (
        <div className="flex flex-wrap gap-1">
          {host.tags.slice(0, 3).map((tag) => (
            <span
              key={tag}
              className="px-2 py-0.5 bg-surface-variant text-on-surface-variant text-[11px] rounded-md"
            >
              {tag}
            </span>
          ))}
          {host.tags.length > 3 && (
            <span className="px-2 py-0.5 bg-surface-variant text-on-surface-variant text-[11px] rounded-md">
              +{host.tags.length - 3}
            </span>
          )}
        </div>
      )}

      {/* Group badge */}
      {host.groupName && !compact && (
        <Badge variant="secondary" className="w-fit text-xs">
          {host.groupName}
        </Badge>
      )}

      {/* Favorite star */}
      {host.favorite && (
        <div className="absolute right-3 top-3">
          <MaterialIcon
            name="star"
            size="sm"
            fill
            className="text-yellow-400"
          />
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between gap-2 mt-auto border-t border-outline-variant pt-3">
        <div className="flex items-center gap-2 text-xs text-on-surface-variant">
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
            className="h-7 px-2 text-xs text-primary"
            onClick={() => onConnect?.(host)}
            disabled={isOffline}
          >
            Connect
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="icon" variant="ghost" className="h-7 w-7">
                <MaterialIcon name="more_horiz" size="sm" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit?.(host)}>
                <MaterialIcon name="edit" size="sm" className="mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDuplicate?.(host)}>
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onToggleFavorite?.(host)}>
                <MaterialIcon
                  name={host.favorite ? "star_border" : "star"}
                  size="sm"
                  fill={host.favorite}
                  className="mr-2"
                />
                {host.favorite ? "Remove from favorites" : "Add to favorites"}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => onDelete?.(host)}
                className="text-destructive"
              >
                <MaterialIcon name="delete" size="sm" className="mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Last connected */}
      {host.lastConnected && !compact && (
        <p className="text-xs text-on-surface-variant/60">
          Last connected: {new Date(host.lastConnected).toLocaleDateString()}
        </p>
      )}
    </div>
  );
}