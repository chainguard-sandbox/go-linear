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
