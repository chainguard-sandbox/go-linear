package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewRelateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewRelateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "relate <issue-id> <related-issue-id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "relate <issue-id> <related-issue-id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		typeFlag := cmd.Flags().Lookup("type")
		if typeFlag == nil {
			t.Fatal("type flag not found")
		}
		if typeFlag.DefValue != "related" {
			t.Errorf("type default = %q, want %q", typeFlag.DefValue, "related")
		}
	})

	t.Run("requires exactly two args", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"ENG-123", "ENG-456"})
		if err != nil {
			t.Errorf("Unexpected error for two args: %v", err)
		}
	})
}

func TestRunRelate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("relate issues json output", func(t *testing.T) {
		cmd := NewRelateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "ENG-456", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "relation-123") {
			t.Errorf("Output should contain relation id, got: %s", output)
		}
	})

	t.Run("relate with type blocks", func(t *testing.T) {
		cmd := NewRelateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "ENG-456", "--type=blocks", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("relate table output", func(t *testing.T) {
		cmd := NewRelateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "ENG-456", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Created") || !strings.Contains(output, "relation") {
			t.Errorf("Table output should show relation created, got: %s", output)
		}
	})
}

func TestNewUnrelateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUnrelateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "unrelate <relation-id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "unrelate <relation-id>")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"relation-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunUnrelate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("unrelate json output", func(t *testing.T) {
		cmd := NewUnrelateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"relation-123", "--yes", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		// JSON output should contain success field or be a message
		if !strings.Contains(output, "success") && !strings.Contains(output, "Deleted") {
			t.Errorf("Output should contain success or Deleted, got: %s", output)
		}
	})

	t.Run("unrelate table output", func(t *testing.T) {
		cmd := NewUnrelateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"relation-123", "--yes", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Deleted") {
			t.Errorf("Table output should show 'Deleted', got: %s", output)
		}
	})
}

func TestNewUpdateRelationCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUpdateRelationCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "update-relation <relation-id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "update-relation <relation-id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		typeFlag := cmd.Flags().Lookup("type")
		if typeFlag == nil {
			t.Fatal("type flag not found")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"relation-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunUpdateRelation(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("update relation json output", func(t *testing.T) {
		cmd := NewUpdateRelationCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"relation-123", "--type=blocks", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "relation-123") {
			t.Errorf("Output should contain relation id, got: %s", output)
		}
	})

	t.Run("update relation table output", func(t *testing.T) {
		cmd := NewUpdateRelationCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"relation-123", "--type=related", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Updated") {
			t.Errorf("Table output should show 'Updated', got: %s", output)
		}
	})
}
