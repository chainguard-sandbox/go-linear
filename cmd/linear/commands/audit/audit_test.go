package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewAuditCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewAuditCommand(factory)
	if cmd.Use != "audit" {
		t.Errorf("Use = %q, want %q", cmd.Use, "audit")
	}
	if len(cmd.Commands()) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(cmd.Commands()))
	}
}

func TestNewListCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}

	expectedFlags := []string{"type", "actor", "ip", "country-code", "created-after", "created-before", "limit"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q not found", flag)
		}
	}
}

func TestRunList(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("list json output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v", err)
		}
		if _, ok := result["nodes"]; !ok {
			t.Error("Expected 'nodes' field in output")
		}
	})

	t.Run("list with type filter", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--type=issue.create"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("list with created-after filter", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--created-after=7d"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("list with created-before filter", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--created-before=7d"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("list with limit", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--limit=10"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestRunListCreatedAfterVariables(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--created-after=7d"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}

	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	createdAt, ok := filter["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("filter[createdAt] = %T, want map", filter["createdAt"])
	}
	gte, ok := createdAt["gte"].(string)
	if !ok {
		t.Fatalf("createdAt[gte] = %T, want string", createdAt["gte"])
	}

	parsed, err := time.Parse(time.RFC3339, gte)
	if err != nil {
		t.Fatalf("createdAt.gte %q is not valid RFC3339: %v", gte, err)
	}

	// Should be approximately 7 days ago — within a 1-day window either side
	expected := time.Now().UTC().Add(-7 * 24 * time.Hour)
	diff := parsed.Sub(expected)
	if diff < -25*time.Hour || diff > 25*time.Hour {
		t.Errorf("createdAt.gte = %v, want ~7 days ago (%v)", parsed, expected)
	}
}

func TestRunListActorResolvesName(t *testing.T) {
	handlers := defaultHandlers()
	handlers["ListUsers"] = testutil.MockUsersResponse

	server, lastVars := testutil.MockServerCapture(t, handlers)
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	// "Test User" will be resolved via ResolveUser → ListUsers mock → "user-123"
	cmd.SetArgs([]string{"--actor=Test User"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	actor, ok := filter["actor"].(map[string]any)
	if !ok {
		t.Fatalf("filter[actor] = %T, want map", filter["actor"])
	}
	id, ok := actor["id"].(map[string]any)
	if !ok {
		t.Fatalf("actor[id] = %T, want map", actor["id"])
	}
	eq, ok := id["eq"].(string)
	if !ok {
		t.Fatalf("id[eq] = %T, want string", id["eq"])
	}
	if eq != "user-123" {
		t.Errorf("actor.id.eq = %q, want %q", eq, "user-123")
	}
}

func TestRunListTypeFilterVariables(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--type=issue.create"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	typ, ok := filter["type"].(map[string]any)
	if !ok {
		t.Fatalf("filter[type] = %T, want map", filter["type"])
	}
	if eq, _ := typ["eq"].(string); !strings.EqualFold(eq, "issue.create") {
		t.Errorf("type.eq = %q, want issue.create", eq)
	}
}

func TestRunListIPFilterVariables(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--ip=1.2.3.4"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	ip, ok := filter["ip"].(map[string]any)
	if !ok {
		t.Fatalf("filter[ip] = %T, want map", filter["ip"])
	}
	if eq, _ := ip["eq"].(string); eq != "1.2.3.4" {
		t.Errorf("ip.eq = %q, want %q", eq, "1.2.3.4")
	}
}

func TestRunListCountryCodeFilterVariables(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--country-code=US"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	cc, ok := filter["countryCode"].(map[string]any)
	if !ok {
		t.Fatalf("filter[countryCode] = %T, want map", filter["countryCode"])
	}
	if eq, _ := cc["eq"].(string); eq != "US" {
		t.Errorf("countryCode.eq = %q, want %q", eq, "US")
	}
}

func TestRunListCreatedBeforeVariables(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--created-before=7d"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	createdAt, ok := filter["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("filter[createdAt] = %T, want map", filter["createdAt"])
	}
	lte, ok := createdAt["lte"].(string)
	if !ok {
		t.Fatalf("createdAt[lte] = %T, want string", createdAt["lte"])
	}
	if _, err := time.Parse(time.RFC3339, lte); err != nil {
		t.Errorf("createdAt.lte %q is not valid RFC3339: %v", lte, err)
	}
	if createdAt["gte"] != nil {
		t.Errorf("createdAt.gte should be absent, got %v", createdAt["gte"])
	}
}

func TestRunListCreatedDateRange(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--created-after=30d", "--created-before=7d"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	filter, ok := vars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables[filter] = %T, want map", vars["filter"])
	}
	createdAt, ok := filter["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("filter[createdAt] = %T, want map", filter["createdAt"])
	}
	if _, ok := createdAt["gte"].(string); !ok {
		t.Errorf("createdAt.gte missing or not string, got %T", createdAt["gte"])
	}
	if _, ok := createdAt["lte"].(string); !ok {
		t.Errorf("createdAt.lte missing or not string, got %T", createdAt["lte"])
	}
}

func TestRunListInvertedDateRange(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	// after=7d is ~April 2026, before=30d is ~April 2026 but earlier — inverted
	cmd.SetArgs([]string{"--created-after=7d", "--created-before=30d"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for inverted date range, got nil")
	}
	if !strings.Contains(err.Error(), "created-after") || !strings.Contains(err.Error(), "created-before") {
		t.Errorf("error %q should mention both flags", err.Error())
	}
}

func TestRunListInvalidCreatedAfter(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--created-after=not-a-date"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for invalid --created-after, got nil")
	}
	if !strings.Contains(err.Error(), "created-after") {
		t.Errorf("error %q should mention 'created-after'", err.Error())
	}
}

func TestRunListActorNotFound(t *testing.T) {
	handlers := defaultHandlers()
	// Return empty users list so resolver fails to find the actor
	handlers["ListUsers"] = `{"data":{"users":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`

	server := testutil.MockServer(t, handlers)
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--actor=nobody"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() expected error when actor not found, got nil")
	}
}

func TestRunListAfterCursorForwarded(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--after=cursor-abc123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	after, ok := vars["after"].(string)
	if !ok {
		t.Fatalf("vars[after] = %T, want string", vars["after"])
	}
	if after != "cursor-abc123" {
		t.Errorf("after = %q, want %q", after, "cursor-abc123")
	}
}

func TestRunTypes(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewTypesCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result []any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON array: %v", err)
	}
	if len(result) == 0 {
		t.Error("Expected non-empty types list")
	}
}
