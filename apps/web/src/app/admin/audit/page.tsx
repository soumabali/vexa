'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import DataTable from '@/components/admin/tables/data-table';
import { useAsyncData } from '@/hooks/useAsyncData';
import { MaterialIcon } from "@/components/ui/material-icon";

interface AuditLog {
  id: string;
  timestamp: string;
  user: string;
  action: string;
  resource: string;
  resourceType: string;
  status: 'success' | 'failure' | 'warning';
  ip: string;
  userAgent: string;
  details: string;
}

export default function AuditLogs() {
  const { data: logsResp } = useAsyncData(async () => {
    const res = await fetch('/api/admin/audit-logs');
    const data = await res.json();
    return data.logs || [];
  });
  const logs = (logsResp as AuditLog[] | null) ?? [];
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('all');

  const handleExport = () => {
    const csv = [
      ['ID', 'Timestamp', 'User', 'Action', 'Resource', 'Type', 'Status', 'IP', 'Details'].join(','),
      ...logs.map((log) =>
        [
          log.id,
          log.timestamp,
          log.user,
          log.action,
          log.resource,
          log.resourceType,
          log.status,
          log.ip,
          `"${log.details}"`,
        ].join(',')
      ),
    ].join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const filteredLogs = logs.filter((log) => {
    const matchesSearch =
      log.user.toLowerCase().includes(search.toLowerCase()) ||
      log.action.toLowerCase().includes(search.toLowerCase()) ||
      log.resource.toLowerCase().includes(search.toLowerCase());
    const matchesFilter = filter === 'all' || log.status === filter;
    return matchesSearch && matchesFilter;
  });

  const columns = [
    { key: 'timestamp', header: 'Timestamp' },
    { key: 'user', header: 'User' },
    { key: 'action', header: 'Action' },
    { key: 'resource', header: 'Resource' },
    {
      key: 'status',
      header: 'Status',
      render: (log: AuditLog) => (
        <Badge
          variant={
            log.status === 'success'
              ? 'default'
              : log.status === 'failure'
              ? 'destructive'
              : 'secondary'
          }
        >
          {log.status}
        </Badge>
      ),
    },
    { key: 'ip', header: 'IP Address' },
    {
      key: 'details',
      header: 'Details',
      render: (log: AuditLog) => (
        <span className="text-sm text-on-surface-variant truncate max-w-[200px]">{log.details}</span>
      ),
    },
  ];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold">Audit Logs</h1>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <MaterialIcon name="search" className="h-4 w-4 text-on-surface-variant" />
              <Input
                placeholder="Search logs..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="max-w-sm"
              />
              <select
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="border rounded-md px-3 py-2"
              >
                <option value="all">All Status</option>
                <option value="success">Success</option>
                <option value="failure">Failure</option>
                <option value="warning">Warning</option>
              </select>
            </div>
            <Button onClick={handleExport}>
              <MaterialIcon name="download" className="h-4 w-4 mr-2" />
              Export CSV
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <DataTable data={filteredLogs} columns={columns} />
        </CardContent>
      </Card>
    </div>
  );
}