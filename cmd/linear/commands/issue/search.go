package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewSearchCommand creates the issue search command.
func NewSearchCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search issues by text query",
		Long: `Search issues using full-text search.

Searches across issue titles and descriptions.

Examples:
  linear issue search "authentication bug"
  linear issue search "performance" --output=json
  linear issue search "login" --limit=10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSearch(cmd, client, args[0])
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of results to return")
	cmd.Flags().String("after", "", "Cursor for pagination")
	cmd.Flags().BoolP("include-archived", "a", false, "Include archived issues")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runSearch(cmd *cobra.Command, client *linear.Client, query string) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	includeArchived, _ := cmd.Flags().GetBool("include-archived")
	var includeArchivedPtr *bool
	if includeArchived {
		includeArchivedPtr = &includeArchived
	}

	// Search issues
	searchResult, err := client.SearchIssues(ctx, query, &first, afterPtr, nil, includeArchivedPtr)
	if err != nil {
		return fmt.Errorf("failed to search issues: %w", err)
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), searchResult, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Found %.0f issues\n\n", searchResult.TotalCount)
		for _, node := range searchResult.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", node.Identifier, node.Title)
			if node.State.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  State: %s\n", node.State.Name)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
