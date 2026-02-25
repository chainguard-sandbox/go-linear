package document

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewUnarchiveCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewUnarchiveCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "unarchive <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "unarchive <id>")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"doc-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunDocumentUnarchive(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("unarchive json output", func(t *testing.T) {
		cmd := NewUnarchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"doc-123"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"success": true`) {
			t.Errorf("Output should contain success: true, got: %s", output)
		}
		if !strings.Contains(output, `"documentId"`) {
			t.Errorf("Output should contain documentId, got: %s", output)
		}
	})
}
