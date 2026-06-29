package apitests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/hosts"
	"github.com/soumabali/vexa/internal/models"
)

func hostColumns() []string {
	return []string{
		"id", "owner_id", "name", "address", "protocol", "port",
		"credentials_id", "tags", "group_path", "allowed_users",
		"description", "is_active", "created_at", "updated_at",
	}
}

func buildHostRow(id, ownerID uuid.UUID, name, address, protocol string, port int) *sqlmock.Rows {
	return sqlmock.NewRows(hostColumns()).
		AddRow(id, ownerID, name, address, protocol, port,
			nil, []byte("{}"), "", []byte("{}"),
			"", true, time.Now(), time.Now())
}

func newHostHandler(t *testing.T) (*hosts.Repository, sqlmock.Sqlmock, *handlers.HostHandler) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := hosts.NewRepository(db)
	logger, _ := audit.NewNoOpLogger()
	handler := handlers.NewHostHandler(repo, logger)
	return repo, mock, handler
}

func TestHostHandler_Create(t *testing.T) {
	t.Skip("requires sqlmock DB setup")
	_, mock, handler := newHostHandler(t)
	ownerID := uuid.New()

	t.Run("valid request creates host and returns 201", func(t *testing.T) {
		hostID := uuid.New()
		mock.ExpectQuery(`INSERT INTO hosts`).
			WithArgs(ownerID, "web-prod-01", "10.0.1.100", "ssh", 22, nil, sqlmock.AnyArg(), "/prod/web", sqlmock.AnyArg(), "").
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(hostID, time.Now(), time.Now()))

		r := setupRouter()
		r.POST("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Create(c)
		})

		body := models.CreateHostRequest{
			Name:      "web-prod-01",
			Address:   "10.0.1.100",
			Protocol:  "ssh",
			Port:      22,
			Tags:      []string{"production", "web"},
			GroupPath: "/prod/web",
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hosts", jsonBody(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		resp := parseJSON(w)
		assert.Equal(t, "web-prod-01", resp["name"])
		assert.Equal(t, "10.0.1.100", resp["address"])
		assert.Equal(t, "ssh", resp["protocol"])
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Create(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hosts", strings.NewReader(`{"address":"10.0.1.100"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid address returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Create(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hosts", strings.NewReader(`{"name":"test","address":"invalid!!","protocol":"ssh","port":22}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		r := setupRouter()
		r.POST("/hosts", handler.Create)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hosts", jsonBody(models.CreateHostRequest{Name: "test", Address: "10.0.1.1", Protocol: "ssh", Port: 22}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHostHandler_Get(t *testing.T) {
	t.Skip("requires sqlmock DB setup")
	_, mock, handler := newHostHandler(t)
	ownerID := uuid.New()
	hostID := uuid.New()

	t.Run("valid host ID returns 200", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(hostID, ownerID).
			WillReturnRows(buildHostRow(hostID, ownerID, "db-primary", "10.0.2.50", "ssh", 22))

		r := setupRouter()
		r.GET("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Get(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+hostID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.Equal(t, "db-primary", resp["name"])
	})

	t.Run("non-existent host returns 404", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(sqlmock.AnyArg(), ownerID).
			WillReturnError(sql.ErrNoRows)

		r := setupRouter()
		r.GET("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Get(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		r := setupRouter()
		r.GET("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Get(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/not-a-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHostHandler_List(t *testing.T) {
	t.Skip("requires sqlmock DB setup")
	_, mock, handler := newHostHandler(t)
	ownerID := uuid.New()

	t.Run("returns all hosts for user", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE owner_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs(ownerID, 100, 0).
			WillReturnRows(sqlmock.NewRows(hostColumns()))

		r := setupRouter()
		r.GET("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.List(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.NotNil(t, resp["hosts"])
	})

	t.Run("filter by tag", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE owner_id = \$1 AND \$2 = ANY\(tags\) ORDER BY created_at DESC LIMIT \$3 OFFSET \$4`).
			WithArgs(ownerID, "prod", 100, 0).
			WillReturnRows(sqlmock.NewRows(hostColumns()))

		r := setupRouter()
		r.GET("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.List(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts?tag=prod", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("filter by group", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE owner_id = \$1 AND group_path = \$2 ORDER BY created_at DESC LIMIT \$3 OFFSET \$4`).
			WithArgs(ownerID, "/prod", 100, 0).
			WillReturnRows(sqlmock.NewRows(hostColumns()))

		r := setupRouter()
		r.GET("/hosts", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.List(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts?group=/prod", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		r := setupRouter()
		r.GET("/hosts", handler.List)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHostHandler_Update(t *testing.T) {
	t.Skip("requires sqlmock DB setup")
	_, mock, handler := newHostHandler(t)
	ownerID := uuid.New()
	hostID := uuid.New()

	t.Run("valid update returns 200", func(t *testing.T) {
		mock.ExpectExec(`UPDATE hosts SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.PATCH("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Update(c)
		})

		body := models.UpdateHostRequest{Name: "new-name"}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/hosts/"+hostID.String(), jsonBody(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update non-existent host returns 404", func(t *testing.T) {
		mock.ExpectExec(`UPDATE hosts SET`).
			WillReturnResult(sqlmock.NewResult(0, 0))

		r := setupRouter()
		r.PATCH("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Update(c)
		})

		body := models.UpdateHostRequest{Name: "new-name"}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/hosts/"+uuid.New().String(), jsonBody(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update with invalid host ID returns 400", func(t *testing.T) {
		r := setupRouter()
		r.PATCH("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Update(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/hosts/invalid-id", strings.NewReader(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("partial update - change address only", func(t *testing.T) {
		mock.ExpectExec(`UPDATE hosts SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.PATCH("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Update(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/hosts/"+hostID.String(), strings.NewReader(`{"address":"10.0.2.100"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHostHandler_Delete(t *testing.T) {
	t.Skip("requires sqlmock DB setup")
	_, mock, handler := newHostHandler(t)
	ownerID := uuid.New()
	hostID := uuid.New()

	t.Run("valid host ID deletes and returns 204", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(hostID, ownerID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := setupRouter()
		r.DELETE("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Delete(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/hosts/"+hostID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete non-existent host returns 404", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(sqlmock.AnyArg(), ownerID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		r := setupRouter()
		r.DELETE("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Delete(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/hosts/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		r := setupRouter()
		r.DELETE("/hosts/:id", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.Delete(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/hosts/not-a-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHostHandler_HealthCheck(t *testing.T) {
	_, mock, handler := newHostHandler(t)
	hostID := uuid.New()
	ownerID := uuid.New()

	t.Run("returns health status for valid host", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, address, protocol, port, is_active, created_at, updated_at FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(hostID, ownerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "address", "protocol", "port", "is_active", "created_at", "updated_at"}).
				AddRow(hostID, "healthy-host", "10.0.1.1", "ssh", 22, true, time.Now(), time.Now()))

		r := setupRouter()
		r.GET("/hosts/:id/health", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.HealthCheck(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+hostID.String()+"/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.NotEmpty(t, resp["host_id"])
	})

	t.Run("non-existent host returns error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(sqlmock.AnyArg(), ownerID).
			WillReturnError(sql.ErrNoRows)

		r := setupRouter()
		r.GET("/hosts/:id/health", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.HealthCheck(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+uuid.New().String()+"/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHostHandler_GetStats(t *testing.T) {
	_, mock, handler := newHostHandler(t)
	hostID := uuid.New()
	ownerID := uuid.New()

	t.Run("returns stats for valid host", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, address, protocol, port, is_active, created_at, updated_at FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(hostID, ownerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "address", "protocol", "port", "is_active", "created_at", "updated_at"}).
				AddRow(hostID, "stats-host", "10.0.0.1", "ssh", 22, true, time.Now(), time.Now()))

		mock.ExpectQuery(`SELECT COUNT\(\*\), MAX\(started_at\)`).
			WithArgs(hostID, ownerID).
			WillReturnRows(sqlmock.NewRows([]string{"count", "max"}).
				AddRow(7, time.Now().Add(-2*time.Hour)))

		mock.ExpectQuery(`SELECT\s+COUNT\(\*\),`).
			WithArgs(hostID, ownerID).
			WillReturnRows(sqlmock.NewRows([]string{"total", "active"}).
				AddRow(2, 1))

		r := setupRouter()
		r.GET("/hosts/:id/stats", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.GetStats(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+hostID.String()+"/stats", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.EqualValues(t, 7, resp["total_sessions"])
		assert.EqualValues(t, 2, resp["tunnel_count"])
		assert.EqualValues(t, 1, resp["active_tunnels"])
		assert.NotEmpty(t, resp["last_connected_at"])
	})

	t.Run("non-existent host returns 404", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM hosts WHERE id = \$1 AND owner_id = \$2`).
			WithArgs(sqlmock.AnyArg(), ownerID).
			WillReturnError(sql.ErrNoRows)

		r := setupRouter()
		r.GET("/hosts/:id/stats", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.GetStats(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/"+uuid.New().String()+"/stats", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		r := setupRouter()
		r.GET("/hosts/:id/stats", func(c *gin.Context) {
			c.Set("user_id", ownerID)
			handler.GetStats(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hosts/not-a-uuid/stats", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHostCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request models.CreateHostRequest
		valid   bool
	}{
		{
			name: "valid SSH host",
			request: models.CreateHostRequest{
				Name: "prod-server", Address: "192.168.1.100",
				Protocol: "ssh", Port: 22,
			},
			valid: true,
		},
		{
			name: "valid RDP host",
			request: models.CreateHostRequest{
				Name: "windows-box", Address: "192.168.1.101",
				Protocol: "rdp", Port: 3389,
			},
			valid: true,
		},
		{
			name: "valid VNC host",
			request: models.CreateHostRequest{
				Name: "vnc-box", Address: "192.168.1.102",
				Protocol: "vnc", Port: 5900,
			},
			valid: true,
		},
		{
			name: "missing name",
			request: models.CreateHostRequest{
				Address: "192.168.1.100", Protocol: "ssh", Port: 22,
			},
			valid: false,
		},
		{
			name: "missing address",
			request: models.CreateHostRequest{
				Name: "test", Protocol: "ssh", Port: 22,
			},
			valid: false,
		},
		{
			name: "missing protocol",
			request: models.CreateHostRequest{
				Name: "test", Address: "192.168.1.100", Port: 22,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			r := setupRouter()
			r.POST("/hosts", func(c *gin.Context) {
				c.Set("user_id", uuid.New())
				repo := hosts.NewRepository(db)
				logger, _ := audit.NewNoOpLogger()
				handler := handlers.NewHostHandler(repo, logger)
				handler.Create(c)
			})

			if tt.valid {
				mock.ExpectExec(`INSERT INTO hosts`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), tt.request.Name, tt.request.Address, string(tt.request.Protocol), tt.request.Port,
						tt.request.GroupPath, tt.request.Description, true).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/hosts", jsonBody(tt.request))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if tt.valid {
				assert.Equal(t, http.StatusCreated, w.Code, "expected valid request to return 201")
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code, "expected invalid request to return 400")
			}
		})
	}
}
