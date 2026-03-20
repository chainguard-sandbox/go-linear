package issue

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewSearchCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewSearchCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "search <query>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "search <query>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		expectedFlags := []string{"limit", "fields", "count", "include-archived"}
		for _, flag := range expectedFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"search term"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunSearch(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("search json output", func(t *testing.T) {
		cmd := NewSearchCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"test"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}

		if _, ok := result["nodes"]; !ok {
			t.Error("JSON output should have 'nodes' field")
		}
	})

	t.Run("search with limit", func(t *testing.T) {
		cmd := NewSearchCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"test", "--limit=10"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("search with field filtering", func(t *testing.T) {
		cmd := NewSearchCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"test", "--fields=id,title"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
