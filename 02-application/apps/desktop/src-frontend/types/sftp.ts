/** SFTP types shared between frontend components and hooks */

export interface SftpFile {
  name: string
  path: string
  size: number
  isDirectory: boolean
  permissions: string
  modifiedTime: string
  owner?: string
  group?: string
  isSymlink?: boolean
  symlinkTarget?: string
}

export interface SftpDirectory {
  path: string
  files: SftpFile[]
}

export type FileTransferStatus = 'queued' | 'preparing' | 'transferring' | 'paused' | 'completed' | 'error' | 'cancelled'

export interface FileTransfer {
  id: string
  filename: string
  remotePath: string
  localPath: string
  direction: 'upload' | 'download'
  status: FileTransferStatus
  progress: number // 0-100
  bytesTransferred: number
  totalBytes: number
  speed: number // bytes per second
  error?: string
  startedAt?: string
  completedAt?: string
}

export interface SyncStatus {
  path: string
  localModified: string
  remoteModified: string
  localSize: number
  remoteSize: number
  state: 'in-sync' | 'local-newer' | 'remote-newer' | 'conflict' | 'missing-local' | 'missing-remote'
}

export interface FileConflict {
  path: string
  localSize: number
  remoteSize: number
  localModified: string
  remoteModified: string
}

export type ConflictResolution = 'use-local' | 'use-remote' | 'skip' | 'rename-local' | 'rename-remote'

export interface ContextMenuItem {
  label: string
  action: string
  icon?: React.ReactNode
  danger?: boolean
  separator?: boolean
  disabled?: boolean
}

export type ViewMode = 'split' | 'remote-only' | 'local-only'
