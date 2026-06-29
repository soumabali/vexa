"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { useAsyncData } from "@/hooks/useAsyncData";
import { authApi } from "@/lib/api/auth";
import { toast } from "sonner";
import { Loader2, Monitor, Trash2, MapPin, Clock } from "lucide-react";

interface SessionMetadata {
  session_id: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  last_active_at: string;
  is_current: boolean;
}

function parseUserAgent(ua: string): { browser: string; os: string } {
  if (!ua) return { browser: "Unknown browser", os: "Unknown OS" };
  const lower = ua.toLowerCase();
  let browser = "Unknown browser";
  if (lower.includes("firefox/")) browser = "Firefox";
  else if (lower.includes("edg/")) browser = "Edge";
  else if (lower.includes("chrome/")) browser = "Chrome";
  else if (lower.includes("safari/") && !lower.includes("chrome/")) browser = "Safari";
  let os = "Unknown OS";
  if (lower.includes("windows")) os = "Windows";
  else if (lower.includes("mac os") || lower.includes("macintosh")) os = "macOS";
  else if (lower.includes("linux")) os = "Linux";
  else if (lower.includes("android")) os = "Android";
  else if (lower.includes("iphone") || lower.includes("ipad")) os = "iOS";
  return { browser, os };
}

function formatRelative(iso: string): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const diff = Date.now() - then;
  const min = Math.round(diff / 60000);
  if (min < 1) return "just now";
  if (min < 60) return `${min} min ago`;
  const hr = Math.round(min / 60);
  if (hr < 24) return `${hr} hr ago`;
  const day = Math.round(hr / 24);
  return `${day} day${day === 1 ? "" : "s"} ago`;
}

export default function SessionsPage() {
  const { data, loading, error, reload } = useAsyncData<SessionMetadata[]>(
    async () => {
      const sessions = await authApi.listSessions();
      return sessions;
    },
    []
  );
  const [revokeTarget, setRevokeTarget] = useState<SessionMetadata | null>(null);
  const [revoking, setRevoking] = useState(false);

  const handleRevoke = async () => {
    if (!revokeTarget) return;
    setRevoking(true);
    try {
      await authApi.revokeSession(revokeTarget.session_id);
      toast.success("Session revoked.");
      setRevokeTarget(null);
      reload();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to revoke session.");
    } finally {
      setRevoking(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Active Sessions</h1>
          <p className="text-muted-foreground mt-2">
            Review and revoke active sessions across your devices.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Devices</CardTitle>
            <CardDescription>
              Sessions that are currently signed in to your account.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex items-center gap-2 text-muted-foreground py-6">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading sessions…
              </div>
            ) : error ? (
              <p className="text-sm text-destructive py-6">
                {error.message || "Failed to load sessions."}
              </p>
            ) : !data || data.length === 0 ? (
              <p className="text-sm text-muted-foreground py-6">
                No active sessions.
              </p>
            ) : (
              <div className="space-y-3">
                {data.map((session) => {
                  const { browser, os } = parseUserAgent(session.user_agent);
                  return (
                    <div
                      key={session.session_id}
                      className="flex items-center justify-between rounded-lg border p-4"
                      data-testid="session-row"
                    >
                      <div className="flex items-center gap-3">
                        <Monitor className="h-5 w-5 text-primary" />
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <p className="font-medium">
                              {browser} on {os}
                            </p>
                            {session.is_current ? (
                              <Badge variant="outline" className="text-green-600 border-green-200">
                                This device
                              </Badge>
                            ) : null}
                          </div>
                          <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <MapPin className="h-3 w-3" />
                              {session.ip_address || "Unknown IP"}
                            </span>
                            <span className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              Last active {formatRelative(session.last_active_at)}
                            </span>
                          </div>
                        </div>
                      </div>
                      {session.is_current ? null : (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setRevokeTarget(session)}
                          aria-label={`Revoke ${browser} session`}
                          data-testid="revoke-button"
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Dialog
        open={revokeTarget !== null}
        onOpenChange={(open) => {
          if (!open && !revoking) setRevokeTarget(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke this session?</DialogTitle>
            <DialogDescription>
              The selected device will be signed out immediately. This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRevokeTarget(null)}
              disabled={revoking}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleRevoke}
              disabled={revoking}
              data-testid="confirm-revoke"
            >
              {revoking ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : null}
              Revoke session
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}