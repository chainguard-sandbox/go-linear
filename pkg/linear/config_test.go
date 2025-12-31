package linear

import (
	"net/http"
	"testing"
	"time"
)

func TestNewDefaultClientConfig(t *testing.T) {
	apiKey := "test-api-key"
	config := NewDefaultClientConfig(apiKey)

	if config.APIKey != apiKey {
		t.Errorf("expected APIKey %q, got %q", apiKey, config.APIKey)
	}

	if config.BaseURL != "https://api.linear.app/graphql" {
		t.Errorf("expected BaseURL https://api.linear.app/graphql, got %q", config.BaseURL)
	}

	if config.UserAgent != "go-linear/0.1.0" {
		t.Errorf("expected UserAgent go-linear/0.1.0, got %q", config.UserAgent)
	}

	if config.HTTPClient == nil {
		t.Fatal("expected HTTPClient to be non-nil")
	}

	if config.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.HTTPClient.Timeout)
	}
}

func TestNewDefaultTransportConfig(t *testing.T) {
	config := NewDefaultTransportConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.InitialBackoff != 1*time.Second {
		t.Errorf("expected InitialBackoff 1s, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff 30s, got %v", config.MaxBackoff)
	}

	if config.MaxRetryDuration != 90*time.Second {
		t.Errorf("expected MaxRetryDuration 90s, got %v", config.MaxRetryDuration)
	}
}

func TestNeedsTransportWrapping(t *testing.T) {
	tests := []struct {
		name     string
		config   *ClientConfig
		expected bool
	}{
		{
			name: "no wrapping needed - nil transport",
			config: &ClientConfig{
				Transport: nil,
			},
			expected: false,
		},
		{
			name: "no wrapping needed - zero values",
			config: &ClientConfig{
				Transport: &TransportConfig{
					MaxRetries: 0,
				},
			},
			expected: false,
		},
		{
			name: "wrapping needed - retries > 0",
			config: &ClientConfig{
				Transport: &TransportConfig{
					MaxRetries: 3,
				},
			},
			expected: true,
		},
		{
			name: "wrapping needed - logger set",
			config: &ClientConfig{
				Logger:    NewLogger(),
				Transport: &TransportConfig{},
			},
			expected: true,
		},
		{
			name: "wrapping needed - rate limit callback",
			config: &ClientConfig{
				OnRateLimit: func(*RateLimitInfo) {},
				Transport:   &TransportConfig{},
			},
			expected: true,
		},
		{
			name: "wrapping needed - metrics enabled",
			config: &ClientConfig{
				Transport: &TransportConfig{
					MetricsEnabled: true,
				},
			},
			expected: true,
		},
		{
			name: "wrapping needed - circuit breaker",
			config: &ClientConfig{
				Transport: &TransportConfig{
					CircuitBreaker: &CircuitBreaker{
						MaxFailures:  5,
						ResetTimeout: 60 * time.Second,
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsTransportWrapping(tt.config)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBuildTransport(t *testing.T) {
	t.Run("no wrapping", func(t *testing.T) {
		config := &ClientConfig{
			HTTPClient: &http.Client{
				Transport: http.DefaultTransport,
			},
			Transport: &TransportConfig{
				MaxRetries: 0,
			},
		}

		result := buildTransport(config)
		if result != http.DefaultTransport {
			t.Error("expected base transport to be returned unchanged")
		}
	})

	t.Run("with wrapping", func(t *testing.T) {
		config := &ClientConfig{
			HTTPClient: &http.Client{
				Transport: http.DefaultTransport,
			},
			Transport: &TransportConfig{
				MaxRetries:       3,
				InitialBackoff:   1 * time.Second,
				MaxBackoff:       30 * time.Second,
				MaxRetryDuration: 90 * time.Second,
			},
		}

		result := buildTransport(config)
		transport, ok := result.(*Transport)
		if !ok {
			t.Fatal("expected Transport wrapper")
		}

		if transport.Base != http.DefaultTransport {
			t.Error("expected Base to be http.DefaultTransport")
		}

		if transport.MaxRetries != 3 {
			t.Errorf("expected MaxRetries 3, got %d", transport.MaxRetries)
		}

		if transport.InitialBackoff != 1*time.Second {
			t.Errorf("expected InitialBackoff 1s, got %v", transport.InitialBackoff)
		}
	})

	t.Run("with all features", func(t *testing.T) {
		cb := &CircuitBreaker{
			MaxFailures:  5,
			ResetTimeout: 60 * time.Second,
		}

		callback := func(*RateLimitInfo) {}

		config := &ClientConfig{
			HTTPClient: &http.Client{
				Transport: http.DefaultTransport,
			},
			Logger:      NewLogger(),
			OnRateLimit: callback,
			Transport: &TransportConfig{
				MaxRetries:       3,
				InitialBackoff:   1 * time.Second,
				MaxBackoff:       30 * time.Second,
				MaxRetryDuration: 90 * time.Second,
				CircuitBreaker:   cb,
				MetricsEnabled:   true,
			},
		}

		result := buildTransport(config)
		transport, ok := result.(*Transport)
		if !ok {
			t.Fatal("expected Transport wrapper")
		}

		if transport.CircuitBreaker != cb {
			t.Error("expected CircuitBreaker to be set")
		}

		if transport.Logger == nil {
			t.Error("expected Logger to be set")
		}

		if !transport.MetricsEnabled {
			t.Error("expected MetricsEnabled to be true")
		}
	})
}
