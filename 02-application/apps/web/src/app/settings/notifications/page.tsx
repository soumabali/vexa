"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { MaterialIcon } from "@/components/ui/material-icon";
import { toast } from "sonner";

interface ChannelToggle {
  id: string;
  label: string;
  desc: string;
  enabled: boolean;
}

interface EventToggle {
  id: string;
  label: string;
  desc: string;
  icon: string;
  enabled: boolean;
}

interface ActivityItem {
  id: string;
  icon: string;
  event: string;
  timestamp: string;
}

export default function NotificationsSettingsPage() {
  const [channels, setChannels] = useState<ChannelToggle[]>([
    { id: "desktop", label: "Desktop Notifications", desc: "Show OS-level alerts for important events.", enabled: true },
    { id: "email", label: "Email Alerts", desc: "Send a digest of critical events to your inbox.", enabled: true },
  ]);
  const [quietHours, setQuietHours] = useState<boolean>(true);
  const [quietFrom, setQuietFrom] = useState<string>("22:00");
  const [quietTo, setQuietTo] = useState<string>("07:00");

  const [events, setEvents] = useState<EventToggle[]>([
    { id: "ssh", label: "SSH Login Alerts", desc: "Notify on every successful or failed SSH login.", icon: "terminal", enabled: true },
    { id: "host-down", label: "Host Down", desc: "Alert when a monitored host stops responding.", icon: "dns", enabled: true },
    { id: "cred-access", label: "Credential Access", desc: "Notify when a stored credential is read or used.", icon: "key", enabled: false },
    { id: "tunnel", label: "Tunnel Status", desc: "Alert on tunnel connect, disconnect, or failure.", icon: "tunnel", enabled: true },
  ]);

  const [slackConnected, setSlackConnected] = useState<boolean>(false);

  const [activity] = useState<ActivityItem[]>([
    { id: "a1", icon: "terminal", event: "SSH login on prod-web-01", timestamp: "2 min ago" },
    { id: "a2", icon: "dns", event: "Host db-primary went down", timestamp: "18 min ago" },
    { id: "a3", icon: "key", event: "Credential ‘deploy-key’ accessed", timestamp: "1 hr ago" },
    { id: "a4", icon: "tunnel", event: "Tunnel tunnel-eu reconnected", timestamp: "3 hr ago" },
  ]);

  const toggleChannel = (id: string) =>
    setChannels((prev) => prev.map((c) => (c.id === id ? { ...c, enabled: !c.enabled } : c)));

  const toggleEvent = (id: string) =>
    setEvents((prev) => prev.map((e) => (e.id === id ? { ...e, enabled: !e.enabled } : e)));

  const handleSave = () => {
    toast.success("Notification preferences saved", {
      icon: <MaterialIcon name="check" size="sm" className="text-primary" />,
    });
  };

  const handleClearActivity = () => {
    toast.success("Recent activity cleared");
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Notifications</h1>
        <p className="text-sm text-on-surface-variant">
          Manage how and when you receive alerts from Vexa.
        </p>
      </div>

      {/* Channels */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="notifications_active" size="sm" className="text-on-surface-variant" />
            Notification Channels
          </CardTitle>
          <CardDescription>Choose where alerts are delivered.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {channels.map((channel) => (
            <div key={channel.id} className="flex items-center justify-between gap-4">
              <div className="space-y-0.5">
                <p className="text-sm font-medium text-on-surface">{channel.label}</p>
                <p className="text-xs text-on-surface-variant">{channel.desc}</p>
              </div>
              <Switch checked={channel.enabled} onCheckedChange={() => toggleChannel(channel.id)} />
            </div>
          ))}

          {/* Quiet hours */}
          <div className="rounded-lg border border-outline-variant bg-surface-container-lowest p-4 space-y-4">
            <div className="flex items-center justify-between gap-4">
              <div className="space-y-0.5">
                <p className="text-sm font-medium text-on-surface">Quiet Hours</p>
                <p className="text-xs text-on-surface-variant">Mute non-critical alerts during set hours.</p>
              </div>
              <Switch checked={quietHours} onCheckedChange={setQuietHours} />
            </div>
            {quietHours && (
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label htmlFor="quiet-from" className="text-on-surface-variant">From</Label>
                  <Input
                    id="quiet-from"
                    type="time"
                    value={quietFrom}
                    onChange={(e) => setQuietFrom(e.target.value)}
                    className="bg-surface-container-lowest border-outline-variant text-on-surface focus-visible:border-primary"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="quiet-to" className="text-on-surface-variant">To</Label>
                  <Input
                    id="quiet-to"
                    type="time"
                    value={quietTo}
                    onChange={(e) => setQuietTo(e.target.value)}
                    className="bg-surface-container-lowest border-outline-variant text-on-surface focus-visible:border-primary"
                  />
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Event triggers */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="tune" size="sm" className="text-on-surface-variant" />
            Event Triggers
          </CardTitle>
          <CardDescription>Select which events generate notifications.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {events.map((event) => (
            <div key={event.id} className="flex items-center justify-between gap-4">
              <div className="flex items-start gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-surface-container-high text-on-surface-variant">
                  <MaterialIcon name={event.icon} size="sm" />
                </div>
                <div className="space-y-0.5">
                  <p className="text-sm font-medium text-on-surface">{event.label}</p>
                  <p className="text-xs text-on-surface-variant">{event.desc}</p>
                </div>
              </div>
              <Switch checked={event.enabled} onCheckedChange={() => toggleEvent(event.id)} />
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Integrations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-on-surface">
            <MaterialIcon name="hub" size="sm" className="text-on-surface-variant" />
            Integrations
          </CardTitle>
          <CardDescription>Route alerts to external services.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between gap-4 rounded-lg border border-outline-variant bg-surface-container-lowest p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-surface-container-high text-on-surface-variant">
                <MaterialIcon name="tag" size="md" />
              </div>
              <div className="space-y-0.5">
                <p className="text-sm font-medium text-on-surface">Slack</p>
                <p className="text-xs text-on-surface-variant">
                  {slackConnected ? "Workspace connected" : "Send alerts to a Slack channel"}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Switch checked={slackConnected} onCheckedChange={setSlackConnected} />
              {!slackConnected && (
                <Button variant="default" className="bg-tertiary text-on-tertiary hover:bg-tertiary/90">
                  Connect Workspace
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Recent activity */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <div className="space-y-1">
            <CardTitle className="flex items-center gap-2 text-on-surface">
              <MaterialIcon name="history" size="sm" className="text-on-surface-variant" />
              Recent Activity
            </CardTitle>
            <CardDescription>Latest notification events.</CardDescription>
          </div>
          <button
            type="button"
            onClick={handleClearActivity}
            className="text-sm text-primary hover:text-primary/80"
          >
            Clear all
          </button>
        </CardHeader>
        <CardContent>
          <ul className="space-y-3">
            {activity.map((item) => (
              <li key={item.id} className="flex items-center gap-3">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-surface-container-high text-on-surface-variant">
                  <MaterialIcon name={item.icon} size="sm" />
                </div>
                <p className="flex-1 text-sm text-on-surface">{item.event}</p>
                <span className="text-xs text-on-surface-variant">{item.timestamp}</span>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={handleSave}>Save Changes</Button>
      </div>
    </div>
  );
}