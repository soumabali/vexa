"use client";

import { useState } from "react";
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
import { MaterialIcon } from "@/components/ui/material-icon";
import { Skeleton } from "@/components/ui/skeleton";

const hostTypeIcons = {
  ssh: "server",
  rdp: "monitor",
  vnc: "monitor",
};

const hostTypeColors = {
  ssh: "bg-secondary-container text-on-secondary-container",
  rdp: "bg-tertiary-container text-on-tertiary-container",
  vnc: "bg-primary-container text-on-primary-container",
};

const statusColors = {
  online: "bg-success",
  offline: "bg-error",
  unknown: "bg-outline-variant",
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
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <Skeleton className="h-8 w-32" />
            <Skeleton className="h-4 w-64" />
          </div>
          <Skeleton className="h-10 w-32 rounded-md" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-outline-variant bg-surface p-4 space-y-3">
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
              <Skeleton className="h-3 w-full" />
              <Skeleton className="h-3 w-2/3" />
              <div className="flex gap-2">
                <Skeleton className="h-8 w-20 rounded-md" />
                <Skeleton className="h-8 w-20 rounded-md" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-headline-lg text-on-surface">Hosts</h1>
          <p className="text-body-md text-on-surface-variant">
            Securely manage remote environments across SSH, RDP, VNC.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setViewMode(viewMode === "grid" ? "list" : "grid")}
            className="h-9 border-outline-variant"
            aria-label="Toggle view mode"
          >
            <MaterialIcon name={viewMode === "grid" ? "list" : "grid_view"} size="sm" />
          </Button>
          <Button onClick={() => setShowCreateDialog(true)} className="h-9 bg-primary text-on-primary rounded-lg">
            <MaterialIcon name="add" size="sm" className="mr-2" />
            Add Host
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex items-center gap-4 flex-wrap">
        <div className="relative flex-1 min-w-[240px] max-w-md">
          <MaterialIcon name="search" size="sm" className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
          <Input
            placeholder="Search hosts..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10 bg-surface-container-low border border-outline-variant"
          />
        </div>
        <div className="flex gap-2 items-center">
          <button
            className="h-9 px-4 rounded-full text-label-md bg-secondary-container text-on-secondary-container"
          >
            All
          </button>
          {["ssh", "rdp", "vnc"].map((type) => (
            <button
              key={type}
              className={`h-9 px-4 rounded-full text-label-md capitalize border border-outline-variant text-on-surface-variant hover:${hostTypeColors[type as keyof typeof hostTypeColors]}`}
            >
              {type.toUpperCase()}
            </button>
          ))}
          <Button
            variant="outline"
            size="sm"
            className="h-9 border-outline-variant ml-2"
          >
            <MaterialIcon name="tune" size="sm" className="mr-2" />
            Filters
          </Button>
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
        <Card className="border-dashed border-2 border-outline-variant bg-surface-container">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <MaterialIcon name="server" size="xl" className="text-on-surface-variant mb-4" />
            <p className="text-headline-sm text-on-surface mb-2">No hosts found</p>
            <p className="text-body-md text-on-surface-variant mb-4">
              Add your first host to get started
            </p>
            <Button onClick={() => setShowCreateDialog(true)} className="bg-primary text-on-primary rounded-lg">
              <MaterialIcon name="add" size="sm" className="mr-2" />
              Add Host
            </Button>
          </CardContent>
        </Card>
      ) : viewMode === "grid" ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredHosts.map((host) => (
            <HostCard
              key={host.id}
              host={host}
              onDelete={setHostToDelete}
              onToggleFavorite={handleToggleFavorite}
              onEdit={(host) => setEditingHost(host)}
            />
          ))}
          <button
            onClick={() => setShowCreateDialog(true)}
            className="border-2 border-dashed border-outline-variant hover:border-primary rounded-xl min-h-[160px] flex flex-col items-center justify-center gap-2 text-on-surface-variant hover:text-primary transition-colors"
          >
            <MaterialIcon name="add" size="lg" />
            <span className="text-label-md">Add Host</span>
          </button>
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

      {/* Footer stats */}
      <div className="border-t border-outline-variant pt-4 flex items-center justify-between text-label-md text-on-surface-variant">
        <span>Total Hosts: {filteredHosts.length}</span>
        <span>Active Sessions: {hosts.filter((h) => h.status === "online").length}</span>
        <span className="text-mono-code">v2.4.1-stable</span>
      </div>

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
            <DialogTitle className="text-error">Delete Host</DialogTitle>
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
              <MaterialIcon name="delete" size="sm" className="mr-2" />
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
  const iconName = hostTypeIcons[host.hostType as keyof typeof hostTypeIcons] || "server";

  return (
    <Card className="bg-surface-container border-outline hover:border-primary/50 transition-colors group" data-testid={`host-card-${host.id}`} data-host-id={host.id} data-host-name={host.name}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-lg ${hostTypeColors[host.hostType as keyof typeof hostTypeColors]}`}>
              <MaterialIcon name={iconName} size="sm" />
            </div>
            <div>
              <CardTitle className="text-base font-semibold text-on-surface">{host.name}</CardTitle>
              <p className="text-xs text-on-surface-variant font-mono-code">
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
                <MaterialIcon name="star" size="sm" fill className="text-warning" />
              ) : (
                <MaterialIcon name="star_border" size="sm" className="text-on-surface-variant" />
              )}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8" data-testid={`host-actions-menu-${host.id}`}>
                  <MaterialIcon name="more_horiz" size="sm" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onEdit(host)}>
                  <MaterialIcon name="edit" size="sm" className="mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDelete(host)} className="text-error focus:text-error">
                  <MaterialIcon name="delete" size="sm" className="mr-2" />
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
            <span className="text-xs capitalize text-on-surface-variant">
              {host.status}
            </span>
          </div>
          <Badge variant="outline" className="text-xs capitalize border-outline-variant text-on-surface-variant">
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
  const iconName = hostTypeIcons[host.hostType as keyof typeof hostTypeIcons] || "server";

  return (
    <div className="flex items-center justify-between p-4 hover:bg-surface-container-high transition-colors group" data-testid={`host-row-${host.id}`} data-host-id={host.id} data-host-name={host.name}>
      <div className="flex items-center gap-4">
        <div className={`p-2 rounded-lg ${hostTypeColors[host.hostType as keyof typeof hostTypeColors]}`}>
          <MaterialIcon name={iconName} size="sm" />
        </div>
        <div>
          <p className="font-semibold text-on-surface">{host.name}</p>
          <p className="text-sm text-on-surface-variant font-mono-code">
            {host.host}:{host.port}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${statusColors[host.status as keyof typeof statusColors] || statusColors.unknown}`} />
          <span className="text-sm capitalize text-on-surface-variant">
            {host.status}
          </span>
        </div>
        <Badge variant="outline" className="capitalize border-outline-variant text-on-surface-variant">
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
              <MaterialIcon name="star" size="sm" fill className="text-warning" />
            ) : (
              <MaterialIcon name="star_border" size="sm" className="text-on-surface-variant" />
            )}
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8" data-testid={`host-actions-menu-${host.id}`}>
                <MaterialIcon name="more_horiz" size="sm" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit(host)}>
                <MaterialIcon name="edit" size="sm" className="mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDelete(host)} className="text-error focus:text-error">
                <MaterialIcon name="delete" size="sm" className="mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  );
}
