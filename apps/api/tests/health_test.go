package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/health"
)

func TestTCPHealthChecker(t *testing.T) {
	checker := health.NewTCPHealthChecker().
		WithTimeout(5 * time.Second).
		WithRetries(2)

	t.Run("check_localhost_ssh", func(t *testing.T) {
		// This test requires an SSH server running locally
		// Skip if not available
		result, err := checker.Check(context.Background(), "127.0.0.1", 22)
		if err != nil {
			t.Skip("No SSH server available locally")
		}
		assert.Equal(t, health.StatusHealthy, result.Status)
		assert.Greater(t, result.LatencyMs(), float64(0))
		assert.True(t, result.IsHealthy())
	})

	t.Run("check_unreachable_host", func(t *testing.T) {
		// Test with a port that's likely closed
		result, err := checker.Check(context.Background(), "127.0.0.1", 65432)
		assert.Error(t, err)
		assert.Equal(t, health.StatusUnhealthy, result.Status)
		assert.False(t, result.IsHealthy())
		assert.GreaterOrEqual(t, result.Attempt, 1)
	})
}

func TestProbeEngine(t *testing.T) {
	engine := health.NewProbeEngine()
	require.NotNil(t, engine)

	t.Run("register_and_get", func(t *testing.T) {
		config := health.DefaultProbeConfig()
		config.Port = 22
		config.Interval = 1 * time.Hour // Long interval so it doesn't run immediately

		engine.Register("host-1", "Test Host", "127.0.0.1", config)

		h, ok := engine.GetHealth("host-1")
		assert.True(t, ok)
		assert.Equal(t, "host-1", h.HostID)
		assert.Equal(t, "Test Host", h.HostName)
		assert.Equal(t, "127.0.0.1", h.Address)
	})

	t.Run("unregister", func(t *testing.T) {
		engine.Unregister("host-1")
		_, ok := engine.GetHealth("host-1")
		assert.False(t, ok)
	})

	t.Run("get_all_health", func(t *testing.T) {
		// Register multiple hosts
		config := health.DefaultProbeConfig()
		engine.Register("host-a", "Host A", "10.0.0.1", config)
		engine.Register("host-b", "Host B", "10.0.0.2", config)

		all := engine.GetAllHealth()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "host-a")
		assert.Contains(t, all, "host-b")
	})
}

func TestBulkCheck(t *testing.T) {
	engine := health.NewProbeEngine()

	t.Run("bulk_check_max_100", func(t *testing.T) {
		hosts := make([]health.HostCheckRequest, 105)
		for i := 0; i < 105; i++ {
			hosts[i] = health.HostCheckRequest{
				HostID:  fmt.Sprintf("host-%d", i),
				Address: "127.0.0.1",
				Port:    22,
			}
		}

		// Should still work but cap at 100
		results := engine.BulkCheck(hosts, 10)
		assert.Len(t, results, 105)
	})

	t.Run("bulk_check_parallel", func(t *testing.T) {
		hosts := []health.HostCheckRequest{
			{HostID: "h1", Address: "127.0.0.1", Port: 22},
			{HostID: "h2", Address: "127.0.0.1", Port: 65432}, // Likely closed
		}

		results := engine.BulkCheck(hosts, 5)
		assert.Len(t, results, 2)
	})
}

func TestProbeConfig(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := health.DefaultProbeConfig()
		assert.Equal(t, health.ProbeTypeTCP, config.Type)
		assert.Equal(t, 10*time.Second, config.Timeout)
		assert.Equal(t, 3, config.Retries)
		assert.Equal(t, 5*time.Minute, config.Interval)
	})

	t.Run("custom_config", func(t *testing.T) {
		config := health.ProbeConfig{
			Type:     health.ProbeTypeTCP,
			Timeout:  30 * time.Second,
			Retries:  5,
			Interval: 1 * time.Minute,
		}
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 5, config.Retries)
	})
}

