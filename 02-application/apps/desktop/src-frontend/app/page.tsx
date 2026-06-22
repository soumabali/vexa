'use client'

import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { HostList } from '@/components/HostList'
import { Terminal } from '@/components/Terminal'
import { SftpPanel } from '@/components/SftpPanel'
import { ConnectionBar } from '@/components/ConnectionBar'
import { SettingsPanel } from '@/components/SettingsPanel'
import { KeyManager } from '@/components/KeyManager'

export default function Home() {
  const [activeHost, setActiveHost] = useState<string | null>(null)
  const [showSettings, setShowSettings] = useState(false)
  const [showKeyManager, setShowKeyManager] = useState(false)
  const [activePanel, setActivePanel] = useState<'terminal' | 'sftp'>('terminal')
  const [sidebarWidth, setSidebarWidth] = useState(280)

  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'K') {
        e.preventDefault()
        setShowKeyManager((prev) => !prev)
      }
    }
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [])

  const handleHostSelect = (hostId: string) => {
    setActiveHost(hostId)
  }

  const handleQuickConnect = () => {
    setActiveHost('quick-connect')
  }

  const handleOpenKeyManager = () => {
    setShowKeyManager(true)
  }

  return (
    <div className="flex h-screen w-screen bg-background text-foreground overflow-hidden">
      {/* Sidebar - Host List */}
      <div
        className="flex-shrink-0 border-r border-border flex flex-col"
        style={{ width: sidebarWidth }}
      >
        <HostList
          activeHost={activeHost}
          onHostSelect={handleHostSelect}
          onQuickConnect={handleQuickConnect}
        />
      </div>

      {/* Resizer */}
      <div
        className="w-1 cursor-col-resize hover:bg-primary/50 transition-colors"
        onMouseDown={(e) => {
          if (typeof window === 'undefined') return
          const startX = e.clientX
          const startWidth = sidebarWidth

          const handleMouseMove = (e: MouseEvent) => {
            const newWidth = Math.max(200, Math.min(400, startWidth + e.clientX - startX))
            setSidebarWidth(newWidth)
          }

          const handleMouseUp = () => {
            if (typeof document !== 'undefined') {
              document.removeEventListener('mousemove', handleMouseMove)
              document.removeEventListener('mouseup', handleMouseUp)
            }
          }

          if (typeof document !== 'undefined') {
            document.addEventListener('mousemove', handleMouseMove)
            document.addEventListener('mouseup', handleMouseUp)
          }
        }}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Connection Bar */}
        <ConnectionBar
          activeHost={activeHost}
          onSettingsClick={() => setShowSettings(true)}
          onKeyManagerClick={() => setShowKeyManager(true)}
        />

        {/* Panel Tabs */}
        {activeHost && (
          <div className="flex items-center gap-1 px-3 py-1.5 bg-secondary/50 border-b border-border">
            <button
              onClick={() => setActivePanel('terminal')}
              className={cn(
                'px-3 py-1 text-xs rounded transition-colors',
                activePanel === 'terminal'
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted'
              )}
            >
              Terminal
            </button>
            <button
              onClick={() => setActivePanel('sftp')}
              className={cn(
                'px-3 py-1 text-xs rounded transition-colors',
                activePanel === 'sftp'
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted'
              )}
            >
              SFTP
            </button>
          </div>
        )}

        {/* Main Panel */}
        <div className="flex-1 min-h-0">
          {activeHost ? (
            activePanel === 'terminal' ? (
              <Terminal hostId={activeHost} />
            ) : (
              <SftpPanel hostId={activeHost} />
            )
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              <div className="text-center">
                <h2 className="text-2xl font-semibold mb-2">Welcome to vexa</h2>
                <p className="text-sm">Select a host from the sidebar or quick connect to get started</p>
                <button
                  onClick={handleQuickConnect}
                  className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                >
                  Quick Connect (Ctrl+Shift+T)
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <SettingsPanel onClose={() => setShowSettings(false)} />
      )}

      {/* Key Manager */}
      {showKeyManager && (
        <KeyManager isOpen={showKeyManager} onClose={() => setShowKeyManager(false)} />
      )}
    </div>
  )
}
