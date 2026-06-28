'use client';

import React, { createContext, useContext, useState, useCallback, useRef } from 'react';
import { FileItem, TransferItem, ClipboardItem, Bookmark, FileFilter, PaneSide } from '@/types/file-manager';
import { filterFiles, sortFiles } from '@/lib/file-utils';

interface FileManagerContextType {
  // Local state
  localPath: string;
  localItems: FileItem[];
  localSelected: Set<string>;
  localSortBy: 'name' | 'size' | 'date' | 'permissions';
  localSortOrder: 'asc' | 'desc';
  localViewMode: 'list' | 'grid' | 'tree';
  localFilter: FileFilter;
  localLoading: boolean;
  localError: string | null;
  
  // Remote state
  remotePath: string;
  remoteItems: FileItem[];
  remoteSelected: Set<string>;
  remoteSortBy: 'name' | 'size' | 'date' | 'permissions';
  remoteSortOrder: 'asc' | 'desc';
  remoteViewMode: 'list' | 'grid' | 'tree';
  remoteFilter: FileFilter;
  remoteLoading: boolean;
  remoteError: string | null;
  
  // Transfer queue
  transfers: TransferItem[];
  
  // Clipboard
  clipboard: ClipboardItem | null;
  
  // Bookmarks
  bookmarks: Bookmark[];
  
  // Active pane
  activePane: PaneSide;
  
  // Preview
  previewItem: FileItem | null;
  previewSide: PaneSide | null;
  
  // Actions
  setLocalPath: (path: string) => void;
  setRemotePath: (path: string) => void;
  setLocalItems: (items: FileItem[]) => void;
  setRemoteItems: (items: FileItem[]) => void;
  selectItem: (side: PaneSide, id: string, multi?: boolean, range?: boolean) => void;
  clearSelection: (side: PaneSide) => void;
  setSort: (side: PaneSide, by: 'name' | 'size' | 'date' | 'permissions') => void;
  setViewMode: (side: PaneSide, mode: 'list' | 'grid' | 'tree') => void;
  setFilter: (side: PaneSide, filter: FileFilter) => void;
  setActivePane: (side: PaneSide) => void;
  setClipboard: (clipboard: ClipboardItem | null) => void;
  addBookmark: (bookmark: Omit<Bookmark, 'id'>) => void;
  removeBookmark: (id: string) => void;
  addTransfer: (transfer: TransferItem) => void;
  updateTransfer: (id: string, updates: Partial<TransferItem>) => void;
  removeTransfer: (id: string) => void;
  setPreviewItem: (item: FileItem | null, side?: PaneSide) => void;
  getFilteredItems: (side: PaneSide) => FileItem[];
}

const FileManagerContext = createContext<FileManagerContextType | undefined>(undefined);

