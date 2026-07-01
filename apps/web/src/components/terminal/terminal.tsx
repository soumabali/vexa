'use client';

import React, { useRef, useEffect, useCallback, useState } from 'react';
import { Terminal as XTerm } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { TerminalSanitizer } from '@/lib/terminal-sanitize';
import 'xterm/css/xterm.css';

export interface TerminalProps {
  id: string;
  wsUrl: string;
  onData?: (data: string) => void;
  onBinary?: (data: Uint8Array) => void;
  onConnectionChange?: (status: TerminalStatus) => void;
  onTitleChange?: (title: string) => void;
  theme?: TerminalTheme;
  fontSize?: number;
  fontFamily?: string;
  scrollback?: number;
  cursorBlink?: boolean;
  cursorStyle?: 'block' | 'bar' | 'underline';
  className?: string;
}

export type TerminalStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface TerminalTheme {
  background: string;
  foreground: string;
  cursor: string;
  cursorAccent: string;
  selectionBackground: string;
  selectionForeground?: string;
  black: string;
  red: string;
  green: string;
  yellow: string;
  blue: string;
  magenta: string;
  cyan: string;
  white: string;
  brightBlack: string;
  brightRed: string;
  brightGreen: string;
  brightYellow: string;
  brightBlue: string;
  brightMagenta: string;
  brightCyan: string;
  brightWhite: string;
}

export const DEFAULT_DARK_THEME: TerminalTheme = {
  background: '#1e1e1e',
  foreground: '#d4d4d4',
  cursor: '#d4d4d4',
  cursorAccent: '#1e1e1e',
  selectionBackground: '#264f78',
  selectionForeground: '#ffffff',
  black: '#000000',
  red: '#cd3131',
  green: '#0dbc79',
  yellow: '#e5e510',
  blue: '#2472c8',
  magenta: '#bc3fbc',
  cyan: '#11a8cd',
  white: '#e5e5e5',
  brightBlack: '#666666',
  brightRed: '#f14c4c',
  brightGreen: '#23d18b',
  brightYellow: '#f5f543',
  brightBlue: '#3b8eea',
  brightMagenta: '#d670d6',
  brightCyan: '#29b8db',
  brightWhite: '#e5e5e5',
};

export const DEFAULT_LIGHT_THEME: TerminalTheme = {
  background: '#ffffff',
  foreground: '#333333',
  cursor: '#333333',
  cursorAccent: '#ffffff',
  selectionBackground: '#add6ff',
  selectionForeground: '#000000',
  black: '#000000',
  red: '#cd3131',
  green: '#00bc00',
  yellow: '#949800',
  blue: '#0451a5',
  magenta: '#bc05bc',
  cyan: '#0598bc',
  white: '#555555',
  brightBlack: '#666666',
  brightRed: '#cd3131',
  brightGreen: '#14ce14',
  brightYellow: '#b5ba00',
  brightBlue: '#0451a5',
  brightMagenta: '#bc05bc',
  brightCyan: '#0598bc',
  brightWhite: '#a5a5a5',
};

export const THEMES: Record<string, TerminalTheme> = {
  dark: DEFAULT_DARK_THEME,
  light: DEFAULT_LIGHT_THEME,
};

const MAX_RECONNECT_ATTEMPTS = 3;
const RECONNECT_BASE_DELAY = 1000;

