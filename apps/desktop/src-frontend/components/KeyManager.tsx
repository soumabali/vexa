'use client'

import { useState, useCallback, useRef } from 'react'
import {
  Key,
  Plus,
  Upload,
  Trash2,
  Copy,
  Eye,
  EyeOff,
  Lock,
  Unlock,
  RefreshCw,
  Server,
  X,
  ChevronRight,
  ChevronDown,
  Check,
  AlertCircle,
  Shield,
  FileKey,
  MoreVertical,
  Search,
  RotateCcw,
  Download,
  Fingerprint,
  KeyRound,
} from 'lucide-react'
import { cn } from '@/lib/utils'

// ── Types ──────────────────────────────────────────

export type KeyAlgorithm = 'rsa' | 'ed25519' | 'ecdsa'
export type KeyStatus = 'active' | 'rotating' | 'expired' | 'compromised'

export interface SSHKey {
  id: string
  name: string
  algorithm: KeyAlgorithm
  bits?: number
  fingerprint: string
  fingerprintSHA256: string
  createdAt: string
  updatedAt: string
  hasPassphrase: boolean
  status: KeyStatus
  publicKey: string
  privateKey: string
  comment?: string
  associatedHosts: string[]
  lastUsedAt?: string
  tags: string[]
}

export interface HostRef {
  id: string
  name: string
  hostname: string
  username: string
}

export interface KeyManagerProps {
  isOpen: boolean
  onClose: () => void
}

// ── Constants ──────────────────────────────────────

const ALGORITHM_OPTIONS: { value: KeyAlgorithm; label: string; bits: number[]; defaultBits: number }[] = [
  { value: 'rsa', label: 'RSA', bits: [2048, 3072, 4096], defaultBits: 4096 },
  { value: 'ed25519', label: 'Ed25519', bits: [256], defaultBits: 256 },
  { value: 'ecdsa', label: 'ECDSA', bits: [256, 384, 521], defaultBits: 256 },
]

const STATUS_COLORS: Record<KeyStatus, string> = {
  active: 'text-green-500',
  rotating: 'text-yellow-500',
  expired: 'text-red-500',
  compromised: 'text-red-600',
}

const STATUS_LABELS: Record<KeyStatus, string> = {
  active: 'Active',
  rotating: 'Rotating',
  expired: 'Expired',
  compromised: 'Compromised',
}

// ── Mock Data (replace with API calls) ─────────────

const mockHosts: HostRef[] = [
  { id: 'host-1', name: 'Production Server', hostname: 'prod.example.com', username: 'admin' },
  { id: 'host-2', name: 'Staging Server', hostname: 'staging.example.com', username: 'deploy' },
  { id: 'host-3', name: 'Development Server', hostname: 'dev.example.com', username: 'developer' },
  { id: 'host-4', name: 'Database Server', hostname: 'db.example.com', username: 'dbadmin' },
]

const mockKeys: SSHKey[] = [
  {
    id: 'key-1',
    name: 'Production Key',
    algorithm: 'ed25519',
    bits: 256,
    fingerprint: 'MD5:ab:cd:ef:12:34:56:78:90:ab:cd:ef:12:34:56:78:90',
    fingerprintSHA256: 'SHA256:abc123def456ghi789jkl012mno345pqr678stu901vwx',
    createdAt: '2026-01-15T10:30:00Z',
    updatedAt: '2026-05-20T14:22:00Z',
    hasPassphrase: true,
    status: 'active',
    publicKey: 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDIhz2GK/XCUj4i6Q5yQJNL1MXMY0RxzPV2QrBqfHrDq production-key',
    privateKey: '-----BEGIN OPENSSH PRIVATE KEY-----\n...',
    comment: 'Production server access',
    associatedHosts: ['host-1'],
    lastUsedAt: '2026-05-27T08:15:00Z',
    tags: ['production', 'critical'],
  },
  {
    id: 'key-2',
    name: 'Staging Key',
    algorithm: 'rsa',
    bits: 4096,
    fingerprint: 'MD5:12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd:ef',
    fingerprintSHA256: 'SHA256:def456ghi789jkl012mno345pqr678stu901vwx234yz',
    createdAt: '2026-02-01T09:00:00Z',
    updatedAt: '2026-05-25T11:00:00Z',
    hasPassphrase: false,
    status: 'rotating',
    publicKey: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC... staging-key',
    privateKey: '-----BEGIN OPENSSH PRIVATE KEY-----\n...',
    comment: 'Staging environment',
    associatedHosts: ['host-2', 'host-3'],
    lastUsedAt: '2026-05-26T16:45:00Z',
    tags: ['staging'],
  },
  {
    id: 'key-3',
    name: 'Legacy RSA Key',
    algorithm: 'rsa',
    bits: 2048,
    fingerprint: 'MD5:ef:12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd',
    fingerprintSHA256: 'SHA256:ghi789jkl012mno345pqr678stu901vwx234yz567abc',
    createdAt: '2025-06-10T08:00:00Z',
    updatedAt: '2025-06-10T08:00:00Z',
    hasPassphrase: true,
    status: 'expired',
    publicKey: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... legacy-key',
    privateKey: '-----BEGIN OPENSSH PRIVATE KEY-----\n...',
    comment: 'Old key - needs rotation',
    associatedHosts: [],
    lastUsedAt: '2025-12-01T10:00:00Z',
    tags: ['legacy', 'needs-rotation'],
  },
]

// ── Helper Components ──────────────────────────────

