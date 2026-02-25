package notification

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/testutil"
)

func TestNewNotificationCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewNotificationCommand(factory)

	if cmd.Use != "notification" {
		t.Errorf("Use = %q, want %q", cmd.Use, "notification")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestNewArchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewArchiveCommand(factory)

	if !strings.HasPrefix(cmd.Use, "archive") {
		t.Errorf("Use = %q, want prefix archive", cmd.Use)
	}
}

func TestRunArchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewArchiveCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"notif-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewUpdateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "update") {
		t.Errorf("Use = %q, want prefix update", cmd.Use)
	}
}

func TestRunUpdate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("mark read", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"notif-123", "--read"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewSubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewSubscribeCommand(factory)

	if cmd.Use != "subscribe" {
		t.Errorf("Use = %q, want %q", cmd.Use, "subscribe")
	}
}

func TestRunSubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("subscribe to project", func(t *testing.T) {
		cmd := NewSubscribeCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		// Use UUID format
		cmd.SetArgs([]string{"--project=00000000-0000-0000-0000-000000000001"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewUnsubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUnsubscribeCommand(factory)

	if !strings.HasPrefix(cmd.Use, "unsubscribe") {
		t.Errorf("Use = %q, want prefix unsubscribe", cmd.Use)
	}
}

func TestRunUnsubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewUnsubscribeCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"sub-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
