package issue

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func TestNewGetCommand(t *testing.T) {
	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test")
	}

	cmd := NewGetCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "get <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "get <id>")
		}

		if cmd.Short == "" {
			t.Error("Short description should not be empty")
		}
	})

	t.Run("flags", func(t *testing.T) {
		fieldsFlag := cmd.Flags().Lookup("fields")
		if fieldsFlag == nil {
			t.Fatal("fields flag not found")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"arg1", "arg2"})
		if err == nil {
			t.Error("Expected error for two args")
		}

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunGet(t *testing.T) {
	// Create test server that returns a mock issue
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123",
					"title": "Test Issue",
					"description": "Test description",
					"priority": 2,
					"state": {"id": "state-1", "name": "In Progress"},
					"team": {"id": "team-1", "key": "ENG", "name": "Engineering"},
					"assignee": {"id": "user-1", "name": "Test User"},
					"createdAt": "2024-01-01T00:00:00.000Z",
					"updatedAt": "2024-01-02T00:00:00.000Z"
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
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ENG-123") {
			t.Errorf("Output should contain identifier, got: %s", output)
		}

		// Verify it's valid JSON
		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})
}

func TestRunGet_ClientError(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"Internal server error"}]}`))
	}))
	defer server.Close()

	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(server.URL))
	}

	cmd := NewGetCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"ENG-123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for server error")
	}
}

func TestRunGet_FieldsFiltering(t *testing.T) {
	// Create test server that returns a mock issue with all fields
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123",
					"title": "Test Issue",
					"description": "Test description",
					"priority": 2,
					"estimate": 5,
					"state": {"id": "state-1", "name": "In Progress"},
					"team": {"id": "team-1", "key": "ENG", "name": "Engineering"},
					"assignee": {"id": "user-1", "name": "Test User"},
					"createdAt": "2024-01-01T00:00:00.000Z",
					"updatedAt": "2024-01-02T00:00:00.000Z"
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
		cmd := NewGetCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--fields=id,title"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		if _, ok := result["id"]; !ok {
			t.Error("Expected 'id' field in output")
		}
		if _, ok := result["title"]; !ok {
			t.Error("Expected 'title' field in output")
		}
		if _, ok := result["description"]; ok {
			t.Error("Did not expect 'description' field in filtered output")
		}
	})
}

// Silence the unused import warning for time package
var _ = time.Now
