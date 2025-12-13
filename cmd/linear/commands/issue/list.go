package issue

import (
	"context"
	"fmt"

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
		Long: `List issues with filtering. Returns 8 default fields per issue.

Filters: --team (name/key), --assignee=me (current user) or email, --state (name), --priority (0-4), --label (name), --created-after=yesterday|7d|2w|2025-12-10

Pagination: --limit (default 50), --after (cursor from pageInfo.endCursor)

Example: go-linear issue list --team=ENG --assignee=me --priority=1 --completed-after=7d --output=json

Returns: {nodes: [{8 issue fields}...], pageInfo: {hasNextPage, endCursor}}
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

	// Labels/Projects
	cmd.Flags().StringArray("label", []string{}, "Label names (repeatable)")

	// Output
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

	// Get pagination parameters
	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	// Query issues - use SearchIssues if we have filters, otherwise use Issues
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
