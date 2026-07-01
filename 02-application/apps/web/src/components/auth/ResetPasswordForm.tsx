"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { resetPasswordSchema, ResetPasswordInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MaterialIcon } from "@/components/ui/material-icon";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import Link from "next/link";

export function ResetPasswordForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordInput>({
    resolver: zodResolver(resetPasswordSchema),
  });

  const onSubmit = async (data: ResetPasswordInput) => {
    setIsLoading(true);
    setError("");
    try {
      // In real app, token comes from URL query state
      await authApi.resetPassword({ token: "temp-token", password: data.password });
      setIsSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reset password");
    } finally {
      setIsLoading(false);
    }
  };

  if (isSuccess) {
    return (
      <div className="w-full space-y-6">
        <div className="flex justify-center">
          <div className="w-12 h-12 rounded-full bg-primary-container flex items-center justify-center">
            <MaterialIcon name="check_circle" size="lg" className="text-on-primary-container" />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <h2 className="text-headline-sm font-bold text-on-surface">Password Reset</h2>
          <p className="text-body-md text-on-surface-variant">
            Your password has been successfully reset. You can now log in with your new password.
          </p>
        </div>
        <div className="flex justify-center pt-2">
          <Link href="/login">
            <Button className="bg-primary text-on-primary hover:bg-primary/90">
              Continue to Login
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <h2 className="text-headline-sm font-bold text-on-surface">Reset Password</h2>
        <p className="text-body-md text-on-surface-variant">
          Choose a new password for your account.
        </p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && <ErrorDisplay message={error} />}
        <div className="space-y-2">
          <Label htmlFor="password" className="text-on-surface">New Password</Label>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              placeholder="Enter new password"
              className="bg-surface-container-high border-outline-variant text-on-surface pr-10"
              {...register("password")}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-on-surface-variant hover:text-on-surface"
              tabIndex={-1}
            >
              <MaterialIcon name={showPassword ? "visibility_off" : "visibility"} size="sm" />
            </button>
          </div>
          {errors.password && (
            <p className="text-sm text-error">{errors.password.message}</p>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="confirmPassword" className="text-on-surface">Confirm Password</Label>
          <Input
            id="confirmPassword"
            type={showPassword ? "text" : "password"}
            placeholder="Confirm new password"
            className="bg-surface-container-high border-outline-variant text-on-surface"
            {...register("confirmPassword")}
          />
          {errors.confirmPassword && (
            <p className="text-sm text-error">{errors.confirmPassword.message}</p>
          )}
        </div>
        <Button
          type="submit"
          className="w-full bg-primary text-on-primary hover:bg-primary/90"
          disabled={isLoading}
        >
          {isLoading ? <LoadingSpinner /> : "Reset Password"}
        </Button>
      </form>
    </div>
  );
}