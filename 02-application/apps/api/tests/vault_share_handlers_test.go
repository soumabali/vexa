package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/vault"
)

func TestVaultShareHandlersTest(t *testing.T) {
	// RegisterShareRoutes must wire all share HTTP endpoints.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := vault.NewShareHandler(nil)
	require.NotNil(t, handler)
	vault.RegisterShareRoutes(r.Group("/api/v1/vault"), handler)

	// Probe each route — OPTIONS would 404 in gin (no router method registered),
	// so use a GET probe and accept either 401 (auth-required) or 404/405.
	probe := func(method, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(""))
		r.ServeHTTP(w, req)
		return w.Code
	}

	// All routes must be registered (NOT 404 with "no route").
	checks := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/vault/shares"},
		{http.MethodGet, "/api/v1/vault/shares/abc"},
		{http.MethodPost, "/api/v1/vault/shares/abc/accept"},
		{http.MethodPost, "/api/v1/vault/shares/abc/reject"},
		{http.MethodPost, "/api/v1/vault/shares/abc/revoke"},
		{http.MethodGet, "/api/v1/vault/shares/stats"},
	}
	for _, c := range checks {
		code := probe(c.method, c.path)
		assert.NotEqual(t, http.StatusNotFound, code,
			"route %s %s should be registered", c.method, c.path)
	}
}

func TestVaultShareHandler_NewShareHandler(t *testing.T) {
	h := vault.NewShareHandler(nil)
	require.NotNil(t, h)
	assert.NotNil(t, h)
}
