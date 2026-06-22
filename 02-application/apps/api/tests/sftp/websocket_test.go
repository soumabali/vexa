package sftp_test

import (
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/sftp"
)

func TestFileTransferMessageStructure(t *testing.T) {
	msg := sftp.FileTransferMessage{
		Type:       "upload",
		SessionID:  "test-session",
		Path:       "/home/test/file.txt",
		Size:       1024,
		Offset:     0,
		ChunkSize:  64 * 1024,
		TransferID: "test-transfer",
		Error:      "",
	}

	assert.Equal(t, "upload", msg.Type)
	assert.Equal(t, "test-session", msg.SessionID)
	assert.Equal(t, int64(1024), msg.Size)
}

func TestFileTransferMessageJSON(t *testing.T) {
	msg := sftp.FileTransferMessage{
		Type:      "progress",
		SessionID: "test-session",
		TransferID: "transfer-1",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"progress"`)
	assert.Contains(t, string(data), `"session_id":"test-session"`)

	var decoded sftp.FileTransferMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, msg.Type, decoded.Type)
	assert.Equal(t, msg.SessionID, decoded.SessionID)
}

func TestBinaryFrameFormat(t *testing.T) {
	// Test binary frame construction
	offset := int64(1024)
	data := []byte("Hello, World!")
	frame := make([]byte, 12+len(data))
	frame[0] = 0x00 // Marker byte
	binary.BigEndian.PutUint64(frame[0:8], uint64(offset))
	binary.BigEndian.PutUint32(frame[8:12], uint32(len(data)))
	copy(frame[12:], data)

	// Verify frame structure
	assert.Equal(t, uint8(0x00), frame[0])
	assert.Equal(t, offset, int64(binary.BigEndian.Uint64(frame[0:8])))
	assert.Equal(t, uint32(len(data)), binary.BigEndian.Uint32(frame[8:12]))
	assert.Equal(t, data, frame[12:])
}

func TestBinaryFrameLargeOffset(t *testing.T) {
	offset := int64(1 << 40) // 1TB offset
	data := []byte{0x01, 0x02, 0x03}

	frame := make([]byte, 12+len(data))
	binary.BigEndian.PutUint64(frame[0:8], uint64(offset))
	binary.BigEndian.PutUint32(frame[8:12], uint32(len(data)))
	copy(frame[12:], data)

	readOffset := int64(binary.BigEndian.Uint64(frame[0:8]))
	assert.Equal(t, offset, readOffset)
}

func TestBinaryFrameEmptyData(t *testing.T) {
	offset := int64(0)
	data := []byte{}

	frame := make([]byte, 12)
	binary.BigEndian.PutUint64(frame[0:8], uint64(offset))
	binary.BigEndian.PutUint32(frame[8:12], uint32(len(data)))

	assert.Equal(t, uint32(0), binary.BigEndian.Uint32(frame[8:12]))
}

func TestProgressStructure(t *testing.T) {
	progress := &sftp.Progress{
		BytesTransferred: 512,
		TotalBytes:       1024,
		Percentage:       50.0,
		SpeedBps:         1024.0,
	}

	data, err := json.Marshal(progress)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"bytes_transferred":512`)
	assert.Contains(t, string(data), `"total_bytes":1024`)
	assert.Contains(t, string(data), `"percentage":50`)
	assert.Contains(t, string(data), `"speed_bps":1024`)
}

func TestProgressCalculationsWebSocket(t *testing.T) {
	tests := []struct {
		name             string
		bytesTransferred int64
		totalBytes       int64
		expectedPercent  float64
	}{
		{"zero size", 0, 0, 0},
		{"empty transfer", 0, 100, 0},
		{"half complete", 50, 100, 50},
		{"complete", 100, 100, 100},
		{"over complete", 150, 100, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var percent float64
			if tt.totalBytes > 0 {
				percent = float64(tt.bytesTransferred) / float64(tt.totalBytes) * 100
			}
			assert.InDelta(t, tt.expectedPercent, percent, 0.01)
		})
	}
}

// BenchmarkBinaryFrame benchmarks binary frame construction
func BenchmarkBinaryFrame(b *testing.B) {
	offset := int64(1024)
	data := make([]byte, 64*1024) // 64KB chunk

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame := make([]byte, 12+len(data))
		binary.BigEndian.PutUint64(frame[0:8], uint64(offset))
		binary.BigEndian.PutUint32(frame[8:12], uint32(len(data)))
		copy(frame[12:], data)
		_ = frame
	}
}

// BenchmarkJSONMarshal benchmarks message marshaling
func BenchmarkJSONMarshal(b *testing.B) {
	msg := sftp.FileTransferMessage{
		Type:      "progress",
		SessionID: "test-session",
		TransferID: "transfer-1",
		Progress: &sftp.Progress{
			BytesTransferred: 1024,
			TotalBytes:       2048,
			Percentage:       50.0,
			SpeedBps:         1024.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
