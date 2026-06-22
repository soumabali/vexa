// Package recording_test - Tests for session recording
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/recording"
)

func TestNewRecorder(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		TerminalWidth:  80,
		TerminalHeight: 24,
		TerminalType:   "xterm-256color",
		Shell:          "/bin/bash",
		RetentionDays:  90,
		EnableCompression: true,
		EnableEncryption:  true,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)
	require.NotNil(t, recorder)

	info := recorder.GetInfo()
	assert.Equal(t, recording.StatusRecording, info.Status)
	assert.Equal(t, config.TerminalWidth, info.TerminalWidth)
	assert.Equal(t, config.TerminalHeight, info.TerminalHeight)
	assert.Equal(t, config.TerminalType, info.TerminalType)
	assert.Equal(t, config.RetentionDays, info.RetentionDays)
	assert.True(t, info.Encrypted)
	assert.WithinDuration(t, time.Now(), info.StartedAt, time.Second)
}

func TestRecorder_RecordOutput(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  false,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	// Record some output
	output := []byte("Hello, World!\n")
	err = recorder.RecordOutput(output)
	require.NoError(t, err)

	// Record more output
	output2 := []byte("Second line of output\n")
	err = recorder.RecordOutput(output2)
	require.NoError(t, err)

	info := recorder.GetInfo()
	assert.Equal(t, 2, info.EventCount)

	// Verify data exists
	data, err := recorder.GetData()
	require.NoError(t, err)
	assert.True(t, len(data) > 0)
}

func TestRecorder_PauseResume(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  false,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	// Pause recording
	err = recorder.Pause()
	require.NoError(t, err)

	// Should fail to record while paused
	err = recorder.RecordOutput([]byte("test"))
	assert.Error(t, err)

	// Resume recording
	err = recorder.Resume()
	require.NoError(t, err)

	// Should succeed after resume
	err = recorder.RecordOutput([]byte("test"))
	require.NoError(t, err)
}

func TestRecorder_Stop(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  false,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	// Record some data
	err = recorder.RecordOutput([]byte("Test output\n"))
	require.NoError(t, err)

	// Stop recording
	err = recorder.Stop()
	require.NoError(t, err)

	info := recorder.GetInfo()
	assert.Equal(t, recording.StatusCompleted, info.Status)
	assert.NotNil(t, info.EndedAt)
	assert.True(t, info.Duration > 0)

	// Get checksum
	checksum := recorder.GetChecksum()
	assert.NotEmpty(t, checksum)
	assert.Equal(t, 64, len(checksum)) // SHA-256 hex = 64 chars
}

func TestRecorder_GetData(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  false,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	// Record output
	output := []byte("Line 1\nLine 2\nLine 3\n")
	err = recorder.RecordOutput(output)
	require.NoError(t, err)

	// Stop
	err = recorder.Stop()
	require.NoError(t, err)

	// Get data
	data, err := recorder.GetData()
	require.NoError(t, err)

	// Verify header exists
	lines := bytes.Split(data, []byte("\n"))
	require.True(t, len(lines) > 1)

	// First line should be JSON header
	var header map[string]interface{}
	err = json.Unmarshal(lines[0], &header)
	require.NoError(t, err)
	assert.Equal(t, float64(2), header["version"])
	assert.NotNil(t, header["width"])
	assert.NotNil(t, header["height"])
}

func TestRecorder_WithCompression(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: true,
		EnableEncryption:  false,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	// Record lots of repetitive data (good for compression)
	for i := 0; i < 100; i++ {
		err = recorder.RecordOutput([]byte("Repetitive output line that should compress well\n"))
		require.NoError(t, err)
	}

	err = recorder.Stop()
	require.NoError(t, err)

	data, err := recorder.GetData()
	require.NoError(t, err)
	assert.True(t, len(data) > 0)
}

func TestRecorder_WithEncryption(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  true,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	err = recorder.RecordOutput([]byte("Secret output that should be encrypted\n"))
	require.NoError(t, err)

	err = recorder.Stop()
	require.NoError(t, err)

	data, err := recorder.GetData()
	require.NoError(t, err)

	// Data should be encrypted (not plain text)
	assert.NotContains(t, string(data), "Secret output")
}

func TestRecordingInfo(t *testing.T) {
	config := recording.RecordingConfig{
		SessionID:      uuid.MustParse("12345678-1234-1234-1234-123456789abc"),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		TerminalWidth:  120,
		TerminalHeight: 40,
		TerminalType:   "screen-256color",
		Shell:          "/bin/zsh",
		RetentionDays:  30,
	}

	recorder, err := recording.NewRecorder(config)
	require.NoError(t, err)

	info := recorder.GetInfo()
	assert.Equal(t, config.SessionID, info.SessionID)
	assert.Equal(t, config.UserID, info.UserID)
	assert.Equal(t, config.HostID, info.HostID)
	assert.Equal(t, 120, info.TerminalWidth)
	assert.Equal(t, 40, info.TerminalHeight)
	assert.Equal(t, "screen-256color", info.TerminalType)
	assert.Equal(t, 30, info.RetentionDays)
}

// Benchmark tests
func BenchmarkRecorder_RecordOutput(b *testing.B) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  false,
	}

	recorder, _ := recording.NewRecorder(config)
	output := []byte("Benchmark output line\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.RecordOutput(output)
	}
}

func BenchmarkRecorder_RecordOutput_Compressed(b *testing.B) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: true,
		EnableEncryption:  false,
	}

	recorder, _ := recording.NewRecorder(config)
	output := []byte("Benchmark output line\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.RecordOutput(output)
	}
}

func BenchmarkRecorder_RecordOutput_Encrypted(b *testing.B) {
	config := recording.RecordingConfig{
		SessionID:      uuid.New(),
		UserID:         uuid.New(),
		HostID:         uuid.New(),
		EnableCompression: false,
		EnableEncryption:  true,
	}

	recorder, _ := recording.NewRecorder(config)
	output := []byte("Benchmark output line\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.RecordOutput(output)
	}
}

// Integration test with storage
func TestStorageManager_SaveAndRetrieve(t *testing.T) {
	// Create local storage for testing
	storage, err := recording.NewLocalStorage("/tmp/test-recordings")
	require.NoError(t, err)
	require.NotNil(t, storage)

	recordingID := uuid.New()
	data := []byte("Test recording data")

	ctx := context.Background()

	// Save
	err = storage.Upload(ctx, recordingID.String(), bytes.NewReader(data), int64(len(data)), "application/x-asciicast")
	require.NoError(t, err)

	// Retrieve
	reader, size, err := storage.Download(ctx, recordingID.String())
	require.NoError(t, err)
	defer reader.Close()

	assert.Equal(t, int64(len(data)), size)

	retrieved, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Cleanup
	err = storage.Delete(ctx, recordingID.String())
	require.NoError(t, err)

	// Verify deleted
	exists, err := storage.Exists(ctx, recordingID.String())
	require.NoError(t, err)
	assert.False(t, exists)
}

// Mock storage for testing
type MockStorage struct {
	data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{data: make(map[string][]byte)}
}

func (m *MockStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	data, _ := io.ReadAll(reader)
	m.data[key] = data
	return nil
}

func (m *MockStorage) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	data, ok := m.data[key]
	if !ok {
		return nil, 0, fmt.Errorf("not found")
	}
	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockStorage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("http://mock/%s", key), nil
}

func TestPlaybackServer(t *testing.T) {
	storage := NewMockStorage()
	playback := recording.NewPlaybackServer(storage, nil)
	require.NotNil(t, playback)
}
