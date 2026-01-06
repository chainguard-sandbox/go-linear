package issue

import (
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewIssueCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewIssueCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "issue" {
			t.Errorf("Use = %q, want %q", cmd.Use, "issue")
		}
		if cmd.Short != "Manage Linear issues" {
			t.Errorf("Short = %q, want %q", cmd.Short, "Manage Linear issues")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		if len(subcommands) == 0 {
			t.Error("Expected subcommands to be added")
		}

		// Check for expected subcommands
		expectedSubcommands := []string{
			"list", "get", "search", "create", "update",
			"batch-update", "delete", "relate", "update-relation",
			"unrelate", "add-label", "remove-label",
		}

		subcommandNames := make(map[string]bool)
		for _, sub := range subcommands {
			subcommandNames[sub.Use] = true
		}

		for _, expected := range expectedSubcommands {
			found := false
			for name := range subcommandNames {
				// Use string matching since Use may include args
				if name == expected || len(name) > len(expected) && name[:len(expected)] == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected subcommand %q not found", expected)
			}
		}
	})
}
