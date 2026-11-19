package document

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// Linear removed initiativeId and teamId from DocumentCreateInput in
// @linear/sdk 103.0.0.
func TestRunCreate_RemovedLinkFlags(t *testing.T) {
	for _, flag := range []string{"initiative", "team"} {
		t.Run(flag, func(t *testing.T) {
			cmd := NewCreateCommand(func() (*linear.Client, error) {
				t.Fatal("no client should be created for a removed flag")
				return nil, nil
			})
			if f := cmd.Flags().Lookup(flag); f == nil || !f.Hidden {
				t.Fatalf("--%s should exist and be hidden", flag)
			}
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{"--title=Doc", "--" + flag + "=X"})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), "--"+flag+" was removed from the Linear API") {
				t.Fatalf("error = %v, want the removal explanation", err)
			}
			if !strings.Contains(err.Error(), "--project") {
				t.Errorf("error should suggest --project or --issue: %v", err)
			}
		})
	}
}
