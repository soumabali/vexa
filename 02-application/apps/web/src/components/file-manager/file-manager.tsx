'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { useToast } from '@/hooks/use-toast';
import { FileList } from './file-list';
import { FileGridView } from './file-grid-view';
import { FileTree } from './file-tree';
import { FilePreview } from './file-preview';
import { FileUpload } from './file-upload';
import { FileToolbar } from './file-toolbar';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import {
  Loader2,
  AlertTriangle,
  Check,
  X,
  Folder,
  File,
  Upload,
  Download,
  Copy,
  Move,
  Trash2,
  Pencil,
  Settings,
  Search,
  Grid3X3,
  List,
  TreePine,
  ChevronRight,
  Home,
  Star,
  Clock,
  HardDrive,
  Cloud,
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

interface FileSystem {
  name: string;
  host: string;
  port: number;
  username: string;
  path: string;
  protocol: 'sftp' | 'scp';
}

interface FileManagerProps {
  type: 'local' | 'remote';
  host: FileSystem | null;
  compact?: boolean;
  onTransferProgress?: (progress: {
    file: string;
    progress: number;
    speed: string;
    eta: string;
  }) => void;
  onTransferComplete?: () => void;
}

type ViewMode = 'list' | 'grid' | 'tree';

export function FileManager({
  type,
  host,
  compact = false,
  onTransferProgress,
  onTransferComplete,
}: FileManagerProps) {
  const { toast } = useToast();
  const [currentPath, setCurrentPath] = useState('/');
  const [files, setFiles] = useState<FileItem[]>([]);
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [isLoading, setIsLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<'name' | 'size' | 'date' | 'type'>('name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [previewFile, setPreviewFile] = useState<FileItem | null>(null);
  const [showUpload, setShowUpload] = useState(false);
  const [clipboard, setClipboard] = useState<{
    files: string[];
    operation: 'copy' | 'move';
    sourcePath: string;
  } | null>(null);
  const [history, setHistory] = useState<string[]>(['/']);
  const [historyIndex, setHistoryIndex] = useState(0);
  const [bookmarks, setBookmarks] = useState<string[]>([]);

  const fetchFiles = useCallback(async () => {
    setIsLoading(true);
    try {
      const url = type === 'local'
        ? `/api/files/local?path=${encodeURIComponent(currentPath)}`
        : `/api/files/remote?host=${host?.host}&path=${encodeURIComponent(currentPath)}`;

      const response = await fetch(url);
      const data = await response.json();
      setFiles(data.files || []);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch files',
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  }, [currentPath, type, host, toast]);

  useEffect(() => {
    Promise.resolve().then(() => fetchFiles());
  }, [fetchFiles]);

  const navigateTo = useCallback((path: string) => {
    const newHistory = history.slice(0, historyIndex + 1);
    newHistory.push(path);
    setHistory(newHistory);
    setHistoryIndex(newHistory.length - 1);
    setCurrentPath(path);
    setSelectedFiles(new Set());
  }, [history, historyIndex]);

  const navigateBack = useCallback(() => {
    if (historyIndex > 0) {
      setHistoryIndex(historyIndex - 1);
      setCurrentPath(history[historyIndex - 1]);
      setSelectedFiles(new Set());
    }
  }, [history, historyIndex]);

  const navigateForward = useCallback(() => {
    if (historyIndex < history.length - 1) {
      setHistoryIndex(historyIndex + 1);
      setCurrentPath(history[historyIndex + 1]);
      setSelectedFiles(new Set());
    }
  }, [history, historyIndex]);

  const navigateUp = useCallback(() => {
    const parent = currentPath.split('/').slice(0, -1).join('/') || '/';
    navigateTo(parent);
  }, [currentPath, navigateTo]);

  const handleFileClick = useCallback((file: FileItem) => {
    if (file.type === 'directory') {
      navigateTo(file.path);
    } else {
      setPreviewFile(file);
    }
  }, [navigateTo]);

  const handleSelect = useCallback((path: string, multi: boolean) => {
    setSelectedFiles((prev) => {
      const next = new Set(prev);
      if (multi) {
        if (next.has(path)) {
          next.delete(path);
        } else {
          next.add(path);
        }
      } else {
        if (next.size === 1 && next.has(path)) {
          next.clear();
        } else {
          next.clear();
          next.add(path);
        }
      }
      return next;
    });
  }, []);

  const handleSelectAll = useCallback(() => {
    setSelectedFiles(new Set(files.map((f) => f.path)));
  }, [files]);

  const handleDelete = useCallback(async () => {
    const selected = Array.from(selectedFiles);
    if (selected.length === 0) return;

    try {
      await fetch('/api/files/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ paths: selected, host: host?.host }),
      });

      toast({
        title: 'Deleted',
        description: `${selected.length} item(s) deleted`,
      });

      setSelectedFiles(new Set());
      fetchFiles();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete files',
        variant: 'destructive',
      });
    }
  }, [selectedFiles, host, fetchFiles, toast]);

  const handleCopy = useCallback(() => {
    setClipboard({
      files: Array.from(selectedFiles),
      operation: 'copy',
      sourcePath: currentPath,
    });
    toast({
      title: 'Copied',
      description: `${selectedFiles.size} item(s) copied to clipboard`,
    });
  }, [selectedFiles, currentPath, toast]);

  const handleCut = useCallback(() => {
    setClipboard({
      files: Array.from(selectedFiles),
      operation: 'move',
      sourcePath: currentPath,
    });
    toast({
      title: 'Cut',
      description: `${selectedFiles.size} item(s) cut to clipboard`,
    });
  }, [selectedFiles, currentPath, toast]);

  const handlePaste = useCallback(async () => {
    if (!clipboard) return;

    try {
      await fetch('/api/files/paste', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          files: clipboard.files,
          operation: clipboard.operation,
          sourcePath: clipboard.sourcePath,
          targetPath: currentPath,
          host: host?.host,
        }),
      });

      toast({
        title: 'Pasted',
        description: `${clipboard.files.length} item(s) ${clipboard.operation === 'copy' ? 'copied' : 'moved'}`,
      });

      setClipboard(null);
      fetchFiles();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to paste files',
        variant: 'destructive',
      });
    }
  }, [clipboard, currentPath, host, fetchFiles, toast]);

  const handleRename = useCallback(async (oldPath: string, newName: string) => {
    try {
      await fetch('/api/files/rename', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          oldPath,
          newName,
          host: host?.host,
        }),
      });

      toast({
        title: 'Renamed',
        description: `${oldPath} → ${newName}`,
      });

      fetchFiles();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to rename file',
        variant: 'destructive',
      });
    }
  }, [host, fetchFiles, toast]);

  const handleChmod = useCallback(async (path: string, permissions: string) => {
    try {
      await fetch('/api/files/chmod', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path,
          permissions,
          host: host?.host,
        }),
      });

      toast({
        title: 'Permissions Updated',
        description: `${path} → ${permissions}`,
      });

      fetchFiles();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update permissions',
        variant: 'destructive',
      });
    }
  }, [host, fetchFiles, toast]);

  const handleDownload = useCallback(async (file: FileItem) => {
    try {
      const url = type === 'local'
        ? `/api/files/download?path=${encodeURIComponent(file.path)}`
        : `/api/files/download?host=${host?.host}&path=${encodeURIComponent(file.path)}`;

      const response = await fetch(url);
      const blob = await response.blob();

      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = file.name;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(downloadUrl);

      toast({
        title: 'Downloaded',
        description: file.name,
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to download file',
        variant: 'destructive',
      });
    }
  }, [type, host, toast]);

  const handleUpload = useCallback(async (files: File[]) => {
    setShowUpload(false);

    for (const file of files) {
      try {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('path', currentPath);
        if (host?.host) {
          formData.append('host', host.host);
        }

        const xhr = new XMLHttpRequest();
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = Math.round((event.loaded / event.total) * 100);
            const speed = `${(event.loaded / 1024 / 1024).toFixed(2)} MB/s`;
            const eta = `${Math.round((event.total - event.loaded) / event.loaded * 10)}s`;
            onTransferProgress?.({
              file: file.name,
              progress,
              speed,
              eta,
            });
          }
        });

        xhr.addEventListener('load', () => {
          if (xhr.status === 200) {
            onTransferComplete?.();
            fetchFiles();
          }
        });

        xhr.open('POST', '/api/files/upload');
        xhr.send(formData);
      } catch (error) {
        toast({
          title: 'Error',
          description: `Failed to upload ${file.name}`,
          variant: 'destructive',
        });
      }
    }
  }, [currentPath, host, fetchFiles, onTransferProgress, onTransferComplete, toast]);

  const toggleBookmark = useCallback((path: string) => {
    setBookmarks((prev) => {
      if (prev.includes(path)) {
        return prev.filter((p) => p !== path);
      }
      return [...prev, path];
    });
  }, []);

  const filteredFiles = files.filter((file) =>
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const sortedFiles = [...filteredFiles].sort((a, b) => {
    let comparison = 0;
    switch (sortBy) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'size':
        comparison = a.size - b.size;
        break;
      case 'date':
        comparison = new Date(a.modified).getTime() - new Date(b.modified).getTime();
        break;
      case 'type':
        comparison = a.type.localeCompare(b.type);
        break;
    }
    return sortOrder === 'asc' ? comparison : -comparison;
  });

  return (
    <div className="space-y-2">
      <FileToolbar
        currentPath={currentPath}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        onNavigateUp={navigateUp}
        onNavigateBack={navigateBack}
        onNavigateForward={navigateForward}
        canNavigateBack={historyIndex > 0}
        canNavigateForward={historyIndex < history.length - 1}
        selectedCount={selectedFiles.size}
        onSelectAll={handleSelectAll}
        onDelete={handleDelete}
        onCopy={handleCopy}
        onCut={handleCut}
        onPaste={handlePaste}
        canPaste={!!clipboard}
        onUpload={() => setShowUpload(true)}
        onRefresh={fetchFiles}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        sortBy={sortBy}
        onSortChange={setSortBy}
        sortOrder={sortOrder}
        onSortOrderChange={() => setSortOrder((o) => o === 'asc' ? 'desc' : 'asc')}
        bookmarks={bookmarks}
        onBookmarkToggle={toggleBookmark}
        onBookmarkNavigate={navigateTo}
        isBookmarked={bookmarks.includes(currentPath)}
      />

      <div className="border rounded-lg">
        {isLoading ? (
          <div className="p-8 flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <>
            {viewMode === 'list' && (
              <FileList
                files={sortedFiles}
                selectedFiles={selectedFiles}
                onSelect={handleSelect}
                onFileClick={handleFileClick}
                onDownload={handleDownload}
                onRename={handleRename}
                onChmod={handleChmod}
                onBookmarkToggle={toggleBookmark}
                bookmarks={bookmarks}
                sortBy={sortBy}
                sortOrder={sortOrder}
                onSortChange={setSortBy}
                onSortOrderChange={() => setSortOrder((o) => o === 'asc' ? 'desc' : 'asc')}
              />
            )}
            {viewMode === 'grid' && (
              <FileGridView
                items={sortedFiles}
                selected={selectedFiles}
                onItemClick={(item, e) => handleSelect(item.path, e.ctrlKey || e.metaKey)}
                onItemDoubleClick={handleFileClick}
              />
            )}
            {viewMode === 'tree' && (
              <FileTree
                files={sortedFiles}
                selectedFiles={selectedFiles}
                onSelect={handleSelect}
                onFileClick={handleFileClick}
                currentPath={currentPath}
                onNavigate={navigateTo}
              />
            )}
          </>
        )}
      </div>

      <div className="text-sm text-muted-foreground flex justify-between">
        <span>
          {sortedFiles.length} item(s)
          {selectedFiles.size > 0 && ` • ${selectedFiles.size} selected`}
        </span>
        <span>
          {formatBytes(sortedFiles.reduce((sum, f) => sum + f.size, 0))}
        </span>
      </div>

      {/* Preview Dialog */}
      <Dialog open={!!previewFile} onOpenChange={() => setPreviewFile(null)}>
        <DialogContent className="max-w-4xl max-h-[90vh]">
          <DialogHeader>
            <DialogTitle>{previewFile?.name}</DialogTitle>
          </DialogHeader>
          {previewFile && (
            <FilePreview
              file={previewFile}
              host={host}
              onDownload={handleDownload}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Upload Dialog */}
      <Dialog open={showUpload} onOpenChange={setShowUpload}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Upload Files</DialogTitle>
          </DialogHeader>
          <FileUpload onUpload={handleUpload} />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}
