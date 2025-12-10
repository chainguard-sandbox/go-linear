package linear

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector holds Prometheus collectors for a client instance.
// Supports both global (shared) and instance-scoped (isolated) metrics.
type MetricsCollector struct {
	requests   *prometheus.CounterVec
	duration   *prometheus.HistogramVec
	errors     *prometheus.CounterVec
	retries    *prometheus.CounterVec
	rateLimit  *prometheus.GaugeVec
	registerer prometheus.Registerer
}

// newMetricsCollector creates a new metrics collector with the given registry.
func newMetricsCollector(reg prometheus.Registerer, suffix string) *MetricsCollector {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	// Add suffix for instance-scoped metrics to avoid collisions
	namePrefix := "linear"
	if suffix != "" {
		namePrefix = "linear_" + suffix
	}

	mc := &MetricsCollector{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: namePrefix + "_requests_total",
				Help: "Total number of Linear API requests",
			},
			[]string{"operation", "status_code"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    namePrefix + "_request_duration_seconds",
				Help:    "Linear API request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: namePrefix + "_errors_total",
				Help: "Total number of Linear API errors",
			},
			[]string{"operation", "error_type"},
		),
		retries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: namePrefix + "_retries_total",
				Help: "Total number of retry attempts",
			},
			[]string{"reason"},
		),
		rateLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: namePrefix + "_rate_limit_remaining",
				Help: "Linear API rate limit remaining capacity",
			},
			[]string{"limit_type"},
		),
		registerer: reg,
	}

	// Register metrics with the registry
	// Ignore errors - metrics may already be registered (e.g., in tests)
	_ = reg.Register(mc.requests)
	_ = reg.Register(mc.duration)
	_ = reg.Register(mc.errors)
	_ = reg.Register(mc.retries)
	_ = reg.Register(mc.rateLimit)

	return mc
}

// recordRequest records a completed HTTP request.
func (mc *MetricsCollector) recordRequest(operation string, statusCode int, duration time.Duration) {
	mc.requests.WithLabelValues(operation, fmt.Sprintf("%d", statusCode)).Inc()
	mc.duration.WithLabelValues(operation).Observe(duration.Seconds())
}

// recordError records an error by type.
func (mc *MetricsCollector) recordError(operation, errorType string) {
	mc.errors.WithLabelValues(operation, errorType).Inc()
}

// recordRetry records a retry attempt.
func (mc *MetricsCollector) recordRetry(reason string) {
	mc.retries.WithLabelValues(reason).Inc()
}

// recordRateLimit updates rate limit gauges.
func (mc *MetricsCollector) recordRateLimit(info *RateLimitInfo) {
	if info != nil {
		mc.rateLimit.WithLabelValues("requests").Set(float64(info.RequestsRemaining))
		mc.rateLimit.WithLabelValues("complexity").Set(float64(info.ComplexityRemaining))
	}
}

// Global metrics collector for backwards compatibility
var (
	defaultMetrics = newGlobalMetrics()
)

func newGlobalMetrics() *MetricsCollector {
	return &MetricsCollector{
		requests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linear_requests_total",
				Help: "Total number of Linear API requests",
			},
			[]string{"operation", "status_code"},
		),
		duration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "linear_request_duration_seconds",
				Help:    "Linear API request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		errors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linear_errors_total",
				Help: "Total number of Linear API errors",
			},
			[]string{"operation", "error_type"},
		),
		retries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "linear_retries_total",
				Help: "Total number of retry attempts",
			},
			[]string{"reason"},
		),
		rateLimit: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "linear_rate_limit_remaining",
				Help: "Linear API rate limit remaining capacity",
			},
			[]string{"limit_type"},
		),
		registerer: prometheus.DefaultRegisterer,
	}
}

// EnableMetrics enables Prometheus metrics collection.
// Metrics are automatically registered with the default Prometheus registry
// via promauto and can be exposed at /metrics using promhttp.Handler().
//
// Call this before creating clients to ensure metrics are registered.
//
// For multi-client scenarios (multiple Linear workspaces), use
// WithMetricsRegistry() to create isolated metrics per client.
//
// Example:
//
//	linear.EnableMetrics()
//	client, _ := linear.NewClient(apiKey, linear.WithMetrics())
//
// See examples/prometheus/main.go for complete integration.
func EnableMetrics() {
	// Global metrics are registered automatically via promauto in newGlobalMetrics().
	// This function exists for explicit opt-in and documentation.
}

// Package-level functions for backwards compatibility - use global collector
func recordRequest(operation string, statusCode int, duration time.Duration) {
	defaultMetrics.recordRequest(operation, statusCode, duration)
}

func recordError(operation, errorType string) {
	defaultMetrics.recordError(operation, errorType)
}

func recordRetry(reason string) {
	defaultMetrics.recordRetry(reason)
}

func recordRateLimit(info *RateLimitInfo) {
	defaultMetrics.recordRateLimit(info)
}
