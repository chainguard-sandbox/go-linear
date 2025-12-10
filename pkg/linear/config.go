package linear

import (
	"log/slog"
	"net/http"
	"time"
)

// ClientConfig holds configuration for a Linear client.
type ClientConfig struct {
	// APIKey is the Linear API key or OAuth token
	APIKey string

	// BaseURL is the Linear API endpoint
	// Default: https://api.linear.app/graphql
	BaseURL string

	// UserAgent for HTTP requests
	// Default: go-linear/VERSION
	UserAgent string

	// HTTPClient is a custom HTTP client
	// Default: created with sensible timeouts and connection pooling
	HTTPClient *http.Client

	// Logger for structured logging
	Logger *slog.Logger

	// CredentialProvider for dynamic credential management
	// If set, APIKey is ignored
	CredentialProvider CredentialProvider

	// Transport configuration
	Transport *TransportConfig

	// OnRateLimit callback for rate limit events
	OnRateLimit func(*RateLimitInfo)
}

// TransportConfig holds transport-layer configuration.
type TransportConfig struct {
	// Retry configuration
	MaxRetries       int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	MaxRetryDuration time.Duration

	// Circuit breaker for fail-fast behavior
	CircuitBreaker *CircuitBreaker

	// Metrics configuration
	MetricsEnabled   bool
	MetricsCollector *MetricsCollector

	// Tracing configuration
	TracingEnabled bool
}

// newDefaultHTTPClient creates an HTTP client with sensible defaults and connection pooling.
func newDefaultHTTPClient() *http.Client {
	baseTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 3, // Match Linear's ~2 req/sec rate limit
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: baseTransport,
	}
}

// NewDefaultClientConfig creates a ClientConfig with sensible defaults.
func NewDefaultClientConfig(apiKey string) *ClientConfig {
	return &ClientConfig{
		APIKey:     apiKey,
		BaseURL:    "https://api.linear.app/graphql",
		UserAgent:  "go-linear/0.1.0",
		HTTPClient: newDefaultHTTPClient(),
	}
}

// NewDefaultTransportConfig creates a TransportConfig with sensible defaults.
func NewDefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		MaxRetries:       3,
		InitialBackoff:   1 * time.Second,
		MaxBackoff:       30 * time.Second,
		MaxRetryDuration: 90 * time.Second,
	}
}

// needsTransportWrapping determines if the Transport wrapper should be applied.
// Wrapping is needed if any observability or retry features are enabled.
func needsTransportWrapping(cfg *ClientConfig) bool {
	if cfg.Transport == nil {
		return false
	}
	return cfg.Transport.MaxRetries > 0 ||
		cfg.Logger != nil ||
		cfg.OnRateLimit != nil ||
		cfg.Transport.MetricsEnabled ||
		cfg.Transport.CircuitBreaker != nil
}

// buildTransport creates the appropriate Transport based on configuration.
// If observability features are enabled, wraps the base transport with retry/logging/metrics.
// Otherwise, returns the base transport unchanged.
func buildTransport(cfg *ClientConfig) http.RoundTripper {
	base := cfg.HTTPClient.Transport

	if !needsTransportWrapping(cfg) {
		return base
	}

	return &Transport{
		Base:             base,
		Logger:           cfg.Logger,
		MaxRetries:       cfg.Transport.MaxRetries,
		InitialBackoff:   cfg.Transport.InitialBackoff,
		MaxBackoff:       cfg.Transport.MaxBackoff,
		MaxRetryDuration: cfg.Transport.MaxRetryDuration,
		OnRateLimit:      cfg.OnRateLimit,
		MetricsEnabled:   cfg.Transport.MetricsEnabled,
		MetricsCollector: cfg.Transport.MetricsCollector,
		CircuitBreaker:   cfg.Transport.CircuitBreaker,
	}
}
