package linear

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics_RequestTracking(t *testing.T) {
	// Reset metrics before test
	requestsTotal.Reset()
	requestDuration.Reset()

	// Record a request
	recordRequest("Issues", 200, 150*time.Millisecond)

	// Verify counter incremented
	count := testutil.CollectAndCount(requestsTotal)
	if count == 0 {
		t.Error("expected requestsTotal to be recorded")
	}

	// Verify labels are correct
	expected := `
		# HELP linear_requests_total Total number of Linear API requests
		# TYPE linear_requests_total counter
		linear_requests_total{operation="Issues",status_code="200"} 1
	`
	if err := testutil.CollectAndCompare(requestsTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}

	// Verify histogram recorded
	histCount := testutil.CollectAndCount(requestDuration)
	if histCount == 0 {
		t.Error("expected requestDuration to be recorded")
	}
}

func TestMetrics_ErrorTracking(t *testing.T) {
	errorsTotal.Reset()

	// Record errors
	recordError("IssueCreate", "graphql")
	recordError("IssueCreate", "network")

	expected := `
		# HELP linear_errors_total Total number of Linear API errors
		# TYPE linear_errors_total counter
		linear_errors_total{error_type="graphql",operation="IssueCreate"} 1
		linear_errors_total{error_type="network",operation="IssueCreate"} 1
	`
	if err := testutil.CollectAndCompare(errorsTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}
}

func TestMetrics_RetryTracking(t *testing.T) {
	retriesTotal.Reset()

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
	if err := testutil.CollectAndCompare(retriesTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("unexpected metric value:\n%v", err)
	}
}

func TestMetrics_RateLimitGauges(t *testing.T) {
	rateLimitRemaining.Reset()

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
	if err := testutil.CollectAndCompare(rateLimitRemaining, strings.NewReader(expected)); err != nil {
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
