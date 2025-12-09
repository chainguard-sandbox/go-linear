package linear

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linear_requests_total",
			Help: "Total number of Linear API requests",
		},
		[]string{"operation", "status_code"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "linear_request_duration_seconds",
			Help:    "Linear API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linear_errors_total",
			Help: "Total number of Linear API errors",
		},
		[]string{"operation", "error_type"},
	)

	retriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "linear_retries_total",
			Help: "Total number of retry attempts",
		},
		[]string{"reason"},
	)

	rateLimitRemaining = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "linear_rate_limit_remaining",
			Help: "Linear API rate limit remaining capacity",
		},
		[]string{"limit_type"},
	)
)

// EnableMetrics enables Prometheus metrics collection.
// Metrics are automatically registered with the default Prometheus registry
// via promauto and can be exposed at /metrics using promhttp.Handler().
//
// Call this before creating clients to ensure metrics are registered.
//
// Example:
//
//	linear.EnableMetrics()
//	client, _ := linear.NewClient(apiKey, linear.WithMetrics())
//
// See examples/prometheus/main.go for complete integration.
func EnableMetrics() {
	// Metrics are registered automatically via promauto.
	// This function exists for explicit opt-in and documentation.
}

// recordRequest records a completed HTTP request.
func recordRequest(operation string, statusCode int, duration time.Duration) {
	requestsTotal.WithLabelValues(operation, fmt.Sprintf("%d", statusCode)).Inc()
	requestDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// recordError records an error by type.
func recordError(operation string, errorType string) {
	errorsTotal.WithLabelValues(operation, errorType).Inc()
}

// recordRetry records a retry attempt.
func recordRetry(reason string) {
	retriesTotal.WithLabelValues(reason).Inc()
}

// recordRateLimit updates rate limit gauges.
func recordRateLimit(info *RateLimitInfo) {
	if info != nil {
		rateLimitRemaining.WithLabelValues("requests").Set(float64(info.RequestsRemaining))
		rateLimitRemaining.WithLabelValues("complexity").Set(float64(info.ComplexityRemaining))
	}
}
