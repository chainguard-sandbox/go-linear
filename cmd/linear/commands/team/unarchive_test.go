package team

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewUnarchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUnarchiveCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "unarchive <name-or-id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "unarchive <name-or-id>")
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

		err = cmd.Args(cmd, []string{"team-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunTeamUnarchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("unarchive json output", func(t *testing.T) {
		cmd := NewUnarchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Engineering", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"success": true`) {
			t.Errorf("Output should contain success: true, got: %s", output)
		}
		if !strings.Contains(output, `"teamId"`) {
			t.Errorf("Output should contain teamId, got: %s", output)
		}
	})

	t.Run("unarchive table output", func(t *testing.T) {
		cmd := NewUnarchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Engineering", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "unarchived successfully") {
			t.Errorf("Output should contain 'unarchived successfully', got: %s", output)
		}
	})
}
