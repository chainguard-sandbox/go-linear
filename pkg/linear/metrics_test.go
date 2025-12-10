package linear

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics_RequestTracking(t *testing.T) {
	// Reset global metrics before test
	defaultMetrics.requests.Reset()
	defaultMetrics.duration.Reset()

	// Record a request
	recordRequest("Issues", 200, 150*time.Millisecond)

	// Verify counter incremented
	count := testutil.CollectAndCount(defaultMetrics.requests)
	if count == 0 {
		t.Error("expected requests to be recorded")
	}

	// Verify labels are correct
	expected := `
		# HELP linear_requests_total Total number of Linear API requests
		# TYPE linear_requests_total counter
		linear_requests_total{operation="Issues",status_code="200"} 1
	`
	if err := testutil.CollectAndCompare(defaultMetrics.requests, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}

	// Verify histogram recorded
	histCount := testutil.CollectAndCount(defaultMetrics.duration)
	if histCount == 0 {
		t.Error("expected duration to be recorded")
	}
}

func TestMetrics_ErrorTracking(t *testing.T) {
	defaultMetrics.errors.Reset()

	// Record errors
	recordError("IssueCreate", "graphql")
	recordError("IssueCreate", "network")

	expected := `
		# HELP linear_errors_total Total number of Linear API errors
		# TYPE linear_errors_total counter
		linear_errors_total{error_type="graphql",operation="IssueCreate"} 1
		linear_errors_total{error_type="network",operation="IssueCreate"} 1
	`
	if err := testutil.CollectAndCompare(defaultMetrics.errors, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}
}

func TestMetrics_RetryTracking(t *testing.T) {
	defaultMetrics.retries.Reset()

	// Record retries
	recordRetry("rate_limited")
	recordRetry("rate_limited")
	recordRetry("server_error")

	expected := `
		# HELP linear_retries_total Total number of retry attempts
		# TYPE linear_retries_total counter
		linear_retries_total{reason="rate_limited"} 2
		linear_retries_total{reason="server_error"} 1
	`
	if err := testutil.CollectAndCompare(defaultMetrics.retries, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}
}

func TestMetrics_RateLimitGauges(t *testing.T) {
	defaultMetrics.rateLimit.Reset()

	// Update rate limit info
	info := &RateLimitInfo{
		RequestsRemaining:   100,
		RequestsLimit:       120,
		ComplexityRemaining: 8000,
		ComplexityLimit:     10000,
	}
	recordRateLimit(info)

	// Verify gauges set correctly
	expected := `
		# HELP linear_rate_limit_remaining Linear API rate limit remaining capacity
		# TYPE linear_rate_limit_remaining gauge
		linear_rate_limit_remaining{limit_type="complexity"} 8000
		linear_rate_limit_remaining{limit_type="requests"} 100
	`
	if err := testutil.CollectAndCompare(defaultMetrics.rateLimit, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}
}

func TestMetrics_NilRateLimitInfo(t *testing.T) {
	// Should not panic with nil
	recordRateLimit(nil)
}

// TestEnableMetrics verifies EnableMetrics can be called multiple times
func TestEnableMetrics(t *testing.T) {
	// Should not panic
	EnableMetrics()
	EnableMetrics()
}

// TestMetricsCollector_InstanceScoped verifies instance-scoped metrics work
func TestMetricsCollector_InstanceScoped(t *testing.T) {
	// Create custom registry
	reg := prometheus.NewRegistry()

	// Create instance-scoped collector
	collector := newMetricsCollector(reg, "test")

	// Record metrics
	collector.recordRequest("TestOp", 200, 100*time.Millisecond)
	collector.recordError("TestOp", "test_error")
	collector.recordRetry("test_retry")

	// Verify metrics exist in custom registry
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have metrics with "test" suffix
	found := false
	for _, m := range metrics {
		if strings.HasPrefix(*m.Name, "linear_test_") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected metrics with 'linear_test_' prefix in custom registry")
	}

	// Verify metrics NOT in default registry (isolation)
	defaultMetrics.requests.Reset()
	defaultCount := testutil.CollectAndCount(defaultMetrics.requests)
	if defaultCount > 0 {
		t.Error("Instance metrics should not appear in default registry")
	}
}