export default function Terminal({
  id,
  wsUrl,
  onData,
  onBinary,
  onConnectionChange,
  onTitleChange,
  theme = DEFAULT_DARK_THEME,
  fontSize = 14,
  fontFamily = 'JetBrains Mono, Fira Code, monospace',
  scrollback = 10000,
  cursorBlink = true,
  cursorStyle = 'block',
  className,
}: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<XTerm | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const sanitizerRef = useRef<TerminalSanitizer>(new TerminalSanitizer());
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const connectRef = useRef<() => void>(() => {});
  const [status, setStatus] = useState<TerminalStatus>('connecting');

  const updateStatus = useCallback(
    (newStatus: TerminalStatus) => {
      setStatus(newStatus);
      onConnectionChange?.(newStatus);
    },
    [onConnectionChange]
  );

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    updateStatus('connecting');

    try {
      const ws = new WebSocket(wsUrl);
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        reconnectAttemptsRef.current = 0;
        sanitizerRef.current.reset();
        updateStatus('connected');
      };

      ws.onmessage = (event) => {
        if (!termRef.current) return;

        if (event.data instanceof ArrayBuffer) {
          const data = new Uint8Array(event.data);
          termRef.current.write(data);
          onBinary?.(data);
        } else if (typeof event.data === 'string') {
          const sanitized = sanitizerRef.current.sanitize(event.data);
          if (!sanitized) return;

          const titleMatch = sanitized.match(/\x1b\]0;(.+?)\x07/);
          if (titleMatch && onTitleChange) {
            onTitleChange(sanitizerRef.current.sanitizeTitle(titleMatch[1]));
          }

          termRef.current.write(sanitized);
          onData?.(sanitized);
        }
      };

      ws.onclose = () => {
        updateStatus('disconnected');
        termRef.current?.writeln('\r\n\x1b[33m[Connection closed]\x1b[0m');

        // Auto-reconnect
        if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          const delay = RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttemptsRef.current);
          reconnectAttemptsRef.current += 1;
          reconnectTimeoutRef.current = setTimeout(() => {
            termRef.current?.writeln(`\r\n\x1b[33m[Reconnecting (${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})...]\x1b[0m`);
            connectRef.current();
          }, delay);
        } else {
          termRef.current?.writeln('\r\n\x1b[31m[Max reconnect attempts reached. Please refresh to reconnect.]\x1b[0m');
          updateStatus('error');
        }
      };

      ws.onerror = () => {
        updateStatus('error');
        termRef.current?.writeln('\r\n\x1b[31m[Connection error]\x1b[0m');
      };

      wsRef.current = ws;
    } catch (err) {
      updateStatus('error');
      termRef.current?.writeln(`\r\n\x1b[31m[Connection failed: ${err}]\x1b[0m`);
    }
  }, [wsUrl, updateStatus, onData, onBinary, onTitleChange]);

  useEffect(() => {
    connectRef.current = connect;
  });

  useEffect(() => {
    if (!containerRef.current) return;

    const term = new XTerm({
      theme,
      fontSize,
      fontFamily,
      scrollback,
      cursorBlink,
      cursorStyle,
      allowProposedApi: true,
      convertEol: true,
      rightClickSelectsWord: false,
    });

    const fitAddon = new FitAddon();
    const webLinksAddon = new WebLinksAddon();

    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);

    term.open(containerRef.current);
    fitAddon.fit();

    // Handle user input
    const disposable = term.onData((data) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(data);
      }
    });

    termRef.current = term;

    // Connect WebSocket
    connect();

    // Handle resize
    const handleResize = () => {
      fitAddon.fit();
      const { cols, rows } = term;
      const resizeMsg = JSON.stringify({ type: 'resize', cols, rows });
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(resizeMsg);
      }
    };

    window.addEventListener('resize', handleResize);

    return () => {
      disposable.dispose();
      webLinksAddon.dispose();
      fitAddon.dispose();
      term.dispose();
      window.removeEventListener('resize', handleResize);

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      wsRef.current?.close();
    };
  }, [id, theme, fontSize, fontFamily, scrollback, cursorBlink, cursorStyle, connect]);

  // Expose methods via ref if needed
  const write = useCallback((data: string | Uint8Array) => {
    termRef.current?.write(data);
  }, []);

  const clear = useCallback(() => {
    termRef.current?.clear();
  }, []);

  const focus = useCallback(() => {
    termRef.current?.focus();
  }, []);

  const getSelection = useCallback(() => {
    return termRef.current?.getSelection() || '';
  }, []);

  return (
    <div className={`terminal-container flex flex-col h-full bg-black p-4 font-mono-code ${className || ''}`}>
      <div
        ref={containerRef}
        className="terminal-viewport flex-1 overflow-hidden text-mono-code"
        style={{ backgroundColor: theme.background }}
      />
    </div>
  );
}

export { Terminal, XTerm, FitAddon, WebLinksAddon };
