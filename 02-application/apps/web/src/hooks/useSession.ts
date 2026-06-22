"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { authApi } from "@/lib/api/auth";

const DEFAULT_ACCESS_TOKEN_TTL_MS = 15 * 60 * 1000;   // 15 minutes
const DEFAULT_REFRESH_THRESHOLD_MS = 2 * 60 * 1000;     // refresh 2 min before expiry
const DEFAULT_WARNING_BEFORE_MS = 60 * 1000;            // warn 1 min before expiry

export interface SessionState {
  isAuthenticated: boolean;
  accessToken: string | null;
  expiresAt: number | null;   // Unix ms
  user: { id: string; email: string; name: string } | null;
}

export interface UseSessionReturn {
  session: SessionState;
  /** Force a token refresh */
  refresh: () => Promise<void>;
  /** Extend session (e.g. user clicks "Stay logged in" on warning) */
  extendSession: (newExpiresAt: number) => void;
  /** Immediately end the session */
  logout: () => Promise<void>;
  /** Time remaining in ms until token expires (0 if not authenticated) */
  timeRemaining: number;
  /** Whether to show the timeout warning overlay */
  showTimeoutWarning: boolean;
  /** Seconds remaining before forced logout (countdown) */
  timeoutCountdown: number;
  /** Dismiss the warning and refresh the session */
  dismissWarning: () => void;
}

/**
 * useSession
 *
 * Manages access token lifecycle:
 * - Stores token + expiry in memory (optionally in sessionStorage)
 * - Auto-refreshes when within `refreshThresholdMs` of expiry
 * - Shows a warning dialog `warningBeforeMs` before expiry
 * - Countdown on the warning, auto-logout when expiry is reached
 *
 * Designed to be used inside a React context provider wrapping the app.
 */
