package issue

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewUpdateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "update <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "update <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		expectedFlags := []string{
			"title", "description", "assignee", "state",
			"priority", "cycle", "project", "parent",
			"add-label", "remove-label", "output",
		}
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

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("update title", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--title=Updated Title", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update multiple fields", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"ENG-123",
			"--title=New Title",
			"--description=New description",
			"--priority=1",
			"--output=json",
		})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("Expected output")
		}
	})

	t.Run("update state", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--state=Done", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("update assignee", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--assignee=me", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--title=Updated", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Updated issue") {
			t.Errorf("Table output should show 'Updated issue', got: %s", output)
		}
	})

	t.Run("invalid output format", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"ENG-123", "--title=Test", "--output=invalid"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid output format")
		}
	})
}
