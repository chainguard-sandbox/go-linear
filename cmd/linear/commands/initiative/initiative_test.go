package initiative

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewInitiativeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewInitiativeCommand(factory)

	if cmd.Use != "initiative" {
		t.Errorf("Use = %q, want %q", cmd.Use, "initiative")
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
		cmd.SetArgs([]string{"--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=table"})
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
		cmd.SetArgs([]string{"Security Initiative", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "init-123") {
			t.Error("Expected initiative id in output")
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--output=table"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		output := buf.String()
		// Validate enhanced fields are displayed
		if !strings.Contains(output, "init-123") {
			t.Error("Expected ID in output")
		}
		if !strings.Contains(output, "Active") {
			t.Error("Expected status in output")
		}
		if !strings.Contains(output, "atRisk") {
			t.Error("Expected health in output")
		}
		if !strings.Contains(output, "Test Owner") {
			t.Error("Expected owner name in output")
		}
		if !strings.Contains(output, "Parent Initiative") {
			t.Error("Expected parent initiative in output")
		}
		if !strings.Contains(output, "2 linked") {
			t.Error("Expected linked projects count in output")
		}
		if !strings.Contains(output, "Test Project 1") {
			t.Error("Expected linked project name in output")
		}
	})
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

	t.Run("json output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--name=Test Initiative", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewUpdateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "update") {
		t.Errorf("Use = %q, want prefix update", cmd.Use)
	}
}

func TestRunUpdate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--name=Updated Initiative", "--output=json"})
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

	t.Run("json output with yes flag", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--yes", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("table output with yes flag", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"Security Initiative", "--yes", "--output=table"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewAddProjectCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewAddProjectCommand(factory)

	if !strings.HasPrefix(cmd.Use, "add-project") {
		t.Errorf("Use = %q, want prefix add-project", cmd.Use)
	}

	for _, flag := range []string{"initiative", "project"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestRunAddProject(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewAddProjectCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--initiative=Security Initiative",
			"--project=Test Project",
			"--output=json",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "link-123") {
			t.Errorf("Expected link-123 in output, got: %s", output)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewAddProjectCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--initiative=Security Initiative",
			"--project=Test Project",
			"--output=table",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Linked") {
			t.Errorf("Expected linked message in output, got: %s", output)
		}
	})
}

func TestNewRemoveProjectCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewRemoveProjectCommand(factory)

	if !strings.HasPrefix(cmd.Use, "remove-project") {
		t.Errorf("Use = %q, want prefix remove-project", cmd.Use)
	}

	for _, flag := range []string{"initiative", "project"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestRunRemoveProject(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewRemoveProjectCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--initiative=Security Initiative",
			"--project=Test Project",
			"--output=json",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") {
			t.Errorf("Expected success in output, got: %s", output)
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewRemoveProjectCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--initiative=Security Initiative",
			"--project=Test Project",
			"--output=table",
		})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Unlinked") {
			t.Errorf("Expected unlinked message in output, got: %s", output)
		}
	})
}
