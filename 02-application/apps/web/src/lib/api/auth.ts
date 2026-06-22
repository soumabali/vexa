import { z } from "zod";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";

// Shared request helper
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    credentials: "include",
    ...options,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({
      message: "An error occurred",
    }));
    throw new Error(error.message || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}

// ─── Login flow ────────────────────────────────────────────────────────────────

export interface LoginStep1Response {
  mfa_required: boolean;
  mfa_token?: string;
  mfa_type?: "totp" | "webauthn" | "backup_code";
  webauthn_challenge?: string;
  webauthn_credential_id?: string;
  // Direct token response (when MFA is not enabled)
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
  token_type?: string;
}

export interface TokenPairResponse {
  access_token: string;
  refresh_token: string;
  expires_in?: number;
  token_type?: string;
}

export interface MFAToken {
  mfa_token: string;
  totp_code: string;
  challenge?: string;
}

export const authApi = {
  // Step 1: submit credentials → may return MFA challenge
  loginStep1: (data: { email: string; password: string; remember_me?: boolean }) =>
    apiRequest<LoginStep1Response>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Step 2: verify TOTP code
  verifyMFA: (data: { mfa_token: string; totp_code: string; challenge?: string }) =>
    apiRequest<TokenPairResponse>("/api/v1/auth/mfa/verify", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Step 2 alternative: WebAuthn assertion
  verifyWebAuthn: (data: {
    mfa_token: string;
    credential_id: string;
    authenticator_data: string;
    client_data: string;
    signature: string;
    timestamp?: number;
  }) =>
    apiRequest<{ token: string; user: { id: string; email: string; name: string } }>("/api/v1/auth/webauthn/verify", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Step 2 alternative: backup code
  verifyBackupCode: (data: { mfa_token: string; code: string }) =>
    apiRequest<{ token: string; user: { id: string; email: string; name: string } }>("/api/v1/auth/mfa/backup-code", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  refreshToken: (data: { refresh_token: string }) =>
    apiRequest<{ token: string; refresh_token?: string }>("/api/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  logout: () =>
    apiRequest<{ message: string }>("/api/v1/auth/logout", {
      method: "POST",
    }),

  // ─── User registration & verification ───────────────────────────────────────

  register: (data: { email: string; password: string }) =>
    apiRequest<{ message: string; userId: string }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  verifyEmail: (data: { email: string; code: string }) =>
    apiRequest<{ message: string }>("/api/v1/auth/verify-email", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  resendVerification: (email: string) =>
    apiRequest<{ message: string }>("/api/v1/auth/resend-verification", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),

  // ─── Password reset ──────────────────────────────────────────────────────────

  forgotPassword: (email: string) =>
    apiRequest<{ message: string }>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),

  verifyResetCode: (data: { email: string; code: string }) =>
    apiRequest<{ message: string; token: string }>("/api/v1/auth/verify-reset-code", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  resetPassword: (data: { token: string; password: string }) =>
    apiRequest<{ message: string }>("/api/v1/auth/reset-password", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // ─── Profile ─────────────────────────────────────────────────────────────────

  getProfile: () =>
    apiRequest<{
      id: string;
      email: string;
      name: string;
      avatar?: string;
      createdAt: string;
    }>("/api/v1/auth/profile"),

  updateProfile: (data: { name: string; email: string }) =>
    apiRequest<{ message: string; user: { id: string; email: string; name: string } }>("/api/v1/auth/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  // ─── User profile API ─────────────────────────────────────────────────
  getUserProfile: () =>
    apiRequest<{ id: string; email: string; role: string; mfa_enabled: boolean; totp_enabled: boolean }>("/api/v1/users/me", {
      method: "GET",
    }),

  updateUserProfile: (data: { name?: string; email?: string }) =>
    apiRequest<{ message: string }>("/api/v1/users/me", {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  changePassword: (data: {
    currentPassword: string;
    newPassword: string;
  }) =>
    apiRequest<{ message: string }>("/api/v1/auth/change-password", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteAccount: (password: string) =>
    apiRequest<{ message: string }>("/api/v1/auth/account", {
      method: "DELETE",
      body: JSON.stringify({ password }),
    }),

  // ─── MFA / 2FA ────────────────────────────────────────────────────────────────

  setup2FA: () =>
    apiRequest<{ secret: string; qrCode: string; uri: string; backupCodes: string[] }>("/api/v1/auth/mfa/setup", {
      method: "POST",
    }),

  verify2FASetup: (totp_code: string) =>
    apiRequest<{ message: string }>("/api/v1/auth/mfa/enable", {
      method: "POST",
      body: JSON.stringify({ totp_code }),
    }),

  disable2FA: () =>
    apiRequest<{ message: string }>("/api/v1/auth/mfa/disable", {
      method: "DELETE",
    }),

  // ─── WebAuthn (security key) ──────────────────────────────────────────────────

  /**
   * Start WebAuthn registration. Returns challenge + credential options so the
   * browser can call navigator.credentials.create({ publicKey: ... })
   */
  registerWebAuthn: () =>
    apiRequest<{
      challenge: string;
      rp: { name: string; id: string };
      user: { id: string; name: string; displayName: string };
      pubKeyCredParams: { type: string; alg: number }[];
      timeout: number;
      attestation?: string;
    }>("/api/v1/auth/webauthn/register/start", {
      method: "POST",
    }),

  /**
   * Complete WebAuthn registration by sending the attestation response.
   */
  completeWebAuthnRegistration: (data: {
    credential_id: string;
    client_data: string;
    attest_data: string;
  }) =>
    apiRequest<{ message: string }>("/api/v1/auth/webauthn/register/complete", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  /**
   * List registered security keys (credential IDs + metadata).
   */
  listWebAuthnDevices: () =>
    apiRequest<{
      devices: {
        id: string;
        name: string;
        created_at: string;
        last_used: string | null;
        credential_id: string;
      }[];
    }>("/api/v1/auth/webauthn/devices"),

  /**
   * Remove a registered security key.
   */
  removeWebAuthnDevice: (deviceId: string) =>
    apiRequest<{ message: string }>(`/auth/webauthn/devices/${deviceId}`, {
      method: "DELETE",
    }),

  // ─── Sessions ────────────────────────────────────────────────────────────────

  getSessions: () =>
    apiRequest<
      {
        id: string;
        device: string;
        browser: string;
        ip: string;
        location: string;
        lastActive: string;
        isCurrent: boolean;
      }[]
    >("/api/v1/auth/sessions"),

  revokeSession: (sessionId: string) =>
    apiRequest<{ message: string }>(`/auth/sessions/${sessionId}`, {
      method: "DELETE",
    }),

  getLoginHistory: () =>
    apiRequest<
      {
        id: string;
        device: string;
        browser: string;
        ip: string;
        location: string;
        timestamp: string;
        status: "success" | "failed";
      }[]
    >("/api/v1/auth/login-history"),

  // ─── Biometric ────────────────────────────────────────────────────────────────

  updateBiometric: (enabled: boolean) =>
    apiRequest<{ message: string }>("/api/v1/auth/biometric", {
      method: "PUT",
      body: JSON.stringify({ enabled }),
    }),
};

// ─── Validation helpers ───────────────────────────────────────────────────────

export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

export const registerSchema = z
  .object({
    email: z.string().email("Invalid email address"),
    password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .regex(/[A-Z]/, "Password must contain at least one uppercase letter")
      .regex(/[a-z]/, "Password must contain at least one lowercase letter")
      .regex(/[0-9]/, "Password must contain at least one number")
      .regex(/[^A-Za-z0-9]/, "Password must contain at least one special character"),
    confirmPassword: z.string(),
    acceptTerms: z.boolean().refine((val) => val === true, {
      message: "You must accept the terms and conditions",
    }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;