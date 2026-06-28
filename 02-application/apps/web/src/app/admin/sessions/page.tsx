'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import DataTable from '@/components/admin/tables/data-table';
import { Power, Eye, PlayCircle } from 'lucide-react';

interface Session {
  id: string;
  user: string;
  host: string;
  protocol: 'ssh' | 'rdp' | 'vnc' | 'sftp';
  status: 'active' | 'idle' | 'closed';
  startedAt: string;
  duration: string;
  clientIP: string;
}

export default function SessionManagement() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeTab, setActiveTab] = useState('active');

  const fetchSessions = async () => {
    try {
      const res = await fetch('/api/admin/sessions');
      const data = await res.json();
      setSessions(data.sessions || []);
    } catch (error) {
      console.error('Failed to fetch sessions:', error);
    }
  };

  useEffect(() => {
    fetchSessions();
  }, []);

  const handleKillSession = async (id: string) => {
    if (!confirm('Are you sure you want to terminate this session?')) return;
    try {
      await fetch(`/api/admin/sessions/${id}/kill`, { method: 'POST' });
      fetchSessions();
    } catch (error) {
      console.error('Failed to kill session:', error);
    }
  };

  const handleViewDetails = (id: string) => {
    window.open(`/admin/sessions/${id}`, '_blank');
  };

  const handlePlayRecording = (id: string) => {
    window.open(`/sessions/recordings?session=${id}`, '_blank');
  };

  const getFilteredSessions = () => {
    switch (activeTab) {
      case 'active':
        return sessions.filter((s) => s.status === 'active');
      case 'idle':
        return sessions.filter((s) => s.status === 'idle');
      case 'all':
        return sessions;
      default:
        return sessions;
    }
  };

  const columns = [
    { key: 'user', header: 'User' },
    { key: 'host', header: 'Host' },
    {
      key: 'protocol',
      header: 'Protocol',
      render: (session: Session) => (
        <Badge variant="outline">{session.protocol.toUpperCase()}</Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (session: Session) => (
        <Badge
          variant={
            session.status === 'active'
              ? 'default'
              : session.status === 'idle'
              ? 'secondary'
              : 'destructive'
          }
        >
          {session.status}
        </Badge>
      ),
    },
    { key: 'startedAt', header: 'Started' },
    { key: 'duration', header: 'Duration' },
    { key: 'clientIP', header: 'Client IP' },
    {
      key: 'actions',
      header: 'Actions',
      render: (session: Session) => (
        <div className="flex gap-2">
          <Button variant="ghost" size="icon" onClick={() => handleViewDetails(session.id)}>
            <Eye className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => handlePlayRecording(session.id)}>
            <PlayCircle className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="text-red-500"
            onClick={() => handleKillSession(session.id)}
          >
            <Power className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold">Session Management</h1>

      <Card>
        <CardHeader>
          <CardTitle>Active Sessions</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="active">Active ({sessions.filter((s) => s.status === 'active').length})</TabsTrigger>
              <TabsTrigger value="idle">Idle ({sessions.filter((s) => s.status === 'idle').length})</TabsTrigger>
              <TabsTrigger value="all">All ({sessions.length})</TabsTrigger>
            </TabsList>
            <TabsContent value={activeTab}>
              <DataTable data={getFilteredSessions()} columns={columns} />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}