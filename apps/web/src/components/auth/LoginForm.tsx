"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { loginSchema, LoginInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MaterialIcon } from "@/components/ui/material-icon";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import Link from "next/link";

export type LoginStep = "credentials" | "mfa";

interface MFAChallenge {
  mfa_token: string;
  mfa_type: "totp" | "webauthn" | "backup_code";
  webauthn_challenge?: string;
  webauthn_credential_id?: string;
}

function saveAuthToken(token: string) {
  if (typeof document !== "undefined") {
    document.cookie = `ssh_access_token=${token}; path=/; max-age=${30*24*60*60}`;
  }
  if (typeof localStorage !== "undefined") {
    localStorage.setItem('access_token', token);
    localStorage.setItem('token', token);
  }
}

export function LoginForm() {
  const router = useRouter();
  const [step, setStep] = useState<LoginStep>("credentials");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [mfaChallenge, setMfaChallenge] = useState<MFAChallenge | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });

  // Step 1: credentials

  const handleCredentialsSubmit = async (data: LoginInput) => {
    if (isLoading) return;
    setIsLoading(true);
    setError("");
    try {
      const result = await authApi.loginStep1({ email: data.email.trim(), password: data.password, remember_me: rememberMe });

      if (result.mfa_required) {
        setMfaChallenge({
          mfa_token: result.mfa_token || "",
          mfa_type: result.mfa_type || "totp",
          webauthn_challenge: result.webauthn_challenge,
          webauthn_credential_id: result.webauthn_credential_id,
        });
        setStep("mfa");
      } else {
        if (result.access_token) {
          saveAuthToken(result.access_token);
        }
        router.push("/hosts");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  // Step 2: TOTP

  const handleTOTPVerify = async (code: string) => {
    if (isLoading || !mfaChallenge) return;
    setIsLoading(true);
    setError("");
    try {
      const result = await authApi.verifyMFA({
        mfa_token: mfaChallenge.mfa_token,
        totp_code: code,
      });
      if (result.access_token) {
        saveAuthToken(result.access_token);
      }
      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "MFA verification failed");
    } finally {
      setIsLoading(false);
    }
  };

  // Step 2-alt: WebAuthn assertion

  const handleWebAuthnLogin = async () => {
    if (isLoading || !mfaChallenge) return;
    setIsLoading(true);
    setError("");
    try {
      const challenge = mfaChallenge.webauthn_challenge || "";
      const assertion = await navigator.credentials.get({
        publicKey: {
          challenge: Uint8Array.from(atob(challenge), (c) => c.charCodeAt(0)),
          rpId: window.location.hostname,
          allowCredentials: mfaChallenge.webauthn_credential_id
            ? [{ id: Uint8Array.from(atob(mfaChallenge.webauthn_credential_id), (c) => c.charCodeAt(0)), type: "public-key" }]
            : [],
          userVerification: "preferred",
        },
      }) as PublicKeyCredential | null;

      if (!assertion) {
        setError("WebAuthn assertion failed");
        setIsLoading(false);
        return;
      }

      const rawIdBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(assertion.rawId))));
      const assertionResponse = assertion.response as AuthenticatorAssertionResponse;
      const clientDataBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.clientDataJSON))));
      const signatureBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.signature))));

      await authApi.verifyWebAuthn({
        mfa_token: mfaChallenge.mfa_token,
        credential_id: rawIdBase64,
        authenticator_data: btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.authenticatorData)))),
        client_data: clientDataBase64,
        signature: signatureBase64,
      });

      // After WebAuthn verification, get token
      const result = await authApi.loginStep1({ email: "", password: "" });
      if (result.access_token) {
        saveAuthToken(result.access_token);
      }
      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "WebAuthn verification failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleBackupCodeVerify = async (code: string) => {
    if (isLoading || !mfaChallenge) return;
    setIsLoading(true);
    setError("");
    try {
      const result = await authApi.verifyBackupCode({
        mfa_token: mfaChallenge.mfa_token,
        code,
      });
      saveAuthToken(result.token);
      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid backup code");
    } finally {
      setIsLoading(false);
    }
  };

  const resetToCredentials = () => {
    setStep("credentials");
    setMfaChallenge(null);
    setError("");
  };

  return (
    <div className="space-y-6">
      {error && <ErrorDisplay message={error} />}

      {step === "credentials" && (
        <form onSubmit={handleSubmit(handleCredentialsSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email" className="text-on-surface">Email</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              className="bg-surface-container-high border-outline-variant text-on-surface"
              {...register("email")}
            />
            {errors.email && <p className="text-sm text-error">{errors.email.message}</p>}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="password" className="text-on-surface">Password</Label>
              <Link href="/forgot-password" className="text-sm text-primary hover:underline">
                Forgot Password?
              </Link>
            </div>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                placeholder="Enter your password"
                className="bg-surface-container-high border-outline-variant text-on-surface pr-10"
                {...register("password")}
              />
              <button
                type="button"
                aria-label={showPassword ? "Hide password" : "Show password"}
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-on-surface-variant hover:text-on-surface focus:outline-none focus-visible:ring-2 focus-visible:ring-primary rounded"
                tabIndex={-1}
              >
                <MaterialIcon name={showPassword ? "visibility_off" : "visibility"} size="sm" />
              </button>
            </div>
            {errors.password && <p className="text-sm text-error">{errors.password.message}</p>}
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="rememberMe"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              className="rounded border-outline-variant"
            />
            <Label htmlFor="rememberMe" className="text-sm text-on-surface-variant cursor-pointer">
              Remember me for 30 days
            </Label>
          </div>

          <Button type="submit" className="w-full bg-primary text-on-primary hover:bg-primary/90" disabled={isLoading}>
            {isLoading ? <LoadingSpinner size="sm" /> : "Sign In"}
          </Button>
        </form>
      )}

      {step === "mfa" && mfaChallenge && (
        <MFAStep
          mfaChallenge={mfaChallenge}
          isLoading={isLoading}
          onTOTPVerify={handleTOTPVerify}
          onWebAuthnVerify={handleWebAuthnLogin}
          onBackupCodeVerify={handleBackupCodeVerify}
          onBack={resetToCredentials}
        />
      )}

      {step === "credentials" && (
        <div className="text-center space-y-2 pt-2">
          <p className="text-sm text-on-surface-variant">
            Don&apos;t have an account?{" "}
            <Link href="/register" className="text-primary hover:underline font-medium">
              Create one
            </Link>
          </p>
        </div>
      )}
    </div>
  );
}

// ── MFA Step Sub-Component ──

function MFAStep({
  mfaChallenge,
  isLoading,
  onTOTPVerify,
  onWebAuthnVerify,
  onBackupCodeVerify,
  onBack,
}: {
  mfaChallenge: MFAChallenge;
  isLoading: boolean;
  onTOTPVerify: (code: string) => Promise<void>;
  onWebAuthnVerify: () => Promise<void>;
  onBackupCodeVerify: (code: string) => Promise<void>;
  onBack: () => void;
}) {
  const [mode, setMode] = useState<"totp" | "webauthn" | "backup_code">(
    mfaChallenge.mfa_type === "webauthn" ? "webauthn" : "totp"
  );
  const [code, setCode] = useState("");

  return (
    <div className="space-y-4">
      {mode === "totp" && (
        <>
          <p className="text-sm text-on-surface-variant text-center">
            Enter the 6-digit code from your authenticator app.
          </p>
          <Input
            type="text"
            inputMode="numeric"
            maxLength={6}
            placeholder="000000"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
            className="bg-surface-container-high border-outline-variant text-on-surface text-center text-2xl tracking-[0.5em]"
          />
          <Button
            onClick={() => onTOTPVerify(code)}
            className="w-full bg-primary text-on-primary hover:bg-primary/90"
            disabled={isLoading || code.length < 6}
          >
            {isLoading ? <LoadingSpinner size="sm" /> : "Verify"}
          </Button>
          {mfaChallenge.webauthn_challenge && (
            <button
              type="button"
              onClick={() => setMode("webauthn")}
              className="w-full text-sm text-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded"
            >
              Use security key instead
            </button>
          )}
        </>
      )}

      {mode === "webauthn" && (
        <>
          <p className="text-sm text-on-surface-variant text-center">
            Use your security key to verify your identity.
          </p>
          <Button
            onClick={onWebAuthnVerify}
            className="w-full bg-secondary-container text-on-secondary-container hover:bg-secondary-container/80"
            disabled={isLoading}
          >
            <MaterialIcon name="usb" size="sm" className="mr-2" />
            {isLoading ? <LoadingSpinner size="sm" /> : "Authenticate with Security Key"}
          </Button>
          <button
            type="button"
            onClick={() => setMode("totp")}
            className="w-full text-sm text-primary hover:underline"
          >
            Back to authenticator app
          </button>
        </>
      )}

      {mode === "backup_code" && (
        <>
          <p className="text-sm text-on-surface-variant text-center">
            Enter one of your 8-digit backup codes.
          </p>
          <Input
            type="text"
            maxLength={8}
            placeholder="XXXX-XXXX"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 8))}
            className="bg-surface-container-high border-outline-variant text-on-surface text-center text-xl tracking-[0.3em]"
          />
          <Button
            onClick={() => onBackupCodeVerify(code)}
            className="w-full bg-primary text-on-primary hover:bg-primary/90"
            disabled={isLoading || code.length < 8}
          >
            {isLoading ? <LoadingSpinner size="sm" /> : "Verify"}
          </Button>
          <button
            type="button"
            onClick={() => setMode("totp")}
            className="w-full text-sm text-primary hover:underline"
          >
            Back to authenticator app
          </button>
        </>
      )}

      <Button variant="ghost" onClick={onBack} className="w-full text-on-surface-variant">
        <MaterialIcon name="arrow_back" size="sm" className="mr-1" />
        Back to login
      </Button>
    </div>
  );
}
