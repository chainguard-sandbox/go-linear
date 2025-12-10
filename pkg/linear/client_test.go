package linear

import (
	"context"
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