function Badge({
  children,
  variant = 'default',
}: {
  children: React.ReactNode
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info'
}) {
  const variants = {
    default: 'bg-secondary text-secondary-foreground',
    success: 'bg-green-500/10 text-green-500',
    warning: 'bg-yellow-500/10 text-yellow-500',
    error: 'bg-red-500/10 text-red-500',
    info: 'bg-blue-500/10 text-blue-500',
  }
  return (
    <span className={cn('inline-flex items-center px-2 py-0.5 rounded text-xs font-medium', variants[variant])}>
      {children}
    </span>
  )
}

function Dialog({
  isOpen,
  onClose,
  title,
  children,
  maxWidth = 'md',
}: {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl'
}) {
  if (!isOpen) return null
  const maxWidths = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-lg', xl: 'max-w-xl' }
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div
        className={cn('bg-background border border-border rounded-lg shadow-lg flex flex-col w-full', maxWidths[maxWidth])}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h3 className="text-lg font-semibold">{title}</h3>
          <button onClick={onClose} className="p-1.5 hover:bg-secondary rounded-md transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-4 overflow-y-auto max-h-[70vh]">{children}</div>
      </div>
    </div>
  )
}

// ── Passphrase Dialog ──────────────────────────────

function PassphraseDialog({
  isOpen,
  onClose,
  onSubmit,
  keyName,
  mode,
}: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (passphrase: string) => void
  keyName: string
  mode: 'set' | 'change' | 'unlock' | 'remove'
}) {
  const [passphrase, setPassphrase] = useState('')
  const [confirmPassphrase, setConfirmPassphrase] = useState('')
  const [showPassphrase, setShowPassphrase] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (mode === 'set' || mode === 'change') {
      if (passphrase.length < 8) {
        setError('Passphrase must be at least 8 characters')
        return
      }
      if (passphrase !== confirmPassphrase) {
        setError('Passphrases do not match')
        return
      }
    }

    onSubmit(passphrase)
    setPassphrase('')
    setConfirmPassphrase('')
    onClose()
  }

  const titles = {
    set: 'Set Passphrase',
    change: 'Change Passphrase',
    unlock: 'Unlock Key',
    remove: 'Remove Passphrase',
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title={titles[mode]} maxWidth="sm">
      <form onSubmit={handleSubmit} className="space-y-4">
        <p className="text-sm text-muted-foreground">
          {mode === 'unlock'
            ? `Enter passphrase to unlock "${keyName}"`
            : mode === 'remove'
            ? `Enter current passphrase to remove protection from "${keyName}"`
            : `Set a passphrase for "${keyName}" to encrypt the private key`}
        </p>

        <div>
          <label className="text-sm font-medium mb-1.5 block">
            {mode === 'change' ? 'Current Passphrase' : mode === 'remove' ? 'Current Passphrase' : 'Passphrase'}
          </label>
          <div className="relative">
            <input
              type={showPassphrase ? 'text' : 'password'}
              value={passphrase}
              onChange={(e) => setPassphrase(e.target.value)}
              className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm pr-10"
              placeholder={mode === 'unlock' ? 'Enter passphrase...' : 'Min 8 characters...'}
              autoFocus
            />
            <button
              type="button"
              onClick={() => setShowPassphrase(!showPassphrase)}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              {showPassphrase ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>

        {(mode === 'set' || mode === 'change') && (
          <div>
            <label className="text-sm font-medium mb-1.5 block">Confirm Passphrase</label>
            <input
              type="password"
              value={confirmPassphrase}
              onChange={(e) => setConfirmPassphrase(e.target.value)}
              className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm"
              placeholder="Confirm passphrase..."
            />
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 text-sm text-red-500">
            <AlertCircle className="w-4 h-4" />
            {error}
          </div>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
          >
            {mode === 'unlock' ? 'Unlock' : mode === 'remove' ? 'Remove' : 'Set Passphrase'}
          </button>
        </div>
      </form>
    </Dialog>
  )
}

// ── Key Generation Dialog ──────────────────────────

function KeyGenerationDialog({
  isOpen,
  onClose,
  onGenerate,
}: {
  isOpen: boolean
  onClose: () => void
  onGenerate: (key: Omit<SSHKey, 'id' | 'fingerprint' | 'fingerprintSHA256' | 'createdAt' | 'updatedAt' | 'publicKey' | 'privateKey'>) => void
}) {
  const [name, setName] = useState('')
  const [algorithm, setAlgorithm] = useState<KeyAlgorithm>('ed25519')
  const [bits, setBits] = useState(256)
  const [comment, setComment] = useState('')
  const [setPassphrase, setSetPassphrase] = useState(true)
  const [isGenerating, setIsGenerating] = useState(false)

  const selectedAlgo = ALGORITHM_OPTIONS.find((a) => a.value === algorithm)!

  const handleGenerate = async () => {
    if (!name.trim()) return
    setIsGenerating(true)

    // Simulate key generation delay
    await new Promise((resolve) => setTimeout(resolve, 1500))

    onGenerate({
      name: name.trim(),
      algorithm,
      bits,
      hasPassphrase: setPassphrase,
      status: 'active',
      comment: comment.trim() || undefined,
      associatedHosts: [],
      tags: [],
    })

    setIsGenerating(false)
    setName('')
    setComment('')
    setAlgorithm('ed25519')
    setBits(256)
    setSetPassphrase(true)
    onClose()
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title="Generate New SSH Key" maxWidth="md">
      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-1.5 block">Key Name *</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm"
            placeholder="e.g., Production Server Key"
            autoFocus
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-1.5 block">Algorithm</label>
          <div className="grid grid-cols-3 gap-2">
            {ALGORITHM_OPTIONS.map((algo) => (
              <button
                key={algo.value}
                onClick={() => {
                  setAlgorithm(algo.value)
                  setBits(algo.defaultBits)
                }}
                className={cn(
                  'px-3 py-2 text-sm rounded-md border transition-colors',
                  algorithm === algo.value
                    ? 'border-primary bg-primary/10 text-foreground'
                    : 'border-border hover:bg-secondary/50 text-muted-foreground'
                )}
              >
                <div className="font-medium">{algo.label}</div>
                <div className="text-xs opacity-70">
                  {algo.value === 'ed25519' ? 'Recommended' : algo.value === 'rsa' ? 'Widely supported' : 'Fast signing'}
                </div>
              </button>
            ))}
          </div>
        </div>

        {selectedAlgo.bits.length > 1 && (
          <div>
            <label className="text-sm font-medium mb-1.5 block">Key Size (bits)</label>
            <div className="flex gap-2">
              {selectedAlgo.bits.map((b) => (
                <button
                  key={b}
                  onClick={() => setBits(b)}
                  className={cn(
                    'px-3 py-1.5 text-sm rounded-md border transition-colors',
                    bits === b
                      ? 'border-primary bg-primary/10 text-foreground'
                      : 'border-border hover:bg-secondary/50 text-muted-foreground'
                  )}
                >
                  {b}
                </button>
              ))}
            </div>
          </div>
        )}

        <div>
          <label className="text-sm font-medium mb-1.5 block">Comment (optional)</label>
          <input
            type="text"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm"
            placeholder="e.g., user@hostname"
          />
        </div>

        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={setPassphrase}
            onChange={(e) => setSetPassphrase(e.target.checked)}
            className="rounded"
          />
          <span className="text-sm">Protect with passphrase (recommended)</span>
        </label>

        {isGenerating && (
          <div className="flex items-center gap-3 p-3 bg-secondary/50 rounded-md">
            <RefreshCw className="w-5 h-5 animate-spin text-primary" />
            <div>
              <div className="text-sm font-medium">Generating key...</div>
              <div className="text-xs text-muted-foreground">This may take a moment for RSA keys</div>
            </div>
          </div>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
            disabled={isGenerating}
          >
            Cancel
          </button>
          <button
            onClick={handleGenerate}
            disabled={!name.trim() || isGenerating}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isGenerating ? 'Generating...' : 'Generate Key'}
          </button>
        </div>
      </div>
    </Dialog>
  )
}

// ── Key Import Dialog ──────────────────────────────

function KeyImportDialog({
  isOpen,
  onClose,
  onImport,
}: {
  isOpen: boolean
  onClose: () => void
  onImport: (key: Partial<SSHKey>) => void
}) {
  const [name, setName] = useState('')
  const [privateKeyContent, setPrivateKeyContent] = useState('')
  const [publicKeyContent, setPublicKeyContent] = useState('')
  const [importMode, setImportMode] = useState<'both' | 'private' | 'public'>('both')
  const [isImporting, setIsImporting] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFileRead(file)
  }

  const handleFileRead = (file: File) => {
    const reader = new FileReader()
    reader.onload = (event) => {
      const content = event.target?.result as string
      if (content.includes('PRIVATE KEY')) {
        setPrivateKeyContent(content)
        if (!name) setName(file.name.replace(/\.(pem|key|ppk)$/, ''))
      } else if (content.includes('ssh-')) {
        setPublicKeyContent(content)
        if (!name) setName(file.name.replace(/\.pub$/, ''))
      }
    }
    reader.readAsText(file)
  }

  const handleImport = async () => {
    if (!name.trim()) return
    setIsImporting(true)

    // Simulate import processing
    await new Promise((resolve) => setTimeout(resolve, 800))

    // Detect algorithm from key content
    let algorithm: KeyAlgorithm = 'rsa'
    if (privateKeyContent.includes('OPENSSH PRIVATE KEY') && privateKeyContent.includes('ed25519')) {
      algorithm = 'ed25519'
    } else if (privateKeyContent.includes('EC PRIVATE KEY') || privateKeyContent.includes('ecdsa')) {
      algorithm = 'ecdsa'
    }

    onImport({
      name: name.trim(),
      algorithm,
      hasPassphrase: false, // Will be detected on first use
      status: 'active',
      publicKey: publicKeyContent || undefined,
      privateKey: privateKeyContent || undefined,
      associatedHosts: [],
      tags: [],
    })

    setIsImporting(false)
    setName('')
    setPrivateKeyContent('')
    setPublicKeyContent('')
    onClose()
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title="Import SSH Key" maxWidth="lg">
      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-1.5 block">Key Name *</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm"
            placeholder="e.g., My Imported Key"
            autoFocus
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-1.5 block">Import Mode</label>
          <div className="flex gap-2">
            {[
              { value: 'both' as const, label: 'Key Pair', desc: 'Private + Public' },
              { value: 'private' as const, label: 'Private Only', desc: 'Will extract public' },
              { value: 'public' as const, label: 'Public Only', desc: 'For verification' },
            ].map((mode) => (
              <button
                key={mode.value}
                onClick={() => setImportMode(mode.value)}
                className={cn(
                  'flex-1 px-3 py-2 text-sm rounded-md border transition-colors text-left',
                  importMode === mode.value
                    ? 'border-primary bg-primary/10 text-foreground'
                    : 'border-border hover:bg-secondary/50 text-muted-foreground'
                )}
              >
                <div className="font-medium">{mode.label}</div>
                <div className="text-xs opacity-70">{mode.desc}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Drag & Drop Area */}
        <div
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleFileDrop}
          onClick={() => fileInputRef.current?.click()}
          className={cn(
            'border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors',
            dragOver
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-muted-foreground hover:bg-secondary/30'
          )}
        >
          <Upload className="w-8 h-8 mx-auto mb-2 text-muted-foreground" />
          <div className="text-sm font-medium">Drop key files here or click to browse</div>
          <div className="text-xs text-muted-foreground mt-1">Supports .pem, .key, .pub files</div>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            accept=".pem,.key,.pub,.ppk"
            onChange={(e) => e.target.files?.[0] && handleFileRead(e.target.files[0])}
          />
        </div>

        {(importMode === 'both' || importMode === 'private') && (
          <div>
            <label className="text-sm font-medium mb-1.5 block">
              Private Key {importMode === 'private' && '*'}
            </label>
            <textarea
              value={privateKeyContent}
              onChange={(e) => setPrivateKeyContent(e.target.value)}
              className="w-full h-32 px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm font-mono text-xs resize-none"
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----\n..."
            />
          </div>
        )}

        {(importMode === 'both' || importMode === 'public') && (
          <div>
            <label className="text-sm font-medium mb-1.5 block">
              Public Key {importMode === 'public' && '*'}
            </label>
            <textarea
              value={publicKeyContent}
              onChange={(e) => setPublicKeyContent(e.target.value)}
              className="w-full h-20 px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm font-mono text-xs resize-none"
              placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... comment"
            />
          </div>
        )}

        {isImporting && (
          <div className="flex items-center gap-3 p-3 bg-secondary/50 rounded-md">
            <RefreshCw className="w-5 h-5 animate-spin text-primary" />
            <span className="text-sm">Processing key...</span>
          </div>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
            disabled={isImporting}
          >
            Cancel
          </button>
          <button
            onClick={handleImport}
            disabled={!name.trim() || (!privateKeyContent && !publicKeyContent) || isImporting}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isImporting ? 'Importing...' : 'Import Key'}
          </button>
        </div>
      </div>
    </Dialog>
  )
}

// ── Host Association Dialog ────────────────────────

function HostAssociationDialog({
  isOpen,
  onClose,
  keyData,
  hosts,
  onUpdate,
}: {
  isOpen: boolean
  onClose: () => void
  keyData: SSHKey | null
  hosts: HostRef[]
  onUpdate: (hostIds: string[]) => void
}) {
  const [selectedHosts, setSelectedHosts] = useState<string[]>([])

  // Initialize selected hosts when dialog opens
  useState(() => {
    if (keyData) setSelectedHosts(keyData.associatedHosts)
  })

  if (!keyData) return null

  const toggleHost = (hostId: string) => {
    setSelectedHosts((prev) =>
      prev.includes(hostId) ? prev.filter((id) => id !== hostId) : [...prev, hostId]
    )
  }

  const handleSave = () => {
    onUpdate(selectedHosts)
    onClose()
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title={`Associate "${keyData.name}" with Hosts`} maxWidth="md">
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Select the hosts that should use this SSH key for authentication.
        </p>

        <div className="border border-border rounded-md divide-y divide-border">
          {hosts.length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">No hosts configured</div>
          ) : (
            hosts.map((host) => {
              const isSelected = selectedHosts.includes(host.id)
              return (
                <button
                  key={host.id}
                  onClick={() => toggleHost(host.id)}
                  className={cn(
                    'w-full flex items-center gap-3 px-4 py-3 text-left transition-colors',
                    isSelected ? 'bg-primary/5' : 'hover:bg-secondary/30'
                  )}
                >
                  <div
                    className={cn(
                      'w-5 h-5 rounded border flex items-center justify-center transition-colors',
                      isSelected ? 'bg-primary border-primary' : 'border-border'
                    )}
                  >
                    {isSelected && <Check className="w-3.5 h-3.5 text-primary-foreground" />}
                  </div>
                  <Server className="w-4 h-4 text-muted-foreground" />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{host.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {host.username}@{host.hostname}
                    </div>
                  </div>
                </button>
              )
            })
          )}
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
          >
            Save Associations
          </button>
        </div>
      </div>
    </Dialog>
  )
}

// ── Key Rotation Dialog ────────────────────────────

function KeyRotationDialog({
  isOpen,
  onClose,
  keyData,
  onRotate,
}: {
  isOpen: boolean
  onClose: () => void
  keyData: SSHKey | null
  onRotate: (oldKeyId: string, newKey: Omit<SSHKey, 'id' | 'fingerprint' | 'fingerprintSHA256' | 'createdAt' | 'updatedAt' | 'publicKey' | 'privateKey'>) => void
}) {
  const [step, setStep] = useState(1)
  const [newKeyName, setNewKeyName] = useState('')
  const [algorithm, setAlgorithm] = useState<KeyAlgorithm>('ed25519')
  const [isRotating, setIsRotating] = useState(false)

  if (!keyData) return null

  const handleStartRotation = () => {
    setNewKeyName(`${keyData.name} (Rotated)`)
    setAlgorithm(keyData.algorithm)
    setStep(2)
  }

  const handleCompleteRotation = async () => {
    setIsRotating(true)
    await new Promise((resolve) => setTimeout(resolve, 2000))

    onRotate(keyData.id, {
      name: newKeyName,
      algorithm,
      hasPassphrase: true,
      status: 'active',
      associatedHosts: keyData.associatedHosts,
      tags: [...keyData.tags, 'rotated'],
    })

    setIsRotating(false)
    setStep(1)
    onClose()
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title="Key Rotation" maxWidth="md">
      {step === 1 && (
        <div className="space-y-4">
          <div className="flex items-start gap-3 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-md">
            <AlertCircle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
            <div className="text-sm">
              <div className="font-medium text-yellow-500">Rotation Warning</div>
              <div className="text-muted-foreground mt-1">
                Key rotation will generate a new key pair and update all associated hosts.
                The old key will be marked as rotating until all hosts are updated.
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <div className="text-sm font-medium">Key to Rotate</div>
            <div className="p-3 bg-secondary/50 rounded-md">
              <div className="text-sm font-medium">{keyData.name}</div>
              <div className="text-xs text-muted-foreground mt-1">
                {keyData.algorithm.toUpperCase()} {keyData.bits && `(${keyData.bits} bits)`} ·{' '}
                {keyData.associatedHosts.length} associated host
                {keyData.associatedHosts.length !== 1 ? 's' : ''}
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleStartRotation}
              className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
            >
              Start Rotation
            </button>
          </div>
        </div>
      )}

      {step === 2 && (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">New Key Name</label>
            <input
              type="text"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              className="w-full px-3 py-2 bg-secondary/50 border border-border rounded-md text-sm"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">New Algorithm</label>
            <div className="grid grid-cols-3 gap-2">
              {ALGORITHM_OPTIONS.map((algo) => (
                <button
                  key={algo.value}
                  onClick={() => setAlgorithm(algo.value)}
                  className={cn(
                    'px-3 py-2 text-sm rounded-md border transition-colors',
                    algorithm === algo.value
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:bg-secondary/50'
                  )}
                >
                  {algo.label}
                </button>
              ))}
            </div>
          </div>

          {isRotating && (
            <div className="space-y-2">
              <div className="flex items-center gap-3 p-3 bg-secondary/50 rounded-md">
                <RefreshCw className="w-5 h-5 animate-spin text-primary" />
                <div className="text-sm">
                  <div className="font-medium">Rotating key...</div>
                  <div className="text-xs text-muted-foreground">Generating new key pair and updating hosts</div>
                </div>
              </div>
              <div className="w-full bg-secondary rounded-full h-2">
                <div className="bg-primary h-2 rounded-full animate-pulse w-3/4" />
              </div>
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <button
              onClick={() => setStep(1)}
              className="px-4 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
              disabled={isRotating}
            >
              Back
            </button>
            <button
              onClick={handleCompleteRotation}
              disabled={isRotating || !newKeyName.trim()}
              className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50"
            >
              {isRotating ? 'Rotating...' : 'Complete Rotation'}
            </button>
          </div>
        </div>
      )}
    </Dialog>
  )
}

// ── Key Details Panel ──────────────────────────────

function KeyDetailsPanel({ keyData, onClose }: { keyData: SSHKey; onClose: () => void }) {
  const [showPrivateKey, setShowPrivateKey] = useState(false)
  const [showPublicKey, setShowPublicKey] = useState(false)

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    // In Tauri app, use: invoke('write_clipboard', { text })
  }

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold">{keyData.name}</h3>
          <div className="flex items-center gap-2 mt-1">
            <Badge variant={keyData.status === 'active' ? 'success' : keyData.status === 'rotating' ? 'warning' : 'error'}>
              {STATUS_LABELS[keyData.status]}
            </Badge>
            <span className="text-xs text-muted-foreground">
              {keyData.algorithm.toUpperCase()}
              {keyData.bits && ` · ${keyData.bits} bits`}
            </span>
          </div>
        </div>
        <button onClick={onClose} className="p-1.5 hover:bg-secondary rounded-md transition-colors">
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Fingerprint */}
      <div className="space-y-2">
        <div className="text-sm font-medium flex items-center gap-2">
          <Fingerprint className="w-4 h-4" />
          Fingerprint
        </div>
        <div className="p-2 bg-secondary/50 rounded-md font-mono text-xs break-all">
          {keyData.fingerprintSHA256}
        </div>
        <div className="p-2 bg-secondary/50 rounded-md font-mono text-xs break-all text-muted-foreground">
          {keyData.fingerprint}
        </div>
      </div>

      {/* Public Key */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="text-sm font-medium flex items-center gap-2">
            <KeyRound className="w-4 h-4" />
            Public Key
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => copyToClipboard(keyData.publicKey)}
              className="p-1.5 hover:bg-secondary rounded-md transition-colors"
              title="Copy public key"
            >
              <Copy className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setShowPublicKey(!showPublicKey)}
              className="p-1.5 hover:bg-secondary rounded-md transition-colors"
              title={showPublicKey ? 'Hide' : 'Show'}
            >
              {showPublicKey ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
            </button>
          </div>
        </div>
        <div className="p-2 bg-secondary/50 rounded-md font-mono text-xs break-all">
          {showPublicKey ? keyData.publicKey : 'ssh-' + keyData.algorithm + ' AAA... ' + (keyData.comment || keyData.name)}
        </div>
      </div>

      {/* Private Key */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="text-sm font-medium flex items-center gap-2">
            <Lock className="w-4 h-4" />
            Private Key
            {keyData.hasPassphrase && (
              <Badge variant="info">
                <Lock className="w-3 h-3 mr-1" />
                Encrypted
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => copyToClipboard(keyData.privateKey)}
              className="p-1.5 hover:bg-secondary rounded-md transition-colors"
              title="Copy private key"
            >
              <Copy className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setShowPrivateKey(!showPrivateKey)}
              className="p-1.5 hover:bg-secondary rounded-md transition-colors"
              title={showPrivateKey ? 'Hide' : 'Show'}
            >
              {showPrivateKey ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
            </button>
          </div>
        </div>
        <div className="p-2 bg-secondary/50 rounded-md font-mono text-xs break-all">
          {showPrivateKey
            ? keyData.privateKey
            : '-----BEGIN OPENSSH PRIVATE KEY-----\n[REDACTED]\n-----END OPENSSH PRIVATE KEY-----'}
        </div>
      </div>

      {/* Metadata */}
      <div className="grid grid-cols-2 gap-3 text-sm">
        <div>
          <div className="text-muted-foreground text-xs">Created</div>
          <div>{new Date(keyData.createdAt).toLocaleDateString()}</div>
        </div>
        <div>
          <div className="text-muted-foreground text-xs">Last Updated</div>
          <div>{new Date(keyData.updatedAt).toLocaleDateString()}</div>
        </div>
        {keyData.lastUsedAt && (
          <div>
            <div className="text-muted-foreground text-xs">Last Used</div>
            <div>{new Date(keyData.lastUsedAt).toLocaleDateString()}</div>
          </div>
        )}
        <div>
          <div className="text-muted-foreground text-xs">Associated Hosts</div>
          <div>{keyData.associatedHosts.length}</div>
        </div>
      </div>

      {/* Tags */}
      {keyData.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {keyData.tags.map((tag) => (
            <Badge key={tag} variant="default">
              {tag}
            </Badge>
          ))}
        </div>
      )}
    </div>
  )
}

// ── Main KeyManager Component ──────────────────────

export function KeyManager({ isOpen, onClose }: KeyManagerProps) {
  const [keys, setKeys] = useState<SSHKey[]>(mockKeys)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterAlgorithm, setFilterAlgorithm] = useState<KeyAlgorithm | 'all'>('all')
  const [selectedKey, setSelectedKey] = useState<SSHKey | null>(null)
  const [viewMode, setViewMode] = useState<'list' | 'details'>('list')

  // Dialog states
  const [showGenerateDialog, setShowGenerateDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [showPassphraseDialog, setShowPassphraseDialog] = useState(false)
  const [showHostAssocDialog, setShowHostAssocDialog] = useState(false)
  const [showRotationDialog, setShowRotationDialog] = useState(false)
  const [passphraseMode, setPassphraseMode] = useState<'set' | 'change' | 'unlock' | 'remove'>('set')
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; keyId: string } | null>(null)

  // Filter keys
  const filteredKeys = keys.filter((key) => {
    const matchesSearch =
      searchQuery === '' ||
      key.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.comment?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.tags.some((t) => t.toLowerCase().includes(searchQuery.toLowerCase()))
    const matchesAlgo = filterAlgorithm === 'all' || key.algorithm === filterAlgorithm
    return matchesSearch && matchesAlgo
  })

  const handleGenerateKey = useCallback(
    (keyData: Omit<SSHKey, 'id' | 'fingerprint' | 'fingerprintSHA256' | 'createdAt' | 'updatedAt' | 'publicKey' | 'privateKey'>) => {
      const now = new Date().toISOString()
      const newKey: SSHKey = {
        ...keyData,
        id: `key-${Date.now()}`,
        fingerprint: `MD5:${Math.random().toString(36).substring(2, 18)}`,
        fingerprintSHA256: `SHA256:${Math.random().toString(36).substring(2, 42)}`,
        createdAt: now,
        updatedAt: now,
        publicKey: `ssh-${keyData.algorithm} AAA... ${keyData.comment || keyData.name}`,
        privateKey: '-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----',
      }
      setKeys((prev) => [...prev, newKey])
    },
    []
  )

  const handleImportKey = useCallback((keyData: Partial<SSHKey>) => {
    const now = new Date().toISOString()
    const newKey: SSHKey = {
      id: `key-${Date.now()}`,
      name: keyData.name || 'Imported Key',
      algorithm: keyData.algorithm || 'rsa',
      bits: keyData.bits,
      fingerprint: `MD5:${Math.random().toString(36).substring(2, 18)}`,
      fingerprintSHA256: `SHA256:${Math.random().toString(36).substring(2, 42)}`,
      createdAt: now,
      updatedAt: now,
      hasPassphrase: keyData.hasPassphrase || false,
      status: 'active',
      publicKey: keyData.publicKey || '',
      privateKey: keyData.privateKey || '',
      comment: keyData.comment,
      associatedHosts: keyData.associatedHosts || [],
      tags: keyData.tags || ['imported'],
    }
    setKeys((prev) => [...prev, newKey])
  }, [])

  const handleDeleteKey = useCallback((keyId: string) => {
    setKeys((prev) => prev.filter((k) => k.id !== keyId))
    if (selectedKey?.id === keyId) {
      setSelectedKey(null)
      setViewMode('list')
    }
  }, [selectedKey])

  const handleUpdateHosts = useCallback((keyId: string, hostIds: string[]) => {
    setKeys((prev) =>
      prev.map((k) =>
        k.id === keyId ? { ...k, associatedHosts: hostIds, updatedAt: new Date().toISOString() } : k
      )
    )
  }, [])

  const handleRotateKey = useCallback(
    (oldKeyId: string, newKeyData: Omit<SSHKey, 'id' | 'fingerprint' | 'fingerprintSHA256' | 'createdAt' | 'updatedAt' | 'publicKey' | 'privateKey'>) => {
      const now = new Date().toISOString()

      // Mark old key as expired
      setKeys((prev) =>
        prev.map((k) =>
          k.id === oldKeyId ? { ...k, status: 'expired' as KeyStatus, updatedAt: now } : k
        )
      )

      // Create new key
      const newKey: SSHKey = {
        ...newKeyData,
        id: `key-${Date.now()}`,
        fingerprint: `MD5:${Math.random().toString(36).substring(2, 18)}`,
        fingerprintSHA256: `SHA256:${Math.random().toString(36).substring(2, 42)}`,
        createdAt: now,
        updatedAt: now,
        publicKey: `ssh-${newKeyData.algorithm} AAA... ${newKeyData.comment || newKeyData.name}`,
        privateKey: '-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----',
      }
      setKeys((prev) => [...prev, newKey])
    },
    []
  )

  const openPassphraseDialog = (mode: 'set' | 'change' | 'unlock' | 'remove', key: SSHKey) => {
    setSelectedKey(key)
    setPassphraseMode(mode)
    setShowPassphraseDialog(true)
  }

  const openHostAssociation = (key: SSHKey) => {
    setSelectedKey(key)
    setShowHostAssocDialog(true)
  }

  const openRotation = (key: SSHKey) => {
    setSelectedKey(key)
    setShowRotationDialog(true)
  }

  const viewKeyDetails = (key: SSHKey) => {
    setSelectedKey(key)
    setViewMode('details')
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-background border border-border rounded-lg shadow-lg flex flex-col w-full max-w-4xl h-[80vh]">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border flex-shrink-0">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Key className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h2 className="text-lg font-semibold">SSH Key Manager</h2>
              <p className="text-xs text-muted-foreground">{keys.length} keys · {keys.filter((k) => k.status === 'active').length} active</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowImportDialog(true)}
              className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-secondary rounded-md transition-colors"
            >
              <Upload className="w-4 h-4" />
              Import
            </button>
            <button
              onClick={() => setShowGenerateDialog(true)}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Generate
            </button>
            <button onClick={onClose} className="p-2 hover:bg-secondary rounded-md transition-colors ml-2">
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Toolbar */}
        <div className="flex items-center gap-3 p-3 border-b border-border flex-shrink-0">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search keys..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 bg-secondary/50 border border-border rounded-md text-sm"
            />
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setFilterAlgorithm('all')}
              className={cn(
                'px-2.5 py-1.5 text-xs rounded-md transition-colors',
                filterAlgorithm === 'all' ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50'
              )}
            >
              All
            </button>
            {(['rsa', 'ed25519', 'ecdsa'] as KeyAlgorithm[]).map((algo) => (
              <button
                key={algo}
                onClick={() => setFilterAlgorithm(algo)}
                className={cn(
                  'px-2.5 py-1.5 text-xs rounded-md transition-colors uppercase',
                  filterAlgorithm === algo ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50'
                )}
              >
                {algo}
              </button>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className="flex flex-1 overflow-hidden">
          {viewMode === 'list' ? (
            <div className="flex-1 overflow-y-auto">
              {filteredKeys.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                  <Key className="w-12 h-12 mb-4 opacity-30" />
                  <p className="text-sm">No keys found</p>
                  <p className="text-xs mt-1">Generate or import a new SSH key to get started</p>
                </div>
              ) : (
                <div className="divide-y divide-border">
                  {filteredKeys.map((key) => (
                    <div
                      key={key.id}
                      className="flex items-center gap-3 px-4 py-3 hover:bg-secondary/30 transition-colors group"
                      onContextMenu={(e) => {
                        e.preventDefault()
                        setContextMenu({ x: e.clientX, y: e.clientY, keyId: key.id })
                      }}
                    >
                      {/* Icon */}
                      <div className="flex-shrink-0">
                        <div
                          className={cn(
                            'w-10 h-10 rounded-lg flex items-center justify-center',
                            key.algorithm === 'ed25519'
                              ? 'bg-green-500/10'
                              : key.algorithm === 'rsa'
                              ? 'bg-blue-500/10'
                              : 'bg-purple-500/10'
                          )}
                        >
                          <FileKey
                            className={cn(
                              'w-5 h-5',
                              key.algorithm === 'ed25519'
                                ? 'text-green-500'
                                : key.algorithm === 'rsa'
                                ? 'text-blue-500'
                                : 'text-purple-500'
                            )}
                          />
                        </div>
                      </div>

                      {/* Info */}
                      <div className="flex-1 min-w-0 cursor-pointer" onClick={() => viewKeyDetails(key)}>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium truncate">{key.name}</span>
                          {key.hasPassphrase && <Lock className="w-3 h-3 text-muted-foreground" />}
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                          <span className={STATUS_COLORS[key.status]}>● {STATUS_LABELS[key.status]}</span>
                          <span>·</span>
                          <span className="uppercase">{key.algorithm}</span>
                          {key.bits && <span>({key.bits})</span>}
                          <span>·</span>
                          <span>{key.associatedHosts.length} host{key.associatedHosts.length !== 1 ? 's' : ''}</span>
                        </div>
                      </div>

                      {/* Fingerprint */}
                      <div className="hidden md:block flex-shrink-0">
                        <div className="text-xs font-mono text-muted-foreground">
                          {key.fingerprintSHA256.slice(0, 20)}...
                        </div>
                      </div>

                      {/* Actions */}
                      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => openPassphraseDialog(key.hasPassphrase ? 'change' : 'set', key)}
                          className="p-1.5 hover:bg-secondary rounded-md transition-colors"
                          title={key.hasPassphrase ? 'Change passphrase' : 'Set passphrase'}
                        >
                          {key.hasPassphrase ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                        </button>
                        <button
                          onClick={() => openHostAssociation(key)}
                          className="p-1.5 hover:bg-secondary rounded-md transition-colors"
                          title="Associate with hosts"
                        >
                          <Server className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => openRotation(key)}
                          className="p-1.5 hover:bg-secondary rounded-md transition-colors"
                          title="Rotate key"
                        >
                          <RotateCcw className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => handleDeleteKey(key.id)}
                          className="p-1.5 hover:bg-red-500/10 text-red-500 rounded-md transition-colors"
                          title="Delete key"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            <div className="flex-1 overflow-y-auto p-4">
              {selectedKey && (
                <KeyDetailsPanel
                  keyData={selectedKey}
                  onClose={() => {
                    setViewMode('list')
                    setSelectedKey(null)
                  }}
                />
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-border text-xs text-muted-foreground flex-shrink-0">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <Shield className="w-3 h-3" />
              Keys are encrypted at rest
            </span>
            <span className="flex items-center gap-1">
              <Key className="w-3 h-3" />
              Ed25519 recommended for new keys
            </span>
          </div>
          {viewMode === 'details' && (
            <button
              onClick={() => {
                setViewMode('list')
                setSelectedKey(null)
              }}
              className="flex items-center gap-1 hover:text-foreground transition-colors"
            >
              <ChevronDown className="w-3 h-3" />
              Back to list
            </button>
          )}
        </div>
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <div
          className="fixed bg-background border border-border rounded-md shadow-lg py-1 z-[60]"
          style={{ top: contextMenu.y, left: contextMenu.x }}
          onClick={() => setContextMenu(null)}
        >
          {(() => {
            const key = keys.find((k) => k.id === contextMenu.keyId)
            if (!key) return null
            return (
              <>
                <button
                  onClick={() => viewKeyDetails(key)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Eye className="w-3.5 h-3.5" /> View Details
                </button>
                <button
                  onClick={() => openPassphraseDialog(key.hasPassphrase ? 'change' : 'set', key)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Lock className="w-3.5 h-3.5" />
                  {key.hasPassphrase ? 'Change Passphrase' : 'Set Passphrase'}
                </button>
                <button
                  onClick={() => openHostAssociation(key)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Server className="w-3.5 h-3.5" /> Associate Hosts
                </button>
                <button
                  onClick={() => openRotation(key)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <RefreshCw className="w-3.5 h-3.5" /> Rotate Key
                </button>
                <div className="border-t border-border my-1" />
                <button
                  onClick={() => {
                    // Export functionality
                    if (typeof document === 'undefined' || typeof window === 'undefined') return;
                    const blob = new Blob([key.publicKey], { type: 'text/plain' })
                    const url = URL.createObjectURL(blob)
                    if (typeof document === 'undefined') return; const a = document.createElement('a')
                    a.href = url
                    a.download = `${key.name}.pub`
                    a.click()
                    URL.revokeObjectURL(url)
                  }}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Download className="w-3.5 h-3.5" /> Export Public Key
                </button>
                <div className="border-t border-border my-1" />
                <button
                  onClick={() => handleDeleteKey(key.id)}
                  className="w-full text-left px-3 py-1.5 text-sm hover:bg-secondary transition-colors text-red-500 flex items-center gap-2"
                >
                  <Trash2 className="w-3.5 h-3.5" /> Delete Key
                </button>
              </>
            )
          })()}
        </div>
      )}

      {/* Sub-Dialogs */}
      <KeyGenerationDialog
        isOpen={showGenerateDialog}
        onClose={() => setShowGenerateDialog(false)}
        onGenerate={handleGenerateKey}
      />

      <KeyImportDialog
        isOpen={showImportDialog}
        onClose={() => setShowImportDialog(false)}
        onImport={handleImportKey}
      />

      <PassphraseDialog
        isOpen={showPassphraseDialog}
        onClose={() => setShowPassphraseDialog(false)}
        onSubmit={(passphrase) => {
          if (selectedKey) {
            setKeys((prev) =>
              prev.map((k) =>
                k.id === selectedKey.id
                  ? {
                      ...k,
                      hasPassphrase: passphraseMode !== 'remove',
                      updatedAt: new Date().toISOString(),
                    }
                  : k
              )
            )
          }
        }}
        keyName={selectedKey?.name || ''}
        mode={passphraseMode}
      />

      <HostAssociationDialog
        isOpen={showHostAssocDialog}
        onClose={() => setShowHostAssocDialog(false)}
        keyData={selectedKey}
        hosts={mockHosts}
        onUpdate={(hostIds) => {
          if (selectedKey) handleUpdateHosts(selectedKey.id, hostIds)
        }}
      />

      <KeyRotationDialog
        isOpen={showRotationDialog}
        onClose={() => setShowRotationDialog(false)}
        keyData={selectedKey}
        onRotate={handleRotateKey}
      />
    </div>
  )
}
