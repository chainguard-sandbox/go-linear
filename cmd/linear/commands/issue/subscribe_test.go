package issue

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewSubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewSubscribeCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "subscribe <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "subscribe <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		if cmd.Flags().Lookup("user") == nil {
			t.Error("Expected flag \"user\" not found")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		if err := cmd.Args(cmd, []string{}); err == nil {
			t.Error("Expected error for no args")
		}
		if err := cmd.Args(cmd, []string{"ENG-123"}); err != nil {
			t.Errorf("Unexpected error for one arg: %v", err)
		}
	})
}

func TestRunSubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("subscribe to issue", func(t *testing.T) {
		cmd := NewSubscribeCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v", err)
		}
		if result["success"] != true {
			t.Error("Expected success: true")
		}
		if result["action"] != "subscribed" {
			t.Errorf("Expected action \"subscribed\", got %q", result["action"])
		}
	})

	t.Run("subscribe with user flag", func(t *testing.T) {
		cmd := NewSubscribeCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123", "--user=test@example.com"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v", err)
		}
		if result["success"] != true {
			t.Error("Expected success: true")
		}
	})
}

func TestNewUnsubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUnsubscribeCommand(factory)

	t.Run("command setup", func(t *testing.T) {
		if cmd.Use != "unsubscribe <id>" {
			t.Errorf("Use = %q, want %q", cmd.Use, "unsubscribe <id>")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		if cmd.Flags().Lookup("user") == nil {
			t.Error("Expected flag \"user\" not found")
		}
	})
}

func TestRunUnsubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()

	factory := testutil.TestFactory(t, server.URL)

	t.Run("unsubscribe from issue", func(t *testing.T) {
		cmd := NewUnsubscribeCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"ENG-123"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v", err)
		}
		if result["success"] != true {
			t.Error("Expected success: true")
		}
		if result["action"] != "unsubscribed" {
			t.Errorf("Expected action \"unsubscribed\", got %q", result["action"])
		}
	})
}
