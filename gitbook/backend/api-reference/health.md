# ❤️ Health & Metrics

Health check and monitoring endpoints. These endpoints do not require authentication.

---

## Health Check

Basic health check to verify the API server is running.

```http
GET /health
```

### Response `200 OK`

```json
{
  "status": "ok"
}
```

---

## Readiness Check

Readiness check to verify all dependencies (database, Redis) are connected and ready.

```http
GET /ready
```

### Response `200 OK`

```json
{
  "status": "ready",
  "database": "connected",
  "redis": "connected"
}
```

### Response `503 Service Unavailable`

```json
{
  "status": "not_ready",
  "database": "connected",
  "redis": "disconnected"
}
```

{% hint style="info" %}
Use `/health` for liveness probes and `/ready` for readiness probes in Kubernetes or Docker health checks.
{% endhint %}

---

## Prometheus Metrics

Exposes metrics in Prometheus text format for scraping.

```http
GET /metrics
```

### Response `200 OK`

```
# HELP aureo_vpn_active_connections Current number of active VPN connections
# TYPE aureo_vpn_active_connections gauge
aureo_vpn_active_connections 38000

# HELP aureo_vpn_node_load_score Current load score of VPN nodes
# TYPE aureo_vpn_node_load_score gauge
aureo_vpn_node_load_score{node="us-east-1"} 25.5

# HELP aureo_vpn_data_transferred_bytes Total data transferred
# TYPE aureo_vpn_data_transferred_bytes counter
aureo_vpn_data_transferred_bytes 1.234567890e+12

# HELP aureo_vpn_connection_duration_seconds Duration of VPN connections
# TYPE aureo_vpn_connection_duration_seconds histogram
aureo_vpn_connection_duration_seconds_bucket{le="60"} 1500
aureo_vpn_connection_duration_seconds_bucket{le="300"} 5000
aureo_vpn_connection_duration_seconds_bucket{le="3600"} 25000
aureo_vpn_connection_duration_seconds_bucket{le="+Inf"} 38000

# HELP aureo_vpn_http_requests_total Total HTTP requests
# TYPE aureo_vpn_http_requests_total counter
aureo_vpn_http_requests_total{method="GET",path="/api/v1/nodes",status="200"} 150000
aureo_vpn_http_requests_total{method="POST",path="/api/v1/sessions",status="201"} 45000
```

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `aureo_vpn_active_connections` | Gauge | Current active VPN connections |
| `aureo_vpn_node_load_score` | Gauge | Load score per node (0-100) |
| `aureo_vpn_data_transferred_bytes` | Counter | Total bytes transferred |
| `aureo_vpn_connection_duration_seconds` | Histogram | VPN connection durations |
| `aureo_vpn_http_requests_total` | Counter | HTTP requests by method, path, status |

---

## Docker Health Check

Example Docker Compose health check configuration:

```yaml
services:
  api-gateway:
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

---

## Grafana Dashboard

When Prometheus and Grafana are configured, metrics are available at:

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (default credentials: `admin`/`admin`)
