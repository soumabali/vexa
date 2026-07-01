'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { FileManager } from '@/components/file-manager';
import { useToast } from '@/hooks/use-toast';
import { useAuth } from "@/hooks/use-auth";
import { useAsyncData } from '@/hooks/useAsyncData';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { MaterialIcon } from "@/components/ui/material-icon";
interface FileItem {
  name: string;
  path: string;
  size: number;
  modified: string;
  permissions: string;
  owner: string;
  group: string;
  type: 'file' | 'directory';
  isSymlink?: boolean;
  target?: string;
  mimeType?: string;
}

interface FileSystem {
  name: string;
  host: string;
  port: number;
  username: string;
  path: string;
  protocol: 'sftp' | 'scp';
}

interface FileManagerPageProps {
  initialHostId?: string;
}

export default function FileManagerPage({ initialHostId }: FileManagerPageProps) {
  const { toast } = useToast();
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('local');
  const [activeHost, setActiveHost] = useState<FileSystem | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [newHostDialog, setNewHostDialog] = useState(false);
  const [transferProgress, setTransferProgress] = useState<{
    active: boolean;
    file: string;
    progress: number;
    speed: string;
    eta: string;
  } | null>(null);

  const { data: hostsData, loading: isLoading, reload: refetchHosts } = useAsyncData<{ hosts: FileSystem[]; initialHost: FileSystem | null }>(async () => {
    const response = await fetch('/api/files/hosts');
    const data = await response.json();
    const hosts = data.hosts || [];
    let initialHost: FileSystem | null = null;
    if (initialHostId) {
      initialHost = hosts.find((h: FileSystem) => h.host === initialHostId) || null;
    }
    return { hosts, initialHost };
  });
  const [hosts, setHosts] = useState<FileSystem[]>(hostsData?.hosts ?? []);

  const connectToHost = async (host: FileSystem) => {
    setIsConnecting(true);
    try {
      const response = await fetch('/api/files/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          host: host.host,
          port: host.port,
          username: host.username,
          protocol: host.protocol,
        }),
      });

      if (!response.ok) {
        throw new Error('Connection failed');
      }

      setActiveHost(host);
      setActiveTab('remote');
      toast({
        title: 'Connected',
        description: `Connected to ${host.name}`,
      });
    } catch (error) {
      toast({
        title: 'Connection Failed',
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setIsConnecting(false);
    }
  };

  const handleTransferProgress = useCallback((progress: {
    file: string;
    progress: number;
    speed: string;
    eta: string;
  }) => {
    setTransferProgress({
      active: true,
      ...progress,
    });
  }, []);

  const handleTransferComplete = useCallback(() => {
    setTransferProgress(null);
    toast({
      title: 'Transfer Complete',
      description: 'File transfer completed successfully',
    });
  }, []);

  return (
    <div className="container mx-auto p-4 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">File Manager</h1>
          <p className="text-on-surface-variant mt-1">
            Manage files across local and remote systems
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => setNewHostDialog(true)}
          >
            <MaterialIcon name="add" className="h-4 w-4 mr-2" />
            Add Host
          </Button>
          <Button variant="outline" onClick={refetchHosts}>
            <MaterialIcon name="refresh" className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      <Separator />

      {/* Transfer Progress */}
      {transferProgress?.active && (
        <Card className="border-blue-200 bg-blue-50">
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <MaterialIcon name="progress_activity" className="h-4 w-4 animate-spin text-blue-500" />
                <span className="font-medium">{transferProgress.file}</span>
              </div>
              <span className="text-sm text-on-surface-variant">
                {transferProgress.speed} • ETA {transferProgress.eta}
              </span>
            </div>
            <div className="w-full bg-blue-200 rounded-full h-2">
              <div
                className="bg-blue-500 h-2 rounded-full transition-all"
                style={{ width: `${transferProgress.progress}%` }}
              />
            </div>
          </CardContent>
        </Card>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="local" className="flex items-center gap-2">
            <MaterialIcon name="storage" className="h-4 w-4" />
            Local
          </TabsTrigger>
          <TabsTrigger value="remote" className="flex items-center gap-2">
            <MaterialIcon name="cloud" className="h-4 w-4" />
            Remote
            {activeHost && (
              <Badge variant="secondary" className="ml-1">
                {activeHost.name}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="dual" className="flex items-center gap-2">
            <MaterialIcon name="content_copy" className="h-4 w-4" />
            Dual Pane
          </TabsTrigger>
        </TabsList>

        <TabsContent value="local" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MaterialIcon name="storage" className="h-5 w-5" />
                Local Files
              </CardTitle>
              <CardDescription>
                Browse and manage files on your local machine
              </CardDescription>
            </CardHeader>
            <CardContent>
              <FileManager
                type="local"
                host={null}
                onTransferProgress={handleTransferProgress}
                onTransferComplete={handleTransferComplete}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="remote" className="space-y-4">
          {!activeHost ? (
            <Card>
              <CardHeader>
                <CardTitle>Connect to Remote Host</CardTitle>
                <CardDescription>
                  Select a host from the list below to connect
                </CardDescription>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="space-y-2">
                    <Skeleton className="h-12 w-full" />
                    <Skeleton className="h-12 w-full" />
                    <Skeleton className="h-12 w-full" />
                  </div>
                ) : hosts.length === 0 ? (
                  <div className="text-center py-8">
                    <MaterialIcon name="dns" className="h-12 w-12 mx-auto text-on-surface-variant mb-4" />
                    <p className="text-on-surface-variant">No hosts configured</p>
                    <Button
                      variant="outline"
                      className="mt-4"
                      onClick={() => setNewHostDialog(true)}
                    >
                      <MaterialIcon name="add" className="h-4 w-4 mr-2" />
                      Add Host
                    </Button>
                  </div>
                ) : (
                  <div className="grid gap-2">
                    {hosts.map((host) => (
                      <Card
                        key={host.host}
                        className="cursor-pointer hover:bg-primary-container transition-colors"
                        onClick={() => connectToHost(host)}
                      >
                        <CardContent className="p-4">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                                <MaterialIcon name="dns" className="h-5 w-5 text-primary" />
                              </div>
                              <div>
                                <p className="font-medium">{host.name}</p>
                                <p className="text-sm text-on-surface-variant">
                                  {host.username}@{host.host}:{host.port}
                                </p>
                              </div>
                            </div>
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={isConnecting}
                            >
                              {isConnecting ? (
                                <MaterialIcon name="progress_activity" className="h-4 w-4 animate-spin" />
                              ) : (
                                <MaterialIcon name="chevron_right" className="h-4 w-4" />
                              )}
                            </Button>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <MaterialIcon name="cloud" className="h-5 w-5" />
                    {activeHost.name}
                  </CardTitle>
                  <CardDescription>
                    {activeHost.username}@{activeHost.host}:{activeHost.port}
                  </CardDescription>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="flex items-center gap-1">
                    <MaterialIcon name="wifi" className="h-3 w-3 text-green-500" />
                    Connected
                  </Badge>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setActiveHost(null)}
                  >
                    <MaterialIcon name="close" className="h-4 w-4" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <FileManager
                  type="remote"
                  host={activeHost}
                  onTransferProgress={handleTransferProgress}
                  onTransferComplete={handleTransferComplete}
                />
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="dual" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MaterialIcon name="content_copy" className="h-5 w-5" />
                Dual Pane Mode
              </CardTitle>
              <CardDescription>
                Compare and transfer files between two locations
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                <div className="border rounded-lg p-4">
                  <h3 className="font-medium mb-2 flex items-center gap-2">
                    <MaterialIcon name="storage" className="h-4 w-4" />
                    Source
                  </h3>
                  <FileManager
                    type="local"
                    host={null}
                    compact
                    onTransferProgress={handleTransferProgress}
                    onTransferComplete={handleTransferComplete}
                  />
                </div>
                <div className="border rounded-lg p-4">
                  <h3 className="font-medium mb-2 flex items-center gap-2">
                    <MaterialIcon name="cloud" className="h-4 w-4" />
                    Destination
                  </h3>
                  {activeHost ? (
                    <FileManager
                      type="remote"
                      host={activeHost}
                      compact
                      onTransferProgress={handleTransferProgress}
                      onTransferComplete={handleTransferComplete}
                    />
                  ) : (
                    <div className="text-center py-8 text-on-surface-variant">
                      <p>Connect to a remote host first</p>
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* New Host Dialog */}
      <Dialog open={newHostDialog} onOpenChange={setNewHostDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add New Host</DialogTitle>
            <DialogDescription>
              Configure a new remote host connection
            </DialogDescription>
          </DialogHeader>
          <NewHostForm
            onSubmit={async (host) => {
              await fetch('/api/files/hosts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(host),
              });
              await refetchHosts();
              setNewHostDialog(false);
            }}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function NewHostForm({ onSubmit }: { onSubmit: (host: FileSystem) => void }) {
  const [host, setHost] = useState<Partial<FileSystem>>({
    port: 22,
    protocol: 'sftp',
  });

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          placeholder="My Server"
          value={host.name || ''}
          onChange={(e) => setHost({ ...host, name: e.target.value })}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="host">Host</Label>
        <Input
          id="host"
          placeholder="example.com"
          value={host.host || ''}
          onChange={(e) => setHost({ ...host, host: e.target.value })}
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="port">Port</Label>
          <Input
            id="port"
            type="number"
            value={host.port}
            onChange={(e) => setHost({ ...host, port: parseInt(e.target.value) })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="protocol">Protocol</Label>
          <Select
            value={host.protocol}
            onValueChange={(v: 'sftp' | 'scp') => setHost({ ...host, protocol: v })}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="sftp">SFTP</SelectItem>
              <SelectItem value="scp">SCP</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="username">Username</Label>
        <Input
          id="username"
          value={host.username || ''}
          onChange={(e) => setHost({ ...host, username: e.target.value })}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="path">Initial Path</Label>
        <Input
          id="path"
          placeholder="/home/user"
          value={host.path || ''}
          onChange={(e) => setHost({ ...host, path: e.target.value })}
        />
      </div>
      <Button
        className="w-full"
        onClick={() => onSubmit(host as FileSystem)}
        disabled={!host.name || !host.host || !host.username}
      >
        Add Host
      </Button>
    </div>
  );
}
