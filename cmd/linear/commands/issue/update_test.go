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
			"due-date", "milestone", "link-pr",
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
	Input      map[string]any
	LinkPRVars map[string]any
}

func captureUpdateServer(t *testing.T) (*httptest.Server, *updateCapture) {
	return captureUpdateServerWithLinkPR(t, `{"data":{"attachmentLinkGitHubPR":{"success":true,"attachment":{"id":"att-1"}}}}`)
}

func captureUpdateServerWithLinkPR(t *testing.T, linkPRResponse string) (*httptest.Server, *updateCapture) {
	t.Helper()
	captured := &updateCapture{}
	handlers := defaultHandlers()
	handlers["AttachmentLinkGitHubPR"] = linkPRResponse
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
		if strings.EqualFold(reqBody.OperationName, "AttachmentLinkGitHubPR") {
			captured.LinkPRVars = reqBody.Variables
		}
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

func TestRunUpdate_LinkPR(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantURL string
	}{
		{
			name:    "link-pr with estimate in nullable path",
			args:    []string{"ENG-123", "--estimate=5", "--link-pr=owner/repo#1"},
			wantURL: "https://github.com/owner/repo/pull/1",
		},
		{
			name:    "link-pr with assignee none in nullable path",
			args:    []string{"ENG-123", "--assignee=none", "--link-pr=owner/repo#2"},
			wantURL: "https://github.com/owner/repo/pull/2",
		},
		{
			name:    "link-pr with title in standard path",
			args:    []string{"ENG-123", "--title=x", "--link-pr=owner/repo#3"},
			wantURL: "https://github.com/owner/repo/pull/3",
		},
		{
			name:    "link-pr with hyphenated owner and dotted repo",
			args:    []string{"ENG-123", "--title=x", "--link-pr=my-org/repo.name#45"},
			wantURL: "https://github.com/my-org/repo.name/pull/45",
		},
		{
			name:    "link-pr with dot-leading repo",
			args:    []string{"ENG-123", "--title=x", "--link-pr=myorg/.github#12"},
			wantURL: "https://github.com/myorg/.github/pull/12",
		},
		{
			name:    "link-pr with full URL passed through unchanged",
			args:    []string{"ENG-123", "--estimate=5", "--link-pr=https://github.com/owner/repo/pull/4"},
			wantURL: "https://github.com/owner/repo/pull/4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, captured := captureUpdateServer(t)
			defer server.Close()
			factory := func() (*linear.Client, error) {
				return linear.NewClient("test_api_key", linear.WithBaseURL(server.URL))
			}

			cmd := NewUpdateCommand(factory)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if captured.LinkPRVars == nil {
				t.Fatal("no AttachmentLinkGitHubPR mutation captured")
			}
			if got := captured.LinkPRVars["url"]; got != tt.wantURL {
				t.Errorf("url = %v, want %q", got, tt.wantURL)
			}
			if got := captured.LinkPRVars["issueId"]; got != "issue-123" {
				t.Errorf("issueId = %v, want %q", got, "issue-123")
			}
		})
	}
}

// A short-format value that is not owner/repo#number must be rejected
// instead of being blindly prefixed with https://github.com/.
func TestRunUpdate_LinkPRInvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		linkPR string
	}{
		{name: "bare word", linkPR: "banana"},
		{name: "missing number", linkPR: "owner/repo"},
		{name: "non-numeric number", linkPR: "owner/repo#abc"},
		{name: "extra path segment", linkPR: "owner/repo/extra#1"},
		{name: "leading hyphen in owner", linkPR: "-owner/repo#1"},
		{name: "trailing hyphen in owner", linkPR: "owner-/repo#1"},
		{name: "schemeless full PR path", linkPR: "owner/repo/pull/5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, captured := captureUpdateServer(t)
			defer server.Close()
			factory := func() (*linear.Client, error) {
				return linear.NewClient("test_api_key", linear.WithBaseURL(server.URL))
			}

			cmd := NewUpdateCommand(factory)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"ENG-123", "--title=x", "--link-pr=" + tt.linkPR})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() succeeded, want invalid --link-pr error")
			}
			if !strings.Contains(err.Error(), "invalid --link-pr value") {
				t.Errorf("error = %v, want it to mention invalid --link-pr value", err)
			}
			if captured.LinkPRVars != nil {
				t.Errorf("AttachmentLinkGitHubPR was called with %v, want no call", captured.LinkPRVars)
			}
			if captured.Input != nil {
				t.Errorf("UpdateIssue ran with input %v, want no mutation for an invalid --link-pr", captured.Input)
			}
		})
	}
}

// A failed link mutation must surface as an error even though the issue
// update itself already succeeded.
func TestRunUpdate_LinkPRError(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "nullable path",
			args: []string{"ENG-123", "--estimate=5", "--link-pr=owner/repo#1"},
		},
		{
			name: "standard path",
			args: []string{"ENG-123", "--title=x", "--link-pr=owner/repo#1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, captured := captureUpdateServerWithLinkPR(t, `{"data":{"attachmentLinkGitHubPR":{"success":false}}}`)
			defer server.Close()
			factory := func() (*linear.Client, error) {
				return linear.NewClient("test_api_key", linear.WithBaseURL(server.URL))
			}

			cmd := NewUpdateCommand(factory)
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error when AttachmentLinkGitHubPR fails")
			}
			if !strings.Contains(err.Error(), "failed to link GitHub PR") {
				t.Errorf("error = %q, want it to mention failing to link the GitHub PR", err)
			}
			if captured.LinkPRVars == nil {
				t.Error("no AttachmentLinkGitHubPR mutation captured; error came from the wrong place")
			}
		})
	}
}
