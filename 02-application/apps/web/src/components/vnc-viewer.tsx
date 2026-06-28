'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface VNCViewerProps {
  sessionId: string;
  width?: number;
  height?: number;
  onDisconnect?: () => void;
}

interface SessionInfo {
  session_id: string;
  status: string;
  framebuffer_w?: number;
  framebuffer_h?: number;
  desktop_name?: string;
}

// VNC key mapping
const VNC_KEYS: Record<string, number> = {
  Backspace: 0xff08,
  Tab: 0xff09,
  Enter: 0xff0d,
  Escape: 0xff1b,
  Insert: 0xff63,
  Delete: 0xffff,
  Home: 0xff50,
  End: 0xff57,
  PageUp: 0xff55,
  PageDown: 0xff56,
  ArrowLeft: 0xff51,
  ArrowUp: 0xff52,
  ArrowRight: 0xff53,
  ArrowDown: 0xff54,
  F1: 0xffbe,
  F2: 0xffbf,
  F3: 0xffc0,
  F4: 0xffc1,
  F5: 0xffc2,
  F6: 0xffc3,
  F7: 0xffc4,
  F8: 0xffc5,
  F9: 0xffc6,
  F10: 0xffc7,
  F11: 0xffc8,
  F12: 0xffc9,
  Shift: 0xffe1,
  Control: 0xffe3,
  Alt: 0xffe9,
  Meta: 0xffeb,
  CapsLock: 0xffe5,
};

