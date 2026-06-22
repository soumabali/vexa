"use client";

import React, { useState } from "react";
import { hostsApi } from "@/lib/api/hosts";
import { HostResponse } from "@/lib/validations/hosts";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { TerminalIcon, LoaderIcon } from "lucide-react";

interface ConnectionButtonProps {
  host: HostResponse;
  variant?: "default" | "outline" | "ghost" | "destructive";
  size?: "default" | "sm" | "lg" | "icon";
  showLabel?: boolean;
  className?: string;
}

export function ConnectionButton({
  host,
  variant = "default",
  size = "default",
  showLabel = true,
  className,
}: ConnectionButtonProps) {
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState("");

  const handleConnect = async () => {
    setIsConnecting(true);
    setError("");
    try {
      const { sessionId, token } = await hostsApi.connect(host.id);
      // Navigate to terminal with the session
      const params = new URLSearchParams({
        session: sessionId,
        token,
        host: host.id,
        type: host.hostType,
      });
      window.location.href = `/terminal?${params.toString()}`;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Connection failed");
      setIsConnecting(false);
    }
  };

  return (
    <div className="flex flex-col gap-1">
      <Button
        variant={variant}
        size={size}
        onClick={handleConnect}
        disabled={isConnecting}
        className={className}
      >
        {isConnecting ? (
          <>
            <LoaderIcon className="h-4 w-4 mr-1 animate-spin" />
            Connecting...
          </>
        ) : showLabel ? (
          <>
            <TerminalIcon className="h-4 w-4 mr-1" />
            Connect
          </>
        ) : (
          <TerminalIcon className="h-4 w-4" />
        )}
      </Button>
      {error && (
        <p className="text-xs text-destructive">{error}</p>
      )}
    </div>
  );
}