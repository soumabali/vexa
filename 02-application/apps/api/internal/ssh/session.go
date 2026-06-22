package ssh

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"github.com/gorilla/websocket"
	"github.com/soumabali/vexa/internal/db"
	"github.com/soumabali/vexa/internal/models"
)

// SessionManager handles SSH session lifecycle
type SessionManager struct {
	proxy *ProxyServer
	db    *db.DB
}

// NewSessionManager creates a new session manager
func NewSessionManager(proxy *ProxyServer, database *db.DB) *SessionManager {
	return &SessionManager{
		proxy: proxy,
		db:    database,
	}
}

// CreateSessionRequest represents a request to create SSH session
type CreateSessionRequest struct {
	HostID      string `json:"host_id" binding:"required,uuid"`
	TerminalType string `json:"terminal_type,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// CreateSessionResponse represents the created session
type CreateSessionResponse struct {
	SessionID   string    `json:"session_id"`
	WSURL       string    `json:"ws_url"`
	HostID      string    `json:"host_id"`
	HostAddress string    `json:"host_address"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ID          string    `json:"id"`
	HostID      string    `json:"host_id"`
	HostAddress string    `json:"host_address"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
	Duration    int64     `json:"duration_seconds"`
}

// CreateSession creates a new SSH session
func (sm *SessionManager) CreateSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get host details — TODO: use hosts.Repository
	// host, err := sm.db.GetHost(c.Request.Context(), req.HostID)
	// if err != nil { ... }
	_ = req
	_ = userID

	// Build SSH config
	// TODO: integrate with hosts repository and vault for real SSH config
	// For now, create a placeholder session
	sshConfig, err := sm.buildSSHConfig(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build SSH config"})
		return
	}

	// Create session
	session, err := sm.proxy.CreateSession(userID.(string), req.HostID, sshConfig)
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	// Connect to host — TODO: get real host address from hosts repository
	hostAddress := fmt.Sprintf("%s:22", req.HostID) // placeholder
	if err := session.Connect(c.Request.Context(), hostAddress); err != nil {
		sm.proxy.RemoveSession(session.ID)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Audit log
	// audit.LogSessionCreated(c.Request.Context(), userID.(string), session.ID, req.HostID)

	wsURL := fmt.Sprintf("wss://%s/api/ssh/connect?session_id=%s", c.Request.Host, session.ID)

	c.JSON(http.StatusCreated, CreateSessionResponse{
		SessionID:   session.ID,
		WSURL:       wsURL,
		HostID:      req.HostID,
		HostAddress: hostAddress,
		CreatedAt:   session.CreatedAt,
		ExpiresAt:   session.CreatedAt.Add(24 * time.Hour),
	})
}

// ListSessions returns all active sessions for user
func (sm *SessionManager) ListSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessions := sm.proxy.ListUserSessions(userID.(string))

	var infos []SessionInfo
	now := time.Now()
	for _, s := range sessions {
		status := "active"
		if s.Client == nil {
			status = "connecting"
		}

		infos = append(infos, SessionInfo{
			ID:          s.ID,
			HostID:      s.HostID,
			HostAddress: s.HostAddress,
			Status:      status,
			CreatedAt:   s.CreatedAt,
			LastActive:  s.LastActive,
			Duration:    int64(now.Sub(s.CreatedAt).Seconds()),
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": infos})
}

// GetSession returns details for a specific session
func (sm *SessionManager) GetSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")
	session, ok := sm.proxy.GetSession(sessionID)
	if !ok || session.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	status := "active"
	if session.Client == nil {
		status = "connecting"
	}

	c.JSON(http.StatusOK, SessionInfo{
		ID:          session.ID,
		HostID:      session.HostID,
		HostAddress: session.HostAddress,
		Status:      status,
		CreatedAt:   session.CreatedAt,
		LastActive:  session.LastActive,
		Duration:    int64(time.Since(session.CreatedAt).Seconds()),
	})
}

// CloseSession terminates a session
func (sm *SessionManager) CloseSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Param("id")
	session, ok := sm.proxy.GetSession(sessionID)
	if !ok || session.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	sm.proxy.RemoveSession(sessionID)

	// audit.LogSessionClosed(c.Request.Context(), userID.(string), sessionID)

	c.JSON(http.StatusOK, gin.H{"message": "session closed"})
}

// HandleWebSocket upgrades HTTP to WebSocket for SSH
func (sm *SessionManager) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	session, ok := sm.proxy.GetSession(sessionID)
	if !ok || session.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	ws, err := sm.proxy.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "websocket upgrade failed"})
		return
	}

	// Handle auto-reconnect
	go sm.handleReconnect(session, ws)

	session.HandleWebSocket(ws)
}

// handleReconnect attempts to reconnect on disconnect
func (sm *SessionManager) handleReconnect(session *Session, ws *websocket.Conn) {
	maxRetries := 3
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		select {
		case <-session.ctx.Done():
			return
		case <-time.After(backoff):
		}

		// Check if session is still active
		if _, ok := sm.proxy.GetSession(session.ID); !ok {
			return
		}

		// Attempt reconnection
		if session.Client != nil {
			// Session is still connected
			return
		}

		// Try to reconnect
		if err := session.Connect(context.Background(), session.HostAddress); err != nil {
			backoff *= 2
			continue
		}

		// Reconnected successfully
		session.HandleWebSocket(ws)
		return
	}
}

// buildSSHConfig builds SSH client config from credentials
func (sm *SessionManager) buildSSHConfig(cred *models.Credential) (*ssh.ClientConfig, error) {
	_ = cred // TODO: integrate with vault credential service
	config := &ssh.ClientConfig{
		Timeout:         10 * time.Second,
		HostKeyCallback: HostKeyCallback(),
	}
	return config, nil
}

// CleanupInactiveSessions removes old sessions periodically
func (sm *SessionManager) CleanupInactiveSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.proxy.CleanupInactive(1 * time.Hour)
	}
}
