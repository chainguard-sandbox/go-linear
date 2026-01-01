package organization

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewOrganizationCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewOrganizationCommand(factory)

	if cmd.Use != "organization" {
		t.Errorf("Use = %q, want %q", cmd.Use, "organization")
	}
}

func TestRunOrganization(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewOrganizationCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "org-123") {
			t.Error("Expected org id in output")
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewOrganizationCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=table"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
