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
import { Copy, RefreshCw, CheckCircle2, ShieldCheck } from "lucide-react";

interface TwoFASetupProps {
  onSuccess?: () => void;
}

export function TwoFASetup({ onSuccess }: TwoFASetupProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [setupData, setSetupData] = useState<{
    secret: string;
    uri: string;
    qrCode: string; // base64 PNG data URL
    backupCodes: string[];
  } | null>(null);
  const [isVerified, setIsVerified] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TwoFactorCodeInput>({
    resolver: zodResolver(twoFactorCodeSchema),
  });

  const handleSetup = async () => {
    setIsLoading(true);
    setError("");
    try {
      const data = await authApi.setup2FA();
      const qrSrc = data.qrCode.startsWith("data:image")
        ? data.qrCode
        : `data:image/png;base64,${data.qrCode}`;
      setSetupData({
        secret: data.secret,
        uri: data.uri,
        qrCode: qrSrc,
        backupCodes: data.backupCodes,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to setup 2FA");
    } finally {
      setIsLoading(false);
    }
  };

  const onVerify = async (data: TwoFactorCodeInput) => {
    setIsLoading(true);
    setError("");
    try {
      await authApi.verify2FASetup(data.code);
      setIsVerified(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid verification code");
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (isVerified) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>2FA Enabled!</CardTitle>
          <CardDescription>
            Two-factor authentication has been successfully enabled on your account.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-3 p-4 bg-green-50 border border-green-200 rounded-lg">
            <CheckCircle2 className="h-8 w-8 text-green-600 shrink-0" />
            <div>
              <p className="font-medium text-green-800">Two-factor authentication is active</p>
              <p className="text-sm text-green-700 mt-0.5">
                You&apos;ll need your authenticator app the next time you log in.
              </p>
            </div>
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium">Save these backup codes in a secure location:</p>
            <p className="text-xs text-muted-foreground">
              Each code can only be used once. Store them somewhere safe.
            </p>
            <div className="grid grid-cols-2 gap-2">
              {setupData?.backupCodes.map((code, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-2 bg-muted rounded text-sm font-mono"
                >
                  <span>{code}</span>
                  <Button size="sm" variant="ghost" onClick={() => copyToClipboard(code)}>
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
          <Button onClick={onSuccess} className="w-full">
            I have saved the backup codes
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!setupData) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Setup Two-Factor Authentication</CardTitle>
          <CardDescription>
            Add an extra layer of security to your account
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && <ErrorDisplay message={error} />}
          <div className="space-y-3">
            <div className="flex items-start gap-3 p-4 bg-blue-50 border border-blue-100 rounded-lg">
              <ShieldCheck className="h-5 w-5 text-blue-600 mt-0.5 shrink-0" />
              <div className="text-sm text-blue-900 space-y-1.5">
                <p className="font-medium">Why enable 2FA?</p>
                <ul className="list-disc list-inside space-y-0.5 text-blue-800">
                  <li>Protects against password theft</li>
                  <li>Required for SSH key management</li>
                  <li>One-time codes from authenticator apps</li>
                </ul>
              </div>
            </div>
            <p className="text-sm text-muted-foreground">
              You&apos;ll need an authenticator app like Google Authenticator, Authy, or 1Password.
            </p>
          </div>
        </CardContent>
        <CardFooter>
          <Button onClick={handleSetup} disabled={isLoading} className="w-full">
            {isLoading ? <LoadingSpinner size="sm" /> : "Continue to Setup"}
          </Button>
        </CardFooter>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Scan QR Code</CardTitle>
        <CardDescription>
          Scan this QR code with your authenticator app, or enter the key manually
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && <ErrorDisplay message={error} />}
        <div className="flex flex-col items-center space-y-4">
          <div className="p-4 bg-white rounded-lg border shadow-sm">
            {setupData.qrCode ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={setupData.qrCode}
                alt="TOTP QR Code"
                width={180}
                height={180}
                className="w-[180px] h-[180px]"
              />
            ) : (
              <div className="w-[180px] h-[180px] flex items-center justify-center bg-gray-100 text-xs text-gray-500 text-center p-4">
                QR Code unavailable
                <br />
                Use manual entry below
              </div>
            )}
          </div>

          <div className="w-full space-y-2">
            <Label className="text-center block">Or enter this key manually:</Label>
            <div className="flex items-center gap-2 p-2 bg-muted rounded">
              <code className="flex-1 text-xs font-mono break-all">{setupData.secret}</code>
              <Button size="sm" variant="ghost" onClick={() => copyToClipboard(setupData.secret)} title="Copy secret">
                <Copy className="h-3 w-3" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground text-center break-all">{setupData.uri}</p>
          </div>

          <form onSubmit={handleSubmit(onVerify)} className="w-full space-y-3">
            <div className="space-y-2">
              <Label htmlFor="code">Enter 6-digit verification code</Label>
              <Input
                id="code"
                placeholder="000000"
                maxLength={6}
                {...register("code")}
                className="text-center text-2xl tracking-widest font-mono"
              />
              {errors.code && (
                <p className="text-sm text-destructive">{errors.code.message}</p>
              )}
            </div>
            <Button type="submit" disabled={isLoading} className="w-full">
              {isLoading ? <LoadingSpinner size="sm" /> : "Enable 2FA"}
            </Button>
          </form>
        </div>
      </CardContent>
      <CardFooter className="flex justify-center">
        <Button variant="ghost" onClick={handleSetup} disabled={isLoading} className="text-muted-foreground">
          <RefreshCw className="h-4 w-4 mr-1" />
          Generate New QR Code
        </Button>
      </CardFooter>
    </Card>
  );
}
