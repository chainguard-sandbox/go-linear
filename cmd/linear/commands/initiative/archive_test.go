package initiative

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewArchiveCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewArchiveCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "archive <name-or-id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "archive <name-or-id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		outputFlag := cmd.Flags().Lookup("output")
		if outputFlag == nil {
			t.Fatal("output flag not found")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"init-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunArchive(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("archive json output", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"success": true`) {
			t.Errorf("Output should contain success: true, got: %s", output)
		}
		if !strings.Contains(output, `"initiativeId"`) {
			t.Errorf("Output should contain initiativeId, got: %s", output)
		}
	})

	t.Run("archive table output", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "archived successfully") {
			t.Errorf("Output should contain 'archived successfully', got: %s", output)
		}
	})
}
