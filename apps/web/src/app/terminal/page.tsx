"use client";

import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { useState, useEffect, useRef, useCallback } from "react";
import { MaterialIcon } from "@/components/ui/material-icon";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useQuery } from "@tanstack/react-query";
import { hostsApi } from "@/lib/api/hosts";
import type { Terminal } from "xterm";
import type { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";

interface Host {
  id: string;
  name: string;
  address: string;
  port: number;
  protocol: string;
}

interface SessionTab {
  id: string;
  label: string;
  hostId?: string;
  hostName?: string;
  status: "idle" | "connecting" | "connected" | "error";
  error?: string;
}

const fetchHosts = async (): Promise<Host[]> => {
  const data = await hostsApi.list();
  return (data.hosts ?? []).map((h) => ({
    id: h.id ?? "",
    name: h.name ?? "",
    address: h.host ?? h.address ?? "",
    port: h.port ?? 22,
    protocol: h.hostType ?? "ssh",
  }));
};

export default function TerminalPage() {
  const [tabs, setTabs] = useState<SessionTab[]>([
    { id: "local", label: "Local", status: "idle" },
  ]);
  const [activeTab, setActiveTab] = useState("local");
  const [selectedHost, setSelectedHost] = useState("");
  const terminalRef = useRef<HTMLDivElement>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const { data: hosts = [] } = useQuery({
    queryKey: ["hosts"],
    queryFn: fetchHosts,
  });

  useEffect(() => {
    if (terminalRef.current && !xtermRef.current) {
      void import("xterm").then(({ Terminal }) => {
        void import("xterm-addon-fit").then(({ FitAddon }) => {
          const term = new Terminal({
            cursorBlink: true,
            fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
            fontSize: 14,
            theme: {
              background: "#000000",
              foreground: "#22c55e",
              cursor: "#22c55e",
            },
          });
          const fitAddon = new FitAddon();
          term.loadAddon(fitAddon);
          term.open(terminalRef.current as HTMLDivElement);
          fitAddon.fit();
          fitAddonRef.current = fitAddon;
          xtermRef.current = term;

          term.writeln("Welcome to vexa terminal.");
          term.writeln("Select a host and click Connect to start a remote session.");
          term.writeln("");

          const handleResize = () => {
            fitAddon.fit();
            if (wsRef.current?.readyState === WebSocket.OPEN) {
              const dims = fitAddon.proposeDimensions();
              if (dims) {
                wsRef.current.send(
                  JSON.stringify({
                    type: "resize",
                    data: { cols: dims.cols, rows: dims.rows },
                  })
                );
              }
            }
          };

          window.addEventListener("resize", handleResize);

          term.onData((data) => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
              wsRef.current.send(JSON.stringify({ type: "data", data }));
            }
          });
        });
      });

      return () => {
        window.removeEventListener("resize", () => null);
        wsRef.current?.close();
        xtermRef.current?.dispose();
      };
    }
  }, []);

  const updateActiveTab = useCallback(
    (updater: (tab: SessionTab) => SessionTab) => {
      setTabs((tabList) =>         tabList.map((t) =>           (t.id === activeTab ? updater(t) : t))
      );
    },
    [activeTab]
  );

  const connectToHost = useCallback(
    async (hostId: string) => {
      const host = hosts.find((h) => h.id === hostId);
      if (!host) return;

      const token = localStorage.getItem("access_token");
      if (!token) {
        updateActiveTab((t) => ({ ...t, status: "error", error: "Not authenticated" }));
        return;
      }

      updateActiveTab((t) => ({ ...t, hostId, hostName: host.name, status: "connecting" }));

      const baseWsUrl = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080";
      const wsUrl = `${baseWsUrl}/api/v1/ws/terminal?host_id=${hostId}&token=${encodeURIComponent(token)}`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        updateActiveTab((t) => ({ ...t, status: "connected" }));
        xtermRef.current?.clear();
        xtermRef.current?.writeln(`Connected to ${host.name} (${host.address}:${host.port}).`);

        const dims = fitAddonRef.current?.proposeDimensions();
        if (dims) {
          ws.send(
            JSON.stringify({
              type: "resize",
              data: { cols: dims.cols, rows: dims.rows },
            })
          );
        }
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as {
            type: string;
            data: string;
          };
          if (message.type === "data") {
            xtermRef.current?.write(message.data);
          } else if (message.type === "error") {
            xtermRef.current?.writeln(`\r\nError: ${message.data}`);
            updateActiveTab((t) => ({ ...t, status: "error", error: message.data }));
          }
        } catch {
          xtermRef.current?.write(event.data);
        }
      };

      ws.onerror = () => {
        updateActiveTab((t) => ({ ...t, status: "error", error: "WebSocket error" }));
      };

      ws.onclose = () => {
        updateActiveTab((t) => (t.status !== "error" ? { ...t, status: "idle" } : t));
      };
    },
    [activeTab, hosts, updateActiveTab]
  );

  const addTab = () => {
    const nextIndex = tabs.length;
    const newTab: SessionTab = {
      id: `session-${nextIndex}`,
      label: nextIndex === 1 ? "Session 1" : `Session ${nextIndex}`,
      status: "idle",
    };
    setTabs((prev) => [...prev, newTab]);
    setActiveTab(newTab.id);
  };

  const closeTab = (id: string) => {
    if (tabs.length <= 1) return;
    const remaining = tabs.filter((t) => t.id !== id);
    setTabs(remaining);
    if (activeTab === id) {
      setActiveTab(remaining[remaining.length - 1].id);
    }
    if (wsRef.current && id === activeTab) {
      wsRef.current.close();
      wsRef.current = null;
    }
  };

  const activeSession = tabs.find((t) => t.id === activeTab);
  const statusColor = {
    idle: "bg-gray-500",
    connecting: "bg-yellow-500",
    connected: "bg-green-500",
    error: "bg-red-500",
  }[activeSession?.status ?? "idle"];

  return (
    <DashboardLayout>
      <div className="h-[calc(100vh-7rem)] bg-background flex flex-col border border-outline-variant rounded-lg overflow-hidden">
        {/* Tab bar */}
        <div className="tab-bar h-14 flex items-center gap-1 px-4 border-b border-outline-variant bg-surface-container shrink-0">
          {tabs.map((tab) => (
            <div
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`group flex items-center gap-2 px-4 h-full text-sm cursor-pointer transition-colors border-b-2 ${
                activeTab === tab.id
                  ? "border-primary bg-surface-container-high text-on-surface"
                  : "border-transparent text-on-surface-variant hover:bg-surface-variant"
              }`}
            >
              <MaterialIcon name="terminal" size="sm" className="text-on-surface-variant" />
              <span>{tab.hostName ?? tab.label}</span>
              {tabs.length > 1 && (
                <button
                  type="button"
                  aria-label={`Close tab ${tab.hostName ?? tab.label}`}
                  onClick={(e) => {
                    e.stopPropagation();
                    closeTab(tab.id);
                  }}
                  className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-surface-variant text-on-surface-variant hover:text-on-surface transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                >
                  <MaterialIcon name="close" size="sm" />
                </button>
              )}
            </div>
          ))}
          <button
            type="button"
            aria-label="New tab"
            className="w-8 h-8 ml-1 flex items-center justify-center rounded text-on-surface-variant hover:bg-surface-variant transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            title="New tab"
            onClick={addTab}
          >
            <MaterialIcon name="add" size="sm" />
          </button>

          {/* Right side: host dropdown + connect */}
          <div className="ml-auto flex items-center gap-2">
            <Select value={selectedHost} onValueChange={setSelectedHost}>
              <SelectTrigger className="w-48 text-xs h-8 bg-surface-container-highest border-outline-variant text-on-surface">
                <SelectValue placeholder="Select host" />
              </SelectTrigger>
              <SelectContent>
                {hosts.map((host) => (
                  <SelectItem key={host.id} value={host.id} className="text-xs">
                    {host.name} ({host.address}:{host.port})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              size="sm"
              className="h-8 text-xs bg-primary-container text-on-primary-container hover:bg-primary-container/90"
              disabled={!selectedHost || activeSession?.status === "connecting"}
              onClick={() => selectedHost && connectToHost(selectedHost)}
            >
              <MaterialIcon name="power" size="sm" className="mr-1" />
              Connect
            </Button>
          </div>
        </div>

        {/* Status bar */}
        <div className="flex items-center justify-between gap-3 px-4 py-2 text-xs border-b border-outline-variant bg-surface-container shrink-0">
          <div className="flex items-center gap-2">
            <span className={`inline-block h-2 w-2 rounded-full ${statusColor}`} />
            <span className="text-on-surface-variant capitalize">
              {activeSession?.status ?? "idle"}
            </span>
            {activeSession?.error && (
              <span className="text-destructive flex items-center gap-1">
                <MaterialIcon name="error" size="sm" />
                {activeSession.error}
              </span>
            )}
          </div>
        </div>

        <Card className="flex-1 rounded-none border-0 bg-black">
          <CardContent className="p-0 h-full">
            <div ref={terminalRef} className="h-full w-full bg-black p-4 font-mono-code" />
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
