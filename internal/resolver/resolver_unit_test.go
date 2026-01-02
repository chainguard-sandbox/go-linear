package resolver

import (
	"context"
	"testing"
	"time"
)

// Unit tests that don't require API access.
// These test UUID passthrough, input validation, and cache behavior.

func TestUUIDRegex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"12345678-1234-1234-1234-123456789abc", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", true},
		{"12345678-1234-1234-1234-123456789ABC", false}, // uppercase not matched
		{"not-a-uuid", false},
		{"12345678123412341234123456789abc", false}, // no dashes
		{"", false},
		{"12345678-1234-1234-1234-123456789ab", false},   // too short
		{"12345678-1234-1234-1234-123456789abcd", false}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := uuidRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("uuidRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolver_UUIDPassthrough(t *testing.T) {
	// Create resolver with nil client - UUID passthrough shouldn't need API
	r := &Resolver{
		client: nil,
		cache:  NewCache(5 * time.Minute),
	}
	ctx := context.Background()

	uuid := "12345678-1234-1234-1234-123456789abc"

	// All resolve methods should pass through UUIDs without API call
	tests := []struct {
		name   string
		method func(context.Context, string) (string, error)
	}{
		{"ResolveTeam", r.ResolveTeam},
		{"ResolveUser", r.ResolveUser},
		{"ResolveState", r.ResolveState},
		{"ResolveLabel", r.ResolveLabel},
		{"ResolveIssue", r.ResolveIssue},
		{"ResolveProject", r.ResolveProject},
		{"ResolveCycle", r.ResolveCycle},
		{"ResolveInitiative", r.ResolveInitiative},
		{"ResolveDocument", r.ResolveDocument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.method(ctx, uuid)
			if err != nil {
				t.Errorf("%s(%q) error = %v", tt.name, uuid, err)
			}
			if got != uuid {
				t.Errorf("%s(%q) = %q, want %q", tt.name, uuid, got, uuid)
			}
		})
	}
}

func TestResolver_EmptyInputValidation(t *testing.T) {
	r := &Resolver{
		client: nil,
		cache:  NewCache(5 * time.Minute),
	}
	ctx := context.Background()

	tests := []struct {
		name   string
		method func(context.Context, string) (string, error)
	}{
		{"ResolveTeam", r.ResolveTeam},
		{"ResolveUser", r.ResolveUser},
		{"ResolveState", r.ResolveState},
		{"ResolveLabel", r.ResolveLabel},
		{"ResolveIssue", r.ResolveIssue},
		{"ResolveProject", r.ResolveProject},
		{"ResolveCycle", r.ResolveCycle},
		{"ResolveInitiative", r.ResolveInitiative},
		{"ResolveDocument", r.ResolveDocument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.method(ctx, "")
			if err == nil {
				t.Errorf("%s(\"\") should return error for empty input", tt.name)
			}
		})
	}
}

func TestResolver_CacheHitSkipsAPI(t *testing.T) {
	// Create resolver with nil client - cache hits shouldn't need API
	cache := NewCache(5 * time.Minute)
	r := &Resolver{
		client: nil,
		cache:  cache,
	}
	ctx := context.Background()

	// Pre-populate cache
	cache.Set("team:engineering", "team-uuid-123")
	cache.Set("user:alice@example.com", "user-uuid-456")
	cache.Set("state:in progress", "state-uuid-789")
	cache.Set("label:bug", "label-uuid-abc")
	cache.Set("project:platform", "project-uuid-def")
	cache.Set("cycle:sprint 42", "cycle-uuid-ghi")
	cache.Set("initiative:q1 goals", "initiative-uuid-jkl")
	cache.Set("document:design doc", "document-uuid-mno")

	// Give async writes time to complete
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name   string
		method func(context.Context, string) (string, error)
		input  string
		want   string
	}{
		{"ResolveTeam", r.ResolveTeam, "Engineering", "team-uuid-123"},
		{"ResolveUser", r.ResolveUser, "alice@example.com", "user-uuid-456"},
		{"ResolveState", r.ResolveState, "In Progress", "state-uuid-789"},
		{"ResolveLabel", r.ResolveLabel, "bug", "label-uuid-abc"},
		{"ResolveProject", r.ResolveProject, "Platform", "project-uuid-def"},
		{"ResolveCycle", r.ResolveCycle, "Sprint 42", "cycle-uuid-ghi"},
		{"ResolveInitiative", r.ResolveInitiative, "Q1 Goals", "initiative-uuid-jkl"},
		{"ResolveDocument", r.ResolveDocument, "Design Doc", "document-uuid-mno"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.method(ctx, tt.input)
			if err != nil {
				t.Errorf("%s(%q) error = %v", tt.name, tt.input, err)
			}
			if got != tt.want {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.input, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Test that New creates a resolver with cache
	r := New(nil)
	if r == nil {
		t.Fatal("New(nil) returned nil")
	}
	if r.cache == nil {
		t.Error("New(nil) resolver has nil cache")
	}
}
