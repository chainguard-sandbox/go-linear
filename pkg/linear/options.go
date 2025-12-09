package linear

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
// This replaces the default client entirely. To configure the default
// client, use WithTimeout, WithTransport, or WithTLSConfig instead.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL for the Linear API.
// Default: https://api.linear.app/graphql
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithUserAgent sets a custom user agent string.
// Default: go-linear/VERSION
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithTimeout sets the HTTP client timeout.
// This creates a new http.Client if one hasn't been set via WithHTTPClient.
// Default: 30 seconds
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient.Timeout == 0 || c.httpClient == http.DefaultClient {
			// Safe: only modify our own client
			c.httpClient.Timeout = timeout
		} else {
			// If user provided custom client, create wrapper to avoid mutation
			oldClient := c.httpClient
			c.httpClient = &http.Client{
				Transport:     oldClient.Transport,
				CheckRedirect: oldClient.CheckRedirect,
				Jar:           oldClient.Jar,
				Timeout:       timeout,
			}
		}
	}
}

// WithTransport sets a custom http.RoundTripper.
// Useful for adding middleware like logging, metrics, or custom retry logic.
//
// Example:
//
//	transport := &linear.Transport{
//	    Base: http.DefaultTransport,
//	    Logger: slog.Default(),
//	    MaxRetries: 3,
//	}
//	client, _ := linear.NewClient(apiKey, linear.WithTransport(transport))
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		if t, ok := c.httpClient.Transport.(*http.Transport); ok {
			c.httpClient.Transport = transport
			// Store original transport for cleanup
			c.baseTransport = t
		} else {
			c.httpClient.Transport = transport
		}
	}
}

// WithTLSConfig sets custom TLS configuration.
// Useful for certificate pinning, custom CA certificates, or TLS version enforcement.
//
// Example:
//
//	tlsConfig := &tls.Config{
//	    MinVersion: tls.VersionTLS12,
//	}
//	client, _ := linear.NewClient(apiKey, linear.WithTLSConfig(tlsConfig))
func WithTLSConfig(config *tls.Config) Option {
	return func(c *Client) {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = config
		}
	}
}

// WithLogger enables structured logging using slog.
// Logs include request/response details, rate limiting info, and retry attempts.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
//	    Level: slog.LevelInfo,
//	}))
//	client, _ := linear.NewClient(apiKey, linear.WithLogger(logger))
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithRetry configures retry behavior for transient failures.
//
// Parameters:
//   - maxRetries: Maximum retry attempts (default: 3)
//   - initialBackoff: Starting backoff duration (default: 1s)
//   - maxBackoff: Maximum backoff duration (default: 30s)
//
// Retries use exponential backoff with jitter.
//
// Example:
//
//	client, _ := linear.NewClient(apiKey,
//	    linear.WithRetry(5, 500*time.Millisecond, 60*time.Second))
func WithRetry(maxRetries int, initialBackoff, maxBackoff time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.initialBackoff = initialBackoff
		c.maxBackoff = maxBackoff
	}
}

// WithRateLimitCallback sets a callback for rate limit events.
// Useful for metrics collection and monitoring.
//
// Example:
//
//	client, _ := linear.NewClient(apiKey,
//	    linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
//	        metrics.RecordRateLimit(info.RequestsRemaining)
//	    }))
func WithRateLimitCallback(callback func(*RateLimitInfo)) Option {
	return func(c *Client) {
		c.onRateLimit = callback
	}
}

// WithMetrics enables Prometheus metrics collection.
// Metrics are automatically registered with the default Prometheus registry
// and can be exposed at /metrics using promhttp.Handler().
//
// Requires calling linear.EnableMetrics() before creating clients.
//
// Metrics collected:
//   - linear_requests_total{operation, status_code}
//   - linear_request_duration_seconds{operation}
//   - linear_errors_total{operation, error_type}
//   - linear_retries_total{reason}
//   - linear_rate_limit_remaining{limit_type}
//
// Example:
//
//	linear.EnableMetrics()
//	client, _ := linear.NewClient(apiKey, linear.WithMetrics())
//
// See examples/prometheus/main.go for complete integration.
func WithMetrics() Option {
	return func(c *Client) {
		c.metricsEnabled = true
	}
}

// WithMaxRetryDuration sets the maximum total time spent retrying requests.
// This prevents unbounded retry loops during prolonged API outages.
//
// Default: 90 seconds
//
// Example:
//
//	// Allow up to 2 minutes of retries total
//	client, _ := linear.NewClient(apiKey,
//	    linear.WithMaxRetryDuration(2*time.Minute))
func WithMaxRetryDuration(duration time.Duration) Option {
	return func(c *Client) {
		c.maxRetryDuration = duration
	}
}
