package discovery

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler contains HTTP handlers for discovery endpoints
type Handler struct {
	scanner *Scanner
}

// NewHandler creates discovery handlers
func NewHandler(scanner *Scanner) *Handler {
	return &Handler{scanner: scanner}
}

// RegisterRoutes registers discovery routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	discovery := router.Group("/discovery")
	discovery.Use(authMiddleware)
	{
		discovery.POST("/scan", h.StartScan)
		discovery.GET("/scan", h.ListScans)
		discovery.GET("/scan/:id", h.GetScan)
		discovery.GET("/scan/:id/status", h.GetScanStatus)
		discovery.GET("/scan/:id/results", h.GetScanResults)
		discovery.POST("/scan/:id/cancel", h.CancelScan)
		discovery.POST("/scan/:id/import", h.ImportResults)
	}
}

// StartScanRequest represents scan start request
type StartScanRequest struct {
	Network string   `json:"network" binding:"required"` // CIDR notation
	Ports   []int    `json:"ports"`                     // Empty = default ports
	Options struct {
		TimeoutMs       int  `json:"timeout_ms"`
		RateLimitMs     int  `json:"rate_limit_ms"`
		Concurrency     int  `json:"concurrency"`
		ICMPFirst       bool `json:"icmp_first"`
		OSScan          bool `json:"os_scan"`
		ResolveHostname bool `json:"resolve_hostname"`
	} `json:"options"`
}

// StartScanResponse represents scan start response
type StartScanResponse struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	Network   string    `json:"network"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// StartScan starts a new network scan
func (h *Handler) StartScan(c *gin.Context) {
	var req StartScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate network
	if !strings.Contains(req.Network, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "network must be CIDR notation (e.g., 192.168.1.0/24)"})
		return
	}

	// Build options
	opts := DefaultScanOptions()
	if req.Options.TimeoutMs > 0 {
		opts.TimeoutMs = req.Options.TimeoutMs
	}
	if req.Options.RateLimitMs > 0 {
		opts.RateLimitMs = req.Options.RateLimitMs
	}
	if req.Options.Concurrency > 0 {
		opts.Concurrency = req.Options.Concurrency
		// Limit max concurrency
		if opts.Concurrency > 200 {
			opts.Concurrency = 200
		}
	}
	opts.ICMPFirst = req.Options.ICMPFirst
	opts.OSScan = req.Options.OSScan
	opts.ResolveHostname = req.Options.ResolveHostname

	// Create job
	job, err := h.scanner.CreateJob(req.Network, req.Ports, opts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start scan in background
	ctx := c.Request.Context()
	if err := h.scanner.StartJob(ctx, job.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, StartScanResponse{
		ID:        job.ID,
		Status:    job.Status,
		Network:   job.Network,
		Message:   "Scan started successfully",
		CreatedAt: job.CreatedAt,
	})
}

// GetScanStatus returns scan progress
func (h *Handler) GetScanStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan ID"})
		return
	}

	job, exists := h.scanner.GetJob(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            job.ID,
		"status":        job.Status,
		"progress":      job.Progress,
		"total_hosts":   job.TotalHosts,
		"scanned_hosts": job.ScannedHosts,
		"found_hosts":   len(job.Results),
		"error":         job.Error,
		"started_at":    job.StartedAt,
		"completed_at":  job.CompletedAt,
	})
}

// GetScanResults returns scan results
func (h *Handler) GetScanResults(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan ID"})
		return
	}

	job, exists := h.scanner.GetJob(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	// Filter by status
	status := c.Query("status")

	var results []ScanResult
	for _, r := range job.Results {
		if status != "" && r.Status != status {
			continue
		}
		results = append(results, r)
	}

	// Paginate
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(results) {
		start = len(results)
	}
	if end > len(results) {
		end = len(results)
	}

	var pageResults []ScanResult
	if start < len(results) {
		pageResults = results[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"scan":        job.ID,
		"status":      job.Status,
		"total":       len(results),
		"page":        page,
		"per_page":    perPage,
		"total_pages": (len(results) + perPage - 1) / perPage,
		"results":     pageResults,
	})
}

// GetScan returns full scan details
func (h *Handler) GetScan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan ID"})
		return
	}

	job, exists := h.scanner.GetJob(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListScans returns all scan jobs
func (h *Handler) ListScans(c *gin.Context) {
	jobs := h.scanner.ListJobs()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(jobs) {
		start = len(jobs)
	}
	if end > len(jobs) {
		end = len(jobs)
	}

	var pageJobs []*ScanJob
	if start < len(jobs) {
		pageJobs = jobs[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"total":       len(jobs),
		"page":        page,
		"per_page":    perPage,
		"total_pages": (len(jobs) + perPage - 1) / perPage,
		"jobs":        pageJobs,
	})
}

// CancelScan canels an active scan
func (h *Handler) CancelScan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan ID"})
		return
	}

	if err := h.scanner.CancelJob(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"status":  "cancelled",
		"message": "Scan cancelled",
	})
}

// ImportResultsRequest represents result import request
type ImportResultsRequest struct {
	ImportAll   bool     `json:"import_all"`
	SelectedIPs []string `json:"selected_ips,omitempty"`
}

// ImportResults imports scan results to hosts
func (h *Handler) ImportResults(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan ID"})
		return
	}

	var req ImportResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	imported, err := h.scanner.ImportResults(id, userID.(string), req.ImportAll, req.SelectedIPs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported_count": len(imported),
		"imported_ips":   imported,
		"message":        "Hosts imported successfully",
	})
}
