//go:build !integration

package initiative

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
	for _, flag := range []string{"initiative", "body"} {
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
			"--initiative=Security Initiative",
			"--body=Test update body",
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
			"--initiative=Security Initiative",
			"--body=Test update body",
			"--health=onTrack",
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

func TestNewStatusUpdateListCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateListCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-list") {
		t.Errorf("Use = %q, want prefix status-update-list", cmd.Use)
	}

	// Check required flags
	if cmd.Flags().Lookup("initiative") == nil {
		t.Error("Expected flag 'initiative'")
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
			"--initiative=Security Initiative",
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
		cmd.SetArgs([]string{"update-123"})

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
}

func TestNewStatusUpdateArchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewStatusUpdateArchiveCommand(factory)

	if !strings.HasPrefix(cmd.Use, "status-update-archive") {
		t.Errorf("Use = %q, want prefix status-update-archive", cmd.Use)
	}
}

func TestRunStatusUpdateArchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output with yes flag", func(t *testing.T) {
		cmd := NewStatusUpdateArchiveCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"update-123", "--yes"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Expected success in output, got: %s", output)
		}
	})
}
