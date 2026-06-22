'use client'

import { useState } from 'react'
import {
  Plus,
  Settings,
  Minimize2,
  Maximize2,
  X,
  Monitor,
  Wifi,
  WifiOff,
  Terminal,
  FolderOpen,
  Key,
} from 'lucide-react'
import { invoke } from '@tauri-apps/api/core'

interface ConnectionBarProps {
  activeHost: string | null
  onSettingsClick: () => void
  onKeyManagerClick?: () => void
}

interface Tab {
  id: string
  hostId: string
  name: string
  connected: boolean
}

const mockTabs: Tab[] = [
  { id: 'tab-1', hostId: 'host-1', name: 'Production Server', connected: true },
  { id: 'tab-2', hostId: 'host-2', name: 'Staging Server', connected: false },
]

export function ConnectionBar({ activeHost, onSettingsClick, onKeyManagerClick }: ConnectionBarProps) {
  const [tabs, setTabs] = useState<Tab[]>(mockTabs)
  const [activeTab, setActiveTab] = useState('tab-1')
  const [isMaximized, setIsMaximized] = useState(false)

  const handleMinimize = async () => {
    try {
      await invoke('minimize_to_tray')
    } catch (e) {
      console.error('Failed to minimize to tray:', e)
    }
  }

  const handleMaximize = async () => {
    try {
      // In a real implementation, this would use Tauri's window API
      setIsMaximized(!isMaximized)
    } catch (e) {
      console.error('Failed to maximize window:', e)
    }
  }

  const handleClose = async () => {
    try {
      await invoke('minimize_to_tray')
    } catch (e) {
      console.error('Failed to close window:', e)
    }
  }

  const handleCloseTab = (tabId: string) => {
    setTabs((prev) => prev.filter((t) => t.id !== tabId))
  }

  return (
    <div className="flex items-center bg-secondary/50 border-b border-border">
      {/* Tabs */}
      <div className="flex-1 flex items-center overflow-x-auto scrollbar-hide">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-3 py-2 text-sm border-r border-border min-w-0 transition-colors ${
              activeTab === tab.id
                ? 'bg-background text-foreground'
                : 'text-muted-foreground hover:text-foreground hover:bg-secondary'
            }`}
          >
            {tab.connected ? (
              <Wifi className="w-3 h-3 text-green-500 flex-shrink-0" />
            ) : (
              <WifiOff className="w-3 h-3 text-muted-foreground flex-shrink-0" />
            )}
            <span className="truncate">{tab.name}</span>
            <button
              onClick={(e) => {
                e.stopPropagation()
                handleCloseTab(tab.id)
              }}
              className="flex-shrink-0 p-0.5 hover:bg-secondary rounded transition-colors"
            >
              <X className="w-3 h-3" />
            </button>
          </button>
        ))}

        {/* New Tab Button */}
        <button
          className="p-2 text-muted-foreground hover:text-foreground transition-colors"
          title="New Tab"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 px-2">
        {/* SFTP Toggle */}
        <button
          className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors"
          title="Toggle SFTP"
        >
          <FolderOpen className="w-4 h-4" />
        </button>

        {/* Key Manager Toggle */}
        <button
          onClick={onKeyManagerClick}
          className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors"
          title="SSH Key Manager"
        >
          <Key className="w-4 h-4" />
        </button>

        {/* Settings */}
        <button
          onClick={onSettingsClick}
          className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors"
          title="Settings"
        >
          <Settings className="w-4 h-4" />
        </button>

        {/* Window Controls */}
        <div className="flex items-center ml-1 border-l border-border pl-1">
          <button
            onClick={handleMinimize}
            className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors"
            title="Minimize to Tray"
          >
            <Minimize2 className="w-4 h-4" />
          </button>
          <button
            onClick={handleMaximize}
            className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-md transition-colors"
            title="Maximize"
          >
            {isMaximized ? (
              <Monitor className="w-4 h-4" />
            ) : (
              <Maximize2 className="w-4 h-4" />
            )}
          </button>
          <button
            onClick={handleClose}
            className="p-1.5 text-muted-foreground hover:text-red-500 hover:bg-red-500/10 rounded-md transition-colors"
            title="Close to Tray"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
