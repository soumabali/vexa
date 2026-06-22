package ssh

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers SSH proxy routes
type Handler struct {
	sessionManager *SessionManager
}

// NewHandler creates a new SSH handler
func NewHandler(sm *SessionManager) *Handler {
	return &Handler{
		sessionManager: sm,
	}
}

// RegisterRoutes registers all SSH-related routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	ssh := r.Group("/ssh")
	{
		ssh.POST("/sessions", h.sessionManager.CreateSession)
		ssh.GET("/sessions", h.sessionManager.ListSessions)
		ssh.GET("/sessions/:id", h.sessionManager.GetSession)
		ssh.DELETE("/sessions/:id", h.sessionManager.CloseSession)
		ssh.GET("/connect", h.sessionManager.HandleWebSocket)
	}
}

// HealthCheck returns SSH proxy health status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}
