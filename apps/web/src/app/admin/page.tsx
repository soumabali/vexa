'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, Users, Server, Shield, Wifi, AlertTriangle, Clock } from 'lucide-react';
import StatsCards from '@/components/admin/stats-cards';
import LineChart from '@/components/admin/charts/line-chart';
import BarChart from '@/components/admin/charts/bar-chart';
import PieChart from '@/components/admin/charts/pie-chart';
import WorldMap from '@/components/admin/maps/world-map';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  activeSessions: number;
  totalHosts: number;
  onlineHosts: number;
  totalBandwidth: string;
  errorsLast24h: number;
  avgLatency: string;
}

interface AlertItem {
  id: string;
  severity: 'critical' | 'warning' | 'info';
  title: string;
  description: string;
  timestamp: string;
}

function isValidStatsPayload(payload: unknown): payload is DashboardStats {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    typeof p.totalUsers === 'number' &&
    typeof p.activeUsers === 'number' &&
    typeof p.activeSessions === 'number' &&
    typeof p.totalHosts === 'number'
  );
}

function isValidAlertPayload(payload: unknown): payload is AlertItem {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    typeof p.id === 'string' &&
    (p.severity === 'critical' || p.severity === 'warning' || p.severity === 'info') &&
    typeof p.title === 'string'
  );
}

export default function AdminDashboard() {
  const [stats, setStats] = useState<DashboardStats>({
    totalUsers: 0,
    activeUsers: 0,
    activeSessions: 0,
    totalHosts: 0,
    onlineHosts: 0,
    totalBandwidth: '0 GB',
    errorsLast24h: 0,
    avgLatency: '0 ms',
  });

  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [wsConnected, setWsConnected] = useState(false);

  const fetchStats = async () => {
    try {
      const res = await fetch('/api/admin/stats');
      const data = await res.json();
      setStats(data);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  useEffect(() => {
    // Fetch initial stats
    fetchStats();
    // WebSocket for real-time updates
    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WS_URL}/api/admin/metrics`);
    ws.onopen = () => setWsConnected(true);
    ws.onclose = () => setWsConnected(false);
    ws.onmessage = (event) => {
      let data: unknown;
      try {
        data = JSON.parse(event.data);
      } catch {
        console.warn('Failed to parse WebSocket message');
        return;
      }
      if (!data || typeof data !== 'object') return;
      const msg = data as Record<string, unknown>;
      if (msg.type === 'stats' && isValidStatsPayload(msg.payload)) {
        setStats(msg.payload);
      } else if (msg.type === 'alert' && isValidAlertPayload(msg.payload)) {
        setAlerts((prev) => [msg.payload as AlertItem, ...prev].slice(0, 50));
      }
    };
    return () => ws.close();
  }, []);

  const statItems = [
    { title: 'Total Users', value: stats.totalUsers, icon: Users, color: 'text-blue-500' },
    { title: 'Active Users', value: stats.activeUsers, icon: Activity, color: 'text-green-500' },
    { title: 'Active Sessions', value: stats.activeSessions, icon: Wifi, color: 'text-purple-500' },
    { title: 'Total Hosts', value: stats.totalHosts, icon: Server, color: 'text-orange-500' },
    { title: 'Online Hosts', value: stats.onlineHosts, icon: Shield, color: 'text-teal-500' },
    { title: 'Bandwidth (24h)', value: stats.totalBandwidth, icon: Clock, color: 'text-indigo-500' },
    { title: 'Errors (24h)', value: stats.errorsLast24h, icon: AlertTriangle, color: 'text-red-500' },
    { title: 'Avg Latency', value: stats.avgLatency, icon: Activity, color: 'text-cyan-500' },
  ];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-sm text-muted-foreground">
            {wsConnected ? 'Live' : 'Disconnected'}
          </span>
        </div>
      </div>

      <StatsCards items={statItems} />

      {alerts.length > 0 && (
        <div className="space-y-2">
          {alerts.slice(0, 3).map((alert) => (
            <Alert key={alert.id} variant={alert.severity === 'critical' ? 'destructive' : 'default'}>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>{alert.title}</AlertTitle>
              <AlertDescription>{alert.description}</AlertDescription>
            </Alert>
          ))}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Connections Over Time</CardTitle>
          </CardHeader>
          <CardContent>
            <LineChart
              data={[
                { label: '00:00', value: 120 },
                { label: '04:00', value: 80 },
                { label: '08:00', value: 250 },
                { label: '12:00', value: 380 },
                { label: '16:00', value: 420 },
                { label: '20:00', value: 350 },
                { label: '23:59', value: 200 },
              ]}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Hosts by Connections</CardTitle>
          </CardHeader>
          <CardContent>
            <BarChart
              data={[
                { label: 'prod-web-01', value: 450 },
                { label: 'prod-db-01', value: 320 },
                { label: 'staging-web', value: 180 },
                { label: 'dev-server', value: 120 },
                { label: 'bastion', value: 90 },
              ]}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Protocol Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <PieChart
              data={[
                { label: 'SSH', value: 65, color: '#10b981' },
                { label: 'RDP', value: 20, color: '#3b82f6' },
                { label: 'VNC', value: 10, color: '#f59e0b' },
                { label: 'SFTP', value: 5, color: '#8b5cf6' },
              ]}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Global Connection Map</CardTitle>
          </CardHeader>
          <CardContent>
            <WorldMap
              points={[
                { lat: 40.7128, lng: -74.006, country: 'US', connections: 450 },
                { lat: 51.5074, lng: -0.1278, country: 'GB', connections: 280 },
                { lat: 52.5200, lng: 13.405, country: 'DE', connections: 200 },
                { lat: 35.6762, lng: 139.6503, country: 'JP', connections: 150 },
                { lat: -33.8688, lng: 151.2093, country: 'AU', connections: 100 },
                { lat: 1.3521, lng: 103.8198, country: 'SG', connections: 80 },
                { lat: 19.0760, lng: 72.8777, country: 'IN', connections: 120 },
              ]}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}