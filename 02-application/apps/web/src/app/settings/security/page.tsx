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
import { Input } from "@/components/ui/input";
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
import { MaterialIcon } from "@/components/ui/material-icon";
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
  const [disableTotpCode, setDisableTotpCode] = useState("");
  const [codesOpen, setCodesOpen] = useState(false);
  const [codesTotp, setCodesTotp] = useState("");
  const [codesLoading, setCodesLoading] = useState(false);
  const [codes, setCodes] = useState<string[] | null>(null);

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
      await authApi.disable2FA(disableTotpCode);
      toast.success("MFA disabled");
      setDisableTotpCode("");
      await refreshProfile();
      setDisableOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to disable MFA");
    } finally {
      setDisableLoading(false);
    }
  };

  const handleRegenerateCodes = async () => {
    setCodesLoading(true);
    try {
      const res = await authApi.regenerateBackupCodes(codesTotp);
      setCodes(res.backup_codes);
      toast.success("Backup codes generated");
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to regenerate backup codes"
      );
    } finally {
      setCodesLoading(false);
    }
  };

  const handleCopyCodes = async () => {
    if (!codes) return;
    try {
      await navigator.clipboard.writeText(codes.join("\n"));
      toast.success("Codes copied to clipboard");
    } catch {
      toast.error("Could not copy to clipboard");
    }
  };

  const handleDownloadCodes = () => {
    if (!codes) return;
    const blob = new Blob([codes.join("\n")], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "vexa-backup-codes.txt";
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };

  const handleCloseCodes = () => {
    // One-time-show: drop codes from state on close so they cannot leak back.
    setCodes(null);
    setCodesTotp("");
    setCodesOpen(false);
  };

  const mfaEnabled = profile?.mfa_enabled ?? false;

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 max-w-4xl space-y-6">
        <div>
          <h1 className="text-headline-lg text-on-surface mb-2">Security Settings</h1>
          <p className="text-body-lg text-on-surface-variant">
            Manage authentication methods and account security
          </p>
        </div>

        {/* MFA card */}
        <Card className="bg-surface-container border border-outline-variant rounded-xl">
          <CardHeader>
            <div className="flex items-center gap-2">
              <MaterialIcon name="shield" className="text-primary" />
              <CardTitle className="text-on-surface">Multi-Factor Authentication</CardTitle>
            </div>
            <CardDescription className="text-on-surface-variant">
              Add an extra layer of security with an authenticator app
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-on-surface">Require 2FA at login</Label>
                <p className="text-sm text-on-surface-variant">Prompt for a verification code when signing in</p>
              </div>
              {profileLoading ? (
                <MaterialIcon name="progress_activity" className="animate-spin text-on-surface-variant" />
              ) : (
                <Switch
                  checked={mfaEnabled}
                  disabled
                  aria-label="Two-factor authentication status"
                />
              )}
            </div>

            {/* Methods list */}
            <div className="space-y-3">
              <div className="flex items-center gap-3 p-4 rounded-lg border border-outline-variant bg-surface-container-lowest">
                <MaterialIcon name="smartphone" className={mfaEnabled ? "text-success" : "text-on-surface-variant"} />
                <div className="flex-1">
                  <p className="font-medium text-on-surface">Authenticator app</p>
                  <p className="text-sm text-on-surface-variant">
                    {mfaEnabled ? "Your account is protected with TOTP at sign-in." : "Google Authenticator, Authy, Microsoft Authenticator, etc."}
                  </p>
                </div>
                {mfaEnabled ? (
                  <Badge className="bg-success/10 text-success border border-success/30">Enabled</Badge>
                ) : (
                  <Badge variant="outline" className="border-outline-variant text-on-surface-variant">Setup required</Badge>
                )}
                {mfaEnabled ? (
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setDisableOpen(true)}
                      className="border-outline-variant text-on-surface"
                    >
                      <MaterialIcon name="key" size="sm" className="mr-2" />
                      Disable
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setCodesOpen(true)}
                      className="border-outline-variant text-on-surface"
                      data-testid="open-backup-codes"
                    >
                      <MaterialIcon name="vpn_key" size="sm" className="mr-2" />
                      Backup codes
                    </Button>
                  </div>
                ) : (
                  <Button size="sm" onClick={() => setSetupOpen(true)} className="bg-primary text-on-primary hover:bg-primary/90">
                    Configure
                  </Button>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* WebAuthn / Passkeys */}
        <Card className="bg-surface-container border border-outline-variant rounded-xl">
          <CardHeader>
            <div className="flex items-center gap-2">
              <MaterialIcon name="fingerprint" className="text-primary" />
              <CardTitle className="text-on-surface">Passkeys / WebAuthn</CardTitle>
            </div>
            <CardDescription className="text-on-surface-variant">
              Use biometrics or security keys for passwordless sign-in
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {webAuthnLoading ? (
              <div className="flex items-center gap-2 text-on-surface-variant">
                <MaterialIcon name="progress_activity" className="animate-spin" />
                Loading passkeys...
              </div>
            ) : webAuthnCreds.length === 0 ? (
              <p className="text-sm text-on-surface-variant">No passkeys registered yet.</p>
            ) : (
              webAuthnCreds.map((cred) => (
                <div
                  key={cred.id}
                  className="flex items-center justify-between rounded-lg border border-outline-variant bg-surface-container-lowest p-4"
                >
                  <div className="flex items-center gap-3">
                    <MaterialIcon name="fingerprint" className="text-primary" />
                    <div>
                      <p className="font-medium text-on-surface">{cred.name || "Unnamed passkey"}</p>
                      <p className="text-xs text-on-surface-variant">
                        {cred.authenticatorAttachment === "platform" ? "Platform" : "Roaming"} authenticator
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveWebAuthn(cred.id)}
                    className="text-error hover:text-error"
                    aria-label="Remove passkey"
                  >
                    <MaterialIcon name="delete" size="sm" />
                  </Button>
                </div>
              ))
            )}

            <Button asChild className="bg-primary text-on-primary hover:bg-primary/90">
              <Link href="/settings/webauthn">
                <MaterialIcon name="add" size="sm" className="mr-2" />
                Manage passkeys
              </Link>
            </Button>
          </CardContent>
        </Card>

        {/* Security audit card */}
        <Card className="bg-surface-container border border-outline-variant rounded-xl">
          <CardHeader>
            <div className="flex items-center gap-2">
              <MaterialIcon name="verified_user" className="text-primary" />
              <CardTitle className="text-on-surface">Security Audit</CardTitle>
            </div>
            <CardDescription className="text-on-surface-variant">
              Checklist of recommended security practices
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <AuditRow ok={mfaEnabled} label="Two-factor authentication enabled" />
            <AuditRow ok={webAuthnCreds.length > 0} label="Passkey registered" />
            <AuditRow ok={!!profile?.email} label="Email verified" />
            <AuditRow ok={false} label="Backup codes saved" warn />
          </CardContent>
        </Card>

        {/* Danger zone */}
        <Card className="bg-surface-container border border-error rounded-xl">
          <CardHeader>
            <div className="flex items-center gap-2 text-error">
              <MaterialIcon name="warning" className="text-error" />
              <CardTitle className="text-error">Danger Zone</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium text-on-surface">Delete account</p>
                <p className="text-sm text-on-surface-variant">Permanently delete your account and all data</p>
              </div>
              <Button variant="destructive" disabled className="bg-error text-on-error hover:bg-error/90">
                Delete Account
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* MFA setup dialog */}
      <Dialog open={setupOpen} onOpenChange={setSetupOpen}>
        <DialogContent className="sm:max-w-lg bg-surface-container-high border-outline-variant">
          <DialogHeader>
            <DialogTitle className="text-on-surface">Set up TOTP</DialogTitle>
            <DialogDescription className="text-on-surface-variant">
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
        <DialogContent className="sm:max-w-md bg-surface-container-high border-outline-variant">
          <DialogHeader>
            <DialogTitle className="text-on-surface">Disable MFA</DialogTitle>
            <DialogDescription className="text-on-surface-variant">
              Enter a current TOTP code to confirm disabling two-factor authentication.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="disable-totp-code" className="text-on-surface">Verification code</Label>
            <Input
              id="disable-totp-code"
              name="totp_code"
              inputMode="numeric"
              maxLength={6}
              autoComplete="one-time-code"
              placeholder="000000"
              value={disableTotpCode}
              onChange={(e) =>
                setDisableTotpCode(e.target.value.replace(/\D/g, "").slice(0, 6))
              }
              disabled={disableLoading}
              autoFocus
              className="text-center text-2xl tracking-widest font-mono bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDisableOpen(false)} disabled={disableLoading} className="border-outline-variant text-on-surface">
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDisableMFA}
              disabled={disableLoading || disableTotpCode.length !== 6}
              className="bg-error text-on-error hover:bg-error/90"
            >
              {disableLoading ? <MaterialIcon name="progress_activity" size="sm" className="animate-spin mr-2" /> : null}
              Disable MFA
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Request backup codes (one-time-show) */}
      <Dialog
        open={codesOpen}
        onOpenChange={(open) => {
          if (open) return;
          if (codesLoading) return;
          if (codes !== null) {
            // Once codes are shown the user must press the explicit
            // "I have saved them" button — no other close path.
            return;
          }
          setCodesOpen(false);
          setCodesTotp("");
        }}
      >
        <DialogContent className="bg-surface-container-high border-outline-variant">
          {codes === null ? (
            <>
              <DialogHeader>
                <DialogTitle className="text-on-surface">Regenerate backup codes</DialogTitle>
                <DialogDescription className="text-on-surface-variant">
                  These codes will replace any existing codes. Save them now —
                  they will only be shown once.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Label htmlFor="backup-codes-totp" className="text-on-surface">Verification code</Label>
                <Input
                  id="backup-codes-totp"
                  inputMode="numeric"
                  pattern="[0-9]{6}"
                  maxLength={6}
                  autoFocus
                  value={codesTotp}
                  onChange={(e) =>
                    setCodesTotp(e.target.value.replace(/[^0-9]/g, ""))
                  }
                  disabled={codesLoading}
                  data-testid="backup-codes-totp"
                  className="bg-surface-container-lowest border-outline-variant text-on-surface focus:border-primary focus:ring-1 focus:ring-primary"
                />
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setCodesOpen(false)}
                  disabled={codesLoading}
                  className="border-outline-variant text-on-surface"
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleRegenerateCodes}
                  disabled={codesLoading || codesTotp.length !== 6}
                  className="bg-primary text-on-primary hover:bg-primary/90"
                  data-testid="submit-backup-codes"
                >
                  {codesLoading ? (
                    <MaterialIcon name="progress_activity" size="sm" className="animate-spin mr-2" />
                  ) : null}
                  Generate codes
                </Button>
              </DialogFooter>
            </>
          ) : (
            <>
              <DialogHeader>
                <DialogTitle className="text-on-surface">Save your new backup codes</DialogTitle>
                <DialogDescription className="text-on-surface-variant">
                  These codes will only be shown once. Store them in a safe
                  place — they are the only way to recover access if you lose
                  your authenticator.
                </DialogDescription>
              </DialogHeader>
              <div
                className="rounded-md border border-outline-variant bg-surface-container-lowest p-3 font-mono text-sm grid grid-cols-2 gap-x-4 gap-y-1 text-on-surface"
                data-testid="backup-codes-list"
              >
                {codes.map((code) => (
                  <span key={code}>{code}</span>
                ))}
              </div>
              <DialogFooter className="gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleCopyCodes}
                  data-testid="copy-backup-codes"
                  className="border-outline-variant text-on-surface"
                >
                  <MaterialIcon name="content_copy" size="sm" className="mr-2" />
                  Copy
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDownloadCodes}
                  data-testid="download-backup-codes"
                  className="border-outline-variant text-on-surface"
                >
                  <MaterialIcon name="download" size="sm" className="mr-2" />
                  Download .txt
                </Button>
                <Button
                  onClick={handleCloseCodes}
                  className="bg-primary text-on-primary hover:bg-primary/90"
                  data-testid="confirm-backup-codes"
                >
                  I have saved them
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}

function AuditRow({ ok, label, warn }: { ok: boolean; label: string; warn?: boolean }) {
  return (
    <div className="flex items-center gap-3">
      {ok ? (
        <MaterialIcon name="check_circle" className="text-success" />
      ) : warn ? (
        <MaterialIcon name="warning" className="text-warning" />
      ) : (
        <MaterialIcon name="error" className="text-error" />
      )}
      <span className={ok ? "text-on-surface" : "text-on-surface-variant"}>{label}</span>
    </div>
  );
}