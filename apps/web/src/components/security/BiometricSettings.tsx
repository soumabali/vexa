"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { twoFactorCodeSchema, TwoFactorCodeInput } from "@/lib/validations/auth";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoadingSpinner } from "../auth/LoadingSpinner";
import { ErrorDisplay } from "../auth/ErrorDisplay";
import { Switch } from "@/components/ui/switch";

export function BiometricSettings() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isEnabled, setIsEnabled] = useState(false);
  const [showSetup, setShowSetup] = useState(false);

  const handleToggle = async (enabled: boolean) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.updateBiometric(enabled);
      setIsEnabled(enabled);
      if (enabled) {
        setShowSetup(true);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update biometric settings");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Biometric Authentication</CardTitle>
        <CardDescription>
          Use Face ID or Touch ID for quick and secure access
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && <ErrorDisplay message={error} />}
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label>Enable Biometric Login</Label>
            <p className="text-sm text-muted-foreground">
              Use Face ID or Touch ID to log in
            </p>
          </div>
          <Switch
            checked={isEnabled}
            onCheckedChange={handleToggle}
            disabled={isLoading}
          />
        </div>

        {showSetup && isEnabled && (
          <div className="p-4 bg-muted rounded-lg space-y-4">
            <p className="text-sm font-medium">Setup Instructions:</p>
            <ul className="text-sm text-muted-foreground space-y-2 list-disc list-inside">
              <li>Go to your device settings</li>
              <li>Enable Face ID / Touch ID</li>
              <li>Return to this app and scan your biometric</li>
            </ul>
            <Button
              className="w-full"
              onClick={() => {
                // In real app, trigger biometric prompt
                alert("Biometric prompt would appear here");
                setShowSetup(false);
              }}
            >
              Setup Biometric
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
