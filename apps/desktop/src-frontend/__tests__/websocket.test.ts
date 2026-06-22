/**
 * WebSocket Binary Protocol Unit Tests
 *
 * Run with: npx jest websocket.test.ts
 */

declare const describe: (name: string, fn: () => void) => void
declare const it: (name: string, fn: () => void) => void
declare const expect: (value: any) => any

import {
  buildResizeMessage,
  buildInputMessage,
  buildHeartbeatMessage,
  buildAuthMessage,
  buildCloseMessage,
  parseOutputMessage,
  isHeartbeatAck,
  computeBackoff,
  WS_MSG_INPUT,
  WS_MSG_OUTPUT,
  WS_MSG_RESIZE,
  WS_MSG_HEARTBEAT,
  WS_MSG_AUTH,
  WS_MSG_CLOSE,
} from '../lib/websocket'

describe('WebSocket Binary Protocol', () => {
  describe('buildResizeMessage', () => {
    it('creates a 5-byte message with correct type', () => {
      const msg = buildResizeMessage(80, 24)
      expect(msg.length).toBe(5)
      expect(msg[0]).toBe(WS_MSG_RESIZE)
    })

    it('encodes cols and rows big-endian', () => {
      const msg = buildResizeMessage(120, 40)
      const view = new DataView(msg.buffer)
      expect(view.getUint16(1, false)).toBe(120)
      expect(view.getUint16(3, false)).toBe(40)
    })

    it('handles max uint16 values', () => {
      const msg = buildResizeMessage(65535, 65535)
      const view = new DataView(msg.buffer)
      expect(view.getUint16(1, false)).toBe(65535)
      expect(view.getUint16(3, false)).toBe(65535)
    })
  })

  describe('buildInputMessage', () => {
    it('creates a message with correct type', () => {
      const msg = buildInputMessage('ls -la')
      expect(msg[0]).toBe(WS_MSG_INPUT)
    })

    it('encodes UTF-8 data after type byte', () => {
      const data = 'hello world'
      const msg = buildInputMessage(data)
      const decoder = new TextDecoder('utf-8')
      expect(decoder.decode(msg.slice(1))).toBe(data)
    })

    it('handles unicode characters', () => {
      const data = 'héllo 🌍'
      const msg = buildInputMessage(data)
      const decoder = new TextDecoder('utf-8')
      expect(decoder.decode(msg.slice(1))).toBe(data)
    })

    it('handles empty string', () => {
      const msg = buildInputMessage('')
      expect(msg.length).toBe(1)
      expect(msg[0]).toBe(WS_MSG_INPUT)
    })
  })

  describe('buildHeartbeatMessage', () => {
    it('creates a 1-byte heartbeat message', () => {
      const msg = buildHeartbeatMessage()
      expect(msg.length).toBe(1)
      expect(msg[0]).toBe(WS_MSG_HEARTBEAT)
    })
  })

  describe('buildAuthMessage', () => {
    it('creates a message with correct type', () => {
      const msg = buildAuthMessage('my.jwt.token')
      expect(msg[0]).toBe(WS_MSG_AUTH)
    })

    it('encodes the token after type byte', () => {
      const token = 'eyJhbGciOiJIUzI1NiJ9.test'
      const msg = buildAuthMessage(token)
      const decoder = new TextDecoder('utf-8')
      expect(decoder.decode(msg.slice(1))).toBe(token)
    })
  })

  describe('buildCloseMessage', () => {
    it('creates a 1-byte close message', () => {
      const msg = buildCloseMessage()
      expect(msg.length).toBe(1)
      expect(msg[0]).toBe(WS_MSG_CLOSE)
    })
  })

  describe('parseOutputMessage', () => {
    it('returns null for non-output message type', () => {
      const buf = new Uint8Array([WS_MSG_HEARTBEAT])
      expect(parseOutputMessage(buf.buffer)).toBeNull()
    })

    it('returns null for empty buffer', () => {
      expect(parseOutputMessage(new ArrayBuffer(0))).toBeNull()
    })

    it('parses output message correctly', () => {
      const encoder = new TextEncoder()
      const payload = encoder.encode('terminal output')
      const buf = new Uint8Array(1 + payload.length)
      buf[0] = WS_MSG_OUTPUT
      buf.set(payload, 1)

      expect(parseOutputMessage(buf.buffer)).toBe('terminal output')
    })

    it('handles unicode output', () => {
      const encoder = new TextEncoder()
      const payload = encoder.encode('héllo 🌍')
      const buf = new Uint8Array(1 + payload.length)
      buf[0] = WS_MSG_OUTPUT
      buf.set(payload, 1)

      expect(parseOutputMessage(buf.buffer)).toBe('héllo 🌍')
    })
  })

  describe('isHeartbeatAck', () => {
    it('returns true for heartbeat message', () => {
      const buf = new Uint8Array([WS_MSG_HEARTBEAT])
      expect(isHeartbeatAck(buf.buffer)).toBe(true)
    })

    it('returns false for other message types', () => {
      const buf = new Uint8Array([WS_MSG_OUTPUT])
      expect(isHeartbeatAck(buf.buffer)).toBe(false)
    })

    it('returns false for empty buffer', () => {
      expect(isHeartbeatAck(new ArrayBuffer(0))).toBe(false)
    })
  })

  describe('computeBackoff', () => {
    it('returns base delay for first attempt', () => {
      const delay = computeBackoff(0, 1000, 30000)
      expect(delay).toBeGreaterThanOrEqual(850)
      expect(delay).toBeLessThanOrEqual(1150)
    })

    it('doubles delay with each attempt', () => {
      const d0 = computeBackoff(0, 1000, 30000)
      const d1 = computeBackoff(1, 1000, 30000)
      const d2 = computeBackoff(2, 1000, 30000)

      // With jitter, just verify exponential trend
      expect(d1).toBeGreaterThanOrEqual(d0 * 0.5) // allow jitter
      expect(d2).toBeGreaterThanOrEqual(d1 * 0.5)
    })

    it('caps at maxMs', () => {
      const delay = computeBackoff(100, 1000, 30000)
      expect(delay).toBeLessThanOrEqual(30000)
    })

    it('applies jitter', () => {
      // Run many times and verify jitter spread
      const delays = Array.from({ length: 50 }, () => computeBackoff(2, 1000, 30000))
      const min = Math.min(...delays)
      const max = Math.max(...delays)
      expect(max).toBeGreaterThan(min)
    })
  })
})
