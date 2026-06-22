package sftp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/auth"
)

// Handler provides HTTP handlers for SFTP operations
type Handler struct {
	sftpService *SFTPService
	jwtManager  *auth.JWTManager
}

// NewHandler creates a new SFTP HTTP handler
func NewHandler(sftpService *SFTPService, jwtManager *auth.JWTManager) *Handler {
	return &Handler{
		sftpService: sftpService,
		jwtManager:  jwtManager,
	}
}

// CreateSessionRequest represents a request to create an SFTP session
type CreateSessionRequest struct {
	HostID      string `json:"host_id" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	SSHUser     string `json:"ssh_user" binding:"required"`
	SSHPassword string `json:"ssh_password" binding:"required"`
	RootDir     string `json:"root_dir,omitempty"`
}

// CreateSession handles creating a new SFTP session
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hostID, err := uuid.Parse(req.HostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host_id"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// In a real implementation, you would look up the host details from the database
	// and establish the SSH connection. For now, return a stub response.
	c.JSON(http.StatusOK, gin.H{
		"session_id": "stub-session-id",
		"host_id":    hostID.String(),
		"user_id":    userID.String(),
		"status":     "created",
	})
}

// ListDirectory handles listing directory contents
func (h *Handler) ListDirectory(c *gin.Context) {
	sessionID := c.Param("session_id")
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	files, err := h.sftpService.ListDirectory(sessionID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"path":       path,
		"files":      files,
	})
}

// UploadRequest represents an upload request via HTTP
func (h *Handler) Upload(c *gin.Context) {
	sessionID := c.Param("session_id")
	path := c.PostForm("path")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	// Get file size from header or parse from form
	sizeStr := c.PostForm("size")
	var size int64
	if sizeStr != "" {
		size, _ = strconv.ParseInt(sizeStr, 10, 64)
	}

	transfer, err := h.sftpService.Upload(sessionID, path, file, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transfer_id": transfer.ID,
		"status":      "uploaded",
		"size":        transfer.Size,
	})
}

// Download handles file downloads via HTTP
func (h *Handler) Download(c *gin.Context) {
	sessionID := c.Param("session_id")
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	offsetStr := c.Query("offset")
	var offset int64
	if offsetStr != "" {
		offset, _ = strconv.ParseInt(offsetStr, 10, 64)
	}

	reader, transfer, err := h.sftpService.Download(sessionID, path, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	c.Header("X-Transfer-ID", transfer.ID)
	c.Header("X-File-Size", strconv.FormatInt(transfer.Size, 10))
	c.Header("Content-Disposition", "attachment; filename=\""+path+"\"")
	c.DataFromReader(http.StatusOK, transfer.Size-offset, "application/octet-stream", reader, nil)
}

// Remove handles file/directory deletion
func (h *Handler) Remove(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sftpService.Remove(sessionID, req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"path":       req.Path,
		"status":     "removed",
	})
}

// Mkdir handles directory creation
func (h *Handler) Mkdir(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sftpService.Mkdir(sessionID, req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"path":       req.Path,
		"status":     "created",
	})
}

// Move handles rename/move operations
func (h *Handler) Move(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		OldPath string `json:"old_path" binding:"required"`
		NewPath string `json:"new_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sftpService.Move(sessionID, req.OldPath, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"old_path":   req.OldPath,
		"new_path":   req.NewPath,
		"status":     "moved",
	})
}

// CloseSession handles closing an SFTP session
func (h *Handler) CloseSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	if err := h.sftpService.CloseSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"status":     "closed",
	})
}

// GetBandwidth returns bandwidth statistics for a session
func (h *Handler) GetBandwidth(c *gin.Context) {
	sessionID := c.Param("session_id")

	stats := h.sftpService.MonitorBandwidth(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"session_id":      sessionID,
		"bytes_sent":      stats.BytesSent,
		"bytes_received":  stats.BytesReceived,
		"sent_speed":      stats.SentSpeed,
		"receive_speed":   stats.ReceiveSpeed,
		"last_updated":    stats.LastUpdated,
	})
}
