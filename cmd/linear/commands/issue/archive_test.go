package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewArchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewArchiveCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "archive <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "archive <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		trashFlag := cmd.Flags().Lookup("trash")
		if trashFlag == nil {
			t.Fatal("trash flag not found")
		}

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

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunArchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("archive json output", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Output should contain success, got: %s", output)
		}
		if !strings.Contains(output, "\"trashed\": false") {
			t.Errorf("Output should contain trashed: false, got: %s", output)
		}
	})

	t.Run("archive with trash flag", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--trash", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "\"trashed\": true") {
			t.Errorf("Output should contain trashed: true, got: %s", output)
		}
	})

	t.Run("archive table output", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "archived") {
			t.Errorf("Table output should contain 'archived', got: %s", output)
		}
	})

	t.Run("archive with trash table output", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--trash", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "moved to trash") {
			t.Errorf("Table output should contain 'moved to trash', got: %s", output)
		}
	})

	t.Run("invalid output format", func(t *testing.T) {
		cmd := NewArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"ENG-123", "--output=invalid"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid output format")
		}
	})
}
