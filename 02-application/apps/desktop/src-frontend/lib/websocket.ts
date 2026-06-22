/**
 * WebSocket Binary Protocol Constants
 *
 * Shared between frontend Terminal and Go backend.
 * Keep in sync with: apps/api/internal/terminal/session.go
 */

export const WS_MSG_INPUT = 0x01
export const WS_MSG_OUTPUT = 0x02
export const WS_MSG_RESIZE = 0x03
export const WS_MSG_HEARTBEAT = 0x04
export const WS_MSG_AUTH = 0x05
export const WS_MSG_CLOSE = 0x06

/**
 * Build a binary resize message: [type:u8, cols:u16be, rows:u16be]
 */
export function buildResizeMessage(cols: number, rows: number): Uint8Array {
  const buf = new ArrayBuffer(5)
  const view = new DataView(buf)
  view.setUint8(0, WS_MSG_RESIZE)
  view.setUint16(1, cols, false) // big-endian
  view.setUint16(3, rows, false)
  return new Uint8Array(buf)
}

/**
 * Build a binary input message: [type:u8, ...utf8-data]
 */
export function buildInputMessage(data: string): Uint8Array {
  const encoder = new TextEncoder()
  const payload = encoder.encode(data)
  const buf = new Uint8Array(1 + payload.length)
  buf[0] = WS_MSG_INPUT
  buf.set(payload, 1)
  return buf
}

/**
 * Build a binary heartbeat message.
 */
export function buildHeartbeatMessage(): Uint8Array {
  return new Uint8Array([WS_MSG_HEARTBEAT])
}

/**
 * Build a binary auth message: [type:u8, ...jwt-token-utf8]
 */
export function buildAuthMessage(token: string): Uint8Array {
  const encoder = new TextEncoder()
  const payload = encoder.encode(token)
  const buf = new Uint8Array(1 + payload.length)
  buf[0] = WS_MSG_AUTH
  buf.set(payload, 1)
  return buf
}

/**
 * Build a binary close message.
 */
export function buildCloseMessage(): Uint8Array {
  return new Uint8Array([WS_MSG_CLOSE])
}

/**
 * Parse an incoming binary output message.
 * Returns the terminal string, or null if not an output message.
 */
export function parseOutputMessage(data: ArrayBuffer): string | null {
  const view = new DataView(data)
  if (view.byteLength < 1) return null
  const msgType = view.getUint8(0)
  if (msgType !== WS_MSG_OUTPUT) return null
  const decoder = new TextDecoder('utf-8')
  return decoder.decode(data.slice(1))
}

/**
 * Parse an incoming heartbeat ack.
 */
export function isHeartbeatAck(data: ArrayBuffer): boolean {
  const view = new DataView(data)
  return view.byteLength >= 1 && view.getUint8(0) === WS_MSG_HEARTBEAT
}

/**
 * Compute exponential backoff with jitter.
 */
export function computeBackoff(
  attempt: number,
  baseMs: number = 1000,
  maxMs: number = 30000
): number {
  const exp = Math.min(attempt, 6)
  const jitter = Math.random() * 0.3 + 0.85 // 0.85–1.15 jitter
  return Math.min(baseMs * Math.pow(2, exp) * jitter, maxMs)
}

/**
 * Connection state type.
 */
export type ConnectionState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected' | 'error'