export function VNCViewer({ sessionId, width = 1024, height = 768, onDisconnect }: VNCViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
  const [quality, setQuality] = useState('high');
  const [viewOnly, setViewOnly] = useState(false);
  const handleBinaryMessageRef = useRef<(data: ArrayBuffer) => void>(() => {});
  const renderScreenUpdateRef = useRef<(data: Uint8Array) => void>(() => {});
  const handleClipboardRef = useRef<(data: Uint8Array) => void>(() => {});
  const [scaleMode, setScaleMode] = useState<'fit' | 'stretch' | 'original'>('fit');
  
  // Framebuffer state
  const fbRef = useRef({
    width: 0,
    height: 0,
    imageData: null as ImageData | null,
  });

  const connect = useCallback(async (params: {
    hostname: string;
    port: number;
    password: string;
  }) => {
    setConnecting(true);
    setError(null);

    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WS_URL}/api/vnc/connect`);
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.onopen = () => {
      ws.send(JSON.stringify({
        session_id: sessionId,
        ...params,
      }));
    };

    ws.onmessage = (event) => {
      if (typeof event.data === 'string') {
        const msg = JSON.parse(event.data);
        if (msg.error) {
          setError(msg.error);
          setConnecting(false);
          ws.close();
          return;
        }
        if (msg.type === 'session_info') {
          setSessionInfo(msg);
          setConnected(true);
          setConnecting(false);
        }
      } else {
        handleBinaryMessageRef.current(event.data as ArrayBuffer);
      }
    };

    ws.onerror = () => {
      setError('WebSocket error');
      setConnecting(false);
      setConnected(false);
    };

    ws.onclose = () => {
      setConnected(false);
      setConnecting(false);
      if (onDisconnect) onDisconnect();
    };
  }, [sessionId, onDisconnect]);

  const handleBinaryMessage = useCallback((data: ArrayBuffer) => {
    const bytes = new Uint8Array(data);
    if (bytes.length === 0) return;

    const frameType = bytes[0];
    const payload = bytes.slice(1);

    switch (frameType) {
      case 0x01: // Screen update
        renderScreenUpdateRef.current(payload);
        break;
      case 0x05: // Clipboard
        handleClipboardRef.current(payload);
        break;
      case 0xFF: // Error
        setError(new TextDecoder().decode(payload));
        break;
    }
  }, []);

  const renderScreenUpdate = useCallback((data: Uint8Array) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Parse VNC framebuffer update
    // [x:2][y:2][w:2][h:2][encoding:4][data...]
    if (data.length < 12) return;

    const x = (data[0] << 8) | data[1];
    const y = (data[2] << 8) | data[3];
    const w = (data[4] << 8) | data[5];
    const h = (data[6] << 8) | data[7];
    // const encoding = (data[8] << 24) | (data[9] << 16) | (data[10] << 8) | data[11];
    const pixelData = data.slice(12);

    // Create image data for the update region
    const imageData = ctx.createImageData(w, h);
    
    // Convert pixel data (32-bit RGB) to ImageData (RGBA)
    for (let i = 0; i < pixelData.length / 4; i++) {
      const offset = i * 4;
      imageData.data[offset] = pixelData[offset + 2];     // R
      imageData.data[offset + 1] = pixelData[offset + 1]; // G
      imageData.data[offset + 2] = pixelData[offset];     // B
      imageData.data[offset + 3] = 255;                   // A
    }

    ctx.putImageData(imageData, x, y);
  }, []);

  const handleClipboard = useCallback((data: Uint8Array) => {
    const text = new TextDecoder().decode(data);
    navigator.clipboard.writeText(text).catch(console.error);
  }, []);

  useEffect(() => {
    handleBinaryMessageRef.current = handleBinaryMessage;
  });
  useEffect(() => {
    renderScreenUpdateRef.current = renderScreenUpdate;
    handleClipboardRef.current = handleClipboard;
  });

  const sendInput = useCallback((inputData: Uint8Array) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    if (viewOnly) return;

    const frame = new Uint8Array(1 + inputData.length);
    frame[0] = 0x02; // Input frame type
    frame.set(inputData, 1);
    wsRef.current.send(frame);
  }, [viewOnly]);

  const sendPointerEvent = useCallback((x: number, y: number, buttonMask: number) => {
    // VNC pointer event: [button_mask:1][x:2][y:2]
    const data = new Uint8Array(5);
    data[0] = buttonMask;
    data[1] = (x >> 8) & 0xFF;
    data[2] = x & 0xFF;
    data[3] = (y >> 8) & 0xFF;
    data[4] = y & 0xFF;
    sendInput(data);
  }, [sendInput]);

  const sendKeyEvent = useCallback((key: string, down: boolean) => {
    const keyCode = VNC_KEYS[key] || key.charCodeAt(0);
    
    // VNC key event: [down_flag:1][key:3]
    const data = new Uint8Array(4);
    data[0] = down ? 1 : 0;
    data[1] = (keyCode >> 16) & 0xFF;
    data[2] = (keyCode >> 8) & 0xFF;
    data[3] = keyCode & 0xFF;
    sendInput(data);
  }, [sendInput]);

  const handleMouseMove = useCallback((e: React.MouseEvent<Element>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const scaleX = fbRef.current.width / rect.width;
    const scaleY = fbRef.current.height / rect.height;

    const x = Math.floor((e.clientX - rect.left) * scaleX);
    const y = Math.floor((e.clientY - rect.top) * scaleY);

    sendPointerEvent(x, y, e.buttons);
  }, [sendPointerEvent]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    const buttonMap: Record<number, number> = { 0: 1, 1: 4, 2: 2 };
    const buttonMask = buttonMap[e.button] || 0;
    handleMouseMove(e);

    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const scaleX = fbRef.current.width / rect.width;
    const scaleY = fbRef.current.height / rect.height;
    const x = Math.floor((e.clientX - rect.left) * scaleX);
    const y = Math.floor((e.clientY - rect.top) * scaleY);
    sendPointerEvent(x, y, buttonMask);
  }, [handleMouseMove, sendPointerEvent]);

  const handleMouseUp = useCallback((e: React.MouseEvent) => {
    handleMouseMove(e);
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const scaleX = fbRef.current.width / rect.width;
    const scaleY = fbRef.current.height / rect.height;
    const x = Math.floor((e.clientX - rect.left) * scaleX);
    const y = Math.floor((e.clientY - rect.top) * scaleY);
    sendPointerEvent(x, y, 0);
  }, [handleMouseMove, sendPointerEvent]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    e.preventDefault();
    sendKeyEvent(e.key, true);
  }, [sendKeyEvent]);

  const handleKeyUp = useCallback((e: React.KeyboardEvent) => {
    sendKeyEvent(e.key, false);
  }, [sendKeyEvent]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setConnected(false);
  }, []);

  const handleQualityChange = useCallback((value: string) => {
    setQuality(value);
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const data = new Uint8Array(2);
      data[0] = 0x07; // Quality frame type
      data[1] = value === 'low' ? 0 : value === 'medium' ? 1 : 2;
      wsRef.current.send(data);
    }
  }, []);

  const handleResize = useCallback((width: number, height: number) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    canvas.width = width;
    canvas.height = height;
    fbRef.current.width = width;
    fbRef.current.height = height;
    
    // Send resize to server
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const data = new Uint8Array(5);
      data[0] = 0x04; // Resize frame type
      data[1] = (width >> 8) & 0xFF;
      data[2] = width & 0xFF;
      data[3] = (height >> 8) & 0xFF;
      data[4] = height & 0xFF;
      wsRef.current.send(data);
    }
  }, []);

  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const getCanvasStyle = () => {
    switch (scaleMode) {
      case 'fit':
        return { maxWidth: '100%', maxHeight: '80vh' };
      case 'stretch':
        return { width: '100%', height: '80vh' };
      case 'original':
        return {};
    }
  };

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <Badge variant="destructive" className="mb-4">Error</Badge>
          <p className="text-destructive mb-4">{error}</p>
          <Button onClick={() => setError(null)}>Retry</Button>
        </div>
      </Card>
    );
  }

  if (!connected) {
    return (
      <Card className="p-6">
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">VNC Connection</h3>
          <VNCConnectForm onConnect={connect} connecting={connecting} />
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          <Badge variant={connected ? "default" : "secondary"}>
            {connected ? "Connected" : "Disconnected"}
          </Badge>
          {sessionInfo?.desktop_name && (
            <span className="text-sm text-muted-foreground">
              {sessionInfo.desktop_name}
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-2 flex-wrap">
          <label className="text-sm">View Only</label>
          <Input
            type="checkbox"
            checked={viewOnly}
            onChange={(e) => setViewOnly(e.target.checked)}
            className="w-4 h-4"
          />
          
          <select
            value={quality}
            onChange={(e) => handleQualityChange(e.target.value)}
            className="text-sm border rounded p-1"
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
          
          <select
            value={scaleMode}
            onChange={(e) => setScaleMode(e.target.value as 'fit' | 'stretch' | 'original')}
            className="text-sm border rounded p-1"
          >
            <option value="fit">Fit</option>
            <option value="stretch">Stretch</option>
            <option value="original">Original</option>
          </select>
          
          <Button variant="destructive" size="sm" onClick={disconnect}>
            Disconnect
          </Button>
        </div>
      </div>

      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="border rounded-lg cursor-crosshair"
        style={getCanvasStyle()}
        onMouseMove={handleMouseMove}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
        onContextMenu={(e) => e.preventDefault()}
        onKeyDown={handleKeyDown}
        onKeyUp={handleKeyUp}
        tabIndex={0}
      />

      {sessionInfo && (
        <div className="text-sm text-muted-foreground">
          Session: {sessionInfo.session_id} | 
          Resolution: {sessionInfo.framebuffer_w}x{sessionInfo.framebuffer_h}
        </div>
      )}
    </div>
  );
}

// VNC Connect Form Component
function VNCConnectForm({
  onConnect,
  connecting
}: {
  onConnect: (params: { hostname: string; port: number; password: string }) => void;
  connecting: boolean;
}) {
  const [hostname, setHostname] = useState('');
  const [port, setPort] = useState(5900);
  const [password, setPassword] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConnect({ hostname, port, password });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Hostname</label>
          <Input
            value={hostname}
            onChange={(e) => setHostname(e.target.value)}
            placeholder="192.168.1.100"
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Port</label>
          <Input
            type="number"
            value={port}
            onChange={(e) => setPort(Number(e.target.value))}
            placeholder="5900"
          />
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Password</label>
        <Input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="••••••••"
        />
      </div>

      <Button type="submit" disabled={connecting} className="w-full">
        {connecting ? 'Connecting...' : 'Connect'}
      </Button>
    </form>
  );
}
