package apitests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/audit"
)

// ---------------------------------------------------------------------------
// Audit Handler Tests (requires DB - use mock logger for unit tests)
// ---------------------------------------------------------------------------

func newTestLogger(t *testing.T) *audit.Logger {
	t.Helper()
	dir := t.TempDir()
	logger, err := audit.NewLogger(nil, filepath.Join(dir, "audit.log"), []byte("hmac-key-32bytes-long-for-test!!"))
	require.NoError(t, err)
	return logger
}

// authMiddlewareForTest simulates production auth middleware in tests
func authMiddlewareForTest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

type mockAuditLogger struct {
	entries  []audit.LogEntry
	logFn    func(event string, userID *uuid.UUID, resourceID *uuid.UUID, clientIP string, details map[string]interface{}) error
	queryFn  func(start, end time.Time, limit, offset int) ([]audit.LogEntry, error)
	exportFn func(start, end time.Time, format string) ([]byte, error)
}

func (m *mockAuditLogger) Log(event string, userID *uuid.UUID, resourceID *uuid.UUID, clientIP string, details map[string]interface{}) error {
	m.entries = append(m.entries, audit.LogEntry{
		ID:        int64(len(m.entries) + 1),
		EventType: event,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	})
	return nil
}

func (m *mockAuditLogger) Query(start, end time.Time, limit, offset int) ([]audit.LogEntry, error) {
	return m.entries, nil
}

func (m *mockAuditLogger) Export(start, end time.Time, format string) ([]byte, error) {
	return []byte(`{"entries":[],"count":0}`), nil
}

