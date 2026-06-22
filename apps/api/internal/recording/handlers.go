// Package recording - HTTP handlers for session recordings
package recording

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles recording HTTP endpoints
type Handler struct {
	recorder      *Recorder
	storage       StorageBackend
	playback      *PlaybackServer
	manager       *RecordingManager
	auth          AuthMiddleware
}

// AuthMiddleware defines authentication interface
type AuthMiddleware interface {
	GetUserID(r *http.Request) (uuid.UUID, error)
	RequireAuth(next http.HandlerFunc) http.HandlerFunc
}

// RecordingManager manages recording lifecycle
type RecordingManager struct {
	storage    StorageBackend
	repository RecordingRepository
}

// RecordingRepository defines database operations
type RecordingRepository interface {
	Create(ctx context.Context, recording *Recording) error
	GetByID(ctx context.Context, id uuid.UUID) (*Recording, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Recording, error)
	ListBySession(ctx context.Context, sessionID uuid.UUID) ([]*Recording, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status RecordingStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, userID uuid.UUID, query string) ([]*Recording, error)
}

// Recording represents a recording entity
type Recording struct {
	ID             uuid.UUID       `json:"id"`
	SessionID      uuid.UUID       `json:"session_id"`
	UserID         uuid.UUID       `json:"user_id"`
	HostID         uuid.UUID       `json:"host_id"`
	Status         RecordingStatus `json:"status"`
	StartedAt      time.Time       `json:"started_at"`
	EndedAt        *time.Time      `json:"ended_at,omitempty"`
	Duration       time.Duration   `json:"duration"`
	Filename       string          `json:"filename"`
	FileSizeBytes  int64           `json:"file_size_bytes"`
	Checksum       string          `json:"checksum"`
	TerminalWidth  int             `json:"terminal_width"`
	TerminalHeight int             `json:"terminal_height"`
	TerminalType   string          `json:"terminal_type"`
	Shell          string          `json:"shell"`
	Encrypted      bool            `json:"encrypted"`
	RetentionDays  int             `json:"retention_days"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	CommandHistory []string        `json:"command_history,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// NewHandler creates a new recording handler
func NewHandler(storage StorageBackend, repository RecordingRepository, auth AuthMiddleware) *Handler {
	playback := NewPlaybackServer(storage, nil)
	manager := &RecordingManager{
		storage:    storage,
		repository: repository,
	}

	return &Handler{
		storage:  storage,
		playback: playback,
		manager:  manager,
		auth:     auth,
	}
}

// RegisterRoutes registers recording routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	requireAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(h.auth.RequireAuth(next.ServeHTTP))
	}

	r.Route("/api/sessions/{sessionID}/recordings", func(r chi.Router) {
		r.Use(requireAuthMiddleware)
		r.Post("/", h.StartRecording)
		r.Get("/", h.ListSessionRecordings)
	})

	r.Route("/api/recordings", func(r chi.Router) {
		r.Use(requireAuthMiddleware)
		r.Get("/", h.ListRecordings)
		r.Get("/search", h.SearchRecordings)
		r.Get("/{recordingID}", h.GetRecording)
		r.Get("/{recordingID}/play", h.PlayRecording)
		r.Get("/{recordingID}/download", h.DownloadRecording)
		r.Post("/{recordingID}/export", h.ExportRecording)
		r.Delete("/{recordingID}", h.DeleteRecording)
	})
}

