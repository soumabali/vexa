'use client'

import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'
import type { SftpFile } from '@/types/sftp'
import {
  FolderPlus,
  Pencil,
  Trash2,
  Shield,
  FileDown,
  FileUp,
  Scissors,
  Copy,
  ClipboardPaste,
} from 'lucide-react'

export type FileAction =
  | 'rename'
  | 'delete'
  | 'chmod'
  | 'download'
  | 'upload'
  | 'new-folder'
  | 'cut'
  | 'copy'
  | 'paste'

interface ContextMenuItem {
  action: FileAction
  label: string
  icon: React.ReactNode
  danger?: boolean
  separator?: boolean
  disabled?: boolean
}

interface FileContextMenuProps {
  file: SftpFile | null
  position: { x: number; y: number }
  isLocal: boolean
  onAction: (action: FileAction) => void
  onClose: () => void
  clipboard?: { files: string[]; action: 'copy' | 'cut' } | null
}

export function FileContextMenu({ file, position, isLocal, onAction, onClose, clipboard }: FileContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (typeof document === 'undefined') return
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => {
      if (typeof document !== 'undefined') {
        document.removeEventListener('mousedown', handleClick)
      }
    }
  }, [onClose])

  const items: ContextMenuItem[] = [
    ...(file
      ? [
          { action: 'copy' as FileAction, label: 'Copy', icon: <Copy size={14} /> },
          { action: 'cut' as FileAction, label: 'Cut', icon: <Scissors size={14} /> },
          ...(clipboard
            ? [{ action: 'paste' as FileAction, label: 'Paste', icon: <ClipboardPaste size={14} /> }]
            : []),
        ]
      : clipboard
        ? [{ action: 'paste' as FileAction, label: 'Paste', icon: <ClipboardPaste size={14} /> }]
        : []),
    { action: 'new-folder' as FileAction, label: 'New Folder', icon: <FolderPlus size={14} />, separator: true },
    ...(file
      ? [
          ...(!isLocal && !file.isDirectory
            ? [{ action: 'download' as FileAction, label: 'Download', icon: <FileDown size={14} /> }]
            : []),
          ...(isLocal && !file.isDirectory
            ? [{ action: 'upload' as FileAction, label: 'Upload', icon: <FileUp size={14} /> }]
            : []),
          { action: 'rename' as FileAction, label: 'Rename', icon: <Pencil size={14} />, separator: true },
          { action: 'chmod' as FileAction, label: 'Set Permissions', icon: <Shield size={14} /> },
          {
            action: 'delete' as FileAction,
            label: 'Delete',
            icon: <Trash2 size={14} />,
            danger: true,
          },
        ]
      : []),
  ]

  return (
    <div
      ref={menuRef}
      className="fixed z-50 min-w-[160px] bg-popover border border-border rounded-md shadow-lg py-1"
      style={{
        left: Math.min(position.x, typeof window !== 'undefined' ? window.innerWidth - 200 : position.x),
        top: Math.min(position.y, typeof window !== 'undefined' ? window.innerHeight - 300 : position.y),
      }}
    >
      {items.map((item, idx) => (
        <div key={`${item.action}-${idx}`}>
          {item.separator && idx > 0 && <div className="my-1 border-t border-border" />}
          <button
            onClick={() => {
              onAction(item.action)
              onClose()
            }}
            disabled={item.disabled}
            className={cn(
              'w-full flex items-center gap-2 px-3 py-1.5 text-sm transition-colors',
              item.danger
                ? 'text-destructive hover:bg-destructive/10'
                : 'text-foreground hover:bg-muted',
              item.disabled && 'opacity-40 cursor-not-allowed'
            )}
          >
            <span className="text-muted-foreground">{item.icon}</span>
            {item.label}
          </button>
        </div>
      ))}
    </div>
  )
}
