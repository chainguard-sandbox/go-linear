package team

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewTeamCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewTeamCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "team" {
			t.Errorf("Use = %q, want %q", cmd.Use, "team")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		subcommands := cmd.Commands()
		if len(subcommands) == 0 {
			t.Error("Expected subcommands to be added")
		}

		expectedSubcommands := []string{"list", "get", "members", "create", "update", "delete"}
		subcommandNames := make(map[string]bool)
		for _, sub := range subcommands {
			subcommandNames[sub.Use] = true
		}

		for _, expected := range expectedSubcommands {
			found := false
			for name := range subcommandNames {
				if name == expected || strings.HasPrefix(name, expected+" ") {
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

func TestNewListCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewListCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "list" {
			t.Errorf("Use = %q, want %q", cmd.Use, "list")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		expectedFlags := []string{"limit", "output", "fields"}
		for _, flag := range expectedFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})
}

func TestRunList(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("list json output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("list table output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ENG") {
			t.Errorf("Table output should contain team key, got: %s", output)
		}
	})

	t.Run("list with limit", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--limit=10", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("invalid output format", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--output=invalid"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid output format")
		}
	})
}

func TestNewGetCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewGetCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if !strings.HasPrefix(cmd.Use, "get") {
			t.Errorf("Use = %q, want prefix %q", cmd.Use, "get")
		}
	})

	t.Run("requires one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"ENG"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunGet(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("get json output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "team-123") {
			t.Errorf("Output should contain team id, got: %s", output)
		}
	})

	t.Run("get table output", func(t *testing.T) {
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ENG") || !strings.Contains(output, "Engineering") {
			t.Errorf("Table output should contain team info, got: %s", output)
		}
	})
}

func TestNewMembersCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewMembersCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "members" {
			t.Errorf("Use = %q, want %q", cmd.Use, "members")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		teamFlag := cmd.Flags().Lookup("team")
		if teamFlag == nil {
			t.Error("team flag not found")
		}
	})
}

func TestNewCreateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewCreateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "create" {
			t.Errorf("Use = %q, want %q", cmd.Use, "create")
		}
	})

	t.Run("required flags", func(t *testing.T) {
		nameFlag := cmd.Flags().Lookup("name")
		if nameFlag == nil {
			t.Error("name flag not found")
		}

		keyFlag := cmd.Flags().Lookup("key")
		if keyFlag == nil {
			t.Error("key flag not found")
		}
	})
}

func TestRunCreate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("create team json output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--name=New Team", "--key=NEW", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "team-new") {
			t.Errorf("Output should contain new team id, got: %s", output)
		}
	})

	t.Run("create team table output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--name=New Team", "--key=NEW", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Created") && !strings.Contains(output, "NEW") {
			t.Errorf("Table output should show created team, got: %s", output)
		}
	})
}

func TestNewUpdateCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if !strings.HasPrefix(cmd.Use, "update") {
			t.Errorf("Use = %q, want prefix %q", cmd.Use, "update")
		}
	})

	t.Run("update flags exist", func(t *testing.T) {
		updateFlags := []string{"name", "description"}
		for _, flag := range updateFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})
}

func TestRunUpdate(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("update team json output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--name=Updated Engineering", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "team-123") {
			t.Errorf("Output should contain team id, got: %s", output)
		}
	})

	t.Run("update team table output", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--name=Updated Engineering", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Updated") || !strings.Contains(output, "Engineering") {
			t.Errorf("Table output should show updated team, got: %s", output)
		}
	})
}

func TestNewDeleteCommand(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)
	cmd := NewDeleteCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if !strings.HasPrefix(cmd.Use, "delete") {
			t.Errorf("Use = %q, want prefix %q", cmd.Use, "delete")
		}
	})

	t.Run("yes flag exists", func(t *testing.T) {
		yesFlag := cmd.Flags().Lookup("yes")
		if yesFlag == nil {
			t.Error("yes flag not found")
		}
	})
}

func TestRunDelete(t *testing.T) {
	server := mockServer(t, defaultHandlers())
	defer server.Close()

	factory := testFactory(t, server.URL)

	t.Run("delete team with confirmation", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--yes", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "success") && !strings.Contains(output, "true") && !strings.Contains(output, "deleted") {
			t.Errorf("Output should indicate success, got: %s", output)
		}
	})

	t.Run("delete team table output", func(t *testing.T) {
		cmd := NewDeleteCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG", "--yes", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(strings.ToLower(output), "deleted") {
			t.Errorf("Table output should show deleted, got: %s", output)
		}
	})
}
