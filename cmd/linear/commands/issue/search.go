package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewSearchCommand creates the issue search command.
func NewSearchCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search issues by text query",
		Long: `Search issues by text. Returns 8 default fields per result. Searches titles and descriptions. Use --count for totals.

Example: go-linear issue search "authentication bug" --limit=20 --output=json

Count: --count returns {"count": N} (see issue_list for details)
Related: issue_list, issue_get`,
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
	cmd.Flags().Bool("count", false, "Return only count, not results (99% token reduction)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,identifier,title,url,state.name,team.key,priority,createdAt) | none | defaults,extra")

	return cmd
}

func runSearch(cmd *cobra.Command, client *linear.Client, query string) error {
	ctx := cmd.Context()

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

	// Check if count mode
	countMode, _ := cmd.Flags().GetBool("count")
	if countMode {
		// Return just the count
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "json":
			result := map[string]any{
				"count": searchResult.TotalCount,
			}
			return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
		case "table":
			fmt.Fprintf(cmd.OutOrStdout(), "%.0f\n", searchResult.TotalCount)
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// Format output (normal mode)
	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults (use same as issue.list)
		defaults := fieldfilter.GetDefaults("issue.list", configOverrides)

		// Parse field selector with defaults (search returns same structure as list)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), searchResult, true, fieldSelector)
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
