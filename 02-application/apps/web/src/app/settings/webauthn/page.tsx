"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import {
  ShieldCheck,
  Fingerprint,
  KeyRound,
  Trash2,
  Smartphone,
  Usb,
  AlertTriangle,
  CheckCircle2,
  Loader2,
} from "lucide-react";
import {
  isWebAuthnSupported,
  isPlatformAuthenticatorAvailable,
  registerCredential,
  listCredentials,
  deleteCredential,
  renameCredential,
  type WebAuthnCredential,
} from "@/lib/webauthn";
import { toast } from "sonner";

export default function WebAuthnSettingsPage() {
  const [supported, setSupported] = useState(false);
  const [platformAvailable, setPlatformAvailable] = useState(false);
  const [credentials, setCredentials] = useState<WebAuthnCredential[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [registerName, setRegisterName] = useState("");
  const [requirePlatform, setRequirePlatform] = useState(false);
  const [renameId, setRenameId] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [deleteId, setDeleteId] = useState<string | null>(null);

  useEffect(() => {
    setSupported(isWebAuthnSupported());
    isPlatformAuthenticatorAvailable().then(setPlatformAvailable);
    fetchCredentials();
  }, []);

  async function fetchCredentials() {
    try {
      const data = await listCredentials();
      setCredentials(data.credentials || []);
    } catch (e) { const err = e as { message?: string };
      setError(err.message || "Failed to load credentials");
    }
  }

  async function handleRegister() {
    setLoading(true);
    setError(null);
    try {
      await registerCredential(registerName || undefined, requirePlatform);
      setRegisterName("");
      setRequirePlatform(false);
      await fetchCredentials();
      toast.success("Passkey registered");
    } catch (e) { const err = e as { message?: string };
      setError(err.message || "Registration failed");
      toast.error(err.message || "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: string) {
    setLoading(true);
    setError(null);
    try {
      await deleteCredential(id);
      setDeleteId(null);
      await fetchCredentials();
      toast.success("Passkey removed");
    } catch (e) { const err = e as { message?: string };
      setError(err.message || "Delete failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleRename(id: string, name: string) {
    setLoading(true);
    setError(null);
    try {
      await renameCredential(id, name);
      setRenameId(null);
      setRenameValue("");
      await fetchCredentials();
      toast.success("Passkey renamed");
    } catch (e) { const err = e as { message?: string };
      setError(err.message || "Rename failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto py-6 max-w-2xl space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Passkeys &amp; WebAuthn</h1>
            <p className="text-muted-foreground">Manage passwordless authentication credentials.</p>
          </div>
          <ShieldCheck className="h-8 w-8 text-primary" />
        </div>

        {!supported && (
          <Card className="border-destructive">
            <CardContent className="flex items-center gap-3 py-6">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              <p className="text-sm text-destructive">Your browser does not support WebAuthn / FIDO2.</p>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Fingerprint className="h-5 w-5" />
              Register New Passkey
            </CardTitle>
            <CardDescription>
              Add a platform authenticator or a roaming authenticator (YubiKey, security key).
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="passkey-name">Passkey Name</Label>
              <Input
                id="passkey-name"
                placeholder="e.g. MacBook Touch ID"
                value={registerName}
                onChange={(e) => setRegisterName(e.target.value)}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                id="require-platform"
                type="checkbox"
                className="h-4 w-4 rounded border-gray-300"
                checked={requirePlatform}
                onChange={(e) => setRequirePlatform(e.target.checked)}
              />
              <Label htmlFor="require-platform" className="text-sm font-normal">
                Require platform authenticator (Touch ID / Windows Hello)
              </Label>
            </div>

            <Button onClick={handleRegister} disabled={!supported || loading} className="w-full sm:w-auto">
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <KeyRound className="mr-2 h-4 w-4" />
              Register Passkey
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Usb className="h-5 w-5" />
              Your Passkeys
            </CardTitle>
            <CardDescription>{credentials.length} passkey{credentials.length !== 1 && "s"} registered.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {credentials.length === 0 && <p className="text-sm text-muted-foreground">No passkeys registered yet.</p>}

            {credentials.map((cred) => (
              <div key={cred.id} className="flex items-center justify-between rounded-lg border p-4">
                <div className="flex items-center gap-3">
                  {cred.authenticatorAttachment === "platform" ? (
                    <Smartphone className="h-5 w-5 text-muted-foreground" />
                  ) : (
                    <Usb className="h-5 w-5 text-muted-foreground" />
                  )}
                  <div>
                    <p className="font-medium">{cred.name || "Unnamed Passkey"}</p>
                    <div className="flex flex-wrap items-center gap-2 mt-1">
                      {cred.isResidentKey && <Badge variant="secondary" className="text-xs">Resident Key</Badge>}
                      {cred.isBackupEligible && <Badge variant="outline" className="text-xs">Backup Eligible</Badge>}
                      {cred.isBackedUp && <Badge variant="outline" className="text-xs">Backed Up</Badge>}
                      <span className="text-xs text-muted-foreground">{new Date(cred.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {renameId === cred.id ? (
                    <div className="flex items-center gap-2">
                      <Input
                        className="w-40"
                        value={renameValue}
                        onChange={(e) => setRenameValue(e.target.value)}
                        placeholder="New name"
                      />
                      <Button size="sm" variant="ghost" onClick={() => handleRename(cred.id, renameValue)} disabled={loading}>
                        <CheckCircle2 className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setRenameId(null);
                          setRenameValue("");
                        }}
                      >
                        Cancel
                      </Button>
                    </div>
                  ) : (
                    <>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setRenameId(cred.id);
                          setRenameValue(cred.name || "");
                        }}
                      >
                        Rename
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-destructive"
                        onClick={() => {
                          setDeleteId(cred.id);
                          handleDelete(cred.id);
                        }}
                        disabled={loading}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {error && (
          <Card className="border-destructive">
            <CardContent className="flex items-center gap-3 py-4">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              <p className="text-sm text-destructive">{error}</p>
            </CardContent>
          </Card>
        )}

        <div className="rounded-lg border bg-muted/50 p-4">
          <h3 className="text-sm font-semibold mb-1">Security Tips</h3>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Register at least one backup passkey (e.g. a YubiKey) in case your platform authenticator is unavailable.</li>
            <li>Passkeys with resident keys allow discoverable credential login without entering your email.</li>
            <li>Backup eligible credentials can be synced across devices by your authenticator vendor.</li>
          </ul>
        </div>
      </div>
    </DashboardLayout>
  );
}
