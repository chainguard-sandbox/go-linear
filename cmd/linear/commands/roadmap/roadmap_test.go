package roadmap

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewRoadmapCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewRoadmapCommand(factory)

	if cmd.Use != "roadmap" {
		t.Errorf("Use = %q, want %q", cmd.Use, "roadmap")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestNewListCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewListCommand(factory)

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
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

func TestNewGetCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewGetCommand(factory)

	if !strings.HasPrefix(cmd.Use, "get") {
		t.Errorf("Use = %q, want prefix get", cmd.Use)
	}
}

func TestRunGet(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"road-123"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "road-123") {
			t.Error("Expected roadmap id in output")
		}
	})
}
