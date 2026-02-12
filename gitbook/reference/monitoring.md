# 📊 Monitoring & Metrics

Aureo VPN exposes Prometheus metrics for monitoring all aspects of the system. Metrics are registered via `promauto` (auto-registration) and served at the `/metrics` endpoint.

Source: `pkg/metrics/metrics.go`

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `METRICS_ENABLED` | `true` | Enable/disable metrics collection |
| `METRICS_PORT` | `9090` | Metrics server port |
| `METRICS_PATH` | `/metrics` | Prometheus scrape endpoint |

---

## Prometheus Metrics Catalog

### HTTP Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_http_requests_total` | Counter | `method`, `path`, `status` | Total number of HTTP requests |
| `aureo_vpn_http_request_duration_seconds` | Histogram | `method`, `path` | HTTP request duration. Buckets: default Prometheus buckets (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10) |

### VPN Connection Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_active_connections` | Gauge | `protocol`, `node` | Number of currently active VPN connections |
| `aureo_vpn_connections_total` | Counter | `protocol`, `node`, `status` | Total VPN connections (status: `success`, `failed`) |
| `aureo_vpn_data_transferred_bytes` | Counter | `direction`, `protocol`, `node` | Total data transferred (direction: `sent`, `received`) |
| `aureo_vpn_connection_duration_seconds` | Histogram | `protocol`, `node` | VPN connection duration. Buckets: 60s, 5m, 10m, 30m, 1h, 2h, 4h, 8h, 24h |

### Node Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_node_status` | Gauge | `node`, `country`, `city` | Node status (1 = online, 0 = offline) |
| `aureo_vpn_node_load_score` | Gauge | `node` | Node load score (0-100, lower is better) |
| `aureo_vpn_node_cpu_usage_percent` | Gauge | `node` | Node CPU usage percentage |
| `aureo_vpn_node_memory_usage_percent` | Gauge | `node` | Node memory usage percentage |
| `aureo_vpn_node_bandwidth_gbps` | Gauge | `node` | Node bandwidth usage in Gbps |

### Authentication Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_login_attempts_total` | Counter | `status` | Total login attempts (status: `success`, `failed`) |
| `aureo_vpn_user_registrations_total` | Counter | (none) | Total user registrations |
| `aureo_vpn_token_generations_total` | Counter | `type` | Total token generations (type: `access`, `refresh`) |

### User Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_active_users` | Gauge | (none) | Number of currently active users |

### P2P Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_p2p_connected_peers` | Gauge | (none) | Number of connected P2P peers |
| `aureo_p2p_known_nodes` | Gauge | (none) | Number of known VPN nodes in the P2P network |

### VPN Traffic Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_bytes_received_total` | Counter | (none) | Total bytes received by VPN (aggregate) |
| `aureo_vpn_bytes_sent_total` | Counter | (none) | Total bytes sent by VPN (aggregate) |

### Database Metrics

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `aureo_vpn_database_queries_total` | Counter | `operation`, `table` | Total database queries (operation: `select`, `insert`, `update`, `delete`) |
| `aureo_vpn_database_query_duration_seconds` | Histogram | `operation`, `table` | Database query duration. Buckets: default Prometheus buckets |

---

## Metrics Collection

### HTTP Metrics Middleware

The `RecordHTTPMetrics()` middleware wraps every request:

```go
func RecordHTTPMetrics() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := prometheus.NewTimer(
            HTTPRequestDuration.WithLabelValues(c.Method(), c.Path()),
        )
        defer start.ObserveDuration()

        err := c.Next()

        HTTPRequestsTotal.WithLabelValues(
            c.Method(), c.Path(), statusCode,
        ).Inc()

        return err
    }
}
```

### Node Metrics Collection

The `metricsCollector` goroutine runs every 15 seconds on each VPN node:

```go
func (s *Service) collectMetrics() {
    // Recalculate load score
    loadScore := node.CalculateLoadScore()

    // Update Prometheus gauges
    metrics.NodeStatus.WithLabelValues(node.Name, node.Country, node.City).Set(status)
    metrics.NodeLoad.WithLabelValues(node.Name).Set(loadScore)
    metrics.NodeCPUUsage.WithLabelValues(node.Name).Set(node.CPUUsage)
    metrics.NodeMemoryUsage.WithLabelValues(node.Name).Set(node.MemoryUsage)
    metrics.NodeBandwidth.WithLabelValues(node.Name).Set(node.BandwidthUsageGbps)
}
```

### Connection Metrics

Updated in the node service when sessions are created/disconnected:

```go
// On session creation
metrics.ActiveConnections.WithLabelValues(protocol, node.Name).Inc()
metrics.ConnectionsTotal.WithLabelValues(protocol, node.Name, "success").Inc()

// On session disconnect
metrics.ActiveConnections.WithLabelValues(protocol, node.Name).Dec()
```

### Auth Metrics

Updated in API handlers:

```go
// On login success
metrics.LoginAttempts.WithLabelValues("success").Inc()

// On login failure
metrics.LoginAttempts.WithLabelValues("failed").Inc()

// On registration
metrics.UserRegistrations.Inc()

// On token refresh
metrics.TokenGenerations.WithLabelValues("access").Inc()
```

---

## Prometheus Endpoint

The metrics endpoint is served via a FastHTTP adapter:

```go
func PrometheusHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
        handler(c.Context())
        return nil
    }
}
```

Scrape configuration for Prometheus:

```yaml
scrape_configs:
  - job_name: 'aureo-vpn'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

---

## Key Dashboards

Recommended Grafana panels:

| Panel | Query |
|---|---|
| Active Connections | `aureo_vpn_active_connections` |
| Request Rate | `rate(aureo_vpn_http_requests_total[5m])` |
| Request Latency P99 | `histogram_quantile(0.99, rate(aureo_vpn_http_request_duration_seconds_bucket[5m]))` |
| Node Load | `aureo_vpn_node_load_score` |
| Login Success Rate | `rate(aureo_vpn_login_attempts_total{status="success"}[5m]) / rate(aureo_vpn_login_attempts_total[5m])` |
| Bandwidth per Node | `aureo_vpn_node_bandwidth_gbps` |
| P2P Peers | `aureo_p2p_connected_peers` |
| DB Query Latency | `histogram_quantile(0.95, rate(aureo_vpn_database_query_duration_seconds_bucket[5m]))` |
