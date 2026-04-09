package issue

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func TestNewUpdateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "update <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "update <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		expectedFlags := []string{
			"title", "description", "assignee", "state",
			"priority", "estimate", "cycle", "project", "parent",
			"add-label", "remove-label",
			"due-date", "milestone",
		}
		for _, flag := range expectedFlags {
			if cmd.Flags().Lookup(flag) == nil {
				t.Errorf("Expected flag %q not found", flag)
			}
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Error("Expected error for no args")
		}

		err = cmd.Args(cmd, []string{"ENG-123"})
		if err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunUpdate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("update title", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--title=Updated Title"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update multiple fields", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{
			"ENG-123",
			"--title=New Title",
			"--description=New description",
			"--priority=1",
		})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("Expected output")
		}
	})

	t.Run("update state", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--state=Done"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("update assignee", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--assignee=me"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("unassign with none", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--assignee=none"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update project by name", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--project=Test Project"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update estimate", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--estimate=5"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("clear estimate with none", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--estimate=none"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update estimate with float", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--estimate=1.5"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}
	})

	t.Run("update cycle by name", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--cycle=Sprint 1"})

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

type updateCapture struct {
	Input map[string]any
}

func captureUpdateServer(t *testing.T) (*httptest.Server, *updateCapture) {
	t.Helper()
	captured := &updateCapture{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var reqBody struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if strings.EqualFold(reqBody.OperationName, "UpdateIssue") {
			if input, ok := reqBody.Variables["input"].(map[string]any); ok {
				captured.Input = input
			}
		}
		handlers := defaultHandlers()
		for key, resp := range handlers {
			if strings.EqualFold(key, reqBody.OperationName) {
				_, _ = w.Write([]byte(resp))
				return
			}
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	return server, captured
}

func TestRunUpdate_EstimatePayload(t *testing.T) {
	t.Run("estimate value sent in mutation", func(t *testing.T) {
		server, captured := captureUpdateServer(t)
		defer server.Close()
		factory := func() (*linear.Client, error) {
			return linear.NewClient("test_api_key", linear.WithBaseURL(server.URL))
		}

		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--estimate=5"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if captured.Input == nil {
			t.Fatal("no UpdateIssue mutation captured")
		}
		if captured.Input["estimate"] != float64(5) {
			t.Errorf("estimate = %v (%T), want float64(5)", captured.Input["estimate"], captured.Input["estimate"])
		}
	})

	t.Run("estimate null sent in mutation when none", func(t *testing.T) {
		server, captured := captureUpdateServer(t)
		defer server.Close()
		factory := func() (*linear.Client, error) {
			return linear.NewClient("test_api_key", linear.WithBaseURL(server.URL))
		}

		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--estimate=none"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if captured.Input == nil {
			t.Fatal("no UpdateIssue mutation captured")
		}
		estimateVal, exists := captured.Input["estimate"]
		if !exists {
			t.Error("estimate key missing from mutation input")
		}
		if estimateVal != nil {
			t.Errorf("estimate = %v, want nil (explicit null)", estimateVal)
		}
	})
}
