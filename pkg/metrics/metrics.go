package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aureo_vpn_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// VPN Connection metrics
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_active_connections",
			Help: "Number of active VPN connections",
		},
		[]string{"protocol", "node"},
	)

	ConnectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_connections_total",
			Help: "Total number of VPN connections",
		},
		[]string{"protocol", "node", "status"},
	)

	DataTransferred = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_data_transferred_bytes",
			Help: "Total data transferred in bytes",
		},
		[]string{"direction", "protocol", "node"},
	)

	ConnectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aureo_vpn_connection_duration_seconds",
			Help:    "VPN connection duration in seconds",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"protocol", "node"},
	)

	// Node metrics
	NodeStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_node_status",
			Help: "VPN node status (1 = online, 0 = offline)",
		},
		[]string{"node", "country", "city"},
	)

	NodeLoad = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_node_load_score",
			Help: "VPN node load score (0-100)",
		},
		[]string{"node"},
	)

	NodeCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_node_cpu_usage_percent",
			Help: "VPN node CPU usage percentage",
		},
		[]string{"node"},
	)

	NodeMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_node_memory_usage_percent",
			Help: "VPN node memory usage percentage",
		},
		[]string{"node"},
	)

	NodeBandwidth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_node_bandwidth_gbps",
			Help: "VPN node bandwidth usage in Gbps",
		},
		[]string{"node"},
	)

	// User metrics
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "aureo_vpn_active_users",
			Help: "Number of active users",
		},
	)

	UserRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "aureo_vpn_user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	// Authentication metrics
	LoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"},
	)

	TokenGenerations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_token_generations_total",
			Help: "Total number of token generations",
		},
		[]string{"type"},
	)

	// Database metrics
	DatabaseQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aureo_vpn_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aureo_vpn_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)
)

// PrometheusHandler returns a Fiber handler for Prometheus metrics
func PrometheusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		handler(c.Context())
		return nil
	}
}

// RecordHTTPMetrics middleware records HTTP metrics
func RecordHTTPMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := prometheus.NewTimer(HTTPRequestDuration.WithLabelValues(c.Method(), c.Path()))
		defer start.ObserveDuration()

		err := c.Next()

		status := c.Response().StatusCode()
		HTTPRequestsTotal.WithLabelValues(c.Method(), c.Path(), string(rune(status))).Inc()

		return err
	}
}
