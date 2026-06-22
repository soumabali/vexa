"use client";

import { useState } from "react";
import { useWebAuthn } from "@/hooks/useWebAuthn";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingSpinner } from "./LoadingSpinner";
import { ErrorDisplay } from "./ErrorDisplay";
import { Badge } from "@/components/ui/badge";
import { KeyRound, Shield, Trash2, Plus, CheckCircle2, XCircle } from "lucide-react";

interface WebAuthnDevice {
  id: string;
  name: string;
  credential_id: string;
  created_at: string;
  last_used: string | null;
}

interface WebAuthnSettingsProps {
  onDeviceRegistered?: () => void;
}

export function WebAuthnSettings({ onDeviceRegistered }: WebAuthnSettingsProps) {
  const { register, listDevices, removeDevice, status, isSupported } = useWebAuthn();
  const [devices, setDevices] = useState<WebAuthnDevice[]>([]);
  const [loading, setLoading] = useState(false);
  const [deviceError, setDeviceError] = useState("");
  const [removingId, setRemovingId] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);
  const [addingName, setAddingName] = useState("");
  const [registrationSuccess, setRegistrationSuccess] = useState(false);

  const loadDevices = async () => {
    setLoading(true);
    setDeviceError("");
    try {
      const devs = await listDevices();
      // Map API response to WebAuthnDevice — credential_id comes from the API device object
      const mapped: WebAuthnDevice[] = devs.map((d: { id?: string; device_id?: string; name: string; credential_id?: string; device_credential_id?: string; created_at: string; last_used?: string | null }) => ({
        id: (d.id ?? d.device_id) as string,
        name: d.name,
        credential_id: (d.credential_id ?? d.device_credential_id ?? "") as string,
        created_at: d.created_at,
        last_used: (d.last_used ?? null) as string | null,
      }));
      setDevices(mapped);
    } catch (err) {
      setDeviceError(err instanceof Error ? err.message : "Failed to load devices");
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async () => {
    if (!addingName.trim()) return;
    setDeviceError("");
    setRegistrationSuccess(false);
    try {
      const userId = "current-user"; // In real app, get from session context
      const credentialId = await register({
        userId,
        userName: addingName,
        userDisplayName: addingName,
      });
      if (credentialId) {
        setRegistrationSuccess(true);
        setAddingName("");
        setShowAddForm(false);
        await loadDevices();
        onDeviceRegistered?.();
      }
    } catch (err) {
      setDeviceError(err instanceof Error ? err.message : "Registration failed");
    }
  };

  const handleRemove = async (deviceId: string) => {
    setRemovingId(deviceId);
    setDeviceError("");
    try {
      await removeDevice(deviceId);
      setDevices(prev => prev.filter(d => d.id !== deviceId));
    } catch (err) {
      setDeviceError(err instanceof Error ? err.message : "Failed to remove device");
    } finally {
      setRemovingId(null);
    }
  };

  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });

  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Security Keys</CardTitle>
          <CardDescription>Manage WebAuthn (FIDO2) security keys and devices</CardDescription>
        </div>
        {!showAddForm && (
          <Button size="sm" onClick={() => { setShowAddForm(true); loadDevices(); }}>
            <Plus className="h-4 w-4 mr-1" />
            Add Key
          </Button>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        {deviceError && <ErrorDisplay message={deviceError} />}

        {registrationSuccess && (
          <div className="flex items-center gap-2 p-3 bg-green-50 border border-green-200 rounded-md text-green-800 text-sm">
            <CheckCircle2 className="h-4 w-4 shrink-0" />
            Security key registered successfully!
          </div>
        )}

        {!isSupported && (
          <div className="flex items-center gap-2 p-4 bg-yellow-50 border border-yellow-200 rounded-md">
            <XCircle className="h-4 w-4 text-yellow-600 shrink-0" />
            <p className="text-sm text-yellow-800">
              WebAuthn is not supported in this browser. Please use Chrome, Firefox, Safari, or Edge.
            </p>
          </div>
        )}

        {showAddForm && (
          <div className="space-y-3 p-4 border rounded-lg bg-muted/30">
            <p className="text-sm font-medium">Register a new security key:</p>
            <div className="space-y-2">
              <label className="text-sm" htmlFor="device-name">Key Name</label>
              <input
                id="device-name"
                type="text"
                placeholder="e.g. YubiKey 5, MacBook Touch ID"
                value={addingName}
                onChange={(e) => setAddingName(e.target.value)}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            </div>
            <div className="flex gap-2">
              <Button onClick={handleRegister} disabled={!addingName.trim() || status === "waiting"} className="flex-1">
                {status === "waiting" ? <LoadingSpinner size="sm" /> : <><Plus className="h-4 w-4 mr-1" /> Register Key</>}
              </Button>
              <Button variant="outline" onClick={() => { setShowAddForm(false); setAddingName(""); }}>
                Cancel
              </Button>
            </div>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-6">
            <LoadingSpinner />
          </div>
        ) : devices.length === 0 && !showAddForm ? (
          <div className="text-center py-8 text-muted-foreground">
            <KeyRound className="h-10 w-10 mx-auto mb-3 opacity-30" />
            <p className="text-sm">No security keys registered.</p>
            <p className="text-xs mt-1">Add a security key above to enable passwordless login.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {devices.map((device) => (
              <div key={device.id} className="flex items-center justify-between p-4 border rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-primary/10 rounded-full">
                    <KeyRound className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-sm">{device.name}</p>
                      {device.last_used && <Badge variant="secondary" className="text-xs">Active</Badge>}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Registered {formatDate(device.created_at)}
                      {device.last_used && ` · Last used ${formatDate(device.last_used)}`}
                    </p>
                    {device.credential_id && (
                      <p className="text-xs font-mono text-muted-foreground mt-0.5">
                        ID: {device.credential_id.slice(0, 20)}...
                      </p>
                    )}
                  </div>
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleRemove(device.id)}
                  disabled={removingId === device.id}
                  className="text-destructive hover:text-destructive"
                >
                  {removingId === device.id ? <LoadingSpinner size="sm" /> : <Trash2 className="h-4 w-4" />}
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}