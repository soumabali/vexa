"use client";

import { useState } from "react";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingSpinner } from "../auth/LoadingSpinner";
import { ErrorDisplay } from "../auth/ErrorDisplay";
import { Laptop, Smartphone, Tablet, X, Monitor } from "lucide-react";

interface Session {
  id: string;
  device: string;
  browser: string;
  ip: string;
  location: string;
  lastActive: string;
  isCurrent: boolean;
}

const mockSessions: Session[] = [
  {
    id: "1",
    device: "MacBook Pro",
    browser: "Chrome 120.0",
    ip: "192.168.1.1",
    location: "San Francisco, CA",
    lastActive: "2024-01-15T10:30:00Z",
    isCurrent: true,
  },
  {
    id: "2",
    device: "iPhone 15",
    browser: "Safari 17.0",
    ip: "192.168.1.2",
    location: "San Francisco, CA",
    lastActive: "2024-01-15T09:15:00Z",
    isCurrent: false,
  },
  {
    id: "3",
    device: "Windows PC",
    browser: "Firefox 121.0",
    ip: "192.168.1.3",
    location: "New York, NY",
    lastActive: "2024-01-14T16:45:00Z",
    isCurrent: false,
  },
];

function DeviceIcon({ device }: { device: string }) {
  if (device.toLowerCase().includes("iphone") || device.toLowerCase().includes("android")) {
    return <Smartphone className="h-5 w-5" />;
  }
  if (device.toLowerCase().includes("ipad") || device.toLowerCase().includes("tablet")) {
    return <Tablet className="h-5 w-5" />;
  }
  if (device.toLowerCase().includes("macbook") || device.toLowerCase().includes("laptop")) {
    return <Laptop className="h-5 w-5" />;
  }
  return <Monitor className="h-5 w-5" />;
}

export function SessionList() {
  const [sessions, setSessions] = useState<Session[]>(mockSessions);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [revokingId, setRevokingId] = useState<string | null>(null);

  const handleRevoke = async (sessionId: string) => {
    setRevokingId(sessionId);
    setError("");
    try {
      await authApi.revokeSession(sessionId);
      setSessions(sessions.filter((s) => s.id !== sessionId));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to revoke session");
    } finally {
      setRevokingId(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Active Sessions</CardTitle>
        <CardDescription>
          Manage your active sessions across devices
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error && <ErrorDisplay message={error} />}
        <div className="space-y-4">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={`flex items-center justify-between p-4 rounded-lg border ${
                session.isCurrent ? "bg-primary/5 border-primary/20" : "bg-card"
              }`}
            >
              <div className="flex items-center gap-4">
                <div className="p-2 bg-muted rounded-full">
                  <DeviceIcon device={session.device} />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{session.device}</p>
                    {session.isCurrent && (
                      <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full">
                        Current
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {session.browser} · {session.location}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    IP: {session.ip} · Last active: {formatDate(session.lastActive)}
                  </p>
                </div>
              </div>
              {!session.isCurrent && (
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleRevoke(session.id)}
                  disabled={revokingId === session.id}
                >
                  {revokingId === session.id ? (
                    <LoadingSpinner size="sm" />
                  ) : (
                    <X className="h-4 w-4" />
                  )}
                </Button>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
