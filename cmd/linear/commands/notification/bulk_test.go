package notification

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewArchiveAllCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewArchiveAllCommand(factory)
	if cmd.Use != "archive-all" {
		t.Errorf("Use = %q, want %q", cmd.Use, "archive-all")
	}
}

func TestRunArchiveAll(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("archive all for issue", func(t *testing.T) {
		cmd := NewArchiveAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001"})
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

	t.Run("requires entity flag", func(t *testing.T) {
		cmd := NewArchiveAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error when no entity flag provided")
		}
	})

	t.Run("rejects multiple entity flags", func(t *testing.T) {
		cmd := NewArchiveAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{
			"--issue=00000000-0000-0000-0000-000000000001",
			"--project=00000000-0000-0000-0000-000000000002",
		})
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error when multiple entity flags provided")
		}
	})
}

func TestRunMarkReadAll(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewMarkReadAllCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001"})
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
}

func TestRunMarkUnreadAll(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewMarkUnreadAllCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001"})
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
}

func TestRunSnoozeAll(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("snooze with future duration", func(t *testing.T) {
		cmd := NewSnoozeAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001", "--until=3d"})
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

	t.Run("snooze with tomorrow", func(t *testing.T) {
		cmd := NewSnoozeAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001", "--until=tomorrow"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("rejects past ISO date", func(t *testing.T) {
		cmd := NewSnoozeAllCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001", "--until=2020-01-01"})
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error for past --until value")
		}
	})
}

func TestRunUnsnoozeAll(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewUnsnoozeAllCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--issue=00000000-0000-0000-0000-000000000001"})
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
}
