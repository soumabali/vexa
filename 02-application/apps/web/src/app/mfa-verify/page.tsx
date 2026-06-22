"use client";

import { useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { twoFactorCodeSchema, TwoFactorCodeInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { ShieldCheck, KeyRound, ArrowLeft } from "lucide-react";

function MFAVerifyContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const mfaToken = searchParams.get("token") || "";
  const returnTo = searchParams.get("return") || "/dashboard";

  const [error, setError] = useState("");
  const [mode, setMode] = useState<"totp" | "backup">("totp");

  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm<TwoFactorCodeInput>({
    resolver: zodResolver(twoFactorCodeSchema),
  });

  const onTOTPSubmit = async (data: TwoFactorCodeInput) => {
    try {
      const result = await authApi.verifyMFA({ mfa_token: mfaToken, totp_code: data.code });
      // Store token and redirect
      sessionStorage.setItem("ssh_access_token", result.access_token);
      router.push(returnTo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid verification code");
    }
  };

  const onBackupCodeSubmit = async () => {
    const code = getValues("code");
    try {
      const result = await authApi.verifyBackupCode({ mfa_token: mfaToken, code });
      sessionStorage.setItem("ssh_access_token", result.token);
      sessionStorage.setItem("ssh_user", JSON.stringify(result.user));
      router.push(returnTo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid backup code");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center space-y-3">
          <div className="flex justify-center">
            <div className="p-3 bg-primary/10 rounded-full">
              <ShieldCheck className="h-8 w-8 text-primary" />
            </div>
          </div>
          <CardTitle>Two-Factor Authentication</CardTitle>
          <CardDescription>
            Enter the 6-digit code from your authenticator app
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}

          {mode === "totp" ? (
            <form onSubmit={handleSubmit(onTOTPSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="code">Verification Code</Label>
                <Input
                  id="code"
                  placeholder="000000"
                  maxLength={6}
                  {...register("code")}
                  className="text-center text-2xl tracking-widest font-mono"
                  autoFocus
                />
                {errors.code && (
                  <p className="text-sm text-destructive">{errors.code.message}</p>
                )}
              </div>
              <Button type="submit" className="w-full">
                Verify
              </Button>
            </form>
          ) : (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="backup-code">Backup Code</Label>
                <Input
                  id="backup-code"
                  placeholder="XXXX-XXXX"
                  {...register("code")}
                  className="font-mono"
                  autoFocus
                />
              </div>
              <Button onClick={onBackupCodeSubmit} className="w-full">
                Verify with Backup Code
              </Button>
            </div>
          )}

          <div className="flex justify-between text-sm">
            {mode === "totp" ? (
              <button
                type="button"
                onClick={() => setMode("backup")}
                className="text-primary hover:underline"
              >
                Use a backup code
              </button>
            ) : (
              <button
                type="button"
                onClick={() => setMode("totp")}
                className="text-primary hover:underline"
              >
                Back to authenticator app
              </button>
            )}
          </div>
        </CardContent>
        <CardFooter className="flex justify-center">
          <a href="/login" className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
            <ArrowLeft className="h-3 w-3" />
            Back to login
          </a>
        </CardFooter>
      </Card>
    </div>
  );
}
export default function MFAVerifyPage() {
  return (
    <Suspense fallback={<div className="p-8 text-center">Loading...</div>}>
      <MFAVerifyContent />
    </Suspense>
  );
}
