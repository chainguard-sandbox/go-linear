package issue

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func TestNewListCommand(t *testing.T) {
	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test")
	}

	cmd := NewListCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "list" {
			t.Errorf("Use = %q, want %q", cmd.Use, "list")
		}

		if cmd.Short == "" {
			t.Error("Short description should not be empty")
		}
	})

	t.Run("pagination flags", func(t *testing.T) {
		limitFlag := cmd.Flags().Lookup("limit")
		if limitFlag == nil {
			t.Fatal("limit flag not found")
		}
		if limitFlag.DefValue != "50" {
			t.Errorf("limit default = %q, want %q", limitFlag.DefValue, "50")
		}

		afterFlag := cmd.Flags().Lookup("after")
		if afterFlag == nil {
			t.Fatal("after flag not found")
		}
	})

	t.Run("field flags", func(t *testing.T) {
		countFlag := cmd.Flags().Lookup("count")
		if countFlag == nil {
			t.Fatal("count flag not found")
		}

		fieldsFlag := cmd.Flags().Lookup("fields")
		if fieldsFlag == nil {
			t.Fatal("fields flag not found")
		}
	})

	t.Run("filter flags exist", func(t *testing.T) {
		expectedFlags := []string{
			"team", "assignee", "state", "priority",
			"created-after", "created-before",
			"label", "creator", "project", "cycle",
			"has-children", "has-suggested-teams",
		}

		for _, flag := range expectedFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})

	t.Run("no args required", func(t *testing.T) {
		// List command should accept no args
		if cmd.Args != nil {
			err := cmd.Args(cmd, []string{})
			if err != nil {
				t.Errorf("list should accept no args: %v", err)
			}
		}
	})
}

func TestRunList(t *testing.T) {
	// Create test server that returns a list of mock issues
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"issues": {
					"nodes": [
						{
							"id": "issue-1",
							"identifier": "ENG-1",
							"title": "First Issue",
							"priority": 1,
							"state": {"id": "state-1", "name": "Todo"},
							"team": {"id": "team-1", "key": "ENG", "name": "Engineering"},
							"createdAt": "2024-01-01T00:00:00.000Z",
							"updatedAt": "2024-01-02T00:00:00.000Z"
						},
						{
							"id": "issue-2",
							"identifier": "ENG-2",
							"title": "Second Issue",
							"priority": 2,
							"state": {"id": "state-2", "name": "In Progress"},
							"team": {"id": "team-1", "key": "ENG", "name": "Engineering"},
							"createdAt": "2024-01-01T00:00:00.000Z",
							"updatedAt": "2024-01-02T00:00:00.000Z"
						}
					],
					"pageInfo": {
						"hasNextPage": false,
						"endCursor": null
					}
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(server.URL))
	}

	t.Run("json output", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()

		// Verify it's valid JSON
		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}

		// Check nodes structure
		if _, ok := result["nodes"]; !ok {
			t.Error("JSON output should have 'nodes' field")
		}
	})

	t.Run("with limit", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--limit=10"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		// Output should be valid JSON
		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})
}

func TestRunList_EmptyResults(t *testing.T) {
	// Create test server that returns empty results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"issues": {
					"nodes": [],
					"pageInfo": {
						"hasNextPage": false,
						"endCursor": null
					}
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(server.URL))
	}

	t.Run("json output empty", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v", err)
		}

		nodes, ok := result["nodes"].([]any)
		if !ok {
			t.Error("Expected nodes array in JSON output")
		}
		if len(nodes) != 0 {
			t.Errorf("Expected empty nodes array, got %d items", len(nodes))
		}
	})
}

func TestRunList_FieldsFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"issues": {
					"nodes": [
						{
							"id": "issue-1",
							"identifier": "ENG-1",
							"title": "Test Issue",
							"description": "Test description",
							"priority": 1,
							"state": {"id": "state-1", "name": "Todo"},
							"team": {"id": "team-1", "key": "ENG"},
							"createdAt": "2024-01-01T00:00:00.000Z",
							"updatedAt": "2024-01-02T00:00:00.000Z"
						}
					],
					"pageInfo": {"hasNextPage": false}
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(server.URL))
	}

	t.Run("specific fields only", func(t *testing.T) {
		cmd := NewListCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--fields=id,title"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		nodes, ok := result["nodes"].([]any)
		if !ok || len(nodes) == 0 {
			t.Fatal("Expected nodes array with items")
		}

		firstNode := nodes[0].(map[string]any)
		if _, ok := firstNode["id"]; !ok {
			t.Error("Expected 'id' field in node")
		}
		if _, ok := firstNode["title"]; !ok {
			t.Error("Expected 'title' field in node")
		}
		// description should be filtered out
		if _, ok := firstNode["description"]; ok {
			t.Error("Did not expect 'description' field in filtered output")
		}
	})
}
