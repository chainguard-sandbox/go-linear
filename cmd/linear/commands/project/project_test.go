package project

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewProjectCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewProjectCommand(factory)

	if cmd.Use != "project" {
		t.Errorf("Use = %q, want %q", cmd.Use, "project")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestNewListCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewListCommand(factory)

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
	for _, flag := range []string{"limit", "output"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestRunList(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

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

	t.Run("invalid output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--output=invalid"})
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error")
		}
	})
}

func TestNewGetCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewGetCommand(factory)

	if !strings.HasPrefix(cmd.Use, "get") {
		t.Errorf("Use = %q, want prefix get", cmd.Use)
	}
}

func TestRunGet(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"proj-123", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "proj-123") {
			t.Error("Expected project id in output")
		}
	})

	t.Run("table output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"proj-123", "--output=table"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewCreateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewCreateCommand(factory)

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create")
	}
	if cmd.Flags().Lookup("name") == nil {
		t.Error("Expected name flag")
	}
}

func TestRunCreate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--name=New Project", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewUpdateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "update") {
		t.Errorf("Use = %q, want prefix update", cmd.Use)
	}
}

func TestRunUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"proj-123", "--name=Updated", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewDeleteCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewDeleteCommand(factory)

	if !strings.HasPrefix(cmd.Use, "delete") {
		t.Errorf("Use = %q, want prefix delete", cmd.Use)
	}
	if cmd.Flags().Lookup("yes") == nil {
		t.Error("Expected yes flag")
	}
}

func TestRunDelete(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("with confirmation", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"proj-123", "--yes", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewMilestoneCreateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewMilestoneCreateCommand(factory)

	if cmd.Use != "milestone-create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "milestone-create")
	}
	for _, flag := range []string{"project", "name"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestRunMilestoneCreate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewMilestoneCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		// Use UUID format to skip resolver lookup
		cmd.SetArgs([]string{"--project=00000000-0000-0000-0000-000000000001", "--name=Q1 2025", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewMilestoneUpdateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewMilestoneUpdateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "milestone-update") {
		t.Errorf("Use = %q, want prefix milestone-update", cmd.Use)
	}
}

func TestRunMilestoneUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("json output", func(t *testing.T) {
		cmd := NewMilestoneUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"milestone-123", "--name=Q2 2025", "--output=json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewMilestoneDeleteCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)
	cmd := NewMilestoneDeleteCommand(factory)

	if !strings.HasPrefix(cmd.Use, "milestone-delete") {
		t.Errorf("Use = %q, want prefix milestone-delete", cmd.Use)
	}
	if cmd.Flags().Lookup("yes") == nil {
		t.Error("Expected yes flag")
	}
}

func TestRunMilestoneDelete(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()
	factory := testFactory(t, server.URL)

	t.Run("with confirmation", func(t *testing.T) {
		cmd := NewMilestoneDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"milestone-123", "--yes"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
