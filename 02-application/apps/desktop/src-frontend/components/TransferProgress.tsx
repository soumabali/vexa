'use client'

import { useState } from 'react'
import { cn } from '@/lib/utils'
import type { FileTransfer } from '@/types/sftp'
import { FileUp, FileDown, X, CheckCircle, AlertCircle, Loader2, Trash2, Minimize2, Maximize2 } from 'lucide-react'

interface TransferProgressProps {
  transfers: FileTransfer[]
  onCancel: (id: string) => void
  onDismiss: (id: string) => void
  onClearCompleted: () => void
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

function formatSpeed(bps: number): string {
  return `${formatBytes(bps)}/s`
}

function TransferItem({ transfer, onCancel, onDismiss }: { transfer: FileTransfer; onCancel: (id: string) => void; onDismiss: (id: string) => void }) {
  const isUpload = transfer.direction === 'upload'
  const isActive = transfer.status === 'queued' || transfer.status === 'preparing' || transfer.status === 'transferring'
  const isDone = transfer.status === 'completed'
  const isError = transfer.status === 'error'

  return (
    <div className={cn('flex items-center gap-2 px-3 py-2 border-b border-border text-sm', isDone && 'opacity-60')}>
      <div className="flex-shrink-0 text-muted-foreground">
        {isUpload ? <FileUp size={16} /> : <FileDown size={16} />}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-0.5">
          <span className="truncate font-medium text-foreground">{transfer.filename}</span>
          <span className="text-xs text-muted-foreground ml-2 flex-shrink-0">
            {formatBytes(transfer.bytesTransferred)} / {formatBytes(transfer.totalBytes)}
          </span>
        </div>
        <div className="relative h-1.5 bg-muted rounded-full overflow-hidden">
          <div
            className={cn(
              'absolute inset-y-0 left-0 rounded-full transition-all duration-300',
              isError && 'bg-destructive',
              isDone && 'bg-green-500',
              isActive && 'bg-primary'
            )}
            style={{ width: `${Math.min(transfer.progress, 100)}%` }}
          />
        </div>
        <div className="flex items-center justify-between mt-0.5 text-xs text-muted-foreground">
          <span className="capitalize">
            {isActive && transfer.speed > 0 ? formatSpeed(transfer.speed) : transfer.status}
          </span>
          <span>{transfer.progress}%</span>
        </div>
      </div>
      <div className="flex-shrink-0 flex items-center gap-1">
        {isActive && (
          <button
            onClick={() => onCancel(transfer.id)}
            className="p-1 hover:bg-muted rounded text-muted-foreground hover:text-foreground transition-colors"
            title="Cancel"
          >
            <X size={14} />
          </button>
        )}
        {(isDone || isError || transfer.status === 'cancelled') && (
          <button
            onClick={() => onDismiss(transfer.id)}
            className="p-1 hover:bg-muted rounded text-muted-foreground hover:text-foreground transition-colors"
            title="Dismiss"
          >
            <Trash2 size={14} />
          </button>
        )}
      </div>
    </div>
  )
}

export function TransferProgress({ transfers, onCancel, onDismiss, onClearCompleted }: TransferProgressProps) {
  const [collapsed, setCollapsed] = useState(false)

  if (transfers.length === 0) return null

  const activeCount = transfers.filter((t) => t.status === 'queued' || t.status === 'preparing' || t.status === 'transferring').length
  const completedCount = transfers.filter((t) => t.status === 'completed' || t.status === 'error' || t.status === 'cancelled').length

  return (
    <div className="border-t border-border bg-card">
      <div
        className="flex items-center justify-between px-3 py-2 bg-secondary/50 cursor-pointer select-none"
        onClick={() => setCollapsed((c) => !c)}
      >
        <div className="flex items-center gap-2 text-sm font-medium">
          <span>Transfers</span>
          {activeCount > 0 && (
            <span className="inline-flex items-center justify-center px-1.5 py-0.5 text-xs bg-primary text-primary-foreground rounded-full">
              {activeCount}
            </span>
          )}
          {completedCount > 0 && (
            <span className="inline-flex items-center justify-center px-1.5 py-0.5 text-xs bg-muted text-muted-foreground rounded-full">
              {completedCount} done
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {completedCount > 0 && (
            <button
              onClick={(e) => {
                e.stopPropagation()
                onClearCompleted()
              }}
              className="text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              Clear completed
            </button>
          )}
          <button className="text-muted-foreground hover:text-foreground transition-colors">
            {collapsed ? <Maximize2 size={14} /> : <Minimize2 size={14} />}
          </button>
        </div>
      </div>

      {!collapsed && (
        <div className="max-h-48 overflow-y-auto">
          {transfers.map((transfer) => (
            <TransferItem key={transfer.id} transfer={transfer} onCancel={onCancel} onDismiss={onDismiss} />
          ))}
        </div>
      )}
    </div>
  )
}
