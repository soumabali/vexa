'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import DataTable from '@/components/admin/tables/data-table';
import { Server, Activity, AlertTriangle } from 'lucide-react';

interface Host {
  id: string;
  name: string;
  hostname: string;
  port: number;
  protocol: 'ssh' | 'rdp' | 'vnc';
  status: 'online' | 'offline' | 'unreachable';
  os: string;
  lastSeen: string;
  healthStatus: 'healthy' | 'warning' | 'critical';
  latency: string;
}

export default function HostManagement() {
  const [hosts, setHosts] = useState<Host[]>([]);

  useEffect(() => {
    fetchHosts();
  }, []);

  const fetchHosts = async () => {
    try {
      const res = await fetch('/api/admin/hosts');
      const data = await res.json();
      setHosts(data.hosts || []);
    } catch (error) {
      console.error('Failed to fetch hosts:', error);
    }
  };

  const handleRunHealthCheck = async (id: string) => {
    try {
      await fetch(`/api/admin/hosts/${id}/health-check`, { method: 'POST' });
      fetchHosts();
    } catch (error) {
      console.error('Failed to run health check:', error);
    }
  };

  const columns = [
    { key: 'name', header: 'Name' },
    { key: 'hostname', header: 'Hostname' },
    { key: 'port', header: 'Port' },
    {
      key: 'protocol',
      header: 'Protocol',
      render: (host: Host) => (
        <Badge variant="outline">{host.protocol.toUpperCase()}</Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (host: Host) => (
        <Badge
          variant={
            host.status === 'online'
              ? 'default'
              : host.status === 'offline'
              ? 'secondary'
              : 'destructive'
          }
        >
          {host.status}
        </Badge>
      ),
    },
    { key: 'os', header: 'OS' },
    {
      key: 'healthStatus',
      header: 'Health',
      render: (host: Host) => (
        <div className="flex items-center gap-2">
          {host.healthStatus === 'healthy' && <Activity className="h-4 w-4 text-green-500" />}
          {host.healthStatus === 'warning' && <AlertTriangle className="h-4 w-4 text-yellow-500" />}
          {host.healthStatus === 'critical' && <AlertTriangle className="h-4 w-4 text-red-500" />}
          <span className="capitalize">{host.healthStatus}</span>
        </div>
      ),
    },
    { key: 'latency', header: 'Latency' },
    { key: 'lastSeen', header: 'Last Seen' },
    {
      key: 'actions',
      header: 'Actions',
      render: (host: Host) => (
        <Button variant="ghost" size="sm" onClick={() => handleRunHealthCheck(host.id)}>
          <Activity className="h-4 w-4 mr-2" />
          Health Check
        </Button>
      ),
    },
  ];

  const stats = [
    { label: 'Total Hosts', value: hosts.length, icon: Server },
    { label: 'Online', value: hosts.filter((h) => h.status === 'online').length, icon: Activity },
    { label: 'Offline', value: hosts.filter((h) => h.status === 'offline').length, icon: AlertTriangle },
    {
      label: 'Unreachable',
      value: hosts.filter((h) => h.status === 'unreachable').length,
      icon: AlertTriangle,
    },
  ];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold">Host Management</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <Card key={index}>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div className="space-y-2">
                  <p className="text-sm font-medium text-muted-foreground">{stat.label}</p>
                  <p className="text-2xl font-bold">{stat.value}</p>
                </div>
                <stat.icon className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Hosts</CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable data={hosts} columns={columns} />
        </CardContent>
      </Card>
    </div>
  );
}
