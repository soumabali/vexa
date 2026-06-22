package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/gateway"
)

var (
	ErrSessionNotFound    = errors.New("SFTP session not found")
	ErrSessionExists      = errors.New("SFTP session already exists")
	ErrTransferInProgress = errors.New("transfer already in progress")
	ErrTransferNotFound   = errors.New("transfer not found")
	ErrBandwidthExceeded  = errors.New("bandwidth limit exceeded")
	ErrFileTooLarge       = errors.New("file exceeds maximum size limit")
	ErrInvalidOffset      = errors.New("invalid offset")
	ErrReadFailed         = errors.New("read operation failed")
	ErrWriteFailed        = errors.New("write operation failed")
)

const (
	MaxFileSize     = 10 * 1024 * 1024 * 1024 // 10GB
	MaxBandwidth    = 100 * 1024 * 1024       // 100MB/s
	DefaultChunkSize = 64 * 1024              // 64KB chunks
	BandwidthWindow  = time.Second
)

type SFTPService struct {
	sessions  map[string]*SFTPClient
	mu        sync.RWMutex
	sanitizer *PathSanitizer
	sshGateway *gateway.SSHGateway
	auditLogger *audit.Logger

	// Bandwidth monitoring
	bandwidthStats map[string]*BandwidthStats
	bandwidthMu    sync.RWMutex

	// Active transfers
	transfers  map[string]*Transfer
	transferMu sync.RWMutex

	// Rate limiters per session
	rateLimiters map[string]*RateLimiter
	rateMu       sync.RWMutex
}

type SFTPClient struct {
	ID         string
	Conn       *ssh.Client
	SFTP       *sftp.Client
	RootDir    string
	UserID     uuid.UUID
	HostID     uuid.UUID
	CreatedAt  time.Time
	LastUsed   time.Time
	active     bool
	mu         sync.RWMutex
}

type Transfer struct {
	ID        string
	SessionID string
	Path      string
	Type      string // upload or download
	Size      int64
	Offset    int64
	StartedAt time.Time
	Completed bool
	Cancelled bool
	mu        sync.RWMutex
}

type RateLimiter struct {
	bytesSent     int64
	bytesReceived int64
	lastWindow    time.Time
	mu            sync.Mutex
}

