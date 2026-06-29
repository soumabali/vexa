"use client";

import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AlertTriangle, Clock } from "lucide-react";

/**
 * SessionProvider
 *
 * Wraps the app and manages global session state.
 * Renders the timeout warning dialog automatically.
 */
export function SessionProvider({ children }: { children: React.ReactNode }) {
  const sessionState = useSession({
    accessTokenTtlMs: 15 * 60 * 1000,
    refreshThresholdMs: 2 * 60 * 1000,
    warningBeforeMs: 60 * 1000,
  });

  return (
    <>
      {children}
      <SessionTimeoutModal sessionState={sessionState} />
    </>
  );
}

/**
 * SessionTimeoutModal
 *
 * Shows a countdown dialog before session expiry.
 * Auto-fires logout when countdown reaches 0.
 */
export function SessionTimeoutModal({
  sessionState,
}: {
  sessionState: ReturnType<typeof useSession>;
}) {
  const { showTimeoutWarning, timeoutCountdown, dismissWarning, logout } = sessionState;
  const [open, setOpen] = useState(false);

  useEffect(() => {
    Promise.resolve().then(() => setOpen(showTimeoutWarning));
  }, [showTimeoutWarning]);

  const handleStayLoggedIn = () => {
    setOpen(false);
    dismissWarning();
  };

  const handleLogoutNow = () => {
    setOpen(false);
    logout();
  };

  const minutes = Math.floor(timeoutCountdown / 60);
  const seconds = timeoutCountdown % 60;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader className="text-center space-y-3">
          <div className="mx-auto p-3 bg-yellow-100 rounded-full w-fit">
            <AlertTriangle className="h-8 w-8 text-yellow-600" />
          </div>
          <DialogTitle>Session Expiring</DialogTitle>
          <DialogDescription className="text-center">
            Your session will expire in{" "}
            <span className="font-mono font-bold text-foreground">
              {minutes > 0 ? `${minutes}m ` : ""}{seconds}s
            </span>
            . Would you like to stay logged in?
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button onClick={handleStayLoggedIn} className="w-full">
            Stay Logged In
          </Button>
          <Button variant="ghost" onClick={handleLogoutNow} className="w-full text-muted-foreground">
            Log Out Now
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ─── Inline session timer badge ─────────────────────────────────────────────

/**
 * SessionTimerBadge
 *
 * Small in-page display showing time remaining.
 * Use inside dashboard or layout to show live token countdown.
 */
export function SessionTimerBadge({ timeRemainingMs }: { timeRemainingMs: number }) {
  const [label, setLabel] = useState("");

  useEffect(() => {
    const update = () => {
      const total = Math.max(0, timeRemainingMs - Date.now());
      const min = Math.floor(total / 60000);
      const sec = Math.floor((total % 60000) / 1000);
      if (min > 5) setLabel("");
      else if (min > 0) setLabel(`${min}m ${sec}s`);
      else setLabel(`${sec}s`);
    };
    update();
    const id = setInterval(update, 1000);
    return () => clearInterval(id);
  }, [timeRemainingMs]);

  if (!label) return null;

  return (
    <div className="flex items-center gap-1.5 text-xs text-yellow-600 bg-yellow-50 px-2 py-1 rounded-full border border-yellow-200">
      <Clock className="h-3 w-3" />
      <span>Session: {label}</span>
    </div>
  );
}