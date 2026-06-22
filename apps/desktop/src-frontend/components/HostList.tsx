'use client'

import { useState, useCallback } from 'react'
import {
  Plus,
  Search,
  Server,
  Folder,
  ChevronRight,
  ChevronDown,
  MoreVertical,
  Star,
  Wifi,
  WifiOff,
  RefreshCw,
} from 'lucide-react'

interface Host {
  id: string
  name: string
  hostname: string
  port: number
  username: string
  protocol: 'ssh' | 'sftp' | 'local'
  status: 'connected' | 'disconnected' | 'connecting'
  favorite: boolean
  folderId?: string
}

interface Folder {
  id: string
  name: string
  expanded: boolean
  hosts: string[]
}

const mockHosts: Host[] = [
  {
    id: 'host-1',
    name: 'Production Server',
    hostname: 'prod.example.com',
    port: 22,
    username: 'admin',
    protocol: 'ssh',
    status: 'connected',
    favorite: true,
  },
  {
    id: 'host-2',
    name: 'Staging Server',
    hostname: 'staging.example.com',
    port: 22,
    username: 'deploy',
    protocol: 'ssh',
    status: 'disconnected',
    favorite: false,
    folderId: 'folder-1',
  },
  {
    id: 'host-3',
    name: 'Development Server',
    hostname: 'dev.example.com',
    port: 22,
    username: 'developer',
    protocol: 'ssh',
    status: 'disconnected',
    favorite: false,
    folderId: 'folder-1',
  },
  {
    id: 'host-4',
    name: 'Database Server',
    hostname: 'db.example.com',
    port: 22,
    username: 'dbadmin',
    protocol: 'ssh',
    status: 'disconnected',
    favorite: true,
  },
  {
    id: 'host-5',
    name: 'Local Terminal',
    hostname: 'localhost',
    port: 0,
    username: '',
    protocol: 'local',
    status: 'disconnected',
    favorite: false,
  },
]

const mockFolders: Folder[] = [
  {
    id: 'folder-1',
    name: 'Web Servers',
    expanded: true,
    hosts: ['host-2', 'host-3'],
  },
]

interface HostListProps {
  activeHost: string | null
  onHostSelect: (hostId: string) => void
  onQuickConnect: () => void
}

