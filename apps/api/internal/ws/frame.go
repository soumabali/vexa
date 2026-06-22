package ws

// Frame types and structures for binary WebSocket protocol.

// TerminalResizePayload represents a terminal resize event.
type TerminalResizePayload struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// ErrorPayload represents an error frame payload.
type ErrorPayload struct {
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}

// Error codes.
const (
	ErrCodeInvalidFrame     uint16 = 1001
	ErrCodeDecryptionFailed uint16 = 1002
	ErrCodeDecompressionFailed uint16 = 1003
	ErrCodeAuthFailed       uint16 = 1004
	ErrCodeSessionExpired   uint16 = 1005
	ErrCodeRateLimited      uint16 = 1006
	ErrCodeInvalidSequence  uint16 = 1007
	ErrCodePayloadTooLarge  uint16 = 1008
	ErrCodeProtocolError    uint16 = 1009
	ErrCodeInternalError    uint16 = 1010
)

// AuthPayload represents authentication frame payload.
type AuthPayload struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id,omitempty"`
}

// ClosePayload represents a close frame payload.
type ClosePayload struct {
	Code   uint16 `json:"code"`
	Reason string `json:"reason"`
}

// ControlPayload represents a generic control frame payload.
type ControlPayload struct {
	Command string            `json:"cmd"`
	Args    map[string]string `json:"args,omitempty"`
}

// NACKPayload represents a negative acknowledgment for retransmission.
type NACKPayload struct {
	MissingSequences []uint32 `json:"missing"`
}

// RetransmitPayload wraps a retransmitted frame.
type RetransmitPayload struct {
	OriginalSequence uint32 `json:"orig_seq"`
	FrameData        []byte `json:"frame_data"`
}

// HeartbeatStats tracks connection health.
type HeartbeatStats struct {
	SentAt     int64 `json:"sent_at"`
	ReceivedAt int64 `json:"received_at,omitempty"`
	LatencyMs  int64 `json:"latency_ms,omitempty"`
}
