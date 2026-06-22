'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen, UnlistenFn } from '@tauri-apps/api/event'
import type { SftpFile, FileTransfer, FileTransferStatus, SyncStatus, FileConflict, ConflictResolution } from '@/types/sftp'

// Generate unique ID for transfers
const generateId = () => Math.random().toString(36).substring(2, 15)

// Progress update from Rust backend
interface ProgressPayload {
  transferId: string
  bytesTransferred: number
  totalBytes: number
  speed: number
}

export function useSftp(hostId: string) {
  const [remotePath, setRemotePath] = useState('/')
  const [remoteFiles, setRemoteFiles] = useState<SftpFile[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [transfers, setTransfers] = useState<FileTransfer[]>([])
  const [syncStatus, setSyncStatus] = useState<SyncStatus[]>([])
  const [pendingConflicts, setPendingConflicts] = useState<FileConflict[]>([])
  const unlistenRef = useRef<UnlistenFn | null>(null)

  // Listen for transfer progress events from Tauri
  useEffect(() => {
    const setupListener = async () => {
      const unlisten = await listen<ProgressPayload>('sftp:transfer-progress', (event) => {
        const { transferId, bytesTransferred, totalBytes, speed } = event.payload
        setTransfers((prev) =>
          prev.map((t) =>
            t.id === transferId
              ? {
                  ...t,
                  bytesTransferred,
                  totalBytes,
                  speed,
                  progress: totalBytes > 0 ? Math.round((bytesTransferred / totalBytes) * 100) : 0,
                }
              : t
          )
        )
      })
      unlistenRef.current = unlisten
    }
    setupListener()
    return () => {
      unlistenRef.current?.()
    }
  }, [hostId])

  // List remote directory
  const listRemote = useCallback(
    async (path: string = remotePath) => {
      setIsLoading(true)
      setError(null)
      try {
        const files = await invoke<SftpFile[]>('sftp_list_directory', { hostId, path })
        setRemoteFiles(files)
        setRemotePath(path)
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setError(message)
        console.error('SFTP list error:', err)
      } finally {
        setIsLoading(false)
      }
    },
    [hostId, remotePath]
  )

  // Navigate up
  const navigateUp = useCallback(() => {
    if (remotePath === '/') return
    const parts = remotePath.split('/').filter(Boolean)
    parts.pop()
    const parent = '/' + parts.join('/')
    listRemote(parent || '/')
  }, [remotePath, listRemote])

  // Navigate to a directory
  const navigateTo = useCallback(
    (path: string) => {
      listRemote(path)
    },
    [listRemote]
  )

  // Upload file
  const uploadFile = useCallback(
    async (localPath: string, remoteTargetPath: string) => {
      const id = generateId()
      const filename = localPath.split(/[/\\]/).pop() || 'unknown'
      const remoteFilePath = remoteTargetPath.endsWith('/') ? `${remoteTargetPath}${filename}` : `${remoteTargetPath}/${filename}`

      const transfer: FileTransfer = {
        id,
        filename,
        remotePath: remoteFilePath,
        localPath,
        direction: 'upload',
        status: 'queued',
        progress: 0,
        bytesTransferred: 0,
        totalBytes: 0,
        speed: 0,
        startedAt: new Date().toISOString(),
      }

      setTransfers((prev) => [transfer, ...prev])

      try {
        setTransfers((prev) => prev.map((t) => (t.id === id ? { ...t, status: 'preparing' } : t)))
        await invoke('sftp_upload_file', { hostId, localPath, remotePath: remoteFilePath, transferId: id })
        setTransfers((prev) =>
          prev.map((t) =>
            t.id === id ? { ...t, status: 'completed', progress: 100, completedAt: new Date().toISOString() } : t
          )
        )
        // Refresh directory
        listRemote()
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setTransfers((prev) => prev.map((t) => (t.id === id ? { ...t, status: 'error', error: message } : t)))
      }
    },
    [hostId, listRemote]
  )

  // Download file
  const downloadFile = useCallback(
    async (remoteFilePath: string, localDir: string) => {
      const id = generateId()
      const filename = remoteFilePath.split('/').pop() || 'unknown'
      const localPath = `${localDir}/${filename}`

      const transfer: FileTransfer = {
        id,
        filename,
        remotePath: remoteFilePath,
        localPath,
        direction: 'download',
        status: 'queued',
        progress: 0,
        bytesTransferred: 0,
        totalBytes: 0,
        speed: 0,
        startedAt: new Date().toISOString(),
      }

      setTransfers((prev) => [transfer, ...prev])

      try {
        setTransfers((prev) => prev.map((t) => (t.id === id ? { ...t, status: 'preparing' } : t)))
        await invoke('sftp_download_file', { hostId, remotePath: remoteFilePath, localDir, transferId: id })
        setTransfers((prev) =>
          prev.map((t) =>
            t.id === id ? { ...t, status: 'completed', progress: 100, completedAt: new Date().toISOString() } : t
          )
        )
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setTransfers((prev) => prev.map((t) => (t.id === id ? { ...t, status: 'error', error: message } : t)))
      }
    },
    [hostId]
  )

  // Rename file
  const renameFile = useCallback(
    async (oldPath: string, newName: string) => {
      try {
        const parent = oldPath.substring(0, oldPath.lastIndexOf('/')) || '/'
        const newPath = parent === '/' ? `/${newName}` : `${parent}/${newName}`
        await invoke('sftp_rename', { hostId, oldPath, newPath })
        listRemote()
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setError(message)
      }
    },
    [hostId, listRemote]
  )

  // Delete file/directory
  const deleteFile = useCallback(
    async (path: string, isDirectory: boolean) => {
      try {
        if (isDirectory) {
          await invoke('sftp_remove_directory', { hostId, path })
        } else {
          await invoke('sftp_remove_file', { hostId, path })
        }
        listRemote()
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setError(message)
      }
    },
    [hostId, listRemote]
  )

  // Change permissions
  const chmodFile = useCallback(
    async (path: string, permissions: string) => {
      try {
        await invoke('sftp_chmod', { hostId, path, permissions })
        listRemote()
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setError(message)
      }
    },
    [hostId, listRemote]
  )

  // Create new directory
  const createDirectory = useCallback(
    async (dirName: string) => {
      try {
        const path = remotePath === '/' ? `/${dirName}` : `${remotePath}/${dirName}`
        await invoke('sftp_create_directory', { hostId, path })
        listRemote()
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        setError(message)
      }
    },
    [hostId, remotePath, listRemote]
  )

  // Cancel a transfer
  const cancelTransfer = useCallback(
    async (transferId: string) => {
      try {
        await invoke('sftp_cancel_transfer', { hostId, transferId })
        setTransfers((prev) => prev.map((t) => (t.id === transferId ? { ...t, status: 'cancelled' } : t)))
      } catch (err) {
        console.error('Failed to cancel transfer:', err)
      }
    },
    [hostId]
  )

  // Clear completed/cancelled/error transfers
  const clearCompleted = useCallback(() => {
    setTransfers((prev) => prev.filter((t) => t.status === 'queued' || t.status === 'preparing' || t.status === 'transferring' || t.status === 'paused'))
  }, [])

  // Resolve a conflict
  const resolveConflict = useCallback(
    async (conflict: FileConflict, resolution: ConflictResolution) => {
      try {
        await invoke('sftp_resolve_conflict', { hostId, path: conflict.path, resolution })
        setPendingConflicts((prev) => prev.filter((c) => c.path !== conflict.path))
        listRemote()
      } catch (err) {
        console.error('Failed to resolve conflict:', err)
      }
    },
    [hostId, listRemote]
  )

  // Dismiss a transfer
  const dismissTransfer = useCallback((transferId: string) => {
    setTransfers((prev) => prev.filter((t) => t.id !== transferId))
  }, [])

  return {
    remotePath,
    remoteFiles,
    isLoading,
    error,
    transfers,
    syncStatus,
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
    setRemotePath,
  }
}
