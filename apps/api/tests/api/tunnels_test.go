package apitests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/wireguard"
)

func newTunnelHandler(t *testing.T) (*wireguard.WireGuardService, sqlmock.Sqlmock, *handlers.TunnelHandler) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})
	handler := handlers.NewTunnelHandler(svc)
	return svc, mock, handler
}

func tunnelColumns() []string {
	return []string{
		"id", "user_id", "host_id", "interface_name",
		"server_private_key", "server_public_key", "client_public_key",
		"preshared_key", "server_ip", "client_ip", "listen_port",
		"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
		"status", "is_enabled", "last_handshake_at", "bytes_sent",
		"bytes_received", "last_rotated_at", "created_at", "updated_at",
	}
}

func buildTunnelRow(id, userID, hostID uuid.UUID) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(tunnelColumns()).
		AddRow(id, userID, hostID, "wg0",
			"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
			51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
			"connected", true, nil, int64(100), int64(200),
			nil, now, now)
}

func setupTunnelRoute(handler *handlers.TunnelHandler, method, path string, fn func(c *gin.Context)) *gin.Engine {
	r := setupRouter()
	r.Handle(method, path, func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		fn(c)
	})
	return r
}

func TestTunnelHandler_Create(t *testing.T) {
	t.Run("creates tunnel successfully", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		userID := uuid.New()
		hostID := uuid.New()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
			WithArgs(userID, hostID.String()).
			WillReturnRows(sqlmock.NewRows(tunnelColumns()))
		mock.ExpectQuery(`INSERT INTO wireguard_tunnels`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(uuid.New(), time.Now(), time.Now()))
		mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.POST("/tunnels", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Create(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tunnels",
			jsonBody(map[string]interface{}{
				"host_id": hostID.String(),
				"port":    51820,
			}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid json returns 400", func(t *testing.T) {
		_, _, handler := newTunnelHandler(t)
		TestInvalidJSON(t, "POST", "/tunnels", handler.Create, "")
	})

	t.Run("missing host_id returns 400", func(t *testing.T) {
		_, _, handler := newTunnelHandler(t)
		r := setupRouter()
		r.POST("/tunnels", func(c *gin.Context) {
			c.Set("user_id", uuid.New())
			handler.Create(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tunnels",
			jsonBody(map[string]interface{}{"port": 51820}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTunnelHandler_Get(t *testing.T) {
	t.Run("returns tunnel", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()
		hostID := uuid.New()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(buildTunnelRow(tunnelID, userID, hostID))

		r := setupRouter()
		r.GET("/tunnels/:id", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Get(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tunnels/"+tunnelID.String(), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found returns 404", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnError(sql.ErrNoRows)

		r := setupRouter()
		r.GET("/tunnels/:id", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Get(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tunnels/"+tunnelID.String(), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTunnelHandler_Update(t *testing.T) {
	t.Run("updates tunnel", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()
		hostID := uuid.New()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(buildTunnelRow(tunnelID, userID, hostID))
		mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.PATCH("/tunnels/:id", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Update(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/tunnels/"+tunnelID.String(),
			jsonBody(map[string]interface{}{
				"allowed_ips": []string{"10.0.0.0/8"},
				"mtu":         1400,
			}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTunnelHandler_Delete(t *testing.T) {
	t.Run("deletes tunnel", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()
		hostID := uuid.New()
		now := time.Now()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(sqlmock.NewRows(tunnelColumns()).
				AddRow(tunnelID, userID, hostID, "wg0",
					"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
					51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
					"connected", true, nil, int64(0), int64(0),
					nil, now, now))
		mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
			WithArgs(tunnelID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.DELETE("/tunnels/:id", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Delete(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/tunnels/"+tunnelID.String(), nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTunnelHandler_Enable(t *testing.T) {
	_, mock, handler := newTunnelHandler(t)
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", false, nil, int64(0), int64(0),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := setupRouter()
	r.POST("/tunnels/:id/enable", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Enable(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tunnels/"+tunnelID.String()+"/enable", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTunnelHandler_Disable(t *testing.T) {
	_, mock, handler := newTunnelHandler(t)
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(0), int64(0),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := setupRouter()
	r.POST("/tunnels/:id/disable", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Disable(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tunnels/"+tunnelID.String()+"/disable", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTunnelHandler_Rotate(t *testing.T) {
	_, mock, handler := newTunnelHandler(t)
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(0), int64(0),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := setupRouter()
	r.POST("/tunnels/:id/rotate", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Rotate(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tunnels/"+tunnelID.String()+"/rotate", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTunnelHandler_Config(t *testing.T) {
	t.Run("returns config as JSON", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()
		hostID := uuid.New()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(buildTunnelRow(tunnelID, userID, hostID))
		mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
			WithArgs(hostID).
			WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("203.0.113.1"))

		r := setupRouter()
		r.GET("/tunnels/:id/config", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Config(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tunnels/"+tunnelID.String()+"/config", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns config as text/plain with Accept header", func(t *testing.T) {
		_, mock, handler := newTunnelHandler(t)
		tunnelID := uuid.New()
		userID := uuid.New()
		hostID := uuid.New()
		now := time.Now()

		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(sqlmock.NewRows(tunnelColumns()).
				AddRow(tunnelID, userID, hostID, "wg0",
					"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
					51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
					"connected", true, nil, int64(0), int64(0),
					nil, now, now))
		mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
			WithArgs(hostID).
			WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("203.0.113.1"))

		r := setupRouter()
		r.GET("/tunnels/:id/config", func(c *gin.Context) {
			c.Set("user_id", userID)
			handler.Config(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tunnels/"+tunnelID.String()+"/config?format=conf", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "[Interface]")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTunnelHandler_Stats(t *testing.T) {
	_, mock, handler := newTunnelHandler(t)
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(100), int64(200),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET bytes_sent`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := setupRouter()
	r.GET("/tunnels/:id/stats", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Stats(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tunnels/"+tunnelID.String()+"/stats", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTunnelHandler_List(t *testing.T) {
	_, mock, handler := newTunnelHandler(t)
	userID := uuid.New()
	hostID := uuid.New()
	tunnelID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(100), int64(200),
				nil, now, now))

	r := setupRouter()
	r.GET("/tunnels", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tunnels?host_id="+hostID.String(), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
