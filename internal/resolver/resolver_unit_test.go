package resolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
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

// Mock HTTP server tests for resolver methods

func newMockResolver(t *testing.T, handler http.HandlerFunc) *Resolver {
	t.Helper()

	// Clear cache to avoid pollution from other tests
	cleanupCacheDir(t)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := linear.NewClient("lin_api_test", linear.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return New(client)
}

func TestResolveTeam_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"teams": map[string]any{
					"nodes": []map[string]any{
						{"id": "team-123", "name": "Engineering", "key": "ENG"},
						{"id": "team-456", "name": "Platform", "key": "PLT"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"by name", "Engineering", "team-123", false},
		{"by key", "ENG", "team-123", false},
		{"case insensitive", "engineering", "team-123", false},
		{"not found", "NonExistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ResolveTeam(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveTeam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveUser_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var reqBody struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)

		// Check if it's a viewer query
		if contains(reqBody.Query, "viewer") {
			resp := map[string]any{
				"data": map[string]any{
					"viewer": map[string]any{
						"id": "user-me", "name": "Current User", "email": "me@example.com",
						"displayName": "me", "active": true, "admin": false, "createdAt": "2024-01-01T00:00:00.000Z",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Users query
		resp := map[string]any{
			"data": map[string]any{
				"users": map[string]any{
					"nodes": []map[string]any{
						{"id": "user-123", "name": "Alice Smith", "email": "alice@example.com", "displayName": "alice"},
						{"id": "user-456", "name": "Bob Jones", "email": "bob@example.com", "displayName": "bob"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"by name", "Alice Smith", "user-123", false},
		{"by email", "alice@example.com", "user-123", false},
		{"by display name", "alice", "user-123", false},
		{"me keyword", "me", "user-me", false},
		{"ME uppercase", "ME", "user-me", false},
		{"not found", "nonexistent@example.com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ResolveUser(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveState_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"workflowStates": map[string]any{
					"nodes": []map[string]any{
						{"id": "state-1", "name": "Backlog", "type": "backlog", "color": "#bbb"},
						{"id": "state-2", "name": "In Progress", "type": "started", "color": "#f00"},
						{"id": "state-3", "name": "Done", "type": "completed", "color": "#0f0"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"by name", "In Progress", "state-2", false},
		{"case insensitive", "in progress", "state-2", false},
		{"not found", "Unknown State", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ResolveState(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveLabel_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"issueLabels": map[string]any{
					"nodes": []map[string]any{
						{"id": "label-1", "name": "bug", "color": "#f00"},
						{"id": "label-2", "name": "feature", "color": "#0f0"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveLabel(ctx, "bug")
	if err != nil {
		t.Errorf("ResolveLabel() error = %v", err)
	}
	if got != "label-1" {
		t.Errorf("ResolveLabel() = %v, want label-1", got)
	}
}

func TestResolveIssue_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// ResolveIssue now uses direct issue lookup instead of search
		resp := map[string]any{
			"data": map[string]any{
				"issue": map[string]any{
					"id":         "issue-123",
					"identifier": "ENG-123",
					"title":      "Test Issue",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveIssue(ctx, "ENG-123")
	if err != nil {
		t.Errorf("ResolveIssue() error = %v", err)
	}
	if got != "issue-123" {
		t.Errorf("ResolveIssue() = %v, want issue-123", got)
	}
}

func TestResolveProject_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"projects": map[string]any{
					"nodes": []map[string]any{
						{"id": "proj-1", "name": "Platform Redesign"},
						{"id": "proj-2", "name": "API v2"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveProject(ctx, "Platform Redesign")
	if err != nil {
		t.Errorf("ResolveProject() error = %v", err)
	}
	if got != "proj-1" {
		t.Errorf("ResolveProject() = %v, want proj-1", got)
	}
}

func TestResolveCycle_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := "Sprint 42"
		resp := map[string]any{
			"data": map[string]any{
				"cycles": map[string]any{
					"nodes": []map[string]any{
						{"id": "cycle-1", "name": &name},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveCycle(ctx, "Sprint 42")
	if err != nil {
		t.Errorf("ResolveCycle() error = %v", err)
	}
	if got != "cycle-1" {
		t.Errorf("ResolveCycle() = %v, want cycle-1", got)
	}
}

func TestResolveInitiative_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"initiatives": map[string]any{
					"nodes": []map[string]any{
						{"id": "init-1", "name": "Q1 Goals"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveInitiative(ctx, "Q1 Goals")
	if err != nil {
		t.Errorf("ResolveInitiative() error = %v", err)
	}
	if got != "init-1" {
		t.Errorf("ResolveInitiative() = %v, want init-1", got)
	}
}

func TestResolveDocument_APICall(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"documents": map[string]any{
					"nodes": []map[string]any{
						{"id": "doc-1", "title": "Design Doc"},
					},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveDocument(ctx, "Design Doc")
	if err != nil {
		t.Errorf("ResolveDocument() error = %v", err)
	}
	if got != "doc-1" {
		t.Errorf("ResolveDocument() = %v, want doc-1", got)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestResolveLabel_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"issueLabels": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveLabel(ctx, "nonexistent")
	if err == nil {
		t.Error("ResolveLabel() should return error for nonexistent label")
	}
}

func TestResolveIssue_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"searchIssues": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveIssue(ctx, "NONEXISTENT-999")
	if err == nil {
		t.Error("ResolveIssue() should return error for nonexistent issue")
	}
}

func TestResolveProject_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"projects": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveProject(ctx, "nonexistent")
	if err == nil {
		t.Error("ResolveProject() should return error for nonexistent project")
	}
}

func TestResolveCycle_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"cycles": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveCycle(ctx, "nonexistent")
	if err == nil {
		t.Error("ResolveCycle() should return error for nonexistent cycle")
	}
}

func TestResolveInitiative_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"initiatives": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveInitiative(ctx, "nonexistent")
	if err == nil {
		t.Error("ResolveInitiative() should return error for nonexistent initiative")
	}
}

func TestResolveDocument_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"documents": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveDocument(ctx, "nonexistent")
	if err == nil {
		t.Error("ResolveDocument() should return error for nonexistent document")
	}
}

func TestResolveUser_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"users": map[string]any{
					"nodes":    []map[string]any{},
					"pageInfo": map[string]any{"hasNextPage": false},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	_, err := resolver.ResolveUser(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("ResolveUser() should return error for nonexistent user")
	}
}
