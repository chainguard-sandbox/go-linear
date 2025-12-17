package issue

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/filter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the issue list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues with filtering",
		Long: `List issues with filtering. Returns 8 default fields per issue. Use --count for totals (99% fewer tokens).

Filters: --team (name/key), --assignee=me (current user) or email, --state (name), --priority (0-4), --label (name), --created-after=yesterday|7d|2w|2025-12-10

Pagination: --limit (default 50), --after (cursor from pageInfo.endCursor)
Count: --count returns just {"count": N} instead of full results

Example: go-linear issue list --team=ENG --assignee=me --priority=1 --output=json

Returns: {nodes: [{8 issue fields}...], pageInfo: {hasNextPage, endCursor}} or {"count": N} with --count
Related: issue_get, issue_create, team_list, user_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	// Pagination
	cmd.Flags().IntP("limit", "l", 50, "Number of issues to return")
	cmd.Flags().String("after", "", "Cursor for pagination")

	// Identity filtering (AI-friendly names)
	cmd.Flags().String("team", "", "Team name or ID (e.g., 'Engineering')")
	cmd.Flags().String("assignee", "", "Assignee name, email, or ID (e.g., 'alice@company.com', 'me')")
	cmd.Flags().String("state", "", "State name or ID (e.g., 'In Progress')")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")

	// Date filtering (AI-friendly)
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("completed-after", "", "Completed after date")
	cmd.Flags().String("completed-before", "", "Completed before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	// Collection filters (repeatable, OR logic)
	cmd.Flags().StringArray("label", []string{}, "Label names (repeatable)")
	cmd.Flags().StringArray("attachment-by", []string{}, "Has attachments by user (name/email/'me', repeatable)")
	cmd.Flags().StringArray("comment-by", []string{}, "Has comments by user (name/email/'me', repeatable)")
	cmd.Flags().StringArray("subscriber", []string{}, "Has subscriber (name/email/'me', repeatable)")

	// Additional date filters (alphabetical)
	cmd.Flags().String("added-to-cycle-after", "", "Added to cycle after date")
	cmd.Flags().String("added-to-cycle-before", "", "Added to cycle before date")
	cmd.Flags().String("archived-after", "", "Archived after date")
	cmd.Flags().String("archived-before", "", "Archived before date")
	cmd.Flags().String("auto-archived-after", "", "Auto-archived after date")
	cmd.Flags().String("auto-archived-before", "", "Auto-archived before date")
	cmd.Flags().String("auto-closed-after", "", "Auto-closed after date")
	cmd.Flags().String("auto-closed-before", "", "Auto-closed before date")
	cmd.Flags().String("canceled-after", "", "Canceled after date")
	cmd.Flags().String("canceled-before", "", "Canceled before date")
	cmd.Flags().String("started-after", "", "Started after date")
	cmd.Flags().String("started-before", "", "Started before date")
	cmd.Flags().String("triaged-after", "", "Triaged after date")
	cmd.Flags().String("triaged-before", "", "Triaged before date")
	cmd.Flags().String("due-after", "", "Due after date")
	cmd.Flags().String("due-before", "", "Due before date")
	cmd.Flags().String("snoozed-until-after", "", "Snoozed until after date")
	cmd.Flags().String("snoozed-until-before", "", "Snoozed until before date")

	// Entity filters (alphabetical)
	cmd.Flags().String("added-to-cycle-period", "", "When added to cycle: before, during, after")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("cycle", "", "Cycle UUID")
	cmd.Flags().String("delegate", "", "Delegated agent name, email, or 'me'")
	cmd.Flags().String("description", "", "Text in description")
	cmd.Flags().String("id", "", "Issue ID (UUID)")
	cmd.Flags().String("last-applied-template", "", "Last applied template UUID")
	cmd.Flags().String("parent", "", "Parent issue ID")
	cmd.Flags().String("project", "", "Project UUID")
	cmd.Flags().String("project-milestone", "", "Project milestone UUID")
	cmd.Flags().String("snoozed-by", "", "Who snoozed (name, email, or 'me')")
	cmd.Flags().String("title", "", "Text in title")

	// Numeric filters
	cmd.Flags().Int("estimate", -1, "Story points/estimate")
	cmd.Flags().Int("number", -1, "Issue number")

	// Boolean filters (alphabetical)
	cmd.Flags().Bool("has-blocked-by", false, "Has blocked-by relations")
	cmd.Flags().Bool("has-blocking", false, "Has blocking relations")
	cmd.Flags().Bool("has-children", false, "Has sub-issues")
	cmd.Flags().Bool("has-duplicate", false, "Has duplicate relations")
	cmd.Flags().Bool("has-needs", false, "Has customer needs")
	cmd.Flags().Bool("has-reactions", false, "Has reactions")
	cmd.Flags().Bool("has-related", false, "Has related relations")
	cmd.Flags().Bool("has-suggested-assignees", false, "Has AI assignee suggestions")
	cmd.Flags().Bool("has-suggested-labels", false, "Has AI label suggestions")
	cmd.Flags().Bool("has-suggested-projects", false, "Has AI project suggestions")
	cmd.Flags().Bool("has-suggested-teams", false, "Has AI team suggestions")

	// Output
	cmd.Flags().Bool("count", false, "Return only count, not results (99% token reduction)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,identifier,title,url,state.name,team.key,priority,createdAt) | none | defaults,extra | id,title,...")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := filter.NewIssueFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	issueFilter := filterBuilder.Build()

	// Check if count mode
	countMode, _ := cmd.Flags().GetBool("count")

	if countMode {
		// Count mode: return just the total count
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

		// Return count
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

	// Normal list mode: Get pagination parameters
	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	// Query issues - use IssuesFiltered if we have filters, otherwise use Issues
	var nodes []any
	var _ any // pageInfo unused for now

	if issueFilter != nil {
		// Use IssuesFiltered with filters (no search term required)
		filteredResult, err := client.IssuesFiltered(ctx, &first, afterPtr, issueFilter)
		if err != nil {
			return fmt.Errorf("failed to filter issues: %w", err)
		}

		// Convert filtered nodes to generic format
		nodes = make([]any, len(filteredResult.Nodes))
		for i, n := range filteredResult.Nodes {
			nodes[i] = n
		}

		// Format output
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

			// Get command defaults
			defaults := fieldfilter.GetDefaults("issue.list", configOverrides)

			// Parse field selector with defaults
			// For list commands, fields apply to items in nodes array, not the wrapper
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}

			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), filteredResult, true, fieldSelector)
		case "table":
			// Display filtered results
			if len(filteredResult.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No issues found")
				return nil
			}
			for _, node := range filteredResult.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", node.Identifier, node.Title)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s (supported: json, table)", output)
		}
	} else {
		// No filters - use regular Issues query
		issues, err := client.Issues(ctx, &first, afterPtr)
		if err != nil {
			return fmt.Errorf("failed to list issues: %w", err)
		}
		nodes = make([]any, len(issues.Nodes))
		for i, n := range issues.Nodes {
			nodes[i] = n
		}
		_ = issues.PageInfo // Unused for now

		// Format output
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

			// Get command defaults
			defaults := fieldfilter.GetDefaults("issue.list", configOverrides)

			// Parse field selector with defaults
			// For list commands, fields apply to items in nodes array, not the wrapper
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}

			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), issues, true, fieldSelector)
		case "table":
			return formatter.FormatIssuesTable(cmd.OutOrStdout(), issues.Nodes)
		default:
			return fmt.Errorf("unsupported output format: %s (supported: json, table)", output)
		}
	}
}
