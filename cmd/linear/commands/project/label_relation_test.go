package project

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewLabelCreateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelCreateCommand(factory)
	if !strings.HasPrefix(cmd.Use, "label-create") {
		t.Errorf("Use = %q, want prefix label-create", cmd.Use)
	}
}

func TestRunLabelCreate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelCreateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--name=Backend", "--color=#ff0000"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestRunLabelList(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestRunLabelListPagination(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--limit=10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with --limit error = %v", err)
	}
}

func TestRunLabelUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"plabel-123", "--name=Updated"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRunLabelDelete(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelDeleteCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"plabel-123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRunRelationCreate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationCreateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"--project=00000000-0000-0000-0000-000000000001",
		"--related-project=00000000-0000-0000-0000-000000000002",
		"--type=blocks",
		"--anchor-type=project",
		"--related-anchor-type=project",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRunRelationCreateInvalidType(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationCreateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--project=00000000-0000-0000-0000-000000000001",
		"--related-project=00000000-0000-0000-0000-000000000002",
		"--type=depends_on",
		"--anchor-type=project",
		"--related-anchor-type=project",
	})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() expected error for invalid --type, got nil")
	}
}

func TestRunRelationCreateInvalidAnchorType(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationCreateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--project=00000000-0000-0000-0000-000000000001",
		"--related-project=00000000-0000-0000-0000-000000000002",
		"--type=blocks",
		"--anchor-type=invalid",
		"--related-anchor-type=project",
	})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() expected error for invalid --anchor-type, got nil")
	}
}

func TestRunRelationList(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestRunRelationListPagination(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationListCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--limit=10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with --limit error = %v", err)
	}
}

func TestRunRelationUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"prel-123", "--type=related"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRunLabelUpdateZeroValuePropagates(t *testing.T) {
	server, lastVars := mockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewLabelUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"plabel-123", "--name="})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables[input] type = %T, want map", vars["input"])
	}
	name, ok := input["name"]
	if !ok {
		t.Error("input.name missing: --name= should propagate empty string, not be dropped")
	}
	if name != "" {
		t.Errorf("input.name = %q, want empty string", name)
	}
}

func TestRunRelationUpdateInvalidType(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"prel-123", "--type=depends_on"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() expected error for invalid --type, got nil")
	}
}

func TestRunRelationDelete(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	cmd := NewRelationDeleteCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"prel-123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
