package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewSuggestionsCommand creates the issue suggestions command.
func NewSuggestionsCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "suggestions <issue-id>",
		Short: "List AI suggestions for an issue",
		Long: `List AI suggestions for issue. Shows suggested assignees, teams, labels, projects. Excludes dismissed by default.

Example: go-linear issue suggestions ENG-123 --output=json

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
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

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

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"issue":       suggestions.Identifier,
			"suggestions": activeSuggestions,
		}, true)
	case "table":
		if len(activeSuggestions) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No AI suggestions")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "AI Suggestions for %s:\n\n", suggestions.Identifier)

		for _, s := range activeSuggestions {
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: ", s.State, s.Type)

			//nolint:exhaustive // Handle known types, default handles rest
			switch string(s.Type) {
			case "assignee":
				if s.SuggestedUser != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", s.SuggestedUser.Name)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "")
				}
			case "team":
				if s.SuggestedTeam != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", s.SuggestedTeam.Name, s.SuggestedTeam.Key)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "")
				}
			case "label":
				if s.SuggestedLabel != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", s.SuggestedLabel.Name)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "")
				}
			case "project":
				if s.SuggestedProject != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", s.SuggestedProject.Name)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "")
				}
			case "relatedIssue", "similarIssue":
				if s.SuggestedIssue != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s - %s\n", s.SuggestedIssue.Identifier, s.SuggestedIssue.Title)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "")
				}
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "(type: %s)\n", s.Type)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
