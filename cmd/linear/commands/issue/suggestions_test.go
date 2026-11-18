package issue

import (
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func TestSuggestionsCommand_Removed(t *testing.T) {
	cmd := NewSuggestionsCommand(func() (*linear.Client, error) {
		t.Fatal("issue suggestions must not create a client: the API was removed")
		return nil, nil
	})
	if !cmd.Hidden {
		t.Error("issue suggestions should be hidden")
	}
	// Old flags must still parse into the removal error, not "unknown flag".
	_, err := runCmd(t, cmd, "ENG-123", "--include-dismissed", "--limit=5")
	if err == nil || !strings.Contains(err.Error(), "issue suggestions was removed from the Linear API") {
		t.Fatalf("error = %v, want the removal explanation", err)
	}
	if !strings.Contains(err.Error(), "--has-suggested-assignees") {
		t.Errorf("error should point to the --has-suggested-* filters: %v", err)
	}
}
