// Package recording - Playback server for session recordings
package recording

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// PlaybackServer handles recording playback streaming
type PlaybackServer struct {
	storage    StorageBackend
	keyManager KeyManager
}

// KeyManager handles encryption key retrieval
type KeyManager interface {
	GetKey(ctx context.Context, keyID string) ([]byte, error)
}

// NewPlaybackServer creates a new playback server
func NewPlaybackServer(storage StorageBackend, keyManager KeyManager) *PlaybackServer {
	return &PlaybackServer{
		storage:    storage,
		keyManager: keyManager,
	}
}

// PlaybackOptions contains playback configuration
type PlaybackOptions struct {
	Speed        float64 `json:"speed,omitempty"`        // Playback speed (1.0 = normal)
	StartOffset  float64 `json:"start_offset,omitempty"` // Start from offset in seconds
	EndOffset    float64 `json:"end_offset,omitempty"`   // End at offset in seconds
	AutoPlay     bool    `json:"autoplay,omitempty"`     // Auto-start playback
	Loop         bool    `json:"loop,omitempty"`         // Loop playback
}

// ServeRecording serves a recording for playback
func (ps *PlaybackServer) ServeRecording(w http.ResponseWriter, r *http.Request, recordingID uuid.UUID, opts PlaybackOptions) error {
	ctx := r.Context()

	// Retrieve recording data
	reader, size, err := ps.storage.Download(ctx, recordingID.String())
	if err != nil {
		return fmt.Errorf("failed to retrieve recording: %w", err)
	}
	defer reader.Close()

	// Detect compression
	var dataReader io.Reader = reader
	bufReader := bufio.NewReader(reader)
	peek, err := bufReader.Peek(2)
	if err != nil {
		return fmt.Errorf("failed to peek data: %w", err)
	}

	// Check for gzip magic number
	if peek[0] == 0x1f && peek[1] == 0x8b {
		gzipReader, err := gzip.NewReader(bufReader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		dataReader = gzipReader
	} else {
		dataReader = bufReader
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/x-asciicast")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, max-age=3600")

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		return ps.serveRange(w, r, dataReader, rangeHeader, size)
	}

	// Apply playback options
	if opts.Speed != 0 || opts.StartOffset != 0 || opts.EndOffset != 0 {
		return ps.serveWithOptions(w, dataReader, opts)
	}

	// Stream directly
	_, err = io.Copy(w, dataReader)
	return err
}

