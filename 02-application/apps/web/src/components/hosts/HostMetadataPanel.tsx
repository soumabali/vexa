"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CopyableField } from "./CopyableField";
import { HostResponse } from "@/lib/validations/hosts";
import { Server, Network, KeyRound, Tag } from "lucide-react";

interface HostMetadataPanelProps {
  host: HostResponse;
}

/**
 * Read-only metadata for a single host: IP/hostname, port, username, OS,
 * tags, group path. Sensitive values (e.g., credentials) never rendered here.
 */
export function HostMetadataPanel({ host }: HostMetadataPanelProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Server className="h-4 w-4" aria-hidden="true" />
          Connection Details
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center gap-2">
          <Network className="h-4 w-4 text-muted-foreground shrink-0" aria-hidden="true" />
          <CopyableField
            label="Address"
            value={host.address}
            ariaLabel="Copy host address"
          />
          <span className="text-sm text-muted-foreground">:</span>
          <CopyableField
            label="Port"
            value={String(host.port)}
            mono
            ariaLabel="Copy port"
          />
        </div>

        {host.username && (
          <div className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-muted-foreground shrink-0" aria-hidden="true" />
            <CopyableField
              label="User"
              value={host.username}
              ariaLabel="Copy username"
            />
          </div>
        )}

        {(host as { os?: string }).os && (
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">OS:</span>
            <span>{(host as { os?: string }).os}</span>
          </div>
        )}

        {host.hostType && (
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">Protocol:</span>
            <Badge variant="outline" className="uppercase">
              {host.hostType}
            </Badge>
          </div>
        )}

        {host.groupName && (
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">Group:</span>
            <span className="font-mono text-xs">{host.groupName}</span>
          </div>
        )}

        {host.tags && host.tags.length > 0 && (
          <div className="flex items-start gap-2">
            <Tag className="h-4 w-4 text-muted-foreground shrink-0 mt-1" aria-hidden="true" />
            <div className="flex flex-wrap gap-1">
              {host.tags.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
