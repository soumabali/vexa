package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/hosts"
	"github.com/soumabali/vexa/internal/models"
)

type HostHandler struct {
	repo        *hosts.Repository
	auditLogger *audit.Logger
}

func NewHostHandler(repo *hosts.Repository, auditLogger *audit.Logger) *HostHandler {
	return &HostHandler{repo: repo, auditLogger: auditLogger}
}

// NewHostHandlerWithLogger creates a host handler using a real audit logger (test helper).
func NewHostHandlerWithLogger(repo *hosts.Repository, auditLogger *audit.Logger) *HostHandler {
	return &HostHandler{repo: repo, auditLogger: auditLogger}
}

func (h *HostHandler) Create(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "host repository not configured"})
		return
	}

	var req models.CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var credID *uuid.UUID
	if req.CredentialsID != nil {
		id, err := uuid.Parse(*req.CredentialsID)
		if err == nil {
			credID = &id
		}
	}

	host := &models.Host{
		ID:            uuid.New(),
		OwnerID:       ownerID,
		Name:          req.Name,
		Address:       req.Address,
		Protocol:      req.Protocol,
		Port:          req.Port,
		CredentialsID: credID,
		Tags:          req.Tags,
		GroupPath:     req.GroupPath,
		Description:   req.Description,
		IsActive:      true,
	}

	if err := h.repo.Create(c.Request.Context(), host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create host"})
		return
	}

	h.auditLogger.Log("host.create", &ownerID, nil, c.ClientIP(), map[string]interface{}{
		"host_id":   host.ID.String(),
		"host_name": host.Name,
		"protocol":  host.Protocol,
	})

	c.JSON(http.StatusCreated, host)
}

func (h *HostHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	userID, _ := c.Get("user_id")
	ownerID := userID.(uuid.UUID)

	host, err := h.repo.GetByID(c.Request.Context(), id, ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	c.JSON(http.StatusOK, host)
}

func (h *HostHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	ownerID := userID.(uuid.UUID)

	tag := c.Query("tag")
	group := c.Query("group")
	limit := 100
	offset := 0

	hosts, err := h.repo.List(c.Request.Context(), ownerID, tag, group, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list hosts"})
		return
	}

	if hosts == nil {
		hosts = []*models.Host{}
	}

	c.JSON(http.StatusOK, gin.H{"hosts": hosts, "total": len(hosts), "page": 1, "pageSize": limit, "totalPages": 1})
}

func (h *HostHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	var req models.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.Port > 0 {
		updates["port"] = req.Port
	}
	if req.CredentialsID != nil {
		cid, _ := uuid.Parse(*req.CredentialsID)
		updates["credentials_id"] = cid
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.GroupPath != "" {
		updates["group_path"] = req.GroupPath
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	userID, _ := c.Get("user_id")
	ownerID := userID.(uuid.UUID)

	updateHost := &models.Host{ID: id}
	if req.Name != "" {
		updateHost.Name = req.Name
	}
	if req.Address != "" {
		updateHost.Address = req.Address
	}
	if req.Protocol != "" {
		updateHost.Protocol = req.Protocol
	}
	if req.Port > 0 {
		updateHost.Port = req.Port
	}
	if req.GroupPath != "" {
		updateHost.GroupPath = req.GroupPath
	}
	if req.Description != "" {
		updateHost.Description = req.Description
	}
	if req.IsActive != nil {
		if *req.IsActive {
			updateHost.Status = "active"
		} else {
			updateHost.Status = "inactive"
		}
	}

	if err := h.repo.Update(c.Request.Context(), updateHost, ownerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update host"})
		return
	}

	h.auditLogger.Log("host.update", &ownerID, nil, c.ClientIP(), map[string]interface{}{
		"host_id": id.String(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "host updated"})
}

func (h *HostHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	userID, _ := c.Get("user_id")
	ownerID := userID.(uuid.UUID)

	if err := h.repo.Delete(c.Request.Context(), id, ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	h.auditLogger.Log("host.delete", &ownerID, nil, c.ClientIP(), map[string]interface{}{
		"host_id": id.String(),
	})

	c.JSON(http.StatusNoContent, nil)
}

func (h *HostHandler) HealthCheck(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID := userID.(uuid.UUID)

	_, err = h.repo.GetByID(c.Request.Context(), id, ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	result := &models.HealthCheckResult{HostID: id, Reachable: false, LatencyMs: 0, CheckedAt: time.Now().UTC()}
	c.JSON(http.StatusOK, result)
}

// GetStats returns aggregate stats for a single host:
// session count, last connection time, tunnel counts.
// Returns 404 if the host does not exist for the calling user.
func (h *HostHandler) GetStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ownerID := userID.(uuid.UUID)

	// Verify ownership / existence first so we return 404 instead of empty stats.
	if _, err := h.repo.GetByID(c.Request.Context(), id, ownerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	stats, err := h.repo.GetStats(c.Request.Context(), id, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch host stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
