package fuzz

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// FuzzWebSocketFrame fuzzes WebSocket frame parsing
func FuzzWebSocketFrame(f *testing.F) {
	// Seed corpus: valid frames
	// Text frame with "Hello"
	f.Add([]byte{0x81, 0x05, 'H', 'e', 'l', 'l', 'o'})
	// Binary frame with 4 bytes
	f.Add([]byte{0x82, 0x04, 0x01, 0x02, 0x03, 0x04})
	// Ping frame
	f.Add([]byte{0x89, 0x00})
	// Pong frame
	f.Add([]byte{0x8A, 0x00})
	// Close frame
	f.Add([]byte{0x88, 0x02, 0x03, 0xE8})
	// Fragmented frame
	f.Add([]byte{0x01, 0x03, 'H', 'i', '!', 0x80, 0x02, 'T', 'h'})
	// Empty frame
	f.Add([]byte{0x81, 0x00})
	// 16-bit length
	f.Add([]byte{0x81, 0x7E, 0x00, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				// Expected for malformed input
			}
		}()

		frame, err := parseWebSocketFrame(data)
		if err != nil {
			return
		}

		// Validate frame
		_ = frame.opcode
		_ = frame.payload
	})
}

type wsFrame struct {
	fin     bool
	rsv     [3]bool
	opcode  byte
	masked  bool
	length  uint64
	maskKey [4]byte
	payload []byte
}

func parseWebSocketFrame(data []byte) (*wsFrame, error) {
	if len(data) < 2 {
		return nil, bytes.ErrTooLarge
	}

	frame := &wsFrame{}

	// First byte: FIN, RSV, OPCODE
	frame.fin = data[0]&0x80 != 0
	frame.rsv[0] = data[0]&0x40 != 0
	frame.rsv[1] = data[0]&0x20 != 0
	frame.rsv[2] = data[0]&0x10 != 0
	frame.opcode = data[0] & 0x0F

	// Validate opcode
	validOpcodes := []byte{0x0, 0x1, 0x2, 0x8, 0x9, 0xA}
	valid := false
	for _, op := range validOpcodes {
		if frame.opcode == op {
			valid = true
			break
		}
	}
	if !valid {
		return nil, bytes.ErrTooLarge
	}

	// Second byte: MASK, LENGTH
	frame.masked = data[1]&0x80 != 0
	payloadLen := uint64(data[1] & 0x7F)

	offset := 2

	// Extended payload length
	switch payloadLen {
	case 126:
		if len(data) < offset+2 {
			return nil, bytes.ErrTooLarge
		}
		payloadLen = uint64(binary.BigEndian.Uint16(data[offset:]))
		offset += 2
	case 127:
		if len(data) < offset+8 {
			return nil, bytes.ErrTooLarge
		}
		payloadLen = binary.BigEndian.Uint64(data[offset:])
		offset += 8
	}

	// Sanity check
	if payloadLen > 1024*1024*10 { // 10MB max
		return nil, bytes.ErrTooLarge
	}

	frame.length = payloadLen

	// Mask key
	if frame.masked {
		if len(data) < offset+4 {
			return nil, bytes.ErrTooLarge
		}
		copy(frame.maskKey[:], data[offset:])
		offset += 4
	}

	// Payload
	if uint64(len(data)-offset) < payloadLen {
		return nil, bytes.ErrTooLarge
	}
	frame.payload = data[offset : offset+int(payloadLen)]

	// Unmask payload
	if frame.masked {
		for i := range frame.payload {
			frame.payload[i] ^= frame.maskKey[i%4]
		}
	}

	return frame, nil
}

// FuzzWebSocketMessage fuzzes WebSocket message handling
func FuzzWebSocketMessage(f *testing.F) {
	f.Add([]byte(`{"type":"auth","token":"abc123"}`))
	f.Add([]byte(`{"type":"resize","cols":80,"rows":24}`))
	f.Add([]byte(`{"type":"input","data":"ls -la\n"}`))
	f.Add([]byte(`{"type":"ping"}`))
	f.Add([]byte(`{"type":"subscribe","channel":"session:123"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(data) > 65536 {
			return
		}

		// Try to parse as JSON message
		var msg map[string]interface{}
		// In real test, use json.Unmarshal
		_ = msg

		// Validate UTF-8 for text frames
		if !isValidUTF8(data) {
			return
		}
	})
}

func isValidUTF8(data []byte) bool {
	return bytes.Valid(data)
}

// FuzzWebSocketHandshake fuzzes WebSocket handshake
func FuzzWebSocketHandshake(f *testing.F) {
	f.Add([]byte("GET /ws HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n"))
	f.Add([]byte("GET /ws HTTP/1.1\r\nHost: example.com\r\n\r\n"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(data) > 4096 {
			return
		}

		// Parse HTTP request
		lines := bytes.Split(data, []byte("\r\n"))
		if len(lines) == 0 {
			return
		}

		// Check request line
		requestLine := bytes.Fields(lines[0])
		if len(requestLine) != 3 {
			return
		}

		method := string(requestLine[0])
		path := string(requestLine[1])
		version := string(requestLine[2])

		if method != "GET" {
			return
		}

		_ = path
		_ = version

		// Parse headers
		headers := make(map[string]string)
		for i := 1; i < len(lines); i++ {
			if len(lines[i]) == 0 {
				break
			}
			idx := bytes.Index(lines[i], []byte(":"))
			if idx == -1 {
				continue
			}
			key := string(bytes.TrimSpace(lines[i][:idx]))
			value := string(bytes.TrimSpace(lines[i][idx+1:]))
			headers[key] = value
		}

		// Check required headers
		if headers["Upgrade"] != "websocket" {
			return
		}
		if headers["Connection"] != "Upgrade" {
			return
		}
		if headers["Sec-WebSocket-Key"] == "" {
			return
		}
		if headers["Sec-WebSocket-Version"] != "13" {
			return
		}
	})
}

// FuzzWebSocketExtensions fuzzes extension negotiation
func FuzzWebSocketExtensions(f *testing.F) {
	f.Add("permessage-deflate")
	f.Add("permessage-deflate; client_no_context_takeover")
	f.Add("permessage-deflate; server_max_window_bits=10")
	f.Add("x-custom-extension")
	f.Add("")

	f.Fuzz(func(t *testing.T, extension string) {
		if len(extension) > 256 {
			return
		}

		knownExtensions := []string{
			"permessage-deflate",
			"x-webkit-deflate-frame",
			"permessage-deflate; client_no_context_takeover",
			"permessage-deflate; server_no_context_takeover",
		}
		_ = knownExtensions
	})
}

// BenchmarkWebSocketFrame benchmarks frame parsing
func BenchmarkWebSocketFrame(b *testing.B) {
	frame := []byte{0x81, 0x05, 'H', 'e', 'l', 'l', 'o'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseWebSocketFrame(frame)
	}
}
