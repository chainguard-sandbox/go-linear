//go:build read

package resolver

import (
	"context"
	"os"
	"testing"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func TestLive_ResolveTeam(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	r := New(client)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "resolve by team name",
			input:   "Engineering",
			wantErr: false,
		},
		{
			name:    "resolve by team key",
			input:   "ENG",
			wantErr: false,
		},
		{
			name:    "nonexistent team",
			input:   "nonexistent-team-xyz",
			wantErr: true,
		},
		{
			name:    "empty team name",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveTeam(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id == "" {
				t.Error("ResolveTeam() returned empty ID")
			}

			// Test cache hit on second call
			if !tt.wantErr {
				id2, err2 := r.ResolveTeam(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveTeam() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveTeam() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveUser(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	r := New(client)
	ctx := context.Background()

	// Test "me" special case
	t.Run("resolve me", func(t *testing.T) {
		id, err := r.ResolveUser(ctx, "me")
		if err != nil {
			t.Errorf("ResolveUser('me') error = %v", err)
		}
		if id == "" {
			t.Error("ResolveUser('me') returned empty ID")
		}
	})

	// Test UUID passthrough
	t.Run("uuid passthrough", func(t *testing.T) {
		uuid := "12345678-1234-1234-1234-123456789abc"
		id, err := r.ResolveUser(ctx, uuid)
		if err != nil {
			t.Errorf("ResolveUser(uuid) error = %v", err)
		}
		if id != uuid {
			t.Errorf("ResolveUser(uuid) = %s, want %s", id, uuid)
		}
	})

	// Test cache
	t.Run("cache works", func(t *testing.T) {
		id1, _ := r.ResolveUser(ctx, "me")
		id2, _ := r.ResolveUser(ctx, "me")
		if id1 != id2 {
			t.Errorf("Cache not working: %s != %s", id1, id2)
		}
	})
}

func TestLive_ResolveState(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	r := New(client)
	ctx := context.Background()

	// Just verify it works - actual state names are workspace-specific
	t.Run("uuid passthrough", func(t *testing.T) {
		uuid := "12345678-1234-1234-1234-123456789abc"
		id, err := r.ResolveState(ctx, uuid)
		if err != nil {
			t.Errorf("ResolveState(uuid) error = %v", err)
		}
		if id != uuid {
			t.Errorf("ResolveState(uuid) = %s, want %s", id, uuid)
		}
	})
}