export function useSession(
  options: {
    accessTokenTtlMs?: number;
    refreshThresholdMs?: number;
    warningBeforeMs?: number;
    onAuthError?: (err: Error) => void;
  } = {}
): UseSessionReturn {
  const {
    accessTokenTtlMs = DEFAULT_ACCESS_TOKEN_TTL_MS,
    refreshThresholdMs = DEFAULT_REFRESH_THRESHOLD_MS,
    warningBeforeMs = DEFAULT_WARNING_BEFORE_MS,
    onAuthError,
  } = options;

  // ── State ──────────────────────────────────────────────────────────────────
  const [session, setSession] = useState<SessionState>({
    isAuthenticated: false,
    accessToken: null,
    expiresAt: null,
    user: null,
  });

  const [showTimeoutWarning, setShowTimeoutWarning] = useState(false);
  const [timeoutCountdown, setTimeoutCountdown] = useState(-1);

  // Refs to avoid stale closures in timers
  const expiresAtRef = useRef<number | null>(null);
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const warningTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // ── Helpers ────────────────────────────────────────────────────────────────

  const clearAllTimers = useCallback(() => {
    if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current);
    if (warningTimerRef.current) clearTimeout(warningTimerRef.current);
    if (countdownIntervalRef.current) clearInterval(countdownIntervalRef.current);
  }, []);

  const scheduleRefresh = useCallback((expiresAt: number) => {
    clearAllTimers();
    const refreshAt = expiresAt - refreshThresholdMs;
    const delay = Math.max(0, refreshAt - Date.now());

    refreshTimerRef.current = setTimeout(async () => {
      try {
        const data = await authApi.refreshToken({ refresh_token: "" });
        const newExpiry = Date.now() + accessTokenTtlMs;
        setSession(prev => ({
          ...prev,
          accessToken: data.token,
          expiresAt: newExpiry,
        }));
        expiresAtRef.current = newExpiry;
        scheduleRefresh(newExpiry);
      } catch (err) {
        // Refresh token invalid — force logout
        onAuthError?.(err instanceof Error ? err : new Error("Session expired"));
        setSession({ isAuthenticated: false, accessToken: null, expiresAt: null, user: null });
      }
    }, delay);
  }, [accessTokenTtlMs, refreshThresholdMs, clearAllTimers, onAuthError]);

  const scheduleWarning = useCallback((expiresAt: number) => {
    clearAllTimers();
    const warnAt = expiresAt - warningBeforeMs;
    const delay = Math.max(0, warnAt - Date.now());

    warningTimerRef.current = setTimeout(() => {
      setShowTimeoutWarning(true);
      setTimeoutCountdown(Math.ceil(warningBeforeMs / 1000));

      // Countdown tick
      countdownIntervalRef.current = setInterval(() => {
        setTimeoutCountdown(prev => {
          if (prev <= 1) {
            // Force logout when countdown reaches 0
            clearAllTimers();
            setSession({ isAuthenticated: false, accessToken: null, expiresAt: null, user: null });
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }, delay);

    // Also schedule refresh
    scheduleRefresh(expiresAt);
  }, [warningBeforeMs, clearAllTimers, scheduleRefresh]);

  const timeRemaining = session.expiresAt ? Math.max(0, session.expiresAt - Date.now()) : 0;

  // ── Public API ────────────────────────────────────────────────────────────

  const refresh = useCallback(async () => {
    try {
      const data = await authApi.refreshToken({ refresh_token: "" });
      const newExpiry = Date.now() + accessTokenTtlMs;
      setSession(prev => ({ ...prev, accessToken: data.token, expiresAt: newExpiry }));
      expiresAtRef.current = newExpiry;
      setShowTimeoutWarning(false);
      clearAllTimers();
      scheduleWarning(newExpiry);
    } catch (err) {
      onAuthError?.(err instanceof Error ? err : new Error("Refresh failed"));
      setSession({ isAuthenticated: false, accessToken: null, expiresAt: null, user: null });
    }
  }, [accessTokenTtlMs, clearAllTimers, scheduleWarning, onAuthError]);

  const extendSession = useCallback((newExpiresAt: number) => {
    setShowTimeoutWarning(false);
    clearAllTimers();
    expiresAtRef.current = newExpiresAt;
    setSession(prev => ({ ...prev, expiresAt: newExpiresAt }));
    scheduleWarning(newExpiresAt);
  }, [clearAllTimers, scheduleWarning]);

  const logout = useCallback(async () => {
    clearAllTimers();
    try {
      await authApi.logout();
    } catch {
      // Best-effort logout
    }
    setSession({ isAuthenticated: false, accessToken: null, expiresAt: null, user: null });
    expiresAtRef.current = null;
  }, [clearAllTimers]);

  const dismissWarning = useCallback(() => {
    refresh();
  }, [refresh]);

  // ── Mount: restore session from sessionStorage ──────────────────────────────
  useEffect(() => {
    try {
      const stored = sessionStorage.getItem("ssh_session");
      if (stored) {
        const parsed = JSON.parse(stored) as { token: string; expiresAt: number; user: { id: string; email: string; name: string } };
        if (parsed.expiresAt > Date.now()) {
          setSession({ isAuthenticated: true, accessToken: parsed.token, expiresAt: parsed.expiresAt, user: parsed.user });
          expiresAtRef.current = parsed.expiresAt;
          scheduleWarning(parsed.expiresAt);
        } else {
          sessionStorage.removeItem("ssh_session");
        }
      }
    } catch {
      // Ignore parse errors
    }
  }, [scheduleWarning]);

  // ── Persist session to sessionStorage on change ───────────────────────────
  useEffect(() => {
    if (session.isAuthenticated && session.accessToken && session.expiresAt) {
      sessionStorage.setItem("ssh_session", JSON.stringify({
        token: session.accessToken,
        expiresAt: session.expiresAt,
        user: session.user,
      }));
    } else {
      sessionStorage.removeItem("ssh_session");
    }
  }, [session]);

  // ── Cleanup on unmount ─────────────────────────────────────────────────────
  useEffect(() => {
    return () => clearAllTimers();
  }, [clearAllTimers]);

  return {
    session,
    refresh,
    extendSession,
    logout,
    timeRemaining,
    showTimeoutWarning,
    timeoutCountdown,
    dismissWarning,
  };
}

// ─── Session timeout warning modal ───────────────────────────────────────────

export interface SessionTimeoutWarningProps {
  countdown: number;      // seconds remaining
  onStayLoggedIn: () => void;
  onLogout: () => void;
}