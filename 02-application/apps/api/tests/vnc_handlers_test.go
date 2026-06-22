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

	"github.com/soumabali/vexa/internal/vnc"
)

func TestVNCSessionsList(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	// Create test session
	p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/vnc/sessions", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var sessions []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &sessions)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestVNCCreateSession(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := map[string]interface{}{
		"host_id":  "host-1",
		"user_id":  "user-1",
		"hostname": "192.168.1.100",
		"port":     5900,
		"password": "pass",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/vnc/sessions", bytes.NewReader(bodyJSON))
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

func TestVNCGetSession(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/vnc/sessions/"+session.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, session.ID, result["id"])
}

func TestVNCGetSessionNotFound(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/vnc/sessions/non-existent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestVNCCloseSession(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/vnc/sessions/"+session.ID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify session is gone
	_, ok := p.GetSession(session.ID)
	assert.False(t, ok)
}

func TestVNCCloseSessionNotFound(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/vnc/sessions/non-existent", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestVNCInvalidMethod(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/vnc/sessions", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestVNCCreateSessionInvalidBody(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/vnc/sessions", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVNCWebSocketUpgrade(t *testing.T) {
	p := vnc.NewProxy()
	handler := vnc.NewHandler(p)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Should fail without proper WebSocket headers
	req := httptest.NewRequest(http.MethodGet, "/api/vnc/connect", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}