// StartRecording starts a new recording for a session
func (h *Handler) StartRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	userID, err := h.auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		TerminalWidth  int    `json:"terminal_width"`
		TerminalHeight int    `json:"terminal_height"`
		TerminalType   string `json:"terminal_type"`
		Shell          string `json:"shell"`
		RetentionDays  int    `json:"retention_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use defaults
	if req.TerminalWidth == 0 {
		req.TerminalWidth = 80
	}
	if req.TerminalHeight == 0 {
		req.TerminalHeight = 24
	}
	if req.TerminalType == "" {
		req.TerminalType = "xterm-256color"
	}
	if req.RetentionDays == 0 {
		req.RetentionDays = 90
	}

	// Create recorder
	config := RecordingConfig{
		SessionID:      sessionID,
		UserID:         userID,
		TerminalWidth:  req.TerminalWidth,
		TerminalHeight: req.TerminalHeight,
		TerminalType:   req.TerminalType,
		Shell:          req.Shell,
		RetentionDays:  req.RetentionDays,
		EnableCompression: true,
		EnableEncryption:  true,
		RealTimeStream:    true,
	}

	recorder, err := NewRecorder(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create recorder: %v", err), http.StatusInternalServerError)
		return
	}

	// Save to database
	recording := &Recording{
		ID:             recorder.id,
		SessionID:      sessionID,
		UserID:         userID,
		Status:         StatusRecording,
		StartedAt:      time.Now(),
		TerminalWidth:  req.TerminalWidth,
		TerminalHeight: req.TerminalHeight,
		TerminalType:   req.TerminalType,
		Shell:          req.Shell,
		RetentionDays:  req.RetentionDays,
		Encrypted:      true,
		Filename:       fmt.Sprintf("%s.cast", recorder.id.String()),
	}

	if err := h.manager.repository.Create(ctx, recording); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save recording: %v", err), http.StatusInternalServerError)
		return
	}

	// Store recorder reference
	h.recorder = recorder

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recording_id": recorder.id,
		"status":         StatusRecording,
		"started_at":     recording.StartedAt,
		"expires_at":     recording.StartedAt.AddDate(0, 0, req.RetentionDays),
	})
}

// StopRecording stops the current recording
func (h *Handler) StopRecording(w http.ResponseWriter, r *http.Request) {
	if h.recorder == nil {
		http.Error(w, "No active recording", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	if err := h.recorder.Stop(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop recording: %v", err), http.StatusInternalServerError)
		return
	}

	// Get recording data
	data, err := h.recorder.GetData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get recording data: %v", err), http.StatusInternalServerError)
		return
	}

	info := h.recorder.GetInfo()

	// Upload to storage
	if err := h.storage.Upload(ctx, info.ID.String(), strings.NewReader(string(data)), int64(len(data)), "application/x-asciicast"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload recording: %v", err), http.StatusInternalServerError)
		return
	}

	// Update database
	if err := h.manager.repository.UpdateStatus(ctx, info.ID, StatusCompleted); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update recording status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recording_id": info.ID,
		"status":       StatusCompleted,
		"duration":     info.Duration.Seconds(),
		"file_size":    len(data),
		"checksum":     info.Checksum,
	})
}

// ListRecordings lists all recordings for the current user
func (h *Handler) ListRecordings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	recordings, err := h.manager.repository.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list recordings: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recordings": recordings,
		"limit":      limit,
		"offset":     offset,
		"total":      len(recordings),
	})
}

// ListSessionRecordings lists recordings for a specific session
func (h *Handler) ListSessionRecordings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	recordings, err := h.manager.repository.ListBySession(ctx, sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list recordings: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recordings)
}

// GetRecording gets a single recording by ID
func (h *Handler) GetRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recordingIDStr := chi.URLParam(r, "recordingID")
	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	recording, err := h.manager.repository.GetByID(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recording)
}

// PlayRecording streams a recording for playback
func (h *Handler) PlayRecording(w http.ResponseWriter, r *http.Request) {
	recordingIDStr := chi.URLParam(r, "recordingID")
	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	// Parse playback options
	opts := PlaybackOptions{}
	if speed := r.URL.Query().Get("speed"); speed != "" {
		opts.Speed, _ = strconv.ParseFloat(speed, 64)
	}
	if start := r.URL.Query().Get("start"); start != "" {
		opts.StartOffset, _ = strconv.ParseFloat(start, 64)
	}
	if end := r.URL.Query().Get("end"); end != "" {
		opts.EndOffset, _ = strconv.ParseFloat(end, 64)
	}
	opts.AutoPlay = r.URL.Query().Get("autoplay") == "true"
	opts.Loop = r.URL.Query().Get("loop") == "true"

	if err := h.playback.ServeRecording(w, r, recordingID, opts); err != nil {
		http.Error(w, fmt.Sprintf("Failed to serve recording: %v", err), http.StatusInternalServerError)
		return
	}
}

// DownloadRecording downloads a recording as .cast file
func (h *Handler) DownloadRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recordingIDStr := chi.URLParam(r, "recordingID")
	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	recording, err := h.manager.repository.GetByID(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	// Set download headers
	w.Header().Set("Content-Type", "application/x-asciicast")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", recording.Filename))
	w.Header().Set("Cache-Control", "private, max-age=3600")

	// Stream recording data
	reader, size, err := h.storage.Download(ctx, recordingID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve recording: %v", err), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	io.Copy(w, reader)
}

// ExportRecording exports a recording in various formats
func (h *Handler) ExportRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recordingIDStr := chi.URLParam(r, "recordingID")
	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "cast"
	}

	recording, err := h.manager.repository.GetByID(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	exporter := NewRecordingExporter(h.storage)

	switch format {
	case "cast":
		w.Header().Set("Content-Type", "application/x-asciicast")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", recording.Filename))
		if err := exporter.ExportToCast(ctx, recordingID, w); err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		}

	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := exporter.ExportToHTML(ctx, recordingID, w, recording.Filename); err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		}

	case "gif":
		w.Header().Set("Content-Type", "image/gif")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gif\"", strings.TrimSuffix(recording.Filename, ".cast")))
		if err := exporter.ExportToGIF(ctx, recordingID, w, 0); err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Unsupported format. Use: cast, html, gif", http.StatusBadRequest)
	}
}

// DeleteRecording soft-deletes a recording
func (h *Handler) DeleteRecording(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recordingIDStr := chi.URLParam(r, "recordingID")
	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	if err := h.manager.repository.Delete(ctx, recordingID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete recording: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchRecordings searches recordings with full-text search
func (h *Handler) SearchRecordings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	recordings, err := h.manager.repository.Search(ctx, userID, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":      query,
		"recordings": recordings,
		"total":      len(recordings),
	})
}