// ServeHTMLOverlay serves an HTML overlay player
func (ps *PlaybackServer) ServeHTMLOverlay(w http.ResponseWriter, r *http.Request, recordingID uuid.UUID) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Session Recording %s</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle/asciinema-player.css">
    <style>
        body { margin: 0; background: #000; }
        #player { width: 100vw; height: 100vh; }
    </style>
</head>
<body>
    <div id="player"></div>
    <script src="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle/asciinema-player.min.js"></script>
    <script>
        AsciinemaPlayer.create('/api/recordings/%s/play', document.getElementById('player'), {
            autoPlay: true,
            loop: false,
            theme: 'monokai',
            fit: 'both'
        });
    </script>
</body>
</html>`, recordingID.String()[:8], recordingID.String())

	w.Write([]byte(html))
}

// serveRange handles HTTP range requests
func (ps *PlaybackServer) serveRange(w http.ResponseWriter, r *http.Request, reader io.Reader, rangeHeader string, totalSize int64) error {
	// Parse range header (simplified)
	parts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range header")
	}

	start, _ := strconv.ParseInt(parts[0], 10, 64)
	end := totalSize - 1
	if parts[1] != "" {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	contentLength := end - start + 1

	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Skip to start position
	if _, err := io.CopyN(io.Discard, reader, start); err != nil {
		return err
	}

	// Copy range
	_, err := io.CopyN(w, reader, contentLength)
	return err
}

// serveWithOptions applies playback options during streaming
func (ps *PlaybackServer) serveWithOptions(w http.ResponseWriter, reader io.Reader, opts PlaybackOptions) error {
	scanner := bufio.NewScanner(reader)
	
	// Read and modify header
	if !scanner.Scan() {
		return fmt.Errorf("empty recording")
	}

	headerLine := scanner.Text()
	var header AsciinemaHeader
	if err := json.Unmarshal([]byte(headerLine), &header); err != nil {
		return fmt.Errorf("invalid header: %w", err)
	}

	// Modify header with playback options
	if opts.Speed != 0 {
		// Speed modification is handled client-side
		header.Title = fmt.Sprintf("%s (speed: %.1fx)", header.Title, opts.Speed)
	}

	modifiedHeader, _ := json.Marshal(header)
	fmt.Fprintf(w, "%s\n", modifiedHeader)

	// Stream events with offset filtering
	for scanner.Scan() {
		line := scanner.Text()
		
		var event []interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip invalid events
		}

		if len(event) < 2 {
			continue
		}

		timestamp, ok := event[0].(float64)
		if !ok {
			continue
		}

		// Filter by time range
		if opts.StartOffset > 0 && timestamp < opts.StartOffset {
			continue
		}
		if opts.EndOffset > 0 && timestamp > opts.EndOffset {
			break
		}

		// Adjust timestamp for start offset
		if opts.StartOffset > 0 {
			event[0] = timestamp - opts.StartOffset
		}

		// Adjust speed
		if opts.Speed > 0 && opts.Speed != 1.0 {
			event[0] = timestamp / opts.Speed
		}

		modifiedEvent, _ := json.Marshal(event)
		fmt.Fprintf(w, "%s\n", modifiedEvent)
	}

	return scanner.Err()
}

// RealTimeStreamer handles real-time recording streaming
type RealTimeStreamer struct {
	recorder *Recorder
	clients  map[string]chan AsciinemaEvent
	mu       sync.RWMutex
}

// NewRealTimeStreamer creates a new real-time streamer
func NewRealTimeStreamer(recorder *Recorder) *RealTimeStreamer {
	rts := &RealTimeStreamer{
		recorder: recorder,
		clients:  make(map[string]chan AsciinemaEvent),
	}
	
	// Set up streaming channel
	streamCh := make(chan AsciinemaEvent, 100)
	recorder.SetStreamChannel(streamCh)
	
	// Start forwarding
	go rts.forwardEvents(streamCh)
	
	return rts
}

// forwardEvents forwards events to all connected clients
func (rts *RealTimeStreamer) forwardEvents(ch <-chan AsciinemaEvent) {
	for event := range ch {
		rts.mu.RLock()
		for _, clientCh := range rts.clients {
			select {
			case clientCh <- event:
			default:
				// Client buffer full, skip
			}
		}
		rts.mu.RUnlock()
	}
}

// AddClient adds a streaming client
func (rts *RealTimeStreamer) AddClient(clientID string) <-chan AsciinemaEvent {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	ch := make(chan AsciinemaEvent, 100)
	rts.clients[clientID] = ch
	return ch
}

// RemoveClient removes a streaming client
func (rts *RealTimeStreamer) RemoveClient(clientID string) {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	if ch, ok := rts.clients[clientID]; ok {
		close(ch)
		delete(rts.clients, clientID)
	}
}

// RecordingExporter handles export functionality
type RecordingExporter struct {
	storage StorageBackend
}

// NewRecordingExporter creates a new exporter
func NewRecordingExporter(storage StorageBackend) *RecordingExporter {
	return &RecordingExporter{storage: storage}
}

// ExportToCast exports a recording as a .cast file
func (e *RecordingExporter) ExportToCast(ctx context.Context, recordingID uuid.UUID, w io.Writer) error {
	reader, _, err := e.storage.Download(ctx, recordingID.String())
	if err != nil {
		return fmt.Errorf("failed to retrieve recording: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(w, reader)
	return err
}

// ExportToGIF exports a recording as an animated GIF (requires external tool)
func (e *RecordingExporter) ExportToGIF(ctx context.Context, recordingID uuid.UUID, w io.Writer, width int) error {
	// This would integrate with a tool like `agg` (asciinema gif generator)
	// For now, return an error indicating external tool needed
	return fmt.Errorf("GIF export requires external tool: install `agg` (asciinema gif generator)")
}

// ExportToHTML exports a recording as a standalone HTML file
func (e *RecordingExporter) ExportToHTML(ctx context.Context, recordingID uuid.UUID, w io.Writer, title string) error {
	reader, _, err := e.storage.Download(ctx, recordingID.String())
	if err != nil {
		return fmt.Errorf("failed to retrieve recording: %w", err)
	}
	defer reader.Close()

	// Encode recording data as base64 for embedding
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return err
	}
	recordingData := base64.StdEncoding.EncodeToString(buf.Bytes())

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle/asciinema-player.css">
    <style>
        body { margin: 0; padding: 20px; background: #1a1a1a; color: #fff; font-family: sans-serif; }
        #player { width: 100%%; max-width: 960px; margin: 0 auto; }
    </style>
</head>
<body>
    <h1>%s</h1>
    <div id="player"></div>
    <script src="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle/asciinema-player.min.js"></script>
    <script>
        const data = atob('%s');
        const blob = new Blob([data], { type: 'application/x-asciicast' });
        const url = URL.createObjectURL(blob);
        AsciinemaPlayer.create(url, document.getElementById('player'), {
            autoPlay: false,
            theme: 'monokai'
        });
    </script>
</body>
</html>`, title, title, recordingData)

	_, err = w.Write([]byte(html))
	return err
}
