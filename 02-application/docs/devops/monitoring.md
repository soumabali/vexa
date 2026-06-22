# vexa — Monitoring & Observability

> **Agent:** DevOps Engineer  > **Status:** Generated 2026-05-28  
> **Version:** 0.1.0

---

## 1. Observability Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Metrics** | Prometheus + Grafana | Performance metrics, dashboards |
| **Logs** | Loki + Grafana | Centralized log aggregation |
| **Tracing** | Jaeger / Tempo | Distributed request tracing |
| **Alerts** | Alertmanager | Alert routing, silencing |
| **Uptime** | Uptime Kuma / Statuspage | External health checks |

---

## 2. Metrics

### 2.1 Application Metrics (Prometheus)

```go
// internal/metrics/metrics.go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "handler", "status"},
    )
    
    ActiveSessions = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "ssh_manager_active_sessions",
            Help: "Number of active SSH/RDP/VNC sessions",
        },
    )
    
    ConnectionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ssh_manager_connections_total",
            Help: "Total connections by protocol",
        },
        []string{"protocol"},
    )
)

func init() {
    prometheus.MustRegister(RequestDuration, ActiveSessions, ConnectionsTotal)
}
```

### 2.2 System Metrics

```yaml
# Node Exporter for system metrics
- job_name: 'node'
  static_configs:
  - targets: ['node-exporter:9100']

# cAdvisor for container metrics
- job_name: 'cadvisor'
  static_configs:
  - targets: ['cadvisor:8080']
```

### 2.3 Custom Business Metrics

```go
// Business metrics
vault_items_total = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "ssh_manager_vault_items_total",
        Help: "Total credentials in vault",
    },
)

users_total = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "ssh_manager_users_total",
        Help: "Total users by role",
    },
    []string{"role"},
)
```

---

## 3. Logging

### 3.1 Structured Logging (Go)

```go
// internal/middleware/logging.go
package middleware

import (
    "log/slog"
    "time"
)

type StructuredLogger struct {
    logger *slog.Logger
}

func (l *StructuredLogger) LogRequest(c *gin.Context, duration time.Duration) {
    l.logger.Info("HTTP request",
        slog.String("method", c.Request.Method),
        slog.String("path", c.Request.URL.Path),
        slog.Int("status", c.Writer.Status()),
        slog.Duration("duration", duration),
        slog.String("ip", c.ClientIP()),
        slog.String("user_agent", c.Request.UserAgent()),
        slog.String("request_id", c.GetString("request_id")),
    )
}
```

### 3.2 Log Levels

| Level | Usage |
|-------|-------|
| `DEBUG` | Development, detailed tracing |
| `INFO` | Normal operations, requests |
| `WARN` | Slow queries, rate limit hits |
| `ERROR` | Failed requests, exceptions |
| `FATAL` | System crashes, data corruption |

### 3.3 Log Aggregation (Loki)

```yaml
# Promtail configuration
scrape_configs:
  - job_name: vexa-api
    static_configs:
    - targets:
        - localhost
      labels:
        job: vexa-api
        __path__: /var/log/vexa/api.log
    pipeline_stages:
    - json:
        expressions:
          level: level
          msg: msg
          method: method
          path: path
          status: status
    - labels:
        level:
        method:
        status:
```

---

## 4. Distributed Tracing

### 4.1 OpenTelemetry Setup

```go
// internal/tracing/tracing.go
package tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() (*sdktrace.TracerProvider, error) {
    exporter, err := jaeger.New(jaeger.WithAgentEndpoint())
    if err != nil {
        return nil, err
    }
    
    provider := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )
    
    otel.SetTracerProvider(provider)
    return provider, nil
}
```

### 4.2 Trace Spans

```go
// In handlers
ctx, span := tracer.Start(c.Request.Context(), "api.auth.login")
defer span.End()

span.SetAttributes(
    attribute.String("user.email", email),
    attribute.Bool("mfa.required", mfaRequired),
)

// Nested spans
_, dbSpan := tracer.Start(ctx, "db.user.find")
user, err := db.FindUser(email)
dbSpan.End()
```

