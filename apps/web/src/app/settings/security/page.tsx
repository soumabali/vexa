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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { authApi } from "@/lib/api/auth";
import { TwoFASetup } from "@/components/security/TwoFASetup";
import { toast } from "sonner";
import Link from "next/link";

interface UserProfile {
  id: string;
  email: string;
  role: string;
  mfa_enabled: boolean;
  totp_enabled: boolean;
}

export default function SecuritySettingsPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [webAuthnCreds, setWebAuthnCreds] = useState<WebAuthnCredential[]>([]);
  const [webAuthnLoading, setWebAuthnLoading] = useState(true);
  const [setupOpen, setSetupOpen] = useState(false);
  const [disableOpen, setDisableOpen] = useState(false);
  const [disableLoading, setDisableLoading] = useState(false);

  useEffect(() => {
    let mounted = true;
    async function load() {
      try {
        const data = await authApi.getUserProfile();
        if (mounted) setProfile(data);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to load profile");
      } finally {
        if (mounted) setProfileLoading(false);
      }
    }
    load();
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    let mounted = true;
    async function loadCreds() {
      try {
        const supported = await isWebAuthnSupported();
        if (!supported) return;
        const data = await listCredentials();
        if (mounted) setWebAuthnCreds(data.credentials || []);
      } catch {
        // WebAuthn may be unavailable; don't block page render
      } finally {
        if (mounted) setWebAuthnLoading(false);
      }
    }
    loadCreds();
    return () => {
      mounted = false;
    };
  }, []);

  const refreshProfile = async () => {
    try {
      const data = await authApi.getUserProfile();
      setProfile(data);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to refresh profile");
    }
  };

  const handleRemoveWebAuthn = async (id: string) => {
    try {
      await deleteCredential(id);
      setWebAuthnCreds((prev) => prev.filter((c) => c.id !== id));
      toast.success("Passkey removed");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to remove passkey");
    }
  };

  const handleDisableMFA = async () => {
    setDisableLoading(true);
    try {
      await authApi.disable2FA();
      toast.success("MFA disabled");
      await refreshProfile();
      setDisableOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to disable MFA");
    } finally {
      setDisableLoading(false);
    }
  };

  const mfaEnabled = profile?.mfa_enabled ?? false;

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
              {profileLoading ? (
                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
              ) : (
                <Switch
                  checked={mfaEnabled}
                  disabled
                  aria-label="Two-factor authentication status"
                />
              )}
            </div>

            {mfaEnabled ? (
              <div className="rounded-md border p-4 space-y-3">
                <div className="flex items-center gap-3">
                  <Smartphone className="h-5 w-5 text-green-600" />
                  <div className="flex-1">
                    <p className="font-medium">Authenticator app is active</p>
                    <p className="text-sm text-muted-foreground">
                      Your account is protected with TOTP at sign-in.
                    </p>
                  </div>
                  <Badge variant="outline" className="text-green-600 border-green-200">Enabled</Badge>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setDisableOpen(true)}
                >
                  <Key className="h-4 w-4 mr-2" />
                  Disable MFA
                </Button>
              </div>
            ) : (
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
                <Button size="sm" onClick={() => setSetupOpen(true)}>Set up authenticator</Button>
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
            {webAuthnLoading ? (
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

      {/* MFA setup dialog */}
      <Dialog open={setupOpen} onOpenChange={setSetupOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Set up TOTP</DialogTitle>
            <DialogDescription>
              Scan the QR code and verify a code to enable two-factor authentication.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-center py-2">
            <TwoFASetup
              onSuccess={async () => {
                await refreshProfile();
                setSetupOpen(false);
              }}
            />
          </div>
        </DialogContent>
      </Dialog>

      {/* Disable MFA confirmation dialog */}
      <Dialog open={disableOpen} onOpenChange={setDisableOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Disable MFA</DialogTitle>
            <DialogDescription>
              This will remove two-factor authentication from your account and lower your security. Are you sure?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDisableOpen(false)} disabled={disableLoading}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDisableMFA} disabled={disableLoading}>
              {disableLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Disable MFA
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}
