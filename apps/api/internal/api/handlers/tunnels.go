package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/models"
	"github.com/soumabali/vexa/internal/security"
	"github.com/soumabali/vexa/internal/wireguard"
)

type TunnelHandler struct {
	svc         *wireguard.WireGuardService
	rateLimiter *security.RateLimiter
}

func NewTunnelHandler(svc *wireguard.WireGuardService) *TunnelHandler {
	return &TunnelHandler{
		svc:         svc,
		rateLimiter: security.NewRateLimiter(10, 20),
	}
}

func (h *TunnelHandler) Create(c *gin.Context) {
	if !h.rateLimiter.Allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	var req models.CreateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	hostID, err := uuid.Parse(req.HostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host_id"})
		return
	}

	tunnel, err := h.svc.CreateTunnel(c.Request.Context(), uid, hostID, &req)
	if err != nil {
		if errors.Is(err, wireguard.ErrTunnelExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, wireguard.ErrMaxTunnelsReached) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tunnel)
}

func (h *TunnelHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	hostID := c.Query("host_id")
	enabledStr := c.Query("enabled")

	var enabled *bool
	if enabledStr == "true" {
		t := true
		enabled = &t
	} else if enabledStr == "false" {
		t := false
		enabled = &t
	}

	tunnels, err := h.svc.ListTunnels(c.Request.Context(), uid, hostID, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tunnels": tunnels})
}

func (h *TunnelHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	tunnel, err := h.svc.GetTunnel(c.Request.Context(), id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tunnel not found"})
		return
	}

	c.JSON(http.StatusOK, tunnel)
}

func (h *TunnelHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	var req models.UpdateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	tunnel, err := h.svc.UpdateTunnel(c.Request.Context(), id, uid, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tunnel)
}

func (h *TunnelHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	if err := h.svc.DeleteTunnel(c.Request.Context(), id, uid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TunnelHandler) Rotate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	tunnel, err := h.svc.RotateKeys(c.Request.Context(), id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tunnel)
}

func (h *TunnelHandler) Config(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	export, err := h.svc.GetConfig(c.Request.Context(), id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	conf := wireguard.GenerateConfigFile(&wireguard.ConfigFileParams{
		ClientPrivateKey:    export.ClientPrivateKey,
		Address:             export.Address,
		DNS:                 export.DNS,
		ServerPublicKey:     export.ServerPublicKey,
		PresharedKey:        export.PresharedKey,
		AllowedIPs:          export.AllowedIPs,
		Endpoint:            export.Endpoint,
		Port:                export.Port,
		PersistentKeepalive: export.PersistentKeepalive,
		MTU:                 export.MTU,
	})

	acceptHeader := c.GetHeader("Accept")
	if acceptHeader == "text/plain" || c.Query("format") == "conf" {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(conf))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config":       conf,
		"export":       export,
		"content_type": "application/x-wireguard-profile",
	})
}

func (h *TunnelHandler) Enable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	if err := h.svc.EnableTunnel(c.Request.Context(), id, uid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tunnel enabled"})
}

func (h *TunnelHandler) Disable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	if err := h.svc.DisableTunnel(c.Request.Context(), id, uid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tunnel disabled"})
}

func (h *TunnelHandler) Stats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tunnel ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	stats, err := h.svc.GetStats(c.Request.Context(), id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
