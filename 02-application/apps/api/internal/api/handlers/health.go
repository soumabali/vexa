package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler serves /health/live and /health/ready.
//
// /health/live  — process is up. Always 200 unless the binary is broken.
// /health/ready — process can serve traffic. 200 only if every dependency
//                 check passes; otherwise 503.
//
// Each dependency check returns a sub-object:
//   { "ok": true|false, "latency_ms": <number>, "error": "..."? }
//
// The aggregate response is:
//
//   {
//     "status": "ready" | "not_ready",
//     "checks": {
//       "db":     { "ok": true,  "latency_ms": 1.234 },
//       "redis":  { "ok": true,  "latency_ms": 0.567 },
//       "config": { "ok": true,  "checks": { "jwt_secret": true } },
//       "disk":   { "ok": true,  "free_pct": 42.1, "path": "..." },
//       "wg":     { "ok": false, "enabled": true, "error": "..." }
//     }
//   }
type HealthHandler struct {
	db             *sql.DB
	redis          *redis.Client
	dataDir        string
	wgCheckEnabled bool
	wgInterface    string
}

// NewHealthHandler wires the health handler. dataDir is the on-disk path
// checked for free space (typically the API's persistent volume mount).
func NewHealthHandler(db *sql.DB, redis *redis.Client, dataDir string) *HealthHandler {
	if dataDir == "" {
		dataDir = os.TempDir()
	}
	iface := os.Getenv("WG_INTERFACE")
	if iface == "" {
		iface = "wg0"
	}
	return &HealthHandler{
		db:             db,
		redis:          redis,
		dataDir:        dataDir,
		wgCheckEnabled: os.Getenv("WG_CHECK_ENABLED") == "1",
		wgInterface:    iface,
	}
}

// checkResult is the wire shape for each dependency.
type checkResult struct {
	OK        bool        `json:"ok"`
	LatencyMs float64     `json:"latency_ms"`
	Error     string      `json:"error,omitempty"`
}

// pingDB measures DB roundtrip latency.
func (h *HealthHandler) pingDB(ctx context.Context) checkResult {
	start := time.Now()
	err := h.db.PingContext(ctx)
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	res := checkResult{OK: err == nil, LatencyMs: latency}
	if err != nil {
		res.Error = err.Error()
	}
	return res
}

// pingRedis measures Redis roundtrip latency.
func (h *HealthHandler) pingRedis(ctx context.Context) checkResult {
	start := time.Now()
	err := h.redis.Ping(ctx).Err()
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	res := checkResult{OK: err == nil, LatencyMs: latency}
	if err != nil {
		res.Error = err.Error()
	}
	return res
}

// checkConfig verifies required env vars are present.
func (h *HealthHandler) checkConfig() gin.H {
	jwtSet := os.Getenv("JWT_SECRET") != "" && os.Getenv("JWT_SECRET") != "change-me-in-production"
	dbSet := os.Getenv("DB_NAME") != "" || os.Getenv("DATABASE_URL") != ""
	checks := gin.H{
		"jwt_secret": jwtSet,
		"db_config":  dbSet,
	}
	return gin.H{
		"ok":     jwtSet && dbSet,
		"checks": checks,
	}
}

// checkDisk verifies at least 10% free on the data dir's filesystem.
func (h *HealthHandler) checkDisk() gin.H {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(h.dataDir, &stat); err != nil {
		return gin.H{
			"ok":    false,
			"error": err.Error(),
			"path":  h.dataDir,
		}
	}
	// BavSize is int64 on Linux; safe to compute as uint64 division.
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	var pct float64
	if total > 0 {
		pct = float64(free) * 100.0 / float64(total)
	}
	ok := pct >= 10.0
	res := gin.H{
		"ok":       ok,
		"path":     h.dataDir,
		"free_pct": pct,
	}
	if !ok {
		res["error"] = "disk space below 10%"
	}
	return res
}

// checkWireGuard optionally probes the live interface via `wg show`.
// Skipped (returns ok=true with enabled=false) when WG_CHECK_ENABLED!=1
// or when `wg` is not in PATH.
func (h *HealthHandler) checkWireGuard() gin.H {
	if !h.wgCheckEnabled {
		return gin.H{"ok": true, "enabled": false}
	}
	if _, err := exec.LookPath("wg"); err != nil {
		return gin.H{
			"ok":       false,
			"enabled":  true,
			"interface": h.wgInterface,
			"error":    "wg binary not found in PATH",
		}
	}
	cmd := exec.Command("wg", "show", h.wgInterface)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return gin.H{
			"ok":        false,
			"enabled":   true,
			"interface": h.wgInterface,
			"error":     strings.TrimSpace(string(out)),
		}
	}
	return gin.H{
		"ok":        true,
		"enabled":   true,
		"interface": h.wgInterface,
	}
}

// Live GET /health/live — liveness probe. Always 200 if the goroutine is
// servicing requests. No dependency checks.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready GET /health/ready — readiness probe. Aggregates DB, Redis, env,
// disk, and (optionally) WireGuard. 503 if any dependency is unhealthy.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbCheck := h.pingDB(ctx)
	redisCheck := h.pingRedis(ctx)
	cfgCheck := h.checkConfig()
	diskCheck := h.checkDisk()
	wgCheck := h.checkWireGuard()

	allOK := dbCheck.OK && redisCheck.OK && cfgCheck["ok"] == true &&
		diskCheck["ok"] == true && wgCheck["ok"] == true

	checks := gin.H{
		"db": gin.H{
			"ok":         dbCheck.OK,
			"latency_ms": dbCheck.LatencyMs,
		},
		"redis": gin.H{
			"ok":         redisCheck.OK,
			"latency_ms": redisCheck.LatencyMs,
		},
		"config": cfgCheck,
		"disk":   diskCheck,
		"wg":     wgCheck,
	}
	if !dbCheck.OK {
		checks["db"].(gin.H)["error"] = dbCheck.Error
	}
	if !redisCheck.OK {
		checks["redis"].(gin.H)["error"] = redisCheck.Error
	}

	status := http.StatusOK
	statusLabel := "ready"
	if !allOK {
		status = http.StatusServiceUnavailable
		statusLabel = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":    statusLabel,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	})
}

// ensure filepath import is used (avoids dead-code warnings if we ever
// surface the data path in logs).
var _ = filepath.Clean