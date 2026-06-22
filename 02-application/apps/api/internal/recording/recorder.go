// Package recording provides session recording in asciinema v2 format
package recording

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RecordingStatus represents the current state of a recording
type RecordingStatus string

const (
	StatusRecording RecordingStatus = "recording"
	StatusPaused    RecordingStatus = "paused"
	StatusCompleted RecordingStatus = "completed"
	StatusFailed    RecordingStatus = "failed"
	StatusDeleted   RecordingStatus = "deleted"
)

// AsciinemaHeader represents the asciinema v2 header
type AsciinemaHeader struct {
	Version      int      `json:"version"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Timestamp    int64    `json:"timestamp,omitempty"`
	Title        string   `json:"title,omitempty"`
	Env          *EnvVars `json:"env,omitempty"`
	Command      string   `json:"command,omitempty"`
}

// EnvVars contains terminal environment variables
type EnvVars struct {
	Shell string `json:"SHELL,omitempty"`
	Term  string `json:"TERM,omitempty"`
}

// AsciinemaEvent represents a single terminal event
type AsciinemaEvent struct {
	Time float64 `json:"time"`
	Type string  `json:"type"` // "o" = output, "i" = input
	Data string  `json:"data"`
}

// RecordingConfig contains configuration for a recording
type RecordingConfig struct {
	SessionID      uuid.UUID
	UserID         uuid.UUID
	HostID         uuid.UUID
	TerminalWidth  int
	TerminalHeight int
	TerminalType   string
	Shell          string
	RetentionDays  int
	EnableCompression bool
	EnableEncryption  bool
	RealTimeStream    bool
}

// Recorder manages the recording lifecycle
type Recorder struct {
	id          uuid.UUID
	config      RecordingConfig
	status      RecordingStatus
	startTime   time.Time
	endTime     *time.Time
	duration    time.Duration

	// I/O
	buffer      *bytes.Buffer
	gzipWriter  *gzip.Writer
	cipherWriter io.Writer
	output      io.Writer

	// Content tracking
	events      []AsciinemaEvent
	eventMu     sync.RWMutex
	commandHistory []string
	contentBuffer  strings.Builder

	// Encryption
	aesGCM      cipher.AEAD
	encKey      []byte

	// Streaming
	streamCh    chan<- AsciinemaEvent
	streamMu    sync.RWMutex

	// Control
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	done        chan struct{}
}

// NewRecorder creates a new session recorder
func NewRecorder(config RecordingConfig) (*Recorder, error) {
	if config.TerminalWidth == 0 {
		config.TerminalWidth = 80
	}
	if config.TerminalHeight == 0 {
		config.TerminalHeight = 24
	}
	if config.TerminalType == "" {
		config.TerminalType = "xterm-256color"
	}
	if config.Shell == "" {
		config.Shell = "/bin/bash"
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 90
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Recorder{
		id:        uuid.New(),
		config:    config,
		status:    StatusRecording,
		startTime: time.Now(),
		buffer:    &bytes.Buffer{},
		events:    make([]AsciinemaEvent, 0, 1000),
		done:      make(chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize output writer
	var output io.Writer = r.buffer

	// Setup compression
	if config.EnableCompression {
		var err error
		r.gzipWriter, err = gzip.NewWriterLevel(output, gzip.BestSpeed)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		output = r.gzipWriter
	}

	// Setup encryption
	if config.EnableEncryption {
		key := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		r.encKey = key

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher: %w", err)
		}

		r.aesGCM, err = cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create AES-GCM: %w", err)
		}

		// Wrap writer with encryption
		r.cipherWriter = &encryptedWriter{
			writer: output,
			aead:   r.aesGCM,
		}
		output = r.cipherWriter
	}

	r.output = output

	// Write asciinema header
	if err := r.writeHeader(); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	return r, nil
}

// writeHeader writes the asciinema v2 header
func (r *Recorder) writeHeader() error {
	header := AsciinemaHeader{
		Version:   2,
		Width:     r.config.TerminalWidth,
		Height:    r.config.TerminalHeight,
		Timestamp: r.startTime.Unix(),
		Title:     fmt.Sprintf("Session %s", r.config.SessionID.String()[:8]),
		Env: &EnvVars{
			Shell: r.config.Shell,
			Term:  r.config.TerminalType,
		},
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.output, "%s\n", headerJSON)
	return err
}

// RecordOutput records terminal output
func (r *Recorder) RecordOutput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusRecording {
		return fmt.Errorf("recording is not active (status: %s)", r.status)
	}

	elapsed := time.Since(r.startTime).Seconds()

	event := AsciinemaEvent{
		Time: elapsed,
		Type: "o",
		Data: string(data),
	}

	// Write to output
	eventJSON, err := json.Marshal([]interface{}{event.Time, event.Type, event.Data})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := fmt.Fprintf(r.output, "%s\n", eventJSON); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// Track event
	r.eventMu.Lock()
	r.events = append(r.events, event)
	r.contentBuffer.Write(data)
	r.eventMu.Unlock()

	// Stream if enabled
	if r.config.RealTimeStream {
		r.streamMu.RLock()
		if r.streamCh != nil {
			select {
			case r.streamCh <- event:
			default:
				// Channel full, drop event
			}
		}
		r.streamMu.RUnlock()
	}

	return nil
}

// RecordInput records user input (optional, for audit)
func (r *Recorder) RecordInput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusRecording {
		return fmt.Errorf("recording is not active")
	}

	elapsed := time.Since(r.startTime).Seconds()

	event := AsciinemaEvent{
		Time: elapsed,
		Type: "i",
		Data: string(data),
	}

	eventJSON, err := json.Marshal([]interface{}{event.Time, event.Type, event.Data})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.output, "%s\n", eventJSON)

	// Track command history (simple heuristic)
	if data[len(data)-1] == '\r' || data[len(data)-1] == '\n' {
		cmd := strings.TrimSpace(string(data))
		if cmd != "" {
			r.commandHistory = append(r.commandHistory, cmd)
		}
	}

	return err
}

// Pause temporarily stops recording
func (r *Recorder) Pause() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusRecording {
		return fmt.Errorf("cannot pause: status is %s", r.status)
	}

	r.status = StatusPaused
	return nil
}

// Resume continues recording
func (r *Recorder) Resume() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusPaused {
		return fmt.Errorf("cannot resume: status is %s", r.status)
	}

	r.status = StatusRecording
	return nil
}

// Stop finalizes the recording
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusRecording && r.status != StatusPaused {
		return fmt.Errorf("cannot stop: status is %s", r.status)
	}

	r.status = StatusCompleted
	now := time.Now()
	r.endTime = &now
	r.duration = now.Sub(r.startTime)

	// Flush gzip
	if r.gzipWriter != nil {
		if err := r.gzipWriter.Close(); err != nil {
			return fmt.Errorf("failed to close gzip: %w", err)
		}
	}

	close(r.done)
	return nil
}

// GetData returns the recorded data
func (r *Recorder) GetData() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.status == StatusRecording {
		// Flush gzip for streaming
		if r.gzipWriter != nil {
			r.gzipWriter.Flush()
		}
	}

	return r.buffer.Bytes(), nil
}

// GetChecksum calculates SHA256 of recording data
func (r *Recorder) GetChecksum() string {
	data, _ := r.GetData()
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GetInfo returns recording metadata
func (r *Recorder) GetInfo() RecordingInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info := RecordingInfo{
		ID:             r.id,
		SessionID:      r.config.SessionID,
		UserID:         r.config.UserID,
		HostID:         r.config.HostID,
		Status:         r.status,
		StartedAt:      r.startTime,
		Duration:       r.duration,
		TerminalWidth:  r.config.TerminalWidth,
		TerminalHeight: r.config.TerminalHeight,
		TerminalType:   r.config.TerminalType,
		RetentionDays:  r.config.RetentionDays,
		Encrypted:      r.config.EnableEncryption,
		FileSizeBytes:  int64(r.buffer.Len()),
		Checksum:       r.GetChecksum(),
	}

	if r.endTime != nil {
		info.EndedAt = *r.endTime
	}

	r.eventMu.RLock()
	info.EventCount = len(r.events)
	info.CommandHistory = r.commandHistory
	r.eventMu.RUnlock()

	return info
}

// SetStreamChannel sets the real-time streaming channel
func (r *Recorder) SetStreamChannel(ch chan<- AsciinemaEvent) {
	r.streamMu.Lock()
	defer r.streamMu.Unlock()
	r.streamCh = ch
}

// RecordingInfo contains recording metadata
type RecordingInfo struct {
	ID             uuid.UUID
	SessionID      uuid.UUID
	UserID         uuid.UUID
	HostID         uuid.UUID
	Status         RecordingStatus
	StartedAt      time.Time
	EndedAt        time.Time
	Duration       time.Duration
	TerminalWidth  int
	TerminalHeight int
	TerminalType   string
	RetentionDays  int
	Encrypted      bool
	FileSizeBytes  int64
	Checksum       string
	EventCount     int
	CommandHistory []string
}

// encryptedWriter wraps an io.Writer with AES-GCM encryption
type encryptedWriter struct {
	writer io.Writer
	aead   cipher.AEAD
	nonce  []byte
	mu     sync.Mutex
}

func (w *encryptedWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.nonce == nil {
		w.nonce = make([]byte, w.aead.NonceSize())
		if _, err := io.ReadFull(rand.Reader, w.nonce); err != nil {
			return 0, err
		}
		if _, err := w.writer.Write(w.nonce); err != nil {
			return 0, err
		}
	}

	ciphertext := w.aead.Seal(nil, w.nonce, p, nil)
	// Increment nonce for next write
	for i := len(w.nonce) - 1; i >= 0; i-- {
		w.nonce[i]++
		if w.nonce[i] != 0 {
			break
		}
	}

	return w.writer.Write(ciphertext)
}

// Size returns the current buffer size
func (r *Recorder) Size() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(r.buffer.Len())
}


