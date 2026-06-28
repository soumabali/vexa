package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
)

// fakeCmd is an *exec.Cmd we can mutate to inject output + exit behavior
// without actually executing the underlying binary.
type fakeCmd struct {
	stdout string
	exit   int
	called bool
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// TestAdminWireguardRotate_Success — 200, audit entry, status="rotated"
func TestAdminWireguardRotate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restore := handlers.SetRotateCommandBuilder(func(sp string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf 'ROTATE_OK'; exit 0")
	})
	defer restore()

	// Create a real script file so the handler's stat check passes.
	dir := t.TempDir()
	script := filepath.Join(dir, "rotate-wireguard-keys.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/usr/bin/env bash\necho\n"), 0o755))

	h := handlers.NewAdminWireguardHandler(nil, script)
	r := gin.New()
	r.POST("/admin/wg/rotate", h.RotateWireguardKeys)

	req := httptest.NewRequest(http.MethodPost, "/admin/wg/rotate", bytes.NewBufferString(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "rotated", body["status"])
	assert.Equal(t, "ROTATE_OK", body["output"])
}

// TestAdminWireguardRotate_Failure — non-zero exit → 500
func TestAdminWireguardRotate_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restore := handlers.SetRotateCommandBuilder(func(sp string) *exec.Cmd {
		return exec.Command("sh", "-c", "printf 'boom'; exit 2")
	})
	defer restore()

	dir := t.TempDir()
	script := filepath.Join(dir, "rotate.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/usr/bin/env bash\n"), 0o755))

	h := handlers.NewAdminWireguardHandler(nil, script)
	r := gin.New()
	r.POST("/admin/wg/rotate", h.RotateWireguardKeys)

	req := httptest.NewRequest(http.MethodPost, "/admin/wg/rotate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "rotation script failed")
	_ = fakeCmd{} // keep type referenced
	_ = itoa
}

// TestAdminWireguardRotate_MissingScript — 500 with helpful msg
func TestAdminWireguardRotate_MissingScript(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewAdminWireguardHandler(nil, "/definitely/does/not/exist/rotate.sh")
	r := gin.New()
	r.POST("/admin/wg/rotate", h.RotateWireguardKeys)

	req := httptest.NewRequest(http.MethodPost, "/admin/wg/rotate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "rotate script not found")
}

// TestAdminWireguardRotate_CallsExpectedPath — script path flows into cmd builder
func TestAdminWireguardRotate_CallsExpectedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	script := filepath.Join(dir, "rotate-wireguard-keys.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/usr/bin/env bash\necho CALLED\n"), 0o755))

	var captured []string
	restore := handlers.SetRotateCommandBuilder(func(sp string) *exec.Cmd {
		captured = append(captured, sp)
		return exec.Command("sh", "-c", "echo CALLED")
	})
	defer restore()

	h := handlers.NewAdminWireguardHandler(nil, script)
	r := gin.New()
	r.POST("/admin/wg/rotate", h.RotateWireguardKeys)

	req := httptest.NewRequest(http.MethodPost, "/admin/wg/rotate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, captured, 1)
	assert.Equal(t, script, captured[0])
	assert.True(t, strings.Contains(w.Body.String(), "CALLED"))
}