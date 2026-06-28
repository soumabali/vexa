"use client";

import { useState, useEffect } from "react";
import { useSession } from "next-auth/react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";

interface Team {
  id: string;
  name: string;
}

interface Share {
  id: string;
  credential_id: string;
  credential_name?: string;
  sender_id: string;
  sender_email?: string;
  recipient_id: string;
  recipient_email?: string;
  permission: "read_only" | "read_write" | "admin";
  status: "pending" | "accepted" | "rejected" | "revoked" | "expired";
  expiry_time?: string;
  created_at: string;
  accepted_at?: string;
  revoked_at?: string;
}

export default function ShareManagementPage() {
  const sessionData = useSession();
  const session = sessionData?.data || null;
  const [activeTab, setActiveTab] = useState("received");
  const [shares, setShares] = useState<Share[]>([]);
  const [loading, setLoading] = useState(false);
  const [showShareDialog, setShowShareDialog] = useState(false);
  const [teams, setTeams] = useState<Team[]>([]);
  const [shareForm, setShareForm] = useState({
    credential_id: "",
    team_id: "",
    permission: "read_only",
    expiry_days: "",
  });

  // Fetch shares
  const fetchShares = async (sent: boolean) => {
    setLoading(true);
    try {
      const res = await fetch(`/api/vault/shares?sent=${sent}`, {
        headers: { Authorization: `Bearer ${session?.accessToken}` },
      });
      if (!res.ok) throw new Error("Failed to fetch shares");
      const data = await res.json();
      setShares(data.shares || []);
    } catch (err) {
      toast.error("Failed to load shares");
    } finally {
      setLoading(false);
    }
  };

  // Fetch teams for share target selector
  const fetchTeams = async () => {
    try {
      const res = await fetch(`/api/teams`, {
        headers: { Authorization: `Bearer ${session?.accessToken}` },
      });
      if (!res.ok) return;
      const data = await res.json();
      setTeams(data.teams || []);
    } catch (err) {
      // non-fatal: dialog shows empty selector
    }
  };

  useEffect(() => {
    const token = session?.accessToken;
    if (!token) return;
    // Defer the async state-setting call out of the effect body to avoid cascading renders.
    const handle = setTimeout(() => {
      void fetchShares(activeTab === "sent");
      void fetchTeams();
    }, 0);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, session]);

  // Create share
  const handleShare = async () => {
    try {
      const res = await fetch(`/api/vault/credentials/${shareForm.credential_id}/share`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${session?.accessToken}`,
        },
        body: JSON.stringify({
          team_id: shareForm.team_id,
          // Backend expects permissions[]; wrap single-select value.
          permissions: [shareForm.permission],
          expiry_days: shareForm.expiry_days ? parseInt(shareForm.expiry_days) : undefined,
        }),
      });
      if (!res.ok) throw new Error("Failed to share");
      toast.success("Credential shared successfully!");
      setShowShareDialog(false);
      fetchShares(false);
    } catch (err) {
      toast.error("Failed to share credential");
    }
  };

  // Accept share
  const handleAccept = async (shareId: string) => {
    try {
      const res = await fetch(`/api/vault/shares/${shareId}/accept`, {
        method: "POST",
        headers: { Authorization: `Bearer ${session?.accessToken}` },
      });
      if (!res.ok) throw new Error("Failed to accept");
      toast.success("Share accepted!");
      fetchShares(false);
    } catch (err) {
      toast.error("Failed to accept share");
    }
  };

  // Revoke share
  const handleRevoke = async (shareId: string) => {
    try {
      const res = await fetch(`/api/vault/shares/${shareId}/revoke`, {
        method: "POST",
        headers: { Authorization: `Bearer ${session?.accessToken}` },
      });
      if (!res.ok) throw new Error("Failed to revoke");
      toast.success("Share revoked!");
      fetchShares(activeTab === "sent");
    } catch (err) {
      toast.error("Failed to revoke share");
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, string> = {
      pending: "bg-yellow-100 text-yellow-800",
      accepted: "bg-green-100 text-green-800",
      rejected: "bg-red-100 text-red-800",
      revoked: "bg-gray-100 text-gray-800",
      expired: "bg-orange-100 text-orange-800",
    };
    return (
      <Badge className={variants[status] || ""}>{status}</Badge>
    );
  };

  const getPermissionLabel = (perm: string) => {
    const labels: Record<string, string> = {
      read_only: "Read Only",
      read_write: "Read & Write",
      admin: "Admin (Can Reshare)",
    };
    return labels[perm] || perm;
  };

  return (
    <div className="container mx-auto py-8 px-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Shared Credentials</h1>
        <Button onClick={() => setShowShareDialog(true)}>
          Share Credential
        </Button>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="mb-4">
          <TabsTrigger value="received">Received</TabsTrigger>
          <TabsTrigger value="sent">Sent</TabsTrigger>
        </TabsList>

        <TabsContent value="received">
          {loading ? (
            <p>Loading...</p>
          ) : (
            <div className="grid gap-4">
              {shares.map((share) => (
                <Card key={share.id}>
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div>
                        <CardTitle>
                          {share.credential_name || "Credential"}
                        </CardTitle>
                        <p className="text-sm text-muted-foreground">
                          From: {share.sender_email || share.sender_id}
                        </p>
                      </div>
                      {getStatusBadge(share.status)}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <p>Permission: {getPermissionLabel(share.permission)}</p>
                      <p>Shared: {new Date(share.created_at).toLocaleDateString()}</p>
                      {share.expiry_time && (
                        <p>Expires: {new Date(share.expiry_time).toLocaleDateString()}</p>
                      )}
                      {share.status === "pending" && (
                        <div className="flex gap-2 mt-4">
                          <Button
                            size="sm"
                            onClick={() => handleAccept(share.id)}
                          >
                            Accept
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRevoke(share.id)}
                          >
                            Reject
                          </Button>
                        </div>
                      )}
                      {share.status === "accepted" && (
                        <div className="flex gap-2 mt-4">
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleRevoke(share.id)}
                          >
                            Revoke Access
                          </Button>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
              {shares.length === 0 && (
                <p className="text-muted-foreground">No shared credentials.</p>
              )}
            </div>
          )}
        </TabsContent>

        <TabsContent value="sent">
          {loading ? (
            <p>Loading...</p>
          ) : (
            <div className="grid gap-4">
              {shares.map((share) => (
                <Card key={share.id}>
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div>
                        <CardTitle>
                          {share.credential_name || "Credential"}
                        </CardTitle>
                        <p className="text-sm text-muted-foreground">
                          To: {share.recipient_email || share.recipient_id}
                        </p>
                      </div>
                      {getStatusBadge(share.status)}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <p>Permission: {getPermissionLabel(share.permission)}</p>
                      <p>Shared: {new Date(share.created_at).toLocaleDateString()}</p>
                      {share.status === "accepted" && (
                        <div className="flex gap-2 mt-4">
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleRevoke(share.id)}
                          >
                            Revoke Access
                          </Button>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
              {shares.length === 0 && (
                <p className="text-muted-foreground">No shared credentials.</p>
              )}
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Share Dialog */}
      {showShareDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold mb-4">Share Credential</h2>
            <div className="space-y-4">
              <div>
                <Label>Credential ID</Label>
                <Input
                  value={shareForm.credential_id}
                  onChange={(e) =>
                    setShareForm({ ...shareForm, credential_id: e.target.value })
                  }
                  placeholder="Enter credential ID"
                />
              </div>
              <div>
                <Label>Recipient Team</Label>
                <Select
                  value={shareForm.team_id}
                  onValueChange={(val: string) =>
                    setShareForm({ ...shareForm, team_id: val })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select team" />
                  </SelectTrigger>
                  <SelectContent>
                    {teams.length === 0 ? (
                      <SelectItem value="__none__" disabled>
                        No teams available
                      </SelectItem>
                    ) : (
                      teams.map((t) => (
                        <SelectItem key={t.id} value={t.id}>
                          {t.name}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Permission</Label>
                <Select
                  value={shareForm.permission}
                  onValueChange={(val: "read_only" | "read_write" | "admin") =>
                    setShareForm({ ...shareForm, permission: val })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="read_only">Read Only</SelectItem>
                    <SelectItem value="read_write">Read & Write</SelectItem>
                    <SelectItem value="admin">Admin (Can Reshare)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Expiry (days, optional)</Label>
                <Input
                  type="number"
                  value={shareForm.expiry_days}
                  onChange={(e) =>
                    setShareForm({ ...shareForm, expiry_days: e.target.value })
                  }
                  placeholder="Never expires"
                />
              </div>
              <div className="flex gap-2">
                <Button onClick={handleShare}>Share</Button>
                <Button
                  variant="outline"
                  onClick={() => setShowShareDialog(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