export function FileManagerProvider({ children }: { children: React.ReactNode }) {
  // Local state
  const [localPath, setLocalPath] = useState('/home');
  const [localItems, setLocalItems] = useState<FileItem[]>([]);
  const [localSelected, setLocalSelected] = useState<Set<string>>(new Set());
  const [localSortBy, setLocalSortBy] = useState<'name' | 'size' | 'date' | 'permissions'>('name');
  const [localSortOrder, setLocalSortOrder] = useState<'asc' | 'desc'>('asc');
  const [localViewMode, setLocalViewMode] = useState<'list' | 'grid' | 'tree'>('list');
  const [localFilter, setLocalFilter] = useState<FileFilter>({ text: '' });
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  // Remote state
  const [remotePath, setRemotePath] = useState('/home');
  const [remoteItems, setRemoteItems] = useState<FileItem[]>([]);
  const [remoteSelected, setRemoteSelected] = useState<Set<string>>(new Set());
  const [remoteSortBy, setRemoteSortBy] = useState<'name' | 'size' | 'date' | 'permissions'>('name');
  const [remoteSortOrder, setRemoteSortOrder] = useState<'asc' | 'desc'>('asc');
  const [remoteViewMode, setRemoteViewMode] = useState<'list' | 'grid' | 'tree'>('list');
  const [remoteFilter, setRemoteFilter] = useState<FileFilter>({ text: '' });
  const [remoteLoading, setRemoteLoading] = useState(false);
  const [remoteError, setRemoteError] = useState<string | null>(null);

  // Transfer queue
  const [transfers, setTransfers] = useState<TransferItem[]>([]);

  // Clipboard
  const [clipboard, setClipboard] = useState<ClipboardItem | null>(null);

  // Bookmarks
  const [bookmarks, setBookmarks] = useState<Bookmark[]>([
    { id: '1', name: 'Home', path: '/home', side: 'local' },
    { id: '2', name: 'Root', path: '/', side: 'local' },
  ]);

  // Active pane
  const [activePane, setActivePane] = useState<PaneSide>('local');

  // Preview
  const [previewItem, setPreviewItem] = useState<FileItem | null>(null);
  const [previewSide, setPreviewSide] = useState<PaneSide | null>(null);

  const lastSelectedRef = useRef<{ local: string | null; remote: string | null }>({ local: null, remote: null });

  const selectItem = useCallback((side: PaneSide, id: string, multi = false, range = false) => {
    if (side === 'local') {
      if (range && lastSelectedRef.current.local) {
        const items = localItems;
        const lastIdx = items.findIndex(i => i.id === lastSelectedRef.current.local);
        const currentIdx = items.findIndex(i => i.id === id);
        if (lastIdx !== -1 && currentIdx !== -1) {
          const start = Math.min(lastIdx, currentIdx);
          const end = Math.max(lastIdx, currentIdx);
          const newSelected = new Set(localSelected);
          for (let i = start; i <= end; i++) {
            newSelected.add(items[i].id);
          }
          setLocalSelected(newSelected);
        }
      } else if (multi) {
        const newSelected = new Set(localSelected);
        if (newSelected.has(id)) {
          newSelected.delete(id);
        } else {
          newSelected.add(id);
          lastSelectedRef.current.local = id;
        }
        setLocalSelected(newSelected);
      } else {
        setLocalSelected(new Set([id]));
        lastSelectedRef.current.local = id;
      }
    } else {
      if (range && lastSelectedRef.current.remote) {
        const items = remoteItems;
        const lastIdx = items.findIndex(i => i.id === lastSelectedRef.current.remote);
        const currentIdx = items.findIndex(i => i.id === id);
        if (lastIdx !== -1 && currentIdx !== -1) {
          const start = Math.min(lastIdx, currentIdx);
          const end = Math.max(lastIdx, currentIdx);
          const newSelected = new Set(remoteSelected);
          for (let i = start; i <= end; i++) {
            newSelected.add(items[i].id);
          }
          setRemoteSelected(newSelected);
        }
      } else if (multi) {
        const newSelected = new Set(remoteSelected);
        if (newSelected.has(id)) {
          newSelected.delete(id);
        } else {
          newSelected.add(id);
          lastSelectedRef.current.remote = id;
        }
        setRemoteSelected(newSelected);
      } else {
        setRemoteSelected(new Set([id]));
        lastSelectedRef.current.remote = id;
      }
    }
  }, [localItems, localSelected, remoteItems, remoteSelected]);

  const clearSelection = useCallback((side: PaneSide) => {
    if (side === 'local') {
      setLocalSelected(new Set());
      lastSelectedRef.current.local = null;
    } else {
      setRemoteSelected(new Set());
      lastSelectedRef.current.remote = null;
    }
  }, []);

  const setSort = useCallback((side: PaneSide, by: 'name' | 'size' | 'date' | 'permissions') => {
    if (side === 'local') {
      if (localSortBy === by) {
        setLocalSortOrder(prev => prev === 'asc' ? 'desc' : 'asc');
      } else {
        setLocalSortBy(by);
        setLocalSortOrder('asc');
      }
    } else {
      if (remoteSortBy === by) {
        setRemoteSortOrder(prev => prev === 'asc' ? 'desc' : 'asc');
      } else {
        setRemoteSortBy(by);
        setRemoteSortOrder('asc');
      }
    }
  }, [localSortBy, remoteSortBy]);

  const setViewMode = useCallback((side: PaneSide, mode: 'list' | 'grid' | 'tree') => {
    if (side === 'local') {
      setLocalViewMode(mode);
    } else {
      setRemoteViewMode(mode);
    }
  }, []);

  const setFilter = useCallback((side: PaneSide, filter: FileFilter) => {
    if (side === 'local') {
      setLocalFilter(filter);
    } else {
      setRemoteFilter(filter);
    }
  }, []);

  const addBookmark = useCallback((bookmark: Omit<Bookmark, 'id'>) => {
    const newBookmark: Bookmark = {
      ...bookmark,
      id: Math.random().toString(36).substring(7),
    };
    setBookmarks(prev => [...prev, newBookmark]);
  }, []);

  const removeBookmark = useCallback((id: string) => {
    setBookmarks(prev => prev.filter(b => b.id !== id));
  }, []);

  const addTransfer = useCallback((transfer: TransferItem) => {
    setTransfers(prev => [...prev, transfer]);
  }, []);

  const updateTransfer = useCallback((id: string, updates: Partial<TransferItem>) => {
    setTransfers(prev =>
      prev.map(t => (t.id === id ? { ...t, ...updates } : t))
    );
  }, []);

  const removeTransfer = useCallback((id: string) => {
    setTransfers(prev => prev.filter(t => t.id !== id));
  }, []);

  const setPreviewItemCallback = useCallback((item: FileItem | null, side?: PaneSide) => {
    setPreviewItem(item);
    setPreviewSide(side || null);
  }, []);

  const getFilteredItems = useCallback((side: PaneSide) => {
    const items = side === 'local' ? localItems : remoteItems;
    const sortBy = side === 'local' ? localSortBy : remoteSortBy;
    const sortOrder = side === 'local' ? localSortOrder : remoteSortOrder;
    const filter = side === 'local' ? localFilter : remoteFilter;
    
    let filtered = filterFiles(items, filter.text);
    if (filter.types) {
      const types = filter.types;
      filtered = filtered.filter(f => types.includes(f.type as 'file' | 'directory' | 'symlink'));
    }
    return sortFiles(filtered, sortBy, sortOrder) as FileItem[];
  }, [localItems, localSortBy, localSortOrder, localFilter, remoteItems, remoteSortBy, remoteSortOrder, remoteFilter]);

  return (
    <FileManagerContext.Provider
      value={{
        localPath,
        localItems,
        localSelected,
        localSortBy,
        localSortOrder,
        localViewMode,
        localFilter,
        localLoading,
        localError,
        remotePath,
        remoteItems,
        remoteSelected,
        remoteSortBy,
        remoteSortOrder,
        remoteViewMode,
        remoteFilter,
        remoteLoading,
        remoteError,
        transfers,
        clipboard,
        bookmarks,
        activePane,
        previewItem,
        previewSide,
        setLocalPath,
        setRemotePath,
        setLocalItems,
        setRemoteItems,
        selectItem,
        clearSelection,
        setSort,
        setViewMode,
        setFilter,
        setActivePane,
        setClipboard,
        addBookmark,
        removeBookmark,
        addTransfer,
        updateTransfer,
        removeTransfer,
        setPreviewItem: setPreviewItemCallback,
        getFilteredItems,
      }}
    >
      {children}
    </FileManagerContext.Provider>
  );
}

export function useFileManager() {
  const context = useContext(FileManagerContext);
  if (!context) {
    throw new Error('useFileManager must be used within a FileManagerProvider');
  }
  return context;
}
