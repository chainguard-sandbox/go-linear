package customview

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewCustomViewCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewCustomViewCommand(factory)
	if cmd.Use != "custom-view" {
		t.Errorf("Use = %q, want %q", cmd.Use, "custom-view")
	}
	if len(cmd.Commands()) != 5 {
		t.Errorf("Expected 5 subcommands, got %d", len(cmd.Commands()))
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

	t.Run("list with fields", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--fields=id,name"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestRunGet(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewGetCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"cv-123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestRunCreate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewCreateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--name=Test View"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRunDelete(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewDeleteCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"cv-123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
