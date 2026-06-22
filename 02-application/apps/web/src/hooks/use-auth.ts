"use client";

import { useEffect, useState } from "react";

interface User {
  id: string;
  email: string;
  role: string;
  mfaEnabled: boolean;
  totpEnabled: boolean;
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchMe() {
      try {
        const res = await fetch("/api/v1/users/me", { credentials: "include" });
        if (res.ok) {
          const data = await res.json();
          setUser({
            id: data.id,
            email: data.email,
            role: data.role,
            mfaEnabled: data.mfa_enabled ?? false,
            totpEnabled: data.totp_enabled ?? false,
          });
        } else {
          setUser(null);
        }
      } catch {
        setUser(null);
      } finally {
        setLoading(false);
      }
    }
    fetchMe();
  }, []);

  return { user, loading, isAuthenticated: !!user };
}

export default useAuth;
