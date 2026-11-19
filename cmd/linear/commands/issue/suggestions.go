package issue

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// NewSuggestionsCommand creates the hidden issue suggestions command.
//
// Linear removed Issue.incomingSuggestions and the IssueSuggestion type from
// its API in @linear/sdk 103.0.0, so there is nothing left to list. The
// command stays, hidden, so scripts get an explanation instead of "unknown
// command". It makes no API request.
func NewSuggestionsCommand(_ cli.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:                "suggestions <issue-id>",
		Short:              "Removed: Linear dropped issue suggestions from its API",
		Hidden:             true,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return cli.ErrRemovedUpstream("issue suggestions",
				"use 'issue list --has-suggested-assignees' (or --has-suggested-teams, -labels, -projects) to find issues that have suggestions")
		},
	}
}
