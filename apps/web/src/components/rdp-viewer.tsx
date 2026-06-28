'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface RDPViewerProps {
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
}

export function RDPViewer({ sessionId, width = 1920, height = 1080, onDisconnect }: RDPViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
  const [quality, setQuality] = useState(80);
  const [fps, setFps] = useState(0);
  const fpsRef = useRef(0);
  const lastTimeRef = useRef<number>(0);
  useEffect(() => {
    if (lastTimeRef.current === 0) {
      lastTimeRef.current = Date.now();
    }
  }, []);

  const handleBinaryMessageRef = useRef<(data: ArrayBuffer) => void>(() => {});
  const renderScreenRef = useRef<(data: Uint8Array) => void>(() => {});
  const handleClipboardRef = useRef<(data: Uint8Array) => void>(() => {});

  const connect = useCallback(async (params: {
    hostname: string;
    port: number;
    username: string;
    password: string;
    domain?: string;
  }) => {
    setConnecting(true);
    setError(null);

    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WS_URL}/api/rdp/connect`);
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.onopen = () => {
      ws.send(JSON.stringify({
        session_id: sessionId,
        ...params,
        width,
        height,
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
        // Binary data - screen update
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
  }, [sessionId, width, height, onDisconnect]);

  const handleBinaryMessage = useCallback((data: ArrayBuffer) => {
    const bytes = new Uint8Array(data);
    if (bytes.length === 0) return;

    const frameType = bytes[0];
    const payload = bytes.slice(1);

    switch (frameType) {
      case 0x01: // Screen update
        renderScreenRef.current(payload);
        break;
      case 0x05: // Clipboard
        handleClipboardRef.current(payload);
        break;
      case 0xFF: // Error
        setError(new TextDecoder().decode(payload));
        break;
    }

    // Update FPS
    fpsRef.current++;
    const now = Date.now();
    if (lastTimeRef.current !== 0 && now - lastTimeRef.current >= 1000) {
      setFps(fpsRef.current);
      fpsRef.current = 0;
      lastTimeRef.current = now;
    }
  }, []);

  const renderScreen = useCallback((data: Uint8Array) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // For FreeRDP output, we'd decode the actual frame data
    // This is a simplified placeholder that draws the raw data
    // In production, you'd use a proper RDP decoder or canvas-based rendering

    // Create image from data
    const imageData = new ImageData(
      new Uint8ClampedArray(data),
      canvas.width,
      canvas.height
    );
    ctx.putImageData(imageData, 0, 0);
  }, []);

  const handleClipboard = useCallback((data: Uint8Array) => {
    const text = new TextDecoder().decode(data);
    navigator.clipboard.writeText(text).catch(console.error);
  }, []);

  useEffect(() => {
    handleBinaryMessageRef.current = handleBinaryMessage;
  });
  useEffect(() => {
    renderScreenRef.current = renderScreen;
    handleClipboardRef.current = handleClipboard;
  });

  const sendInput = useCallback((inputData: Uint8Array) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;

    const frame = new Uint8Array(1 + inputData.length);
    frame[0] = 0x02; // Input frame type
    frame.set(inputData, 1);
    wsRef.current.send(frame);
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent<Element>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = Math.floor((e.clientX - rect.left) * (width / rect.width));
    const y = Math.floor((e.clientY - rect.top) * (height / rect.height));

    // Send mouse move: [button_mask, x_high, x_low, y_high, y_low]
    const data = new Uint8Array(5);
    data[0] = e.buttons; // button mask
    data[1] = (x >> 8) & 0xFF;
    data[2] = x & 0xFF;
    data[3] = (y >> 8) & 0xFF;
    data[4] = y & 0xFF;
    sendInput(data);
  }, [width, height, sendInput]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    handleMouseMove(e);
  }, [handleMouseMove]);

  const handleMouseUp = useCallback((e: React.MouseEvent) => {
    handleMouseMove(e);
  }, [handleMouseMove]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    e.preventDefault();

    // Send key event: [down_flag, key_high, key_low]
    const data = new Uint8Array(3);
    data[0] = 1; // key down
    data[1] = (e.keyCode >> 8) & 0xFF;
    data[2] = e.keyCode & 0xFF;
    sendInput(data);
  }, [sendInput]);

  const handleKeyUp = useCallback((e: React.KeyboardEvent) => {
    const data = new Uint8Array(3);
    data[0] = 0; // key up
    data[1] = (e.keyCode >> 8) & 0xFF;
    data[2] = e.keyCode & 0xFF;
    sendInput(data);
  }, [sendInput]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setConnected(false);
  }, []);

  const handleQualityChange = useCallback((value: number) => {
    setQuality(value);
    // Send quality update
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const data = new Uint8Array(2);
      data[0] = 0x07; // Quality frame type
      data[1] = value;
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
          <h3 className="text-lg font-semibold">RDP Connection</h3>
          <RDPConnectForm onConnect={connect} connecting={connecting} />
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Badge variant={connected ? "default" : "secondary"}>
            {connected ? "Connected" : "Disconnected"}
          </Badge>
          <span className="text-sm text-muted-foreground">{fps} FPS</span>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-sm">Quality:</label>
          <Input
            type="range"
            min="30"
            max="100"
            value={quality}
            onChange={(e) => handleQualityChange(Number(e.target.value))}
            className="w-24"
          />
          <span className="text-sm w-8">{quality}%</span>
          <Button variant="destructive" size="sm" onClick={disconnect}>
            Disconnect
          </Button>
        </div>
      </div>

      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        className="border rounded-lg cursor-crosshair max-w-full"
        style={{ aspectRatio: `${width}/${height}` }}
        onMouseMove={handleMouseMove}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
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

// RDP Connect Form Component
function RDPConnectForm({
  onConnect,
  connecting
}: {
  onConnect: (params: { hostname: string; port: number; username: string; password: string; domain?: string }) => void;
  connecting: boolean;
}) {
  const [hostname, setHostname] = useState('');
  const [port, setPort] = useState(3389);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [domain, setDomain] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConnect({ hostname, port, username, password, domain });
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
            placeholder="3389"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Username</label>
          <Input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Administrator"
            required
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Domain</label>
          <Input
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="WORKGROUP"
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
          required
        />
      </div>

      <Button type="submit" disabled={connecting} className="w-full">
        {connecting ? 'Connecting...' : 'Connect'}
      </Button>
    </form>
  );
}
