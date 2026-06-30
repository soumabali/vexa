"use client";

import React, { useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { CredentialCard } from "./credential-card";
import { CredentialForm } from "./credential-form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Search,
  Plus,
  Folder,
  Tag,
  Star,
  Filter,
  Grid3X3,
  List,
  SortAsc,
  Lock,
} from "lucide-react";

export type CredentialType = "ssh-key" | "password" | "api-key" | "certificate" | "note";

export interface Credential {
  id: string;
  name: string;
  type: CredentialType;
  username?: string;
  host?: string;
  port?: number;
  tags: string[];
  folder: string;
  isFavorite: boolean;
  lastUsed?: Date;
  createdAt: Date;
  updatedAt: Date;
  metadata?: Record<string, unknown>;
  privateKey?: string;
  password?: string;
  apiKey?: string;
  certificate?: string;
  note?: string;
}

interface CredentialListProps {
  credentials: Credential[];
  onAdd: (cred: Omit<Credential, "id" | "createdAt" | "updatedAt">) => void;
  onUpdate: (id: string, cred: Partial<Credential>) => void;
  onDelete: (id: string) => void;
  onFavorite: (id: string) => void;
}

const typeIcons: Record<CredentialType, React.ReactNode> = {
  "ssh-key": <Lock className="h-4 w-4" />,
  password: <Lock className="h-4 w-4" />,
  "api-key": <Lock className="h-4 w-4" />,
  certificate: <Lock className="h-4 w-4" />,
  note: <Lock className="h-4 w-4" />,
};

const typeColors: Record<CredentialType, string> = {
  "ssh-key": "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
  password: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
  "api-key": "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
  certificate: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  note: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
};