type BandwidthStats struct {
	BytesSent     int64
	BytesReceived int64
	SentSpeed     float64
	ReceiveSpeed  float64
	LastUpdated   time.Time
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

type Progress struct {
	BytesTransferred int64   `json:"bytes_transferred"`
	TotalBytes       int64   `json:"total_bytes"`
	Percentage       float64 `json:"percentage"`
	SpeedBps         float64 `json:"speed_bps"`
}

func NewSFTPService(sshGateway *gateway.SSHGateway, auditLogger *audit.Logger) *SFTPService {
	return &SFTPService{
		sessions:       make(map[string]*SFTPClient),
		sanitizer:      NewPathSanitizer(),
		sshGateway:     sshGateway,
		auditLogger:    auditLogger,
		bandwidthStats: make(map[string]*BandwidthStats),
		transfers:      make(map[string]*Transfer),
		rateLimiters:   make(map[string]*RateLimiter),
	}
}

// CreateSession creates a new SFTP session by connecting to a host via SSH
func (s *SFTPService) CreateSession(hostID, userID uuid.UUID, host, sshUser, password string, port int, rootDir string) (*SFTPClient, error) {
	ctx := context.Background()

	// Get SSH connection from gateway
	conn, err := s.sshGateway.GetConnection(ctx, host, port, sshUser, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH connection: %w", err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(conn.Client)
	if err != nil {
		s.sshGateway.ReleaseConnection(conn)
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	session := &SFTPClient{
		ID:        uuid.New().String(),
		Conn:      conn.Client,
		SFTP:      sftpClient,
		RootDir:   rootDir,
		UserID:    userID,
		HostID:    hostID,
		CreatedAt: time.Now().UTC(),
		LastUsed:  time.Now().UTC(),
		active:    true,
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	// Initialize bandwidth stats
	s.bandwidthMu.Lock()
	s.bandwidthStats[session.ID] = &BandwidthStats{LastUpdated: time.Now().UTC()}
	s.bandwidthMu.Unlock()

	// Initialize rate limiter
	s.rateMu.Lock()
	s.rateLimiters[session.ID] = &RateLimiter{lastWindow: time.Now().UTC()}
	s.rateMu.Unlock()

	// Audit log
	if s.auditLogger != nil {
		s.auditLogger.Log(
			"sftp_session_created",
			&userID, nil, "",
			map[string]interface{}{
				"session_id": session.ID,
				"host_id":    hostID.String(),
				"host":       host,
			},
		)
	}

	return session, nil
}

// ListDirectory lists files in a directory with path sanitization
func (s *SFTPService) ListDirectory(sessionID, path string) ([]FileInfo, error) {
	session, err := s.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	sanitizedPath, err := s.sanitizer.SanitizePath(session.RootDir, path)
	if err != nil {
		s.logAudit(session.UserID, "sftp_list_directory_denied", map[string]interface{}{
			"session_id": sessionID,
			"path":       path,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("path sanitization failed: %w", err)
	}

	entries, err := session.SFTP.ReadDir(sanitizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			Mode:    entry.Mode(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	s.updateLastUsed(sessionID)
	s.logAudit(session.UserID, "sftp_list_directory", map[string]interface{}{
		"session_id": sessionID,
		"path":       sanitizedPath,
		"count":      len(files),
	})

	return files, nil
}

// Upload uploads data from a reader to a file path with progress tracking
func (s *SFTPService) Upload(sessionID, path string, reader io.Reader, size int64) (*Transfer, error) {
	if size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	session, err := s.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	sanitizedPath, err := s.sanitizer.SanitizePath(session.RootDir, path)
	if err != nil {
		s.logAudit(session.UserID, "sftp_upload_denied", map[string]interface{}{
			"session_id": sessionID,
			"path":       path,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("path sanitization failed: %w", err)
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(sanitizedPath)
	if err := session.SFTP.MkdirAll(parentDir); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}

	transfer := &Transfer{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Path:      sanitizedPath,
		Type:      "upload",
		Size:      size,
		Offset:    0,
		StartedAt: time.Now().UTC(),
	}

	s.transferMu.Lock()
	s.transfers[transfer.ID] = transfer
	s.transferMu.Unlock()

	// Create file
	file, err := session.SFTP.Create(sanitizedPath)
	if err != nil {
		s.transferMu.Lock()
		delete(s.transfers, transfer.ID)
		s.transferMu.Unlock()
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy with bandwidth monitoring
	buf := make([]byte, DefaultChunkSize)
	var written int64
	startTime := time.Now()

	for {
		select {
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				// Check rate limit
				if err := s.checkRateLimit(sessionID, int64(n), "upload"); err != nil {
					return transfer, err
				}

				_, writeErr := file.Write(buf[:n])
				if writeErr != nil {
					return transfer, fmt.Errorf("%w: %v", ErrWriteFailed, writeErr)
				}

				written += int64(n)
				s.updateBandwidthStats(sessionID, int64(n), 0)

				transfer.mu.Lock()
				transfer.Offset = written
				transfer.mu.Unlock()
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return transfer, fmt.Errorf("read error: %w", err)
			}
		}
	}

	elapsed := time.Since(startTime).Seconds()
	transfer.mu.Lock()
	transfer.Completed = true
	transfer.mu.Unlock()

	s.logAudit(session.UserID, "sftp_upload_complete", map[string]interface{}{
		"session_id":    sessionID,
		"path":          sanitizedPath,
		"size":          written,
		"duration_sec":  elapsed,
		"transfer_id":   transfer.ID,
	})

	s.updateLastUsed(sessionID)
	return transfer, nil
}

// Download downloads a file starting from an offset with resume support
func (s *SFTPService) Download(sessionID, path string, offset int64) (io.ReadCloser, *Transfer, error) {
	session, err := s.getSession(sessionID)
	if err != nil {
		return nil, nil, err
	}

	sanitizedPath, err := s.sanitizer.SanitizePath(session.RootDir, path)
	if err != nil {
		s.logAudit(session.UserID, "sftp_download_denied", map[string]interface{}{
			"session_id": sessionID,
			"path":       path,
			"error":      err.Error(),
		})
		return nil, nil, fmt.Errorf("path sanitization failed: %w", err)
	}

	// Check file info
	info, err := session.SFTP.Stat(sanitizedPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > MaxFileSize {
		return nil, nil, ErrFileTooLarge
	}

	if offset < 0 || offset > info.Size() {
		return nil, nil, ErrInvalidOffset
	}

	transfer := &Transfer{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Path:      sanitizedPath,
		Type:      "download",
		Size:      info.Size(),
		Offset:    offset,
		StartedAt: time.Now().UTC(),
	}

	s.transferMu.Lock()
	s.transfers[transfer.ID] = transfer
	s.transferMu.Unlock()

	file, err := session.SFTP.Open(sanitizedPath)
	if err != nil {
		s.transferMu.Lock()
		delete(s.transfers, transfer.ID)
		s.transferMu.Unlock()
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Seek to offset if needed
	if offset > 0 {
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			file.Close()
			s.transferMu.Lock()
			delete(s.transfers, transfer.ID)
			s.transferMu.Unlock()
			return nil, nil, fmt.Errorf("failed to seek: %w", err)
		}
	}

	s.logAudit(session.UserID, "sftp_download_start", map[string]interface{}{
		"session_id":  sessionID,
		"path":        sanitizedPath,
		"size":        info.Size(),
		"offset":      offset,
		"transfer_id": transfer.ID,
	})

	s.updateLastUsed(sessionID)

	// Wrap the file reader to track bandwidth
	return &bandwidthReader{
		ReadCloser: file,
		service:    s,
		sessionID:  sessionID,
		transfer:   transfer,
	}, transfer, nil
}

// Remove deletes a file or directory
func (s *SFTPService) Remove(sessionID, path string) error {
	session, err := s.getSession(sessionID)
	if err != nil {
		return err
	}

	sanitizedPath, err := s.sanitizer.SanitizePath(session.RootDir, path)
	if err != nil {
		s.logAudit(session.UserID, "sftp_remove_denied", map[string]interface{}{
			"session_id": sessionID,
			"path":       path,
			"error":      err.Error(),
		})
		return fmt.Errorf("path sanitization failed: %w", err)
	}

	// Get info to determine if it's a file or directory
	info, err := session.SFTP.Stat(sanitizedPath)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		err = session.SFTP.RemoveDirectory(sanitizedPath)
	} else {
		err = session.SFTP.Remove(sanitizedPath)
	}

	if err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	s.logAudit(session.UserID, "sftp_remove", map[string]interface{}{
		"session_id": sessionID,
		"path":       sanitizedPath,
		"is_dir":     info.IsDir(),
	})

	s.updateLastUsed(sessionID)
	return nil
}

// Mkdir creates a directory
func (s *SFTPService) Mkdir(sessionID, path string) error {
	session, err := s.getSession(sessionID)
	if err != nil {
		return err
	}

	sanitizedPath, err := s.sanitizer.SanitizePath(session.RootDir, path)
	if err != nil {
		s.logAudit(session.UserID, "sftp_mkdir_denied", map[string]interface{}{
			"session_id": sessionID,
			"path":       path,
			"error":      err.Error(),
		})
		return fmt.Errorf("path sanitization failed: %w", err)
	}

	if err := session.SFTP.MkdirAll(sanitizedPath); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	s.logAudit(session.UserID, "sftp_mkdir", map[string]interface{}{
		"session_id": sessionID,
		"path":       sanitizedPath,
	})

	s.updateLastUsed(sessionID)
	return nil
}

// Move renames or moves a file/directory
func (s *SFTPService) Move(sessionID, oldPath, newPath string) error {
	session, err := s.getSession(sessionID)
	if err != nil {
		return err
	}

	sanitizedOldPath, err := s.sanitizer.SanitizePath(session.RootDir, oldPath)
	if err != nil {
		s.logAudit(session.UserID, "sftp_move_denied", map[string]interface{}{
			"session_id": sessionID,
			"old_path":   oldPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("old path sanitization failed: %w", err)
	}

	sanitizedNewPath, err := s.sanitizer.SanitizePath(session.RootDir, newPath)
	if err != nil {
		s.logAudit(session.UserID, "sftp_move_denied", map[string]interface{}{
			"session_id": sessionID,
			"new_path":   newPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("new path sanitization failed: %w", err)
	}

	if err := session.SFTP.Rename(sanitizedOldPath, sanitizedNewPath); err != nil {
		return fmt.Errorf("failed to move: %w", err)
	}

	s.logAudit(session.UserID, "sftp_move", map[string]interface{}{
		"session_id": sessionID,
		"old_path":   sanitizedOldPath,
		"new_path":   sanitizedNewPath,
	})

	s.updateLastUsed(sessionID)
	return nil
}

// CloseSession cleanly closes an SFTP session
func (s *SFTPService) CloseSession(sessionID string) error {
	s.mu.Lock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.Unlock()
		return ErrSessionNotFound
	}
	delete(s.sessions, sessionID)
	s.mu.Unlock()

	session.mu.Lock()
	session.active = false
	session.mu.Unlock()

	if session.SFTP != nil {
		session.SFTP.Close()
	}

	if session.Conn != nil {
		session.Conn.Close()
	}

	// Clean up bandwidth stats
	s.bandwidthMu.Lock()
	delete(s.bandwidthStats, sessionID)
	s.bandwidthMu.Unlock()

	// Clean up rate limiter
	s.rateMu.Lock()
	delete(s.rateLimiters, sessionID)
	s.rateMu.Unlock()

	// Cancel active transfers
	s.transferMu.Lock()
	for id, t := range s.transfers {
		if t.SessionID == sessionID {
			t.Cancelled = true
			delete(s.transfers, id)
		}
	}
	s.transferMu.Unlock()

	s.logAudit(session.UserID, "sftp_session_closed", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// GetSession returns a session by ID
func (s *SFTPService) GetSession(sessionID string) (*SFTPClient, error) {
	return s.getSession(sessionID)
}

// getSession is the internal method to get a session
func (s *SFTPService) getSession(sessionID string) (*SFTPClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	session.mu.RLock()
	active := session.active
	session.mu.RUnlock()

	if !active {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// updateLastUsed updates the last used timestamp
func (s *SFTPService) updateLastUsed(sessionID string) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if exists {
		session.mu.Lock()
		session.LastUsed = time.Now().UTC()
		session.mu.Unlock()
	}
}

// MonitorBandwidth returns bandwidth stats for a session
func (s *SFTPService) MonitorBandwidth(sessionID string) *BandwidthStats {
	s.bandwidthMu.RLock()
	defer s.bandwidthMu.RUnlock()

	stats, exists := s.bandwidthStats[sessionID]
	if !exists {
		return &BandwidthStats{}
	}

	// Return a copy
	return &BandwidthStats{
		BytesSent:     stats.BytesSent,
		BytesReceived: stats.BytesReceived,
		SentSpeed:     stats.SentSpeed,
		ReceiveSpeed:  stats.ReceiveSpeed,
		LastUpdated:   stats.LastUpdated,
	}
}

// updateBandwidthStats updates bandwidth statistics
func (s *SFTPService) updateBandwidthStats(sessionID string, bytesReceived, bytesSent int64) {
	s.bandwidthMu.Lock()
	defer s.bandwidthMu.Unlock()

	stats, exists := s.bandwidthStats[sessionID]
	if !exists {
		stats = &BandwidthStats{LastUpdated: time.Now().UTC()}
		s.bandwidthStats[sessionID] = stats
	}

	now := time.Now().UTC()
	elapsed := now.Sub(stats.LastUpdated).Seconds()

	if elapsed > 0 {
		atomic.AddInt64(&stats.BytesReceived, bytesReceived)
		atomic.AddInt64(&stats.BytesSent, bytesSent)

		stats.ReceiveSpeed = float64(bytesReceived) / elapsed
		stats.SentSpeed = float64(bytesSent) / elapsed
		stats.LastUpdated = now
	}
}

// checkRateLimit enforces per-session bandwidth limits
func (s *SFTPService) checkRateLimit(sessionID string, bytes int64, direction string) error {
	s.rateMu.RLock()
	limiter, exists := s.rateLimiters[sessionID]
	s.rateMu.RUnlock()

	if !exists {
		return nil
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now().UTC()
	if now.Sub(limiter.lastWindow) > BandwidthWindow {
		// Reset window
		limiter.bytesSent = 0
		limiter.bytesReceived = 0
		limiter.lastWindow = now
	}

	if direction == "upload" {
		limiter.bytesSent += bytes
		if limiter.bytesSent > MaxBandwidth {
			return ErrBandwidthExceeded
		}
	} else {
		limiter.bytesReceived += bytes
		if limiter.bytesReceived > MaxBandwidth {
			return ErrBandwidthExceeded
		}
	}

	return nil
}

// logAudit logs an audit event
func (s *SFTPService) logAudit(userID uuid.UUID, eventType string, details map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}

	s.auditLogger.Log(
		eventType,
		&userID, nil, "",
		details,
	)
}

// bandwidthReader wraps an io.ReadCloser to track bandwidth
type bandwidthReader struct {
	io.ReadCloser
	service   *SFTPService
	sessionID string
	transfer  *Transfer
}

func (r *bandwidthReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.service.updateBandwidthStats(r.sessionID, int64(n), 0)
		r.transfer.mu.Lock()
		r.transfer.Offset += int64(n)
		r.transfer.mu.Unlock()
	}
	return n, err
}

func (r *bandwidthReader) Close() error {
	r.transfer.mu.Lock()
	r.transfer.Completed = true
	r.transfer.mu.Unlock()
	return r.ReadCloser.Close()
}

// ListActiveSessions returns all active SFTP sessions
func (s *SFTPService) ListActiveSessions() []*SFTPClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*SFTPClient, 0, len(s.sessions))
	for _, session := range s.sessions {
		session.mu.RLock()
		active := session.active
		session.mu.RUnlock()
		if active {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetTransfer returns an active transfer by ID
func (s *SFTPService) GetTransfer(transferID string) (*Transfer, error) {
	s.transferMu.RLock()
	defer s.transferMu.RUnlock()

	transfer, exists := s.transfers[transferID]
	if !exists {
		return nil, ErrTransferNotFound
	}
	return transfer, nil
}

// CancelTransfer cancels an active transfer
func (s *SFTPService) CancelTransfer(transferID string) error {
	s.transferMu.Lock()
	defer s.transferMu.Unlock()

	transfer, exists := s.transfers[transferID]
	if !exists {
		return ErrTransferNotFound
	}

	transfer.Cancelled = true
	delete(s.transfers, transferID)

	return nil
}
