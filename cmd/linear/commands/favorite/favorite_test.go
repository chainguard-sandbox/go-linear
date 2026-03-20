package favorite

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewFavoriteCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewFavoriteCommand(factory)

	if cmd.Use != "favorite" {
		t.Errorf("Use = %q, want %q", cmd.Use, "favorite")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestNewCreateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewCreateCommand(factory)

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create")
	}
}

func TestRunCreate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("favorite issue", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		// Use UUID format to skip resolver
		cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewDeleteCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewDeleteCommand(factory)

	if !strings.HasPrefix(cmd.Use, "delete") {
		t.Errorf("Use = %q, want prefix delete", cmd.Use)
	}
}

func TestRunDelete(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewDeleteCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"fav-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
