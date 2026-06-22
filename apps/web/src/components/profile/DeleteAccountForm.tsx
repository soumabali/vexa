"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { LoadingSpinner } from "../auth/LoadingSpinner";
import { ErrorDisplay } from "../auth/ErrorDisplay";
import { AlertTriangle } from "lucide-react";

const deleteAccountSchema = z.object({
  password: z.string().min(1, "Password is required"),
  confirmDelete: z.enum(["DELETE"]),
});

type DeleteAccountInput = z.output<typeof deleteAccountSchema>;

export function DeleteAccountForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DeleteAccountInput>({
    resolver: zodResolver(deleteAccountSchema),
  });

  const onSubmit = async (data: DeleteAccountInput) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.deleteAccount(data.password);
      setIsSuccess(true);
      reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete account");
    } finally {
      setIsLoading(false);
    }
  };

  if (isSuccess) {
    return (
      <Card className="w-full max-w-md border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Account Deleted</CardTitle>
          <CardDescription>
            Your account has been permanently deleted.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md border-destructive">
      <CardHeader>
        <CardTitle className="text-destructive">Delete Account</CardTitle>
        <CardDescription>
          This action cannot be undone. All your data will be permanently removed.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              Deleting your account will permanently remove all your SSH keys, connections, and settings.
            </AlertDescription>
          </Alert>
          {error && <ErrorDisplay message={error} />}
          <div className="space-y-2">
            <Label htmlFor="password">Current Password</Label>
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
            <Label htmlFor="confirmDelete">Type "DELETE" to confirm</Label>
            <Input
              id="confirmDelete"
              placeholder="DELETE"
              {...register("confirmDelete")}
            />
            {errors.confirmDelete && (
              <p className="text-sm text-destructive">{errors.confirmDelete.message}</p>
            )}
          </div>
        </CardContent>
        <CardFooter>
          <Button type="submit" variant="destructive" disabled={isLoading}>
            {isLoading ? <LoadingSpinner size="sm" /> : "Delete Account"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}
