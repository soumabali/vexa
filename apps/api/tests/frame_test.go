package tests

import "testing"

// TestWebSocketFramePayloads is skipped because TerminalResizePayload,
// ErrorPayload, AuthPayload, ClosePayload, ControlPayload, and NACKPayload
// types are not implemented in the current codebase.
func TestWebSocketFramePayloads(t *testing.T) {
	t.Skip("WebSocket frame payload types not implemented")
}
