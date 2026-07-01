"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { forgotPasswordSchema, ForgotPasswordInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MaterialIcon } from "@/components/ui/material-icon";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import Link from "next/link";

export function ForgotPasswordForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordInput>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  const onSubmit = async (data: ForgotPasswordInput) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.forgotPassword(data.email);
      setIsSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send reset email");
    } finally {
      setIsLoading(false);
    }
  };

  if (isSuccess) {
    return (
      <div className="w-full space-y-6">
        <div className="flex justify-center">
          <div className="w-12 h-12 rounded-full bg-primary-container flex items-center justify-center">
            <MaterialIcon name="mark_email_read" size="lg" className="text-on-primary-container" />
          </div>
        </div>
        <div className="space-y-2 text-center">
          <h2 className="text-headline-sm font-bold text-on-surface">Check Your Email</h2>
          <p className="text-body-md text-on-surface-variant">
            If an account exists with that email, we&apos;ve sent a verification code to reset your password.
          </p>
        </div>
        <div className="text-center space-y-2">
          <p className="text-sm text-on-surface-variant">
            Didn&apos;t receive an email? Check your spam folder or{" "}
            <button
              onClick={() => setIsSuccess(false)}
              className="text-primary underline"
            >
              try again
            </button>
          </p>
        </div>
        <div className="flex justify-center pt-2">
          <Link href="/reset-password">
            <Button variant="outline">Enter Verification Code</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <h2 className="text-headline-sm font-bold text-on-surface">Forgot Password?</h2>
        <p className="text-body-md text-on-surface-variant">
          Enter your email and we&apos;ll send you a reset link.
        </p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && <ErrorDisplay message={error} />}
        <div className="space-y-2">
          <Label htmlFor="email" className="text-on-surface">Email</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            className="bg-surface-container-high border-outline-variant text-on-surface"
            {...register("email")}
          />
          {errors.email && (
            <p className="text-sm text-error">{errors.email.message}</p>
          )}
        </div>
        <Button
          type="submit"
          className="w-full bg-primary text-on-primary hover:bg-primary/90"
          disabled={isLoading}
        >
          {isLoading ? <LoadingSpinner /> : "Send Reset Link"}
        </Button>
      </form>
      <div className="text-center space-y-2 pt-2">
        <p className="text-sm text-on-surface-variant">
          Remember your password?{" "}
          <Link href="/login" className="text-primary hover:underline font-medium">
            Back to login
          </Link>
        </p>
      </div>
    </div>
  );
}