func TestICMPHealthChecker(t *testing.T) {
	checker := health.NewICMPHealthChecker().
		WithTimeout(10 * time.Second).
		WithCount(2)

	t.Run("ping_localhost", func(t *testing.T) {
		result, err := checker.Check(context.Background(), "127.0.0.1")
		if err != nil {
			t.Skip("ICMP requires root privileges")
		}
		assert.Equal(t, health.StatusHealthy, result.Status)
		assert.Greater(t, result.LatencyMs(), float64(0))
		assert.LessOrEqual(t, result.PacketLoss, float64(100))
	})

	t.Run("ping_unreachable", func(t *testing.T) {
		// Use a non-routable address
		result, _ := checker.Check(context.Background(), "192.0.2.1") // TEST-NET-1
		// Will likely fail or timeout
		assert.NotNil(t, result)
	})
}

// ----- /health/live & /health/ready handler tests (P4 #4) -----

// We need a real *sql.DB because PingContext dereferences internals even
// when called on a zero value. Open a sqlite memory DB? vexa only ships
// postgres drivers; instead, build a real DB via the standard sql package
// against an unreachable driver — but PingContext will still segfault on
// the nil mutex inside (*DB).conn.
//
// Workaround: skip the test if we can't open any DB. Use sql.Open with
// the postgres driver and a dummy DSN; PingContext will return an error
// rather than panic, which is fine for shape assertions.
func newTestHealthHandler(t *testing.T) *handlers.HealthHandler {
	t.Helper()
	db, err := sql.Open("postgres", "host=127.0.0.1 port=1 user=u dbname=d sslmode=disable connect_timeout=1")
	if err != nil {
		t.Skipf("postgres driver unavailable: %v", err)
	}
	return handlers.NewHealthHandler(db, redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}), os.TempDir())
}

func TestHealthHandler_LiveAlwaysOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHealthHandler(t)
	r := gin.New()
	r.GET("/health/live", h.Live)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "alive", body["status"])
}

func TestHealthHandler_ReadyShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHealthHandler(t)
	r := gin.New()
	r.GET("/health/ready", h.Ready)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Contains(t, body, "status")
	require.Contains(t, body, "checks")
	checks, _ := body["checks"].(map[string]interface{})
	require.NotNil(t, checks)
	for _, key := range []string{"db", "redis", "config", "disk", "wg"} {
		_, ok := checks[key]
		assert.True(t, ok, "missing check key: %s", key)
	}
}

func TestHealthHandler_ReadyFailsWhenJWTMissing(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DB_NAME", "vexa")
	gin.SetMode(gin.TestMode)
	h := newTestHealthHandler(t)
	r := gin.New()
	r.GET("/health/ready", h.Ready)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	checks := body["checks"].(map[string]interface{})
	cfg := checks["config"].(map[string]interface{})
	cfgChecks := cfg["checks"].(map[string]interface{})
	assert.Equal(t, false, cfgChecks["jwt_secret"])
}

func TestHealthHandler_DiskCheckNonEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHealthHandler(t)
	r := gin.New()
	r.GET("/health/ready", h.Ready)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	checks := body["checks"].(map[string]interface{})
	disk := checks["disk"].(map[string]interface{})
	_, hasPath := disk["path"]
	_, hasPct := disk["free_pct"]
	assert.True(t, hasPath)
	assert.True(t, hasPct)
}

func TestHealthHandler_WGDisabledByDefault(t *testing.T) {
	t.Setenv("WG_CHECK_ENABLED", "")
	gin.SetMode(gin.TestMode)
	h := newTestHealthHandler(t)
	r := gin.New()
	r.GET("/health/ready", h.Ready)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	checks := body["checks"].(map[string]interface{})
	wg := checks["wg"].(map[string]interface{})
	assert.Equal(t, false, wg["enabled"])
	assert.Equal(t, true, wg["ok"])
}
