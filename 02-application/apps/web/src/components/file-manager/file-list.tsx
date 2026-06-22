'use client';

import React, { useState } from 'react';
import { Checkbox } from '@/components/ui/checkbox';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
} from '@/components/ui/context-menu';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Folder,
  File,
  FileCode,
  FileText,
  FileImage,
  FileVideo,
  FileAudio,
  FileArchive,
  FileX,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Download,
  Pencil,
  Trash2,
  Star,
  StarOff,
  Copy,
  Scissors,
  ClipboardPaste,
  Info,
} from 'lucide-react';

interface FileItem {
  id: string;
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

interface FileListProps {
  files: FileItem[];
  selectedFiles: Set<string>;
  onSelect: (path: string, multi: boolean) => void;
  onFileClick: (file: FileItem) => void;
  onDownload: (file: FileItem) => void;
  onRename: (oldPath: string, newName: string) => void;
  onChmod: (path: string, permissions: string) => void;
  onBookmarkToggle: (path: string) => void;
  bookmarks: string[];
  sortBy: 'name' | 'size' | 'date' | 'type';
  sortOrder: 'asc' | 'desc';
  onSortChange: (sortBy: 'name' | 'size' | 'date' | 'type') => void;
  onSortOrderChange: () => void;
}

export function FileList({
  files,
  selectedFiles,
  onSelect,
  onFileClick,
  onDownload,
  onRename,
  onChmod,
  onBookmarkToggle,
  bookmarks,
  sortBy,
  sortOrder,
  onSortChange,
  onSortOrderChange,
}: FileListProps) {
  const [renameDialog, setRenameDialog] = useState<{ file: FileItem; newName: string } | null>(null);
  const [chmodDialog, setChmodDialog] = useState<{ file: FileItem; permissions: string } | null>(null);
  const [infoDialog, setInfoDialog] = useState<FileItem | null>(null);

  const getFileIcon = (file: FileItem) => {
    if (file.type === 'directory') {
      return <Folder className="h-4 w-4 text-blue-500" />;
    }
    if (file.isSymlink) {
      return <FileX className="h-4 w-4 text-purple-500" />;
    }
    const ext = file.name.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'js':
      case 'ts':
      case 'jsx':
      case 'tsx':
      case 'go':
      case 'rs':
      case 'py':
      case 'java':
      case 'cpp':
      case 'c':
        return <FileCode className="h-4 w-4 text-yellow-500" />;
      case 'txt':
      case 'md':
      case 'log':
      case 'json':
      case 'yaml':
      case 'yml':
      case 'xml':
        return <FileText className="h-4 w-4 text-gray-500" />;
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'svg':
      case 'webp':
        return <FileImage className="h-4 w-4 text-green-500" />;
      case 'mp4':
      case 'avi':
      case 'mov':
      case 'mkv':
      case 'webm':
        return <FileVideo className="h-4 w-4 text-red-500" />;
      case 'mp3':
      case 'wav':
      case 'flac':
      case 'aac':
      case 'ogg':
        return <FileAudio className="h-4 w-4 text-pink-500" />;
      case 'zip':
      case 'tar':
      case 'gz':
      case 'bz2':
      case '7z':
      case 'rar':
        return <FileArchive className="h-4 w-4 text-orange-500" />;
      default:
        return <File className="h-4 w-4 text-gray-400" />;
    }
  };

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '-';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const formatDate = (date: string): string => {
    return new Date(date).toLocaleString();
  };

  const getSortIcon = (column: 'name' | 'size' | 'date' | 'type') => {
    if (sortBy !== column) return <ArrowUpDown className="h-3 w-3" />;
    return sortOrder === 'asc' ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />;
  };

