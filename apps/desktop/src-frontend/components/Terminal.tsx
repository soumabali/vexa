'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import type { Terminal as XTermType, ITerminalAddon } from 'xterm'

interface TerminalProps {
  hostId: string
}

// Dynamically import xterm only on client side
let xtermModule: typeof import('xterm') | null = null
let xtermAddons: {
  FitAddon: typeof import('xterm-addon-fit').FitAddon
  WebLinksAddon: typeof import('xterm-addon-web-links').WebLinksAddon
  SearchAddon: typeof import('xterm-addon-search').SearchAddon
  Unicode11Addon: typeof import('xterm-addon-unicode11').Unicode11Addon
} | null = null

async function loadXTerm() {
  if (!xtermModule) {
    xtermModule = await import('xterm')
  }
  if (!xtermAddons) {
    const [fit, webLinks, search, unicode11] = await Promise.all([
      import('xterm-addon-fit'),
      import('xterm-addon-web-links'),
      import('xterm-addon-search'),
      import('xterm-addon-unicode11'),
    ])
    xtermAddons = {
      FitAddon: fit.FitAddon,
      WebLinksAddon: webLinks.WebLinksAddon,
      SearchAddon: search.SearchAddon,
      Unicode11Addon: unicode11.Unicode11Addon,
    }
  }
  return { xtermModule, xtermAddons }
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type ConnectionState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected' | 'error'

interface TerminalTheme {
  background: string
  foreground: string
  cursor: string
  cursorAccent: string
  selectionBackground: string
  selectionForeground: string
  black: string
  red: string
  green: string
  yellow: string
  blue: string
  magenta: string
  cyan: string
  white: string
  brightBlack: string
  brightRed: string
  brightGreen: string
  brightYellow: string
  brightBlue: string
  brightMagenta: string
  brightCyan: string
  brightWhite: string
}

const themes: Record<string, TerminalTheme> = {
  dark: {
    background: '#0c0c0c',
    foreground: '#cccccc',
    cursor: '#ffffff',
    cursorAccent: '#000000',
    selectionBackground: '#264f78',
    selectionForeground: '#ffffff',
    black: '#0c0c0c',
    red: '#c50f1f',
    green: '#13a10e',
    yellow: '#c19c00',
    blue: '#0037da',
    magenta: '#881798',
    cyan: '#3a96dd',
    white: '#cccccc',
    brightBlack: '#767676',
    brightRed: '#e74856',
    brightGreen: '#16c60c',
    brightYellow: '#f9f1a5',
    brightBlue: '#3b78ff',
    brightMagenta: '#b4009e',
    brightCyan: '#61d6d6',
    brightWhite: '#f2f2f2',
  },
  light: {
    background: '#ffffff',
    foreground: '#333333',
    cursor: '#000000',
    cursorAccent: '#ffffff',
    selectionBackground: '#add6ff',
    selectionForeground: '#000000',
    black: '#000000',
    red: '#c21b00',
    green: '#338c00',
    yellow: '#b8860b',
    blue: '#0037da',
    magenta: '#881798',
    cyan: '#3a96dd',
    white: '#cccccc',
    brightBlack: '#767676',
    brightRed: '#e74856',
    brightGreen: '#16c60c',
    brightYellow: '#f9f1a5',
    brightBlue: '#3b78ff',
    brightMagenta: '#b4009e',
    brightCyan: '#61d6d6',
    brightWhite: '#f2f2f2',
  },
  dracula: {
    background: '#282a36',
    foreground: '#f8f8f2',
    cursor: '#f8f8f2',
    cursorAccent: '#282a36',
    selectionBackground: '#44475a',
    selectionForeground: '#f8f8f2',
    black: '#000000',
    red: '#ff5555',
    green: '#50fa7b',
    yellow: '#f1fa8c',
    blue: '#bd93f9',
    magenta: '#ff79c6',
    cyan: '#8be9fd',
    white: '#bfbfbf',
    brightBlack: '#4d4d4d',
    brightRed: '#ff6e67',
    brightGreen: '#5af78e',
    brightYellow: '#fffc67',
    brightBlue: '#caa9fa',
    brightMagenta: '#ff92d0',
    brightCyan: '#a5f3fe',
    brightWhite: '#e6e6e6',
  },
}

export function Terminal({ hostId }: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTermType | null>(null)
  const fitAddonRef = useRef<any>(null)
  const searchAddonRef = useRef<any>(null)
  const [connectionState, setConnectionState] = useState<ConnectionState>('idle')
  const [error, setError] = useState<string | null>(null)
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [theme, setTheme] = useState('dark')
  const [fontSize, setFontSize] = useState(14)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [showPortForward, setShowPortForward] = useState(false)
  const [localPort, setLocalPort] = useState('')
  const [remoteHost, setRemoteHost] = useState('localhost')
  const [remotePort, setRemotePort] = useState('')
  const [forwards, setForwards] = useState<Array<{ id: string; localPort: string; remoteHost: string; remotePort: string }>>([])
  const [isRecording, setIsRecording] = useState(false)
  const [recordingPath, setRecordingPath] = useState<string | null>(null)
  const [showMacroPanel, setShowMacroPanel] = useState(false)
  const [macros, setMacros] = useState<Array<{ id: string; name: string; commands: string[] }>>([])
  const [newMacroName, setNewMacroName] = useState('')
  const [isRecordingMacro, setIsRecordingMacro] = useState(false)
  const macroCommandsRef = useRef<string[]>([])

  // Initialize terminal
  useEffect(() => {
    if (typeof window === 'undefined') return
    if (!containerRef.current) return

    let isMounted = true
    let ws: WebSocket | null = null

    const initTerminal = async () => {
      try {
        const { xtermModule: xm, xtermAddons: xa } = await loadXTerm()
        if (!isMounted || !containerRef.current) return

        const { Terminal } = xm!
        const { FitAddon, WebLinksAddon, SearchAddon, Unicode11Addon } = xa!

        const term = new Terminal({
          theme: themes[theme],
          fontSize,
          fontFamily: 'JetBrains Mono, Fira Code, monospace',
          cursorStyle: 'block',
          cursorBlink: true,
          scrollback: 10000,
          allowProposedApi: true,
        })

        const fitAddon = new FitAddon()
        const searchAddon = new SearchAddon()
        const webLinksAddon = new WebLinksAddon()
        const unicode11Addon = new Unicode11Addon()

        term.loadAddon(fitAddon)
        term.loadAddon(searchAddon)
        term.loadAddon(webLinksAddon)
        term.loadAddon(unicode11Addon)

        term.open(containerRef.current)
        fitAddon.fit()

        xtermRef.current = term
        fitAddonRef.current = fitAddon
        searchAddonRef.current = searchAddon

        // Handle resize
        const resizeObserver = new ResizeObserver(() => {
          fitAddon.fit()
          if (ws?.readyState === WebSocket.OPEN && sessionId) {
            const { cols, rows } = term
            ws.send(JSON.stringify({ type: 'resize', sessionId, cols, rows }))
          }
        })
        resizeObserver.observe(containerRef.current)

        // Connect to WebSocket
        connectWebSocket(term)

        return () => {
          resizeObserver.disconnect()
        }
      } catch (err) {
        console.error('Failed to initialize terminal:', err)
        setError('Failed to initialize terminal')
      }
    }

    const connectWebSocket = (term: XTermType) => {
      setConnectionState('connecting')
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      ws = new WebSocket(`${protocol}//${window.location.host}/api/ssh`)
      wsRef.current = ws

      ws.onopen = () => {
        setConnectionState('connected')
        reconnectAttempts.current = 0
        ws!.send(JSON.stringify({ type: 'connect', hostId }))
      }

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data)
        switch (data.type) {
          case 'data':
            term.write(data.data)
            break
          case 'session':
            setSessionId(data.sessionId)
            break
          case 'error':
            setError(data.message)
            setConnectionState('error')
            break
          case 'connected':
            setConnectionState('connected')
            break
          case 'disconnected':
            setConnectionState('disconnected')
            break
        }
      }

      ws.onclose = () => {
        if (reconnectAttempts.current < maxReconnectAttempts) {
          setConnectionState('reconnecting')
          reconnectAttempts.current++
          setTimeout(() => connectWebSocket(term), 1000 * reconnectAttempts.current)
        } else {
          setConnectionState('disconnected')
        }
      }

      ws.onerror = () => {
        setConnectionState('error')
      }

      // Handle terminal input
      term.onData((data) => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'data', sessionId, data }))
        }
      })
    }

    initTerminal()

    return () => {
      isMounted = false
      if (ws) {
        ws.close()
      }
      if (xtermRef.current) {
        xtermRef.current.dispose()
        xtermRef.current = null
      }
    }
  }, [hostId, theme, fontSize])

  // Handle theme change
  const handleThemeChange = useCallback((newTheme: string) => {
    setTheme(newTheme)
    if (xtermRef.current) {
      xtermRef.current.options.theme = themes[newTheme]
    }
  }, [])

  // Handle font size change
  const handleFontSizeChange = useCallback((delta: number) => {
    setFontSize((prev) => {
      const newSize = Math.max(8, Math.min(24, prev + delta))
      if (xtermRef.current) {
        xtermRef.current.options.fontSize = newSize
      }
      return newSize
    })
  }, [])

  // Search
  const handleSearch = useCallback(() => {
    if (searchAddonRef.current) {
      searchAddonRef.current.findNext(searchQuery)
    }
  }, [searchQuery])

  const handleSearchPrevious = useCallback(() => {
    if (searchAddonRef.current) {
      searchAddonRef.current.findPrevious(searchQuery)
    }
  }, [searchQuery])

  // Port forwarding
  const handleAddForward = useCallback(() => {
    if (localPort && remotePort) {
      const newForward = {
        id: Date.now().toString(),
        localPort,
        remoteHost,
        remotePort,
      }
      setForwards((prev) => [...prev, newForward])
      // Send to backend
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'port-forward',
            sessionId,
            action: 'add',
            ...newForward,
          })
        )
      }
      setLocalPort('')
      setRemotePort('')
    }
  }, [localPort, remoteHost, remotePort, sessionId])

  const handleRemoveForward = useCallback((id: string) => {
    setForwards((prev) => prev.filter((f) => f.id !== id))
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          type: 'port-forward',
          sessionId,
          action: 'remove',
          forwardId: id,
        })
      )
    }
  }, [sessionId])

  // Session recording
  const toggleRecording = useCallback(async () => {
    if (isRecording) {
      // Stop recording
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'recording',
            sessionId,
            action: 'stop',
          })
        )
      }
      setIsRecording(false)
    } else {
      // Start recording
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'recording',
            sessionId,
            action: 'start',
          })
        )
      }
      setIsRecording(true)
    }
  }, [isRecording, sessionId])

  // Macros
  const startRecordingMacro = useCallback(() => {
    setIsRecordingMacro(true)
    macroCommandsRef.current = []
  }, [])

  const stopRecordingMacro = useCallback(() => {
    setIsRecordingMacro(false)
    if (newMacroName && macroCommandsRef.current.length > 0) {
      setMacros((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          name: newMacroName,
          commands: macroCommandsRef.current,
        },
      ])
      setNewMacroName('')
    }
  }, [newMacroName])

  const runMacro = useCallback((macroId: string) => {
    const macro = macros.find((m) => m.id === macroId)
    if (macro && wsRef.current?.readyState === WebSocket.OPEN) {
      macro.commands.forEach((cmd) => {
        wsRef.current!.send(
          JSON.stringify({
            type: 'data',
            sessionId,
            data: cmd + '\r',
          })
        )
      })
    }
  }, [macros, sessionId])

  const deleteMacro = useCallback((macroId: string) => {
    setMacros((prev) => prev.filter((m) => m.id !== macroId))
  }, [])

  // Copy/Paste
  const handleCopy = useCallback(() => {
    if (xtermRef.current) {
      const selection = xtermRef.current.getSelection()
      if (selection) {
        navigator.clipboard.writeText(selection)
      }
    }
  }, [])

  const handlePaste = useCallback(async () => {
    try {
      const text = await navigator.clipboard.readText()
      if (xtermRef.current && wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'data',
            sessionId,
            data: text,
          })
        )
      }
    } catch (err) {
      console.error('Failed to paste:', err)
    }
  }, [sessionId])

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-1.5 bg-secondary/50 border-b border-border">
        <div className="flex items-center gap-1">
          <span
            className={`w-2 h-2 rounded-full ${
              connectionState === 'connected'
                ? 'bg-green-500'
                : connectionState === 'connecting' || connectionState === 'reconnecting'
                ? 'bg-yellow-500'
                : connectionState === 'error'
                ? 'bg-red-500'
                : 'bg-gray-500'
            }`}
          />
          <span className="text-xs text-muted-foreground">
            {connectionState === 'connected'
              ? 'Connected'
              : connectionState === 'connecting'
              ? 'Connecting...'
              : connectionState === 'reconnecting'
              ? `Reconnecting (${reconnectAttempts.current}/${maxReconnectAttempts})...`
              : connectionState === 'error'
              ? 'Error'
              : 'Disconnected'}
          </span>
        </div>

        <div className="flex-1" />

        <div className="flex items-center gap-1">
          {/* Theme selector */}
          <select
            value={theme}
            onChange={(e) => handleThemeChange(e.target.value)}
            className="text-xs bg-background border border-border rounded px-2 py-1"
          >
            <option value="dark">Dark</option>
            <option value="light">Light</option>
            <option value="dracula">Dracula</option>
          </select>

          {/* Font size */}
          <button
            onClick={() => handleFontSizeChange(-1)}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
            title="Decrease font size"
          >
            A-
          </button>
          <span className="text-xs text-muted-foreground w-6 text-center">
            {fontSize}
          </span>
          <button
            onClick={() => handleFontSizeChange(1)}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
            title="Increase font size"
          >
            A+
          </button>

          {/* Search */}
          <button
            onClick={() => setIsSearchOpen(!isSearchOpen)}
            className={`text-xs px-2 py-1 rounded ${
              isSearchOpen ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
            }`}
            title="Search"
          >
            🔍
          </button>

          {/* Port forward */}
          <button
            onClick={() => setShowPortForward(!showPortForward)}
            className={`text-xs px-2 py-1 rounded ${
              showPortForward ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
            }`}
            title="Port forwarding"
          >
            🔄
          </button>

          {/* Recording */}
          <button
            onClick={toggleRecording}
            className={`text-xs px-2 py-1 rounded ${
              isRecording ? 'bg-red-500 text-white' : 'hover:bg-muted'
            }`}
            title={isRecording ? 'Stop recording' : 'Start recording'}
          >
            {isRecording ? '⏹️' : '⏺️'}
          </button>

          {/* Macros */}
          <button
            onClick={() => setShowMacroPanel(!showMacroPanel)}
            className={`text-xs px-2 py-1 rounded ${
              showMacroPanel ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
            }`}
            title="Macros"
          >
            ⚡
          </button>

          {/* Copy/Paste */}
          <button
            onClick={handleCopy}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
            title="Copy"
          >
            📋
          </button>
          <button
            onClick={handlePaste}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
            title="Paste"
          >
            📄
          </button>
        </div>
      </div>

      {/* Search bar */}
      {isSearchOpen && (
        <div className="flex items-center gap-2 px-3 py-1.5 bg-secondary/50 border-b border-border">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Search..."
            className="flex-1 text-xs bg-background border border-border rounded px-2 py-1"
          />
          <button
            onClick={handleSearchPrevious}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
          >
            ⬆️
          </button>
          <button
            onClick={handleSearch}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
          >
            ⬇️
          </button>
          <button
            onClick={() => setIsSearchOpen(false)}
            className="text-xs px-2 py-1 hover:bg-muted rounded"
          >
            ✕
          </button>
        </div>
      )}

      {/* Port forward panel */}
      {showPortForward && (
        <div className="px-3 py-2 bg-secondary/50 border-b border-border">
          <div className="flex items-center gap-2 mb-2">
            <input
              type="text"
              value={localPort}
              onChange={(e) => setLocalPort(e.target.value)}
              placeholder="Local port"
              className="w-20 text-xs bg-background border border-border rounded px-2 py-1"
            />
            <span className="text-xs text-muted-foreground">→</span>
            <input
              type="text"
              value={remoteHost}
              onChange={(e) => setRemoteHost(e.target.value)}
              placeholder="Remote host"
              className="w-28 text-xs bg-background border border-border rounded px-2 py-1"
            />
            <span className="text-xs text-muted-foreground">:</span>
            <input
              type="text"
              value={remotePort}
              onChange={(e) => setRemotePort(e.target.value)}
              placeholder="Remote port"
              className="w-20 text-xs bg-background border border-border rounded px-2 py-1"
            />
            <button
              onClick={handleAddForward}
              className="text-xs px-2 py-1 bg-primary text-primary-foreground rounded hover:bg-primary/90"
            >
              Add
            </button>
          </div>
          {forwards.length > 0 && (
            <div className="space-y-1">
              {forwards.map((forward) => (
                <div
                  key={forward.id}
                  className="flex items-center justify-between text-xs bg-background rounded px-2 py-1"
                >
                  <span>
                    localhost:{forward.localPort} → {forward.remoteHost}:{forward.remotePort}
                  </span>
                  <button
                    onClick={() => handleRemoveForward(forward.id)}
                    className="text-red-500 hover:text-red-700"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Macro panel */}
      {showMacroPanel && (
        <div className="px-3 py-2 bg-secondary/50 border-b border-border">
          <div className="flex items-center gap-2 mb-2">
            <input
              type="text"
              value={newMacroName}
              onChange={(e) => setNewMacroName(e.target.value)}
              placeholder="Macro name"
              className="flex-1 text-xs bg-background border border-border rounded px-2 py-1"
            />
            <button
              onClick={isRecordingMacro ? stopRecordingMacro : startRecordingMacro}
              className={`text-xs px-2 py-1 rounded ${
                isRecordingMacro ? 'bg-red-500 text-white' : 'bg-primary text-primary-foreground'
              }`}
            >
              {isRecordingMacro ? 'Stop' : 'Record'}
            </button>
          </div>
          {macros.length > 0 && (
            <div className="space-y-1">
              {macros.map((macro) => (
                <div
                  key={macro.id}
                  className="flex items-center justify-between text-xs bg-background rounded px-2 py-1"
                >
                  <span>{macro.name}</span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => runMacro(macro.id)}
                      className="text-xs px-1 py-0.5 bg-primary text-primary-foreground rounded hover:bg-primary/90"
                    >
                      Run
                    </button>
                    <button
                      onClick={() => deleteMacro(macro.id)}
                      className="text-xs px-1 py-0.5 text-red-500 hover:text-red-700"
                    >
                      ✕
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="px-3 py-2 bg-destructive/10 border-b border-destructive/30">
          <div className="flex items-center justify-between">
            <span className="text-xs text-destructive">{error}</span>
            <button
              onClick={() => setError(null)}
              className="text-xs text-destructive hover:text-destructive/80"
            >
              ✕
            </button>
          </div>
        </div>
      )}

      {/* Terminal container */}
      <div ref={containerRef} className="flex-1 min-h-0 p-2" />
    </div>
  )
}
