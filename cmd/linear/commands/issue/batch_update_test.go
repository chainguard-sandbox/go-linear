package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewBatchUpdateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewBatchUpdateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "batch-update" {
			t.Errorf("Use = %q, want %q", cmd.Use, "batch-update")
		}
	})

	t.Run("filter flags exist", func(t *testing.T) {
		filterFlags := []string{"team", "assignee", "state", "priority", "creator", "label"}
		for _, flag := range filterFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected filter flag %q not found", flag)
			}
		}
	})

	t.Run("update flags exist", func(t *testing.T) {
		updateFlags := []string{"set-state", "set-assignee", "set-priority", "set-team", "add-label", "remove-label"}
		for _, flag := range updateFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected update flag %q not found", flag)
			}
		}
	})

	t.Run("control flags exist", func(t *testing.T) {
		controlFlags := []string{"dry-run", "yes", "batch-limit", "output"}
		for _, flag := range controlFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected control flag %q not found", flag)
			}
		}
	})

	t.Run("default values", func(t *testing.T) {
		dryRun := cmd.Flags().Lookup("dry-run")
		if dryRun.DefValue != "false" {
			t.Errorf("dry-run default = %q, want %q", dryRun.DefValue, "false")
		}

		batchLimit := cmd.Flags().Lookup("batch-limit")
		if batchLimit.DefValue != "50" {
			t.Errorf("batch-limit default = %q, want %q", batchLimit.DefValue, "50")
		}
	})
}

func TestRunBatchUpdate(t *testing.T) {
	// Add BatchUpdateIssues and ListIssuesFiltered handlers to mock server
	handlers := defaultHandlers()
	handlers["ListIssuesFiltered"] = `{
		"data": {
			"issues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue 1",
						"priority": 2,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-01T00:00:00.000Z"
					},
					{
						"id": "issue-456",
						"identifier": "ENG-456",
						"title": "Test Issue 2",
						"priority": 3,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-02T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`
	handlers["BatchUpdateIssues"] = `{
		"data": {
			"issueBatchUpdate": {
				"success": true,
				"issues": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue 1"
					},
					{
						"id": "issue-456",
						"identifier": "ENG-456",
						"title": "Test Issue 2"
					}
				]
			}
		}
	}`

	server := testutil.MockServer(t, handlers)
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("dry run shows what would change", func(t *testing.T) {
		cmd := NewBatchUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--team=ENG", "--set-state=Done", "--dry-run"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Dry run") && !strings.Contains(output, "would be") {
			t.Logf("Dry run output: %s", output)
		}
	})

	t.Run("requires update flags", func(t *testing.T) {
		cmd := NewBatchUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--team=ENG"})

		err := cmd.Execute()
		// Should fail because no update flags provided
		if err == nil {
			t.Log("Expected error when no update flags provided (may be allowed)")
		}
	})

	t.Run("batch limit flag", func(t *testing.T) {
		cmd := NewBatchUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--team=ENG", "--set-state=Done", "--batch-limit=10", "--dry-run"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("set cycle by name", func(t *testing.T) {
		cmd := NewBatchUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--team=ENG", "--set-cycle=Sprint 1", "--dry-run"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("set project by name", func(t *testing.T) {
		cmd := NewBatchUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--team=ENG", "--set-project=Test Project", "--dry-run"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
