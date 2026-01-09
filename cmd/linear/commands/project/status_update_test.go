//go:build !integration

package project

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewStatusUpdateCreateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateCreateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-create") {
		t.Errorf("Use = %q, want prefix status-update-create", cmd.Use)
	}

	// Check required flags
	for _, flag := range []string{"project", "body"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestRunStatusUpdateCreate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewStatusUpdateCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--project=Test Project",
			"--body=Test update body",
			"--output=json",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
	})

	t.Run("json output with health", func(t *testing.T) {
		cmd := NewStatusUpdateCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--project=Test Project",
			"--body=Test update body",
			"--health=onTrack",
			"--output=json",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewStatusUpdateCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--project=Test Project",
			"--body=Test update body",
			"--output=table",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Created project status update") {
			t.Errorf("Expected success message in output, got: %s", output)
		}
	})
}

func TestNewStatusUpdateListCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateListCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-list") {
		t.Errorf("Use = %q, want prefix status-update-list", cmd.Use)
	}

	// Check required flags
	if cmd.Flags().Lookup("project") == nil {
		t.Error("Expected flag 'project'")
	}
}

func TestRunStatusUpdateList(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewStatusUpdateListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--project=Test Project",
			"--output=json",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewStatusUpdateListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--project=Test Project",
			"--output=table",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
	})
}

func TestNewStatusUpdateGetCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateGetCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-get") {
		t.Errorf("Use = %q, want prefix status-update-get", cmd.Use)
	}
}

func TestRunStatusUpdateGet(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewStatusUpdateGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"update-123", "--output=json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
		if !strings.Contains(output, "Test update body") {
			t.Errorf("Expected body in output, got: %s", output)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewStatusUpdateGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"update-123", "--output=table"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "update-123") {
			t.Errorf("Expected update-123 in output, got: %s", output)
		}
	})
}

func TestNewStatusUpdateDeleteCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateDeleteCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-delete") {
		t.Errorf("Use = %q, want prefix status-update-delete", cmd.Use)
	}
}

func TestRunStatusUpdateDelete(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output with yes flag", func(t *testing.T) {
		cmd := NewStatusUpdateDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"update-123", "--yes", "--output=json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Expected success in output, got: %s", output)
		}
	})

	t.Run("table output with yes flag", func(t *testing.T) {
		cmd := NewStatusUpdateDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"update-123", "--yes", "--output=table"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "deleted") {
			t.Errorf("Expected deletion confirmation in output, got: %s", output)
		}
	})
}
