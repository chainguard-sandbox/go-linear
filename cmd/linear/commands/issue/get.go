package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the issue get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single issue by ID",
		Long: `Get detailed information about a specific issue.

Retrieve full issue details including description, state, assignee, labels, and timestamps.

Parameters:
  <id>: Issue identifier (e.g., ENG-123) or UUID (required)

Output (--output=json):
  Returns JSON with: id, identifier, title, description, state, assignee, priority, labels, createdAt, updatedAt

Examples:
  # Get issue by identifier
  linear issue get ENG-123

  # Get issue with JSON output
  linear issue get ENG-123 --output=json

TIP: Use 'linear issue list' to discover issue identifiers

Related Commands:
  - linear issue list - List and filter issues
  - linear issue update - Modify issue fields`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := context.Background()

	issue, err := client.Issue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), issue, true)
	case "table":
		// Simple table output for single issue
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", issue.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title:       %s\n", issue.Title)
		fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", issue.State.Name)
		if issue.Assignee != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Assignee:    %s\n", issue.Assignee.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Priority:    %.0f\n", issue.Priority)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", issue.UpdatedAt)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
