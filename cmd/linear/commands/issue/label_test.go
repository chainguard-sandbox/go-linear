package issue

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewAddLabelCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewAddLabelCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "add-label <issue-id> <label>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "add-label <issue-id> <label>")
		}
	})

	t.Run("requires exactly two args", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err == nil {
			t.Error("Expected error for one arg")
		}

		err = cmd.Args(cmd, []string{"ENG-123", "bug"})
		if err != nil {
			t.Errorf("Unexpected error for two args: %v", err)
		}
	})
}

func TestRunAddLabel(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("add label json output", func(t *testing.T) {
		cmd := NewAddLabelCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "bug", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "issue-123") {
			t.Errorf("Output should contain issue id, got: %s", output)
		}
	})

	t.Run("add label table output", func(t *testing.T) {
		cmd := NewAddLabelCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "bug", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Added label") {
			t.Errorf("Table output should show 'Added label', got: %s", output)
		}
	})
}

func TestNewRemoveLabelCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewRemoveLabelCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "remove-label <issue-id> <label>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "remove-label <issue-id> <label>")
		}
	})

	t.Run("requires exactly two args", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"ENG-123", "bug"})
		if err != nil {
			t.Errorf("Unexpected error for two args: %v", err)
		}
	})
}

func TestRunRemoveLabel(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("remove label json output", func(t *testing.T) {
		cmd := NewRemoveLabelCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "bug", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "issue-123") {
			t.Errorf("Output should contain issue id, got: %s", output)
		}
	})

	t.Run("remove label table output", func(t *testing.T) {
		cmd := NewRemoveLabelCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "bug", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Removed label") {
			t.Errorf("Table output should show 'Removed label', got: %s", output)
		}
	})
}
