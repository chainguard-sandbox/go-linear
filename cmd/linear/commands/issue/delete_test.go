package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewDeleteCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewDeleteCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "delete <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "delete <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		yesFlag := cmd.Flags().Lookup("yes")
		if yesFlag == nil {
			t.Fatal("yes flag not found")
		}

		permanentFlag := cmd.Flags().Lookup("permanent")
		if permanentFlag == nil {
			t.Fatal("permanent flag not found")
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

func TestRunDelete(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("delete with confirmation", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--yes"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Output should contain success, got: %s", output)
		}
	})

	t.Run("delete with permanent flag", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--yes", "--permanent"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "\"permanent\": true") {
			t.Errorf("Output should contain permanent: true, got: %s", output)
		}
	})
}
