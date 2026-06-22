'use client'

import { useState, useCallback, useRef } from 'react'
import { cn } from '@/lib/utils'
import type { SftpFile } from '@/types/sftp'
import {
  Folder,
  FileText,
  ChevronRight,
  ChevronDown,
  HardDrive,
  ArrowUp,
  RefreshCw,
  Loader2,
} from 'lucide-react'

interface SftpFileTreeProps {
  files: SftpFile[]
  currentPath: string
  isLoading: boolean
  selectedPaths: Set<string>
  expandedPaths: Set<string>
  onSelect: (file: SftpFile) => void
  onToggleExpand: (path: string) => void
  onNavigateUp: () => void
  onNavigateTo: (path: string) => void
  onRefresh: () => void
  onContextMenu: (e: React.MouseEvent, file: SftpFile) => void
  title: string
  icon?: React.ReactNode
  onDragStart?: (e: React.DragEvent, file: SftpFile) => void
  onDragOver?: (e: React.DragEvent, file?: SftpFile) => void
  onDrop?: (e: React.DragEvent, file?: SftpFile) => void
  dragOverPath?: string | null
}

function FileIcon({ file }: { file: SftpFile }) {
  if (file.isDirectory) return <Folder size={16} className="text-yellow-500" />
  if (file.isSymlink) return <FileText size={16} className="text-purple-400" />
  return <FileText size={16} className="text-blue-400" />
}

function formatSize(size: number): string {
  if (size === 0) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`
  return `${(size / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function SftpFileTree({
  files,
  currentPath,
  isLoading,
  selectedPaths,
  expandedPaths,
  onSelect,
  onToggleExpand,
  onNavigateUp,
  onNavigateTo,
  onRefresh,
  onContextMenu,
  title,
  icon,
  onDragStart,
  onDragOver,
  onDrop,
  dragOverPath,
}: SftpFileTreeProps) {
  const [sortColumn, setSortColumn] = useState<'name' | 'size' | 'modified'>('name')
  const [sortAsc, setSortAsc] = useState(true)

  const handleSort = useCallback((col: 'name' | 'size' | 'modified') => {
    setSortColumn((prev) => {
      if (prev === col) {
        setSortAsc((a) => !a)
        return prev
      }
      setSortAsc(true)
      return col
    })
  }, [])

  const sortedFiles = [...files].sort((a, b) => {
    if (a.isDirectory !== b.isDirectory) return a.isDirectory ? -1 : 1
    let cmp = 0
    if (sortColumn === 'name') cmp = a.name.localeCompare(b.name)
    else if (sortColumn === 'size') cmp = a.size - b.size
    else if (sortColumn === 'modified') cmp = new Date(a.modifiedTime).getTime() - new Date(b.modifiedTime).getTime()
    return sortAsc ? cmp : -cmp
  })

  return (
    <div className="flex flex-col h-full border border-border rounded-md overflow-hidden bg-card">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-secondary/50 border-b border-border">
        <div className="flex items-center gap-2">
          {icon || <HardDrive size={16} className="text-muted-foreground" />}
          <span className="text-sm font-medium">{title}</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={onNavigateUp}
            disabled={currentPath === '/'}
            className="p-1 hover:bg-muted rounded text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors"
            title="Parent directory"
          >
            <ArrowUp size={14} />
          </button>
          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="p-1 hover:bg-muted rounded text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors"
            title="Refresh"
          >
            {isLoading ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
          </button>
        </div>
      </div>

      {/* Breadcrumb / Path bar */}
      <div className="flex items-center px-3 py-1.5 bg-background border-b border-border text-xs text-muted-foreground overflow-hidden">
        <span className="truncate select-all">{currentPath || '/'}</span>
      </div>

      {/* Column headers */}
      <div className="grid grid-cols-[1fr_80px_140px] gap-2 px-3 py-1.5 bg-secondary/30 border-b border-border text-xs font-medium text-muted-foreground select-none">
        <button onClick={() => handleSort('name')} className="text-left flex items-center gap-1 hover:text-foreground">
          Name {sortColumn === 'name' && (sortAsc ? '↑' : '↓')}
        </button>
        <button onClick={() => handleSort('size')} className="text-right flex items-center justify-end gap-1 hover:text-foreground">
          Size {sortColumn === 'size' && (sortAsc ? '↑' : '↓')}
        </button>
        <button onClick={() => handleSort('modified')} className="text-right flex items-center justify-end gap-1 hover:text-foreground">
          Modified {sortColumn === 'modified' && (sortAsc ? '↑' : '↓')}
        </button>
      </div>

      {/* File list */}
      <div className="flex-1 overflow-y-auto">
        {sortedFiles.length === 0 && !isLoading && (
          <div className="flex items-center justify-center h-32 text-sm text-muted-foreground">
            Directory is empty
          </div>
        )}

        {sortedFiles.map((file) => {
          const isSelected = selectedPaths.has(file.path)
          const isDragOver = dragOverPath === file.path

          return (
            <div
              key={file.path}
              className={cn(
                'grid grid-cols-[1fr_80px_140px] gap-2 px-3 py-1.5 text-sm items-center cursor-pointer select-none border-b border-border/50 transition-colors',
                isSelected && 'bg-primary/10',
                isDragOver && 'bg-primary/20',
                !isSelected && 'hover:bg-muted/50'
              )}
              onClick={() => onSelect(file)}
              onDoubleClick={() => {
                if (file.isDirectory) onNavigateTo(file.path)
              }}
              onContextMenu={(e) => {
                e.preventDefault()
                onContextMenu(e, file)
              }}
              draggable={!file.isDirectory}
              onDragStart={(e) => onDragStart?.(e, file)}
              onDragOver={(e) => {
                e.preventDefault()
                onDragOver?.(e, file)
              }}
              onDragLeave={() => onDragOver?.({} as React.DragEvent, undefined)}
              onDrop={(e) => {
                e.preventDefault()
                onDrop?.(e, file)
              }}
            >
              <div className="flex items-center gap-2 min-w-0">
                <FileIcon file={file} />
                <span className="truncate" title={file.name}>{file.name}</span>
              </div>
              <div className="text-right text-xs text-muted-foreground">
                {file.isDirectory ? '-' : formatSize(file.size)}
              </div>
              <div className="text-right text-xs text-muted-foreground">
                {formatDate(file.modifiedTime)}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
