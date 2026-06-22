"use client";

import { useCallback, useState } from "react";
import { authApi } from "@/lib/api/auth";

export type WebAuthnStatus = "idle" | "waiting" | "success" | "error";

interface WebAuthnRegistrationOptions {
  userId: string;
  userName: string;
  userDisplayName: string;
}

interface WebAuthnCredentialDescriptor {
  id: ArrayBuffer;
  type: "public-key";
  transports?: AuthenticatorTransport[];
}

interface UseWebAuthnReturn {
  status: WebAuthnStatus;
  error: string | null;
  /**
   * Begin WebAuthn registration. Returns the registered credential ID.
   */
  register: (options: WebAuthnRegistrationOptions) => Promise<string | null>;
  /**
   * Assert an existing credential (login / MFA).
   */
  assert: (credentialId: ArrayBuffer, challenge: ArrayBuffer, allowList?: WebAuthnCredentialDescriptor[]) => Promise<PublicKeyCredential | null>;
  /**
   * List registered security keys.
   */
  listDevices: () => Promise<{ id: string; name: string; created_at: string; last_used: string | null }[]>;
  /**
   * Remove a registered device.
   */
  removeDevice: (deviceId: string) => Promise<void>;
  /**
   * Check if WebAuthn is supported in this browser.
   */
  isSupported: boolean;
}

/**
 * useWebAuthn
 *
 * Provides a typed interface for WebAuthn registration and assertion.
 * Handles base64url encoding/decoding, credential creation & getting,
 * and maps to the vexa backend API.
 */
export function useWebAuthn(): UseWebAuthnReturn {
  const [status, setStatus] = useState<WebAuthnStatus>("idle");
  const [error, setError] = useState<string | null>(null);

  // ── Browser support check ───────────────────────────────────────────────────
  const isSupported = typeof navigator !== "undefined" &&
    typeof navigator.credentials !== "undefined" &&
    typeof PublicKeyCredential !== "undefined";

  // ── Helpers ────────────────────────────────────────────────────────────────

  /** Encode ArrayBuffer to base64url string (URL-safe, no padding) */
  const arrayBufferToBase64Url = (buffer: ArrayBuffer): string => {
    const bytes = new Uint8Array(buffer);
    let str = "";
    for (const b of bytes) str += String.fromCharCode(b);
    return btoa(str).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  };

  /** Decode base64url string to ArrayBuffer */
  const base64UrlToArrayBuffer = (str: string): ArrayBuffer => {
    str = str.replace(/-/g, "+").replace(/_/g, "/");
    while (str.length % 4) str += "=";
    const bin = atob(str);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes.buffer;
  };

  // ── Registration ──────────────────────────────────────────────────────────

  const register = useCallback(async (options: WebAuthnRegistrationOptions): Promise<string | null> => {
    setStatus("idle");
    setError(null);
    if (!isSupported) {
      setError("WebAuthn is not supported in this browser");
      return null;
    }

    try {
      // 1. Get registration options from backend
      const pubKeyOptions = await authApi.registerWebAuthn();

      const createCredentialOptions: PublicKeyCredentialCreationOptions = {
        challenge: base64UrlToArrayBuffer(pubKeyOptions.challenge),
        rp: {
          name: pubKeyOptions.rp.name,
          id: pubKeyOptions.rp.id,
        },
        user: {
          id: new TextEncoder().encode(options.userId),
          name: options.userName,
          displayName: options.userDisplayName,
        },
        pubKeyCredParams: pubKeyOptions.pubKeyCredParams.map(p => ({
          type: p.type as "public-key",
          alg: p.alg as number,
        })),
        timeout: pubKeyOptions.timeout,
        attestation: pubKeyOptions.attestation === "none" ? "none" : "indirect",
        authenticatorSelection: {
          authenticatorAttachment: "platform",
          userVerification: "preferred",
          residentKey: "preferred",
        },
      };

      // 2. Create credential in browser
      setStatus("waiting");
      const credential = (await navigator.credentials.create({
        publicKey: createCredentialOptions,
      })) as PublicKeyCredential | null;

      if (!credential) {
        setStatus("error");
        setError("No credential created — user cancelled or not supported");
        return null;
      }

      // 3. Decode attestation response
      const attestationResponse = credential.response as AuthenticatorAttestationResponse;
      const credentialId = arrayBufferToBase64Url(credential.rawId);

      // 4. Send attestation to backend for verification & storage
      await authApi.completeWebAuthnRegistration({
        credential_id: credentialId,
        client_data: arrayBufferToBase64Url(attestationResponse.clientDataJSON),
        attest_data: arrayBufferToBase64Url(attestationResponse.attestationObject),
      });

      setStatus("success");
      return credentialId;
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "WebAuthn registration failed");
      return null;
    }
  }, [isSupported]);

  // ── Assertion (login / MFA) ─────────────────────────────────────────────────

  const assert = useCallback(async (
    credentialId: ArrayBuffer,
    challenge: ArrayBuffer,
    allowList?: WebAuthnCredentialDescriptor[]
  ): Promise<PublicKeyCredential | null> => {
    setStatus("idle");
    setError(null);
    if (!isSupported) {
      setError("WebAuthn is not supported in this browser");
      return null;
    }

    try {
      setStatus("waiting");
      const publicKeyOptions: PublicKeyCredentialRequestOptions = {
        challenge,
        allowCredentials: allowList ?? [{
          id: credentialId,
          type: "public-key",
        }],
        userVerification: "preferred",
        timeout: 60000,
      };

      const credential = (await navigator.credentials.get({
        publicKey: publicKeyOptions,
      })) as PublicKeyCredential | null;

      if (!credential) {
        setStatus("error");
        setError("No credential — user cancelled or not supported");
        return null;
      }

      setStatus("success");
      return credential;
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "WebAuthn assertion failed");
      return null;
    }
  }, [isSupported]);

  // ── Device management ──────────────────────────────────────────────────────

  const listDevices = useCallback(async () => {
    const result = await authApi.listWebAuthnDevices();
    return result.devices;
  }, []);

  const removeDevice = useCallback(async (deviceId: string) => {
    await authApi.removeWebAuthnDevice(deviceId);
  }, []);

  return {
    status,
    error,
    register,
    assert,
    listDevices,
    removeDevice,
    isSupported,
  };
}

// ─── Helper: encode base64url ────────────────────────────────────────────────

export function base64UrlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let str = "";
  for (const b of bytes) str += String.fromCharCode(b);
  return btoa(str).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

export function base64UrlDecode(str: string): ArrayBuffer {
  str = str.replace(/-/g, "+").replace(/_/g, "/");
  while (str.length % 4) str += "=";
  const bin = atob(str);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return bytes.buffer;
}