export function CredentialList({
  credentials,
  onAdd,
  onUpdate,
  onDelete,
  onFavorite,
}: CredentialListProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [sortBy, setSortBy] = useState<"name" | "date" | "type" | "favorite">("name");
  const [filterType, setFilterType] = useState<CredentialType | "all">("all");
  const [filterFolder, setFilterFolder] = useState<string | "all">("all");
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [showFavoritesOnly, setShowFavoritesOnly] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [selectedCredential, setSelectedCredential] = useState<Credential | null>(null);

  const folders = Array.from(new Set(credentials.map((c) => c.folder)));
  const allTags = Array.from(new Set(credentials.flatMap((c) => c.tags)));

  const filteredCredentials = credentials
    .filter((cred) => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesSearch =
          cred.name.toLowerCase().includes(query) ||
          cred.username?.toLowerCase().includes(query) ||
          cred.host?.toLowerCase().includes(query) ||
          cred.tags.some((t) => t.toLowerCase().includes(query));
        if (!matchesSearch) return false;
      }
      if (filterType !== "all" && cred.type !== filterType) return false;
      if (filterFolder !== "all" && cred.folder !== filterFolder) return false;
      if (selectedTag && !cred.tags.includes(selectedTag)) return false;
      if (showFavoritesOnly && !cred.isFavorite) return false;
      return true;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case "name":
          return a.name.localeCompare(b.name);
        case "date":
          return b.updatedAt.getTime() - a.updatedAt.getTime();
        case "type":
          return a.type.localeCompare(b.type);
        case "favorite":
          return (b.isFavorite ? 1 : 0) - (a.isFavorite ? 1 : 0);
        default:
          return 0;
      }
    });

  return (
    <div className="flex flex-col h-full gap-4 bg-surface-container rounded-xl p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
            <Input
              placeholder="Search credentials..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant={showFavoritesOnly ? "default" : "outline"}
              size="sm"
              onClick={() => setShowFavoritesOnly(!showFavoritesOnly)}
            >
              <Star className="h-4 w-4 mr-1" />
              Favorites
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Filter className="h-4 w-4 mr-1" />
                  Filter
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56 bg-surface-container-high border-outline-variant">
                <DropdownMenuItem onClick={() => setFilterType("all")}>
                  All Types
                </DropdownMenuItem>
                {(["ssh-key", "password", "api-key", "certificate", "note"] as CredentialType[]).map(
                  (type) => (
                    <DropdownMenuItem
                      key={type}
                      onClick={() => setFilterType(type)}
                    >
                      <span className="mr-2">{typeIcons[type]}</span>
                      {type.replace("-", " ").toUpperCase()}
                    </DropdownMenuItem>
                  )
                )}
                <DropdownMenuItem onClick={() => setFilterFolder("all")}>
                  All Folders
                </DropdownMenuItem>
                {folders.map((folder) => (
                  <DropdownMenuItem
                    key={folder}
                    onClick={() => setFilterFolder(folder)}
                  >
                    <Folder className="h-4 w-4 mr-2" />
                    {folder}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <SortAsc className="h-4 w-4 mr-1" />
                  Sort
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="bg-surface-container-high border-outline-variant">
                <DropdownMenuItem onClick={() => setSortBy("name")}>
                  Name
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSortBy("date")}>
                  Last Updated
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSortBy("type")}>
                  Type
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSortBy("favorite")}>
                  Favorites First
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <div className="flex items-center border border-outline-variant rounded-md">
              <Button
                variant={viewMode === "grid" ? "secondary" : "ghost"}
                size="sm"
                className="rounded-none rounded-l-md"
                onClick={() => setViewMode("grid")}
              >
                <Grid3X3 className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === "list" ? "secondary" : "ghost"}
                size="sm"
                className="rounded-none rounded-r-md"
                onClick={() => setViewMode("list")}
              >
                <List className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>

        <Button onClick={() => setIsAddDialogOpen(true)} className="bg-primary text-on-primary hover:bg-primary/90">
          <Plus className="h-4 w-4 mr-2" />
          Add Credential
        </Button>
      </div>

      {/* Tags */}
      {allTags.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <Tag className="h-4 w-4 text-on-surface-variant" />
          {allTags.map((tag) => (
            <Badge
              key={tag}
              variant={selectedTag === tag ? "default" : "outline"}
              className="cursor-pointer"
              onClick={() =>
                setSelectedTag(selectedTag === tag ? null : tag)
              }
            >
              {tag}
            </Badge>
          ))}
        </div>
      )}

      {/* Results Count */}
      <div className="text-sm text-on-surface-variant">
        {filteredCredentials.length} credential
        {filteredCredentials.length !== 1 ? "s" : ""}
        {searchQuery && ` matching "${searchQuery}"`}
      </div>

      {/* Credential Grid/List */}
      <ScrollArea className="flex-1">
        <div
          className={
            viewMode === "grid"
              ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
              : "flex flex-col gap-2"
          }
        >
          <AnimatePresence>
            {filteredCredentials.map((credential) => (
              <motion.div
                key={credential.id}
                layout
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                transition={{ duration: 0.2 }}
              >
                <CredentialCard
                  credential={credential}
                  onUpdate={onUpdate}
                  onDelete={onDelete}
                  onFavorite={onFavorite}
                  onClick={() => setSelectedCredential(credential)}
                  viewMode={viewMode}
                />
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </ScrollArea>

      {/* Add Dialog */}
      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto bg-surface-container-high border-outline-variant">
          <DialogHeader>
            <DialogTitle>Add New Credential</DialogTitle>
          </DialogHeader>
          <CredentialForm
            onSubmit={(data) => {
              onAdd(data);
              setIsAddDialogOpen(false);
            }}
            onCancel={() => setIsAddDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Detail Dialog */}
      {selectedCredential && (
        <CredentialDetailDialog
          credential={selectedCredential}
          onClose={() => setSelectedCredential(null)}
          onUpdate={onUpdate}
          onDelete={onDelete}
          onFavorite={onFavorite}
        />
      )}
    </div>
  );
}

function CredentialDetailDialog({
  credential,
  onClose,
  onUpdate,
  onDelete,
  onFavorite,
}: {
  credential: Credential;
  onClose: () => void;
  onUpdate: (id: string, data: Partial<Credential>) => void;
  onDelete: (id: string) => void;
  onFavorite: (id: string) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto bg-surface-container-high border-outline-variant">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {credential.name}
            <Badge className={typeColors[credential.type]}>
              {typeIcons[credential.type]}
              {credential.type.replace("-", " ").toUpperCase()}
            </Badge>
          </DialogTitle>
        </DialogHeader>

        {isEditing ? (
          <CredentialForm
            credential={credential}
            onSubmit={(data) => {
              onUpdate(credential.id, data);
              setIsEditing(false);
            }}
            onCancel={() => setIsEditing(false)}
          />
        ) : (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium text-on-surface-variant">
                  Username
                </label>
                <p>{credential.username || "—"}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-on-surface-variant">
                  Host
                </label>
                <p>{credential.host || "—"}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-on-surface-variant">
                  Port
                </label>
                <p>{credential.port || "—"}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-on-surface-variant">
                  Folder
                </label>
                <p>{credential.folder}</p>
              </div>
            </div>

            <div>
              <label className="text-sm font-medium text-on-surface-variant">
                Tags
              </label>
              <div className="flex flex-wrap gap-2 mt-1">
                {credential.tags.map((tag) => (
                  <Badge key={tag} variant="outline">
                    {tag}
                  </Badge>
                ))}
              </div>
            </div>

            {credential.lastUsed && (
              <div>
                <label className="text-sm font-medium text-on-surface-variant">
                  Last Used
                </label>
                <p>{credential.lastUsed.toLocaleString()}</p>
              </div>
            )}

            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                className="border-outline-variant"
                onClick={() => onFavorite(credential.id)}
              >
                <Star
                  className={`h-4 w-4 mr-2 ${
                    credential.isFavorite ? "fill-yellow-400" : ""
                  }`}
                />
                {credential.isFavorite ? "Unfavorite" : "Favorite"}
              </Button>
              <Button variant="outline" className="border-outline-variant" onClick={() => setIsEditing(true)}>
                Edit
              </Button>
              <Button
                variant="destructive"
                className="bg-error text-on-error hover:bg-error/90"
                onClick={() => {
                  onDelete(credential.id);
                  onClose();
                }}
              >
                Delete
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}