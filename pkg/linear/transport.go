package linear

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
)

// Transport wraps an http.RoundTripper to add production features like
// rate limiting detection, retry logic, and observability.
type Transport struct {
	// Base is the underlying http.RoundTripper.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// Logger for structured logging.
	// If nil, logging is disabled.
	Logger *slog.Logger

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

	// MetricsEnabled enables Prometheus metrics collection.
	// Metrics are recorded for requests, errors, retries, and rate limits.
	MetricsEnabled bool
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

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff with jitter
			backoff := calculateBackoff(attempt, initialBackoff, maxBackoff)

			if t.OnRetry != nil {
				t.OnRetry(attempt, lastErr)
			}

			if t.Logger != nil {
				t.Logger.LogAttrs(req.Context(), slog.LevelWarn,
					"retrying request",
					slog.Int("attempt", attempt),
					slog.Int("max_retries", maxRetries),
					slog.Duration("backoff", backoff),
					slog.String("error", lastErr.Error()),
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

		startTime := time.Now()
		resp, err := base.RoundTrip(req)
		duration := time.Since(startTime)

		operation := "graphql" // Default operation name

		if err != nil {
			lastErr = err
			if t.Logger != nil {
				t.Logger.LogAttrs(req.Context(), slog.LevelError,
					"request failed",
					slog.String("method", req.Method),
					slog.String("url", req.URL.String()),
					slog.Duration("duration", duration),
					slog.String("error", err.Error()),
				)
			}
			if t.MetricsEnabled {
				recordError(operation, "network")
			}
			// Retry network errors
			if attempt < maxRetries && isRetryable(err) {
				if t.MetricsEnabled {
					recordRetry("network_error")
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
				recordRateLimit(rateLimitInfo)
			}

			if t.Logger != nil {
				t.Logger.LogAttrs(req.Context(), slog.LevelDebug,
					"rate limit info",
					slog.Int("requests_remaining", rateLimitInfo.RequestsRemaining),
					slog.Int("requests_limit", rateLimitInfo.RequestsLimit),
					slog.Int("complexity_remaining", rateLimitInfo.ComplexityRemaining),
					slog.Int("complexity_limit", rateLimitInfo.ComplexityLimit),
				)
			}
		}

		// Record metrics
		if t.MetricsEnabled {
			recordRequest(operation, resp.StatusCode, duration)
		}

		// Log successful request
		if t.Logger != nil {
			level := slog.LevelInfo
			if resp.StatusCode >= 400 {
				level = slog.LevelWarn
			}

			// Build attributes with request ID if available
			requestID := resp.Header.Get("X-Request-ID")
			if requestID != "" {
				t.Logger.LogAttrs(req.Context(), level,
					"request completed",
					slog.String("method", req.Method),
					slog.String("url", req.URL.String()),
					slog.Int("status", resp.StatusCode),
					slog.Duration("duration", duration),
					slog.String("request_id", requestID),
				)
			} else {
				t.Logger.LogAttrs(req.Context(), level,
					"request completed",
					slog.String("method", req.Method),
					slog.String("url", req.URL.String()),
					slog.Int("status", resp.StatusCode),
					slog.Duration("duration", duration),
				)
			}
		}

		// Handle 429 Rate Limit
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (429)")

			if attempt < maxRetries {
				if t.MetricsEnabled {
					recordRetry("rate_limited")
				}

				// Use Retry-After header if present
				if retryAfter := parseRetryAfter(resp); retryAfter > 0 {
					if t.Logger != nil {
						t.Logger.LogAttrs(req.Context(), slog.LevelWarn,
							"rate limited, waiting",
							slog.Duration("retry_after", retryAfter),
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
					recordRetry("server_error")
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
func calculateBackoff(attempt int, initial, max time.Duration) time.Duration {
	backoff := float64(initial) * math.Pow(2, float64(attempt-1))
	if backoff > float64(max) {
		backoff = float64(max)
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
