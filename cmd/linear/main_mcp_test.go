package main

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands"
)

// findCommand walks the command tree and returns the first command whose
// immediate name matches.
func findCommand(root *cobra.Command, name string) *cobra.Command {
	if root.Name() == name {
		return root
	}
	for _, sub := range root.Commands() {
		if got := findCommand(sub, name); got != nil {
			return got
		}
	}
	return nil
}

// TestMCPExclusion guards the security-relevant invariant that the `graphql`
// escape hatch is NOT exposed as an MCP tool, while ordinary commands are. A
// command is exposed only when a selector's CmdSelector returns true; with our
// single ExcludeCmdsContaining("graphql") selector that means "path must not
// contain graphql".
func TestMCPExclusion(t *testing.T) {
	root := commands.NewRootCommand()
	selectors := mcpSelectors()
	if len(selectors) == 0 || selectors[0].CmdSelector == nil {
		t.Fatal("mcpSelectors() must define a CmdSelector")
	}
	sel := selectors[0].CmdSelector

	graphqlCmd := findCommand(root, "graphql")
	if graphqlCmd == nil {
		t.Fatal("graphql command not registered on the root command")
	}
	if sel(graphqlCmd) {
		t.Errorf("graphql command (%q) must be EXCLUDED from MCP tools but the selector accepted it", graphqlCmd.CommandPath())
	}

	// A representative ordinary command must remain exposed (guards against an
	// over-broad exclusion).
	issueCmd := findCommand(root, "issue")
	if issueCmd == nil {
		t.Fatal("issue command not registered on the root command")
	}
	if !sel(issueCmd) {
		t.Errorf("issue command (%q) must remain exposed as an MCP tool but the selector rejected it", issueCmd.CommandPath())
	}
}