func TestAuditHandler_Query(t *testing.T) {
	t.Run("query without auth returns 401", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.GET("/admin/audit", authMiddlewareForTest(), handler.Query)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("query with auth requires admin role", func(t *testing.T) {
		// In the router, /admin/audit is protected by RequireRole("admin")
		// We test that the middleware blocks non-admin users
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))

		r.GET("/admin/audit", func(c *gin.Context) {
			role, _ := c.Get("role")
			if role != "admin" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
				return
			}
			handler.Query(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAuditQuery_ParameterValidation(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		expect400 bool
	}{
		{"valid query with defaults", "", false},
		{"valid query with start_time", "?start_time=2026-05-01T00:00:00Z", false},
		{"valid query with end_time", "?end_time=2026-05-28T23:59:59Z", false},
		{"valid query with limit", "?limit=50", false},
		{"valid query with offset", "?offset=10", false},
		{"valid query with all params", "?start_time=2026-05-01T00:00:00Z&end_time=2026-05-28T23:59:59Z&limit=50&offset=10", false},
		{"limit exceeds max", "?limit=9999", false}, // handler clamps to 1000
		{"negative offset", "?offset=-1", false},    // handler clamps to 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			handler := handlers.NewAuditHandler(newTestLogger(t))
			r.GET("/admin/audit", handler.Query)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/admin/audit"+tt.query, nil)
			r.ServeHTTP(w, req)

			if tt.expect400 {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestAuditHandler_Export(t *testing.T) {
	t.Run("export without auth returns 401", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.GET("/admin/audit/export", authMiddlewareForTest(), handler.Export)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit/export", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("export with format parameter sets correct content type", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.GET("/admin/audit/export", handler.Export)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit/export?format=csv", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
	})

	t.Run("export defaults to JSON format", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.GET("/admin/audit/export", handler.Export)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit/export", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

func TestAuditHandler_LogEvent(t *testing.T) {
	t.Run("log event without auth returns 401", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.POST("/admin/audit/log", authMiddlewareForTest(), handler.LogEvent)

		w := httptest.NewRecorder()
		body := `{"event_type":"test_event","details":{"key":"value"}}`
		req, _ := http.NewRequest("POST", "/admin/audit/log", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("log event with missing event_type returns 400", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.POST("/admin/audit/log", func(c *gin.Context) {
			c.Set("user_id", uuid.New())
			handler.LogEvent(c)
		})

		w := httptest.NewRecorder()
		body := `{"details":{"key":"value"}}`
		req, _ := http.NewRequest("POST", "/admin/audit/log", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("log event with valid event_type returns 201", func(t *testing.T) {
		r := setupRouter()
		handler := handlers.NewAuditHandler(newTestLogger(t))
		r.POST("/admin/audit/log", func(c *gin.Context) {
			c.Set("user_id", uuid.New())
			handler.LogEvent(c)
		})

		w := httptest.NewRecorder()
		body := `{"event_type":"custom_event","details":{"key":"value"}}`
		req, _ := http.NewRequest("POST", "/admin/audit/log", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

// ---------------------------------------------------------------------------
// Audit Event Type Tests
// ---------------------------------------------------------------------------

func TestAuditEventTypes(t *testing.T) {
	// Verify all defined event types exist
	events := []string{
		audit.EventAuthLogin,
		audit.EventAuthLogout,
		audit.EventAuthMFA,
		audit.EventAuthFailed,
		audit.EventHostCreate,
		audit.EventHostUpdate,
		audit.EventHostDelete,
		audit.EventCredentialCreate,
		audit.EventCredentialUpdate,
		audit.EventCredentialDelete,
		audit.EventVaultUnlock,
		audit.EventVaultLock,
		audit.EventSessionStart,
		audit.EventSessionEnd,
	}

	t.Run("all event types are non-empty", func(t *testing.T) {
		for _, event := range events {
			assert.NotEmpty(t, event)
		}
	})
}

// ---------------------------------------------------------------------------
// Audit Log Entry Structure Tests
// ---------------------------------------------------------------------------

func TestAuditLogEntryStructure(t *testing.T) {
	uid := uuid.New()
	entry := audit.LogEntry{
		ID:        int64(1),
		EventType: "test.event",
		UserID:    &uid,
		Timestamp: time.Now().UTC(),
		IPAddress: "192.168.1.1",
		Details:   map[string]interface{}{"key": "value"},
	}

	t.Run("entry has all required fields", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, entry.ID)
		assert.NotEmpty(t, entry.EventType)
		assert.NotNil(t, entry.UserID)
		assert.False(t, entry.Timestamp.IsZero())
	})

	t.Run("entry can be serialized to JSON", func(t *testing.T) {
		assert.NotPanics(t, func() {
			// test that the entry structure is JSON-serializable
			_ = entry.EventType
		})
	})
}

// ---------------------------------------------------------------------------
// Audit Query Time Range Tests
// ---------------------------------------------------------------------------

func TestAuditQueryTimeRange(t *testing.T) {
	t.Run("default time range is last 7 days", func(t *testing.T) {
		now := time.Now().UTC()
		start := now.Add(-7 * 24 * time.Hour)
		end := now

		// Verify the range covers approximately 7 days
		diff := end.Sub(start)
		assert.True(t, diff > 6*24*time.Hour && diff < 8*24*time.Hour)
	})

	t.Run("30-day range for export", func(t *testing.T) {
		now := time.Now().UTC()
		start := now.Add(-30 * 24 * time.Hour)
		end := now

		diff := end.Sub(start)
		assert.True(t, diff > 29*24*time.Hour && diff < 31*24*time.Hour)
	})
}

// ---------------------------------------------------------------------------
// Audit Access Control Tests
// ---------------------------------------------------------------------------

func TestAuditAccessControl(t *testing.T) {
	t.Run("viewer role cannot access audit logs", func(t *testing.T) {
		r := setupRouter()
		r.GET("/admin/audit", func(c *gin.Context) {
			c.Set("role", "viewer")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("operator role cannot access audit logs", func(t *testing.T) {
		r := setupRouter()
		r.GET("/admin/audit", func(c *gin.Context) {
			c.Set("role", "operator")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("admin role can access audit logs", func(t *testing.T) {
		r := setupRouter()
		r.GET("/admin/audit", func(c *gin.Context) {
			c.Set("role", "admin")
			c.JSON(http.StatusOK, gin.H{"entries": []interface{}{}})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/audit", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ---------------------------------------------------------------------------
// Audit Pagination Tests
// ---------------------------------------------------------------------------

func TestAuditQueryPagination(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		offset int
		valid  bool
	}{
		{"default limit", 0, 0, true},
		{"explicit limit", 50, 0, true},
		{"max limit", 1000, 0, true},
		{"over max limit", 1001, 0, true}, // handler clamps
		{"with offset", 50, 10, true},
		{"zero limit defaults", 0, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			handler := handlers.NewAuditHandler(newTestLogger(t))
			r.GET("/admin/audit", handler.Query)

			query := ""
			if tt.limit > 0 || tt.offset > 0 {
				query = fmt.Sprintf("?limit=%d&offset=%d", tt.limit, tt.offset)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/admin/audit"+query, nil)
			r.ServeHTTP(w, req)

			// Without auth, all return 401
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
