package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/gateway"
	"github.com/soumabali/vexa/internal/models"
	"github.com/soumabali/vexa/internal/terminal"
	"github.com/soumabali/vexa/internal/vault"
)

// hostRepository is the subset of hosts.Repository methods we need.
type hostRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*models.Host, error)
}

type wsRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*wsTokenBucket
}

type wsTokenBucket struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
}

func newWSTokenBucket(capacity float64, refillRate float64) *wsTokenBucket {
	return &wsTokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *wsTokenBucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func newWSRateLimiter() *wsRateLimiter {
	rl := &wsRateLimiter{
		visitors: make(map[string]*wsTokenBucket),
	}
	go rl.cleanup()
	return rl
}

func (rl *wsRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, exists := rl.visitors[ip]
	if !exists {
		bucket = newWSTokenBucket(5, 5.0/60.0)
		rl.visitors[ip] = bucket
	}
	return bucket.allow()
}

func (rl *wsRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, bucket := range rl.visitors {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

type TerminalHandler struct {
	sessionManager *terminal.SessionManager
	upgrader       websocket.Upgrader
	jwtManager     *auth.JWTManager
	sessionStore   *auth.SessionStore
	allowedOrigins []string
	wsRateLim      *wsRateLimiter
	hostRepo       hostRepository
	credService    *vault.CredentialService
	sshGateway     *gateway.SSHGateway
	auditLogger    *audit.Logger
}

// CredentialData holds resolved SSH credentials.
type CredentialData struct {
	Username   string
	Password   string
	PrivateKey string
	Passphrase string
}

func NewTerminalHandler(
	sessionManager *terminal.SessionManager,
	jwtManager *auth.JWTManager,
	sessionStore *auth.SessionStore,
	allowedOrigins []string,
	hostRepo hostRepository,
	credService *vault.CredentialService,
	sshGateway *gateway.SSHGateway,
	auditLogger *audit.Logger,
) *TerminalHandler {
	if sessionManager == nil {
		panic("sessionManager is required")
	}
	if jwtManager == nil {
		panic("jwtManager is required")
	}
	return &TerminalHandler{
		sessionManager: sessionManager,
		jwtManager:     jwtManager,
		sessionStore:   sessionStore,
		allowedOrigins: allowedOrigins,
		wsRateLim:      newWSRateLimiter(),
		hostRepo:       hostRepo,
		credService:    credService,
		sshGateway:     sshGateway,
		auditLogger:    auditLogger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				for _, allowed := range allowedOrigins {
					if strings.EqualFold(origin, allowed) {
						return true
					}
				}
				log.Printf("WebSocket connection rejected from origin: %s", origin)
				return false
			},
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
		},
	}
}

// authenticateWebSocket extracts and validates the JWT token from WebSocket query parameter or Authorization header
func (h *TerminalHandler) authenticateWebSocket(r *http.Request) (*auth.Claims, error) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = parts[1]
			}
		}
	}

	if tokenString == "" {
		return nil, errors.New("authentication token required")
	}

	claims, err := h.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	ip := c.ClientIP()
	if !h.wsRateLim.allow(ip) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many WebSocket connection attempts"})
		return
	}

	claims, err := h.authenticateWebSocket(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if h.sessionStore != nil {
		revoked, _ := h.sessionStore.IsAccessTokenRevoked(c.Request.Context(), claims.ID)
		if revoked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			return
		}
	}

	userID := claims.UserID

	hostIDStr := c.Query("host_id")
	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	protocol := c.Query("protocol")
	if protocol == "" {
		protocol = "ssh"
	}

	// Optional host/credential resolution; if hostRepo is nil we skip and let the session be a placeholder.
	var host *models.Host
	var credData *CredentialData
	if h.hostRepo != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		host, err = h.hostRepo.GetByID(ctx, hostID, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "host not found"})
			return
		}

		if host.CredentialsID != nil && h.credService != nil {
			decrypted, err := h.credService.DecryptData(ctx, userID, *host.CredentialsID)
			if err == nil && decrypted != nil {
				credData = &CredentialData{
					Username:   decrypted.Username,
					Password:   decrypted.Password,
					PrivateKey: decrypted.PrivateKey,
					Passphrase: decrypted.Passphrase,
				}
			}
		}
	}

	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
		return
	}

	session, err := h.sessionManager.CreateSession(userID, hostID, protocol, ws)
	if err != nil {
		ws.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("terminal.connect", &userID, &hostID, c.ClientIP(), map[string]interface{}{
			"protocol": protocol,
			"session":  session.ID,
		})
	}

	// Connect real SSH if gateway and credentials available
	if h.sshGateway != nil && credData != nil && credData.Username != "" {
		go h.connectSSH(session, host.Address, host.Port, credData)
	} else {
		// No live backend: put session into "local" placeholder mode
		msg := terminal.Message{Type: "data", Data: json.RawMessage(`"Welcome to SSH Manager terminal. No live SSH backend configured for this session."`)}
		data, _ := json.Marshal(msg)
		if err := session.WebSocket.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("failed to write welcome message: %v", err)
		}
	}
}

func (h *TerminalHandler) connectSSH(session *terminal.Session, address string, port int, creds *CredentialData) {
	if port <= 0 {
		port = 22
	}
	var authMethods []ssh.AuthMethod
	if creds.PrivateKey != "" {
		passphrase := []byte(creds.Passphrase)
		var signer ssh.Signer
		var err error
		if creds.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(creds.PrivateKey), passphrase)
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(creds.PrivateKey))
		}
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}
	if creds.Password != "" {
		authMethods = append(authMethods, ssh.Password(creds.Password))
	}

	if err := session.ConnectSSHWithAuth(address, creds.Username, authMethods, port); err != nil {
		msg := terminal.Message{Type: "error", Data: json.RawMessage(`"` + err.Error() + `"`)}
		data, _ := json.Marshal(msg)
		if err := session.WebSocket.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("failed to write ssh error: %v", err)
		}
	}
}

func (h *TerminalHandler) ListSessions(c *gin.Context) {
	sessions := h.sessionManager.ListSessions()
	result := make([]map[string]interface{}, 0)
	for _, s := range sessions {
		stats := s.GetStats()
		result = append(result, map[string]interface{}{
			"id":         stats.ID.String(),
			"user_id":    stats.UserID.String(),
			"host_id":    stats.HostID.String(),
			"protocol":   stats.Protocol,
			"started_at": stats.StartedAt,
			"bytes_in":   stats.BytesIn,
			"bytes_out":  stats.BytesOut,
			"status":     stats.Status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sessions": result})
}

func (h *TerminalHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")
	h.sessionManager.RemoveSession(sessionID)
	c.JSON(http.StatusOK, gin.H{"message": "session closed"})
}
