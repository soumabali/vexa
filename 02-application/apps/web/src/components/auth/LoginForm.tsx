"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { loginSchema, LoginInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import Link from "next/link";
import { KeyRound, Eye, EyeOff, ShieldCheck } from "lucide-react";

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

  // ── Step 1: credentials ──────────────────────────────────────────────────

  const handleCredentialsSubmit = async (data: LoginInput) => {
    if (isLoading) return;
    setIsLoading(true);
    setError("");
    try {
      const response = await authApi.loginStep1({
        email: data.email.trim(),
        password: data.password,
        remember_me: rememberMe,
      });

      if (response.access_token) {
        saveAuthToken(response.access_token);
        router.push("/hosts");
      } else if (response.mfa_required && response.mfa_token) {
        setMfaChallenge({
          mfa_token: response.mfa_token,
          mfa_type: (response.mfa_type as "totp" | "webauthn" | "backup_code") ?? "totp",
          webauthn_challenge: response.webauthn_challenge,
          webauthn_credential_id: response.webauthn_credential_id,
        });
        setStep("mfa");
      } else {
        router.push("/hosts");
      }
    } catch {
      setError("Invalid email or password");
    } finally {
      setIsLoading(false);
    }
  };

  // ── Step 2: TOTP code ────────────────────────────────────────────────────

  const handleMFASubmit = async (code: string) => {
    if (!mfaChallenge) return;
    setIsLoading(true);
    setError("");
    try {
      await authApi.verifyMFA({
        mfa_token: mfaChallenge.mfa_token,
        code,
        challenge: mfaChallenge.webauthn_challenge,
      });
      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "MFA verification failed");
    } finally {
      setIsLoading(false);
    }
  };

  // ── Step 2-alt: WebAuthn assertion ──────────────────────────────────────

  const handleWebAuthn = async () => {
    if (!mfaChallenge?.webauthn_challenge) return;
    setIsLoading(true);
    setError("");
    try {
      const challengeBytes = Uint8Array.from(atob(mfaChallenge.webauthn_challenge), c => c.charCodeAt(0));
      const allowCredentials = mfaChallenge.webauthn_credential_id
        ? [{
            id: Uint8Array.from(atob(mfaChallenge.webauthn_credential_id), c => c.charCodeAt(0)),
            type: "public-key" as const,
          }]
        : [];

      const credential = (await navigator.credentials.get({
        publicKey: {
          challenge: challengeBytes,
          allowCredentials,
          userVerification: "preferred",
        },
      })) as PublicKeyCredential | null;

      if (!credential) {
        setError("No credential — user cancelled or browser blocked");
        setIsLoading(false);
        return;
      }

      const rawIdBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(credential.rawId))));
      const assertionResponse = credential.response as AuthenticatorAssertionResponse;
      const clientDataBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.clientDataJSON))));
      const signatureBase64 = btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.signature))));

      await authApi.verifyWebAuthn({
        mfa_token: mfaChallenge.mfa_token,
        credential_id: rawIdBase64,
        authenticator_data: btoa(String.fromCharCode(...Array.from(new Uint8Array(assertionResponse.authenticatorData)))),
        client_data: clientDataBase64,
        signature: signatureBase64,
      });

      router.push("/hosts");
    } catch (err) {
      setError(err instanceof Error ? err.message : "WebAuthn verification failed");
    } finally {
      setIsLoading(false);
    }
  };

  // ── Step 2-alt: backup code ─────────────────────────────────────────────

  const handleBackupCode = async (code: string) => {
    if (!mfaChallenge) return;
    setIsLoading(true);
    setError("");
    try {
      await authApi.verifyBackupCode({ mfa_token: mfaChallenge.mfa_token, code });
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
    <Card className="w-full max-w-md">
      <CardHeader className="text-center space-y-3">
        <div className="flex justify-center">
          <div className="p-3 bg-primary/10 rounded-full">
            <ShieldCheck className="h-8 w-8 text-primary" />
          </div>
        </div>
        <CardTitle className="text-2xl">
          {step === "credentials" ? "Welcome Back" : "Two-Factor Authentication"}
        </CardTitle>
        <CardDescription>
          {step === "credentials"
            ? "Sign in to your vexa account"
            : "Enter the code from your authenticator app"}
        </CardDescription>
      </CardHeader>

      <CardContent>
        {error && <ErrorDisplay message={error} />}

        {step === "credentials" && (
          <form onSubmit={handleSubmit(handleCredentialsSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" autoComplete="email" placeholder="you@example.com" {...register("email")} />
              {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <Link href="/forgot-password" className="text-xs text-primary hover:underline">
                  Forgot password?
                </Link>
              </div>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  placeholder="••••••••"
                  {...register("password")}
                  className="pr-10"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && <p className="text-sm text-destructive">{errors.password.message}</p>}
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="rememberMe"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="rounded border-gray-300"
              />
              <Label htmlFor="rememberMe" className="text-sm cursor-pointer">
                Remember me for 30 days
              </Label>
            </div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? <LoadingSpinner size="sm" /> : "Sign In"}
            </Button>
          </form>
        )}

        {step === "mfa" && (
          <MFAStep
            mfaChallenge={mfaChallenge}
            onSubmit={handleMFASubmit}
            onWebAuthn={handleWebAuthn}
            onBackupCode={handleBackupCode}
            onBack={resetToCredentials}
            isLoading={isLoading}
          />
        )}
      </CardContent>

      <CardFooter className="flex justify-center">
        {step === "credentials" && (
          <p className="text-sm text-muted-foreground">
            Don&apos;t have an account?{" "}
            <Link href="/register" className="text-primary hover:underline">Create one</Link>
          </p>
        )}
      </CardFooter>
    </Card>
  );
}

// ─── MFA Step ────────────────────────────────────────────────────────────────

function MFAStep({
  mfaChallenge,
  onSubmit,
  onWebAuthn,
  onBackupCode,
  onBack,
  isLoading,
}: {
  mfaChallenge: MFAChallenge | null;
  onSubmit: (code: string) => void;
  onWebAuthn: () => void;
  onBackupCode?: (code: string) => void;
  onBack: () => void;
  isLoading: boolean;
}) {
  const [code, setCode] = useState("");
  const [mode, setMode] = useState<"totp" | "backup">("totp");

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value.replace(/\D/g, "").slice(0, 6);
    setCode(val);
    if (val.length === 6 && mode === "totp") {
      onSubmit(val);
    }
  };

  return (
    <div className="space-y-4">
      {mode === "totp" ? (
        <>
          <div className="space-y-2">
            <Label htmlFor="totp-code">Verification Code</Label>
            <Input
              id="totp-code"
              placeholder="000000"
              value={code}
              onChange={handleCodeChange}
              maxLength={6}
              className="text-center text-2xl tracking-widest font-mono"
              disabled={isLoading}
              autoFocus
            />
          </div>

          {mfaChallenge?.mfa_type === "webauthn" && onWebAuthn && (
            <Button variant="outline" className="w-full" onClick={onWebAuthn} disabled={isLoading}>
              <KeyRound className="h-4 w-4 mr-2" />
              Use Security Key
            </Button>
          )}

          {onBackupCode && (
            <button
              type="button"
              onClick={() => setMode("backup")}
              className="w-full text-sm text-primary hover:underline"
            >
              Use a backup code instead
            </button>
          )}
        </>
      ) : (
        <>
          <div className="space-y-2">
            <Label htmlFor="backup-code">Backup Code</Label>
            <Input
              id="backup-code"
              placeholder="XXXX-XXXX"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={isLoading}
              autoFocus
            />
          </div>
          <Button
            onClick={() => onBackupCode?.(code)}
            className="w-full"
            disabled={isLoading || code.length < 8}
          >
            {isLoading ? <LoadingSpinner size="sm" /> : "Verify with Backup Code"}
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

      <Button variant="ghost" onClick={onBack} className="w-full text-muted-foreground">
        ← Back to login
      </Button>
    </div>
  );
}