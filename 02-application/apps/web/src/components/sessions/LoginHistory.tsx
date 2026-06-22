"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingSpinner } from "../auth/LoadingSpinner";
import { CheckCircle2, XCircle, Laptop, Smartphone, Tablet, Monitor } from "lucide-react";

interface LoginEvent {
  id: string;
  device: string;
  browser: string;
  ip: string;
  location: string;
  timestamp: string;
  status: "success" | "failed";
}

const mockHistory: LoginEvent[] = [
  {
    id: "1",
    device: "MacBook Pro",
    browser: "Chrome 120.0",
    ip: "192.168.1.1",
    location: "San Francisco, CA",
    timestamp: "2024-01-15T10:30:00Z",
    status: "success",
  },
  {
    id: "2",
    device: "iPhone 15",
    browser: "Safari 17.0",
    ip: "192.168.1.2",
    location: "San Francisco, CA",
    timestamp: "2024-01-15T09:15:00Z",
    status: "success",
  },
  {
    id: "3",
    device: "Unknown",
    browser: "Chrome 119.0",
    ip: "203.0.113.1",
    location: "New York, NY",
    timestamp: "2024-01-14T23:45:00Z",
    status: "failed",
  },
  {
    id: "4",
    device: "Windows PC",
    browser: "Firefox 121.0",
    ip: "192.168.1.3",
    location: "New York, NY",
    timestamp: "2024-01-14T16:45:00Z",
    status: "success",
  },
  {
    id: "5",
    device: "iPad Pro",
    browser: "Safari 17.0",
    ip: "192.168.1.4",
    location: "Los Angeles, CA",
    timestamp: "2024-01-13T14:20:00Z",
    status: "success",
  },
];

function DeviceIcon({ device }: { device: string }) {
  if (device.toLowerCase().includes("iphone") || device.toLowerCase().includes("android")) {
    return <Smartphone className="h-4 w-4" />;
  }
  if (device.toLowerCase().includes("ipad") || device.toLowerCase().includes("tablet")) {
    return <Tablet className="h-4 w-4" />;
  }
  if (device.toLowerCase().includes("macbook") || device.toLowerCase().includes("laptop")) {
    return <Laptop className="h-4 w-4" />;
  }
  return <Monitor className="h-4 w-4" />;
}

export function LoginHistory() {
  const [history] = useState<LoginEvent[]>(mockHistory);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Login History</CardTitle>
        <CardDescription>
          Recent login attempts to your account
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {history.map((event) => (
            <div
              key={event.id}
              className="flex items-center justify-between p-3 rounded-lg border bg-card"
            >
              <div className="flex items-center gap-3">
                <div className="p-2 bg-muted rounded-full">
                  <DeviceIcon device={event.device} />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-medium">{event.device}</p>
                    {event.status === "success" ? (
                      <CheckCircle2 className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-destructive" />
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {event.browser} · {event.location}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    IP: {event.ip}
                  </p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xs text-muted-foreground">
                  {formatDate(event.timestamp)}
                </p>
                <span
                  className={`text-xs px-2 py-0.5 rounded-full ${
                    event.status === "success"
                      ? "bg-green-100 text-green-800"
                      : "bg-red-100 text-red-800"
                  }`}
                >
                  {event.status === "success" ? "Success" : "Failed"}
                </span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
