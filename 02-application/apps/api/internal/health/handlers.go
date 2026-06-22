package health

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HealthService manages health check HTTP handlers
type HealthService struct {
	DB     *sql.DB
	Engine *ProbeEngine
}

// NewHealthService creates service
func NewHealthService(db *sql.DB, engine *ProbeEngine) *HealthService {
	return &HealthService{DB: db, Engine: engine}
}

// GetHostHealthHandler GET /api/hosts/:id/health
func (s *HealthService) GetHostHealthHandler(c *gin.Context) {
	hostID := c.Param("id")

	// Get host from database
	var host struct {
		ID      string `db:"id"`
		Name    string `db:"name"`
		Address string `db:"address"`
		Port    int    `db:"port"`
	}

	err := s.DB.QueryRow(
		"SELECT id, name, address, COALESCE(port, 22) FROM hosts WHERE id = $1",
		hostID,
	).Scan(&host.ID, &host.Name, &host.Address, &host.Port)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Check if engine has this host registered
	health, ok := s.Engine.GetHealth(hostID)
	if ok {
		c.JSON(http.StatusOK, health)
		return
	}

	// On-demand check if not registered
	config := DefaultProbeConfig()
	config.Port = host.Port

	checker := NewTCPHealthChecker().WithTimeout(10 * time.Second).WithRetries(3)
	ctx := c.Request.Context()

	result, err := checker.Check(ctx, host.Address, host.Port)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"host_id":   hostID,
			"host_name": host.Name,
			"address":   host.Address,
			"status":    StatusUnhealthy,
			"error":     err.Error(),
			"last_checked": time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"host_id":      hostID,
		"host_name":    host.Name,
		"address":      host.Address,
		"status":       result.Status,
		"latency_ms":   result.LatencyMs(),
		"attempt":      result.Attempt,
		"last_checked": result.CheckedAt,
	})
}

// BulkHealthCheckHandler POST /api/hosts/bulk-health
func (s *HealthService) BulkHealthCheckHandler(c *gin.Context) {
	var req struct {
		Hosts      []HostCheckRequest `json:"hosts" binding:"required,min=1,max=100"`
		MaxParallel int               `json:"max_parallel,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate host IDs and get addresses from DB if needed
	hosts := make([]HostCheckRequest, 0, len(req.Hosts))
	for _, h := range req.Hosts {
		if h.Address == "" {
			// Look up address from DB
			var address string
			var port int
			err := s.DB.QueryRow(
				"SELECT address, COALESCE(port, 22) FROM hosts WHERE id = $1",
				h.HostID,
			).Scan(&address, &port)
			if err != nil {
				continue // Skip invalid hosts
			}
			h.Address = address
			if h.Port == 0 {
				h.Port = port
			}
		}
		if h.Port == 0 {
			h.Port = 22 // Default SSH port
		}
		hosts = append(hosts, h)
	}

	// Run bulk check
	results := s.Engine.BulkCheck(hosts, req.MaxParallel)

	// Count summary
	summary := struct {
		Total    int `json:"total"`
		Healthy  int `json:"healthy"`
		Unhealthy int `json:"unhealthy"`
		Unknown  int `json:"unknown"`
	}{
		Total: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		default:
			summary.Unknown++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"results": results,
	})
}

// RegisterHealthCheckHandler POST /api/hosts/:id/health/register
func (s *HealthService) RegisterHealthCheckHandler(c *gin.Context) {
	hostID := c.Param("id")
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Verify ownership
	var ownerID uuid.UUID
	var hostName, address string
	var port int
	err := s.DB.QueryRow(
		"SELECT user_id, name, address, COALESCE(port, 22) FROM hosts WHERE id = $1",
		hostID,
	).Scan(&ownerID, &hostName, &address, &port)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	if ownerID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your host"})
		return
	}

	var req struct {
		Interval string `json:"interval,omitempty"` // e.g., "5m"
		Timeout  string `json:"timeout,omitempty"`  // e.g., "10s"
		Retries  int    `json:"retries,omitempty"`
	}
	c.ShouldBindJSON(&req)

	// Parse durations
	interval := 5 * time.Minute
	if req.Interval != "" {
		if d, err := time.ParseDuration(req.Interval); err == nil {
			interval = d
		}
	}

	timeout := 10 * time.Second
	if req.Timeout != "" {
		if d, err := time.ParseDuration(req.Timeout); err == nil {
			timeout = d
		}
	}

	retries := 3
	if req.Retries > 0 {
		retries = req.Retries
	}

	config := ProbeConfig{
		Type:     ProbeTypeTCP,
		Host:     address,
		Port:     port,
		Timeout:  timeout,
		Retries:  retries,
		Interval: interval,
	}

	// Register with engine
	s.Engine.Register(hostID, hostName, address, config)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Health check registered",
		"host_id":  hostID,
		"interval": interval.String(),
		"timeout":  timeout.String(),
		"retries":  retries,
	})
}

// UnregisterHealthCheckHandler DELETE /api/hosts/:id/health
func (s *HealthService) UnregisterHealthCheckHandler(c *gin.Context) {
	hostID := c.Param("id")
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Verify ownership
	var ownerID uuid.UUID
	err := s.DB.QueryRow(
		"SELECT user_id FROM hosts WHERE id = $1", hostID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	if ownerID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your host"})
		return
	}

	s.Engine.Unregister(hostID)
	c.JSON(http.StatusOK, gin.H{"message": "Health check unregistered"})
}

// GetHealthHistoryHandler GET /api/hosts/:id/health/history
func (s *HealthService) GetHealthHistoryHandler(c *gin.Context) {
	hostID := c.Param("id")

	hours := 24 // Default 24h
	if h := c.Query("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 168 {
			hours = parsed
		}
	}

	health, ok := s.Engine.GetHealth(hostID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no health data for this host"})
		return
	}

	// Filter history by time range
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	filtered := make([]ProbeResult, 0)
	for _, h := range health.History {
		if h.CheckedAt.After(cutoff) {
			filtered = append(filtered, h)
		}
	}

	// Calculate uptime percentage
	uptime := calculateUptime(filtered)

	c.JSON(http.StatusOK, gin.H{
		"host_id":      hostID,
		"hours":        hours,
		"uptime_percent": uptime,
		"total_checks": len(filtered),
		"current_status": health.Status,
		"history":      filtered,
	})
}

// calculateUptime calculates uptime percentage from history
func calculateUptime(history []ProbeResult) float64 {
	if len(history) == 0 {
		return 0
	}

	healthy := 0
	for _, h := range history {
		if h.Status == StatusHealthy {
			healthy++
		}
	}

	return float64(healthy) / float64(len(history)) * 100
}

// GetGlobalHealthHandler GET /api/health/global
func (s *HealthService) GetGlobalHealthHandler(c *gin.Context) {
	all := s.Engine.GetAllHealth()

	summary := struct {
		Total     int `json:"total"`
		Healthy   int `json:"healthy"`
		Unhealthy int `json:"unhealthy"`
		Unknown   int `json:"unknown"`
	}{Total: len(all)}

	for _, h := range all {
		switch h.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		default:
			summary.Unknown++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"hosts":   all,
	})
}
