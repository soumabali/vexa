"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import {
  Shield,
  Key,
  Fingerprint,
  Trash2,
  Plus,
  AlertTriangle,
  Loader2,
  Smartphone,
} from "lucide-react";
import { useEffect, useState } from "react";
import {
  WebAuthnCredential,
  isWebAuthnSupported,
  listCredentials,
  deleteCredential,
} from "@/lib/webauthn";
import { toast } from "sonner";
import Link from "next/link";

export default function SecuritySettingsPage() {
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false);
  const [webAuthnCreds, setWebAuthnCreds] = useState<WebAuthnCredential[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    async function load() {
      try {
        const data = await listCredentials();
        if (mounted) setWebAuthnCreds(data.credentials || []);
      } catch {
        // WebAuthn may be unavailable; don't block page render
      } finally {
        if (mounted) setLoading(false);
      }
    }
    load();
    return () => {
      mounted = false;
    };
  }, []);

  const handleRemoveWebAuthn = async (id: string) => {
    try {
      await deleteCredential(id);
      setWebAuthnCreds((prev) => prev.filter((c) => c.id !== id));
      toast.success("Passkey removed");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to remove passkey");
    }
  };

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 max-w-4xl space-y-6">
        <div>
          <h1 className="text-2xl font-bold">Security Settings</h1>
          <p className="text-muted-foreground">Manage authentication methods and account security</p>
        </div>

        {/* 2FA TOTP */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              <CardTitle>Two-Factor Authentication (TOTP)</CardTitle>
            </div>
            <CardDescription>
              Add an extra layer of security with an authenticator app
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Require 2FA at login</Label>
                <p className="text-sm text-muted-foreground">Prompt for a verification code when signing in</p>
              </div>
              <Switch checked={twoFactorEnabled} onCheckedChange={setTwoFactorEnabled} />
            </div>

            {twoFactorEnabled && (
              <div className="rounded-md border p-4 space-y-3">
                <div className="flex items-center gap-3">
                  <Smartphone className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1">
                    <p className="font-medium">Authenticator app</p>
                    <p className="text-sm text-muted-foreground">
                      Google Authenticator, Authy, Microsoft Authenticator, etc.
                    </p>
                  </div>
                  <Badge variant="outline">Setup required</Badge>
                </div>
                <Button size="sm" disabled>Set up authenticator</Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* WebAuthn */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Fingerprint className="h-5 w-5" />
              <CardTitle>Passkeys / WebAuthn</CardTitle>
            </div>
            <CardDescription>Use biometrics or security keys for passwordless sign-in</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading ? (
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading passkeys...
              </div>
            ) : webAuthnCreds.length === 0 ? (
              <p className="text-sm text-muted-foreground">No passkeys registered yet.</p>
            ) : (
              webAuthnCreds.map((cred) => (
                <div
                  key={cred.id}
                  className="flex items-center justify-between rounded-lg border p-4"
                >
                  <div className="flex items-center gap-3">
                    <Fingerprint className="h-5 w-5 text-primary" />
                    <div>
                      <p className="font-medium">{cred.name || "Unnamed passkey"}</p>
                      <p className="text-xs text-muted-foreground">
                        {cred.authenticatorAttachment === "platform" ? "Platform" : "Roaming"} authenticator
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveWebAuthn(cred.id)}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              ))
            )}

            <Button asChild>
              <Link href="/settings/webauthn">
                <Plus className="h-4 w-4 mr-2" />
                Manage passkeys
              </Link>
            </Button>
          </CardContent>
        </Card>

        {/* Danger zone */}
        <Card className="border-destructive">
          <CardHeader>
            <div className="flex items-center gap-2 text-destructive">
              <AlertTriangle className="h-5 w-5" />
              <CardTitle className="text-destructive">Danger Zone</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Delete account</p>
                <p className="text-sm text-muted-foreground">Permanently delete your account and all data</p>
              </div>
              <Button variant="destructive" disabled>Delete Account</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
