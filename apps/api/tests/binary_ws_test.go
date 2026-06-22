package tests

import "testing"

// TestBinaryFrameProtocol is skipped because the binary WebSocket frame protocol
// (FrameHeader, FrameTypeSSH, HeaderSize, DecodeHeader, MaxFrameSize) is not yet implemented.
func TestBinaryFrameProtocol(t *testing.T) {
	t.Skip("binary WebSocket frame protocol not implemented")
}
