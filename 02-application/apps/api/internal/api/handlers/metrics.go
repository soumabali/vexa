package handlers

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler serves /metrics in Prometheus text format.
//
// The exporter collects standard process / Go runtime metrics plus the
// application-specific series listed in P4 #5:
//
//   - vexa_http_requests_total{method,path,status}      counter
//   - vexa_http_request_duration_seconds{method,path}    histogram
//   - vexa_uptime_seconds                                gauge
//
// DB / Redis pool gauges are registered by the middleware that owns the
// connection pools (see RegisterPoolGauges).
type MetricsHandler struct {
	registry  *prometheus.Registry
	startTime time.Time

	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	uptime       prometheus.Gauge

	dbPool    prometheus.Gauge
	redisPool prometheus.Gauge

	poolGaugesRegistered atomic.Bool
}

// NewMetricsHandler constructs the handler. It does NOT register on the
// default Prometheus registry — the API keeps its own isolated registry
// so test runs and multi-instance setups don't collide on /metrics.
func NewMetricsHandler() *MetricsHandler {
	reg := prometheus.NewRegistry()
	start := time.Now()

	httpReq := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vexa_http_requests_total",
			Help: "Total HTTP requests handled, partitioned by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)
	httpDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "vexa_http_request_duration_seconds",
			Help:    "HTTP request handler duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	uptime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vexa_uptime_seconds",
		Help: "Seconds since the API process started.",
	})
	dbPool := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vexa_db_pool_active_connections",
		Help: "Active (in-use) connections in the primary database pool.",
	})
	redisPool := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vexa_redis_pool_active_connections",
		Help: "Active connections in the primary Redis pool.",
	})

	reg.MustRegister(httpReq, httpDur, uptime, dbPool, redisPool)
	// Standard process + Go collectors — useful for debugging, low cost.
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(prometheus.NewGoCollector())

	return &MetricsHandler{
		registry:     reg,
		startTime:    start,
		httpRequests: httpReq,
		httpDuration: httpDur,
		uptime:       uptime,
		dbPool:       dbPool,
		redisPool:    redisPool,
	}
}

// Handler returns a gin.HandlerFunc that exposes /metrics in Prometheus
// text format. Call this from the router.
func (h *MetricsHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.uptime.Set(time.Since(h.startTime).Seconds())
		promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}).ServeHTTP(c.Writer, c.Request)
	}
}

// PrometheusMiddleware records per-request metrics. Mount it before the
// other handlers so path normalization (c.FullPath) returns the route
// template rather than the raw URL — this prevents label cardinality
// explosions from path params.
//
// Usage:
//
//	r.Use(metricsHandler.PrometheusMiddleware())
func (h *MetricsHandler) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		status := statusToString(c.Writer.Status())
		h.httpRequests.WithLabelValues(c.Request.Method, path, status).Inc()
		h.httpDuration.WithLabelValues(c.Request.Method, path).
			Observe(time.Since(start).Seconds())
	}
}

// SetDBPoolActive updates the DB pool gauge. Wire from the place that
// owns the *sql.DB (typically main, where we already know the pool size).
func (h *MetricsHandler) SetDBPoolActive(n int) {
	h.dbPool.Set(float64(n))
}

// SetRedisPoolActive updates the Redis pool gauge. Wire from the place
// that owns the redis.Client (typically main).
func (h *MetricsHandler) SetRedisPoolActive(n int) {
	h.redisPool.Set(float64(n))
}

func statusToString(code int) string {
	switch {
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

// EnsureResponse makes sure the handler returns 200 even when no metrics
// have been recorded yet (helps liveness probes from monitoring systems
// that poll /metrics).
func (h *MetricsHandler) EnsureResponse(c *gin.Context) {
	if c.Writer.Status() == http.StatusOK && c.Writer.Size() == 0 {
		h.Handler()(c)
		return
	}
	c.Status(http.StatusOK)
}