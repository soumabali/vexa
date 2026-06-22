package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/rdp"
)

func TestRDPSessionsList(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	// Create test session
	g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/rdp/sessions", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var sessions []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &sessions)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestRDPCreateSession(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := map[string]interface{}{
		"host_id":  "host-1",
		"user_id":  "user-1",
		"hostname": "192.168.1.100",
		"port":     3389,
		"username": "admin",
		"password": "pass",
		"domain":   "WORKGROUP",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/rdp/sessions", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var session map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &session)
	require.NoError(t, err)
	assert.NotEmpty(t, session["id"])
	assert.Equal(t, "disconnected", session["status"])
}

func TestRDPGetSession(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/rdp/sessions/"+session.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, session.ID, result["id"])
}

func TestRDPGetSessionNotFound(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/rdp/sessions/non-existent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRDPCloseSession(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/rdp/sessions/"+session.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify session is gone
	_, ok := g.GetSession(session.ID)
	assert.False(t, ok)
}

func TestRDPCloseSessionNotFound(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/rdp/sessions/non-existent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRDPInvalidMethod(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/rdp/sessions", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// WebSocket tests
func TestRDPWebSocketUpgrade(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := httptest.NewServer(mux)
	defer server.Close()

	// wsURL would be used for a proper WebSocket test with initial params
	_ = "ws" + strings.TrimPrefix(server.URL, "http") + "/api/rdp/connect"

	// This would need a proper WebSocket test with initial params
	// For now, just verify the endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/rdp/connect", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Should fail without proper WebSocket headers
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRDPCreateSessionInvalidBody(t *testing.T) {
	g := rdp.NewGateway()
	handler := rdp.NewHandler(g)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/rdp/sessions", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}