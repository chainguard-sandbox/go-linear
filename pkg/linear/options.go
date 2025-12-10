package linear

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
// This replaces the default client entirely. To configure the default
// client, use WithTimeout, WithTransport, or WithTLSConfig instead.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.config.HTTPClient = client
	}
}

// WithBaseURL sets a custom base URL for the Linear API.
// Default: https://api.linear.app/graphql
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.config.BaseURL = url
	}
}

// WithUserAgent sets a custom user agent string.
// Default: go-linear/VERSION
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.config.UserAgent = ua
	}
}

// WithTimeout sets the HTTP client timeout.
// This creates a new http.Client if one hasn't been set via WithHTTPClient.
// Default: 30 seconds
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.config.HTTPClient.Timeout == 0 || c.config.HTTPClient == http.DefaultClient {
			// Safe: only modify our own client
			c.config.HTTPClient.Timeout = timeout
		} else {
			// If user provided custom client, create wrapper to avoid mutation
			oldClient := c.config.HTTPClient
			c.config.HTTPClient = &http.Client{
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
		c.config.HTTPClient.Transport = transport
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
		if transport, ok := c.config.HTTPClient.Transport.(*http.Transport); ok {
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
		c.config.Logger = logger
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
		c.config.Transport.MaxRetries = maxRetries
		c.config.Transport.InitialBackoff = initialBackoff
		c.config.Transport.MaxBackoff = maxBackoff
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
		c.config.OnRateLimit = callback
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
		c.config.Transport.MetricsEnabled = true
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
		c.config.Transport.MaxRetryDuration = duration
	}
}

// WithMetricsRegistry enables instance-scoped Prometheus metrics with a custom registry.
// Use this for multi-client scenarios to isolate metrics per client/workspace.
//
// The suffix parameter is appended to metric names to avoid collisions:
//   - linear_<suffix>_requests_total
//   - linear_<suffix>_request_duration_seconds
//
// Example:
//
//	// Production workspace metrics
//	prodReg := prometheus.NewRegistry()
//	prodClient, _ := linear.NewClient(prodKey,
//	    linear.WithMetricsRegistry(prodReg, "prod"))
//
//	// Staging workspace metrics (separate)
//	stageReg := prometheus.NewRegistry()
//	stageClient, _ := linear.NewClient(stageKey,
//	    linear.WithMetricsRegistry(stageReg, "staging"))
//
// Each client's metrics are isolated and can be exposed separately.
func WithMetricsRegistry(reg prometheus.Registerer, suffix string) Option {
	return func(c *Client) {
		c.config.Transport.MetricsCollector = newMetricsCollector(reg, suffix)
		c.config.Transport.MetricsEnabled = true
	}
}

// WithTracing enables OpenTelemetry distributed tracing.
// Requires OpenTelemetry SDK to be initialized in your application.
//
// Example:
//
//	import "go.opentelemetry.io/otel"
//
//	// Initialize OTel (usually in main)
//	tp := initTracer()  // Your tracer provider
//	otel.SetTracerProvider(tp)
//
//	// Enable tracing in client
//	client, _ := linear.NewClient(apiKey, linear.WithTracing())
func WithTracing() Option {
	return func(c *Client) {
		c.config.Transport.TracingEnabled = true
	}
}

// WithCircuitBreaker enables circuit breaker pattern for fail-fast behavior.
// Prevents cascading failures by stopping requests when error rate is high.
//
// The circuit breaker has three states:
//   - Closed: Normal operation, requests allowed
//   - Open: Circuit tripped, requests fail fast with ErrCircuitOpen
//   - HalfOpen: Testing if service recovered
//
// Example:
//
//	cb := &linear.CircuitBreaker{
//	    MaxFailures:  5,
//	    ResetTimeout: 60 * time.Second,
//	}
//	client, _ := linear.NewClient(apiKey, linear.WithCircuitBreaker(cb))
//
// The circuit opens after MaxFailures consecutive failures.
// After ResetTimeout, it enters HalfOpen state to test recovery.
func WithCircuitBreaker(cb *CircuitBreaker) Option {
	return func(c *Client) {
		c.config.Transport.CircuitBreaker = cb
	}
}

// WithCredentialProvider enables dynamic credential management.
// Useful for credential rotation, secret managers, or token refresh.
// When set, the apiKey parameter to NewClient can be empty.
//
// Example with AWS Secrets Manager:
//
//	type SecretsProvider struct {
//	    secretName string
//	    client     *secretsmanager.SecretsManager
//	}
//
//	func (p *SecretsProvider) GetCredential(ctx context.Context) (string, error) {
//	    result, err := p.client.GetSecretValue(&secretsmanager.GetSecretValueInput{
//	        SecretId: aws.String(p.secretName),
//	    })
//	    if err != nil {
//	        return "", err
//	    }
//	    return *result.SecretString, nil
//	}
//
//	provider := &SecretsProvider{secretName: "linear-api-key"}
//	client, _ := linear.NewClient("", linear.WithCredentialProvider(provider))
//
// The provider is called on client creation and automatically on 401 errors for credential refresh.
func WithCredentialProvider(provider CredentialProvider) Option {
	return func(c *Client) {
		c.credentialProvider = newCredentialCache(provider)
	}
}
