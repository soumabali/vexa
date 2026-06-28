# Monitoring — vexa (P4 #5)

Prometheus + Grafana stack, gated behind a `monitoring` compose profile so
the base stack stays light.

## Run

```bash
# Start API + web + databases + monitoring
docker compose --profile monitoring up -d

# Stop just the monitoring side
docker compose --profile monitoring stop
```

| Service     | Port  | URL                                |
| ----------- | ----- | ---------------------------------- |
| Prometheus  | 9090  | http://localhost:9090              |
| Grafana     | 3001  | http://localhost:3001 (admin/admin) |

Grafana is provisioned with Prometheus as the default datasource.

## Metrics

The API exposes Prometheus text format at `GET /metrics` (P4 #5):

| Metric                                     | Type      | Labels                  |
| ------------------------------------------ | --------- | ----------------------- |
| `vexa_http_requests_total`                 | counter   | method, path, status    |
| `vexa_http_request_duration_seconds`       | histogram | method, path            |
| `vexa_uptime_seconds`                      | gauge     | —                       |
| `vexa_db_pool_active_connections`          | gauge     | —                       |
| `vexa_redis_pool_active_connections`       | gauge     | —                       |
| (standard Go runtime + process collectors) | mixed     | —                       |

`/metrics` is **not** public: it is bound to the API service inside the
compose network. Do not expose it via Traefik without auth — request
counts leak business signals.

## CORS

When running Grafana locally, add `http://localhost:3001` to
`ALLOWED_ORIGINS` in `.env` (already done in `.env.example`).

## Dashboard recommendations

A minimal but useful Grafana dashboard for vexa:

- Request rate (`rate(vexa_http_requests_total[1m])`) per status class
- 95th / 99th percentile latency
  (`histogram_quantile(0.95, rate(vexa_http_request_duration_seconds_bucket[5m]))`)
- DB / Redis pool gauges
- API uptime
- Per-route error ratio
  (`sum by (path) (rate(vexa_http_requests_total{status="5xx"}[5m]))`)

## Alert rules (minimum)

Save under `docker/prometheus-alerts.yml` and reference it from
`prometheus.yml` (`rule_files:`).

```yaml
groups:
  - name: vexa
    interval: 30s
    rules:
      - alert: VexaAPI5xxRate
        expr: sum(rate(vexa_http_requests_total{status="5xx"}[5m])) > 1
        for: 5m
        annotations:
          summary: "vexa API returning >1 5xx/sec for 5m"

      - alert: VexaAPIDown
        expr: up{job="vexa-api"} == 0
        for: 1m
        annotations:
          summary: "vexa API scrape target down"

      - alert: VexaHighLatency
        expr: |
          histogram_quantile(
            0.95,
            rate(vexa_http_request_duration_seconds_bucket[5m])
          ) > 1
        for: 10m
        annotations:
          summary: "p95 latency above 1s"
```

## Local verification

```bash
# Bring everything up
docker compose --profile monitoring up -d

# Verify scrape
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job:.labels.job, health}'

# Tail a series
curl -s http://localhost:9090/api/v1/query?query=vexa_uptime_seconds | jq .
```