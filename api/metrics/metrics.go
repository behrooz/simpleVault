package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestDuration tracks the duration of HTTP requests
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestTotal tracks the total number of HTTP requests
	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestInFlight tracks the number of in-flight HTTP requests
	HTTPRequestInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// MongoDBOperationsDuration tracks MongoDB operation duration
	MongoDBOperationsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mongodb_operation_duration_seconds",
			Help:    "Duration of MongoDB operations in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "collection"},
	)

	// MongoDBOperationsTotal tracks the total number of MongoDB operations
	MongoDBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mongodb_operations_total",
			Help: "Total number of MongoDB operations",
		},
		[]string{"operation", "collection", "status"},
	)

	// ActiveConnections tracks the number of active connections
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)
)

// PrometheusMiddleware returns a Gin middleware that records Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint from being tracked
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		// Increment in-flight requests
		HTTPRequestInFlight.Inc()
		defer HTTPRequestInFlight.Dec()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := strconv.Itoa(c.Writer.Status())

		// Get path (normalize to avoid high cardinality)
		path := normalizePath(c.FullPath())

		// Record metrics
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		HTTPRequestTotal.WithLabelValues(c.Request.Method, path, status).Inc()
	}
}

// normalizePath normalizes the path to avoid high cardinality in metrics
// Replaces path parameters with placeholders
func normalizePath(path string) string {
	if path == "" {
		return "unknown"
	}
	// For Gin routes, the path might contain parameter names like :id
	// We'll keep the route pattern as-is since Gin provides the route pattern
	// Common patterns:
	// /api/v1/secrets/:id -> /api/v1/secrets/:id
	// This helps group metrics by route pattern rather than individual IDs
	return path
}

// RecordMongoDBOperation records metrics for MongoDB operations
func RecordMongoDBOperation(operation, collection string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	MongoDBOperationsDuration.WithLabelValues(operation, collection).Observe(duration.Seconds())
	MongoDBOperationsTotal.WithLabelValues(operation, collection, status).Inc()
}
