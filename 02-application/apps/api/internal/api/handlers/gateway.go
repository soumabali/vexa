package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/gateway"
)

type GatewayHandler struct {
	sshGateway *gateway.SSHGateway
	rdpGateway *gateway.RDPGateway
	vncGateway *gateway.VNCGateway
	auditLogger *audit.Logger
}

func NewGatewayHandler(ssh *gateway.SSHGateway, rdp *gateway.RDPGateway, vnc *gateway.VNCGateway, audit *audit.Logger) *GatewayHandler {
	return &GatewayHandler{sshGateway: ssh, rdpGateway: rdp, vncGateway: vnc, auditLogger: audit}
}

func (h *GatewayHandler) SSHConnect(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" binding:"required,min=1,max=65535"`
		User     string `json:"user" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	conn, err := h.sshGateway.GetConnection(c.Request.Context(), req.Host, req.Port, req.User, req.Password)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	h.auditLogger.Log("host.connect", &uid, nil, c.ClientIP(), map[string]interface{}{
		"protocol": "ssh",
		"host":     req.Host,
		"port":     req.Port,
	})

	c.JSON(http.StatusOK, gin.H{
		"connection_id": conn.ID,
		"status":        "connected",
	})
}

func (h *GatewayHandler) RDPConnect(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" default:"3389"`
		User     string `json:"user" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "RDP connect endpoint - WebSocket upgrade required", "protocol": "rdp"})
}

func (h *GatewayHandler) VNCConnect(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" default:"5900"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "VNC connect endpoint - WebSocket upgrade required", "protocol": "vnc"})
}