  const handleHeaderClick = (column: 'name' | 'size' | 'date' | 'type') => {
    if (sortBy === column) {
      onSortOrderChange();
    } else {
      onSortChange(column);
    }
  };

  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-8">
              <Checkbox
                checked={files.length > 0 && selectedFiles.size === files.length}
                onCheckedChange={() => {
                  if (selectedFiles.size === files.length) {
                    files.forEach((f) => onSelect(f.path, true));
                  } else {
                    files.forEach((f) => {
                      if (!selectedFiles.has(f.path)) {
                        onSelect(f.path, true);
                      }
                    });
                  }
                }}
              />
            </TableHead>
            <TableHead
              className="cursor-pointer hover:bg-accent/50"
              onClick={() => handleHeaderClick('name')}
            >
              <div className="flex items-center gap-1">
                Name
                {getSortIcon('name')}
              </div>
            </TableHead>
            <TableHead
              className="cursor-pointer hover:bg-accent/50 w-24"
              onClick={() => handleHeaderClick('size')}
            >
              <div className="flex items-center gap-1">
                Size
                {getSortIcon('size')}
              </div>
            </TableHead>
            <TableHead
              className="cursor-pointer hover:bg-accent/50 w-40"
              onClick={() => handleHeaderClick('date')}
            >
              <div className="flex items-center gap-1">
                Modified
                {getSortIcon('date')}
              </div>
            </TableHead>
            <TableHead className="w-24">Permissions</TableHead>
            <TableHead className="w-24">Owner</TableHead>
            <TableHead className="w-24">Group</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {files.map((file) => (
            <ContextMenu key={file.path}>
              <ContextMenuTrigger>
                <TableRow
                  className={`cursor-pointer ${
                    selectedFiles.has(file.path) ? 'bg-accent' : ''
                  }`}
                  onClick={(e) => {
                    if (e.ctrlKey || e.metaKey) {
                      onSelect(file.path, true);
                    } else if (e.shiftKey) {
                      // Range selection logic would go here
                      onSelect(file.path, false);
                    } else {
                      onSelect(file.path, false);
                      onFileClick(file);
                    }
                  }}
                >
                  <TableCell>
                    <Checkbox
                      checked={selectedFiles.has(file.path)}
                      onCheckedChange={() => onSelect(file.path, true)}
                      onClick={(e) => e.stopPropagation()}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getFileIcon(file)}
                      <span className="font-medium">{file.name}</span>
                      {file.isSymlink && (
                        <Badge variant="outline" className="text-xs">
                          → {file.target}
                        </Badge>
                      )}
                      {bookmarks.includes(file.path) && (
                        <Star className="h-3 w-3 text-yellow-500 fill-yellow-500" />
                      )}
                    </div>
                  </TableCell>
                  <TableCell>{formatSize(file.size)}</TableCell>
                  <TableCell>{formatDate(file.modified)}</TableCell>
                  <TableCell>
                    <code className="text-xs bg-muted px-1 py-0.5 rounded">
                      {file.permissions}
                    </code>
                  </TableCell>
                  <TableCell>{file.owner}</TableCell>
                  <TableCell>{file.group}</TableCell>
                </TableRow>
              </ContextMenuTrigger>
              <ContextMenuContent className="w-48">
                <ContextMenuItem onClick={() => onFileClick(file)}>
                  {file.type === 'directory' ? (
                    <><Folder className="h-4 w-4 mr-2" /> Open</>
                  ) : (
                    <><File className="h-4 w-4 mr-2" /> Preview</>
                  )}
                </ContextMenuItem>
                {file.type === 'file' && (
                  <ContextMenuItem onClick={() => onDownload(file)}>
                    <Download className="h-4 w-4 mr-2" /> Download
                  </ContextMenuItem>
                )}
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => setRenameDialog({ file, newName: file.name })}>
                  <Pencil className="h-4 w-4 mr-2" /> Rename
                </ContextMenuItem>
                <ContextMenuItem onClick={() => setChmodDialog({ file, permissions: file.permissions })}>
                  <Info className="h-4 w-4 mr-2" /> Permissions
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onBookmarkToggle(file.path)}>
                  {bookmarks.includes(file.path) ? (
                    <><StarOff className="h-4 w-4 mr-2" /> Remove Bookmark</>
                  ) : (
                    <><Star className="h-4 w-4 mr-2" /> Bookmark</>
                  )}
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => onSelect(file.path, true)}>
                  <Copy className="h-4 w-4 mr-2" /> Copy
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onSelect(file.path, true)}>
                  <Scissors className="h-4 w-4 mr-2" /> Cut
                </ContextMenuItem>
                <ContextMenuItem onClick={() => setInfoDialog(file)}>
                  <Info className="h-4 w-4 mr-2" /> Properties
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem
                  className="text-red-600"
                  onClick={() => onSelect(file.path, true)}
                >
                  <Trash2 className="h-4 w-4 mr-2" /> Delete
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          ))}
        </TableBody>
      </Table>

      {/* Rename Dialog */}
      {renameDialog && (
        <Dialog open onOpenChange={() => setRenameDialog(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Rename</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <Input
                value={renameDialog.newName}
                onChange={(e) =>
                  setRenameDialog({ ...renameDialog, newName: e.target.value })
                }
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setRenameDialog(null)}>
                  Cancel
                </Button>
                <Button
                  onClick={() => {
                    onRename(renameDialog.file.path, renameDialog.newName);
                    setRenameDialog(null);
                  }}
                >
                  Rename
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Chmod Dialog */}
      {chmodDialog && (
        <Dialog open onOpenChange={() => setChmodDialog(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Change Permissions</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <Input
                value={chmodDialog.permissions}
                onChange={(e) =>
                  setChmodDialog({ ...chmodDialog, permissions: e.target.value })
                }
                placeholder="e.g., 755"
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setChmodDialog(null)}>
                  Cancel
                </Button>
                <Button
                  onClick={() => {
                    onChmod(chmodDialog.file.path, chmodDialog.permissions);
                    setChmodDialog(null);
                  }}
                >
                  Change
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Info Dialog */}
      {infoDialog && (
        <Dialog open onOpenChange={() => setInfoDialog(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>File Properties</DialogTitle>
            </DialogHeader>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Name:</span>
                <span>{infoDialog.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Path:</span>
                <span className="font-mono">{infoDialog.path}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Size:</span>
                <span>{formatSize(infoDialog.size)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Type:</span>
                <span>{infoDialog.type}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Modified:</span>
                <span>{formatDate(infoDialog.modified)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Permissions:</span>
                <code className="bg-muted px-1 rounded">{infoDialog.permissions}</code>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Owner:</span>
                <span>{infoDialog.owner}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Group:</span>
                <span>{infoDialog.group}</span>
              </div>
              {infoDialog.isSymlink && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Target:</span>
                  <span>{infoDialog.target}</span>
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
