package viewer

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewViewerCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewViewerCommand(factory)

	if cmd.Use != "viewer" {
		t.Errorf("Use = %q, want %q", cmd.Use, "viewer")
	}
}

func TestRunViewer(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewViewerCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "viewer-123") {
			t.Error("Expected viewer id in output")
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewViewerCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=table"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
