package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewUnarchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
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

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunUnarchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("unarchive json output", func(t *testing.T) {
		cmd := NewUnarchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Output should contain success, got: %s", output)
		}
	})
}
