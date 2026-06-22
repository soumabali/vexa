"use client";

import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { HostResponse } from "@/lib/validations/hosts";
import { HostForm } from "./HostForm";
import { hostsApi } from "@/lib/api/hosts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import {
  SearchIcon,
  PlusIcon,
  ServerIcon,
  MonitorIcon,
  LayoutGridIcon,
  ListIcon,
  TrashIcon,
  PencilIcon,
  MoreHorizontalIcon,
  StarIcon,
  StarOffIcon,
} from "lucide-react";

const hostTypeIcons = {
  ssh: ServerIcon,
  rdp: MonitorIcon,
  vnc: MonitorIcon,
};

const hostTypeColors = {
  ssh: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  rdp: "bg-green-500/20 text-green-400 border-green-500/30",
  vnc: "bg-purple-500/20 text-purple-400 border-purple-500/30",
};

const statusColors = {
  online: "bg-green-500",
  offline: "bg-red-500",
  unknown: "bg-gray-500",
};

export function HostList() {
  const [actionError, setActionError] = useState("");
  const [search, setSearch] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("list");
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [hostToDelete, setHostToDelete] = useState<HostResponse | null>(null);
  const [editingHost, setEditingHost] = useState<HostResponse | null>(null);

  const {
    data: hosts = [],
    isLoading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: ["hosts"],
    queryFn: async () => {
      const result = await hostsApi.list();
      return result.hosts || [];
    },
    staleTime: 0,
    refetchOnWindowFocus: false,
  });

  const filteredHosts = search.trim()
    ? hosts.filter((h) => {
        const term = search.toLowerCase();
        return (
          h.name.toLowerCase().includes(term) ||
          h.host.toLowerCase().includes(term) ||
          h.hostType.toLowerCase().includes(term)
        );
      })
    : hosts;

  const handleDelete = async (host: HostResponse) => {
    try {
      await hostsApi.delete(host.id);
      await refetch();
      setHostToDelete(null);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to delete host");
    }
  };

  const handleToggleFavorite = async (host: HostResponse) => {
    try {
      await hostsApi.toggleFavorite(host.id);
      await refetch();
    } catch (err) {
      console.error("Failed to toggle favorite:", err);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Hosts</h1>
          <p className="text-sm text-muted-foreground">
            Manage your SSH, RDP, and VNC connections
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setViewMode(viewMode === "grid" ? "list" : "grid")}
            className="h-9"
          >
            {viewMode === "grid" ? (
              <ListIcon className="h-4 w-4" />
            ) : (
              <LayoutGridIcon className="h-4 w-4" />
            )}
          </Button>
          <Button onClick={() => setShowCreateDialog(true)} className="h-9">
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Host
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search hosts..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10 bg-surface-container border-outline"
          />
        </div>
        <div className="flex gap-2">
          {["ssh", "rdp", "vnc"].map((type) => (
            <Button
              key={type}
              variant="outline"
              size="sm"
              className={`h-9 capitalize ${hostTypeColors[type as keyof typeof hostTypeColors]}`}
            >
              {type}
            </Button>
          ))}
        </div>
      </div>

      {/* Error */}
      {(actionError || (queryError && queryError.message)) && (
        <ErrorDisplay
          message={actionError || (queryError instanceof Error ? queryError.message : "Failed to fetch hosts")}
        />
      )}

      {/* Hosts Grid/List */}
      {filteredHosts.length === 0 ? (
        <Card className="border-dashed border-2 border-outline bg-surface-container">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <ServerIcon className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium mb-2">No hosts found</p>
            <p className="text-sm text-muted-foreground mb-4">
              Add your first host to get started
            </p>
            <Button onClick={() => setShowCreateDialog(true)}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Host
            </Button>
          </CardContent>
        </Card>
      ) : viewMode === "grid" ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredHosts.map((host) => (
            <HostCard
              key={host.id}
              host={host}
              onDelete={setHostToDelete}
              onToggleFavorite={handleToggleFavorite}
              onEdit={(host) => setEditingHost(host)}
            />
          ))}
        </div>
      ) : (
        <Card className="bg-surface-container border-outline">
          <CardContent className="p-0">
            <div className="divide-y divide-outline-variant">
              {filteredHosts.map((host) => (
                <HostRow
                  key={host.id}
                  host={host}
                  onDelete={setHostToDelete}
                  onToggleFavorite={handleToggleFavorite}
                  onEdit={(host) => setEditingHost(host)}
                />
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Create / Edit Host Dialog */}
      <Dialog
        open={showCreateDialog || !!editingHost}
        onOpenChange={(open) => {
          if (!open) {
            setShowCreateDialog(false);
            setEditingHost(null);
          }
        }}
      >
        <DialogContent className="bg-surface-container border-outline max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingHost ? 'Edit Host' : 'Add New Host'}</DialogTitle>
            <DialogDescription>
              Configure a new SSH, RDP, or VNC connection.
            </DialogDescription>
          </DialogHeader>
          <HostForm
            host={editingHost ?? undefined}
            onSuccess={async () => {
              await refetch();
              setShowCreateDialog(false);
              setEditingHost(null);
            }}
            onCancel={() => {
              setShowCreateDialog(false);
              setEditingHost(null);
            }}
          />
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={!!hostToDelete} onOpenChange={() => setHostToDelete(null)}>
        <DialogContent className="bg-surface-container border-outline">
          <DialogHeader>
            <DialogTitle className="text-destructive">Delete Host</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-semibold">{hostToDelete?.name}</span>? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setHostToDelete(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => hostToDelete && handleDelete(hostToDelete)}
            >
              <TrashIcon className="h-4 w-4 mr-2" />
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Host Card (Grid View)
function HostCard({
  host,
  onDelete,
  onToggleFavorite,
  onEdit,
}: {
  host: HostResponse;
  onDelete: (host: HostResponse) => void;
  onToggleFavorite: (host: HostResponse) => void;
  onEdit: (host: HostResponse) => void;
}) {
  const Icon = hostTypeIcons[host.hostType as keyof typeof hostTypeIcons] || ServerIcon;

  return (
    <Card className="bg-surface-container border-outline hover:border-primary/50 transition-colors group" data-testid={`host-card-${host.id}`} data-host-id={host.id} data-host-name={host.name}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-lg ${hostTypeColors[host.hostType as keyof typeof hostTypeColors]}`}>
              <Icon className="h-5 w-5" />
            </div>
            <div>
              <CardTitle className="text-base font-semibold">{host.name}</CardTitle>
              <p className="text-xs text-muted-foreground font-mono">
                {host.host}:{host.port}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onToggleFavorite(host)}
            >
              {host.favorite ? (
                <StarIcon className="h-4 w-4 text-yellow-400" />
              ) : (
                <StarOffIcon className="h-4 w-4" />
              )}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8" data-testid={`host-actions-menu-${host.id}`}>
                  <MoreHorizontalIcon className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onEdit(host)}>
                  <PencilIcon className="h-4 w-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDelete(host)} className="text-destructive focus:text-destructive">
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${statusColors[host.status as keyof typeof statusColors] || statusColors.unknown}`} />
            <span className="text-xs capitalize text-muted-foreground">
              {host.status}
            </span>
          </div>
          <Badge variant="outline" className="text-xs capitalize">
            {host.hostType}
          </Badge>
        </div>
      </CardContent>
    </Card>
  );
}

// Host Row (List View)
function HostRow({
  host,
  onDelete,
  onToggleFavorite,
  onEdit,
}: {
  host: HostResponse;
  onDelete: (host: HostResponse) => void;
  onToggleFavorite: (host: HostResponse) => void;
  onEdit: (host: HostResponse) => void;
}) {
  const Icon = hostTypeIcons[host.hostType as keyof typeof hostTypeIcons] || ServerIcon;

  return (
    <div className="flex items-center justify-between p-4 hover:bg-surface-container-high transition-colors group" data-testid={`host-row-${host.id}`} data-host-id={host.id} data-host-name={host.name}>
      <div className="flex items-center gap-4">
        <div className={`p-2 rounded-lg ${hostTypeColors[host.hostType as keyof typeof hostTypeColors]}`}>
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <p className="font-semibold">{host.name}</p>
          <p className="text-sm text-muted-foreground font-mono">
            {host.host}:{host.port}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${statusColors[host.status as keyof typeof statusColors] || statusColors.unknown}`} />
          <span className="text-sm capitalize text-muted-foreground">
            {host.status}
          </span>
        </div>
        <Badge variant="outline" className="capitalize">
          {host.hostType}
        </Badge>
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => onToggleFavorite(host)}
          >
            {host.favorite ? (
              <StarIcon className="h-4 w-4 text-yellow-400" />
            ) : (
              <StarOffIcon className="h-4 w-4" />
            )}
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8" data-testid={`host-actions-menu-${host.id}`}>
                <MoreHorizontalIcon className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit(host)}>
                <PencilIcon className="h-4 w-4 mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDelete(host)} className="text-destructive focus:text-destructive">
                <TrashIcon className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  );
}
