import {
  startRegistration,
  startAuthentication,
} from "@simplewebauthn/browser";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type RegistrationJSON = any;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AuthenticationJSON = any;

const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080") + "/api/v1";

export interface WebAuthnCredential {
  id: string;
  name: string;
  transport: string[];
  isResidentKey: boolean;
  isBackupEligible: boolean;
  isBackedUp: boolean;
  authenticatorAttachment?: string;
  createdAt: string;
  lastUsedAt?: string;
}

export interface WebAuthnRegisterBeginResponse {
  options: RegistrationJSON;
  challenge: string;
}

export interface WebAuthnLoginBeginResponse {
  options: AuthenticationJSON;
  challenge: string;
}

function getAuthHeaders(): Record<string, string> {
  const token = typeof window !== "undefined" ? localStorage.getItem("access_token") : null;
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

export async function beginRegistration(name?: string, requirePlatform?: boolean): Promise<WebAuthnRegisterBeginResponse> {
  const res = await fetch(`${API_BASE}/auth/webauthn/register/begin`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...getAuthHeaders() },
    body: JSON.stringify({ name, require_platform: requirePlatform }),
  });
  return handleResponse<WebAuthnRegisterBeginResponse>(res);
}

export async function finishRegistration(
  credential: RegistrationJSON,
  name?: string
): Promise<{ credential: WebAuthnCredential; message: string }> {
  const res = await fetch(`${API_BASE}/auth/webauthn/register/finish`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...getAuthHeaders() },
    body: JSON.stringify({
      name,
      id: credential.id,
      rawId: credential.rawId,
      type: credential.type,
      response: credential.response,
      authenticatorAttachment: credential.authenticatorAttachment,
      clientExtensionResults: credential.clientExtensionResults,
    }),
  });
  return handleResponse(res);
}

export async function registerCredential(name?: string, requirePlatform?: boolean) {
  const { options } = await beginRegistration(name, requirePlatform);
  const credential = await startRegistration({ optionsJSON: options });
  return finishRegistration(credential, name);
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

export async function beginLogin(email?: string): Promise<WebAuthnLoginBeginResponse> {
  const res = await fetch(`${API_BASE}/auth/webauthn/login/begin`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  return handleResponse<WebAuthnLoginBeginResponse>(res);
}

export async function finishLogin(credential: AuthenticationJSON): Promise<{
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}> {
  const res = await fetch(`${API_BASE}/auth/webauthn/login/finish`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: credential.id,
      rawId: credential.rawId,
      type: credential.type,
      response: credential.response,
      authenticatorAttachment: credential.authenticatorAttachment,
      clientExtensionResults: credential.clientExtensionResults,
    }),
  });
  return handleResponse(res);
}

export async function loginWithWebAuthn(email?: string) {
  const { options } = await beginLogin(email);
  const credential = await startAuthentication({ optionsJSON: options });
  return finishLogin(credential);
}

// ---------------------------------------------------------------------------
// Credential Management
// ---------------------------------------------------------------------------

export async function listCredentials(): Promise<{ credentials: WebAuthnCredential[] }> {
  const res = await fetch(`${API_BASE}/auth/webauthn/credentials`, {
    headers: getAuthHeaders(),
  });
  return handleResponse(res);
}

export async function deleteCredential(id: string): Promise<{ message: string }> {
  const res = await fetch(`${API_BASE}/auth/webauthn/credentials/${id}`, {
    method: "DELETE",
    headers: getAuthHeaders(),
  });
  return handleResponse(res);
}

export async function renameCredential(id: string, name: string): Promise<{ message: string }> {
  const res = await fetch(`${API_BASE}/auth/webauthn/credentials/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json", ...getAuthHeaders() },
    body: JSON.stringify({ name }),
  });
  return handleResponse(res);
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function isWebAuthnSupported(): boolean {
  return typeof window !== "undefined" && "PublicKeyCredential" in window;
}

export async function isPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (!isWebAuthnSupported()) return false;
  try {
    return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
  } catch {
    return false;
  }
}

export function isConditionalMediationAvailable(): boolean {
  if (!isWebAuthnSupported()) return false;
  return "isConditionalMediationAvailable" in PublicKeyCredential;
}
