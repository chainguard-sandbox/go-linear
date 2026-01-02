package linear

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("lin_api_test",
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "lin_api_test123",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:   "with custom timeout",
			apiKey: "lin_api_test123",
			opts: []Option{
				WithTimeout(60 * time.Second),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClient_Viewer(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "lin_api_test" {
			t.Errorf("Authorization = %q, want %q", auth, "lin_api_test")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"Test","email":"test@example.com","displayName":"test","createdAt":"2024-01-01T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}}}`))
	})

	viewer, err := client.Viewer(context.Background())
	if err != nil {
		t.Fatalf("Viewer() error = %v", err)
	}

	if viewer.ID != "u1" {
		t.Errorf("ID = %q, want %q", viewer.ID, "u1")
	}
	if viewer.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", viewer.Email, "test@example.com")
	}
}

func TestClient_Close(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Should not error on first call
	if err := client.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Should be idempotent (safe to call multiple times)
	if err := client.Close(); err != nil {
		t.Errorf("Close() second call error = %v, want nil", err)
	}

	// Should be safe to call even after close
	if err := client.Close(); err != nil {
		t.Errorf("Close() third call error = %v, want nil", err)
	}
}

func TestClientOptions(t *testing.T) {
	// Test WithHTTPClient
	t.Run("WithHTTPClient", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}
		client, err := NewClient("lin_api_test", WithHTTPClient(customClient))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithUserAgent
	t.Run("WithUserAgent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != "custom-agent/1.0" {
				t.Errorf("User-Agent = %q, want %q", ua, "custom-agent/1.0")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"Test","email":"test@example.com","displayName":"test","createdAt":"2024-01-01T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}}}`))
		}))
		defer server.Close()

		client, err := NewClient("lin_api_test",
			WithBaseURL(server.URL),
			WithUserAgent("custom-agent/1.0"),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()

		_, err = client.Viewer(context.Background())
		if err != nil {
			t.Errorf("Viewer() error = %v", err)
		}
	})

	// Test WithLogger
	t.Run("WithLogger", func(t *testing.T) {
		logger := NewLogger()
		client, err := NewClient("lin_api_test", WithLogger(logger))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithRetry
	t.Run("WithRetry", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithRetry(5, 100*time.Millisecond, 1*time.Second))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithRateLimitCallback
	t.Run("WithRateLimitCallback", func(t *testing.T) {
		callback := func(info *RateLimitInfo) {}
		client, err := NewClient("lin_api_test", WithRateLimitCallback(callback))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithMaxRetryDuration
	t.Run("WithMaxRetryDuration", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithMaxRetryDuration(5*time.Minute))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithCircuitBreaker
	t.Run("WithCircuitBreaker", func(t *testing.T) {
		cb := &CircuitBreaker{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		}
		client, err := NewClient("lin_api_test", WithCircuitBreaker(cb))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithTransport
	t.Run("WithTransport", func(t *testing.T) {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		client, err := NewClient("lin_api_test", WithTransport(transport))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithMetrics
	t.Run("WithMetrics", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithMetrics())
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithTracing
	t.Run("WithTracing", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithTracing())
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"contains substring", "hello world", "world", true},
		{"does not contain", "hello world", "foo", false},
		{"empty substring", "hello", "", true},
		{"case sensitive", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"401 error", errors.New("HTTP 401 Unauthorized"), true},
		{"authentication error", errors.New("authentication failed"), true},
		{"unauthorized error", errors.New("unauthorized access"), true},
		{"random error", errors.New("Something went wrong"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("isAuthError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