---

## 5. Alerting

### 5.1 Prometheus Alert Rules

```yaml
# infra/monitoring/alerts.yml
groups:
- name: vexa-critical
  rules:
  - alert: APIDown
    expr: up{job="vexa-api"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "API is down"
      description: "API instance {{ $labels.instance }} has been down for 1 minute"
  
  - alert: DatabaseDown
    expr: pg_up{job="postgres"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database is down"
  
  - alert: HighErrorRate
    expr: |
      (
        sum(rate(http_requests_total{status=~"5.."}[5m]))
        /
        sum(rate(http_requests_total[5m]))
      ) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High error rate: {{ $value | humanizePercentage }}"
  
  - alert: SlowQueries
    expr: |
      histogram_quantile(0.99,
        sum(rate(http_request_duration_seconds_bucket[5m])) by (le)
      ) > 2
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "P99 latency > 2s"
  
  - alert: DiskSpaceLow
    expr: |
      (
        node_filesystem_avail_bytes{mountpoint="/"}
        /
        node_filesystem_size_bytes{mountpoint="/"}
      ) < 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Disk space < 10%"
  
  - alert: MemoryHigh
    expr: |
      (
        node_memory_MemAvailable_bytes
        /
        node_memory_MemTotal_bytes
      ) < 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Memory available < 10%"

- name: vexa-security
  rules:
  - alert: FailedLoginAttempts
    expr: |
      sum(rate(ssh_manager_failed_logins_total[5m])) by (ip) > 10
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "High failed login attempts from {{ $labels.ip }}"
  
  - alert: SuspiciousActivity
    expr: |
      ssh_manager_audit_logs_total{action="connect", success="false"} > 50
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Suspicious connection activity detected"
```

### 5.2 Alertmanager Configuration

```yaml
# alertmanager.yml
global:
  slack_api_url: '${SLACK_WEBHOOK_URL}'
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@vexa.local'

route:
  receiver: 'default'
  routes:
  - match:
      severity: critical
    receiver: 'pagerduty-critical'
    continue: true
  - match:
      severity: warning
    receiver: 'slack-warnings'
    continue: true

receivers:
- name: 'default'
  slack_configs:
  - channel: '#alerts'
    title: '{{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

- name: 'pagerduty-critical'
  pagerduty_configs:
  - service_key: '${PAGERDUTY_KEY}'
    severity: critical

- name: 'slack-warnings'
  slack_configs:
  - channel: '#warnings'
    title: '{{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
    send_resolved: true

inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  equal: ['alertname', 'instance']
```

---

## 6. Dashboards

### 6.1 Grafana Dashboard (JSON Model)

```json
{
  "dashboard": {
    "title": "vexa - Overview",
    "tags": ["vexa"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{ method }} {{ handler }}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Error Rate",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "Errors"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Active Sessions",
        "type": "stat",
        "targets": [
          {
            "expr": "ssh_manager_active_sessions",
            "legendFormat": "Sessions"
          }
        ],
        "gridPos": {"h": 4, "w": 6, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "Connections by Protocol",
        "type": "piechart",
        "targets": [
          {
            "expr": "ssh_manager_connections_total",
            "legendFormat": "{{ protocol }}"
          }
        ],
        "gridPos": {"h": 4, "w": 6, "x": 6, "y": 8}
      },
      {
        "id": 5,
        "title": "P95 Latency",
        "type": "timeseries",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{ handler }}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 12}
      },
      {
        "id": 6,
        "title": "Database Connections",
        "type": "timeseries",
        "targets": [
          {
            "expr": "pg_stat_activity_count",
            "legendFormat": "Connections"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 12}
      }
    ]
  }
}
```

### 6.2 Security Dashboard