export function HostList({ activeHost, onHostSelect, onQuickConnect }: HostListProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [hosts] = useState<Host[]>(mockHosts)
  const [folders, setFolders] = useState<Folder[]>(mockFolders)
  const [showNewHostModal, setShowNewHostModal] = useState(false)
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; hostId: string } | null>(null)

  const toggleFolder = useCallback((folderId: string) => {
    setFolders((prev) =>
      prev.map((f) => (f.id === folderId ? { ...f, expanded: !f.expanded } : f))
    )
  }, [])

  const filteredHosts = hosts.filter(
    (host) =>
      searchQuery === '' ||
      host.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      host.hostname.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const favoriteHosts = filteredHosts.filter((h) => h.favorite && !h.folderId)
  const regularHosts = filteredHosts.filter((h) => !h.favorite && !h.folderId)

  const getStatusIcon = (status: Host['status']) => {
    switch (status) {
      case 'connected':
        return <Wifi className="w-3 h-3 text-green-500" />
      case 'connecting':
        return <RefreshCw className="w-3 h-3 text-yellow-500 animate-spin" />
      case 'disconnected':
      default:
        return <WifiOff className="w-3 h-3 text-muted-foreground" />
    }
  }

  const getProtocolIcon = (protocol: Host['protocol']) => {
    switch (protocol) {
      case 'ssh':
        return <Server className="w-4 h-4" />
      case 'sftp':
        return <Server className="w-4 h-4 text-blue-500" />
      case 'local':
        return <Server className="w-4 h-4 text-purple-500" />
      default:
        return <Server className="w-4 h-4" />
    }
  }

  const HostItem = ({ host }: { host: Host }) => (
    <button
      onClick={() => onHostSelect(host.id)}
      onContextMenu={(e) => {
        e.preventDefault()
        setContextMenu({ x: e.clientX, y: e.clientY, hostId: host.id })
      }}
      className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-secondary/50 transition-colors group ${
        activeHost === host.id ? 'bg-secondary' : ''
      }`}
    >
      <div className="flex-shrink-0">{getProtocolIcon(host.protocol)}</div>
      <div className="flex-1 text-left min-w-0">
        <div className="truncate font-medium">{host.name}</div>
        <div className="truncate text-xs text-muted-foreground">
          {host.hostname}
          {host.port !== 0 && `:${host.port}`}
        </div>
      </div>
      <div className="flex items-center gap-1">
        {host.favorite && <Star className="w-3 h-3 text-yellow-500 fill-yellow-500" />}
        {getStatusIcon(host.status)}
      </div>
    </button>
  )

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-border">
        <h2 className="text-sm font-semibold">Hosts</h2>
        <button
          onClick={() => setShowNewHostModal(true)}
          className="p-1.5 hover:bg-secondary rounded-md transition-colors"
          title="Add new host"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>

      {/* Search */}
      <div className="p-2 border-b border-border">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search hosts..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 bg-secondary/50 border border-border rounded-md text-sm"
          />
        </div>
      </div>

      {/* Quick Connect */}
      <button
        onClick={onQuickConnect}
        className="flex items-center gap-2 w-full px-3 py-2.5 text-sm hover:bg-secondary/50 transition-colors border-b border-border"
      >
        <Plus className="w-4 h-4 text-primary" />
        <span className="font-medium">Quick Connect</span>
        <span className="text-xs text-muted-foreground ml-auto">Ctrl+Shift+T</span>
      </button>

      {/* Host List */}
      <div className="flex-1 overflow-y-auto">
        {/* Favorites */}
        {favoriteHosts.length > 0 && (
          <div className="py-1">
            <div className="px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Favorites
            </div>
            {favoriteHosts.map((host) => (
              <HostItem key={host.id} host={host} />
            ))}
          </div>
        )}

        {/* Folders */}
        {folders.map((folder) => (
          <div key={folder.id} className="py-1">
            <button
              onClick={() => toggleFolder(folder.id)}
              className="flex items-center gap-1.5 w-full px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider hover:text-foreground transition-colors"
            >
              {folder.expanded ? (
                <ChevronDown className="w-3 h-3" />
              ) : (
                <ChevronRight className="w-3 h-3" />
              )}
              <Folder className="w-3 h-3" />
              {folder.name}
            </button>
            {folder.expanded &&
              filteredHosts
                .filter((h) => h.folderId === folder.id)
                .map((host) => <HostItem key={host.id} host={host} />)}
          </div>
        ))}

        {/* Regular Hosts */}
        {regularHosts.length > 0 && (
          <div className="py-1">
            <div className="px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              All Hosts
            </div>
            {regularHosts.map((host) => (
              <HostItem key={host.id} host={host} />
            ))}
          </div>
        )}
      </div>

      {/* New Host Modal (placeholder) */}
      {showNewHostModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background border border-border rounded-lg p-6 w-96 max-w-[90vw]">
            <h3 className="text-lg font-semibold mb-4">New Host</h3>
            <div className="space-y-3">
              <div>
                <label className="text-sm text-muted-foreground">Name</label>
                <input
                  type="text"
                  className="w-full mt-1 px-3 py-2 bg-secondary border border-border rounded-md text-sm"
                  placeholder="My Server"
                />
              </div>
              <div>
                <label className="text-sm text-muted-foreground">Hostname</label>
                <input
                  type="text"
                  className="w-full mt-1 px-3 py-2 bg-secondary border border-border rounded-md text-sm"
                  placeholder="example.com"
                />
              </div>
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="text-sm text-muted-foreground">Port</label>
                  <input
                    type="number"
                    defaultValue={22}
                    className="w-full mt-1 px-3 py-2 bg-secondary border border-border rounded-md text-sm"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-sm text-muted-foreground">Protocol</label>
                  <select className="w-full mt-1 px-3 py-2 bg-secondary border border-border rounded-md text-sm">
                    <option value="ssh">SSH</option>
                    <option value="sftp">SFTP</option>
                  </select>
                </div>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button
                onClick={() => setShowNewHostModal(false)}
                className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  setShowNewHostModal(false)
                  // TODO: Add host logic
                }}
                className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
              >
                Add Host
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Context Menu */}
      {contextMenu && (
        <div
          className="fixed bg-background border border-border rounded-md shadow-lg py-1 z-50"
          style={{ top: contextMenu.y, left: contextMenu.x }}
          onClick={() => setContextMenu(null)}
        >
          <button className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors">
            Edit
          </button>
          <button className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors">
            Duplicate
          </button>
          <div className="border-t border-border my-1" />
          <button className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors text-red-500">
            Delete
          </button>
        </div>
      )}
    </div>
  )
}
