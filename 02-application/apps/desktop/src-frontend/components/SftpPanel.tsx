'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { useSftp } from '@/hooks/useSftp'
import { useLocalFiles } from '@/hooks/useLocalFiles'
import { SftpFileTree } from '@/components/SftpFileTree'
import { FileContextMenu } from '@/components/FileContextMenu'
import { TransferProgress } from '@/components/TransferProgress'
import type { SftpFile, FileConflict, ViewMode } from '@/types/sftp'
import type { FileAction } from '@/components/FileContextMenu'
import {
  FolderSync,
  Split,
  Monitor,
  Server,
  AlertTriangle,
  CheckCircle,
  X,
  Loader2,
  Plus,
  Pencil,
  Shield,
  Trash2,
} from 'lucide-react'

interface SftpPanelProps {
  hostId: string
}

export function SftpPanel({ hostId }: SftpPanelProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('split')
  const [selectedRemote, setSelectedRemote] = useState<Set<string>>(new Set())
  const [selectedLocal, setSelectedLocal] = useState<Set<string>>(new Set())
  const [contextMenu, setContextMenu] = useState<{
    file: SftpFile | null
    position: { x: number; y: number }
    isLocal: boolean
  } | null>(null)
  const [renameDialog, setRenameDialog] = useState<{ file: SftpFile; isLocal: boolean } | null>(null)
  const [chmodDialog, setChmodDialog] = useState<SftpFile | null>(null)
  const [newFolderDialog, setNewFolderDialog] = useState<{ isLocal: boolean } | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<{ file: SftpFile; isLocal: boolean } | null>(null)
  const [dragOverRemote, setDragOverRemote] = useState<string | null>(null)
  const [dragOverLocal, setDragOverLocal] = useState<string | null>(null)
  const [clipboard, setClipboard] = useState<{ files: string[]; action: 'copy' | 'cut' } | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const {
    remotePath,
    remoteFiles,
    isLoading: remoteLoading,
    error: remoteError,
    transfers,
    pendingConflicts,
    listRemote,
    navigateUp,
    navigateTo,
    uploadFile,
    downloadFile,
    renameFile,
    deleteFile,
    chmodFile,
    createDirectory,
    cancelTransfer,
    clearCompleted,
    resolveConflict,
    dismissTransfer,
  } = useSftp(hostId)

  const {
    localPath,
    localFiles,
    isLoading: localLoading,
    error: localError,
    listLocal,
    navigateLocalUp,
    navigateLocalTo,
    openFilePicker,
  } = useLocalFiles()

  // Initialize on mount
  useEffect(() => {
    listRemote()
    listLocal()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hostId])

  // Handle remote file selection
  const handleRemoteSelect = useCallback((file: SftpFile) => {
    setSelectedRemote((prev) => {
      const next = new Set(prev)
      if (next.has(file.path)) next.delete(file.path)
      else next.add(file.path)
      return next
    })
  }, [])

  // Handle local file selection
  const handleLocalSelect = useCallback((file: SftpFile) => {
    setSelectedLocal((prev) => {
      const next = new Set(prev)
      if (next.has(file.path)) next.delete(file.path)
      else next.add(file.path)
      return next
    })
  }, [])

  // Context menu handler
  const handleContextMenu = useCallback(
    (e: React.MouseEvent, file: SftpFile, isLocal: boolean) => {
      e.preventDefault()
      setContextMenu({ file, position: { x: e.clientX, y: e.clientY }, isLocal })
    },
    []
  )

  // Execute action from context menu
  const handleFileAction = useCallback(
    (action: FileAction, file: SftpFile | null, isLocal: boolean) => {
      if (action === 'rename' && file) {
        setRenameDialog({ file, isLocal })
      } else if (action === 'delete' && file) {
        setDeleteConfirm({ file, isLocal })
      } else if (action === 'chmod' && file) {
        setChmodDialog(file)
      } else if (action === 'new-folder') {
        setNewFolderDialog({ isLocal })
      } else if (action === 'download' && file) {
        downloadFile(file.path, localPath)
      } else if (action === 'upload' && file) {
        uploadFile(file.path, remotePath)
      } else if (action === 'copy' && file) {
        setClipboard({ files: [file.path], action: 'copy' })
      } else if (action === 'cut' && file) {
        setClipboard({ files: [file.path], action: 'cut' })
      }
    },
    [downloadFile, uploadFile, localPath, remotePath]
  )

  // Confirm rename
  const confirmRename = useCallback(
    async (newName: string) => {
      if (!renameDialog) return
      const { file, isLocal } = renameDialog
      if (!newName.trim() || newName === file.name) {
        setRenameDialog(null)
        return
      }
      if (isLocal) {
        // Local rename via Tauri (future)
        console.warn('Local rename not yet implemented via Tauri')
      } else {
        await renameFile(file.path, newName.trim())
      }
      setRenameDialog(null)
    },
    [renameDialog, renameFile]
  )

  // Confirm delete
  const confirmDelete = useCallback(async () => {
    if (!deleteConfirm) return
    const { file, isLocal } = deleteConfirm
    if (isLocal) {
      console.warn('Local delete not yet implemented via Tauri')
    } else {
      await deleteFile(file.path, file.isDirectory)
    }
    setDeleteConfirm(null)
  }, [deleteConfirm, deleteFile])

  // Confirm chmod
  const confirmChmod = useCallback(
    async (permissions: string) => {
      if (!chmodDialog) return
      await chmodFile(chmodDialog.path, permissions)
      setChmodDialog(null)
    },
    [chmodDialog, chmodFile]
  )

  // Confirm new folder
  const confirmNewFolder = useCallback(
    async (name: string) => {
      if (!newFolderDialog) return
      if (!name.trim()) {
        setNewFolderDialog(null)
        return
      }
      if (newFolderDialog.isLocal) {
        console.warn('Local new folder not yet implemented via Tauri')
      } else {
        await createDirectory(name.trim())
      }
      setNewFolderDialog(null)
    },
    [newFolderDialog, createDirectory]
  )

  // Drag & drop: local → remote (upload)
  const handleLocalDragStart = useCallback((e: React.DragEvent, file: SftpFile) => {
    if (file.isDirectory) {
      e.preventDefault()
      return
    }
    e.dataTransfer.setData('text/plain', JSON.stringify({ type: 'local', path: file.path, name: file.name }))
    e.dataTransfer.effectAllowed = 'copy'
  }, [])

  const handleRemoteDragOver = useCallback((e: React.DragEvent, file?: SftpFile) => {
    e.preventDefault()
    const dropPath = file?.isDirectory ? file.path : remotePath
    setDragOverRemote(dropPath)
    e.dataTransfer.dropEffect = 'copy'
  }, [remotePath])

  const handleRemoteDrop = useCallback(
    async (e: React.DragEvent, file?: SftpFile) => {
      e.preventDefault()
      setDragOverRemote(null)
      const data = e.dataTransfer.getData('text/plain')
      if (!data) return
      try {
        const parsed = JSON.parse(data)
        if (parsed.type === 'local') {
          const targetPath = file?.isDirectory ? file.path : remotePath
          await uploadFile(parsed.path, targetPath)
        }
      } catch {
        // Try plain text path fallback
        const path = data
        if (path) {
          const targetPath = file?.isDirectory ? file.path : remotePath
          await uploadFile(path, targetPath)
        }
      }
    },
    [remotePath, uploadFile]
  )

  // Drag & drop: remote → local (download)
  const handleRemoteDragStart = useCallback((e: React.DragEvent, file: SftpFile) => {
    if (file.isDirectory) {
      e.preventDefault()
      return
    }
    e.dataTransfer.setData('text/plain', JSON.stringify({ type: 'remote', path: file.path, name: file.name }))
    e.dataTransfer.effectAllowed = 'copy'
  }, [])

  const handleLocalDragOver = useCallback((e: React.DragEvent, file?: SftpFile) => {
    e.preventDefault()
    const dropPath = file?.isDirectory ? file.path : localPath
    setDragOverLocal(dropPath)
    e.dataTransfer.dropEffect = 'copy'
  }, [localPath])

  const handleLocalDrop = useCallback(
    async (e: React.DragEvent, file?: SftpFile) => {
      e.preventDefault()
      setDragOverLocal(null)
      const data = e.dataTransfer.getData('text/plain')
      if (!data) return
      try {
        const parsed = JSON.parse(data)
        if (parsed.type === 'remote') {
          const targetDir = file?.isDirectory ? file.path : localPath
          await downloadFile(parsed.path, targetDir)
        }
      } catch {
        const path = data
        if (path) {
          const targetDir = file?.isDirectory ? file.path : localPath
          await downloadFile(path, targetDir)
        }
      }
    },
    [localPath, downloadFile]
  )

  // Resolve all conflicts
  const handleResolveAll = useCallback(
    async (resolution: 'use-local' | 'use-remote' | 'skip') => {
      for (const conflict of pendingConflicts) {
        await resolveConflict(conflict, resolution)
      }
    },
    [pendingConflicts, resolveConflict]
  )

  const showRemote = viewMode === 'split' || viewMode === 'remote-only'
  const showLocal = viewMode === 'split' || viewMode === 'local-only'

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-3 py-2 bg-secondary/50 border-b border-border">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode('split')}
            className={cn(
              'p-1.5 rounded text-xs flex items-center gap-1 transition-colors',
              viewMode === 'split' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            )}
            title="Split view"
          >
            <Split size={14} /> Split
          </button>
          <button
            onClick={() => setViewMode('remote-only')}
            className={cn(
              'p-1.5 rounded text-xs flex items-center gap-1 transition-colors',
              viewMode === 'remote-only' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            )}
            title="Remote only"
          >
            <Server size={14} /> Remote
          </button>
          <button
            onClick={() => setViewMode('local-only')}
            className={cn(
              'p-1.5 rounded text-xs flex items-center gap-1 transition-colors',
              viewMode === 'local-only' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            )}
            title="Local only"
          >
            <Monitor size={14} /> Local
          </button>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => handleFileAction('new-folder', null, false)}
            className="p-1.5 rounded text-xs flex items-center gap-1 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          >
            <Plus size={14} /> New Folder
          </button>
          <button
            onClick={() => {
              listRemote()
              listLocal()
            }}
            className="p-1.5 rounded text-xs flex items-center gap-1 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          >
            <FolderSync size={14} /> Sync
          </button>
        </div>
      </div>

      {/* Conflict banner */}
      {pendingConflicts.length > 0 && (
        <div className="flex items-center gap-3 px-3 py-2 bg-yellow-500/10 border-b border-yellow-500/20 text-sm">
          <AlertTriangle size={16} className="text-yellow-500 flex-shrink-0" />
          <span className="flex-1">{pendingConflicts.length} file conflict{pendingConflicts.length > 1 ? 's' : ''} detected</span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => handleResolveAll('use-local')}
              className="px-2 py-1 text-xs bg-background border border-border rounded hover:bg-muted transition-colors"
            >
              Use Local
            </button>
            <button
              onClick={() => handleResolveAll('use-remote')}
              className="px-2 py-1 text-xs bg-background border border-border rounded hover:bg-muted transition-colors"
            >
              Use Remote
            </button>
            <button
              onClick={() => handleResolveAll('skip')}
              className="px-2 py-1 text-xs bg-background border border-border rounded hover:bg-muted transition-colors"
            >
              Skip All
            </button>
          </div>
        </div>
      )}

      {/* Error banners */}
      {remoteError && (
        <div className="flex items-center gap-2 px-3 py-2 bg-destructive/10 border-b border-destructive/20 text-sm text-destructive">
          <AlertTriangle size={14} />
          <span>Remote: {remoteError}</span>
          <button onClick={() => listRemote()} className="ml-auto text-xs underline">Retry</button>
        </div>
      )}
      {localError && (
        <div className="flex items-center gap-2 px-3 py-2 bg-destructive/10 border-b border-destructive/20 text-sm text-destructive">
          <AlertTriangle size={14} />
          <span>Local: {localError}</span>
          <button onClick={() => listLocal()} className="ml-auto text-xs underline">Retry</button>
        </div>
      )}

      {/* File browser panes */}
      <div className="flex-1 flex min-h-0 overflow-hidden">
        {showLocal && (
          <div className={cn('flex flex-col min-w-0', showRemote && 'w-1/2 border-r border-border')}>
            <SftpFileTree
              files={localFiles}
              currentPath={localPath}
              isLoading={localLoading}
              selectedPaths={selectedLocal}
              expandedPaths={new Set()}
              onSelect={handleLocalSelect}
              onToggleExpand={() => {}}
              onNavigateUp={navigateLocalUp}
              onNavigateTo={navigateLocalTo}
              onRefresh={() => listLocal()}
              onContextMenu={(e, file) => handleContextMenu(e, file, true)}
              title="Local Files"
              icon={<Monitor size={16} className="text-muted-foreground" />}
              onDragStart={handleLocalDragStart}
              onDragOver={handleLocalDragOver}
              onDrop={handleLocalDrop}
              dragOverPath={dragOverLocal}
            />
          </div>
        )}

        {showRemote && (
          <div className={cn('flex flex-col min-w-0', showLocal && 'w-1/2')}>
            <SftpFileTree
              files={remoteFiles}
              currentPath={remotePath}
              isLoading={remoteLoading}
              selectedPaths={selectedRemote}
              expandedPaths={new Set()}
              onSelect={handleRemoteSelect}
              onToggleExpand={() => {}}
              onNavigateUp={navigateUp}
              onNavigateTo={navigateTo}
              onRefresh={listRemote}
              onContextMenu={(e, file) => handleContextMenu(e, file, false)}
              title="Remote Files"
              icon={<Server size={16} className="text-muted-foreground" />}
              onDragStart={handleRemoteDragStart}
              onDragOver={handleRemoteDragOver}
              onDrop={handleRemoteDrop}
              dragOverPath={dragOverRemote}
            />
          </div>
        )}
      </div>

      {/* Transfer progress panel */}
      <TransferProgress
        transfers={transfers}
        onCancel={cancelTransfer}
        onDismiss={dismissTransfer}
        onClearCompleted={clearCompleted}
      />

      {/* Context Menu */}
      {contextMenu && (
        <FileContextMenu
          file={contextMenu.file}
          position={contextMenu.position}
          isLocal={contextMenu.isLocal}
          clipboard={clipboard}
          onAction={(action) => {
            if (contextMenu.file) {
              handleFileAction(action, contextMenu.file, contextMenu.isLocal)
            } else {
              handleFileAction(action, null, contextMenu.isLocal)
            }
            setContextMenu(null)
          }}
          onClose={() => setContextMenu(null)}
        />
      )}

      {/* Rename Dialog */}
      {renameDialog && (
        <Dialog
          title="Rename"
          icon={<Pencil size={16} />}
          onClose={() => setRenameDialog(null)}
          onConfirm={() => confirmRename(inputRef.current?.value || '')}
          confirmText="Rename"
        >
          <input
            ref={inputRef}
            defaultValue={renameDialog.file.name}
            className="w-full px-3 py-2 bg-background border border-border rounded text-sm"
            placeholder="New name"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') confirmRename(inputRef.current?.value || '')
              if (e.key === 'Escape') setRenameDialog(null)
            }}
          />
        </Dialog>
      )}

      {/* Chmod Dialog */}
      {chmodDialog && (
        <Dialog
          title="Set Permissions"
          icon={<Shield size={16} />}
          onClose={() => setChmodDialog(null)}
          onConfirm={() => confirmChmod(inputRef.current?.value || '644')}
          confirmText="Apply"
        >
          <input
            ref={inputRef}
            defaultValue={chmodDialog.permissions}
            className="w-full px-3 py-2 bg-background border border-border rounded text-sm font-mono"
            placeholder="e.g. 755, 644"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') confirmChmod(inputRef.current?.value || '644')
              if (e.key === 'Escape') setChmodDialog(null)
            }}
          />
          <p className="text-xs text-muted-foreground mt-1">Enter numeric permissions (e.g. 755, 644)</p>
        </Dialog>
      )}

      {/* New Folder Dialog */}
      {newFolderDialog && (
        <Dialog
          title="New Folder"
          icon={<Plus size={16} />}
          onClose={() => setNewFolderDialog(null)}
          onConfirm={() => confirmNewFolder(inputRef.current?.value || '')}
          confirmText="Create"
        >
          <input
            ref={inputRef}
            className="w-full px-3 py-2 bg-background border border-border rounded text-sm"
            placeholder="Folder name"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') confirmNewFolder(inputRef.current?.value || '')
              if (e.key === 'Escape') setNewFolderDialog(null)
            }}
          />
        </Dialog>
      )}

      {/* Delete Confirmation */}
      {deleteConfirm && (
        <Dialog
          title="Confirm Delete"
          icon={<Trash2 size={16} className="text-destructive" />}
          onClose={() => setDeleteConfirm(null)}
          onConfirm={confirmDelete}
          confirmText="Delete"
          confirmDanger
        >
          <p className="text-sm">
            Are you sure you want to delete{' '}
            <strong>{deleteConfirm.file.name}</strong>
            {deleteConfirm.file.isDirectory && ' and all its contents'}?
          </p>
          <p className="text-xs text-muted-foreground mt-2">This action cannot be undone.</p>
        </Dialog>
      )}
    </div>
  )
}

// Simple reusable dialog component
interface DialogProps {
  title: string
  icon: React.ReactNode
  children: React.ReactNode
  onClose: () => void
  onConfirm: () => void
  confirmText: string
  confirmDanger?: boolean
}

function Dialog({ title, icon, children, onClose, onConfirm, confirmText, confirmDanger }: DialogProps) {
  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50">
      <div className="bg-card border border-border rounded-lg shadow-xl w-full max-w-sm mx-4 overflow-hidden">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
          <span className="text-muted-foreground">{icon}</span>
          <span className="font-medium">{title}</span>
          <button onClick={onClose} className="ml-auto text-muted-foreground hover:text-foreground">
            <X size={16} />
          </button>
        </div>
        <div className="px-4 py-3">{children}</div>
        <div className="flex justify-end gap-2 px-4 py-3 border-t border-border">
          <button
            onClick={onClose}
            className="px-3 py-1.5 text-sm bg-secondary text-secondary-foreground rounded hover:bg-secondary/80 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className={cn(
              'px-3 py-1.5 text-sm rounded transition-colors',
              confirmDanger
                ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            )}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
