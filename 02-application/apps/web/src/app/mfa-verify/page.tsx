"use client";

import { useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { twoFactorCodeSchema, TwoFactorCodeInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MaterialIcon } from "@/components/ui/material-icon";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import { AuthLayout } from "@/components/auth/AuthLayout";

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
      const result = await authApi.verifyMFA({ mfa_token: mfaToken, totp_code: code });
      sessionStorage.setItem("ssh_access_token", result.access_token);
      router.push(returnTo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid backup code");
    }
  };

  return (
    <AuthLayout title="Two-Factor Authentication" description="Verify your identity to continue">
      <div className="w-full space-y-6">
        <div className="flex justify-center">
          <div className="w-14 h-14 rounded-full bg-primary-container flex items-center justify-center">
            <MaterialIcon name="shield_lock" size="lg" className="text-on-primary-container" />
          </div>
        </div>

        {error && <ErrorDisplay message={error} />}

        {mode === "totp" ? (
          <form onSubmit={handleSubmit(onTOTPSubmit)} className="space-y-4">
            <div className="space-y-2">
              <p className="text-body-md text-on-surface-variant text-center">
                Enter the 6-digit code from your authenticator app.
              </p>
              <Input
                type="text"
                inputMode="numeric"
              maxLength={6}
              placeholder="000000"
              className="bg-surface-container-high border-outline-variant text-on-surface text-center text-2xl tracking-[0.5em]"
              {...register("code")}
            />
            {errors.code && (
              <p className="text-sm text-error text-center">{errors.code.message}</p>
            )}
          </div>
          <Button
            type="submit"
            className="w-full bg-primary text-on-primary hover:bg-primary/90"
          >
            Verify Code
          </Button>
        </form>
      ) : (
        <form onSubmit={handleSubmit(onBackupCodeSubmit)} className="space-y-4">
          <div className="space-y-2">
            <p className="text-body-md text-on-surface-variant text-center">
              Enter one of your backup codes.
            </p>
            <Input
              type="text"
              placeholder="Enter backup code"
              className="bg-surface-container-high border-outline-variant text-on-surface"
              {...register("code")}
            />
            {errors.code && (
              <p className="text-sm text-error text-center">{errors.code.message}</p>
            )}
          </div>
          <Button
            type="submit"
            className="w-full bg-primary text-on-primary hover:bg-primary/90"
          >
            Verify Backup Code
          </Button>
        </form>
      )}

      <div className="text-center pt-2">
        <button
          onClick={() => setMode(mode === "totp" ? "backup" : "totp")}
          className="text-sm text-primary hover:underline"
        >
          {mode === "totp" ? "Use backup code instead" : "Use authenticator code instead"}
        </button>
      </div>
    </div>
    </AuthLayout>
  );
}

export default function MFAVerifyPage() {
  return (
    <Suspense>
      <MFAVerifyContent />
    </Suspense>
  );
}