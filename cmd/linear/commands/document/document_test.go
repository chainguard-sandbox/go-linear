package document

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewDocumentCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewDocumentCommand(factory)

	if cmd.Use != "document" {
		t.Errorf("Use = %q, want %q", cmd.Use, "document")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestRunList(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestRunGet(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"doc-123"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "doc-123") {
			t.Error("Expected document id in output")
		}
	})
}

func TestRunCreate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--title=New Document",
			"--content=Test content",
			"--project=Test Project",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "doc-new") {
			t.Error("Expected document id in output")
		}
	})
}

func TestRunUpdate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"doc-123",
			"--title=Updated Document",
		})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "Updated Document") {
			t.Error("Expected updated title in output")
		}
	})
}

func TestRunDelete(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("with yes flag", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"doc-123", "--yes"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "success") {
			t.Error("Expected success in output")
		}
	})
}
