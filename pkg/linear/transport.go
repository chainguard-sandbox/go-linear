package linear

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/chainguard-dev/clog"
)

// Transport wraps an http.RoundTripper to add production features like
// rate limiting detection, retry logic, and observability.
type Transport struct {
	// Base is the underlying http.RoundTripper.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// Logger for structured logging.
	// If nil, logging is disabled.
	Logger *clog.Logger

	// MaxRetries is the maximum number of retry attempts for transient errors.
	// Default: 3
	MaxRetries int

	// InitialBackoff is the initial backoff duration for retries.
	// Default: 1 second
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	// Default: 30 seconds
	MaxBackoff time.Duration

	// OnRateLimit is called when a rate limit is detected.
	// Can be used for metrics collection.
	OnRateLimit func(limit *RateLimitInfo)

	// OnRetry is called before each retry attempt.
	OnRetry func(attempt int, err error)

	// MaxRetryDuration is the maximum total time spent retrying.
	// Prevents unbounded retry loops during prolonged outages.
	// Default: 90 seconds
	MaxRetryDuration time.Duration

	// MetricsEnabled enables Prometheus metrics collection.
	// Metrics are recorded for requests, errors, retries, and rate limits.
	MetricsEnabled bool

	// MetricsCollector is the metrics collector to use.
	// If nil, uses global default collector.
	MetricsCollector *MetricsCollector

	// CircuitBreaker prevents cascading failures.
	// If nil, circuit breaker is disabled.
	CircuitBreaker *CircuitBreaker
}

// RateLimitInfo contains rate limit information from response headers.
type RateLimitInfo struct {
	// Request-based limits
	RequestsLimit     int
	RequestsRemaining int
	RequestsReset     time.Time

	// Complexity-based limits
	ComplexityLimit     int
	ComplexityRemaining int
	ComplexityReset     time.Time

	// Retry-After header (for 429 responses)
	RetryAfter time.Duration
}

// RoundTrip implements the http.RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	// Clone request body for retries (prevent request body exhaustion)
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	maxRetries := t.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	initialBackoff := t.InitialBackoff
	if initialBackoff == 0 {
		initialBackoff = 1 * time.Second
	}

	maxBackoff := t.MaxBackoff
	if maxBackoff == 0 {
		maxBackoff = 30 * time.Second
	}

	maxRetryDuration := t.MaxRetryDuration
	if maxRetryDuration == 0 {
		maxRetryDuration = 90 * time.Second
	}

	retryStartTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check circuit breaker before each attempt
		if t.CircuitBreaker != nil {
			if err := t.CircuitBreaker.Allow(); err != nil {
				return nil, err
			}
		}

		// Check if we've exceeded total retry time
		if attempt > 0 && time.Since(retryStartTime) > maxRetryDuration {
			return nil, fmt.Errorf("retry duration exceeded (%v): %w", maxRetryDuration, lastErr)
		}

		if attempt > 0 {
			// Restore request body for retry
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			// Calculate exponential backoff with jitter
			backoff := calculateBackoff(attempt, initialBackoff, maxBackoff)

			if t.OnRetry != nil {
				t.OnRetry(attempt, lastErr)
			}

			if t.Logger != nil {
				t.Logger.WarnContext(req.Context(), "retrying request",
					"attempt", attempt,
					"max_retries", maxRetries,
					"backoff", backoff,
					"error", lastErr.Error(),
				)
			}

			timer := time.NewTimer(backoff)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			case <-timer.C:
			}
		}

		// Extract operation name from GraphQL request BEFORE sending
		operation := extractOperationName(req)

		startTime := time.Now()
		resp, err := base.RoundTrip(req)
		duration := time.Since(startTime)

		if err != nil {
			lastErr = err

			// Record failure in circuit breaker
			if t.CircuitBreaker != nil {
				t.CircuitBreaker.RecordFailure()
			}

			if t.Logger != nil {
				t.Logger.ErrorContext(req.Context(), "request failed",
					"method", req.Method,
					"url", req.URL.String(),
					"duration", duration,
					"error", err.Error(),
				)
			}
			if t.MetricsEnabled {
				t.getMetricsCollector().recordError(operation, "network")
			}
			// Retry network errors
			if attempt < maxRetries && isRetryable(err) {
				if t.MetricsEnabled {
					t.getMetricsCollector().recordRetry("network_error")
				}
				continue
			}
			// Max retries exceeded or non-retryable error
			if attempt >= maxRetries {
				return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
			}
			return nil, err
		}

		// Check rate limiting
		if rateLimitInfo := parseRateLimitHeaders(resp); rateLimitInfo != nil {
			if t.OnRateLimit != nil {
				t.OnRateLimit(rateLimitInfo)
			}

			if t.MetricsEnabled {
				t.getMetricsCollector().recordRateLimit(rateLimitInfo)
			}

			if t.Logger != nil {
				t.Logger.DebugContext(req.Context(), "rate limit info",
					"requests_remaining", rateLimitInfo.RequestsRemaining,
					"requests_limit", rateLimitInfo.RequestsLimit,
					"complexity_remaining", rateLimitInfo.ComplexityRemaining,
					"complexity_limit", rateLimitInfo.ComplexityLimit,
				)
			}
		}

		// Record success in circuit breaker
		if t.CircuitBreaker != nil && resp.StatusCode < 500 {
			t.CircuitBreaker.RecordSuccess()
		}

		// Record metrics
		if t.MetricsEnabled {
			t.getMetricsCollector().recordRequest(operation, resp.StatusCode, duration)
		}

		// Log successful request
		if t.Logger != nil {
			ctx := req.Context()
			requestID := resp.Header.Get("X-Request-ID")
			if resp.StatusCode >= 400 {
				if requestID != "" {
					t.Logger.WarnContext(ctx, "request completed",
						"method", req.Method,
						"url", req.URL.String(),
						"status", resp.StatusCode,
						"duration", duration,
						"request_id", requestID,
					)
				} else {
					t.Logger.WarnContext(ctx, "request completed",
						"method", req.Method,
						"url", req.URL.String(),
						"status", resp.StatusCode,
						"duration", duration,
					)
				}
			} else {
				if requestID != "" {
					t.Logger.InfoContext(ctx, "request completed",
						"method", req.Method,
						"url", req.URL.String(),
						"status", resp.StatusCode,
						"duration", duration,
						"request_id", requestID,
					)
				} else {
					t.Logger.InfoContext(ctx, "request completed",
						"method", req.Method,
						"url", req.URL.String(),
						"status", resp.StatusCode,
						"duration", duration,
					)
				}
			}
		}

		// Handle 429 Rate Limit
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (429)")

			if attempt < maxRetries {
				if t.MetricsEnabled {
					t.getMetricsCollector().recordRetry("rate_limited")
				}

				// Use Retry-After header if present
				if retryAfter := parseRetryAfter(resp); retryAfter > 0 {
					if t.Logger != nil {
						t.Logger.WarnContext(req.Context(), "rate limited, waiting",
							"retry_after", retryAfter,
						)
					}

					timer := time.NewTimer(retryAfter)
					select {
					case <-req.Context().Done():
						timer.Stop()
						_ = resp.Body.Close()
						return nil, req.Context().Err()
					case <-timer.C:
					}
				}
				_ = resp.Body.Close()
				continue
			}
		}

		// Handle 5xx server errors with retry
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			lastErr = fmt.Errorf("server error (%d)", resp.StatusCode)
			if attempt < maxRetries {
				if t.MetricsEnabled {
					t.getMetricsCollector().recordRetry("server_error")
				}
				_ = resp.Body.Close()
				continue
			}
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateBackoff calculates exponential backoff with jitter.
func calculateBackoff(attempt int, initial, maxDuration time.Duration) time.Duration {
	backoff := float64(initial) * math.Pow(2, float64(attempt-1))
	if backoff > float64(maxDuration) {
		backoff = float64(maxDuration)
	}

	// Add jitter: ±25% of backoff using crypto/rand
	jitter := backoff * 0.25
	backoff += (cryptoRandFloat64()*2 - 1) * jitter

	return time.Duration(backoff)
}

