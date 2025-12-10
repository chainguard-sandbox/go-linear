package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

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
		Long: `List issues from Linear with comprehensive filtering options.

Supports filtering by team, assignee, state, priority, dates, and labels.
AI agents can use parameter-rich flags for precise queries.

Parameters:
  --team: Team name (e.g., "Engineering") or UUID - use 'linear team list' to discover names
  --assignee: User email, name, or 'me' - use 'linear user list' to discover users
  --state: State name (e.g., "In Progress") - use 'linear state list' for available states
  --priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low
  --completed-after/before: ISO8601 (2025-12-10), relative (yesterday, today), or duration (7d, 2w, 3m)
  --label: Label names - use 'linear label list' to discover available labels

Output (--output=json):
  Returns JSON with: nodes (array of issues), pageInfo (pagination info)
  Each issue has: identifier, title, state, assignee, priority, team, dates

Examples:
  # List my urgent issues
  linear issue list --assignee=me --priority=1

  # Find completed issues from yesterday
  linear issue list --team=Engineering --completed-after=yesterday --completed-before=today --output=json

  # List issues in specific state
  linear issue list --state="In Progress" --team=Engineering

  # Complex multi-filter query
  linear issue list --team=Engineering --priority=1 --created-after=7d --label=bug --output=json

Common Errors:
  - "team not found": Check spelling or use 'linear team list' to see available teams
  - "ambiguous match": Use team key (e.g., "ENG") instead of full name`,
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
	var nodes []interface{}
	var _ interface{} // pageInfo unused for now

	if issueFilter != nil {
		// Use SearchIssues with empty query and filters
		searchResult, err := client.SearchIssues(ctx, "", &first, afterPtr, issueFilter, nil)
		if err != nil {
			return fmt.Errorf("failed to search issues: %w", err)
		}

		// Convert search nodes to list format for display
		// SearchIssues returns different node types but with same structure
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "json":
			return formatter.FormatJSON(cmd.OutOrStdout(), searchResult, true)
		case "table":
			// For table, we need to display the search results
			// The structure is similar but uses SearchIssues_SearchIssues_Nodes
			fmt.Fprintf(cmd.OutOrStdout(), "Found %d issues\n\n", len(searchResult.Nodes))
			// Format the search results in a simple way for now
			for _, node := range searchResult.Nodes {
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
		nodes = make([]interface{}, len(issues.Nodes))
		for i, n := range issues.Nodes {
			nodes[i] = n
		}
		_ = issues.PageInfo // Unused for now

		// Format output
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "json":
			return formatter.FormatJSON(cmd.OutOrStdout(), issues, true)
		case "table":
			return formatter.FormatIssuesTable(cmd.OutOrStdout(), issues.Nodes)
		default:
			return fmt.Errorf("unsupported output format: %s (supported: json, table)", output)
		}
	}
}
