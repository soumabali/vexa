package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/models"
)

type AuditHandler struct {
	logger *audit.Logger
}

func NewAuditHandler(logger *audit.Logger) *AuditHandler {
	return &AuditHandler{logger: logger}
}

func (h *AuditHandler) Query(c *gin.Context) {
	if h.logger == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audit logger not configured"})
		return
	}
	var query models.AuditQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.StartTime.IsZero() {
		query.StartTime = time.Now().UTC().Add(-7 * 24 * time.Hour)
	}
	if query.EndTime.IsZero() {
		query.EndTime = time.Now().UTC()
	}
	if query.Limit <= 0 || query.Limit > 1000 {
		query.Limit = 100
	}

	entries, err := h.logger.Query(query.StartTime, query.EndTime, query.Limit, query.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries, "count": len(entries)})
}

func (h *AuditHandler) Export(c *gin.Context) {
	if h.logger == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audit logger not configured"})
		return
	}
	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	start := time.Now().UTC().Add(-30 * 24 * time.Hour)
	end := time.Now().UTC()

	data, err := h.logger.Export(start, end, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contentType := "application/json"
	filename := "audit-export.json"
	if format == "csv" {
		contentType = "text/csv"
		filename = "audit-export.csv"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *AuditHandler) LogEvent(c *gin.Context) {
	if h.logger == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audit logger not configured"})
		return
	}
	var req struct {
		EventType string                 `json:"event_type" binding:"required"`
		Details   map[string]interface{} `json:"details"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	if err := h.logger.Log(req.EventType, &uid, nil, c.ClientIP(), req.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "event logged"})
}