```json
{
  "dashboard": {
    "title": "vexa - Security",
    "panels": [
      {
        "title": "Failed Login Attempts",
        "type": "timeseries",
        "targets": [
          {
            "expr": "sum(rate(ssh_manager_failed_logins_total[5m])) by (ip)",
            "legendFormat": "{{ ip }}"
          }
        ]
      },
      {
        "title": "Audit Events",
        "type": "table",
        "targets": [
          {
            "expr": "ssh_manager_audit_logs_total",
            "format": "table"
          }
        ]
      },
      {
        "title": "Vault Operations",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(ssh_manager_vault_operations_total[5m])",
            "legendFormat": "{{ operation }}"
          }
        ]
      }
    ]
  }
}
```

---

## 7. Health Endpoints

### 7.1 Liveness Probe

```go
// GET /health/live
func LivenessHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
```

### 7.2 Readiness Probe

```go
// GET /health/ready
func ReadinessHandler(c *gin.Context) {
    checks := map[string]bool{
        "database": db.Ping() == nil,
        "redis":    redisClient.Ping(ctx).Err() == nil,
    }
    
    for _, ok := range checks {
        if !ok {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "status": "not ready",
                "checks": checks,
            })
            return
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
        "checks": checks,
    })
}
```

### 7.3 Startup Probe

```go
// GET /health/startup
func StartupHandler(c *gin.Context) {
    // Check all dependencies are available
    if !allDependenciesReady() {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "starting"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "started"})
}
```

---

## 8. SLOs / SLIs

| SLO | Target | Measurement |
|-----|--------|-------------|
| **Availability** | 99.9% | Uptime over 30 days |
| **Latency (P95)** | < 500ms | API response time |
| **Latency (P99)** | < 2s | API response time |
| **Error Rate** | < 0.1% | 5xx errors / total requests |
| **Database Latency** | < 50ms | Query execution time |
| **Session Establishment** | < 3s | SSH/RDP/VNC connection time |

---

## 9. Runbooks

### 9.1 API High Latency

```markdown
## Symptoms
- P95 latency > 500ms
- P99 latency > 2s

## Diagnosis
1. Check database: `pg_stat_activity` for slow queries
2. Check Redis: connection pool saturation
3. Check goroutine leaks: `go tool pprof`

## Resolution
1. Scale API pods: `kubectl scale deployment api --replicas=5`
2. Restart affected pods: `kubectl rollout restart deployment/api`
3. Check for N+1 queries in recent deployments
```

### 9.2 Database Connection Pool Exhaustion

```markdown
## Symptoms
- "too many connections" errors
- Requests timing out

## Diagnosis
1. `SELECT count(*) FROM pg_stat_activity;`
2. `SELECT * FROM pg_stat_activity WHERE state = 'active';`

## Resolution
1. Increase pool size: `DB_MAX_CONNECTIONS=50`
2. Kill idle connections: `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle';`
3. Add connection pooler (PgBouncer)
```

### 9.3 High Memory Usage

```markdown
## Symptoms
- OOMKilled pods
- Memory usage > 80%

## Diagnosis
1. `kubectl top pods`
2. Check for memory leaks in pprof heap profile
3. Check session cache growth

## Resolution
1. Increase memory limit: `kubectl set resources`
2. Reduce cache TTL
3. Profile and fix memory leak
```

---

## 10. External Monitoring

### 10.1 Uptime Kuma

```yaml
# docker-compose.monitoring.yml
version: '3.8'
services:
  uptime-kuma:
    image: louislam/uptime-kuma:latest
    volumes:
      - uptime-data:/app/data
    ports:
      - "3001:3001"

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./infra/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana-data:/var/lib/grafana
      - ./infra/monitoring/dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"

  loki:
    image: grafana/loki:latest
    volumes:
      - ./infra/monitoring/loki.yml:/etc/loki/local-config.yaml
      - loki-data:/loki
    ports:
      - "3100:3100"

volumes:
  uptime-data:
  prometheus-data:
  grafana-data:
  loki-data:
```

### 10.2 Blackbox Exporter

```yaml
# blackbox.yml
modules:
  http_2xx:
    prober: http
    timeout: 5s
    http:
      valid_http_versions: ["HTTP/1.1", "HTTP/2.0"]
      valid_status_codes: [200, 301, 302]
      method: GET
      fail_if_ssl: false
```
