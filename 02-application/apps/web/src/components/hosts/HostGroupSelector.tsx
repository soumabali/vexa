"use client";

import React, { useState } from "react";
import { hostsApi } from "@/lib/api/hosts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { LoadingSpinner } from "@/components/auth/LoadingSpinner";
import { ErrorDisplay } from "@/components/auth/ErrorDisplay";
import {
  PlusIcon,
  PencilIcon,
  TrashIcon,
  FolderIcon,
  CheckIcon,
  XIcon,
} from "lucide-react";

interface HostGroup {
  id: string;
  name: string;
  color?: string;
  count: number;
}

interface HostGroupSelectorProps {
  groups: HostGroup[];
  onGroupsChange: (groups: HostGroup[]) => void;
}

const colorPresets = [
  "#EF4444", "#F97316", "#EAB308", "#22C55E",
  "#14B8A6", "#3B82F6", "#8B5CF6", "#EC4899",
];

export function HostGroupSelector({ groups, onGroupsChange }: HostGroupSelectorProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [editingGroup, setEditingGroup] = useState<HostGroup | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newGroupName, setNewGroupName] = useState("");
  const [newGroupColor, setNewGroupColor] = useState(colorPresets[5]);
  const [deletingGroupId, setDeletingGroupId] = useState<string | null>(null);

  const handleCreateGroup = async () => {
    if (!newGroupName.trim()) return;
    setIsLoading(true);
    setError("");
    try {
      const created = await hostsApi.createGroup({
        name: newGroupName.trim(),
        color: newGroupColor,
      });
      onGroupsChange([...groups, { ...created, count: 0 }]);
      setNewGroupName("");
      setNewGroupColor(colorPresets[5]);
      setShowCreateDialog(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create group");
    } finally {
      setIsLoading(false);
    }
  };

  const handleUpdateGroup = async () => {
    if (!editingGroup || !newGroupName.trim()) return;
    setIsLoading(true);
    setError("");
    try {
      const updated = await hostsApi.updateGroup(editingGroup.id, {
        name: newGroupName.trim(),
        color: newGroupColor,
      });
      onGroupsChange(
        groups.map((g) =>
          g.id === editingGroup.id ? { ...updated, count: g.count } : g
        )
      );
      setEditingGroup(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update group");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteGroup = async (group: HostGroup) => {
    setDeletingGroupId(group.id);
    try {
      await hostsApi.deleteGroup(group.id);
      onGroupsChange(groups.filter((g) => g.id !== group.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete group");
    } finally {
      setDeletingGroupId(null);
    }
  };

  const startEdit = (group: HostGroup) => {
    setEditingGroup(group);
    setNewGroupName(group.name);
    setNewGroupColor(group.color || colorPresets[5]);
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {groups.length} group{groups.length !== 1 ? "s" : ""}
        </p>
        <Button size="sm" onClick={() => setShowCreateDialog(true)}>
          <PlusIcon className="h-4 w-4 mr-1" />
          New Group
        </Button>
      </div>

      {/* Error */}
      {error && <ErrorDisplay message={error} />}

      {/* Group list */}
      {groups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-8 text-center">
          <FolderIcon className="h-10 w-10 text-muted-foreground/50 mb-3" />
          <p className="text-sm text-muted-foreground mb-3">No groups yet</p>
          <Button size="sm" variant="outline" onClick={() => setShowCreateDialog(true)}>
            <PlusIcon className="h-4 w-4 mr-1" />
            Create Group
          </Button>
        </div>
      ) : (
        <div className="space-y-2 max-h-[300px] overflow-y-auto">
          {groups.map((group) => (
            <div
              key={group.id}
              className="flex items-center justify-between rounded-lg border bg-card p-3 hover:bg-accent/50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div
                  className="h-3 w-3 rounded-full"
                  style={{ backgroundColor: group.color || "#6B7280" }}
                />
                <div>
                  <p className="font-medium text-sm">{group.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {group.count} host{group.count !== 1 ? "s" : ""}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  onClick={() => startEdit(group)}
                >
                  <PencilIcon className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-destructive hover:text-destructive"
                  onClick={() => handleDeleteGroup(group)}
                  disabled={deletingGroupId === group.id}
                >
                  {deletingGroupId === group.id ? (
                    <LoadingSpinner className="h-3.5 w-3.5" />
                  ) : (
                    <TrashIcon className="h-3.5 w-3.5" />
                  )}
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Group Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Group</DialogTitle>
            <DialogDescription>
              Give your group a name and choose a color
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="group-name">Name</Label>
              <Input
                id="group-name"
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                placeholder="Production Servers"
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label>Color</Label>
              <div className="flex items-center gap-2">
                <div
                  className="h-8 w-8 rounded border"
                  style={{ backgroundColor: newGroupColor }}
                />
                <div className="flex gap-1">
                  {colorPresets.map((color) => (
                    <button
                      key={color}
                      type="button"
                      onClick={() => setNewGroupColor(color)}
                      className="h-6 w-6 rounded border-2 transition-all"
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateGroup} disabled={isLoading || !newGroupName.trim()}>
              {isLoading ? <LoadingSpinner className="h-4 w-4 mr-2" /> : null}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Group Dialog */}
      <Dialog
        open={!!editingGroup}
        onOpenChange={(open) => !open && setEditingGroup(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Group</DialogTitle>
            <DialogDescription>
              Update group name and color
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="edit-group-name">Name</Label>
              <Input
                id="edit-group-name"
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                placeholder="Production Servers"
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label>Color</Label>
              <div className="flex items-center gap-2">
                <div
                  className="h-8 w-8 rounded border"
                  style={{ backgroundColor: newGroupColor }}
                />
                <div className="flex gap-1">
                  {colorPresets.map((color) => (
                    <button
                      key={color}
                      type="button"
                      onClick={() => setNewGroupColor(color)}
                      className="h-6 w-6 rounded border-2 transition-all"
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingGroup(null)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateGroup} disabled={isLoading || !newGroupName.trim()}>
              {isLoading ? <LoadingSpinner className="h-4 w-4 mr-2" /> : null}
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}