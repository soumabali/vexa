package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
)

func TestMetricsHandler_ExposesPrometheusText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMetricsHandler()

	r := gin.New()
	r.Use(h.PrometheusMiddleware())
	r.GET("/metrics", h.Handler())
	r.GET("/probe", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Generate some traffic so counters/histograms have values.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Required series (presence, not exact values).
	assert.True(t, strings.Contains(body, "vexa_http_requests_total"),
		"missing vexa_http_requests_total")
	assert.True(t, strings.Contains(body, "vexa_http_request_duration_seconds"),
		"missing vexa_http_request_duration_seconds")
	assert.True(t, strings.Contains(body, "vexa_uptime_seconds"),
		"missing vexa_uptime_seconds")
	assert.True(t, strings.Contains(body, "vexa_db_pool_active_connections"),
		"missing vexa_db_pool_active_connections")
	assert.True(t, strings.Contains(body, "vexa_redis_pool_active_connections"),
		"missing vexa_redis_pool_active_connections")
}

func TestMetricsHandler_RecordsRequestsByPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMetricsHandler()

	r := gin.New()
	r.Use(h.PrometheusMiddleware())
	r.GET("/metrics", h.Handler())
	r.GET("/users/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "u")
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
		r.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// We expect the templated path, NOT the literal /users/42 — prevents
	// label cardinality explosions.
	assert.True(t, strings.Contains(body, `path="/users/:id"`),
		"metrics should label by route template, got body: %s", body)
	assert.False(t, strings.Contains(body, `path="/users/42"`),
		"metrics should NOT contain literal /users/42 (cardinality leak)")
}

func TestMetricsHandler_PoolGauges(t *testing.T) {
	h := handlers.NewMetricsHandler()
	h.SetDBPoolActive(7)
	h.SetRedisPoolActive(3)

	r := gin.New()
	r.GET("/metrics", h.Handler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "vexa_db_pool_active_connections 7"),
		"db pool gauge not set, body: %s", body)
	assert.True(t, strings.Contains(body, "vexa_redis_pool_active_connections 3"),
		"redis pool gauge not set, body: %s", body)
}