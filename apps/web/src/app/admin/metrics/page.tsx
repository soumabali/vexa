'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import LineChart from '@/components/admin/charts/line-chart';
import { useAsyncData } from '@/hooks/useAsyncData';
import { MaterialIcon } from "@/components/ui/material-icon";

interface MetricData {
  timestamp: string;
  cpu: number;
  memory: number;
  network: number;
  disk: number;
}

interface AlertRule {
  id: string;
  name: string;
  metric: string;
  threshold: number;
  operator: '>' | '<' | '=';
  status: 'active' | 'inactive';
  severity: 'critical' | 'warning' | 'info';
}

export default function MetricsPage() {
  const { data: metricsSeed } = useAsyncData(async () => {
    const res = await fetch('/api/admin/metrics/history');
    const data = await res.json();
    return data.metrics || [];
  });
  const [metrics, setMetrics] = useState<MetricData[]>(() => (metricsSeed as MetricData[] | null) ?? []);
  const { data: alertsResp } = useAsyncData(async () => {
    const res = await fetch('/api/admin/alerts');
    const data = await res.json();
    return data.alerts || [];
  });
  const alerts = (alertsResp as AlertRule[] | null) ?? [];
  const [wsConnected, setWsConnected] = useState(false);

  useEffect(() => {
    if (metricsSeed && metrics.length === 0) {
      Promise.resolve().then(() => {
        if (metricsSeed && metrics.length === 0) {
          setMetrics(metricsSeed as MetricData[]);
        }
      });
    }
  }, [metricsSeed, metrics.length]);

  useEffect(() => {
    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WS_URL}/api/admin/metrics`);
    ws.onopen = () => setWsConnected(true);
    ws.onclose = () => setWsConnected(false);
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'metric') {
        setMetrics((prev) => [...prev.slice(-100), data.payload]);
      }
    };
    return () => ws.close();
  }, []);

  const getLatestValue = (key: keyof MetricData) => {
    if (metrics.length === 0) return 0;
    return metrics[metrics.length - 1][key] as number;
  };

  const statCards = [
    {
      title: 'CPU Usage',
      value: `${getLatestValue('cpu').toFixed(1)}%`,
      icon: "memory",
      color: 'text-blue-500',
    },
    {
      title: 'Memory Usage',
      value: `${getLatestValue('memory').toFixed(1)}%`,
      icon: "memory",
      color: 'text-green-500',
    },
    {
      title: 'Network I/O',
      value: `${getLatestValue('network').toFixed(1)} MB/s`,
      icon: "lan",
      color: 'text-purple-500',
    },
    {
      title: 'Disk I/O',
      value: `${getLatestValue('disk').toFixed(1)} MB/s`,
      icon: "storage",
      color: 'text-orange-500',
    },
  ];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">System Metrics</h1>
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-sm text-on-surface-variant">
            {wsConnected ? 'Live' : 'Disconnected'}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((card, index) => (
          <Card key={index}>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div className="space-y-2">
                  <p className="text-sm font-medium text-on-surface-variant">{card.title}</p>
                  <p className="text-2xl font-bold">{card.value}</p>
                </div>
              <MaterialIcon name={card.icon} className={`h-8 w-8 ${card.color}`} />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Tabs defaultValue="cpu">
        <TabsList>
          <TabsTrigger value="cpu">CPU</TabsTrigger>
          <TabsTrigger value="memory">Memory</TabsTrigger>
          <TabsTrigger value="network">Network</TabsTrigger>
          <TabsTrigger value="disk">Disk</TabsTrigger>
        </TabsList>

        <TabsContent value="cpu">
          <Card>
            <CardHeader>
              <CardTitle>CPU Usage Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              <LineChart
                data={metrics.map((m) => ({ label: m.timestamp, value: m.cpu }))}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="memory">
          <Card>
            <CardHeader>
              <CardTitle>Memory Usage Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              <LineChart
                data={metrics.map((m) => ({ label: m.timestamp, value: m.memory }))}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="network">
          <Card>
            <CardHeader>
              <CardTitle>Network I/O Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              <LineChart
                data={metrics.map((m) => ({ label: m.timestamp, value: m.network }))}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="disk">
          <Card>
            <CardHeader>
              <CardTitle>Disk I/O Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              <LineChart
                data={metrics.map((m) => ({ label: m.timestamp, value: m.disk }))}
              />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Card>
        <CardHeader>
          <CardTitle>Alert Rules</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {alerts.map((alert) => (
              <div
                key={alert.id}
                className="flex items-center justify-between p-4 border rounded-lg"
              >
                <div className="space-y-1">
                  <p className="font-medium">{alert.name}</p>
                  <p className="text-sm text-on-surface-variant">
                    {alert.metric} {alert.operator} {alert.threshold}
                  </p>
                </div>
                <div className="flex items-center gap-4">
                  <Badge
                    variant={alert.severity === 'critical' ? 'destructive' : 'default'}
                  >
                    {alert.severity}
                  </Badge>
                  <Badge variant={alert.status === 'active' ? 'default' : 'secondary'}>
                    {alert.status}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