// cryptoRandFloat64 returns a random float64 in [0, 1) using crypto/rand.
func cryptoRandFloat64() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to zero jitter on error
		return 0.5
	}
	// Convert to uint64 and normalize to [0, 1)
	return float64(binary.BigEndian.Uint64(b[:])>>11) / float64(1<<53)
}

// isRetryable determines if an error should trigger a retry.
func isRetryable(err error) bool {
	// Don't retry context cancellation or deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	// For now, retry all other errors. Could be made more sophisticated.
	return true
}

// parseRateLimitHeaders extracts rate limit information from response headers.
func parseRateLimitHeaders(resp *http.Response) *RateLimitInfo {
	h := resp.Header

	// Check if any rate limit headers are present
	if h.Get("X-RateLimit-Requests-Limit") == "" &&
		h.Get("X-RateLimit-Complexity-Limit") == "" {
		return nil
	}

	info := &RateLimitInfo{}

	// Parse request-based limits
	if v := h.Get("X-RateLimit-Requests-Limit"); v != "" {
		info.RequestsLimit, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Requests-Remaining"); v != "" {
		info.RequestsRemaining, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Requests-Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			info.RequestsReset = time.Unix(ts, 0)
		}
	}

	// Parse complexity-based limits
	if v := h.Get("X-RateLimit-Complexity-Limit"); v != "" {
		info.ComplexityLimit, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Complexity-Remaining"); v != "" {
		info.ComplexityRemaining, _ = strconv.Atoi(v)
	}
	if v := h.Get("X-RateLimit-Complexity-Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			info.ComplexityReset = time.Unix(ts, 0)
		}
	}

	return info
}

// extractOperationName extracts the GraphQL operation name from the request body.
// Returns "graphql" if unable to parse.
func extractOperationName(req *http.Request) string {
	if req.Body == nil {
		return "graphql"
	}

	// Read body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "graphql"
	}

	// Parse GraphQL request to extract operationName
	var gqlReq struct {
		OperationName string `json:"operationName"`
	}
	if err := json.Unmarshal(body, &gqlReq); err == nil && gqlReq.OperationName != "" {
		// Restore body for actual request
		req.Body = io.NopCloser(bytes.NewReader(body))
		return gqlReq.OperationName
	}

	// Restore body even on parse failure
	req.Body = io.NopCloser(bytes.NewReader(body))
	return "graphql"
}

// getMetricsCollector returns the metrics collector to use.
// Returns the instance collector if set, otherwise the global default.
func (t *Transport) getMetricsCollector() *MetricsCollector {
	if t.MetricsCollector != nil {
		return t.MetricsCollector
	}
	return defaultMetrics
}

// parseRetryAfter parses the Retry-After header.
func parseRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return 0
}
