package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewTemplateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewTemplateCommand(factory)

	if cmd.Use != "template" {
		t.Errorf("Use = %q, want %q", cmd.Use, "template")
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
		cmd.SetArgs([]string{"tpl-123"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "tpl-123") {
			t.Error("Expected template id in output")
		}
	})

	t.Run("templateData in default output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"tpl-123"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "templateData") {
			t.Error("Expected templateData in default get output")
		}
	})
}
