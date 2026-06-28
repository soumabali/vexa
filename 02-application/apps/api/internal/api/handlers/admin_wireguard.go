package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
)

// AdminWireguardHandler exposes admin-only WireGuard control endpoints
// (currently: key rotation trigger). It shells out to the rotation script
// and records an audit entry on success or failure.
type AdminWireguardHandler struct {
	auditLogger *audit.Logger
	scriptPath  string // absolute path to rotate-wireguard-keys.sh
}

// NewAdminWireguardHandler wires the handler with the audit logger and
// resolves the rotation script path. The script path is taken from
// env WG_ROTATE_SCRIPT or defaults to ../../scripts/rotate-wireguard-keys.sh
// relative to the API binary's working directory.
func NewAdminWireguardHandler(auditLogger *audit.Logger, scriptPath string) *AdminWireguardHandler {
	if scriptPath == "" {
		scriptPath = os.Getenv("WG_ROTATE_SCRIPT")
	}
	if scriptPath == "" {
		// Fallback: search upward from CWD for scripts/rotate-wireguard-keys.sh
		cwd, _ := os.Getwd()
		for d := cwd; d != "/" && d != "."; d = filepath.Dir(d) {
			candidate := filepath.Join(d, "scripts", "rotate-wireguard-keys.sh")
			if _, err := os.Stat(candidate); err == nil {
				scriptPath = candidate
				break
			}
		}
	}
	return &AdminWireguardHandler{
		auditLogger: auditLogger,
		scriptPath:  scriptPath,
	}
}

// RotateWireguardKeys POST /admin/wg/rotate
// Triggers the rotation script. Returns 202 on async start; the script
// itself is synchronous but may take a few seconds while it syncs the
// live interface.
func (h *AdminWireguardHandler) RotateWireguardKeys(c *gin.Context) {
	if h.scriptPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "rotate script not configured (WG_ROTATE_SCRIPT)",
		})
		return
	}
	if _, err := os.Stat(h.scriptPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("rotate script not found: %s", h.scriptPath),
		})
		return
	}

	userID := uuid.Nil
	if v, ok := c.Get("user_id"); ok {
		if s, ok := v.(string); ok {
			if u, err := uuid.Parse(s); err == nil {
				userID = u
			}
		}
	}

	// Build command. We use a small indirection so tests can override exec.Command.
	cmd := h.buildCommand(h.scriptPath)

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	// Audit
	details := map[string]interface{}{
		"script":   h.scriptPath,
		"output":   output,
		"exit_ok":  err == nil,
	}
	uid := userID
	if h.auditLogger != nil {
		_ = h.auditLogger.Log(
			"admin.wg.rotate",
			&uid,
			nil,
			c.ClientIP(),
			details,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "rotation script failed",
			"output": output,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "rotated",
		"output": output,
	})
}

// buildCommand is overridable for tests.
var execCommandBuilder = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// SetRotateCommandBuilder lets tests inject a fake command builder.
// Returns a restore function.
func SetRotateCommandBuilder(b func(scriptPath string) *exec.Cmd) func() {
	prev := execCommandBuilder
	execCommandBuilder = func(name string, args ...string) *exec.Cmd {
		// script path comes through as the first arg; map back to script form.
		if len(args) > 0 {
			return b(args[0])
		}
		return b(name)
	}
	return func() { execCommandBuilder = prev }
}

// internalBuildCommand is the production builder that calls exec.Command.
// It is overridable via SetRotateCommandBuilder.
func (h *AdminWireguardHandler) buildCommand(scriptPath string) *exec.Cmd {
	return execCommandBuilder(scriptPath)
}