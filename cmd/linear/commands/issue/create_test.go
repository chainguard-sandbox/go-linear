package issue

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewCreateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewCreateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "create" {
			t.Errorf("Use = %q, want %q", cmd.Use, "create")
		}
		if cmd.Short == "" {
			t.Error("Short description should not be empty")
		}
	})

	t.Run("required flags", func(t *testing.T) {
		titleFlag := cmd.Flags().Lookup("title")
		if titleFlag == nil {
			t.Fatal("title flag not found")
		}
	})

	t.Run("optional flags", func(t *testing.T) {
		expectedFlags := []string{
			"team", "description", "assignee", "state",
			"priority", "label", "cycle", "project",
			"parent", "estimate", "output",
			"due-date", "milestone",
		}
		for _, flag := range expectedFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})
}

func TestRunCreate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("create with required fields", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--title=Test Issue", "--team=ENG", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ENG-999") {
			t.Errorf("Output should contain new issue identifier, got: %s", output)
		}
	})

	t.Run("create with all optional fields", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"--title=Full Issue",
			"--team=ENG",
			"--description=Test description",
			"--assignee=me",
			"--state=Todo",
			"--priority=1",
			"--label=bug",
			"--output=json",
		})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("create table output", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--title=Test Issue", "--team=ENG", "--output=table"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Created issue") {
			t.Errorf("Table output should show 'Created issue', got: %s", output)
		}
	})

	t.Run("create without team fails", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--title=Test Issue"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error without team")
		}
	})

	t.Run("invalid output format", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--title=Test", "--team=ENG", "--output=invalid"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid output format")
		}
	})

	t.Run("create with cycle by name", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--title=Test Issue", "--team=ENG", "--cycle=Sprint 1", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("create with project by name", func(t *testing.T) {
		cmd := NewCreateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--title=Test Issue", "--team=ENG", "--project=Test Project", "--output=json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})
}
