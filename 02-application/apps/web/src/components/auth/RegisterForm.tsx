"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { registerSchema, RegisterInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MaterialIcon } from "@/components/ui/material-icon";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import Link from "next/link";

export function RegisterForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterInput) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.register({
        email: data.email,
        password: data.password,
      });
      setIsSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
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
            We&apos;ve sent a verification link to your email. Please check your inbox and click the link to verify your account.
          </p>
        </div>
        <div className="text-center space-y-2">
          <p className="text-sm text-on-surface-variant">
            Didn&apos;t receive the email? Check your spam folder or{" "}
            <button
              onClick={() => {}}
              className="text-primary underline"
            >
              resend
            </button>
          </p>
        </div>
        <div className="flex justify-center pt-2">
          <Link href="/login">
            <Button variant="outline">Back to Login</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <h2 className="text-headline-sm font-bold text-on-surface">Create an Account</h2>
        <p className="text-body-md text-on-surface-variant">
          Enter your details to create your vexa account
        </p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && <ErrorDisplay message={error} />}
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            {...register("email")}
          />
          {errors.email && (
            <p className="text-sm text-destructive">{errors.email.message}</p>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            placeholder="••••••••"
            {...register("password")}
          />
          {errors.password && (
            <p className="text-sm text-destructive">{errors.password.message}</p>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="confirmPassword">Confirm Password</Label>
          <Input
            id="confirmPassword"
            type="password"
            placeholder="••••••••"
            {...register("confirmPassword")}
          />
          {errors.confirmPassword && (
            <p className="text-sm text-destructive">{errors.confirmPassword.message}</p>
          )}
        </div>
        <div className="flex items-center space-x-2">
          <Checkbox id="acceptTerms" {...register("acceptTerms")} />
          <Label htmlFor="acceptTerms" className="text-sm">
            I agree to the{" "}
            <Link href="/terms" className="text-primary underline">
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link href="/privacy" className="text-primary underline">
              Privacy Policy
            </Link>
          </Label>
        </div>
        {errors.acceptTerms && (
          <p className="text-sm text-destructive">{errors.acceptTerms.message}</p>
        )}
        <div className="flex flex-col gap-4 pt-2">
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? (
              <LoadingSpinner size="sm" />
            ) : (
              "Create Account"
            )}
          </Button>
          <p className="text-sm text-on-surface-variant text-center">
            Already have an account?{" "}
            <Link href="/login" className="text-primary underline">Sign in</Link>
          </p>
        </div>
      </form>
    </div>
  );
}