package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewSuggestionsCommand creates the issue suggestions command.
func NewSuggestionsCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "suggestions <issue-id>",
		Short: "List AI suggestions for an issue",
		Long: `List AI suggestions for issue. Shows suggested assignees, teams, labels, projects. Excludes dismissed by default.

Example: go-linear issue suggestions ENG-123

Related: issue_get, issue_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSuggestions(cmd, client, args[0], paginationFlags)
		},
	}

	cmd.Flags().Bool("include-dismissed", false, "Include dismissed suggestions")
	paginationFlags.Bind(cmd, 50)

	return cmd
}

func runSuggestions(cmd *cobra.Command, client *linear.Client, issueInput string, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve issue
	issueID, err := res.ResolveIssue(ctx, issueInput)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	first := paginationFlags.LimitPtr()
	suggestions, err := client.IssueSuggestions(ctx, issueID, first, nil)
	if err != nil {
		return fmt.Errorf("failed to get issue suggestions: %w", err)
	}

	includeDismissed, _ := cmd.Flags().GetBool("include-dismissed")

	// Filter out dismissed unless requested
	activeSuggestions := suggestions.IncomingSuggestions.Nodes
	if !includeDismissed {
		filtered := make([]*intgraphql.GetIssueSuggestionsForIssue_Issue_IncomingSuggestions_Nodes, 0, len(activeSuggestions))
		for _, s := range activeSuggestions {
			if s.State != "dismissed" {
				filtered = append(filtered, s)
			}
		}
		activeSuggestions = filtered
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"issue":       suggestions.Identifier,
		"suggestions": activeSuggestions,
	}, true)
}
