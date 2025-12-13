package issue

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCountCommand creates the issue count command.
func NewCountCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count issues with filtering",
		Long: `Count issues. Returns total count (99% less tokens than listing).

Uses same filters as issue_list. Returns just the count, not issue details.

Filters: --team, --assignee=me, --priority (0-4), --state, --label, --created-after=yesterday|7d

Example: go-linear issue count --team=ENG --priority=1 --state="In Progress" --output=json

Related: issue_list, issue_search`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCount(cmd, client)
		},
	}

	// Same filter flags as issue list
	cmd.Flags().String("team", "", "Team name or ID")
	cmd.Flags().String("assignee", "", "Assignee name, email, or 'me'")
	cmd.Flags().String("state", "", "State name or ID")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	cmd.Flags().String("created-after", "", "Created after date")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("completed-after", "", "Completed after date")
	cmd.Flags().String("completed-before", "", "Completed before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")
	cmd.Flags().StringArray("label", []string{}, "Label names (repeatable)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCount(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build filter from flags (same as issue list)
	filterBuilder := filter.NewIssueFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	issueFilter := filterBuilder.Build()

	count := 0

	if issueFilter != nil {
		// With filters: Paginate through all filtered results
		var after *string
		pageSize := int64(100)

		for {
			result, err := client.IssuesFiltered(ctx, &pageSize, after, issueFilter)
			if err != nil {
				return fmt.Errorf("failed to count issues: %w", err)
			}

			count += len(result.Nodes)

			if !result.PageInfo.HasNextPage {
				break
			}
			after = result.PageInfo.EndCursor
		}
	} else {
		// No filters: Use iterator
		it := linear.NewIssueIterator(client, 100)
		for {
			_, err := it.Next(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to iterate issues: %w", err)
			}
			count++
		}
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		result := map[string]any{
			"count": count,
		}
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", count)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
