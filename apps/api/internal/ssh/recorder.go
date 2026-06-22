package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// SessionRecorder records SSH sessions to asciinema format
type SessionRecorder struct {
	mu         sync.RWMutex
	recordings map[string]*Recording
	baseDir    string
}

// Recording represents an active recording
type Recording struct {
	ID         string
	SessionID  string
	UserID     string
	HostID     string
	StartTime  time.Time
	EndTime    *time.Time
	File       *os.File
	Events     []RecordEvent
	mu         sync.RWMutex
	closed     bool
}

// RecordEvent represents a single terminal event
type RecordEvent struct {
	Time      float64 `json:"time" yaml:"time"`
	EventType string  `json:"type" yaml:"type"`
	Data      string  `json:"data" yaml:"data"`
}

// AsciinemaHeader represents the asciinema file header
type AsciinemaHeader struct {
	Version   int               `json:"version" yaml:"version"`
	Width     int               `json:"width" yaml:"width"`
	Height    int               `json:"height" yaml:"height"`
	Timestamp int64             `json:"timestamp" yaml:"timestamp"`
	Env       map[string]string `json:"env" yaml:"env"`
	Title     string            `json:"title,omitempty" yaml:"title,omitempty"`
}

// NewSessionRecorder creates a new session recorder
func NewSessionRecorder(baseDir string) (*SessionRecorder, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recordings directory: %w", err)
	}

	return &SessionRecorder{
		recordings: make(map[string]*Recording),
		baseDir:    baseDir,
	}, nil
}

// StartRecording starts recording a session
func (sr *SessionRecorder) StartRecording(sessionID, userID, hostID string, width, height int) (*Recording, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.recordings[sessionID]; exists {
		return nil, fmt.Errorf("recording already exists for session %s", sessionID)
	}

	filename := fmt.Sprintf("%s/%s_%s.cast", sr.baseDir, sessionID, time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create recording file: %w", err)
	}

	recording := &Recording{
		ID:        fmt.Sprintf("rec_%s", sessionID),
		SessionID: sessionID,
		UserID:    userID,
		HostID:    hostID,
		StartTime: time.Now(),
		File:      file,
		Events:    make([]RecordEvent, 0, 1000),
	}

	// Write header
	header := AsciinemaHeader{
		Version:   2,
		Width:     width,
		Height:    height,
		Timestamp: time.Now().Unix(),
		Env: map[string]string{
			"SHELL": "/bin/bash",
			"TERM":  "xterm-256color",
		},
		Title: fmt.Sprintf("SSH Session %s", sessionID[:8]),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to marshal header: %w", err)
	}

	if _, err := file.Write(headerJSON); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write newline: %w", err)
	}

	sr.recordings[sessionID] = recording
	return recording, nil
}

// RecordOutput records SSH stdout output
func (r *Recording) RecordOutput(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	elapsed := time.Since(r.StartTime).Seconds()
	event := RecordEvent{
		Time:      elapsed,
		EventType: "o",
		Data:      string(data),
	}

	r.Events = append(r.Events, event)

	// Write to file immediately
	eventJSON, _ := json.Marshal(event)
	r.File.Write(eventJSON)
	r.File.WriteString("\n")
}

// RecordInput records user input
func (r *Recording) RecordInput(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	elapsed := time.Since(r.StartTime).Seconds()
	event := RecordEvent{
		Time:      elapsed,
		EventType: "i",
		Data:      string(data),
	}

	r.Events = append(r.Events, event)

	// Write to file
	eventJSON, _ := json.Marshal(event)
	r.File.Write(eventJSON)
	r.File.WriteString("\n")
}

// Stop stops the recording
func (r *Recording) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	now := time.Now()
	r.EndTime = &now

	return r.File.Close()
}

// GetRecording returns a recording by session ID
func (sr *SessionRecorder) GetRecording(sessionID string) (*Recording, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	rec, ok := sr.recordings[sessionID]
	return rec, ok
}

// StopRecording stops recording for a session
func (sr *SessionRecorder) StopRecording(sessionID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	rec, ok := sr.recordings[sessionID]
	if !ok {
		return fmt.Errorf("no recording found for session %s", sessionID)
	}

	if err := rec.Stop(); err != nil {
		return err
	}

	delete(sr.recordings, sessionID)
	return nil
}

// ListRecordings returns all recordings
func (sr *SessionRecorder) ListRecordings() []*Recording {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	recordings := make([]*Recording, 0, len(sr.recordings))
	for _, rec := range sr.recordings {
		recordings = append(recordings, rec)
	}
	return recordings
}

// CleanupOldRecordings removes recordings older than maxAge
func (sr *SessionRecorder) CleanupOldRecordings(maxAge time.Duration) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	now := time.Now()
	for id, rec := range sr.recordings {
		if rec.EndTime != nil && now.Sub(*rec.EndTime) > maxAge {
			rec.Stop()
			delete(sr.recordings, id)
		}
	}

	// Also cleanup files
	entries, err := os.ReadDir(sr.baseDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			os.Remove(fmt.Sprintf("%s/%s", sr.baseDir, entry.Name()))
		}
	}

	return nil
}

// GetRecordingFile returns the file path for a recording
func (r *Recording) GetRecordingFile() string {
	if r.File != nil {
		return r.File.Name()
	}
	return ""
}

// RecordingStats represents recording statistics
type RecordingStats struct {
	TotalRecordings   int           `json:"total_recordings"`
	ActiveRecordings  int           `json:"active_recordings"`
	TotalDuration     time.Duration `json:"total_duration"`
	TotalSize         int64         `json:"total_size_bytes"`
	StoragePath       string        `json:"storage_path"`
}

// GetStats returns recording statistics
func (sr *SessionRecorder) GetStats() (*RecordingStats, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	stats := &RecordingStats{
		TotalRecordings: len(sr.recordings),
		StoragePath:     sr.baseDir,
	}

	for _, rec := range sr.recordings {
		if rec.EndTime == nil {
			stats.ActiveRecordings++
			stats.TotalDuration += time.Since(rec.StartTime)
		} else {
			stats.TotalDuration += rec.EndTime.Sub(rec.StartTime)
		}
	}

	// Calculate total file size
	entries, err := os.ReadDir(sr.baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		stats.TotalSize += info.Size()
	}

	return stats, nil
}
